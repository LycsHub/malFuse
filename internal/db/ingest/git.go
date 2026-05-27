package ingest

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type FileChange struct {
	Status string // A, M, D
	Path   string
}

func HeadHash(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func Diff(dir, fromHash, toHash string) ([]FileChange, error) {
	cmd := exec.Command("git", "diff", "--name-status", fromHash+".."+toHash)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff: %w", err)
	}

	var changes []FileChange
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		changes = append(changes, FileChange{
			Status: parts[0],
			Path:   parts[1],
		})
	}
	return changes, nil
}

func Clone(repoURL, dir string, depth int) error {
	if _, err := os.Stat(dir); err == nil {
		return nil // already exists
	}
	args := []string{"clone"}
	if depth > 0 {
		args = append(args, "--depth", fmt.Sprint(depth))
	}
	args = append(args, repoURL, dir)
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Fetch(dir string) error {
	cmd := exec.Command("git", "fetch", "origin", "main")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func RemoteHeadHash(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "origin/main")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse origin/main: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func CheckoutOSV(dir string) error {
	cmd := exec.Command("git", "checkout", "origin/main", "--", "osv/malicious/")
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
