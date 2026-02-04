package logger

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
)

func TestZapLogger_JSON(t *testing.T) {
	tmpFile := "test_log.json"
	defer os.Remove(tmpFile)

	cfg := &Config{
		Level:    "debug",
		Format:   "json",
		Output:   "file",
		FilePath: tmpFile,
	}

	logger, _ := InitLogger(cfg, "test_id", "test_name", "v1.0.0")

	// Use Helper to simulate real usage
	h := log.NewHelper(logger)
	// Infow passes keyvals directly
	h.Infow("key1", "value1", "msg", "hello world")

	// Read file
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	t.Logf("Log content: %s", string(content))

	// Verify JSON
	var logEntry map[string]interface{}
	if err := json.Unmarshal(content, &logEntry); err != nil {
		t.Errorf("Log output is not valid JSON: %v", err)
	}

	// Verify fields
	if logEntry["level"] != "info" {
		t.Errorf("Expected level info, got %v", logEntry["level"])
	}
	if logEntry["msg"] != "hello world" {
		t.Errorf("Expected msg 'hello world', got %v", logEntry["msg"])
	}
	if logEntry["key1"] != "value1" {
		t.Errorf("Expected key1 'value1', got %v", logEntry["key1"])
	}
	// Verify Kratos fields
	if logEntry["service.id"] != "test_id" {
		t.Errorf("Expected service.id 'test_id', got %v", logEntry["service.id"])
	}
	if _, ok := logEntry["ts"]; !ok {
		t.Errorf("Expected ts field")
	}
	if _, ok := logEntry["caller"]; !ok {
		t.Errorf("Expected caller field")
	}
}

func TestZapLogger_Console(t *testing.T) {
    // Just ensure it doesn't panic
    cfg := &Config{
        Level:  "debug",
        Format: "text",
        Output: "stdout",
    }
    logger, _ := InitLogger(cfg, "test_id", "test_name", "v1.0.0")
    h := log.NewHelper(logger)
    h.Info("This is a console log test")
}
