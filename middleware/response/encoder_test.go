package response

import (
	"net/http/httptest"
	"testing"
)

func TestTraceIDFromRequestUsesPropagatedHeader(t *testing.T) {
	request := httptest.NewRequest("GET", "/", nil)
	request.Header.Set("X-Trace-Id", "trace-1")
	if traceID := traceIDFromRequest(request); traceID != "trace-1" {
		t.Fatalf("trace ID = %q, want trace-1", traceID)
	}
}
