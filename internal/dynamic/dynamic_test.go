package dynamic

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"path/filepath"
	"testing"
	"time"

	"malFuse/internal/engine"
)

type fakeRunner struct {
	calls     int
	output    string
	workspace string
}

func (r *fakeRunner) Run(_ context.Context, _ engine.Request, workspace string, _ Config) (string, error) {
	r.calls++
	r.workspace = workspace
	return r.output, nil
}

func TestClassifyOutputBlocksSignal(t *testing.T) {
	result := classifyOutput("ok\nMALFUSE_DYNAMIC_SIGNAL:network\n")
	if !result.Block {
		t.Fatal("expected block")
	}
	if result.Reason != "dynamic:network" {
		t.Errorf("expected dynamic:network, got %s", result.Reason)
	}
}

func TestAnalyzerCachesByArchiveHash(t *testing.T) {
	archive := makeTarGz(t, map[string]string{
		"pkg/package.json": `{"scripts":{"postinstall":"curl http://example.test"}}`,
	})
	runner := &fakeRunner{output: "MALFUSE_DYNAMIC_SIGNAL:network"}
	analyzer := &Analyzer{
		Config: Config{
			Runtime:      "msb",
			Timeout:      time.Second,
			MaxTotalSize: 1024 * 1024,
			CacheEnabled: true,
			Network:      "none",
			NPMImage:     "node:22-alpine",
			PyPIImage:    "python:3.12-alpine",
		},
		Runner: runner,
		cache:  make(map[string]engine.DynamicResult),
	}

	req := engine.Request{Name: "pkg", Ecosystem: "npm"}
	first := analyzer.Analyze(context.Background(), req, archive)
	second := analyzer.Analyze(context.Background(), req, archive)

	if !first.Block || first.Reason != "dynamic:network" {
		t.Fatalf("expected first result dynamic network block, got %+v", first)
	}
	if !second.Block || second.Reason != "dynamic:network" {
		t.Fatalf("expected cached result dynamic network block, got %+v", second)
	}
	if runner.calls != 1 {
		t.Errorf("expected runner called once, got %d", runner.calls)
	}
	if runner.workspace == "" {
		t.Error("expected runner workspace")
	}
}

func TestAnalyzerPassesCleanOutput(t *testing.T) {
	archive := makeTarGz(t, map[string]string{
		"pkg/package.json": `{"scripts":{"postinstall":"echo ok"}}`,
	})
	runner := &fakeRunner{}
	analyzer := &Analyzer{
		Config: Config{
			Runtime:      "msb",
			Timeout:      time.Second,
			MaxTotalSize: 1024 * 1024,
			CacheEnabled: true,
			Network:      "none",
			NPMImage:     "node:22-alpine",
			PyPIImage:    "python:3.12-alpine",
		},
		Runner: runner,
		cache:  make(map[string]engine.DynamicResult),
	}

	result := analyzer.Analyze(context.Background(), engine.Request{Name: "pkg", Ecosystem: "npm"}, archive)
	if result.Block {
		t.Fatalf("expected clean output pass, got %+v", result)
	}
}

func makeTarGz(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	for name, content := range files {
		data := []byte(content)
		hdr := &tar.Header{
			Name: filepath.ToSlash(name),
			Mode: 0o644,
			Size: int64(len(data)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("WriteHeader: %v", err)
		}
		if _, err := io.Copy(tw, bytes.NewReader(data)); err != nil {
			t.Fatalf("copy: %v", err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tar close: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}
	return buf.Bytes()
}
