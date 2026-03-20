# Initial Setup Guide

Step-by-step guide to set up bash-pilot with a real-world SSH environment.

<br/>

## Table of Contents

- [Install](#install)
- [Create Config File](#create-config-file)
- [Verify Setup](#verify-setup)
- [Example: Typical Environment](#example-typical-environment)
- [Troubleshooting](#troubleshooting)

<br/>

## Install

```bash
# Homebrew (recommended)
brew install somaz94/tap/bash-pilot

# Or via curl (latest)
curl -sSL https://raw.githubusercontent.com/somaz94/bash-pilot/main/scripts/install.sh | bash

# Or via curl (specific version)
curl -sL https://github.com/somaz94/bash-pilot/releases/download/v0.1.0/bash-pilot_0.1.0_linux_amd64.tar.gz | tar xz
sudo mv bash-pilot /usr/local/bin/

# Or via Go
go install github.com/somaz94/bash-pilot/cmd@latest
```

Verify installation:

```bash
bash-pilot version
```

<br/>

## Create Config File

The easiest way is to auto-generate from your existing SSH config:

```bash
# Auto-generate config from ~/.ssh/config
bash-pilot init

# Overwrite existing config
bash-pilot init --force
```

This analyzes your `~/.ssh/config`, auto-detects host groups (git, cloud, k8s, on-prem), and writes the result to `~/.config/bash-pilot/config.yaml`.

### Manual Setup

Alternatively, create the config manually:

```bash
mkdir -p ~/.config/bash-pilot
```

Write your config at `~/.config/bash-pilot/config.yaml`.

Below is an example based on a typical mixed environment (GitHub personal/work accounts, AWS cloud instances, on-prem servers, and a Kubernetes cluster):

```yaml
ssh:
  # Optional: defaults to ~/.ssh/config
  # config_file: ~/.ssh/config

  groups:
    git:
      pattern:
        - "github.com*"
        - "git-codecommit*"
        - "gitlab"
      label: "Git Services"

    cloud:
      pattern:
        - "web-server"
        - "ci-server"
      label: "AWS Cloud"

    k8s:
      pattern:
        - "k8s-*"
      label: "Kubernetes Cluster"

    on-prem:
      pattern:
        - "nas*"
        - "server*"
        - "build-machine"
      label: "On-Premise Servers"

  ping:
    timeout: 5s
    parallel: 10
```

### What Each Group Does

| Group | Pattern | Matches |
|-------|---------|---------|
| `git` | `github.com*`, `git-codecommit*`, `gitlab` | GitHub accounts, CodeCommit, self-hosted GitLab |
| `cloud` | `web-server`, `ci-server` | AWS EC2 instances (public IP) |
| `k8s` | `k8s-*` | k8s-control-01, k8s-worker-01/02/03 |
| `on-prem` | `nas*`, `server*`, `build-machine`, ... | NAS, internal servers, build machines |

> **Tip:** Hosts not matching any pattern are auto-detected by IP range and hostname. See [Auto-Detection Rules](CONFIGURATION.md#auto-detection-rules) for details.

<br/>

## Verify Setup

Run these commands to verify everything works:

### 1. List hosts

```bash
# Color output (default)
bash-pilot ssh list

# Expected output:
# Git Services (3 hosts)
#   github.com-personal       github.com               user1         id_rsa_personal
#   github.com-work           github.com               user2         id_rsa_work
#   git-codecommit...         git-codecommit...        AKID...       id_rsa_codecommit
#
# AWS Cloud (2 hosts)
#   web-server                54.123.45.67             ec2-user      my-region.pem
#   ci-server                 54.123.45.68             ec2-user      my-region.pem
#
# Kubernetes Cluster (4 hosts)
#   k8s-control-01            10.0.1.10                admin         id_rsa_infra
#   k8s-worker-01             10.0.1.11                admin         id_rsa_infra
#   ...
#
# On-Premise Servers (5 hosts)
#   nas                       192.168.1.10             user          id_rsa_office
#   server1                   192.168.1.20             admin         id_rsa_office
#   ...
```

### 2. Test connectivity

```bash
# Ping all hosts
bash-pilot ssh ping

# Ping only Kubernetes nodes
bash-pilot ssh ping k8s-*

# Ping only cloud instances
bash-pilot ssh ping web-server ci-server
```

### 3. Security audit

```bash
bash-pilot ssh audit

# Expected warnings:
# WARN  Shared key: id_rsa_office used by 8 hosts
# WARN  Key permission too open: ~/.ssh/some_key (0644, want 0600)
```

### 4. JSON output for scripting

```bash
# Export host inventory
bash-pilot ssh list -o json | jq '.[] | .name'

# Get unreachable hosts
bash-pilot ssh ping -o json | jq '.[] | select(.reachable == false) | .name'
```

<br/>

## Example: Typical Environment

A typical SSH environment might look like this:

```
~/.ssh/config
├── Git (3 hosts)
│   ├── github.com-personal      → github.com (personal)
│   ├── github.com-work          → github.com (work)
│   └── git-codecommit.*         → AWS CodeCommit
│
├── Cloud (2 hosts)
│   ├── web-server               → 54.123.45.67 (AWS)
│   └── ci-server                → 54.123.45.68 (AWS)
│
├── Kubernetes (4 hosts)
│   ├── k8s-control-01           → 10.0.1.10
│   ├── k8s-worker-01            → 10.0.1.11
│   ├── k8s-worker-02            → 10.0.1.12
│   └── k8s-worker-03            → 10.0.1.13
│
└── On-Premise (5 hosts)
    ├── nas                      → 192.168.1.10
    ├── server1~3                → 192.168.1.20~40
    ├── gitlab                   → 192.168.1.50
    └── ...
```

Without any config, bash-pilot auto-detects these groups by analyzing hostnames and IP ranges. The config file lets you customize group names, labels, and patterns for your specific setup.

<br/>

## Troubleshooting

### Config not loading

```bash
# Check config path
bash-pilot ssh list --config ~/.config/bash-pilot/config.yaml

# Validate YAML syntax
cat ~/.config/bash-pilot/config.yaml | python3 -c "import sys,yaml; yaml.safe_load(sys.stdin)"
```

### Hosts not appearing

- Ensure `~/.ssh/config` exists and has valid `Host` blocks
- Wildcard-only entries (`Host *`) are skipped by design
- Check that host names match your group patterns

### Ping timeouts

```bash
# Increase timeout for slow networks
# In config.yaml:
ssh:
  ping:
    timeout: 10s
```

### Permission warnings in audit

```bash
# Fix key permissions
chmod 600 ~/.ssh/id_rsa_*
chmod 600 ~/.ssh/*.pem
```
