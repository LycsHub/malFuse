package ingest

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type ParsedPackage struct {
	ID        string
	Name      string
	Version   string
	Ecosystem string
	Published string
}

type osvReport struct {
	ID        string         `json:"id"`
	Published string         `json:"published"`
	Affected  []osvAffected  `json:"affected"`
}

type osvAffected struct {
	Package osvPackage `json:"package"`
	Versions []string  `json:"versions"`
}

type osvPackage struct {
	Ecosystem string `json:"ecosystem"`
	Name      string `json:"name"`
}

func ParseOSV(data []byte) ([]ParsedPackage, error) {
	var report osvReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("parse OSV JSON: %w", err)
	}

	var results []ParsedPackage
	for _, affected := range report.Affected {
		ecosystem := normalizeEcosystem(affected.Package.Ecosystem)
		if affected.Package.Name == "" || ecosystem == "" {
			continue
		}
		if len(affected.Versions) == 0 {
			results = append(results, ParsedPackage{
				ID:        report.ID,
				Name:      affected.Package.Name,
				Ecosystem: ecosystem,
				Published: report.Published,
			})
		} else {
			for _, v := range affected.Versions {
				results = append(results, ParsedPackage{
					ID:        report.ID,
					Name:      affected.Package.Name,
					Version:   v,
					Ecosystem: ecosystem,
					Published: report.Published,
				})
			}
		}
	}
	return results, nil
}

func normalizeEcosystem(ecosystem string) string {
	switch ecosystem {
	case "PyPI":
		return "pypi"
	case "VSCode:https://open-vsx.org":
		return "vscode:open-vsx.org"
	default:
		return strings.ToLower(ecosystem)
	}
}

func ParseFile(path string) ([]ParsedPackage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", path, err)
	}
	return ParseOSV(data)
}

func ListFiles(baseDir string) ([]string, error) {
	var files []string
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".json") {
			return nil
		}

		relPath, _ := filepath.Rel(baseDir, path)
		if strings.HasPrefix(relPath, "withdrawn/") || strings.HasPrefix(relPath, "unmergable/") {
			return nil
		}

		files = append(files, path)
		return nil
	})
	return files, err
}

func ParseFiles(paths []string) ([]ParsedPackage, error) {
	var all []ParsedPackage
	for _, path := range paths {
		pkgs, err := ParseFile(path)
		if err != nil {
			log.Printf("[WARN] failed to parse %s: %v", path, err)
			continue
		}
		all = append(all, pkgs...)
	}
	return all, nil
}
