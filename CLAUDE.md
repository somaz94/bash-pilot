# CLAUDE.md - bash-pilot

A powerful CLI toolkit for bash power users — SSH management, Git multi-profile, environment health checks, and smart prompt.

## Build & Test

```bash
make build           # Build binary
make test            # Run unit tests
make test-unit       # go test ./... -v -race -cover
make cover           # Generate coverage report
make cover-html      # Open coverage in browser
make fmt             # go fmt
make vet             # go vet
make install         # Install to /usr/local/bin
```

- Do not include `Co-Authored-By` lines in commit messages.
- Use Conventional Commits (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `ci:`, `chore:`)
- Do not push to remote. Only commit. The user will push manually.
- Do not modify git config.

## Key Concepts

- **SSH Module**: Parses `~/.ssh/config`, lists hosts with grouping, runs parallel connectivity checks, audits security issues
- **Git Module**: Manages multi-profile git identities (user/email/SSH key per directory), detects and cleans gitconfig issues (duplicate safe.directory, etc.)
- **Env Module**: Scans shell config (PATH duplicates, missing dirs), dotfile health checks, slow-loading source detection
- **Prompt Module**: Lightweight bash prompt with git branch, SSH host context, k8s context display
- **Config**: YAML-based configuration at `~/.config/bash-pilot/config.yaml`

## CLI Commands

| Command | Description |
|---------|-------------|
| `bash-pilot ssh list` | List SSH hosts with grouping and status |
| `bash-pilot ssh ping` | Test connectivity to SSH hosts (parallel) |
| `bash-pilot ssh audit` | Audit SSH config for security issues |
| `bash-pilot git profiles` | List configured git profiles |
| `bash-pilot git doctor` | Diagnose gitconfig issues |
| `bash-pilot git clean` | Clean up duplicate/stale gitconfig entries |
| `bash-pilot env check` | Scan shell environment for issues |
| `bash-pilot env path` | Analyze PATH for duplicates and missing dirs |
| `bash-pilot prompt init` | Output bash prompt configuration |
| `bash-pilot version` | Show version info |

## Global Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--config` | `-c` | `~/.config/bash-pilot/config.yaml` | Config file path |
| `--output` | `-o` | `color` | Output format: `color`, `plain`, `json`, `table` |
| `--no-color` | | `false` | Disable color output |

## Project Structure

```
cmd/
  main.go              # Entry point
  cli/
    root.go            # Cobra root command + global flags
    ssh.go             # SSH subcommands (list, ping, audit)
    git.go             # Git subcommands (profiles, doctor, clean)
    env.go             # Env subcommands (check, path)
    prompt.go          # Prompt subcommand (init)
    version.go         # Version subcommand
internal/
  ssh/
    parser.go          # SSH config parser (~/.ssh/config)
    host.go            # Host data model and grouping
    ping.go            # Parallel SSH connectivity checker
    audit.go           # SSH security auditor
  git/
    profile.go         # Git profile manager (includeIF, conditional configs)
    doctor.go          # Gitconfig issue detector
    clean.go           # Gitconfig cleanup (safe.directory dedup, etc.)
  env/
    scanner.go         # Shell environment scanner
    path.go            # PATH analyzer
  prompt/
    prompt.go          # Bash prompt generator
  config/
    config.go          # YAML config loader
  report/
    output.go          # Shared output formatters (color/plain/json/table)
scripts/
  install.sh           # Installation script
```

## Workflow After Code Changes

After modifying any code, always follow this order:

1. **Tests first** — Write or update tests for the changed code. Run `make test` and ensure all tests pass.
2. **Documentation second** — Update the relevant documentation:
   - `README.md` — Quick Start, feature list, usage examples
   - `CLAUDE.md` — Key Concepts, CLI Commands table, Project Structure

Never skip tests or leave them for later. Every code change must have corresponding test coverage before documentation is updated.

- Communicate with the user in Korean.
- All documentation and code comments must be written in English.
