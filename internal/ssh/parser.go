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
	var seen map[string]bool // parameters already set on current

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
		lower := strings.ToLower(key)
		// ssh hands ProxyCommand to the shell verbatim; every other keyword parsed here ends at a "# comment".
		if lower != "proxycommand" {
			value = stripTrailingComment(value)
		}

		switch lower {
		case "match":
			// A Match block's settings are conditional and belong to no Host above it.
			current = nil
			continue
		case "host":
			// "Host *" holds defaults and a comment-only Host matches nothing; neither is a host.
			if value == "*" || value == "" {
				current = nil
				continue
			}
			hosts = append(hosts, Host{Name: value})
			current = &hosts[len(hosts)-1]
			seen = map[string]bool{}
			continue
		}
		if current == nil || value == "" {
			continue
		}
		// ssh_config(5): the first value obtained wins, and ProxyJump and ProxyCommand share one
		// slot. IdentityFile is cumulative in ssh; its first entry, which ssh tries first, is kept.
		slot := lower
		if slot == "proxycommand" {
			slot = "proxyjump"
		}
		if seen[slot] {
			continue
		}
		switch lower {
		case "hostname":
			current.Hostname = value
		case "user":
			current.User = value
		case "identityfile":
			current.IdentityFile = expandPath(value)
		case "port":
			current.Port = value
		case "proxyjump":
			current.ProxyJump = value
		case "proxycommand":
			current.ProxyCommand = value
		case "forwardagent":
			current.ForwardAgent = strings.EqualFold(value, "yes")
		default:
			continue
		}
		seen[slot] = true
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

// stripTrailingComment cuts value at an unquoted '#' that starts a word, as
// OpenSSH's argument splitter does; a '#' inside a word or quotes is kept.
func stripTrailingComment(value string) string {
	var quote byte
	wordStart := true
	for i := 0; i < len(value); i++ {
		c := value[i]
		switch {
		case c == '\\' && i+1 < len(value) &&
			(strings.IndexByte(`'"\`, value[i+1]) >= 0 || quote == 0 && value[i+1] == ' '):
			i++ // an escaped quote or space neither toggles quoting nor ends the word
			wordStart = false
			continue
		case quote != 0:
			if c == quote {
				quote = 0
			}
		case c == '"' || c == '\'':
			quote = c
		case c == '#' && wordStart:
			return strings.TrimSpace(value[:i])
		}
		wordStart = quote == 0 && (c == ' ' || c == '\t')
	}
	return value
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
