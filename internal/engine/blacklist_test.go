package engine

import (
	"context"
	"testing"
)

func TestBlacklistNameOnlyMatch(t *testing.T) {
	check := BlacklistCheck([]BlacklistEntry{
		{Name: "malicious-pkg"},
	})

	result := check(context.Background(), Request{Name: "malicious-pkg"})
	if !result.Block {
		t.Error("expected Block true for blacklisted name")
	}
	if result.Reason != "blacklist" {
		t.Errorf("expected Reason blacklist, got %s", result.Reason)
	}
}

func TestBlacklistNameOnlyNoMatch(t *testing.T) {
	check := BlacklistCheck([]BlacklistEntry{
		{Name: "malicious-pkg"},
	})

	result := check(context.Background(), Request{Name: "safe-pkg"})
	if result.Block {
		t.Error("expected Block false for non-blacklisted name")
	}
}

func TestBlacklistNameAndVersionMatch(t *testing.T) {
	check := BlacklistCheck([]BlacklistEntry{
		{Name: "bad-lib", Version: "2.0.0"},
	})

	result := check(context.Background(), Request{Name: "bad-lib", Version: "2.0.0"})
	if !result.Block {
		t.Error("expected Block true for matching name and version")
	}
}

func TestBlacklistNameMatchVersionMismatch(t *testing.T) {
	check := BlacklistCheck([]BlacklistEntry{
		{Name: "bad-lib", Version: "2.0.0"},
	})

	result := check(context.Background(), Request{Name: "bad-lib", Version: "1.5.0"})
	if result.Block {
		t.Error("expected Block false when version does not match")
	}
}

func TestBlacklistNameAndVersionWhenNoVersionInRequest(t *testing.T) {
	check := BlacklistCheck([]BlacklistEntry{
		{Name: "bad-lib", Version: "2.0.0"},
	})

	result := check(context.Background(), Request{Name: "bad-lib"})
	if result.Block {
		t.Error("expected Block false when request has no version but entry has version constraint")
	}
}
