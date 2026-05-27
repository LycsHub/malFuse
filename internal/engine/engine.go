package engine

import "context"

type Request struct {
	Name      string
	Version   string
	Ecosystem string
	RawPath   string
}

type Result struct {
	Block  bool
	Reason string
}

type CheckFunc func(ctx context.Context, req Request) Result

type Engine struct {
	checks []CheckFunc
}

func New(checks ...CheckFunc) *Engine {
	return &Engine{checks: checks}
}

func (e *Engine) Check(ctx context.Context, req Request) Result {
	for _, check := range e.checks {
		select {
		case <-ctx.Done():
			return Result{Block: true, Reason: "context_cancelled"}
		default:
		}
		result := check(ctx, req)
		if result.Block {
			return result
		}
	}
	return Result{Block: false}
}
