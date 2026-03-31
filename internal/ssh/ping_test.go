package ssh

import (
	"net"
	"testing"
	"time"
)

func TestPingHosts_Reachable(t *testing.T) {
	// Start a local TCP listener to simulate a reachable SSH host.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	_, port, _ := net.SplitHostPort(ln.Addr().String())

	hosts := []Host{
		{Name: "local-test", Hostname: "127.0.0.1", Port: port},
	}

	results := PingHosts(hosts, 2*time.Second, 5)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].OK {
		t.Errorf("expected OK=true for local listener, got error: %s", results[0].Error)
	}
	if results[0].Latency <= 0 {
		t.Error("expected positive latency")
	}
}

func TestPingHosts_Unreachable(t *testing.T) {
	hosts := []Host{
		{Name: "unreachable", Hostname: "192.0.2.1", Port: "22"},
	}

	// Very short timeout to fail fast.
	results := PingHosts(hosts, 100*time.Millisecond, 5)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].OK {
		t.Error("expected OK=false for unreachable host")
	}
	if results[0].Error == "" {
		t.Error("expected error message")
	}
}

func TestPingHosts_ConnectionRefused(t *testing.T) {
	// Open a listener to get a free port, then close it so the port is refused.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	_, port, _ := net.SplitHostPort(ln.Addr().String())
	ln.Close()

	hosts := []Host{
		{Name: "refused", Hostname: "127.0.0.1", Port: port},
	}

	results := PingHosts(hosts, 2*time.Second, 5)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].OK {
		t.Error("expected OK=false for refused connection")
	}
	if results[0].Error == "" {
		t.Error("expected error message")
	}
	// connection refused is not a timeout — error should not say "timeout"
	if len(results[0].Error) >= 7 && results[0].Error[:7] == "timeout" {
		t.Errorf("connection refused should not report as timeout, got: %s", results[0].Error)
	}
}

func TestPingHosts_DefaultPort(t *testing.T) {
	// Host with no port should default to 22.
	hosts := []Host{
		{Name: "no-port", Hostname: "192.0.2.1"},
	}

	results := PingHosts(hosts, 100*time.Millisecond, 5)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	// Should fail but not panic.
	if results[0].OK {
		t.Error("expected failure for unreachable host")
	}
}

func TestPingHosts_NoHostname(t *testing.T) {
	// Host with no hostname should use Name as address.
	hosts := []Host{
		{Name: "localhost", Port: "0"},
	}

	results := PingHosts(hosts, 100*time.Millisecond, 5)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestPingHosts_Parallel(t *testing.T) {
	// Test that multiple hosts are pinged concurrently.
	hosts := []Host{
		{Name: "h1", Hostname: "192.0.2.1"},
		{Name: "h2", Hostname: "192.0.2.2"},
		{Name: "h3", Hostname: "192.0.2.3"},
	}

	start := time.Now()
	results := PingHosts(hosts, 200*time.Millisecond, 10)
	elapsed := time.Since(start)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// If run in parallel, total time should be close to timeout, not 3x timeout.
	if elapsed > 500*time.Millisecond {
		t.Errorf("parallel ping took %v, expected < 500ms (should run concurrently)", elapsed)
	}
}

func TestPingHosts_ZeroParallel(t *testing.T) {
	// parallel=0 should default to 10.
	hosts := []Host{
		{Name: "test", Hostname: "192.0.2.1"},
	}

	results := PingHosts(hosts, 100*time.Millisecond, 0)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}
