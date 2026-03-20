package ssh

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// PingResult holds the result of a connectivity check.
type PingResult struct {
	Host    Host          `json:"host"`
	OK      bool          `json:"ok"`
	Latency time.Duration `json:"latency"`
	Error   string        `json:"error,omitempty"`
}

// PingHosts tests connectivity to multiple hosts in parallel.
func PingHosts(hosts []Host, timeout time.Duration, parallel int) []PingResult {
	if parallel <= 0 {
		parallel = 10
	}

	results := make([]PingResult, len(hosts))
	sem := make(chan struct{}, parallel)
	var wg sync.WaitGroup

	for i, h := range hosts {
		wg.Add(1)
		go func(idx int, host Host) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			results[idx] = pingHost(host, timeout)
		}(i, h)
	}

	wg.Wait()
	return results
}

// pingHost tests TCP connectivity to a single host.
func pingHost(h Host, timeout time.Duration) PingResult {
	result := PingResult{Host: h}

	addr := h.Hostname
	if addr == "" {
		addr = h.Name
	}

	port := h.Port
	if port == "" {
		port = "22"
	}

	target := net.JoinHostPort(addr, port)
	start := time.Now()

	conn, err := net.DialTimeout("tcp", target, timeout)
	result.Latency = time.Since(start)

	if err != nil {
		result.OK = false
		result.Error = fmt.Sprintf("timeout (%s)", addr)
		return result
	}
	conn.Close()

	result.OK = true
	return result
}
