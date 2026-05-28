package linker

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"malFuse/internal/logger"
)

type LinkConfig struct {
	ProxyHost    string
	ProxyPort    string
	PypiUpstream string
	NpmUpstream  string
}

func (c LinkConfig) PipURL() string {
	return fmt.Sprintf("http://%s:%s/pypi/simple/", c.ProxyHost, c.ProxyPort)
}

func (c LinkConfig) NpmURL() string {
	return fmt.Sprintf("http://%s:%s/npm/", c.ProxyHost, c.ProxyPort)
}

type Backup struct {
	PipOriginal string `json:"pip_original"`
	NpmOriginal string `json:"npm_original"`
}

func Link(config LinkConfig, which string) error {
	backup := &Backup{}
	backupPath := getBackupPath()

	if existing, err := loadBackup(backupPath); err == nil {
		backup = existing
	}

	if which == "" || which == "pip" {
		// get current value
		out, _ := exec.Command("pip", "config", "get", "global.index-url").Output()
		if backup.PipOriginal == "" {
			backup.PipOriginal = cleanOutput(out)
		}
		cmd := exec.Command("pip", "config", "set", "global.index-url", config.PipURL())
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("pip config set: %w", err)
		}
		logger.Info("pip configured", "url", config.PipURL())
	}

	if which == "" || which == "npm" {
		out, _ := exec.Command("npm", "config", "get", "registry").Output()
		if backup.NpmOriginal == "" {
			backup.NpmOriginal = cleanOutput(out)
		}
		cmd := exec.Command("npm", "config", "set", "registry", config.NpmURL())
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("npm config set: %w", err)
		}
		logger.Info("npm configured", "url", config.NpmURL())
	}

	return saveBackup(backupPath, backup)
}

func Unlink(which string) error {
	backupPath := getBackupPath()
	backup, err := loadBackup(backupPath)
	if err != nil {
		return fmt.Errorf("no link backup found: %w", err)
	}

	if (which == "" || which == "pip") && backup.PipOriginal != "" {
		cmd := exec.Command("pip", "config", "set", "global.index-url", backup.PipOriginal)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("pip unlink: %w", err)
		}
		logger.Info("pip restored", "url", backup.PipOriginal)
		backup.PipOriginal = ""
	}

	if (which == "" || which == "npm") && backup.NpmOriginal != "" {
		cmd := exec.Command("npm", "config", "set", "registry", backup.NpmOriginal)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("npm unlink: %w", err)
		}
		logger.Info("npm restored", "url", backup.NpmOriginal)
		backup.NpmOriginal = ""
	}

	if backup.PipOriginal == "" && backup.NpmOriginal == "" {
		os.Remove(backupPath)
		return nil
	}
	return saveBackup(backupPath, backup)
}

func getBackupPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".malfuse_backup.json")
}

func saveBackup(path string, b *Backup) error {
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func loadBackup(path string) (*Backup, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var b Backup
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

func cleanOutput(out []byte) string {
	if len(out) == 0 {
		return ""
	}
	// trim trailing newline
	if out[len(out)-1] == '\n' {
		return string(out[:len(out)-1])
	}
	return string(out)
}
