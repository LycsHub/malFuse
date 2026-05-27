package engine

import (
	"context"
	"log"

	"malFuse/internal/osv"
)

type OSVQuerier interface {
	Query(ctx context.Context, name, ecosystem, version string) (osv.Result, error)
}

func OSVCheck(client OSVQuerier) CheckFunc {
	return func(ctx context.Context, req Request) Result {
		result, err := client.Query(ctx, req.Name, req.Ecosystem, req.Version)
		if err != nil {
			log.Printf("[WARN] OSV check failed for %s: %v", req.Name, err)
			return Result{Block: false}
		}
		if result.Vulnerable {
			return Result{Block: true, Reason: "osv"}
		}
		return Result{Block: false}
	}
}
