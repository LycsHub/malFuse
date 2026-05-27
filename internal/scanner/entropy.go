package scanner

import "math"

type ScanResult struct {
	Block  bool
	Reason string
}

func shannonEntropy(data []byte) float64 {
	if len(data) == 0 {
		return 0.0
	}
	freq := make(map[byte]int)
	for _, b := range data {
		freq[b]++
	}
	var entropy float64
	invLen := 1.0 / float64(len(data))
	for _, count := range freq {
		p := float64(count) * invLen
		entropy -= p * math.Log2(p)
	}
	return entropy
}

func entropyCheck(data []byte, threshold float64) ScanResult {
	if shannonEntropy(data) > threshold {
		return ScanResult{Block: true, Reason: "entropy"}
	}
	return ScanResult{Block: false}
}
