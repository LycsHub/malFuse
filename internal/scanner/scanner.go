package scanner

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"log"
	"path/filepath"
	"strings"
)

type ScanConfig struct {
	MaxFileSize       int64
	MaxTotalSize      int64
	EntropyEnabled    bool
	EntropyThreshold  float64
	ObfuscationEnabled bool
	ObfuscationMinB64 int
	ObfuscationMinHex int
	NetworkEnabled    bool
	AllowPrivateIPs   bool
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
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("[WARN] scanner: tar read error: %v", err)
			return ScanResult{Block: false}
		}

		if !isInstallScript(hdr.Name) {
			continue
		}

		if hdr.Size > cfg.MaxFileSize {
			continue
		}

		content, err := io.ReadAll(io.LimitReader(tr, cfg.MaxFileSize))
		if err != nil {
			log.Printf("[WARN] scanner: read file %s: %v", hdr.Name, err)
			continue
		}

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
	}
	return ScanResult{Block: false}
}

func isInstallScript(name string) bool {
	base := strings.ToLower(filepath.Base(name))
	switch base {
	case "setup.py", "preinstall.js", "postinstall.js", "install.js",
		"preinstall.sh", "postinstall.sh", "install.sh":
		return true
	}
	return false
}
