package engine

import "io"

type ScanResult struct {
	Block  bool
	Reason string
}

type StreamChecker interface {
	StreamCheck(req Request, body io.Reader) ScanResult
}
