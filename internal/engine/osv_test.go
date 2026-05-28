package engine

import (
	"context"
	"errors"
	"testing"

	"malFuse/internal/osv"
)

type testOSVClient struct {
	result osv.Result
	err    error
	calls  []osvCall
}

type osvCall struct {
	name, ecosystem, version string
}

func (c *testOSVClient) Query(_ context.Context, name, ecosystem, version string) (osv.Result, error) {
	c.calls = append(c.calls, osvCall{name, ecosystem, version})
	return c.result, c.err
}

func TestOSVCheckVulnerable(t *testing.T) {
	client := &testOSVClient{
		result: osv.Result{Vulnerable: true},
	}
	check := OSVCheck(client, true)

	result := check(context.Background(), Request{Name: "bad-pkg", Ecosystem: "pypi"})
	if !result.Block {
		t.Error("expected Block true for vulnerable package")
	}
	if result.Reason != "osv" {
		t.Errorf("expected Reason osv, got %s", result.Reason)
	}
}

func TestOSVCheckNotVulnerable(t *testing.T) {
	client := &testOSVClient{
		result: osv.Result{Vulnerable: false},
	}
	check := OSVCheck(client, true)

	result := check(context.Background(), Request{Name: "safe-pkg", Ecosystem: "pypi"})
	if result.Block {
		t.Error("expected Block false for safe package")
	}
}

func TestOSVCheckFailOpen(t *testing.T) {
	client := &testOSVClient{
		err: errors.New("api unreachable"),
	}
	check := OSVCheck(client, true)

	result := check(context.Background(), Request{Name: "pkg", Ecosystem: "pypi"})
	if result.Block {
		t.Error("expected Block false on API error (fail-open)")
	}
}

func TestOSVCheckUsesEcosystem(t *testing.T) {
	client := &testOSVClient{
		result: osv.Result{Vulnerable: false},
	}
	check := OSVCheck(client, true)

	check(context.Background(), Request{Name: "pkg", Version: "1.0", Ecosystem: "npm"})

	if len(client.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(client.calls))
	}
	if client.calls[0].name != "pkg" {
		t.Errorf("expected name pkg, got %s", client.calls[0].name)
	}
	if client.calls[0].ecosystem != "npm" {
		t.Errorf("expected ecosystem npm, got %s", client.calls[0].ecosystem)
	}
	if client.calls[0].version != "1.0" {
		t.Errorf("expected version 1.0, got %s", client.calls[0].version)
	}
}
