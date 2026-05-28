package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"malFuse/internal/db/schema"
	"malFuse/internal/engine"
)

func TestIntegrationBlockMaliciousInDB(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := schema.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	schema.InsertOrReplace(db, schema.MaliciousPackage{
		Name: "evil-pkg", Ecosystem: "pypi",
		Published: "2024-01-01", Source: "MAL-1",
	})

	eng := engine.New(engine.MaliciousDBCheck(db))
	handler := New(eng, map[string]RouteEntry{
		"/pypi/": {Upstream: &url.URL{Scheme: "http", Host: "127.0.0.1:1"}, Ecosystem: "pypi"},
	})

	req := httptest.NewRequest("GET", "/pypi/simple/evil-pkg/", nil)
	req = req.WithContext(context.Background())
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "malicious-db" {
		t.Errorf("expected body 'malicious-db', got %q", rec.Body.String())
	}
}

func TestIntegrationMissingDBGraceful(t *testing.T) {
	eng := engine.New(engine.MaliciousDBCheck(nil))
	handler := New(eng, map[string]RouteEntry{
		"/pypi/": {Upstream: &url.URL{Scheme: "http", Host: "127.0.0.1:1"}, Ecosystem: "pypi"},
	})

	req := httptest.NewRequest("GET", "/pypi/simple/requests/", nil)
	req = req.WithContext(context.Background())
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// nil DB → MaliciousDBCheck returns PASS → proxy tries upstream → 502 (unreachable)
	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected 502 (missing DB skips check, upstream unreachable), got %d", rec.Code)
	}
}

func TestIntegrationUnmatchedRoute(t *testing.T) {
	handler := New(engine.New(), map[string]RouteEntry{
		"/pypi/": {Upstream: &url.URL{Scheme: "http", Host: "example.com"}, Ecosystem: "pypi"},
	})

	req := httptest.NewRequest("GET", "/unknown/something", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected 502 for unmatched route, got %d", rec.Code)
	}
}

func TestIntegrationGracefulShutdown(t *testing.T) {
	handler := New(engine.New(), map[string]RouteEntry{
		"/pypi/": {Upstream: &url.URL{Scheme: "http", Host: "example.com"}, Ecosystem: "pypi"},
	})

	srv := &http.Server{
		Addr:    "127.0.0.1:0",
		Handler: handler,
	}

	done := make(chan error, 1)
	go func() {
		done <- srv.ListenAndServe()
	}()

	time.Sleep(50 * time.Millisecond)

	go func() {
		time.Sleep(100 * time.Millisecond)
		srv.Shutdown(context.Background())
	}()

	select {
	case err := <-done:
		if err != http.ErrServerClosed {
			t.Errorf("expected ErrServerClosed, got %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Error("server did not shut down within 3s")
		srv.Close()
	}
}
