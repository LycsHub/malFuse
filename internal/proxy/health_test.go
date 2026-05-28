package proxy

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"net/url"
	"testing"

	"malFuse/internal/engine"
)

var errPing = errors.New("ping failed")

type mockDBPinger struct {
	pingErr error
}

func (m *mockDBPinger) Ping() error { return m.pingErr }

func TestHealthOK(t *testing.T) {
	db := &mockDBPinger{pingErr: nil}
	handler := New(engine.New(), map[string]RouteEntry{
		"/pypi/": {Upstream: &url.URL{Scheme: "http", Host: "example.com"}, Ecosystem: "pypi"},
	})
	handler.SetDBPinger(db)

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %v", resp["status"])
	}
	if resp["db"] != true {
		t.Errorf("expected db true, got %v", resp["db"])
	}
	if resp["uptime"] == nil {
		t.Error("expected uptime field")
	}
}

func TestHealthDBDegraded(t *testing.T) {
	db := &mockDBPinger{pingErr: errPing}
	handler := New(engine.New(), map[string]RouteEntry{
		"/pypi/": {Upstream: &url.URL{Scheme: "http", Host: "example.com"}, Ecosystem: "pypi"},
	})
	handler.SetDBPinger(db)

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["status"] != "degraded" {
		t.Errorf("expected status degraded, got %v", resp["status"])
	}
	if resp["db"] != false {
		t.Errorf("expected db false, got %v", resp["db"])
	}
}

func TestHealthNoDB(t *testing.T) {
	handler := New(engine.New(), map[string]RouteEntry{
		"/pypi/": {Upstream: &url.URL{Scheme: "http", Host: "example.com"}, Ecosystem: "pypi"},
	})

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status ok without DB, got %v", resp["status"])
	}
	if resp["db"] != true {
		t.Errorf("expected db true when no pinger, got %v", resp["db"])
	}
}
