package engine

import (
	"context"
	"database/sql"
	"log"

	"malFuse/internal/db/schema"
)

func MaliciousDBCheck(db *sql.DB) CheckFunc {
	return func(ctx context.Context, req Request) Result {
		if db == nil {
			log.Printf("[WARN] malicious-db: database not available, skipping check for %s", req.Name)
			return Result{Block: false}
		}

		found, err := schema.Lookup(db, req.Name, req.Ecosystem, req.Version)
		if err != nil {
			log.Printf("[WARN] malicious-db: lookup failed for %s: %v", req.Name, err)
			return Result{Block: false}
		}

		if found {
			return Result{Block: true, Reason: "malicious-db"}
		}
		return Result{Block: false}
	}
}
