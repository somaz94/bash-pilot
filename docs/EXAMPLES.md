# Examples

Hands-on examples for bash-pilot.

<br/>

## Table of Contents

- [Quick Demo](#quick-demo)
- [SSH List](#ssh-list)
- [SSH Ping](#ssh-ping)
- [SSH Audit](#ssh-audit)
- [Git Profiles](#git-profiles)
- [Git Doctor](#git-doctor)
- [Git Clean](#git-clean)
- [Env Check](#env-check)
- [Env Path](#env-path)
- [Prompt Init](#prompt-init)
- [Prompt Show](#prompt-show)
- [Snapshot](#snapshot)
- [Diff](#diff)
- [Setup](#setup)
- [Migrate Export](#migrate-export)
- [Migrate Import](#migrate-import)
- [Doctor](#doctor)
- [Scripting with JSON](#scripting-with-json)

<br/>

## Init

### Auto-generate config

```bash
$ bash-pilot init
Config generated: /home/user/.config/bash-pilot/config.yaml
Detected 4 groups from 15 hosts:
  git        3 hosts
  cloud      2 hosts
  k8s        4 hosts
  on-prem    6 hosts

Edit the config to customize group patterns and labels.
```

### Preview without overwriting

```bash
$ bash-pilot init
Config already exists: /home/user/.config/bash-pilot/config.yaml
Use --force to overwrite.

Generated config (preview):
---
ssh:
  groups:
    git:
      pattern:
        - github.com-personal
        - github.com-work
    cloud:
      pattern:
        - web-server
        - ci-server
  ...
```

<br/>

## Quick Demo

Run the built-in demo to see all features:

```bash
make demo          # Run demo
make demo-clean    # Clean up
make demo-all      # Run demo and clean up automatically
```

<br/>

## SSH List

<br/>

### Basic grouping

```bash
$ bash-pilot ssh list
┌─ GIT ──────────────────────────────────────────────
│   github.com-personal     github.com           user1           id_rsa_personal
│   github.com-work         github.com           user2           id_rsa_work
└────────────────────────────────────────────────────

┌─ CLOUD ────────────────────────────────────────────
│   web-server              54.123.45.67         ec2-user        my-region.pem
│   ci-server               54.123.45.68         ec2-user        my-region.pem
└────────────────────────────────────────────────────

┌─ K8S ──────────────────────────────────────────────
│   k8s-control-01          10.0.1.10            admin           id_rsa_infra
│   k8s-worker-01           10.0.1.11            admin           id_rsa_infra
│   k8s-worker-02           10.0.1.12            admin           id_rsa_infra
│   k8s-worker-03           10.0.1.13            admin           id_rsa_infra
└────────────────────────────────────────────────────

┌─ ON-PREM ──────────────────────────────────────────
│   nas                     192.168.1.10         user            id_rsa_office
│   server1                 192.168.1.20         admin           id_rsa_office
│   ...
└────────────────────────────────────────────────────
```

<br/>

### JSON output for scripting

```bash
$ bash-pilot ssh list -o json | jq '.[].name'
"git"
"cloud"
"k8s"
"on-prem"
```

<br/>

## SSH Ping

<br/>

### Test all hosts

```bash
$ bash-pilot ssh ping
✓ github.com-personal   0.12s
✓ nas                    0.02s
✗ web-server             timeout (54.123.45.67)
✓ k8s-control-01         0.01s
```

<br/>

### Filter by pattern

```bash
$ bash-pilot ssh ping "k8s-*"
✓ k8s-control-01         0.01s
✓ k8s-worker-01          0.01s
✓ k8s-worker-02          0.01s
✓ k8s-worker-03          0.01s
```

<br/>

### CI/CD connectivity check

```bash
# Fail CI if any host is unreachable
bash-pilot ssh ping -o json | jq -e '[.[] | select(.ok == false)] | length == 0'
```

<br/>

## SSH Audit

```bash
$ bash-pilot ssh audit
! id_rsa_office: used by 8 hosts (consider per-host keys)
! my-region.pem: permissions 0644 (should be 0600)
✓ id_rsa_personal: permissions OK (0600)
✓ id_rsa_work: permissions OK (0600)
```

<br/>

## Git Profiles

### List profiles with active indicator

```bash
$ bash-pilot git profiles
┌─ GIT PROFILES ────────────────────────────────────
│   global (User)        user@gmail.com             (global)
│ → work                 user@company.com           ~/work
│   personal             user@gmail.com             ~/personal
└────────────────────────────────────────────────────
```

The `→` arrow indicates the active profile based on your current directory.

<br/>

### JSON output

```bash
$ bash-pilot git profiles -o json
[
  {
    "name": "global (User)",
    "email": "user@gmail.com",
    "active": false
  },
  {
    "name": "work",
    "directory": "~/work",
    "email": "user@company.com",
    "sign_key": "ABC123",
    "active": true
  }
]
```

<br/>

## Git Doctor

### Diagnose gitconfig issues

```bash
$ bash-pilot git doctor
┌─ GIT DOCTOR ──────────────────────────────────────
! [safe.directory] Duplicate safe.directory: /home/user/repo1 (3 times)
! [includeIf] Include target not found: ~/.gitconfig-old
! [permissions] Gitconfig too open: ~/.gitconfig (mode 0644, recommend 0600)
✓ [user.identity] No global user.email (using includeIf profiles)
└────────────────────────────────────────────────────
```

<br/>

## Git Clean

### Dry run (preview)

```bash
$ bash-pilot git clean --dry-run
┌─ DRY RUN — entries that would be removed ─────────
!   safe.directory=/home/user/repo1 (line 15)
!   safe.directory=/old/project (line 22, directory not found)
└────────────────────────────────────────────────────
```

### Actual cleanup

```bash
$ bash-pilot git clean
┌─ CLEANED ─────────────────────────────────────────
!   safe.directory=/home/user/repo1 (line 15)
!   safe.directory=/old/project (line 22, directory not found)
└────────────────────────────────────────────────────
✓ Backup saved to: /home/user/.gitconfig.bak
```

<br/>

## Env Check

### Shell environment health scan

```bash
$ bash-pilot env check
┌─ ENV CHECK: editor ───────────────────────────────
│ ✓ Editor: vim
└────────────────────────────────────────────────────

┌─ ENV CHECK: git ──────────────────────────────────
│ ✓ git user.email: user@gmail.com
│ ✓ git user.name: Demo User
└────────────────────────────────────────────────────

┌─ ENV CHECK: home ─────────────────────────────────
│ ✓ /home/user/.ssh: OK
│ ✓ /home/user/.config: OK
│ ✓ /home/user/.bashrc: exists
└────────────────────────────────────────────────────

┌─ ENV CHECK: shell ────────────────────────────────
│ ✓ Shell: /bin/bash
│ ✓ Bash version: GNU bash, version 5.2.15...
└────────────────────────────────────────────────────

┌─ ENV CHECK: ssh-agent ────────────────────────────
│ ✓ ssh-agent: 2 key(s) loaded
└────────────────────────────────────────────────────

┌─ ENV CHECK: tools ────────────────────────────────
│ ✓ git: /usr/bin/git
│ ✓ ssh: /usr/bin/ssh
│ ✓ curl: /usr/bin/curl
│ ✓ docker: /usr/bin/docker
│ ! kubectl: not found
│ ! helm: not found
└────────────────────────────────────────────────────

✓ Summary: 12 ok, 2 warnings, 0 errors
```

<br/>

## Env Path

### PATH analysis

```bash
$ bash-pilot env path
┌─ PATH ENTRIES (8 total) ──────────────────────────
│ ✓ [ 1] /usr/local/bin
│ ✓ [ 2] /usr/bin
│ ✓ [ 3] /bin
│ ✓ [ 4] /usr/sbin
│ ✓ [ 5] /sbin
│ ✓ [ 6] /home/user/.local/bin
│ ✓ [ 7] /usr/local/go/bin
│ ✗ [ 8] /old/removed/path
└────────────────────────────────────────────────────

┌─ MISSING DIRECTORIES ────────────────────────────
│ ✗ /old/removed/path
└────────────────────────────────────────────────────

! 8 entries, 0 duplicates, 1 missing
```

<br/>

## Prompt Init

### Generate and apply smart prompt

```bash
# Apply minimal prompt (git only)
$ eval "$(bash-pilot prompt init)"

# Your prompt now shows:
user@host ~/project (main *) ❯

# Apply full prompt (git + k8s)
$ eval "$(bash-pilot prompt init --theme full)"

# Your prompt now shows:
user@host ~/project [prod-cluster:monitoring] (main *) ❯
```

### Persist in shell profile

```bash
# Add to ~/.bashrc or ~/.bash_profile
echo 'eval "$(bash-pilot prompt init --theme full)"' >> ~/.bashrc
source ~/.bashrc
```

<br/>

## Prompt Show

### Preview prompt components

```bash
$ bash-pilot prompt show
┌─ PROMPT COMPONENTS ──────────────────────────────
│ ✓ user@host:   user@hostname
│ ✓ directory:   ~/projects/bash-pilot
│ ✓ git:         main *
└────────────────────────────────────────────────────

$ bash-pilot prompt show --theme full
┌─ PROMPT COMPONENTS ──────────────────────────────
│ ✓ user@host:   user@hostname
│ ✓ directory:   ~/projects/bash-pilot
│ ✓ git:         main *
│ ✓ k8s:         prod-cluster:monitoring
└────────────────────────────────────────────────────
```

<br/>

## Snapshot

### Capture environment snapshot

```bash
$ bash-pilot snapshot > my-env.json
$ cat my-env.json | jq '.tools[] | .name + " " + .version'
"git git version 2.42.0"
"ssh OpenSSH_9.0p1"
"curl curl 8.1.2"
"docker Docker version 24.0.6"
"go go1.22.0"
"node v20.10.0"
```

### Preview snapshot summary

```bash
$ bash-pilot snapshot --summary
┌─ ENVIRONMENT SNAPSHOT ──────────────────────────
Hostname:  my-macbook
OS/Arch:   darwin/arm64
Shell:     /bin/zsh
Tools:     8 installed
SSH Keys:  3
K8s:       2 context(s)
PATH:      15 entries
Brew:      42 packages
Captured:  2024-01-15T10:30:00Z
└────────────────────────────────────────────────
```

<br/>

## Diff

### Compare environments

```bash
$ bash-pilot diff my-env.json
┌─ DIFF vs my-macbook (2024-01-15T10:30:00Z) ────
✓ [System] all match
  [Tools]
  ~ git                  git version 2.42.0 → git version 2.43.0
  - terraform            1.6.0 (not in current)
  + rust                 1.75.0 (new)
✓ [Git] all match
✓ [SSH Keys] all match
  [K8s Contexts]
  - old-cluster          (not in current)
└────────────────────────────────────────────────

! 15 match, 1 changed, 2 missing, 1 new
```

### Filter by section

```bash
# Compare only SSH keys
$ bash-pilot diff my-env.json --only ssh

# Compare only git and tools
$ bash-pilot diff my-env.json --only git,tools
```

### JSON diff for CI

```bash
# Check if environments match
bash-pilot diff baseline.json -o json | jq '.summary.mismatch + .summary.missing'
```

<br/>

## Setup

### Preview install plan

```bash
$ bash-pilot setup teammate-env.json --dry-run
┌─ SETUP PLAN (dry-run) ──────────────────────────
  terraform            brew install terraform
  helm                 brew install helm
  jq                   brew install jq
  htop                 brew install htop
└────────────────────────────────────────────────

! 4 tool(s) to install, 0 skipped

Run without --dry-run to install.
```

### Install missing tools

```bash
$ bash-pilot setup teammate-env.json
┌─ SETUP ─────────────────────────────────────────
  terraform            installed
  helm                 installed
  jq                   installed
  htop                 installed
└────────────────────────────────────────────────

✓ 4 installed, 0 skipped, 0 failed
```

### Install only specific categories

```bash
# Install only missing tools (no brew packages)
$ bash-pilot setup teammate-env.json --only tools

# Install only missing brew packages
$ bash-pilot setup teammate-env.json --only brew
```

### Onboarding workflow

```bash
# Senior engineer saves their environment
bash-pilot snapshot > team-baseline.json

# New team member compares
bash-pilot diff team-baseline.json

# New team member installs missing tools
bash-pilot setup team-baseline.json
```

<br/>

## Migrate Export

### Export config for migration

```bash
$ bash-pilot migrate export > my-config.json
$ cat my-config.json | jq '.ssh.hosts[].name'
"github.com-personal"
"github.com-work"
"server1"
"k8s-control-01"
```

### Check what will be exported

```bash
$ bash-pilot migrate export | jq '{
  ssh_hosts: (.ssh.hosts | length),
  ssh_keys: (.ssh.keys | length),
  git_profiles: (.git.profiles | length)
}'
{
  "ssh_hosts": 14,
  "ssh_keys": 5,
  "git_profiles": 2
}
```

<br/>

## Migrate Import

### Preview import on new machine

```bash
$ bash-pilot migrate import my-config.json --dry-run
┌─ MIGRATE IMPORT (dry-run) ──────────────────────
  SSH:
    14 host(s) added, 0 skipped
    3 key(s) to generate:
      ssh-keygen -t ed25519 -f /home/newuser/.ssh/id_ed25519
      ssh-keygen -t rsa -f /home/newuser/.ssh/id_rsa_work
      ssh-keygen -t ed25519 -f /home/newuser/.ssh/id_ed25519_deploy
  Git:
    global user.name/email configured
    2 profile directory(s) created
    wrote ~/.gitconfig-work
    wrote ~/.gitconfig-personal
└────────────────────────────────────────────────

Run without --dry-run to apply.
```

### Import only specific sections

```bash
# Import only SSH config
$ bash-pilot migrate import my-config.json --only ssh

# Import only Git profiles
$ bash-pilot migrate import my-config.json --only git
```

### Full machine migration workflow

```bash
# On old machine:
bash-pilot snapshot > baseline.json
bash-pilot migrate export > config.json

# On new machine:
bash-pilot setup baseline.json          # install tools
bash-pilot migrate import config.json   # import SSH + git config
# Generate the SSH keys that were listed:
ssh-keygen -t ed25519 -f ~/.ssh/id_ed25519
```

<br/>

## Doctor

### Full system diagnostics

```bash
$ bash-pilot doctor
┌─ DOCTOR: SSH ─────────────────────────────────────
│ ✓ used by 1 host(s)
│ ! used by 14 hosts (consider per-host keys)
│ ✓ permissions OK (0600)
└────────────────────────────────────────────────────

┌─ DOCTOR: GIT ─────────────────────────────────────
│ ✓ No issues found in gitconfig
└────────────────────────────────────────────────────

┌─ DOCTOR: ENV (SHELL) ─────────────────────────────
│ ✓ Shell: /bin/bash
│ ✓ Bash version: GNU bash, version 5.2.15...
└────────────────────────────────────────────────────

┌─ DOCTOR: ENV (TOOLS) ─────────────────────────────
│ ✓ git: /usr/bin/git
│ ✓ ssh: /usr/bin/ssh
│ ✓ curl: /usr/bin/curl
│ ! kubectl: not found
└────────────────────────────────────────────────────

! Total: 2 issue(s) — SSH: 1, Git: 0, Env: 1
```

<br/>

## Scripting with JSON

<br/>

### List all unreachable hosts

```bash
bash-pilot ssh ping -o json | jq -r '.[] | select(.ok == false) | .host.name'
```

<br/>

### Get hosts by group

```bash
bash-pilot ssh list -o json | jq -r '.[] | select(.name == "k8s") | .hosts[].hostname'
```

<br/>

### Audit to markdown report

```bash
echo "# SSH Audit Report"
echo ""
echo "| Key | Status | Detail |"
echo "|-----|--------|--------|"
bash-pilot ssh audit -o json | jq -r '.findings[] | "| \(.key) | \(.severity) | \(.message) |"'
```
# test
