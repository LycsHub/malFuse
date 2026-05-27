package scanner

import (
	"io"

	"malFuse/internal/engine"
)

type StreamChecker struct {
	Config ScanConfig
}

func (s *StreamChecker) StreamCheck(req engine.Request, body io.Reader) engine.ScanResult {
	result := Scan(body, req.Name, s.Config)
	return engine.ScanResult{Block: result.Block, Reason: result.Reason}
}
