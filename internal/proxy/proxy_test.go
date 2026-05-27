package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"malFuse/internal/engine"
)

type mockEngine struct {
	block  bool
	reason string
}

func (e *mockEngine) Check(_ context.Context, _ engine.Request) engine.Result {
	return engine.Result{Block: e.block, Reason: e.reason}
}

func TestHandlerBlockedRequestReturns403(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("upstream should not be called for blocked requests")
	}))
	defer upstream.Close()

	upstreamURL, _ := url.Parse(upstream.URL)
	handler := New(
		&mockEngine{block: true, reason: "blacklist"},
		map[string]RouteEntry{
			"/pypi/": {Upstream: upstreamURL, Ecosystem: "pypi"},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/pypi/simple/requests/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "blacklist") {
		t.Errorf("expected body to contain 'blacklist', got %s", body)
	}
}

func TestHandlerPassedRequestForwards(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("upstream body"))
	}))
	defer upstream.Close()

	upstreamURL, _ := url.Parse(upstream.URL)
	handler := New(
		&mockEngine{block: false},
		map[string]RouteEntry{
			"/pypi/": {Upstream: upstreamURL, Ecosystem: "pypi"},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/pypi/simple/requests/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "upstream body" {
		t.Errorf("expected body 'upstream body', got %s", rec.Body.String())
	}
}

func TestHandlerUnmatchedRouteReturns502(t *testing.T) {
	handler := New(
		&mockEngine{},
		map[string]RouteEntry{
			"/pypi/": {Upstream: &url.URL{Host: "pypi.org"}, Ecosystem: "pypi"},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/unknown/something", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", rec.Code)
	}
}

func TestHandlerUpstreamUnreachableReturns502(t *testing.T) {
	handler := New(
		&mockEngine{block: false},
		map[string]RouteEntry{
			"/pypi/": {Upstream: &url.URL{Scheme: "http", Host: "127.0.0.1:19999"}, Ecosystem: "pypi"},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/pypi/simple/requests/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected 502 for unreachable upstream, got %d", rec.Code)
	}
}

func TestRouteMatchingStripsPrefix(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/simple/requests/" {
			t.Errorf("expected path /simple/requests/, got %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	upstreamURL, _ := url.Parse(upstream.URL)
	handler := New(
		&mockEngine{block: false},
		map[string]RouteEntry{
			"/pypi/": {Upstream: upstreamURL, Ecosystem: "pypi"},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/pypi/simple/requests/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestPackageNameExtractionPyPI(t *testing.T) {
	tests := []struct {
		path   string
		want   string
		wantOK bool
	}{
		{"/simple/requests/", "requests", true},
		{"/simple/django/", "django", true},
		{"/simple/", "", false},
		{"/other/", "", false},
	}
	for _, tt := range tests {
		name, ok := extractPyPIPackageName(tt.path)
		if ok != tt.wantOK {
			t.Errorf("extractPyPIPackageName(%q) ok = %v, want %v", tt.path, ok, tt.wantOK)
		}
		if name != tt.want {
			t.Errorf("extractPyPIPackageName(%q) = %q, want %q", tt.path, name, tt.want)
		}
	}
}

func TestPackageNameExtractionNPM(t *testing.T) {
	tests := []struct {
		path   string
		want   string
		wantOK bool
	}{
		{"/left-pad", "left-pad", true},
		{"/@scope/pkg", "@scope/pkg", true},
		{"/", "", false},
	}
	for _, tt := range tests {
		name, ok := extractNPMPackageName(tt.path)
		if ok != tt.wantOK {
			t.Errorf("extractNPMPackageName(%q) ok = %v, want %v", tt.path, ok, tt.wantOK)
		}
		if name != tt.want {
			t.Errorf("extractNPMPackageName(%q) = %q, want %q", tt.path, name, tt.want)
		}
	}
}
