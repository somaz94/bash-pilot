# Usage

Complete guide for using bash-pilot CLI.

<br/>

## Table of Contents

- [SSH Module](#ssh-module)
- [Global Flags](#global-flags)
- [Output Formats](#output-formats)

<br/>

## SSH Module

<br/>

### ssh list

List SSH hosts grouped by type (git, cloud, k8s, on-prem).

```bash
# List all hosts
bash-pilot ssh list

# JSON output
bash-pilot ssh list -o json

# Plain text (no color)
bash-pilot ssh list --no-color
```

**Output columns:** Host Name, Hostname/IP, User, Identity File

<br/>

### ssh ping

Test TCP connectivity to SSH hosts in parallel.

```bash
# Ping all hosts
bash-pilot ssh ping

# Ping only k8s hosts
bash-pilot ssh ping "k8s-*"

# Ping staging hosts
bash-pilot ssh ping "staging-*"

# JSON output for CI
bash-pilot ssh ping -o json
```

<br/>

### ssh audit

Security audit for SSH keys and configuration.

```bash
# Run full audit
bash-pilot ssh audit

# JSON output
bash-pilot ssh audit -o json
```

**Checks performed:**

| Check | Severity | Description |
|-------|----------|-------------|
| Shared keys | warn | Identity file used by more than 3 hosts |
| File permissions | warn | Key file permissions not 0600 |
| Missing keys | fail | Identity file does not exist |
| No identity file | warn | Host has no IdentityFile directive |

<br/>

## Global Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--config` | `-c` | `~/.config/bash-pilot/config.yaml` | Config file path |
| `--output` | `-o` | `color` | Output format: `color`, `plain`, `json`, `table` |
| `--no-color` | | `false` | Disable color output |

<br/>

## Output Formats

<br/>

### color (default)

ANSI color-coded output with box drawing characters for grouping.

```
┌─ GIT ──────────────────────────────────────────────
│   github.com-somaz94     github.com       somaz     id_rsa_somaz94
└────────────────────────────────────────────────────
```

<br/>

### plain

Same as color but without ANSI escape codes. Useful for piping.

```bash
bash-pilot ssh list --no-color
```

<br/>

### json

Machine-readable JSON output for scripting and CI/CD.

```bash
bash-pilot ssh list -o json | jq '.[].hosts[].name'
```

<br/>

### table

Aligned table with headers and separators.

```bash
bash-pilot ssh list -o table
```
