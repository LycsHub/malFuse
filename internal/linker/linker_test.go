package linker

import (
	"testing"
)

func TestSaveRestoreBackup(t *testing.T) {
	path := t.TempDir() + "/backup.json"
	t.Setenv("HOME", t.TempDir())

	b := &Backup{
		PipOriginal:  "https://pypi.org/simple/",
		NpmOriginal:  "https://registry.npmjs.org/",
		PnpmOriginal: "https://registry.npmjs.org/",
		YarnOriginal: "https://registry.yarnpkg.com/",
	}

	if err := saveBackup(path, b); err != nil {
		t.Fatalf("saveBackup: %v", err)
	}

	loaded, err := loadBackup(path)
	if err != nil {
		t.Fatalf("loadBackup: %v", err)
	}
	if loaded.PipOriginal != b.PipOriginal {
		t.Errorf("expected PipOriginal %s, got %s", b.PipOriginal, loaded.PipOriginal)
	}
	if loaded.PnpmOriginal != b.PnpmOriginal {
		t.Errorf("expected PnpmOriginal %s, got %s", b.PnpmOriginal, loaded.PnpmOriginal)
	}
	if loaded.YarnOriginal != b.YarnOriginal {
		t.Errorf("expected YarnOriginal %s, got %s", b.YarnOriginal, loaded.YarnOriginal)
	}
}

func TestLoadBackupMissing(t *testing.T) {
	_, err := loadBackup("/nonexistent/backup.json")
	if err == nil {
		t.Error("expected error for missing backup")
	}
}

func TestGetBackupPath(t *testing.T) {
	t.Setenv("HOME", "/home/test")
	path := getBackupPath()
	if path != "/home/test/.malfuse_backup.json" {
		t.Errorf("expected /home/test/.malfuse_backup.json, got %s", path)
	}
}

func TestLinkConfig(t *testing.T) {
	lc := LinkConfig{
		ProxyHost: "127.0.0.1",
		ProxyPort: "8080",
		PypiUpstream: "https://pypi.org/simple/",
		NpmUpstream: "https://registry.npmjs.org/",
	}
	if lc.PipURL() != "http://127.0.0.1:8080/pypi/simple/" {
		t.Errorf("expected PipURL, got %s", lc.PipURL())
	}
	if lc.NpmURL() != "http://127.0.0.1:8080/npm/" {
		t.Errorf("expected NpmURL, got %s", lc.NpmURL())
	}
}
