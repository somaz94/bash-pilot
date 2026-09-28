# CLAUDE.md - bash-pilot

A powerful CLI toolkit for bash power users — SSH management, Git multi-profile, environment health checks, and smart prompt.

## Build & Test

```bash
make build           # Build binary
make test            # Run unit tests
make test-unit       # go test ./... -v -race -cover
make cover           # Generate coverage report
make cover-html      # Open coverage in browser
make lint            # golangci-lint (config in .golangci.yml)
make fmt             # go fmt
make vet             # go vet
make install         # Install to /usr/local/bin
```

## Key Concepts

- **SSH Module**: Parses `~/.ssh/config`, lists hosts with grouping, runs parallel connectivity checks, audits security issues
- **Git Module**: Manages multi-profile git identities (email and signing key per includeIf directory), detects and cleans gitconfig issues (duplicate safe.directory, etc.)
- **Env Module**: Checks shell, common tools, ssh-agent, git identity, home dotfiles, and editor; analyzes PATH for duplicates and missing dirs
- **Prompt Module**: Lightweight bash prompt with user@host, git branch, and k8s context display
- **Snapshot Module**: Captures system, tools, git, SSH, k8s, and brew state to JSON; `diff` compares a snapshot to the current machine, `setup` installs what is missing
- **Migrate Module**: Exports SSH hosts and git profiles to portable JSON and imports them on another machine
- **Config**: YAML-based configuration at `~/.config/bash-pilot/config.yaml`
- **Report**: Shared output formatters (color/plain/json/table) and severity rendering

## CLI Commands

| Command | Description |
|---------|-------------|
| `bash-pilot ssh list` | List SSH hosts with grouping |
| `bash-pilot ssh ping` | Test connectivity to SSH hosts (parallel) |
| `bash-pilot ssh audit` | Audit SSH config for security issues |
| `bash-pilot git profiles` | List configured git profiles |
| `bash-pilot git doctor` | Diagnose gitconfig issues |
| `bash-pilot git clean` | Clean up duplicate/stale gitconfig entries |
| `bash-pilot env check` | Scan shell environment for issues |
| `bash-pilot env path` | Analyze PATH for duplicates and missing dirs |
| `bash-pilot prompt init` | Output bash prompt configuration |
| `bash-pilot prompt show` | Preview prompt components for current environment |
| `bash-pilot init` | Generate config from existing SSH config (`--force` overwrites) |
| `bash-pilot doctor` | Full system diagnostics (SSH + Git + Env) |
| `bash-pilot snapshot` | Capture environment snapshot |
| `bash-pilot diff <snapshot-file>` | Compare snapshot against current environment |
| `bash-pilot setup <snapshot-file>` | Install missing tools from a snapshot |
| `bash-pilot migrate export` | Export SSH + Git config to portable JSON |
| `bash-pilot migrate import <config-file>` | Import SSH + Git config from portable JSON |
| `bash-pilot version` | Show version info |

## Global Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--config` | `-c` | `~/.config/bash-pilot/config.yaml` | Config file path |
| `--output` | `-o` | `color` | Output format: `color`, `plain`, `json`, `table` |
| `--no-color` | | `false` | Disable color output |

## Project Structure

`cmd/main.go` is the entry point. `cmd/cli/` holds one file per top-level Cobra command, and the file name is the command (`root.go` holds the root command and global flags; `diff` shares `snapshot.go` with `snapshot`). `internal/<module>/` holds the logic for each module named in Key Concepts, with tests beside the code, and `internal/testutil/` holds shared test helpers. `scripts/` holds the install, demo, and helper scripts.

List the directories to see the current file set; this section deliberately does not copy it.

## Workflow After Code Changes

After modifying any code, always follow this order:

1. **Tests first** — Write or update tests for the changed code. Run `make test` and ensure all tests pass.
2. **Documentation second** — Update the relevant documentation:
   - `README.md` — Quick Start, feature list, usage examples
   - `CLAUDE.md` — Key Concepts, CLI Commands table, Project Structure (only when a new top-level directory or module is added)

Never skip tests or leave them for later. Every code change must have corresponding test coverage before documentation is updated.
