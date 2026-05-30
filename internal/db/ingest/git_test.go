package ingest

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func setupTempGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "test"},
	}
	for _, args := range cmds {
		c := exec.Command(args[0], args[1:]...)
		c.Dir = dir
		if err := c.Run(); err != nil {
			t.Fatalf("git %v: %v", args[1:], err)
		}
	}

	// Create an initial commit so diff works
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("test"), 0644)
	c := exec.Command("git", "add", ".")
	c.Dir = dir
	c.Run()
	c = exec.Command("git", "commit", "-m", "init")
	c.Dir = dir
	c.Run()

	return dir
}

func TestHeadHash(t *testing.T) {
	dir := setupTempGitRepo(t)
	hash, err := HeadHash(dir)
	if err != nil {
		t.Fatalf("HeadHash() error: %v", err)
	}
	if len(hash) != 40 {
		t.Errorf("expected 40-char hash, got %d", len(hash))
	}
}

func TestDiff(t *testing.T) {
	dir := setupTempGitRepo(t)
	hash1, _ := HeadHash(dir)

	// Create a new file to create a diff
	os.WriteFile(filepath.Join(dir, "new.txt"), []byte("hello"), 0644)
	c := exec.Command("git", "add", "new.txt")
	c.Dir = dir
	c.Run()
	c = exec.Command("git", "commit", "-m", "add new.txt")
	c.Dir = dir
	c.Run()

	hash2, _ := HeadHash(dir)

	changes, err := Diff(dir, hash1, hash2)
	if err != nil {
		t.Fatalf("Diff() error: %v", err)
	}
	if len(changes) != 1 {
		t.Errorf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Status != "A" || changes[0].Path != "new.txt" {
		t.Errorf("expected A new.txt, got %s %s", changes[0].Status, changes[0].Path)
	}
}

func TestDiffEmpty(t *testing.T) {
	dir := setupTempGitRepo(t)
	hash, _ := HeadHash(dir)

	changes, err := Diff(dir, hash, hash)
	if err != nil {
		t.Fatalf("Diff() error: %v", err)
	}
	if len(changes) != 0 {
		t.Errorf("expected 0 changes for same hash, got %d", len(changes))
	}
}

func TestDiffUnavailableCommit(t *testing.T) {
	dir := setupTempGitRepo(t)

	// Use a fake commit hash that will never exist
	_, err := Diff(dir, "0000000000000000000000000000000000000000", "HEAD")
	if err == nil {
		t.Fatal("expected error for unavailable commit, got nil")
	}
}
