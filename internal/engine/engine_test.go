package engine

import (
	"context"
	"testing"
	"time"
)

func TestRequestAndResultTypes(t *testing.T) {
	req := Request{
		Name:      "test-pkg",
		Version:   "1.0.0",
		Ecosystem: "pypi",
		RawPath:   "/pypi/simple/test-pkg/",
	}
	if req.Name != "test-pkg" {
		t.Errorf("expected Name test-pkg, got %s", req.Name)
	}
	if req.Version != "1.0.0" {
		t.Errorf("expected Version 1.0.0, got %s", req.Version)
	}

	result := Result{Block: true, Reason: "blacklist"}
	if !result.Block {
		t.Error("expected Block true")
	}
	if result.Reason != "blacklist" {
		t.Errorf("expected Reason blacklist, got %s", result.Reason)
	}
}

func TestEngineCheckSequentialShortCircuit(t *testing.T) {
	callOrder := []string{}

	e := &Engine{
		checks: []CheckFunc{
			func(ctx context.Context, req Request) Result {
				callOrder = append(callOrder, "first")
				return Result{Block: true, Reason: "first_check"}
			},
			func(ctx context.Context, req Request) Result {
				callOrder = append(callOrder, "second")
				return Result{Block: true, Reason: "second_check"}
			},
		},
	}

	result := e.Check(context.Background(), Request{Name: "pkg"})
	if !result.Block {
		t.Error("expected Block true")
	}
	if result.Reason != "first_check" {
		t.Errorf("expected Reason first_check, got %s", result.Reason)
	}
	if len(callOrder) != 1 || callOrder[0] != "first" {
		t.Errorf("expected only first check called, got %v", callOrder)
	}
}

func TestEngineCheckAllPass(t *testing.T) {
	e := &Engine{
		checks: []CheckFunc{
			func(ctx context.Context, req Request) Result {
				return Result{Block: false}
			},
			func(ctx context.Context, req Request) Result {
				return Result{Block: false}
			},
		},
	}

	result := e.Check(context.Background(), Request{Name: "pkg"})
	if result.Block {
		t.Error("expected Block false when all checks pass")
	}
}

func TestEngineCheckContextCancellation(t *testing.T) {
	blocked := false
	e := &Engine{
		checks: []CheckFunc{
			func(ctx context.Context, req Request) Result {
				select {
				case <-ctx.Done():
					blocked = true
					return Result{Block: false}
				case <-time.After(100 * time.Millisecond):
					return Result{Block: false}
				}
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	result := e.Check(ctx, Request{Name: "pkg"})
	if result.Block {
		t.Error("expected Block false (no checks block, just ctx cancelled)")
	}
	// The check itself returned false, but ctx was done
	if !blocked {
		t.Error("expected check to detect context cancellation")
	}
}

func TestEngineCheckContextDeadlineExceeded(t *testing.T) {
	e := &Engine{
		checks: []CheckFunc{
			func(ctx context.Context, req Request) Result {
				select {
				case <-ctx.Done():
					return Result{Block: true, Reason: "context_cancelled"}
				case <-time.After(time.Second):
					return Result{Block: false}
				}
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	result := e.Check(ctx, Request{Name: "pkg"})
	if !result.Block {
		t.Error("expected Block true on context deadline")
	}
	if result.Reason != "context_cancelled" {
		t.Errorf("expected Reason context_cancelled, got %s", result.Reason)
	}
}
