package metrics

import (
	"context"
	"errors"
	"strconv"
	"time"

	kratosErrors "github.com/go-kratos/kratos/v2/errors"

	pkgErrors "github.com/gaoyong06/go-pkg/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HTTPStatusCodeGetter interface {
	// GetHTTPStatusCode 将业务错误映射为 HTTP 标准状态码（100-599）。
	// 用途：用于将 Kratos error / 业务错误码转换为 Prometheus 采集所需的 status 标签。
	GetHTTPStatusCode(err error) int
}

type DefaultHTTPStatusCodeGetter struct {
	serviceName string
}

func NewDefaultHTTPStatusCodeGetter(serviceName string) HTTPStatusCodeGetter {
	return &DefaultHTTPStatusCodeGetter{serviceName: serviceName}
}

func (g *DefaultHTTPStatusCodeGetter) GetHTTPStatusCode(err error) int {
	if err == nil {
		return 200
	}

	kerr := kratosErrors.FromError(err)
	if kerr == nil {
		return 500
	}

	code := int(kerr.Code)

	if mapping := pkgErrors.GetHTTPStatusMapping(g.serviceName); len(mapping) > 0 {
		if status, ok := mapping[code]; ok {
			return status
		}
	}

	if status, ok := pkgErrors.DefaultHTTPStatusMapping()[code]; ok {
		return status
	}

	if code >= 100 && code <= 599 {
		return code
	}

	return 500
}

type Prometheus struct {
	registry                *prometheus.Registry
	serviceName             string
	statusCodeGetter        HTTPStatusCodeGetter
	httpRequestsTotal       *prometheus.CounterVec
	httpRequestDurationSecs *prometheus.HistogramVec
	httpInFlight            prometheus.Gauge
}

// NewPrometheus 创建 Prometheus 指标采集器。
// 默认会注册“服务自身维度”的基础指标：
// - Go runtime：GC、goroutines、heap 等
// - process：CPU、RSS、fd 等
// - build_info：构建信息
//
// 同时初始化 HTTP 指标：
// - xinyuan_http_requests_total：请求总量（按 service/method/operation/status）
// - xinyuan_http_request_duration_seconds：请求耗时直方图（同维度）
// - xinyuan_http_in_flight_requests：当前并发中的请求数
func NewPrometheus(serviceName string, statusCodeGetter HTTPStatusCodeGetter) *Prometheus {
	registry := prometheus.NewRegistry()
	registerStandardCollectors(registry)

	factory := promauto.With(registry)

	return &Prometheus{
		registry:         registry,
		serviceName:      serviceName,
		statusCodeGetter: statusCodeGetter,
		httpRequestsTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "xinyuan_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"service", "method", "operation", "status"},
		),
		httpRequestDurationSecs: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "xinyuan_http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service", "method", "operation", "status"},
		),
		httpInFlight: factory.NewGauge(
			prometheus.GaugeOpts{
				Name: "xinyuan_http_in_flight_requests",
				Help: "Current number of in-flight HTTP requests",
			},
		),
	}
}

func NewPrometheusDefaultRegistry(serviceName string, statusCodeGetter HTTPStatusCodeGetter) *Prometheus {
	registry, ok := prometheus.DefaultRegisterer.(*prometheus.Registry)
	if !ok || registry == nil {
		registry = prometheus.NewRegistry()
	}
	registerStandardCollectors(registry)

	factory := promauto.With(registry)

	return &Prometheus{
		registry:         registry,
		serviceName:      serviceName,
		statusCodeGetter: statusCodeGetter,
		httpRequestsTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "xinyuan_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"service", "method", "operation", "status"},
		),
		httpRequestDurationSecs: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "xinyuan_http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service", "method", "operation", "status"},
		),
		httpInFlight: factory.NewGauge(
			prometheus.GaugeOpts{
				Name: "xinyuan_http_in_flight_requests",
				Help: "Current number of in-flight HTTP requests",
			},
		),
	}
}

// Registry 返回该采集器使用的 registry。
// 用途：如果服务想把指标挂到独立端口或与其他 registry 合并，可以直接复用该 registry。
func (p *Prometheus) Registry() *prometheus.Registry {
	return p.registry
}

// HTTPMiddleware 返回 Kratos middleware，用于采集 HTTP 请求指标。
// 标签维度：
// - service：服务名
// - method：HTTP 方法
// - operation：Kratos Operation（通常是路由/接口名）
// - status：HTTP 状态码（来自错误映射；成功默认 200）
func (p *Prometheus) HTTPMiddleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			method, operation := p.extractHTTPLabels(ctx)
			start := time.Now()
			p.httpInFlight.Inc()
			defer func() {
				p.httpInFlight.Dec()

				status := 500
				if recovered := recover(); recovered != nil {
					elapsed := time.Since(start).Seconds()
					statusLabel := strconv.Itoa(status)
					p.httpRequestsTotal.WithLabelValues(p.serviceName, method, operation, statusLabel).Inc()
					p.httpRequestDurationSecs.WithLabelValues(p.serviceName, method, operation, statusLabel).Observe(elapsed)
					panic(recovered)
				}

				status = p.deriveHTTPStatusCode(err)
				elapsed := time.Since(start).Seconds()
				statusLabel := strconv.Itoa(status)
				p.httpRequestsTotal.WithLabelValues(p.serviceName, method, operation, statusLabel).Inc()
				p.httpRequestDurationSecs.WithLabelValues(p.serviceName, method, operation, statusLabel).Observe(elapsed)
			}()

			reply, err = handler(ctx, req)
			return reply, err
		}
	}
}

// RegisterHTTPMetricsEndpoint 在 Kratos HTTP Server 上注册 /metrics 端点。
// 该端点会输出该采集器 registry 中注册的全部指标。
func (p *Prometheus) RegisterHTTPMetricsEndpoint(srv *kratoshttp.Server) {
	handler := promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{})
	srv.HandleFunc("/metrics", handler.ServeHTTP)
}

func registerStandardCollectors(registry *prometheus.Registry) {
	collectors := []prometheus.Collector{
		prometheus.NewGoCollector(),
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
		prometheus.NewBuildInfoCollector(),
	}
	for _, c := range collectors {
		if err := registry.Register(c); err != nil {
			var alreadyRegistered prometheus.AlreadyRegisteredError
			if errors.As(err, &alreadyRegistered) {
				continue
			}
			panic(err)
		}
	}
}

// extractHTTPLabels 从 transport context 中提取 method 与 operation。
// 如果无法获取 transport 或请求信息，则返回 UNKNOWN。
func (p *Prometheus) extractHTTPLabels(ctx context.Context) (method string, operation string) {
	method = "UNKNOWN"
	operation = "UNKNOWN"

	tr, ok := transport.FromServerContext(ctx)
	if !ok || tr == nil {
		return method, operation
	}

	operation = tr.Operation()

	if ht, ok := tr.(*kratoshttp.Transport); ok && ht.Request() != nil {
		method = ht.Request().Method
	}

	return method, operation
}

// deriveHTTPStatusCode 根据业务错误推导 HTTP 状态码，用于 metrics 的 status 标签。
// - err == nil：200
// - 未提供 statusCodeGetter：500
// - statusCodeGetter 返回非合法范围：500
func (p *Prometheus) deriveHTTPStatusCode(err error) int {

	// 成功兜底 ： err == nil 明确返回 200（不依赖业务映射）
	if err == nil {
		return 200
	}

	// 依赖兜底 ： statusCodeGetter == nil 时返回 500（避免空指针/无标签）
	if p.statusCodeGetter == nil {
		return 500
	}

	// 合法性兜底 ：把 GetHTTPStatusCode 的返回值裁剪到 100–599，否则回退 500
	code := 500
	func() {
		defer func() {
			if recover() != nil {
				code = 500
			}
		}()
		code = p.statusCodeGetter.GetHTTPStatusCode(err)
	}()
	if code < 100 || code > 599 {
		return 500
	}
	return code
}
