package scanner

import (
	"net"
	"regexp"
)

var (
	urlPattern = regexp.MustCompile(`https?://[^\s'")]+`)
	ipPattern  = regexp.MustCompile(`\b(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\b`)
)

func networkCheck(data []byte, allowPrivateIPs bool) ScanResult {
	s := string(data)

	if urls := urlPattern.FindAllString(s, -1); len(urls) > 0 {
		return ScanResult{Block: true, Reason: "network"}
	}

	if ips := ipPattern.FindAllStringSubmatch(s, -1); len(ips) > 0 {
		for _, match := range ips {
			ip := net.ParseIP(match[1])
			if ip == nil {
				continue
			}
			if ip.IsPrivate() || ip.IsLoopback() {
				if !allowPrivateIPs {
					return ScanResult{Block: true, Reason: "network"}
				}
				continue
			}
			return ScanResult{Block: true, Reason: "network"}
		}
	}

	return ScanResult{Block: false}
}
