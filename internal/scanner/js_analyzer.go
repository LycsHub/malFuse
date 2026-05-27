package scanner

import (
	"archive/tar"
	"encoding/json"
	"io"
	"path/filepath"
	"regexp"
	"strings"
)

func analyzeJSArchive(tr *tar.Reader, cfg ScanConfig) ScanResult {
	var pkgJSON []byte
	var jsFiles map[string][]byte = make(map[string][]byte)
	var fileList []string

	// First pass: collect package.json and all JS files
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return ScanResult{Block: false}
		}
		fileList = append(fileList, hdr.Name)

		if hdr.Size > cfg.MaxFileSize {
			continue
		}

		content, _ := io.ReadAll(io.LimitReader(tr, cfg.MaxFileSize))
		base := filepath.Base(hdr.Name)

		if base == "package.json" {
			pkgJSON = content
		} else if strings.HasSuffix(base, ".js") {
			jsFiles[filepath.Base(hdr.Name)] = content
		}
	}

	// Analyze package.json scripts
	if pkgJSON != nil {
		scripts, err := parsePackageJSONScripts(pkgJSON)
		if err == nil {
			for _, cmd := range scripts {
				if r := runDetectors([]byte(cmd), cfg); r.Block {
					return r
				}
			}

			// Collect referenced JS files for second pass
			refs := extractReferencedJSFiles(scripts)
			for _, ref := range refs {
				base := filepath.Base(ref)
				if content, ok := jsFiles[base]; ok {
					if r := runDetectors(content, cfg); r.Block {
						return r
					}
				}
			}
		}
	}

	return ScanResult{Block: false}
}

var nodeJSRefPattern = regexp.MustCompile(`node\s+([^\s&|;]+\.js)`)

func parsePackageJSONScripts(data []byte) (map[string]string, error) {
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}
	return pkg.Scripts, nil
}

func extractReferencedJSFiles(scripts map[string]string) []string {
	seen := make(map[string]bool)
	var refs []string
	for _, cmd := range scripts {
		matches := nodeJSRefPattern.FindAllStringSubmatch(cmd, -1)
		for _, m := range matches {
			name := strings.TrimPrefix(m[1], "./")
			if !seen[name] {
				seen[name] = true
				refs = append(refs, name)
			}
		}
	}
	return refs
}
