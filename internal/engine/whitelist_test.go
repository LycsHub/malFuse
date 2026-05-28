package engine

import (
	"context"
	"testing"

	"malFuse/internal/db/schema"
)

func TestWhitelistCheckSkipsMalicious(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, _ := schema.Open(dbPath)
	defer db.Close()

	schema.InsertWhitelist(db, "my-pkg", "pypi", "1.0")
	schema.InsertOrReplace(db, schema.MaliciousPackage{Name: "my-pkg", Ecosystem: "pypi", Version: "1.0"})

	whitelistCheck := WhitelistCheck(db)
	maliciousCheck := MaliciousDBCheck(db)

	// whitelist should skip before malicious check runs
	result := whitelistCheck(context.Background(), Request{Name: "my-pkg", Ecosystem: "pypi", Version: "1.0"})
	if !result.Skip {
		t.Error("expected Skip=true for whitelisted package")
	}
	if result.Reason != "whitelist" {
		t.Errorf("expected Reason whitelist, got %s", result.Reason)
	}

	// malicious check SHOULD block if run directly
	result = maliciousCheck(context.Background(), Request{Name: "my-pkg", Ecosystem: "pypi", Version: "1.0"})
	if !result.Block {
		t.Error("expected Block=true from malicious check for un-whitelisted call")
	}
}

func TestWhitelistCheckNotWhitelisted(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, _ := schema.Open(dbPath)
	defer db.Close()

	check := WhitelistCheck(db)
	result := check(context.Background(), Request{Name: "evil-pkg", Ecosystem: "pypi"})

	if result.Skip {
		t.Error("expected Skip=false for non-whitelisted package")
	}
	if result.Block {
		t.Error("expected Block=false for non-whitelisted package")
	}
}

func TestWhitelistCheckNilDB(t *testing.T) {
	check := WhitelistCheck(nil)
	result := check(context.Background(), Request{Name: "pkg", Ecosystem: "pypi"})
	if result.Skip || result.Block {
		t.Error("expected no Skip/Block with nil DB")
	}
}

func TestPipelineSkipStopsAtWhitelist(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, _ := schema.Open(dbPath)
	defer db.Close()

	schema.InsertWhitelist(db, "my-pkg", "pypi", "")

	called := false
	blockingCheck := CheckFunc(func(ctx context.Context, req Request) Result {
		called = true
		return Result{Block: true, Reason: "test"}
	})

	eng := New(WhitelistCheck(db), blockingCheck)
	result := eng.Check(context.Background(), Request{Name: "my-pkg", Ecosystem: "pypi"})

	if result.Block {
		t.Error("expected no Block when whitelisted")
	}
	if !result.Skip {
		t.Error("expected Skip=true from whitelist in pipeline")
	}
	if called {
		t.Error("blocking check should not have been called (skipped by whitelist)")
	}
}
