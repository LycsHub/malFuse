package config

import (
	"encoding/json"
	"fmt"
	"time"
)

type Config struct {
	Port      string          `json:"port"`
	Host      string          `json:"host"`
	Routing   []Route         `json:"routing"`
	Blacklist BlacklistConfig `json:"blacklist"`
	Cooldown  CooldownConfig  `json:"cooldown"`
	Typo      TypoConfig      `json:"typo"`
	OSV       OSVConfig       `json:"osv"`
}

type Route struct {
	Prefix    string `json:"prefix"`
	Upstream  string `json:"upstream"`
	Ecosystem string `json:"ecosystem"`
}

type BlacklistConfig struct {
	Entries []BlacklistEntry `json:"entries"`
}

type BlacklistEntry struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type CooldownConfig struct {
	Enabled  bool          `json:"enabled"`
	Duration time.Duration `json:"duration"`
}

func (c *CooldownConfig) UnmarshalJSON(data []byte) error {
	type alias struct {
		Enabled  bool   `json:"enabled"`
		Duration string `json:"duration"`
	}
	a := alias{}
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	c.Enabled = a.Enabled
	if a.Duration != "" {
		d, err := time.ParseDuration(a.Duration)
		if err != nil {
			return fmt.Errorf("invalid cooldown duration: %w", err)
		}
		c.Duration = d
	}
	return nil
}

type TypoConfig struct {
	Enabled   bool `json:"enabled"`
	Threshold int  `json:"threshold"`
}

type OSVConfig struct {
	Enabled bool          `json:"enabled"`
	TTL     time.Duration `json:"ttl"`
	BaseURL string        `json:"base_url"`
}

func (o *OSVConfig) UnmarshalJSON(data []byte) error {
	type alias struct {
		Enabled bool   `json:"enabled"`
		TTL     string `json:"ttl"`
		BaseURL string `json:"base_url"`
	}
	a := alias{}
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	o.Enabled = a.Enabled
	o.BaseURL = a.BaseURL
	if a.TTL != "" {
		d, err := time.ParseDuration(a.TTL)
		if err != nil {
			return fmt.Errorf("invalid osv ttl: %w", err)
		}
		o.TTL = d
	}
	return nil
}

func Default() *Config {
	return &Config{
		Port: "8080",
		Host: "127.0.0.1",
		OSV: OSVConfig{
			BaseURL: "https://api.osv.dev",
		},
	}
}

func Load(data []byte) (*Config, error) {
	cfg := Default()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	if len(c.Routing) == 0 {
		return fmt.Errorf("at least one route is required")
	}
	for i, r := range c.Routing {
		if r.Prefix == "" {
			return fmt.Errorf("route %d: prefix is required", i)
		}
		if r.Upstream == "" {
			return fmt.Errorf("route %d: upstream is required", i)
		}
		if r.Ecosystem == "" {
			return fmt.Errorf("route %d: ecosystem is required", i)
		}
	}
	for i, e := range c.Blacklist.Entries {
		if e.Name == "" {
			return fmt.Errorf("blacklist entry %d: name is required", i)
		}
	}
	return nil
}
