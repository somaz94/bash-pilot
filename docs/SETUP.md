# Initial Setup Guide

Step-by-step guide to set up bash-pilot with a real-world SSH environment.

<br/>

## Table of Contents

- [Install](#install)
- [Create Config File](#create-config-file)
- [Verify Setup](#verify-setup)
- [Example: Real Environment](#example-real-environment)
- [Troubleshooting](#troubleshooting)

<br/>

## Install

```bash
# Homebrew (recommended)
brew install somaz94/tap/bash-pilot

# Or via curl
curl -sSL https://raw.githubusercontent.com/somaz94/bash-pilot/main/scripts/install.sh | bash

# Or via Go
go install github.com/somaz94/bash-pilot/cmd@latest
```

Verify installation:

```bash
bash-pilot version
```

<br/>

## Create Config File

Create the config directory and file:

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
        - "test-server"
        - "jenkins"
      label: "AWS Frankfurt"

    k8s:
      pattern:
        - "k8s-*"
      label: "Kubernetes Cluster"

    on-prem:
      pattern:
        - "nas*"
        - "server*"
        - "mac-mini"
        - "projectm-svn"
        - "openclaw"
      label: "On-Premise Servers"

  ping:
    timeout: 5s
    parallel: 10
```

### What Each Group Does

| Group | Pattern | Matches |
|-------|---------|---------|
| `git` | `github.com*`, `git-codecommit*`, `gitlab` | GitHub accounts, CodeCommit, self-hosted GitLab |
| `cloud` | `test-server`, `jenkins` | AWS EC2 instances (public IP) |
| `k8s` | `k8s-*` | k8s-control-01, k8s-compute-01/02/03 |
| `on-prem` | `nas*`, `server*`, `mac-mini`, ... | NAS, internal servers, build machines |

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
#   github.com-somaz94        github.com               somaz         id_rsa_somaz94
#   github.com-somaz940829    github.com               somaz-devops  id_rsa_somaz940829
#   git-codecommit...         git-codecommit...        APKA...       id_rsa_codecommit
#
# AWS Frankfurt (2 hosts)
#   test-server               3.65.182.184             ec2-user      frankfurt-habby-1704.pem
#   jenkins                   18.159.54.27             ec2-user      frankfurt-habby-1704.pem
#
# Kubernetes Cluster (4 hosts)
#   k8s-control-01            10.10.10.17              concrit       id_rsa_concrit
#   k8s-compute-01            10.10.10.18              concrit       id_rsa_concrit
#   ...
#
# On-Premise Servers (8 hosts)
#   nas                       10.10.10.5               somaz         id_rsa_concrit
#   server1                   10.10.10.10              concrit       id_rsa_concrit
#   ...
```

### 2. Test connectivity

```bash
# Ping all hosts
bash-pilot ssh ping

# Ping only Kubernetes nodes
bash-pilot ssh ping k8s-*

# Ping only cloud instances
bash-pilot ssh ping test-server jenkins
```

### 3. Security audit

```bash
bash-pilot ssh audit

# Expected warnings:
# WARN  Shared key: id_rsa_concrit used by 12 hosts
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

## Example: Real Environment

A typical SSH environment might look like this:

```
~/.ssh/config
├── Git (3 hosts)
│   ├── github.com-somaz94        → github.com (personal)
│   ├── github.com-somaz940829    → github.com (work)
│   └── git-codecommit.*          → AWS CodeCommit
│
├── Cloud (2 hosts)
│   ├── test-server               → 3.65.182.184 (AWS Frankfurt)
│   └── jenkins                   → 18.159.54.27 (AWS Frankfurt)
│
├── Kubernetes (4 hosts)
│   ├── k8s-control-01            → 10.10.10.17
│   ├── k8s-compute-01            → 10.10.10.18
│   ├── k8s-compute-02            → 10.10.10.19
│   └── k8s-compute-03            → 10.10.10.22
│
└── On-Premise (8 hosts)
    ├── nas                       → 10.10.10.5
    ├── server1~4                 → 10.10.10.10~15
    ├── gitlab                    → 10.10.10.60
    ├── mac-mini                  → 10.10.10.80
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
