# bash-pilot

[![CI](https://github.com/somaz94/bash-pilot/actions/workflows/ci.yml/badge.svg)](https://github.com/somaz94/bash-pilot/actions/workflows/ci.yml)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Latest Tag](https://img.shields.io/github/v/tag/somaz94/bash-pilot)](https://github.com/somaz94/bash-pilot/tags)
[![Top Language](https://img.shields.io/github/languages/top/somaz94/bash-pilot)](https://github.com/somaz94/bash-pilot)

A powerful CLI toolkit for bash power users — SSH management, Git multi-profile, environment health checks, and smart prompt.

> For detailed documentation, see the [docs/](docs/) folder:
>
> [Setup](docs/SETUP.md) |
> [Usage](docs/USAGE.md) |
> [Configuration](docs/CONFIGURATION.md) |
> [Examples](docs/EXAMPLES.md) |
> [Deployment](docs/DEPLOYMENT.md) |
> [Development](docs/DEVELOPMENT.md) |
> [Use Cases](docs/USE-CASES.md)

<br/>

## Features

<br/>

### SSH Module
- **`bash-pilot ssh list`** — List SSH hosts with automatic grouping (git, cloud, k8s, on-prem)
- **`bash-pilot ssh ping [pattern]`** — Parallel connectivity testing with latency measurement
- **`bash-pilot ssh audit`** — Security audit (shared keys, file permissions, missing keys)

<br/>

### Git Module (coming soon)
- **`bash-pilot git profiles`** — Multi-profile git identity management
- **`bash-pilot git doctor`** — Diagnose gitconfig issues (duplicate safe.directory, etc.)
- **`bash-pilot git clean`** — Clean up stale/duplicate gitconfig entries

<br/>

### Env Module (coming soon)
- **`bash-pilot env check`** — Shell environment health scan
- **`bash-pilot env path`** — PATH analysis (duplicates, missing directories)

<br/>

### Prompt Module (coming soon)
- **`bash-pilot prompt init`** — Smart bash prompt with git branch and k8s context

<br/>

## Quick Start

<br/>

### Install

```bash
# Homebrew
brew install somaz94/tap/bash-pilot

# Scoop (Windows)
scoop bucket add somaz94 https://github.com/somaz94/scoop-bucket
scoop install bash-pilot

# curl (latest)
curl -sSL https://raw.githubusercontent.com/somaz94/bash-pilot/main/scripts/install.sh | bash

# curl (specific version)
curl -sL https://github.com/somaz94/bash-pilot/releases/download/v0.1.0/bash-pilot_0.1.0_linux_amd64.tar.gz | tar xz
sudo mv bash-pilot /usr/local/bin/

# Go install
go install github.com/somaz94/bash-pilot/cmd@latest

# From source
git clone https://github.com/somaz94/bash-pilot.git
cd bash-pilot && make install
```

<br/>

### Upgrade

```bash
# Homebrew
brew update && brew upgrade bash-pilot

# Scoop
scoop update bash-pilot

# curl (re-run installer)
curl -sSL https://raw.githubusercontent.com/somaz94/bash-pilot/main/scripts/install.sh | bash

# Go install
go install github.com/somaz94/bash-pilot/cmd@latest
```

<br/>

### Uninstall

```bash
# Homebrew
brew uninstall bash-pilot

# Scoop
scoop uninstall bash-pilot

# Manual
sudo rm /usr/local/bin/bash-pilot

# Optional: remove config
rm -rf ~/.config/bash-pilot
```

<br/>

### Basic Usage

```bash
# Auto-generate config from existing ~/.ssh/config
bash-pilot init

# List SSH hosts grouped by type
bash-pilot ssh list

# Test connectivity to all hosts
bash-pilot ssh ping

# Test only k8s hosts
bash-pilot ssh ping "k8s-*"

# Security audit
bash-pilot ssh audit

# JSON output
bash-pilot ssh list -o json
```

<br/>

## Configuration

Auto-generate from your SSH config, or create manually:

```bash
# Auto-generate (recommended)
bash-pilot init

# Or create manually
mkdir -p ~/.config/bash-pilot
```

Config file at `~/.config/bash-pilot/config.yaml`:

```yaml
ssh:
  groups:
    git:
      pattern: ["github.com*", "gitlab*", "git-codecommit*"]
    cloud:
      pattern: ["test-server", "jenkins"]
      label: "AWS Frankfurt"
    k8s:
      pattern: ["k8s-*"]
    on-prem:
      pattern: ["server*", "nas*", "mac-mini"]
  ping:
    timeout: 5s
    parallel: 10
```

<br/>

## Global Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--config` | `-c` | `~/.config/bash-pilot/config.yaml` | Config file path |
| `--output` | `-o` | `color` | Output format: `color`, `plain`, `json`, `table` |
| `--no-color` | | `false` | Disable color output |

<br/>

## Development

```bash
make build        # Build binary
make test         # Run tests with race detection
make cover        # Generate coverage report
make demo         # Run demo
make demo-all     # Run demo and clean up
make help         # Show all targets
```

<br/>

## License

Apache License 2.0 — see [LICENSE](LICENSE) for details.
