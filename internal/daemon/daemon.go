package daemon

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func WritePID(path string, pid int) error {
	return os.WriteFile(path, []byte(strconv.Itoa(pid)+"\n"), 0644)
}

func ReadPID(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("read pid file: %w", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("parse pid: %w", err)
	}
	return pid, nil
}

func RemovePID(path string) {
	os.Remove(path)
}

func IsRunning(pid int) bool {
	return SendSignal(pid, syscall.Signal(0)) == nil
}

func SendSignal(pid int, sig os.Signal) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Signal(sig)
}
