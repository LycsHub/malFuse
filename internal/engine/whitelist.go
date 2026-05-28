package engine

import (
	"context"
	"database/sql"

	"malFuse/internal/db/schema"
)

func WhitelistCheck(db *sql.DB) CheckFunc {
	return func(ctx context.Context, req Request) Result {
		if db == nil {
			return Result{Block: false}
		}
		found, _ := schema.IsWhitelisted(db, req.Name, req.Ecosystem, req.Version)
		if found {
			return Result{Skip: true, Reason: "whitelist"}
		}
		return Result{Block: false}
	}
}
