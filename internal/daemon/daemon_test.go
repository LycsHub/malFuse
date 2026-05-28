package daemon

import (
	"os"
	"strings"
	"syscall"
	"testing"
)

func TestWriteReadRemovePIDFile(t *testing.T) {
	path := t.TempDir() + "/test.pid"

	if err := WritePID(path, 12345); err != nil {
		t.Fatalf("WritePID: %v", err)
	}
	data, _ := os.ReadFile(path)
	if strings.TrimSpace(string(data)) != "12345" {
		t.Errorf("expected 12345, got %s", data)
	}

	pid, err := ReadPID(path)
	if err != nil {
		t.Fatalf("ReadPID: %v", err)
	}
	if pid != 12345 {
		t.Errorf("expected 12345, got %d", pid)
	}

	RemovePID(path)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("pid file should be removed")
	}
}

func TestReadPIDMissing(t *testing.T) {
	_, err := ReadPID("/nonexistent/test.pid")
	if err == nil {
		t.Error("expected error for missing pid file")
	}
}

func TestReadPIDInvalid(t *testing.T) {
	path := t.TempDir() + "/test.pid"
	os.WriteFile(path, []byte("not-a-number"), 0644)
	_, err := ReadPID(path)
	if err == nil {
		t.Error("expected error for invalid pid")
	}
}

func TestIsRunning(t *testing.T) {
	if !IsRunning(os.Getpid()) {
		t.Error("current process should be running")
	}
	if IsRunning(0) {
		t.Error("PID 0 should not be running")
	}
}

func TestSendSignal(t *testing.T) {
	// Send 0 signal to self — should succeed
	if err := SendSignal(os.Getpid(), syscall.Signal(0)); err != nil {
		t.Errorf("SendSignal to self: %v", err)
	}
}
