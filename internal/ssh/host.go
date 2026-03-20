package ssh

import (
	"net"
	"path/filepath"
	"sort"
	"strings"

	"github.com/somaz94/bash-pilot/internal/config"
)

// Host represents a parsed SSH host entry.
type Host struct {
	Name         string `json:"name"`
	Hostname     string `json:"hostname"`
	User         string `json:"user"`
	Port         string `json:"port,omitempty"`
	IdentityFile string `json:"identity_file"`
	ProxyJump    string `json:"proxy_jump,omitempty"`
	ForwardAgent bool   `json:"forward_agent,omitempty"`
}

// KeyName returns the basename of the identity file.
func (h Host) KeyName() string {
	if h.IdentityFile == "" {
		return ""
	}
	return filepath.Base(h.IdentityFile)
}

// HostGroup represents a named group of hosts.
type HostGroup struct {
	Name  string `json:"name"`
	Label string `json:"label,omitempty"`
	Hosts []Host `json:"hosts"`
}

// GroupHosts organizes hosts into groups based on config patterns.
// Hosts that don't match any pattern are grouped by auto-detection.
func GroupHosts(hosts []Host, sshCfg config.SSHConfig) []HostGroup {
	grouped := make(map[string][]Host)
	labels := make(map[string]string)
	matched := make(map[int]bool)

	// First pass: match hosts against configured patterns.
	for i, h := range hosts {
		for groupName, group := range sshCfg.Groups {
			for _, pattern := range group.Pattern {
				if matchPattern(h.Name, pattern) {
					grouped[groupName] = append(grouped[groupName], h)
					if group.Label != "" {
						labels[groupName] = group.Label
					}
					matched[i] = true
					break
				}
			}
			if matched[i] {
				break
			}
		}
	}

	// Second pass: auto-group unmatched hosts.
	for i, h := range hosts {
		if matched[i] {
			continue
		}
		group := autoDetectGroup(h)
		grouped[group] = append(grouped[group], h)
	}

	// Build sorted group list.
	order := []string{"git", "cloud", "k8s", "on-prem", "other"}
	var result []HostGroup
	seen := make(map[string]bool)

	for _, name := range order {
		if hosts, ok := grouped[name]; ok {
			result = append(result, HostGroup{
				Name:  name,
				Label: labels[name],
				Hosts: hosts,
			})
			seen[name] = true
		}
	}

	// Add any remaining groups not in the predefined order.
	var remaining []string
	for name := range grouped {
		if !seen[name] {
			remaining = append(remaining, name)
		}
	}
	sort.Strings(remaining)
	for _, name := range remaining {
		result = append(result, HostGroup{
			Name:  name,
			Label: labels[name],
			Hosts: grouped[name],
		})
	}

	return result
}

// autoDetectGroup guesses a group based on host properties.
func autoDetectGroup(h Host) string {
	name := strings.ToLower(h.Name)
	hostname := strings.ToLower(h.Hostname)

	// Git hosts.
	if strings.Contains(name, "github") || strings.Contains(name, "gitlab") ||
		strings.Contains(name, "codecommit") || strings.Contains(name, "bitbucket") {
		return "git"
	}

	// Kubernetes hosts.
	if strings.HasPrefix(name, "k8s-") || strings.Contains(name, "kube") ||
		strings.Contains(name, "master") || strings.Contains(name, "node") {
		return "k8s"
	}

	// Cloud vs on-prem based on IP.
	if hostname != "" {
		ip := net.ParseIP(hostname)
		if ip != nil {
			if isPrivateIP(ip) {
				return "on-prem"
			}
			return "cloud"
		}
		// If hostname is a FQDN with cloud provider keywords.
		if strings.Contains(hostname, "amazonaws.com") || strings.Contains(hostname, "compute.google") ||
			strings.Contains(hostname, "azure") {
			return "cloud"
		}
	}

	return "other"
}

// isPrivateIP checks if an IP is in a private range.
func isPrivateIP(ip net.IP) bool {
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}
	for _, cidr := range privateRanges {
		_, network, _ := net.ParseCIDR(cidr)
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// matchPattern checks if a hostname matches a glob-like pattern.
func matchPattern(name, pattern string) bool {
	matched, _ := filepath.Match(pattern, name)
	if matched {
		return true
	}
	// Also try matching with the pattern as a prefix.
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(name, pattern[:len(pattern)-1])
	}
	return false
}
