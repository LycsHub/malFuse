package engine

import "context"

type BlacklistEntry struct {
	Name    string
	Version string
}

func BlacklistCheck(entries []BlacklistEntry) CheckFunc {
	return func(ctx context.Context, req Request) Result {
		for _, entry := range entries {
			if req.Name != entry.Name {
				continue
			}
			if entry.Version != "" && entry.Version != req.Version {
				continue
			}
			return Result{Block: true, Reason: "blacklist"}
		}
		return Result{Block: false}
	}
}
