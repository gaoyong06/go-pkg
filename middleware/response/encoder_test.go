package response

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	kratosErrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/require"
)

func TestTraceIDFromRequestUsesPropagatedHeader(t *testing.T) {
	request := httptest.NewRequest("GET", "/", nil)
	request.Header.Set("X-Trace-Id", "trace-1")
	if traceID := traceIDFromRequest(request); traceID != "trace-1" {
		t.Fatalf("trace ID = %q, want trace-1", traceID)
	}
}

func TestNewErrorEncoderUsesRequestLocaleResolver(t *testing.T) {
	handler := NewDefaultErrorHandler()
	encoder := NewErrorEncoder(
		handler,
		WithErrorMessageResolver(func(r *http.Request, err error, _ bool) string {
			if r.Header.Get("Accept-Language") == "zh-CN" {
				return "本地化错误"
			}
			return "Localized error"
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "zh-CN")
	resp := httptest.NewRecorder()
	encoder(resp, req, kratosErrors.New(http.StatusBadRequest, "INVALID_ARGUMENT", "stable message"))

	var body ResponseStructure
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &body))
	require.Equal(t, "本地化错误", body.ErrorMessage)
	require.Equal(t, "Accept-Language", resp.Header().Get("Vary"))
}

func TestNewErrorEncoderKeepsTraceIDAndUsesRequestMetadata(t *testing.T) {
	handler := NewDefaultErrorHandler()
	encoder := NewErrorEncoder(
		handler,
		WithErrorLogger(log.NewStdLogger(io.Discard)),
		WithErrorRequestMetadata(func(r *http.Request) map[string]string {
			return map[string]string{"device_id": r.Header.Get("X-Device-Id")}
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/homepagetab/v1/bootstrap", nil)
	req.Header.Set("X-Trace-Id", "trace-test-001")
	req.Header.Set("X-Device-Id", "device-test-001")
	resp := httptest.NewRecorder()

	encoder(resp, req, errors.New("sql: connection refused"))

	require.Equal(t, http.StatusInternalServerError, resp.Code)
	require.Equal(t, "trace-test-001", resp.Header().Get("X-Trace-Id"))
	var body ResponseStructure
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &body))
	require.False(t, body.Success)
	require.Equal(t, "UNKNOWN_ERROR", body.ErrorCode)
	require.Equal(t, "trace-test-001", body.TraceId)
	require.NotContains(t, body.ErrorMessage, "connection refused")
}
