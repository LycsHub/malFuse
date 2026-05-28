package logger

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestInitTextStdout(t *testing.T) {
	Init(Config{Level: "info", Format: "text", Output: "stdout"})
	if logrusLogger == nil {
		t.Fatal("logger not initialized")
	}
}

func TestInitJSONFile(t *testing.T) {
	path := t.TempDir() + "/test.log"
	defer os.Remove(path)

	Init(Config{Level: "info", Format: "json", Output: path})

	Debug("test debug", "key", "value")
	Info("test info", "key", "value")
	Warn("test warn", "key", "value")
	Error("test error", "key", "value")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, `"level":"info"`) {
		t.Error("expected JSON log with level=info")
	}
	if !strings.Contains(content, "test info") {
		t.Error("expected info message in log")
	}
	if strings.Contains(content, "test debug") {
		t.Error("debug should be filtered out at info level")
	}
}

func TestLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	Init(Config{Level: "warn", Format: "text", Output: "stdout"})
	logrusLogger.SetOutput(&buf)

	Info("should not appear")
	Warn("should appear", "key", 1)

	out := buf.String()
	if strings.Contains(out, "should not appear") {
		t.Error("info log should be filtered at warn level")
	}
	if !strings.Contains(out, "should appear") {
		t.Error("warn log should appear at warn level")
	}
}

func TestFatal(t *testing.T) {
	// Fatal calls os.Exit — tested via integration/manual
	// Just ensure Init works and doesn't panic
	Init(Config{Level: "info", Format: "text", Output: "stdout"})
}

func TestResetForTests(t *testing.T) {
	Init(Config{Level: "info", Format: "text", Output: "stdout"})
	Reset()
	if logrusLogger != nil {
		t.Error("expected nil after Reset")
	}
}
