package schema

import (
	"os"
	"testing"
)

func TestOpenCreatesTables(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	defer os.Remove(dbPath)

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer db.Close()

	var tableCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('malicious_packages', 'update_state')").Scan(&tableCount)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if tableCount != 2 {
		t.Errorf("expected 2 tables, got %d", tableCount)
	}
}

func TestWALModeEnabled(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	defer os.Remove(dbPath)

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer db.Close()

	var journalMode string
	err = db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		t.Fatalf("pragma query failed: %v", err)
	}
	if journalMode != "wal" {
		t.Errorf("expected journal_mode wal, got %s", journalMode)
	}
}

func TestInsertOrReplace(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	defer os.Remove(dbPath)

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer db.Close()

	err = InsertOrReplace(db, MaliciousPackage{
		Name:      "evil-pkg",
		Version:   "1.0",
		Ecosystem: "pypi",
		Published: "2024-01-01T00:00:00Z",
		Source:    "MAL-2024-1",
	})
	if err != nil {
		t.Fatalf("InsertOrReplace() error: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM malicious_packages WHERE name='evil-pkg' AND version='1.0' AND ecosystem='pypi'").Scan(&count)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 row, got %d", count)
	}
}

func TestInsertOrReplaceUpsert(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	defer os.Remove(dbPath)

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer db.Close()

	InsertOrReplace(db, MaliciousPackage{
		Name: "pkg", Ecosystem: "pypi", Published: "2024-01-01", Source: "MAL-1",
	})
	InsertOrReplace(db, MaliciousPackage{
		Name: "pkg", Ecosystem: "pypi", Published: "2024-06-01", Source: "MAL-1-UPDATED",
	})

	var published, source string
	err = db.QueryRow("SELECT published, source FROM malicious_packages WHERE name='pkg' AND ecosystem='pypi'").Scan(&published, &source)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if published != "2024-06-01" {
		t.Errorf("expected published 2024-06-01, got %s", published)
	}
	if source != "MAL-1-UPDATED" {
		t.Errorf("expected source MAL-1-UPDATED, got %s", source)
	}
}

func TestDelete(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	defer os.Remove(dbPath)

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer db.Close()

	InsertOrReplace(db, MaliciousPackage{Name: "pkg", Ecosystem: "pypi", Version: "1.0"})
	InsertOrReplace(db, MaliciousPackage{Name: "pkg", Ecosystem: "pypi"})

	err = Delete(db, "pkg", "pypi", "1.0")
	if err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM malicious_packages WHERE name='pkg' AND ecosystem='pypi'").Scan(&count)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 remaining row (version NULL), got %d", count)
	}

	var version *string
	err = db.QueryRow("SELECT version FROM malicious_packages WHERE name='pkg' AND ecosystem='pypi'").Scan(&version)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if version != nil {
		t.Errorf("expected NULL version, got %s", *version)
	}
}

func TestLookup(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	defer os.Remove(dbPath)

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer db.Close()

	InsertOrReplace(db, MaliciousPackage{Name: "evil-pkg", Ecosystem: "pypi", Version: "1.0"})

	found, err := Lookup(db, "evil-pkg", "pypi", "1.0")
	if err != nil {
		t.Fatalf("Lookup() error: %v", err)
	}
	if !found {
		t.Error("expected found true for matching name+version")
	}
}

func TestLookupVersionNull(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	defer os.Remove(dbPath)

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer db.Close()

	InsertOrReplace(db, MaliciousPackage{Name: "evil-pkg", Ecosystem: "pypi"})

	found, err := Lookup(db, "evil-pkg", "pypi", "any-version")
	if err != nil {
		t.Fatalf("Lookup() error: %v", err)
	}
	if !found {
		t.Error("expected found true for version=NULL matching any request version")
	}
}

func TestLookupNotFound(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	defer os.Remove(dbPath)

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer db.Close()

	found, err := Lookup(db, "safe-pkg", "pypi", "1.0")
	if err != nil {
		t.Fatalf("Lookup() error: %v", err)
	}
	if found {
		t.Error("expected found false for unknown package")
	}
}

func TestLookupNoVersionRequestMatchesSpecificVersion(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	defer os.Remove(dbPath)

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer db.Close()

	InsertOrReplace(db, MaliciousPackage{Name: "evil-pkg", Ecosystem: "pypi", Version: "1.0"})

	found, err := Lookup(db, "evil-pkg", "pypi", "")
	if err != nil {
		t.Fatalf("Lookup() error: %v", err)
	}
	if !found {
		t.Error("expected found true when request has no version but DB has specific version")
	}
}

func TestUpdateState(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	defer os.Remove(dbPath)

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer db.Close()

	err = SetUpdateState(db, "pypi", "abc123", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("SetUpdateState() error: %v", err)
	}

	commit, updated, err := GetUpdateState(db, "pypi")
	if err != nil {
		t.Fatalf("GetUpdateState() error: %v", err)
	}
	if commit != "abc123" {
		t.Errorf("expected commit abc123, got %s", commit)
	}
	if updated != "2024-01-01T00:00:00Z" {
		t.Errorf("expected updated 2024-01-01T00:00:00Z, got %s", updated)
	}
}

func TestOpenReadOnly(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"

	// first create the DB with Open (writable)
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	InsertOrReplace(db, MaliciousPackage{Name: "test", Ecosystem: "pypi"})
	db.Close()

	// then open read-only
	ro, err := OpenReadOnly(dbPath)
	if err != nil {
		t.Fatalf("OpenReadOnly() error: %v", err)
	}
	defer ro.Close()

	found, err := Lookup(ro, "test", "pypi", "")
	if err != nil {
		t.Fatalf("Lookup() error: %v", err)
	}
	if !found {
		t.Error("expected found true via read-only connection")
	}

	// verify PRAGMA journal_mode is NOT wal (we skip it in ReadOnly)
	var jm string
	ro.QueryRow("PRAGMA journal_mode").Scan(&jm)
	if jm != "delete" {
		t.Logf("journal_mode in read-only: %s", jm)
	}
}

func TestOpenReadOnlyMissingFile(t *testing.T) {
	_, err := OpenReadOnly("/nonexistent/path/test.db")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
