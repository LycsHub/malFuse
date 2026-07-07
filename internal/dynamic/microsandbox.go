package dynamic

import (
	"context"
	"os/exec"
	"strings"

	"malFuse/internal/engine"
)

type MicrosandboxRunner struct {
	Binary string
}

func (r MicrosandboxRunner) Run(ctx context.Context, req engine.Request, workspace string, cfg Config) (string, error) {
	binary := r.Binary
	if binary == "" {
		binary = "msb"
	}

	args := []string{
		"run",
		"--rm",
		"--cpus", "1",
		"--memory", "512m",
		"-v", workspace + ":/workspace",
		"--workdir", "/workspace",
	}
	if cfg.Network == "" || cfg.Network == "none" {
		args = append(args, "--no-net")
	}
	args = append(args, imageFor(req, cfg), "sh", "-c", scriptFor(req))

	cmd := exec.CommandContext(ctx, binary, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}
