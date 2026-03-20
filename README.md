# bash-pilot

A powerful CLI toolkit for bash power users — SSH management, Git multi-profile, environment health checks, and smart prompt.

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

## Installation

<br/>

### Homebrew

```bash
brew tap somaz94/tap
brew install bash-pilot
```

<br/>

### From source

```bash
git clone https://github.com/somaz94/bash-pilot.git
cd bash-pilot
make install
```

<br/>

### Quick build

```bash
make build
./bin/bash-pilot version
```

<br/>

## Quick Start

```bash
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

Create `~/.config/bash-pilot/config.yaml`:

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
make fmt          # Format code
make vet          # Run go vet
make test         # Run tests with race detection
make cover        # Generate coverage report
make cover-html   # Open coverage in browser
```

<br/>

## License

Apache License 2.0 — see [LICENSE](LICENSE) for details.
