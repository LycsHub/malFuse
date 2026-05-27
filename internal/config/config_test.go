package config

import (
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	jsonData := `{
  "port": "9090",
  "host": "0.0.0.0",
  "db_path": "malfuse.db",
  "routing": [
    {"prefix": "/pypi/", "upstream": "https://pypi.org", "ecosystem": "pypi"},
    {"prefix": "/npm/", "upstream": "https://registry.npmjs.org", "ecosystem": "npm"}
  ],
  "cooldown": {
    "enabled": true,
    "duration": "24h"
  },
  "typo": {
    "enabled": true,
    "threshold": 3
  },
  "osv": {
    "enabled": true,
    "ttl": "30m",
    "base_url": "https://api.osv.dev"
  }
}`

	cfg, err := Load([]byte(jsonData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.Port)
	}
	if cfg.Host != "0.0.0.0" {
		t.Errorf("expected host 0.0.0.0, got %s", cfg.Host)
	}
	if cfg.DBPath != "malfuse.db" {
		t.Errorf("expected db_path malfuse.db, got %s", cfg.DBPath)
	}

	if len(cfg.Routing) != 2 {
		t.Fatalf("expected 2 routes, got %d", len(cfg.Routing))
	}

	if cfg.Cooldown.Enabled != true {
		t.Errorf("expected cooldown enabled true")
	}
	expectedCooldown := 24 * time.Hour
	if cfg.Cooldown.Duration != expectedCooldown {
		t.Errorf("expected cooldown duration %v, got %v", expectedCooldown, cfg.Cooldown.Duration)
	}

	if cfg.Typo.Enabled != true {
		t.Errorf("expected typo enabled true")
	}
	if cfg.Typo.Threshold != 3 {
		t.Errorf("expected typo threshold 3, got %d", cfg.Typo.Threshold)
	}

	if cfg.OSV.Enabled != true {
		t.Errorf("expected osv enabled true")
	}
	expectedTTL := 30 * time.Minute
	if cfg.OSV.TTL != expectedTTL {
		t.Errorf("expected osv ttl %v, got %v", expectedTTL, cfg.OSV.TTL)
	}
	if cfg.OSV.BaseURL != "https://api.osv.dev" {
		t.Errorf("expected osv base_url https://api.osv.dev, got %s", cfg.OSV.BaseURL)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	jsonData := `{
  "routing": [
    {"prefix": "/pypi/", "upstream": "https://pypi.org", "ecosystem": "pypi"}
  ]
}`

	cfg, err := Load([]byte(jsonData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("expected default port 8080, got %s", cfg.Port)
	}
	if cfg.Host != "127.0.0.1" {
		t.Errorf("expected default host 127.0.0.1, got %s", cfg.Host)
	}
	if cfg.DBPath != "malfuse.db" {
		t.Errorf("expected default db_path malfuse.db, got %s", cfg.DBPath)
	}
	if cfg.OSV.BaseURL != "https://api.osv.dev" {
		t.Errorf("expected default osv base_url, got %s", cfg.OSV.BaseURL)
	}
}
