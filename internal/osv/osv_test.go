package osv

import (
	"testing"
	"time"
)

func TestCacheHit(t *testing.T) {
	cache := newCache(1 * time.Hour)

	cache.set("pkg:pypi", Result{Vulnerable: true})
	result, ok := cache.get("pkg:pypi")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if !result.Vulnerable {
		t.Error("expected Vulnerable true from cache")
	}
}

func TestCacheMiss(t *testing.T) {
	cache := newCache(1 * time.Hour)

	_, ok := cache.get("pkg:pypi")
	if ok {
		t.Fatal("expected cache miss")
	}
}

func TestCacheExpiry(t *testing.T) {
	cache := newCache(1 * time.Millisecond)

	cache.set("pkg:pypi", Result{Vulnerable: true})
	time.Sleep(2 * time.Millisecond)

	_, ok := cache.get("pkg:pypi")
	if ok {
		t.Fatal("expected cache miss after TTL expiry")
	}
}

func TestCacheSeparateEcosystems(t *testing.T) {
	cache := newCache(1 * time.Hour)

	cache.set("lodash:npm", Result{Vulnerable: true})
	cache.set("lodash:pypi", Result{Vulnerable: false})

	npmResult, _ := cache.get("lodash:npm")
	pypiResult, _ := cache.get("lodash:pypi")

	if !npmResult.Vulnerable {
		t.Error("expected npm result Vulnerable true")
	}
	if pypiResult.Vulnerable {
		t.Error("expected pypi result Vulnerable false")
	}
}

func TestCacheKey(t *testing.T) {
	key := cacheKey("requests", "pypi")
	if key != "requests:pypi" {
		t.Errorf("expected cache key 'requests:pypi', got %s", key)
	}
}
