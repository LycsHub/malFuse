package dynamic

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"malFuse/internal/engine"
	"malFuse/internal/logger"
)

const signalPrefix = "MALFUSE_DYNAMIC_SIGNAL:"

type Config struct {
	Runtime      string
	Timeout      time.Duration
	MaxTotalSize int64
	CacheEnabled bool
	Network      string
	NPMImage     string
	PyPIImage    string
}

type Runner interface {
	Run(ctx context.Context, req engine.Request, workspace string, cfg Config) (string, error)
}

type Analyzer struct {
	Config Config
	Runner Runner

	mu    sync.Mutex
	cache map[string]engine.DynamicResult
}

func New(cfg Config) *Analyzer {
	return &Analyzer{
		Config: cfg,
		Runner: MicrosandboxRunner{
			Binary: cfg.Runtime,
		},
		cache: make(map[string]engine.DynamicResult),
	}
}

func (a *Analyzer) Analyze(ctx context.Context, req engine.Request, archive []byte) engine.DynamicResult {
	if a.Runner == nil {
		a.Runner = MicrosandboxRunner{Binary: a.Config.Runtime}
	}
	if a.cache == nil {
		a.cache = make(map[string]engine.DynamicResult)
	}

	sum := sha256.Sum256(archive)
	key := hex.EncodeToString(sum[:])
	if a.Config.CacheEnabled {
		if result, ok := a.cacheGet(key); ok {
			return result
		}
	}

	if a.Config.MaxTotalSize > 0 && int64(len(archive)) > a.Config.MaxTotalSize {
		logger.Warn("dynamic scan skipped: archive too large",
			"package", req.Name,
			"ecosystem", req.Ecosystem,
			"size", len(archive),
		)
		result := engine.DynamicResult{Block: false}
		a.cacheSet(key, result)
		return result
	}

	workspace, err := os.MkdirTemp("", "malfuse-dynamic-*")
	if err != nil {
		logger.Warn("dynamic scan skipped: tempdir failed", "error", err)
		return engine.DynamicResult{Block: false}
	}
	defer os.RemoveAll(workspace)

	srcDir := filepath.Join(workspace, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		logger.Warn("dynamic scan skipped: mkdir failed", "error", err)
		return engine.DynamicResult{Block: false}
	}
	if err := unpackArchive(bytes.NewReader(archive), int64(len(archive)), srcDir); err != nil {
		logger.Warn("dynamic scan skipped: unpack failed", "error", err)
		return engine.DynamicResult{Block: false}
	}
	if err := writeGuards(filepath.Join(workspace, "guard", "bin")); err != nil {
		logger.Warn("dynamic scan skipped: guard setup failed", "error", err)
		return engine.DynamicResult{Block: false}
	}

	runCtx := ctx
	cancel := func() {}
	if a.Config.Timeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, a.Config.Timeout)
	}
	defer cancel()

	output, err := a.Runner.Run(runCtx, req, workspace, a.Config)
	result := classifyOutput(output)
	if result.Block {
		logger.Warn("dynamic scan blocked",
			"package", req.Name,
			"ecosystem", req.Ecosystem,
			"reason", result.Reason,
			"evidence", strings.Join(result.Evidence, " | "),
		)
		a.cacheSet(key, result)
		return result
	}
	if err != nil {
		logger.Warn("dynamic scan failed open",
			"package", req.Name,
			"ecosystem", req.Ecosystem,
			"error", err,
		)
	}

	result = engine.DynamicResult{Block: false}
	a.cacheSet(key, result)
	return result
}

func (a *Analyzer) cacheGet(key string) (engine.DynamicResult, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	result, ok := a.cache[key]
	return result, ok
}

func (a *Analyzer) cacheSet(key string, result engine.DynamicResult) {
	if !a.Config.CacheEnabled {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cache[key] = result
}

func classifyOutput(output string) engine.DynamicResult {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, signalPrefix) {
			continue
		}
		reason := strings.TrimPrefix(line, signalPrefix)
		if reason == "" {
			reason = "suspicious_process"
		}
		return engine.DynamicResult{
			Block:    true,
			Reason:   "dynamic:" + reason,
			Evidence: []string{line},
		}
	}
	return engine.DynamicResult{Block: false}
}

func unpackArchive(r io.Reader, size int64, dst string) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	if zr, err := zip.NewReader(bytes.NewReader(data), size); err == nil {
		return unpackZip(zr, dst)
	}
	if gr, err := gzip.NewReader(bytes.NewReader(data)); err == nil {
		defer gr.Close()
		return unpackTar(gr, dst)
	}
	return unpackTar(bytes.NewReader(data), dst)
}

func unpackZip(zr *zip.Reader, dst string) error {
	for _, f := range zr.File {
		path, ok := safeJoin(dst, f.Name)
		if !ok {
			continue
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(path, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		in, err := f.Open()
		if err != nil {
			return err
		}
		if err := writeFile(path, in, f.FileInfo().Mode()); err != nil {
			in.Close()
			return err
		}
		in.Close()
	}
	return nil
}

func unpackTar(r io.Reader, dst string) error {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		path, ok := safeJoin(dst, hdr.Name)
		if !ok {
			continue
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, 0o755); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			if err := writeFile(path, tr, os.FileMode(hdr.Mode)); err != nil {
				return err
			}
		}
	}
}

func safeJoin(base, name string) (string, bool) {
	clean := filepath.Clean(name)
	if clean == "." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || filepath.IsAbs(clean) {
		return "", false
	}
	path := filepath.Join(base, clean)
	rel, err := filepath.Rel(base, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return path, true
}

func writeFile(path string, r io.Reader, mode os.FileMode) error {
	out, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, r)
	return err
}

func writeGuards(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	wrappers := map[string]string{
		"curl": "network",
		"wget": "network",
		"nc":   "network",
		"ncat": "network",
		"cat":  "credential_access",
	}
	for name, reason := range wrappers {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(guardScript(reason)), 0o755); err != nil {
			return err
		}
	}
	return nil
}

func guardScript(reason string) string {
	if reason == "credential_access" {
		return fmt.Sprintf("#!/bin/sh\ncase \"$*\" in\n*id_rsa*|*.ssh*|*.npmrc*|*.pypirc*|*pip.conf*|*credentials*) echo %scredential_access; exit 42;;\nesac\nexec /bin/cat \"$@\"\n", signalPrefix)
	}
	return fmt.Sprintf("#!/bin/sh\ncase \"$*\" in\n*id_rsa*|*.ssh*|*.npmrc*|*.pypirc*|*pip.conf*|*credentials*) echo %scredential_access; exit 42;;\nesac\necho %s%s\nexit 42\n", signalPrefix, signalPrefix, reason)
}

type packageJSON struct {
	Scripts map[string]string `json:"scripts"`
}

func npmScriptsCommand() string {
	return `set -eu
export HOME=/workspace/home
export PATH=/workspace/guard/bin:$PATH
mkdir -p "$HOME"
pkg="$(find /workspace/src -name package.json -type f | head -n 1)"
if [ -z "$pkg" ]; then exit 0; fi
cd "$(dirname "$pkg")"
for script in preinstall install postinstall; do
  if node -e "const p=require('./package.json'); process.exit(p.scripts && p.scripts['$script'] ? 0 : 1)"; then
    npm run "$script" --ignore-scripts=false
  fi
done`
}

func pypiInstallCommand() string {
	return `set -eu
export HOME=/workspace/home
export PATH=/workspace/guard/bin:$PATH
mkdir -p "$HOME"
root="$(find /workspace/src -name setup.py -o -name pyproject.toml | head -n 1)"
if [ -z "$root" ]; then exit 0; fi
cd "$(dirname "$root")"
python -m pip install --no-index --no-deps --no-build-isolation .`
}

func scriptFor(req engine.Request) string {
	switch req.Ecosystem {
	case "npm":
		return npmScriptsCommand()
	case "pypi":
		return pypiInstallCommand()
	default:
		return "exit 0"
	}
}

func imageFor(req engine.Request, cfg Config) string {
	switch req.Ecosystem {
	case "npm":
		return cfg.NPMImage
	case "pypi":
		return cfg.PyPIImage
	default:
		return cfg.NPMImage
	}
}

func parsePackageJSONScripts(data []byte) map[string]string {
	pkg := packageJSON{}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil
	}
	return pkg.Scripts
}
