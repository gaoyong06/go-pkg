package metrics

import (
	"context"
	"time"

	pkgresponse "github.com/gaoyong06/go-pkg/middleware/response"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/metadata"
	grpcstatus "google.golang.org/grpc/status"
)

type ClientMetrics struct {
	serviceName string
	requests    *prometheus.CounterVec
	duration    *prometheus.HistogramVec
}

func NewClientMetricsDefaultRegistry(serviceName string) *ClientMetrics {
	return NewClientMetrics(serviceName, prometheus.DefaultRegisterer)
}

func NewClientMetrics(serviceName string, registerer prometheus.Registerer) *ClientMetrics {
	requests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "xinyuan_upstream_requests_total",
		Help: "Total number of upstream client requests",
	}, []string{"service", "upstream", "operation", "result"})
	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "xinyuan_upstream_request_duration_seconds",
		Help:    "Duration of upstream client requests in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"service", "upstream", "operation", "result"})

	requests = registerCounterVec(registerer, requests)
	duration = registerHistogramVec(registerer, duration)
	return &ClientMetrics{serviceName: serviceName, requests: requests, duration: duration}
}

func (m *ClientMetrics) Middleware(upstream string) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (reply any, err error) {
			if traceID := pkgresponse.GetTraceIdFromContext(ctx); traceID != "" {
				ctx = metadata.AppendToOutgoingContext(ctx, "x-trace-id", traceID)
			}
			operation := "UNKNOWN"
			if clientTransport, ok := transport.FromClientContext(ctx); ok && clientTransport.Operation() != "" {
				operation = clientTransport.Operation()
			}
			start := time.Now()
			defer func() {
				result := grpcstatus.Code(err).String()
				if recovered := recover(); recovered != nil {
					result = "PANIC"
					m.Observe(upstream, operation, result, time.Since(start))
					panic(recovered)
				}
				m.Observe(upstream, operation, result, time.Since(start))
			}()
			return handler(ctx, req)
		}
	}
}

func (m *ClientMetrics) Observe(upstream, operation, result string, elapsed time.Duration) {
	m.requests.WithLabelValues(m.serviceName, upstream, operation, result).Inc()
	m.duration.WithLabelValues(m.serviceName, upstream, operation, result).Observe(elapsed.Seconds())
}

func registerCounterVec(registerer prometheus.Registerer, collector *prometheus.CounterVec) *prometheus.CounterVec {
	if err := registerer.Register(collector); err != nil {
		if existing, ok := err.(prometheus.AlreadyRegisteredError); ok {
			if counter, ok := existing.ExistingCollector.(*prometheus.CounterVec); ok {
				return counter
			}
		}
		panic(err)
	}
	return collector
}

func registerHistogramVec(registerer prometheus.Registerer, collector *prometheus.HistogramVec) *prometheus.HistogramVec {
	if err := registerer.Register(collector); err != nil {
		if existing, ok := err.(prometheus.AlreadyRegisteredError); ok {
			if histogram, ok := existing.ExistingCollector.(*prometheus.HistogramVec); ok {
				return histogram
			}
		}
		panic(err)
	}
	return collector
}
