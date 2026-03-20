package ssh

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// ParseConfig reads and parses an SSH config file into a list of Host entries.
func ParseConfig(path string) ([]Host, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(home, ".ssh", "config")
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var hosts []Host
	var current *Host

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle Include directives (skip for now).
		if strings.HasPrefix(strings.ToLower(line), "include ") {
			continue
		}

		key, value := parseKeyValue(line)
		if key == "" {
			continue
		}

		switch strings.ToLower(key) {
		case "host":
			// Skip wildcard-only entries like "Host *".
			if value == "*" {
				current = nil
				continue
			}
			h := Host{
				Name: value,
			}
			hosts = append(hosts, h)
			current = &hosts[len(hosts)-1]
		case "hostname":
			if current != nil {
				current.Hostname = value
			}
		case "user":
			if current != nil {
				current.User = value
			}
		case "identityfile":
			if current != nil {
				current.IdentityFile = expandPath(value)
			}
		case "port":
			if current != nil {
				current.Port = value
			}
		case "proxycommand", "proxyjump":
			if current != nil {
				current.ProxyJump = value
			}
		case "forwardagent":
			if current != nil {
				current.ForwardAgent = strings.EqualFold(value, "yes")
			}
		}
	}

	return hosts, scanner.Err()
}

// parseKeyValue splits "Key Value" or "Key=Value" into key and value.
func parseKeyValue(line string) (string, string) {
	// Try equals-separated first (e.g., "Host=value").
	if idx := strings.Index(line, "="); idx > 0 {
		return strings.TrimSpace(line[:idx]), strings.TrimSpace(line[idx+1:])
	}

	// Try space/tab-separated (e.g., "Host value").
	parts := strings.SplitN(line, " ", 2)
	if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}

	// Try tab-separated.
	parts = strings.SplitN(line, "\t", 2)
	if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}

	return line, ""
}

// expandPath replaces ~ with the home directory.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
