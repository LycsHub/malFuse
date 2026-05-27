package ingest

import (
	"testing"
)

func TestParsePyPIOSVFile(t *testing.T) {
	data := []byte(`{
  "modified": "2024-10-24T01:01:57Z",
  "published": "2024-06-25T13:32:04Z",
  "schema_version": "1.5.0",
  "id": "MAL-2024-4725",
  "summary": "Malicious code in 3m-promo-gen-api (PyPI)",
  "affected": [
    {
      "package": {
        "ecosystem": "PyPI",
        "name": "3m-promo-gen-api",
        "purl": "pkg:pypi/3m-promo-gen-api"
      },
      "versions": ["1.0"]
    }
  ]
}`)

	results, err := ParseOSV(data)
	if err != nil {
		t.Fatalf("ParseOSV() error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.Name != "3m-promo-gen-api" {
		t.Errorf("expected Name '3m-promo-gen-api', got %s", r.Name)
	}
	if r.Ecosystem != "pypi" {
		t.Errorf("expected Ecosystem 'pypi', got %s", r.Ecosystem)
	}
	if r.Version != "1.0" {
		t.Errorf("expected Version '1.0', got %s", r.Version)
	}
	if r.Published != "2024-06-25T13:32:04Z" {
		t.Errorf("expected Published '2024-06-25T13:32:04Z', got %s", r.Published)
	}
	if r.ID != "MAL-2024-4725" {
		t.Errorf("expected ID 'MAL-2024-4725', got %s", r.ID)
	}
}

func TestParseNPMOSVFile(t *testing.T) {
	data := []byte(`{
  "id": "MAL-2023-1",
  "published": "2023-01-01T00:00:00Z",
  "affected": [
    {
      "package": {
        "ecosystem": "npm",
        "name": "evil-npm-pkg"
      },
      "versions": ["1.0.0", "2.0.0"]
    }
  ]
}`)

	results, err := ParseOSV(data)
	if err != nil {
		t.Fatalf("ParseOSV() error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results (one per version), got %d", len(results))
	}
	if results[0].Name != "evil-npm-pkg" {
		t.Errorf("expected Name 'evil-npm-pkg', got %s", results[0].Name)
	}
	if results[0].Ecosystem != "npm" {
		t.Errorf("expected Ecosystem 'npm', got %s", results[0].Ecosystem)
	}
}

func TestParseOSVNoVersionsCreatesNullVersion(t *testing.T) {
	data := []byte(`{
  "id": "MAL-2023-99",
  "published": "2023-01-01T00:00:00Z",
  "affected": [
    {
      "package": {
        "ecosystem": "PyPI",
        "name": "bad-pkg"
      },
      "versions": []
    }
  ]
}`)

	results, err := ParseOSV(data)
	if err != nil {
		t.Fatalf("ParseOSV() error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Version != "" {
		t.Errorf("expected empty Version for no versions, got %s", results[0].Version)
	}
}

func TestParseInvalidJSON(t *testing.T) {
	_, err := ParseOSV([]byte(`{invalid}`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestNormalizeEcosystem(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"PyPI", "pypi"},
		{"npm", "npm"},
		{"Go", "go"},
		{"crates.io", "crates.io"},
	}
	for _, tt := range tests {
		result := normalizeEcosystem(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeEcosystem(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
