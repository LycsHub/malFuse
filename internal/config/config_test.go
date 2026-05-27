package config

import (
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	jsonData := `{
  "port": "9090",
  "host": "0.0.0.0",
  "routing": [
    {"prefix": "/pypi/", "upstream": "https://pypi.org", "ecosystem": "pypi"},
    {"prefix": "/npm/", "upstream": "https://registry.npmjs.org", "ecosystem": "npm"}
  ],
  "blacklist": {
    "entries": [
      {"name": "malicious-pkg"},
      {"name": "bad-lib", "version": "2.0.0"}
    ]
  },
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

	if len(cfg.Routing) != 2 {
		t.Fatalf("expected 2 routes, got %d", len(cfg.Routing))
	}
	if cfg.Routing[0].Prefix != "/pypi/" {
		t.Errorf("expected first route prefix /pypi/, got %s", cfg.Routing[0].Prefix)
	}
	if cfg.Routing[0].Upstream != "https://pypi.org" {
		t.Errorf("expected first route upstream https://pypi.org, got %s", cfg.Routing[0].Upstream)
	}
	if cfg.Routing[0].Ecosystem != "pypi" {
		t.Errorf("expected first route ecosystem pypi, got %s", cfg.Routing[0].Ecosystem)
	}

	if len(cfg.Blacklist.Entries) != 2 {
		t.Fatalf("expected 2 blacklist entries, got %d", len(cfg.Blacklist.Entries))
	}
	if cfg.Blacklist.Entries[0].Name != "malicious-pkg" {
		t.Errorf("expected first blacklist entry name malicious-pkg, got %s", cfg.Blacklist.Entries[0].Name)
	}
	if cfg.Blacklist.Entries[0].Version != "" {
		t.Errorf("expected first blacklist entry version empty, got %s", cfg.Blacklist.Entries[0].Version)
	}
	if cfg.Blacklist.Entries[1].Name != "bad-lib" {
		t.Errorf("expected second blacklist entry name bad-lib, got %s", cfg.Blacklist.Entries[1].Name)
	}
	if cfg.Blacklist.Entries[1].Version != "2.0.0" {
		t.Errorf("expected second blacklist entry version 2.0.0, got %s", cfg.Blacklist.Entries[1].Version)
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
	if cfg.OSV.BaseURL != "https://api.osv.dev" {
		t.Errorf("expected default osv base_url, got %s", cfg.OSV.BaseURL)
	}
}

func TestValidateBlacklist(t *testing.T) {
	tests := []struct {
		name    string
		entries []BlacklistEntry
		wantErr bool
	}{
		{"valid", []BlacklistEntry{{Name: "pkg"}}, false},
		{"empty name", []BlacklistEntry{{Name: ""}}, true},
		{"valid with version", []BlacklistEntry{{Name: "pkg", Version: "1.0"}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Routing:   []Route{{Prefix: "/pypi/", Upstream: "https://pypi.org", Ecosystem: "pypi"}},
				Blacklist: BlacklistConfig{Entries: tt.entries},
			}
			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
