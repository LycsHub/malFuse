package engine

import (
	"context"
	"errors"
	"testing"
	"time"
)

type testMetadataFetcher struct {
	time time.Time
	err  error
}

func (f *testMetadataFetcher) FetchPublishTime(_ context.Context, name, ecosystem string) (time.Time, error) {
	return f.time, f.err
}

func TestCooldownCheckRecentlyPublished(t *testing.T) {
	fetcher := &testMetadataFetcher{
		time: time.Now().Add(-1 * time.Hour),
	}
	check := CooldownCheck(fetcher, 48*time.Hour)

	result := check(context.Background(), Request{Name: "new-pkg", Ecosystem: "pypi"})
	if !result.Block {
		t.Error("expected Block true for recently published package (1h old, 48h cooldown)")
	}
	if result.Reason != "cooldown" {
		t.Errorf("expected Reason cooldown, got %s", result.Reason)
	}
}

func TestCooldownCheckOldPackage(t *testing.T) {
	fetcher := &testMetadataFetcher{
		time: time.Now().Add(-72 * time.Hour),
	}
	check := CooldownCheck(fetcher, 48*time.Hour)

	result := check(context.Background(), Request{Name: "old-pkg", Ecosystem: "pypi"})
	if result.Block {
		t.Error("expected Block false for old package (72h old, 48h cooldown)")
	}
}

func TestCooldownCheckFailClosedOnError(t *testing.T) {
	fetcher := &testMetadataFetcher{
		err: errors.New("metadata not found"),
	}
	check := CooldownCheck(fetcher, 48*time.Hour)

	result := check(context.Background(), Request{Name: "pkg", Ecosystem: "pypi"})
	if !result.Block {
		t.Error("expected Block true on metadata error (fail-closed)")
	}
	if result.Reason != "cooldown" {
		t.Errorf("expected Reason cooldown, got %s", result.Reason)
	}
}
