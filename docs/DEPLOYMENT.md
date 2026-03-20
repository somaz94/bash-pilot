# Deployment

Guide for releasing and distributing bash-pilot.

<br/>

## Release Flow

A single tag push triggers the entire release pipeline automatically:

```
git tag v1.0.0 && git push origin v1.0.0
    └→ GitHub Actions (release.yml)
        └→ GoReleaser
            ├→ GitHub Releases (linux/darwin/windows x amd64/arm64)
            ├→ Homebrew tap update (somaz94/homebrew-tap)
            └→ Scoop bucket update (somaz94/scoop-bucket)
```

<br/>

## Distribution Channels

<br/>

### 1. GitHub Releases (Default)

Automatically built by GoReleaser.

```bash
# Install latest
curl -sSL https://raw.githubusercontent.com/somaz94/bash-pilot/main/scripts/install.sh | bash

# Install specific version
curl -sL https://github.com/somaz94/bash-pilot/releases/download/v0.1.0/bash-pilot_0.1.0_linux_amd64.tar.gz | tar xz
sudo mv bash-pilot /usr/local/bin/
```

**Available archive naming pattern:** `bash-pilot_{version}_{os}_{arch}.tar.gz`

**Supported platforms:**

| OS | Architecture |
|----|-------------|
| Linux | amd64, arm64 |
| macOS (Darwin) | amd64, arm64 |
| Windows | amd64, arm64 |

<br/>

### 2. Homebrew (macOS / Linux)

```bash
brew install somaz94/tap/bash-pilot
```

<br/>

### 3. Go Install

```bash
go install github.com/somaz94/bash-pilot/cmd@latest
```

<br/>

## Secrets Configuration

| Secret | Purpose | Scope |
|--------|---------|-------|
| `PAT_TOKEN` | Cross-repo write access (Homebrew tap) | GoReleaser |
| `GITHUB_TOKEN` | Release creation, dependabot auto-merge | Auto-provided |

<br/>

## Step-by-Step: First Release

1. Ensure all tests pass:
   ```bash
   make test
   make build
   ```

2. Create and push a tag:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```

3. GitHub Actions triggers `release.yml` → GoReleaser runs automatically

4. Verify:
   - Check [GitHub Releases](https://github.com/somaz94/bash-pilot/releases) for binaries
   - Check `somaz94/homebrew-tap` for formula commit
