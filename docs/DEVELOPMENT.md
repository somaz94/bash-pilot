# Development

Guide for building, testing, and contributing to bash-pilot.

<br/>

## Table of Contents

- [Prerequisites](#prerequisites)
- [Project Structure](#project-structure)
- [Build](#build)
- [Testing](#testing)
- [Demo](#demo)
- [CI/CD Workflows](#cicd-workflows)
- [Conventions](#conventions)

<br/>

## Prerequisites

- Go 1.26+
- Make

<br/>

## Project Structure

```
.
├── cmd/
│   ├── main.go                    # Entry point
│   └── cli/
│       ├── root.go                # Root command with global flags
│       ├── ssh.go                 # SSH subcommands (list, ping, audit)
│       ├── git.go                 # Git subcommands (profiles, doctor, clean)
│       ├── env.go                 # Env subcommands (check, path)
│       ├── prompt.go              # Prompt subcommands (init, show)
│       ├── doctor.go              # Cross-module diagnostics
│       ├── snapshot.go            # Snapshot & diff subcommands
│       ├── setup.go               # Setup subcommand (install missing tools)
│       ├── migrate.go             # Migrate subcommands (export, import)
│       ├── init.go                # Init command (auto-generate config)
│       └── version.go             # Version subcommand
├── internal/
│   ├── ssh/
│   │   ├── parser.go             # SSH config parser (~/.ssh/config)
│   │   ├── parser_test.go
│   │   ├── host.go               # Host data model and grouping
│   │   ├── host_test.go
│   │   ├── ping.go               # Parallel SSH connectivity checker
│   │   ├── ping_test.go
│   │   ├── audit.go              # SSH security auditor
│   │   └── audit_test.go
│   ├── git/
│   │   ├── gitconfig.go           # Gitconfig parser, profiles, doctor, clean
│   │   └── gitconfig_test.go
│   ├── env/
│   │   ├── env.go                 # Environment check, PATH analysis
│   │   └── env_test.go
│   ├── prompt/
│   │   ├── prompt.go              # Prompt script generation
│   │   ├── helpers.go             # Git branch, k8s context helpers
│   │   ├── prompt_test.go
│   │   └── helpers_test.go
│   ├── snapshot/
│   │   ├── snapshot.go            # Environment capture (tools, git, SSH, k8s, brew)
│   │   ├── snapshot_test.go
│   │   ├── diff.go                # Snapshot comparison logic
│   │   ├── diff_test.go
│   │   ├── setup.go               # Install missing tools from snapshot
│   │   └── setup_test.go
│   ├── migrate/
│   │   ├── types.go               # Portable migration data structures
│   │   ├── export.go              # Export SSH + Git config
│   │   ├── export_test.go
│   │   ├── import.go              # Import with path translation
│   │   └── import_test.go
│   ├── config/
│   │   ├── config.go             # YAML config loader
│   │   └── config_test.go
│   └── report/
│       ├── output.go             # Output formatters (color/plain/json/table)
│       └── output_test.go
├── scripts/
│   ├── demo.sh                   # Demo script
│   ├── demo-clean.sh             # Demo cleanup
│   └── install.sh                # curl installer
├── docs/                         # Documentation
├── .github/
│   ├── workflows/                # CI/CD workflows
│   ├── dependabot.yml            # Dependency updates
│   └── release.yml               # Release note categories
├── .goreleaser.yml               # Multi-platform build + Homebrew + Scoop
├── Makefile                      # Build, test, demo
└── go.mod
```

<br/>

### Key Directories

| Directory | Description |
|-----------|-------------|
| `cmd/cli/` | Cobra CLI commands and flag definitions |
| `internal/ssh/` | SSH config parsing, host grouping, connectivity testing, security audit |
| `internal/git/` | Gitconfig parsing, multi-profile management, diagnostics, cleanup |
| `internal/env/` | Shell environment health check, PATH analysis |
| `internal/prompt/` | Smart prompt generation (git branch, k8s context) |
| `internal/snapshot/` | Environment snapshot capture and diff comparison |
| `internal/migrate/` | Cross-machine SSH + Git config migration |
| `internal/config/` | YAML configuration loader with defaults |
| `internal/report/` | Output formatting (color, plain, JSON, table) |

<br/>

## Build

```bash
make build           # Build binary → ./bin/bash-pilot
make clean           # Remove build artifacts
make install         # Install to /usr/local/bin
```

<br/>

## Testing

```bash
make test            # Run unit tests (alias)
make test-unit       # go test ./... -v -race -cover
make cover           # Generate coverage report
make cover-html      # Open coverage report in browser
```

<br/>

### Test Coverage

| Package | Coverage |
|---------|----------|
| `internal/ssh` | 96.1% |
| `internal/git` | 94.2% |
| `internal/env` | 99.2% |
| `internal/prompt` | 99.0% |
| `internal/snapshot` | 95.0% |
| `internal/migrate` | 91.7% |
| `internal/config` | 82.4% |
| `internal/report` | 100% |

<br/>

## Demo

```bash
make demo            # Run demo (creates temp SSH config, tests all commands)
make demo-clean      # Clean up demo resources
make demo-all        # Run demo and clean up automatically
```

<br/>

## Workflow

```bash
make check-gh        # Verify gh CLI is installed and authenticated
make branch name=git-module   # Create feature branch from main
make pr title="feat: add git module"   # Test → push → create PR (auto-generates body)
```

`make pr` automatically:
1. Runs `go test ./... -race -cover` and `go vet`
2. Pushes the branch to origin
3. Generates PR body from commit history (categorized by `feat:`, `fix:`, `test:`, `docs:`)
4. Detects changed test packages and builds a test plan checklist
5. Creates the PR via `gh pr create`

<br/>

## CI/CD Workflows

| Workflow | Trigger | Description |
|----------|---------|-------------|
| `ci.yml` | push (main), PR, dispatch | Unit tests → Build → Version verify |
| `release.yml` | tag push `v*` | GoReleaser (binaries + Homebrew + Scoop) |
| `gitlab-mirror.yml` | push to main | Mirror to GitLab |
| `changelog-generator.yml` | after release, PR merge | Auto-generate CHANGELOG.md |
| `contributors.yml` | after changelog | Auto-generate CONTRIBUTORS.md |
| `stale-issues.yml` | daily cron | Auto-close stale issues |
| `dependabot-auto-merge.yml` | PR (dependabot) | Auto-merge minor/patch updates |
| `issue-greeting.yml` | issue opened | Welcome message |

<br/>

### Workflow Chain

```
tag push v* → Create release (GoReleaser)
                └→ Generate changelog
                      └→ Generate Contributors
```

<br/>

## Conventions

- **Commits**: Conventional Commits (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `ci:`, `chore:`)
- **Secrets**: `PAT_TOKEN` (cross-repo ops), `GITHUB_TOKEN` (releases)
- **paths-ignore**: `.github/workflows/**`, `**/*.md`
