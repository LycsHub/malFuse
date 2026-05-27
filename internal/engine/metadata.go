package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type RegistryMetadataFetcher struct {
	HTTPClient *http.Client
	Routes     []RouteConfig
}

type RouteConfig struct {
	Prefix    string
	Upstream  string
	Ecosystem string
}

func NewRegistryMetadataFetcher(routes []RouteConfig) *RegistryMetadataFetcher {
	return &RegistryMetadataFetcher{
		HTTPClient: &http.Client{Timeout: 2 * time.Second},
		Routes:     routes,
	}
}

func (f *RegistryMetadataFetcher) FetchPublishTime(ctx context.Context, name, ecosystem string) (time.Time, error) {
	upstream, ok := f.upstreamFor(ecosystem)
	if !ok {
		return time.Time{}, fmt.Errorf("no upstream configured for ecosystem %s", ecosystem)
	}

	metadataURL, err := f.buildMetadataURL(upstream, ecosystem, name)
	if err != nil {
		return time.Time{}, fmt.Errorf("build metadata URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, metadataURL, nil)
	if err != nil {
		return time.Time{}, fmt.Errorf("create request: %w", err)
	}

	resp, err := f.HTTPClient.Do(req)
	if err != nil {
		return time.Time{}, fmt.Errorf("fetch metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return time.Time{}, fmt.Errorf("metadata fetch returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return time.Time{}, fmt.Errorf("read metadata: %w", err)
	}

	return f.parsePublishTime(ecosystem, name, body)
}

func (f *RegistryMetadataFetcher) upstreamFor(ecosystem string) (string, bool) {
	for _, r := range f.Routes {
		if r.Ecosystem == ecosystem {
			return r.Upstream, true
		}
	}
	return "", false
}

func (f *RegistryMetadataFetcher) buildMetadataURL(upstream, ecosystem, name string) (string, error) {
	base := upstream
	if base[len(base)-1] == '/' {
		base = base[:len(base)-1]
	}
	switch ecosystem {
	case "pypi":
		return fmt.Sprintf("%s/pypi/%s/json", base, name), nil
	case "npm":
		return fmt.Sprintf("%s/%s", base, name), nil
	default:
		return "", fmt.Errorf("unsupported ecosystem %s", ecosystem)
	}
}

func (f *RegistryMetadataFetcher) parsePublishTime(ecosystem, name string, body []byte) (time.Time, error) {
	switch ecosystem {
	case "pypi":
		return f.parsePyPITime(body, name)
	case "npm":
		return f.parseNPMTime(body, name)
	default:
		return time.Time{}, fmt.Errorf("unsupported ecosystem %s", ecosystem)
	}
}

type pypiResponse struct {
	Info struct {
		Name       string `json:"name"`
		UploadTime string `json:"upload_time"`
	} `json:"info"`
}

func (f *RegistryMetadataFetcher) parsePyPITime(body []byte, name string) (time.Time, error) {
	var resp pypiResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return time.Time{}, fmt.Errorf("parse PyPI JSON: %w", err)
	}
	if resp.Info.UploadTime == "" {
		return time.Time{}, fmt.Errorf("missing upload_time for %s", name)
	}
	t, err := time.Parse("2006-01-02T15:04:05", resp.Info.UploadTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse upload_time %q: %w", resp.Info.UploadTime, err)
	}
	return t, nil
}

type npmResponse struct {
	Time map[string]string `json:"time"`
}

func (f *RegistryMetadataFetcher) parseNPMTime(body []byte, name string) (time.Time, error) {
	var resp npmResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return time.Time{}, fmt.Errorf("parse NPM JSON: %w", err)
	}
	created, ok := resp.Time["created"]
	if !ok || created == "" {
		return time.Time{}, fmt.Errorf("missing created time for %s", name)
	}
	t, err := time.Parse(time.RFC3339, created)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse created time %q: %w", created, err)
	}
	return t, nil
}
