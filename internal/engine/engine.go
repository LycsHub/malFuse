package engine

import (
	"context"

	"malFuse/internal/logger"
)

type Request struct {
	Name      string
	Version   string
	Ecosystem string
	RawPath   string
}

type Result struct {
	Block  bool
	Reason string
	Skip   bool
}

type CheckFunc func(ctx context.Context, req Request) Result

type namedCheck struct {
	name  string
	check CheckFunc
}

type Engine struct {
	checks []namedCheck
}

func New(checks ...CheckFunc) *Engine {
	e := &Engine{}
	for _, c := range checks {
		e.checks = append(e.checks, namedCheck{check: c})
	}
	return e
}

func (e *Engine) AddNamed(name string, check CheckFunc) {
	e.checks = append(e.checks, namedCheck{name: name, check: check})
}

func (e *Engine) Check(ctx context.Context, req Request) Result {
	for _, nc := range e.checks {
		select {
		case <-ctx.Done():
			return Result{Block: true, Reason: "context_cancelled"}
		default:
		}
		result := nc.check(ctx, req)

		name := nc.name
		if name == "" {
			name = "check"
		}
		if result.Block {
			logger.Debug("check result", "check", name, "result", "block", "reason", result.Reason)
		} else if result.Skip {
			logger.Debug("check result", "check", name, "result", "skip", "reason", result.Reason)
		} else {
			logger.Debug("check result", "check", name, "result", "pass")
		}

		if result.Block || result.Skip {
			return result
		}
	}
	return Result{Block: false}
}
