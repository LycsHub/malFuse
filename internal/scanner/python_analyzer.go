package scanner

import (
	"regexp"
	"strings"
)

var pyprojectBuildBackendPattern = regexp.MustCompile(`(?m)^build-backend\s*=\s*"([^"]+)"`)

func scanPthContent(content []byte, cfg ScanConfig) ScanResult {
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "import ") || strings.HasPrefix(line, "import\t") {
			if r := runDetectors([]byte(line), cfg); r.Block {
				return r
			}
		}
	}
	return ScanResult{Block: false}
}

func scanPyprojectToml(content []byte, cfg ScanConfig) ScanResult {
	matches := pyprojectBuildBackendPattern.FindAllSubmatch(content, -1)
	for _, m := range matches {
		if len(m) >= 2 {
			if r := runDetectors([]byte(m[1]), cfg); r.Block {
				return r
			}
		}
	}
	return ScanResult{Block: false}
}
