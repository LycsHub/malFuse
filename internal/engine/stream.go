package engine

import (
	"context"
	"io"
)

type ScanResult struct {
	Block  bool
	Reason string
}

type StreamChecker interface {
	StreamCheck(req Request, body io.Reader) ScanResult
}

type DynamicResult struct {
	Block    bool
	Reason   string
	Evidence []string
}

type DynamicAnalyzer interface {
	Analyze(ctx context.Context, req Request, archive []byte) DynamicResult
}
