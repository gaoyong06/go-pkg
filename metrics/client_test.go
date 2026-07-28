package metrics

import (
	"context"
	"testing"

	pkgresponse "github.com/gaoyong06/go-pkg/middleware/response"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/metadata"
)

func TestClientMetricsRecordsUpstreamResult(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewClientMetrics("console", registry)
	handler := metrics.Middleware("passport")(func(context.Context, any) (any, error) {
		return "ok", nil
	})

	if _, err := handler(context.Background(), nil); err != nil {
		t.Fatalf("handler() error = %v", err)
	}
	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v", err)
	}
	if len(families) != 2 {
		t.Fatalf("metric families = %d, want 2", len(families))
	}
}

func TestClientMetricsPropagatesTraceID(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewClientMetrics("console", registry)
	handler := metrics.Middleware("passport")(func(ctx context.Context, _ any) (any, error) {
		outgoing, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			t.Fatal("outgoing metadata is missing")
		}
		values := outgoing.Get("x-trace-id")
		if len(values) != 1 || values[0] != "trace-1" {
			t.Fatalf("trace metadata = %v", values)
		}
		return nil, nil
	})
	ctx := pkgresponse.SetTraceIdToContext(context.Background(), "trace-1")
	if _, err := handler(ctx, nil); err != nil {
		t.Fatalf("handler() error = %v", err)
	}
}
