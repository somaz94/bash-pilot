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
make pr title="feat: add git module"   # Test → push → create PR
```

<br/>

## CI/CD Workflows

| Workflow | Trigger | Description |
|----------|---------|-------------|
| `ci.yml` | push, PR, dispatch | Unit tests → Build → Version verify |
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
