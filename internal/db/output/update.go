package output

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"malFuse/internal/db/ingest"
	"malFuse/internal/db/schema"
	"malFuse/internal/logger"
)

type fileState struct {
	Ecosystems map[string]string `json:"ecosystems"`
}

func RunUpdate(db *sql.DB, repoDir, mode, sqlOutput, repoProxy string) error {
	repoURL := "https://github.com/ossf/malicious-packages"
	if repoProxy != "" {
		repoURL = fmt.Sprintf("https://%s/%s", strings.TrimRight(repoProxy, "/"), repoURL)
	}
	if err := ingest.Clone(repoURL, repoDir, 1); err != nil {
		return fmt.Errorf("clone repo: %w", err)
	}

	if err := ingest.Fetch(repoDir); err != nil {
		return fmt.Errorf("fetch repo: %w", err)
	}

	if err := ingest.CheckoutOSV(repoDir); err != nil {
		return fmt.Errorf("checkout osv: %w", err)
	}

	ecosystems, err := discoverEcosystems(repoDir)
	if err != nil {
		return fmt.Errorf("discover ecosystems: %w", err)
	}

	var allNew, allDeleted []ingest.ParsedPackage
	skippedEcosystems := 0

	for _, eco := range ecosystems {
		prevCommit := getPrevCommit(db, repoDir, eco)

		var addedFiles []string
		var deletedFiles []string

		remoteHash, err := ingest.RemoteHeadHash(repoDir)
		if err != nil {
			return fmt.Errorf("get remote head: %w", err)
		}

		if prevCommit == "" {
			ecoDir := filepath.Join(repoDir, "osv", "malicious", eco)
			files, err := ingest.ListFiles(ecoDir)
			if err != nil {
				return fmt.Errorf("list files for %s: %w", eco, err)
			}
			addedFiles = files
		} else {
			changes, err := ingest.Diff(repoDir, prevCommit, remoteHash)
			if err != nil {
				return fmt.Errorf("diff for %s: %w", eco, err)
			}

			ecoPrefix := fmt.Sprintf("osv/malicious/%s/", eco)
			for _, ch := range changes {
				if !strings.HasPrefix(ch.Path, ecoPrefix) {
					continue
				}
				if ch.Status == "D" {
					deletedFiles = append(deletedFiles, filepath.Join(repoDir, ch.Path))
				} else if ch.Status == "A" || ch.Status == "M" {
					addedFiles = append(addedFiles, filepath.Join(repoDir, ch.Path))
				}
			}
		}

		if len(addedFiles) == 0 && len(deletedFiles) == 0 {
			skippedEcosystems++
			continue
		}

		if prevCommit == "" {
			logger.Info("scanning ecosystem", "ecosystem", eco, "files", len(addedFiles))
		}

		newPkgs, _ := ingest.ParseFiles(addedFiles)

		var deletedPkgs []ingest.ParsedPackage
		for _, f := range deletedFiles {
			name, version := parseFromPath(repoDir, eco, f)
			if name != "" {
				deletedPkgs = append(deletedPkgs, ingest.ParsedPackage{
					Name:      name,
					Version:   version,
					Ecosystem: eco,
				})
			}
		}

		allNew = append(allNew, newPkgs...)
		allDeleted = append(allDeleted, deletedPkgs...)

		if mode == "sql" {
			// Accumulated below
		} else {
			if len(deletedPkgs) > 0 {
				if err := DeletePackages(db, deletedPkgs); err != nil {
					return err
				}
			}
			if len(newPkgs) > 0 {
				if err := UpsertPackages(db, newPkgs); err != nil {
					return err
				}
			}
			logger.Info("ecosystem updated", "ecosystem", eco,
				"added", len(newPkgs), "deleted", len(deletedPkgs))
		}

		savePrevCommit(db, repoDir, eco, remoteHash)
	}

	if mode == "sql" && (len(allNew) > 0 || len(allDeleted) > 0) {
		if err := WriteSQLFile(sqlOutput, allNew, allDeleted); err != nil {
			return err
		}
	}

	if skippedEcosystems > 0 {
		logger.Info("up to date", "ecosystems", skippedEcosystems)
	}
	logger.Info("update summary", "added", len(allNew), "deleted", len(allDeleted))

	return nil
}

func getPrevCommit(db *sql.DB, repoDir, eco string) string {
	if db != nil {
		commit, _, _ := schema.GetUpdateState(db, eco)
		if commit != "" {
			return commit
		}
	}

	// Fallback: read from state file for SQL mode
	stateFile := filepath.Join(repoDir, ".malfuse_state")
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return ""
	}
	var s fileState
	if err := json.Unmarshal(data, &s); err != nil {
		return ""
	}
	return s.Ecosystems[eco]
}

func savePrevCommit(db *sql.DB, repoDir, eco, commit string) {
	if db != nil {
		schema.SetUpdateState(db, eco, commit, time.Now().Format(time.RFC3339))
	}

	// Also save to state file
	stateFile := filepath.Join(repoDir, ".malfuse_state")
	s := fileState{Ecosystems: make(map[string]string)}
	data, _ := os.ReadFile(stateFile)
	json.Unmarshal(data, &s)
	s.Ecosystems[eco] = commit
	newData, _ := json.Marshal(s)
	os.WriteFile(stateFile, newData, 0644)
}

func parseFromPath(repoDir, ecosystem, path string) (string, string) {
	rel, _ := filepath.Rel(repoDir, path)
	parts := strings.Split(filepath.ToSlash(rel), "/")
	if len(parts) >= 5 {
		return parts[3], ""
	}
	return "", ""
}

func discoverEcosystems(repoDir string) ([]string, error) {
	maliciousDir := filepath.Join(repoDir, "osv", "malicious")
	entries, err := os.ReadDir(maliciousDir)
	if err != nil {
		return nil, fmt.Errorf("read malicious dir: %w", err)
	}
	var ecosystems []string
	for _, entry := range entries {
		if entry.IsDir() {
			ecosystems = append(ecosystems, entry.Name())
		}
	}
	return ecosystems, nil
}
