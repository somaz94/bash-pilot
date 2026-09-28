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
	} else {
		path = expandPath(path)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var hosts []Host
	var current *Host

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Include is not followed: hosts defined in included files are not listed.
		if strings.HasPrefix(strings.ToLower(line), "include ") {
			continue
		}

		key, value := parseKeyValue(line)
		if key == "" {
			continue
		}

		switch strings.ToLower(key) {
		case "host":
			// "Host *" holds global defaults, not a host; nil current drops its directives.
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
		case "proxyjump":
			// ssh_config(5): whichever of ProxyJump / ProxyCommand comes first wins.
			if current != nil && current.ProxyCommand == "" {
				current.ProxyJump = value
			}
		case "proxycommand":
			if current != nil && current.ProxyJump == "" {
				current.ProxyCommand = value
			}
		case "forwardagent":
			if current != nil {
				current.ForwardAgent = strings.EqualFold(value, "yes")
			}
		}
	}

	return hosts, scanner.Err()
}

// parseKeyValue splits an ssh_config line into keyword and argument. Per
// ssh_config(5) the keyword ends at the first whitespace or '=', and the
// separator is whitespace with at most one '='; later '=' belong to the value.
func parseKeyValue(line string) (string, string) {
	idx := strings.IndexAny(line, " \t=")
	if idx == -1 {
		return line, ""
	}
	rest := strings.TrimLeft(line[idx:], " \t")
	rest = strings.TrimPrefix(rest, "=")
	return line[:idx], strings.TrimSpace(rest)
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
