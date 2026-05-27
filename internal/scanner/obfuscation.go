package scanner

import (
	"regexp"
)

var (
	b64Pattern = regexp.MustCompile(`[A-Za-z0-9+/=]{100,}`)
	evalPattern = regexp.MustCompile(`\b(eval|exec|Function)\s*\(`)
)

func obfuscationCheck(data []byte, b64MinLength int) ScanResult {
	s := string(data)

	if b64Pattern.MatchString(s) {
		return ScanResult{Block: true, Reason: "obfuscation"}
	}

	if countHexEscapes(s) >= b64MinLength {
		return ScanResult{Block: true, Reason: "obfuscation"}
	}

	if evalPattern.MatchString(s) {
		return ScanResult{Block: true, Reason: "obfuscation"}
	}

	return ScanResult{Block: false}
}

func countHexEscapes(s string) int {
	count := 0
	maxRun := 0
	for i := 0; i < len(s)-3; i++ {
		if s[i] == '\\' && s[i+1] == 'x' &&
			isHex(s[i+2]) && isHex(s[i+3]) {
			count++
			if count > maxRun {
				maxRun = count
			}
			i += 3
		} else {
			count = 0
		}
	}
	return maxRun
}

func isHex(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}
