package engine

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"malFuse/internal/db/schema"
)

func TestMaliciousDBCheckFound(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	defer os.Remove(dbPath)
	db, err := schema.Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer db.Close()

	schema.InsertOrReplace(db, schema.MaliciousPackage{Name: "evil-pkg", Ecosystem: "pypi", Version: "1.0"})

	check := MaliciousDBCheck(db)
	result := check(context.Background(), Request{Name: "evil-pkg", Ecosystem: "pypi", Version: "1.0"})

	if !result.Block {
		t.Error("expected Block true for known malicious package")
	}
	if result.Reason != "malicious-db" {
		t.Errorf("expected Reason malicious-db, got %s", result.Reason)
	}
}

func TestMaliciousDBCheckNotFound(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	defer os.Remove(dbPath)
	db, err := schema.Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer db.Close()

	check := MaliciousDBCheck(db)
	result := check(context.Background(), Request{Name: "safe-pkg", Ecosystem: "pypi"})

	if result.Block {
		t.Error("expected Block false for unknown package")
	}
}

func TestMaliciousDBCheckNilDB(t *testing.T) {
	check := MaliciousDBCheck(nil)
	result := check(context.Background(), Request{Name: "evil-pkg", Ecosystem: "pypi"})

	if result.Block {
		t.Error("expected Block false when db is nil (graceful fallback)")
	}
}

func TestMaliciousDBCheckClosedDB(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	defer os.Remove(dbPath)
	db, _ := sql.Open("sqlite", dbPath)
	db.Close()

	check := MaliciousDBCheck(db)
	result := check(context.Background(), Request{Name: "evil-pkg", Ecosystem: "pypi"})

	if result.Block {
		t.Error("expected Block false when db is closed (graceful fallback)")
	}
}
