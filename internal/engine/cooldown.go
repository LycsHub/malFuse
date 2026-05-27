package engine

import (
	"context"
	"log"
	"time"
)

type MetadataFetcher interface {
	FetchPublishTime(ctx context.Context, name, ecosystem string) (time.Time, error)
}

func CooldownCheck(fetcher MetadataFetcher, duration time.Duration) CheckFunc {
	return func(ctx context.Context, req Request) Result {
		publishTime, err := fetcher.FetchPublishTime(ctx, req.Name, req.Ecosystem)
		if err != nil {
			log.Printf("[WARN] Cooldown metadata fetch failed for %s: %v", req.Name, err)
			return Result{Block: true, Reason: "cooldown"}
		}
		if time.Since(publishTime) < duration {
			return Result{Block: true, Reason: "cooldown"}
		}
		return Result{Block: false}
	}
}
