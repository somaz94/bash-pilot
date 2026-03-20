package ssh

import (
	"net"
	"testing"

	"github.com/somaz94/bash-pilot/internal/config"
)

func TestGroupHosts(t *testing.T) {
	hosts := []Host{
		{Name: "github.com-somaz94", Hostname: "github.com", User: "somaz"},
		{Name: "test-server", Hostname: "3.65.182.184", User: "ec2-user"},
		{Name: "nas", Hostname: "10.10.10.5", User: "somaz"},
		{Name: "k8s-control-01", Hostname: "10.10.10.17", User: "concrit"},
	}

	cfg := config.SSHConfig{
		Groups: map[string]config.SSHGroup{
			"git": {Pattern: []string{"github.com*"}},
			"k8s": {Pattern: []string{"k8s-*"}},
		},
	}

	groups := GroupHosts(hosts, cfg)

	groupMap := make(map[string]HostGroup)
	for _, g := range groups {
		groupMap[g.Name] = g
	}

	if g, ok := groupMap["git"]; !ok || len(g.Hosts) != 1 {
		t.Errorf("expected git group with 1 host, got %v", groupMap["git"])
	}
	if g, ok := groupMap["k8s"]; !ok || len(g.Hosts) != 1 {
		t.Errorf("expected k8s group with 1 host, got %v", groupMap["k8s"])
	}
	if g, ok := groupMap["cloud"]; !ok || len(g.Hosts) != 1 {
		t.Errorf("expected cloud group with 1 host (test-server), got %v", groupMap["cloud"])
	}
	if g, ok := groupMap["on-prem"]; !ok || len(g.Hosts) != 1 {
		t.Errorf("expected on-prem group with 1 host (nas), got %v", groupMap["on-prem"])
	}
}

func TestGroupHosts_WithLabel(t *testing.T) {
	hosts := []Host{
		{Name: "staging-web", Hostname: "52.78.100.10"},
	}
	cfg := config.SSHConfig{
		Groups: map[string]config.SSHGroup{
			"cloud": {Pattern: []string{"staging-*"}, Label: "AWS Seoul"},
		},
	}

	groups := GroupHosts(hosts, cfg)
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if groups[0].Label != "AWS Seoul" {
		t.Errorf("Label = %q, want %q", groups[0].Label, "AWS Seoul")
	}
}

func TestGroupHosts_CustomGroup(t *testing.T) {
	// Custom group name not in predefined order should still appear.
	hosts := []Host{
		{Name: "custom-host", Hostname: "10.0.0.1"},
	}
	cfg := config.SSHConfig{
		Groups: map[string]config.SSHGroup{
			"myteam": {Pattern: []string{"custom-*"}},
		},
	}

	groups := GroupHosts(hosts, cfg)
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if groups[0].Name != "myteam" {
		t.Errorf("Name = %q, want %q", groups[0].Name, "myteam")
	}
}

func TestGroupHosts_EmptyConfig(t *testing.T) {
	hosts := []Host{
		{Name: "github.com-test", Hostname: "github.com"},
		{Name: "server1", Hostname: "10.10.10.1"},
	}
	cfg := config.SSHConfig{}

	groups := GroupHosts(hosts, cfg)
	if len(groups) == 0 {
		t.Error("expected auto-detected groups")
	}
}

func TestAutoDetectGroup(t *testing.T) {
	tests := []struct {
		host Host
		want string
	}{
		// Git hosts.
		{Host{Name: "github.com-somaz94", Hostname: "github.com"}, "git"},
		{Host{Name: "gitlab", Hostname: "10.10.10.60"}, "git"},
		{Host{Name: "my-bitbucket", Hostname: "bitbucket.org"}, "git"},
		{Host{Name: "codecommit-repo", Hostname: "git-codecommit.us-east-1.amazonaws.com"}, "git"},
		// Kubernetes hosts.
		{Host{Name: "k8s-control-01", Hostname: "10.10.10.17"}, "k8s"},
		{Host{Name: "kube-master", Hostname: "10.0.0.1"}, "k8s"},
		// Cloud hosts (public IP).
		{Host{Name: "test-server", Hostname: "3.65.182.184"}, "cloud"},
		// Cloud hosts (FQDN).
		{Host{Name: "aws-instance", Hostname: "ec2-1-2-3-4.compute.amazonaws.com"}, "cloud"},
		{Host{Name: "gcp-vm", Hostname: "my-vm.compute.google.com"}, "cloud"},
		{Host{Name: "azure-vm", Hostname: "my-vm.azure.com"}, "cloud"},
		// On-prem hosts.
		{Host{Name: "nas", Hostname: "10.10.10.5"}, "on-prem"},
		{Host{Name: "server1", Hostname: "192.168.1.100"}, "on-prem"},
		{Host{Name: "internal", Hostname: "172.16.0.50"}, "on-prem"},
		// Other (no hostname).
		{Host{Name: "random", Hostname: ""}, "other"},
		// Other (non-IP, non-cloud FQDN).
		{Host{Name: "myhost", Hostname: "myhost.local"}, "other"},
	}

	for _, tt := range tests {
		got := autoDetectGroup(tt.host)
		if got != tt.want {
			t.Errorf("autoDetectGroup(%q, hostname=%q) = %q, want %q", tt.host.Name, tt.host.Hostname, got, tt.want)
		}
	}
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    bool
	}{
		{"github.com-somaz94", "github.com*", true},
		{"k8s-control-01", "k8s-*", true},
		{"test-server", "k8s-*", false},
		{"nas", "nas*", true},
		{"nas", "nas", true},
		// Exact match via filepath.Match.
		{"server1", "server1", true},
		{"server1", "server2", false},
		// Non-matching wildcard.
		{"prod-web", "staging-*", false},
	}

	for _, tt := range tests {
		got := matchPattern(tt.name, tt.pattern)
		if got != tt.want {
			t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.name, tt.pattern, got, tt.want)
		}
	}
}

func TestKeyName(t *testing.T) {
	h := Host{IdentityFile: "/home/user/.ssh/id_rsa_somaz94"}
	if got := h.KeyName(); got != "id_rsa_somaz94" {
		t.Errorf("KeyName() = %q, want %q", got, "id_rsa_somaz94")
	}

	h2 := Host{}
	if got := h2.KeyName(); got != "" {
		t.Errorf("KeyName() = %q, want empty", got)
	}
}

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{"10.10.10.5", true},
		{"172.16.0.1", true},
		{"192.168.1.1", true},
		{"3.65.182.184", false},
		{"8.8.8.8", false},
	}

	for _, tt := range tests {
		ip := net.ParseIP(tt.ip)
		if ip == nil {
			t.Fatalf("failed to parse IP: %s", tt.ip)
		}
		got := isPrivateIP(ip)
		if got != tt.want {
			t.Errorf("isPrivateIP(%s) = %v, want %v", tt.ip, got, tt.want)
		}
	}
}
