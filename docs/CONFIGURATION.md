# Configuration

Reference for all bash-pilot configuration options.

<br/>

## Table of Contents

- [Config File](#config-file)
- [SSH Section](#ssh-section)
- [Git Section](#git-section)
- [CLI Flags](#cli-flags)
- [Auto-Detection Rules](#auto-detection-rules)

<br/>

## Config File

Default location: `~/.config/bash-pilot/config.yaml`

Override with `--config` or `-c` flag:

```bash
bash-pilot ssh list --config /path/to/config.yaml
```

<br/>

### Full Example

```yaml
ssh:
  config_file: ~/.ssh/config   # Optional: override SSH config path
  groups:
    git:
      pattern: ["github.com*", "gitlab*", "git-codecommit*"]
    cloud:
      pattern: ["staging-*", "prod-*"]
      label: "AWS Seoul"
    k8s:
      pattern: ["k8s-*"]
    on-prem:
      pattern: ["server*", "nas*", "mac-mini"]
  ping:
    timeout: 5s
    parallel: 10

git:
  profiles:
    work:
      directory: ~/gitlab-project
      email: user@company.com
      key: ~/.ssh/id_rsa_work
    personal:
      directory: ~/PrivateWork
      email: user@gmail.com
      key: ~/.ssh/id_rsa_personal
```

<br/>

## SSH Section

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ssh.config_file` | string | `~/.ssh/config` | Path to SSH config file |
| `ssh.groups.<name>.pattern` | []string | — | Glob patterns to match host names |
| `ssh.groups.<name>.label` | string | — | Display label for the group |
| `ssh.ping.timeout` | duration | `5s` | TCP connection timeout per host |
| `ssh.ping.parallel` | int | `10` | Maximum concurrent ping goroutines |

<br/>

## Git Section

The git module reads profiles directly from `~/.gitconfig` `includeIf` directives. No additional config is needed in `config.yaml`.

### CLI Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--gitconfig` | string | auto-detect | Path to gitconfig file |
| `--dry-run` | bool | `false` | Preview changes without modifying (clean only) |

### How Profiles Are Detected

bash-pilot reads `~/.gitconfig` and finds `[includeIf "gitdir:..."]` sections:

```gitconfig
[includeIf "gitdir:~/work/"]
    path = ~/.gitconfig-work

[includeIf "gitdir:~/personal/"]
    path = ~/.gitconfig-personal
```

Each included config provides the email and optional signing key for that directory.

<br/>

## CLI Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--config` | `-c` | string | `~/.config/bash-pilot/config.yaml` | Config file path |
| `--output` | `-o` | string | `color` | Output format |
| `--no-color` | | bool | `false` | Disable color output |

<br/>

## Auto-Detection Rules

When hosts don't match any configured group pattern, bash-pilot auto-detects the group:

| Rule | Group | Detection Method |
|------|-------|-----------------|
| Git hosts | `git` | Name contains `github`, `gitlab`, `codecommit`, `bitbucket` |
| Kubernetes | `k8s` | Name starts with `k8s-` or contains `kube`, `master`, `node` |
| Cloud | `cloud` | Public IP address, or hostname contains `amazonaws.com`, `compute.google`, `azure` |
| On-premise | `on-prem` | Private IP ranges: `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16` |
| Other | `other` | No match found |
