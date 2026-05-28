package engine

import (
	"context"

	"malFuse/internal/logger"
	"malFuse/internal/osv"
)

type OSVQuerier interface {
	Query(ctx context.Context, name, ecosystem, version string) (osv.Result, error)
}

func OSVCheck(client OSVQuerier, blockOnVuln bool) CheckFunc {
	return func(ctx context.Context, req Request) Result {
		result, err := client.Query(ctx, req.Name, req.Ecosystem, req.Version)
		if err != nil {
			logger.Warn("OSV check failed", "package", req.Name, "error", err)
			return Result{Block: false}
		}
		if result.Vulnerable {
			if blockOnVuln {
				return Result{Block: true, Reason: "osv"}
			}
			logger.Info("OSV: vulnerability found", "package", req.Name, "block_on_vuln", false)
		}
		return Result{Block: false}
	}
}
