package engine

import (
	"context"
	"time"

	"malFuse/internal/logger"
)

type MetadataFetcher interface {
	FetchPublishTime(ctx context.Context, name, ecosystem string) (time.Time, error)
}

func CooldownCheck(fetcher MetadataFetcher, duration time.Duration) CheckFunc {
	return func(ctx context.Context, req Request) Result {
		publishTime, err := fetcher.FetchPublishTime(ctx, req.Name, req.Ecosystem)
		if err != nil {
			logger.Warn("cooldown metadata fetch failed", "package", req.Name, "error", err)
			return Result{Block: true, Reason: "cooldown"}
		}
		if time.Since(publishTime) < duration {
			return Result{Block: true, Reason: "cooldown"}
		}
		return Result{Block: false}
	}
}
