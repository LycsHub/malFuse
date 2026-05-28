package engine

import (
	"context"

	"malFuse/internal/logger"
	"malFuse/internal/osv"
)

type OSVQuerier interface {
	Query(ctx context.Context, name, ecosystem, version string) (osv.Result, error)
}

func OSVCheck(client OSVQuerier) CheckFunc {
	return func(ctx context.Context, req Request) Result {
		result, err := client.Query(ctx, req.Name, req.Ecosystem, req.Version)
		if err != nil {
			logger.Warn("OSV check failed", "package", req.Name, "error", err)
			return Result{Block: false}
		}
		if result.Vulnerable {
			return Result{Block: true, Reason: "osv"}
		}
		return Result{Block: false}
	}
}
