# Usage

Complete guide for using bash-pilot CLI.

<br/>

## Table of Contents

- [Init](#init)
- [SSH Module](#ssh-module)
- [Git Module](#git-module)
- [Env Module](#env-module)
- [Prompt Module](#prompt-module)
- [Snapshot Module](#snapshot-module)
- [Migrate Module](#migrate-module)
- [Doctor (Cross-Module)](#doctor-cross-module)
- [Global Flags](#global-flags)
- [Output Formats](#output-formats)

<br/>

## Init

Auto-generate config from your existing SSH config.

```bash
# Generate ~/.config/bash-pilot/config.yaml
bash-pilot init

# Overwrite existing config
bash-pilot init --force
```

Analyzes `~/.ssh/config`, auto-detects host groups (git, cloud, k8s, on-prem), and writes the config file. If a config already exists, shows a preview without overwriting unless `--force` is used.

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

## Git Module

<br/>

### git profiles

List git identity profiles from `~/.gitconfig` includeIf directives.

```bash
# List all profiles
bash-pilot git profiles

# JSON output
bash-pilot git profiles -o json

# Specify gitconfig path
bash-pilot git profiles --gitconfig /path/to/.gitconfig
```

Shows active profile based on current working directory. Profiles are detected from `includeIf "gitdir:..."` directives.

<br/>

### git doctor

Diagnose common gitconfig issues.

```bash
# Run diagnostics
bash-pilot git doctor

# JSON output
bash-pilot git doctor -o json
```

**Checks performed:**

| Check | Severity | Description |
|-------|----------|-------------|
| Duplicate safe.directory | warn | Same directory listed multiple times |
| Missing includeIf target | warn | Included config file does not exist |
| No user.email | warn | No global email configured (unless using includeIf) |
| Duplicate remote URLs | warn | Same URL configured in multiple remotes |
| File permissions | warn | Gitconfig permissions too open (not 0600) |

<br/>

### git clean

Remove duplicate and stale entries from gitconfig.

```bash
# Preview what would be removed
bash-pilot git clean --dry-run

# Actually clean up
bash-pilot git clean

# JSON output
bash-pilot git clean -o json
```

**What gets cleaned:**
- Duplicate `safe.directory` entries
- Stale `safe.directory` entries (directory no longer exists)

A backup is automatically created at `~/.gitconfig.bak` before any changes.

<br/>

## Env Module

<br/>

### env check

Shell environment health scan — checks shell, tools, SSH agent, git config, editor, and home directory.

```bash
# Run full environment check
bash-pilot env check

# JSON output
bash-pilot env check -o json
```

**Checks performed:**

| Check | Category | Description |
|-------|----------|-------------|
| Shell | shell | SHELL env var, bash version (warns on 3.x) |
| Common tools | tools | git, ssh, curl (required); make, docker, kubectl, helm, terraform, go, node, python3 (optional) |
| SSH agent | ssh-agent | SSH_AUTH_SOCK set, socket exists, keys loaded |
| Git config | git | Global user.email and user.name |
| Home directory | home | ~/.ssh permissions, ~/.config exists, shell profile exists |
| Editor | editor | EDITOR or VISUAL env var set |

<br/>

### env path

Analyze the PATH environment variable.

```bash
# Show all PATH entries with status
bash-pilot env path

# JSON output
bash-pilot env path -o json
```

**Reports:**
- All PATH entries with existence check
- Duplicate directories (by canonical path)
- Missing directories

<br/>

## Prompt Module

<br/>

### prompt init

Generate a smart bash prompt script with git branch, dirty status, exit code indicator, and optional k8s context.

```bash
# Preview the generated script
bash-pilot prompt init

# Apply immediately
eval "$(bash-pilot prompt init)"

# Full theme (git + k8s context)
eval "$(bash-pilot prompt init --theme full)"

# Full theme without k8s
eval "$(bash-pilot prompt init --theme full --no-k8s)"

# Persist in your shell profile
echo 'eval "$(bash-pilot prompt init)"' >> ~/.bashrc
```

**Themes:**

| Theme | Description |
|-------|-------------|
| `minimal` (default) | user@host, directory, git branch/dirty, exit code |
| `full` | All of minimal + k8s context:namespace |

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--theme` | `minimal` | Prompt theme: `minimal`, `full` |
| `--no-k8s` | `false` | Disable k8s context (even in full theme) |

<br/>

### prompt show

Preview what components would appear in the prompt for your current environment.

```bash
# Show current prompt components
bash-pilot prompt show

# Full theme preview
bash-pilot prompt show --theme full

# JSON output
bash-pilot prompt show -o json
```

<br/>

## Snapshot Module

<br/>

### snapshot

Capture a full snapshot of the current environment to JSON — tools, versions, shell, git identity, SSH keys, k8s contexts, PATH, and brew packages.

```bash
# Save snapshot to file
bash-pilot snapshot > my-env.json

# Preview snapshot summary
bash-pilot snapshot --summary

# JSON output (default)
bash-pilot snapshot
```

**Captured data:**

| Category | Details |
|----------|---------|
| System | OS, architecture, hostname, timestamp |
| Shell | SHELL, version (bash), EDITOR/VISUAL, TERM |
| Tools | git, ssh, curl, make, docker, kubectl, helm, terraform, go, node, python3 (name, path, version) |
| Git | Global user.email, user.name, includeIf profiles |
| SSH Keys | Key names, fingerprints, types |
| K8s | Context names, current context |
| PATH | All PATH entries |
| Brew | Installed formulae (macOS only) |

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--summary` | `false` | Show human-readable summary instead of full JSON |

<br/>

### diff

Compare a saved environment snapshot against the current environment.

```bash
# Compare snapshot
bash-pilot diff my-env.json

# JSON output
bash-pilot diff my-env.json -o json
```

**Comparison sections:** System, Tools, Git, SSH Keys, K8s Contexts, Brew Packages

**Entry statuses:**

| Status | Symbol | Description |
|--------|--------|-------------|
| match | (hidden) | Value identical in both |
| mismatch | `~` | Value changed |
| missing | `-` | In snapshot but not current |
| extra | `+` | In current but not snapshot |

<br/>

### setup

Install missing tools from a saved snapshot — turns `diff` results into actions.

```bash
# Preview what would be installed
bash-pilot setup my-env.json --dry-run

# Install missing tools
bash-pilot setup my-env.json

# JSON output
bash-pilot setup my-env.json -o json
```

**What gets installed:**
- Missing tools detected in the Tools section of the diff (via brew on macOS, apt/snap on Linux)
- Missing brew packages (macOS only)

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--dry-run` | `false` | Preview install plan without executing |

**Action statuses:**

| Status | Description |
|--------|-------------|
| pending | Will be installed (shown in dry-run) |
| installed | Successfully installed |
| skipped | No known install command for this tool/OS |
| failed | Install command returned an error |

<br/>

## Migrate Module

<br/>

### migrate export

Export SSH hosts, key references, and Git profiles to a portable JSON format. Private keys are NOT included.

```bash
# Export to file
bash-pilot migrate export > my-config.json

# Pipe to JSON viewer
bash-pilot migrate export | jq .
```

**Exported data:**

| Category | Details |
|----------|---------|
| SSH Hosts | name, hostname, user, port, identity file (tilde-relative), proxy, forward agent |
| SSH Keys | name, type, path (no private key content) |
| Git Identity | global user.name, user.email |
| Git Profiles | includeIf profile name, directory, email, signing key |

All paths are stored as `~/`-relative for cross-platform portability.

<br/>

### migrate import

Import SSH hosts and Git profiles from a migrate config file. Paths are automatically translated to the local home directory.

```bash
# Preview changes
bash-pilot migrate import my-config.json --dry-run

# Apply changes
bash-pilot migrate import my-config.json

# JSON output
bash-pilot migrate import my-config.json -o json
```

**Import behavior:**

| Item | Action |
|------|--------|
| SSH hosts | Appended to `~/.ssh/config` (existing hosts skipped) |
| SSH keys | Lists keys to generate with `ssh-keygen` commands |
| Git identity | Sets global user.name/email via `git config --global` |
| Git profiles | Creates directory, writes profile config, adds includeIf to `~/.gitconfig` |

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--dry-run` | `false` | Preview changes without applying |

**Safety:**
- Existing SSH hosts are never overwritten
- Existing git profile configs are never overwritten
- SSH private keys are never copied — only key generation commands are shown
- `~/.ssh/config` is appended to, not replaced

<br/>

## Doctor (Cross-Module)

Run all diagnostic checks in a single command — combines SSH audit, Git doctor, and Env check.

```bash
# Full system diagnostics
bash-pilot doctor

# JSON output
bash-pilot doctor -o json
```

**Sections reported:**
- SSH: shared keys, file permissions, missing keys
- Git: duplicate safe.directory, missing includeIf targets, permissions
- Env: shell, tools, SSH agent, git config, editor, home directory

Ends with a summary: `Total: N issue(s) — SSH: X, Git: Y, Env: Z`

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
│   github.com-personal    github.com       user1     id_rsa_personal
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
