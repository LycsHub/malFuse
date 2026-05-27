package output

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverEcosystems(t *testing.T) {
	dir := t.TempDir()
	maliciousDir := filepath.Join(dir, "osv", "malicious")
	os.MkdirAll(filepath.Join(maliciousDir, "npm"), 0755)
	os.MkdirAll(filepath.Join(maliciousDir, "pypi"), 0755)
	os.MkdirAll(filepath.Join(maliciousDir, "go"), 0755)
	os.WriteFile(filepath.Join(maliciousDir, "README.md"), []byte("not a dir"), 0644)

	ecos, err := discoverEcosystems(dir)
	if err != nil {
		t.Fatalf("discoverEcosystems() error: %v", err)
	}
	if len(ecos) != 3 {
		t.Errorf("expected 3 ecosystems, got %d: %v", len(ecos), ecos)
	}
	found := make(map[string]bool)
	for _, e := range ecos {
		found[e] = true
	}
	for _, want := range []string{"npm", "pypi", "go"} {
		if !found[want] {
			t.Errorf("expected ecosystem %s not found", want)
		}
	}
}

func TestParseFromPath(t *testing.T) {
	repoDir := "/tmp/repo"
	path := filepath.Join(repoDir, "osv/malicious/pypi/requests/MAL-2024-1.json")
	name, version := parseFromPath(repoDir, "pypi", path)
	if name != "requests" {
		t.Errorf("expected name 'requests', got '%s'", name)
	}
	if version != "" {
		t.Errorf("expected empty version, got '%s'", version)
	}
}
