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
