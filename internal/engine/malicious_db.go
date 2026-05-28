package engine

import (
	"context"
	"database/sql"

	"malFuse/internal/db/schema"
	"malFuse/internal/logger"
)

func MaliciousDBCheck(db *sql.DB) CheckFunc {
	return func(ctx context.Context, req Request) Result {
		if db == nil {
			logger.Warn("malicious-db: database not available, skipping check", "package", req.Name)
			return Result{Block: false}
		}

		found, err := schema.Lookup(db, req.Name, req.Ecosystem, req.Version)
		if err != nil {
			logger.Warn("malicious-db: lookup failed", "package", req.Name, "error", err)
			return Result{Block: false}
		}

		if found {
			return Result{Block: true, Reason: "malicious-db"}
		}
		return Result{Block: false}
	}
}
