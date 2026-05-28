package scanner

import (
	"archive/tar"
	"compress/gzip"
	"io"

	"malFuse/internal/logger"
)

type ScanConfig struct {
	MaxFileSize        int64
	MaxTotalSize       int64
	EntropyEnabled     bool
	EntropyThreshold   float64
	ObfuscationEnabled bool
	ObfuscationMinB64  int
	ObfuscationMinHex  int
	NetworkEnabled     bool
	AllowPrivateIPs    bool
}

func Scan(r io.Reader, pkgName string, cfg ScanConfig) ScanResult {
	lr := io.LimitReader(r, cfg.MaxTotalSize)

	gr, err := gzip.NewReader(lr)
	if err == nil {
		defer gr.Close()
		return scanTar(gr, cfg)
	}

	return scanTar(r, cfg)
}

func scanTar(r io.Reader, cfg ScanConfig) ScanResult {
	tr := tar.NewReader(r)

	// Determine ecosystem by scanning files
	hasPackageJSON := false

	// Single-pass: buffer all readable files
	type fileEntry struct {
		name    string
		content []byte
	}
	var files []fileEntry
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Warn("scanner: tar read error", "error", err)
			return ScanResult{Block: false}
		}
		if hdr.Size > cfg.MaxFileSize {
			continue
		}
		content, err := io.ReadAll(io.LimitReader(tr, cfg.MaxFileSize))
		if err != nil {
			continue
		}
		files = append(files, fileEntry{name: hdr.Name, content: content})
	}

	// Detect ecosystem
	for _, f := range files {
		base := filename(f.name)
		if base == "package.json" || hasExt(base, ".js") {
			hasPackageJSON = true
			break
		}
	}

	if hasPackageJSON {
		foundPkgJSON := false
		for _, f := range files {
			if filename(f.name) == "package.json" {
				foundPkgJSON = true
				scripts, err := parsePackageJSONScripts(f.content)
				if err != nil {
					continue
				}
				for _, cmd := range scripts {
					if r := runDetectors([]byte(cmd), cfg); r.Block {
						return r
					}
				}
				refs := extractReferencedJSFiles(scripts)
				for _, ref := range refs {
					for _, f2 := range files {
						if filename(f2.name) == filename(ref) {
							if r := runDetectors(f2.content, cfg); r.Block {
								return r
							}
						}
					}
				}
			}
		}
		// If no package.json found but .js files exist, scan them directly
		if !foundPkgJSON {
			for _, f := range files {
				if hasExt(filename(f.name), ".js") {
					if r := runDetectors(f.content, cfg); r.Block {
						return r
					}
				}
			}
		}
		return ScanResult{Block: false}
	}

	// Python ecosystem: scan known entry points
	for _, f := range files {
		base := filename(f.name)
		switch {
		case base == "setup.py" || base == "__init__.py":
			if r := runDetectors(f.content, cfg); r.Block {
				return r
			}
		case hasExt(base, ".pth"):
			if r := scanPthContent(f.content, cfg); r.Block {
				return r
			}
		case base == "pyproject.toml":
			if r := scanPyprojectToml(f.content, cfg); r.Block {
				return r
			}
		}
	}

	return ScanResult{Block: false}
}

func runDetectors(content []byte, cfg ScanConfig) ScanResult {
	if cfg.EntropyEnabled {
		if result := entropyCheck(content, cfg.EntropyThreshold); result.Block {
			return result
		}
	}
	if cfg.ObfuscationEnabled {
		if result := obfuscationCheck(content, cfg.ObfuscationMinB64); result.Block {
			return result
		}
	}
	if cfg.NetworkEnabled {
		if result := networkCheck(content, cfg.AllowPrivateIPs); result.Block {
			return result
		}
	}
	return ScanResult{Block: false}
}

func filename(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}

func hasExt(name, ext string) bool {
	return len(name) > len(ext) && name[len(name)-len(ext):] == ext
}
