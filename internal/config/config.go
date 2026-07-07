package config

import (
	"encoding/json"
	"fmt"
	"time"
)

type Config struct {
	Port        string            `json:"port"`
	Host        string            `json:"host"`
	DBPath      string            `json:"db_path"`
	PIDFile     string            `json:"pid_file"`
	RepoProxy   string            `json:"repo_proxy"`
	Logging     LoggingConfig     `json:"logging"`
	Routing     []Route           `json:"routing"`
	Cooldown    CooldownConfig    `json:"cooldown"`
	Typo        TypoConfig        `json:"typo"`
	OSV         OSVConfig         `json:"osv"`
	ScriptScan  ScriptScanConfig  `json:"script_scan"`
	DynamicScan DynamicScanConfig `json:"dynamic_scan"`
}

type Route struct {
	Prefix    string `json:"prefix"`
	Upstream  string `json:"upstream"`
	Ecosystem string `json:"ecosystem"`
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
	Enabled     bool          `json:"enabled"`
	BlockOnVuln bool          `json:"block_on_vuln"`
	TTL         time.Duration `json:"ttl"`
	BaseURL     string        `json:"base_url"`
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
		Port:    "8080",
		Host:    "127.0.0.1",
		DBPath:  "malfuse.db",
		PIDFile: "malfuse.pid",
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		},
		OSV: OSVConfig{
			BaseURL: "https://api.osv.dev",
		},
		DynamicScan: defaultDynamicScanConfig(),
	}
}

func Load(data []byte) (*Config, error) {
	cfg := Default()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	cfg.applyDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) applyDefaults() {
	if c.DynamicScan.Runtime == "" {
		c.DynamicScan.Runtime = "msb"
	}
	if c.DynamicScan.Timeout == 0 {
		c.DynamicScan.Timeout = 30 * time.Second
	}
	if c.DynamicScan.MaxTotalSize == 0 {
		c.DynamicScan.MaxTotalSize = 50 * 1024 * 1024
	}
	if c.DynamicScan.Network == "" {
		c.DynamicScan.Network = "none"
	}
	if c.DynamicScan.NPMImage == "" {
		c.DynamicScan.NPMImage = "node:22-alpine"
	}
	if c.DynamicScan.PyPIImage == "" {
		c.DynamicScan.PyPIImage = "python:3.12-alpine"
	}
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
	return nil
}

type ScriptScanConfig struct {
	Enabled      bool              `json:"enabled"`
	MaxFileSize  int64             `json:"max_file_size"`
	MaxTotalSize int64             `json:"max_total_size"`
	Entropy      EntropyConfig     `json:"entropy"`
	Obfuscation  ObfuscationConfig `json:"obfuscation"`
	Network      NetworkConfig     `json:"network"`
}

type EntropyConfig struct {
	Enabled   bool    `json:"enabled"`
	Threshold float64 `json:"threshold"`
}

type ObfuscationConfig struct {
	Enabled         bool `json:"enabled"`
	Base64MinLength int  `json:"base64_min_length"`
	HexMinLength    int  `json:"hex_min_length"`
}

type NetworkConfig struct {
	Enabled         bool `json:"enabled"`
	AllowPrivateIPs bool `json:"allow_private_ips"`
}

type DynamicScanConfig struct {
	Enabled      bool          `json:"enabled"`
	Runtime      string        `json:"runtime"`
	Timeout      time.Duration `json:"timeout"`
	MaxTotalSize int64         `json:"max_total_size"`
	CacheEnabled bool          `json:"cache_enabled"`
	Network      string        `json:"network"`
	NPMImage     string        `json:"npm_image"`
	PyPIImage    string        `json:"pypi_image"`
}

func defaultDynamicScanConfig() DynamicScanConfig {
	return DynamicScanConfig{
		Runtime:      "msb",
		Timeout:      30 * time.Second,
		MaxTotalSize: 50 * 1024 * 1024,
		CacheEnabled: true,
		Network:      "none",
		NPMImage:     "node:22-alpine",
		PyPIImage:    "python:3.12-alpine",
	}
}

func (d *DynamicScanConfig) UnmarshalJSON(data []byte) error {
	type alias struct {
		Enabled      bool   `json:"enabled"`
		Runtime      string `json:"runtime"`
		Timeout      string `json:"timeout"`
		MaxTotalSize int64  `json:"max_total_size"`
		CacheEnabled *bool  `json:"cache_enabled"`
		Network      string `json:"network"`
		NPMImage     string `json:"npm_image"`
		PyPIImage    string `json:"pypi_image"`
	}

	defaults := defaultDynamicScanConfig()
	a := alias{CacheEnabled: &defaults.CacheEnabled}
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	d.Enabled = a.Enabled
	d.Runtime = a.Runtime
	d.MaxTotalSize = a.MaxTotalSize
	if a.CacheEnabled != nil {
		d.CacheEnabled = *a.CacheEnabled
	}
	d.Network = a.Network
	d.NPMImage = a.NPMImage
	d.PyPIImage = a.PyPIImage

	if a.Timeout != "" {
		timeout, err := time.ParseDuration(a.Timeout)
		if err != nil {
			return fmt.Errorf("invalid dynamic_scan timeout: %w", err)
		}
		d.Timeout = timeout
	}
	return nil
}

type LoggingConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
	Output string `json:"output"`
}
