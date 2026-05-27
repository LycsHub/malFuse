package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"malFuse/internal/db/schema"
	"malFuse/internal/engine"
)

func TestMaliciousDBIntegration(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := schema.Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer db.Close()

	schema.InsertOrReplace(db, schema.MaliciousPackage{
		Name:      "evil-pkg",
		Ecosystem: "pypi",
		Version:   "1.0",
		Published: "2024-01-01",
		Source:    "MAL-2024-1",
	})
	schema.InsertOrReplace(db, schema.MaliciousPackage{
		Name:      "bad-lib",
		Ecosystem: "pypi",
		Published: "2024-02-01",
		Source:    "MAL-2024-2",
	})

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	upstreamURL, _ := url.Parse(upstream.URL)

	eng := engine.New(engine.MaliciousDBCheck(db))
	handler := New(eng, map[string]RouteEntry{
		"/pypi/": {Upstream: upstreamURL, Ecosystem: "pypi"},
	})

	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{"package with version in db blocked", "/pypi/simple/evil-pkg/", http.StatusForbidden},
		{"package without version in db blocked", "/pypi/simple/bad-lib/", http.StatusForbidden},
		{"safe package passed", "/pypi/simple/requests/", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			// Pass context so engine doesn't cancel
			req = req.WithContext(context.Background())
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected %d, got %d (body: %s)", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}
