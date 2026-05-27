package engine

import (
	"context"
	_ "embed"
	"strings"
)

//go:embed packages.txt
var popularPackagesData string

func levenshteinDistance(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	d := make([][]int, la+1)
	for i := range d {
		d[i] = make([]int, lb+1)
		d[i][0] = i
	}
	for j := 0; j <= lb; j++ {
		d[0][j] = j
	}

	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			d[i][j] = min(
				d[i-1][j]+1,
				d[i][j-1]+1,
				d[i-1][j-1]+cost,
			)
		}
	}
	return d[la][lb]
}

func min(vals ...int) int {
	m := vals[0]
	for _, v := range vals[1:] {
		if v < m {
			m = v
		}
	}
	return m
}

func TypoCheck(popular []string, threshold int) CheckFunc {
	if popular == nil {
		popular = loadEmbeddedPackages()
	}
	popMap := make(map[string]struct{}, len(popular))
	for _, p := range popular {
		popMap[p] = struct{}{}
	}

	return func(ctx context.Context, req Request) Result {
		if len(req.Name) < 3 {
			return Result{Block: false}
		}
		if _, ok := popMap[req.Name]; ok {
			return Result{Block: false}
		}
		for _, pkg := range popular {
			if levenshteinDistance(req.Name, pkg) <= threshold {
				return Result{Block: true, Reason: "typo-squatting"}
			}
		}
		return Result{Block: false}
	}
}

func loadEmbeddedPackages() []string {
	lines := strings.Split(strings.TrimSpace(popularPackagesData), "\n")
	packages := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			packages = append(packages, line)
		}
	}
	return packages
}
