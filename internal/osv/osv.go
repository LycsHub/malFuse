package osv

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Result struct {
	Vulnerable bool
}

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	cache      *cache
}

func NewClient(baseURL string, ttl time.Duration) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 2 * time.Second,
		},
		cache: newCache(ttl),
	}
}

type queryRequest struct {
	Package  packageRef `json:"package"`
	Version  string     `json:"version,omitempty"`
}

type packageRef struct {
	Name      string `json:"name"`
	Ecosystem string `json:"ecosystem"`
}

type queryResponse struct {
	Vulns []vuln `json:"vulns"`
}

type vuln struct {
	ID string `json:"id"`
}

func (c *Client) Query(ctx context.Context, name, ecosystem, version string) (Result, error) {
	key := cacheKey(name, ecosystem)
	if result, ok := c.cache.get(key); ok {
		return result, nil
	}

	result, err := c.queryAPI(ctx, name, ecosystem, version)
	if err != nil {
		return Result{Vulnerable: false}, err
	}

	c.cache.set(key, result)
	return result, nil
}

func (c *Client) queryAPI(ctx context.Context, name, ecosystem, version string) (Result, error) {
	reqBody := queryRequest{
		Package: packageRef{
			Name:      name,
			Ecosystem: ecosystem,
		},
		Version: version,
	}

	bodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return Result{}, fmt.Errorf("marshal request: %w", err)
	}

	// Use URL with query params: GET is official, POST also works
	queryURL, err := url.JoinPath(c.BaseURL, "/v1/query")
	if err != nil {
		return Result{}, fmt.Errorf("build URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, queryURL+"?package.name="+url.QueryEscape(name)+"&package.ecosystem="+url.QueryEscape(ecosystem), nil)
	if err != nil {
		return Result{}, fmt.Errorf("create request: %w", err)
	}
	_ = bodyJSON // keep for POST if needed

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("api request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("api returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{}, fmt.Errorf("read response: %w", err)
	}

	var qr queryResponse
	if err := json.Unmarshal(body, &qr); err != nil {
		return Result{}, fmt.Errorf("parse response: %w", err)
	}

	return Result{Vulnerable: len(qr.Vulns) > 0}, nil
}

type cacheEntry struct {
	result    Result
	expiresAt time.Time
}

type cache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	ttl     time.Duration
}

func newCache(ttl time.Duration) *cache {
	return &cache{
		entries: make(map[string]cacheEntry),
		ttl:     ttl,
	}
}

func (c *cache) get(key string) (Result, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[key]
	if !ok {
		return Result{}, false
	}
	if time.Now().After(entry.expiresAt) {
		return Result{}, false
	}
	return entry.result, true
}

func (c *cache) set(key string, result Result) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = cacheEntry{
		result:    result,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func cacheKey(name, ecosystem string) string {
	return fmt.Sprintf("%s:%s", name, ecosystem)
}
