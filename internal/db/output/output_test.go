package output

import (
	"os"
	"strings"
	"testing"

	"malFuse/internal/db/ingest"
	"malFuse/internal/db/schema"
)

func TestDirectModeUpsert(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	defer os.Remove(dbPath)

	db, err := schema.Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer db.Close()

	pkgs := []ingest.ParsedPackage{
		{ID: "MAL-1", Name: "evil-pkg", Ecosystem: "pypi", Version: "1.0", Published: "2024-01-01"},
		{ID: "MAL-2", Name: "bad-lib", Ecosystem: "pypi", Published: "2024-01-01"},
	}

	err = UpsertPackages(db, pkgs)
	if err != nil {
		t.Fatalf("UpsertPackages() error: %v", err)
	}

	found, _ := schema.Lookup(db, "evil-pkg", "pypi", "1.0")
	if !found {
		t.Error("expected evil-pkg to be found")
	}
	found, _ = schema.Lookup(db, "bad-lib", "pypi", "any")
	if !found {
		t.Error("expected bad-lib (no version) to match any version")
	}
}

func TestDirectModeDelete(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	defer os.Remove(dbPath)

	db, err := schema.Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer db.Close()

	schema.InsertOrReplace(db, schema.MaliciousPackage{Name: "evil-pkg", Ecosystem: "pypi", Version: "1.0"})

	err = DeletePackages(db, []ingest.ParsedPackage{
		{Name: "evil-pkg", Ecosystem: "pypi", Version: "1.0"},
	})
	if err != nil {
		t.Fatalf("DeletePackages() error: %v", err)
	}

	found, _ := schema.Lookup(db, "evil-pkg", "pypi", "1.0")
	if found {
		t.Error("expected evil-pkg to be deleted")
	}
}

func TestGenerateSQL(t *testing.T) {
	pkgs := []ingest.ParsedPackage{
		{ID: "MAL-1", Name: "evil-pkg", Ecosystem: "pypi", Version: "1.0", Published: "2024-01-01"},
		{ID: "MAL-2", Name: "bad-lib", Ecosystem: "npm", Published: "2024-02-01"},
	}

	sql := GenerateInsertSQL(pkgs)
	if !strings.Contains(sql, "INSERT OR REPLACE INTO malicious_packages") {
		t.Error("expected INSERT statement")
	}
	if !strings.Contains(sql, "'evil-pkg'") {
		t.Error("expected evil-pkg in SQL")
	}
	if !strings.Contains(sql, "'bad-lib'") {
		t.Error("expected bad-lib in SQL")
	}
}

func TestGenerateDeleteSQL(t *testing.T) {
	pkgs := []ingest.ParsedPackage{
		{Name: "evil-pkg", Ecosystem: "pypi", Version: "1.0"},
		{Name: "bad-lib", Ecosystem: "npm"},
	}

	sql := GenerateDeleteSQL(pkgs)
	if !strings.Contains(sql, "DELETE FROM malicious_packages") {
		t.Error("expected DELETE statement")
	}
	if !strings.Contains(sql, "'evil-pkg'") {
		t.Error("expected evil-pkg in delete SQL")
	}
}

func TestWriteSQLFile(t *testing.T) {
	outPath := t.TempDir() + "/updates.sql"
	pkgs := []ingest.ParsedPackage{
		{ID: "MAL-1", Name: "pkg", Ecosystem: "pypi", Version: "1.0", Published: "2024-01-01"},
	}

	err := WriteSQLFile(outPath, pkgs, nil)
	if err != nil {
		t.Fatalf("WriteSQLFile() error: %v", err)
	}

	data, _ := os.ReadFile(outPath)
	if len(data) == 0 {
		t.Error("expected non-empty SQL file")
	}
}
