# Use Cases

Real-world scenarios for bash-pilot.

<br/>

## Table of Contents

- [SSH Host Inventory](#ssh-host-inventory)
- [Infrastructure Health Check](#infrastructure-health-check)
- [Security Compliance Audit](#security-compliance-audit)
- [Onboarding New Team Members](#onboarding-new-team-members)
- [Environment Snapshot and Diff](#environment-snapshot-and-diff)
- [New Machine Setup](#new-machine-setup)
- [Config Migration Across Machines](#config-migration-across-machines)
- [CI/CD Pipeline Integration](#cicd-pipeline-integration)
- [Multi-Environment Management](#multi-environment-management)

<br/>

## SSH Host Inventory

**Scenario:** You have 20+ hosts in `~/.ssh/config` and need a quick overview.

```bash
# Visual overview with automatic grouping
bash-pilot ssh list

# Export as JSON for documentation
bash-pilot ssh list -o json > ssh-inventory.json

# Count hosts per group
bash-pilot ssh list -o json | jq '.[] | {group: .name, count: (.hosts | length)}'
```

**Before bash-pilot:** Manually scroll through `~/.ssh/config` or run `grep Host ~/.ssh/config`.

**After:** Instant grouped view with host details at a glance.

<br/>

## Infrastructure Health Check

**Scenario:** Morning check — are all servers reachable?

```bash
# Quick connectivity check for all hosts
bash-pilot ssh ping

# Check only production k8s nodes
bash-pilot ssh ping "k8s-*"

# Check only cloud instances
bash-pilot ssh ping "staging-*"
```

**Automation example (crontab):**

```bash
# Alert on unreachable hosts every 30 minutes
*/30 * * * * bash-pilot ssh ping -o json | jq '[.[] | select(.ok==false)]' | grep -q '"name"' && notify-send "SSH host down!"
```

<br/>

## Security Compliance Audit

**Scenario:** Security review — are SSH keys properly managed?

```bash
# Run full audit
bash-pilot ssh audit

# Example findings:
# ! id_rsa_shared: used by 8 hosts (consider per-host keys)
# ! staging.pem: permissions 0644 (should be 0600)
# ✓ id_rsa_personal: permissions OK (0600)
```

**Common findings and fixes:**

| Finding | Fix |
|---------|-----|
| Key used by N hosts | Generate per-host or per-group keys |
| Permissions 0644 | `chmod 0600 ~/.ssh/key_file` |
| Key file not found | Remove stale entry or restore key |
| No IdentityFile specified | Add `IdentityFile` directive |

<br/>

## Environment Snapshot and Diff

**Scenario:** "It works on my machine" — something's different on your colleague's setup but you don't know what.

```bash
# Capture your working environment
bash-pilot snapshot > my-env.json

# Share the file with your colleague, they run:
bash-pilot diff my-env.json

# Output:
# ✓ [System] all match
# [Tools]
#   ~ node                 v20.10.0 → v18.17.0
#   - terraform            1.6.0 (not in current)
# ✓ [Git] all match
# [K8s Contexts]
#   - prod-cluster         (not in current)
#
# ! 20 match, 1 changed, 2 missing, 0 new
```

**Before bash-pilot:** Manually compare `node --version`, `terraform --version`, `kubectl config get-contexts`... one by one.

**After:** Single command shows every difference across tools, versions, git, SSH, k8s, and brew packages.

<br/>

## New Machine Setup

**Scenario:** New MacBook or Linux server — need to install the same tools as your current machine.

```bash
# On your existing machine: save snapshot
bash-pilot snapshot > team-baseline.json

# On the new machine: preview what's missing
bash-pilot setup team-baseline.json --dry-run

# Output:
# ┌─ SETUP PLAN (dry-run) ──────────────────────────
#   terraform            brew install terraform
#   helm                 brew install helm
#   htop                 brew install htop
# └────────────────────────────────────────────────
# ! 3 tool(s) to install, 0 skipped

# Install everything
bash-pilot setup team-baseline.json
```

**Full onboarding workflow:**

```bash
# 1. Senior engineer saves baseline
bash-pilot snapshot > team-baseline.json

# 2. New team member checks what's different
bash-pilot diff team-baseline.json

# 3. Install missing tools automatically
bash-pilot setup team-baseline.json

# 4. Verify everything matches
bash-pilot diff team-baseline.json
# ✓ all match
```

**Before bash-pilot:** Share a wiki page with install instructions, hope everyone follows all steps correctly.

**After:** One snapshot file + two commands = identical environment.

<br/>

## Onboarding New Team Members

**Scenario:** New developer needs to understand the SSH infrastructure.

```bash
# Show them the full host inventory
bash-pilot ssh list

# Export for wiki/docs
bash-pilot ssh list -o json | jq -r '
  .[] | "## \(.name)\n" + (.hosts[] | "- \(.name) → \(.hostname) (user: \(.user))")
' > ssh-hosts.md

# Verify their connectivity
bash-pilot ssh ping
```

<br/>

## Config Migration Across Machines

**Scenario:** New MacBook, new Linux server, or just want the same SSH hosts and Git profiles everywhere.

```bash
# On your current machine: export config
bash-pilot migrate export > my-config.json

# On the new machine: preview what will be imported
bash-pilot migrate import my-config.json --dry-run

# Apply — SSH hosts added, Git profiles created, key generation commands shown
bash-pilot migrate import my-config.json

# Generate the missing SSH keys
ssh-keygen -t ed25519 -f ~/.ssh/id_ed25519
ssh-keygen -t ed25519 -f ~/.ssh/id_ed25519_work
```

**What makes this different from `cp ~/.ssh/config`:**
- Paths are automatically translated (`/Users/somaz/` → `/home/somaz/`)
- Existing hosts are not overwritten
- SSH keys are not copied (security) — only generation commands are provided
- Git includeIf profiles and directories are auto-created

**Complete new machine setup (3 commands):**

```bash
bash-pilot setup baseline.json           # Install missing tools
bash-pilot migrate import my-config.json  # Import SSH + Git config
bash-pilot diff baseline.json            # Verify everything matches
```

**Before bash-pilot:** Manually copy files, fix paths, regenerate keys, edit gitconfig.

**After:** Three commands. Done.

<br/>

## CI/CD Pipeline Integration

**Scenario:** Pre-deployment check that target servers are reachable.

```yaml
# GitHub Actions example
- name: Check deployment targets
  run: |
    bash-pilot ssh ping "prod-*" -o json | \
      jq -e '[.[] | select(.ok == false)] | length == 0' || \
      (echo "Some hosts are unreachable!" && exit 1)
```

```bash
# Jenkins / shell script
if ! bash-pilot ssh ping "prod-*" -o json | jq -e 'all(.ok)'; then
  echo "ABORT: Not all production hosts are reachable"
  exit 1
fi
```

<br/>

## Multi-Environment Management

**Scenario:** Managing dev, staging, and production environments with different SSH configs.

```yaml
# ~/.config/bash-pilot/config.yaml
ssh:
  groups:
    dev:
      pattern: ["dev-*"]
      label: "Development"
    staging:
      pattern: ["staging-*"]
      label: "AWS Staging"
    prod:
      pattern: ["prod-*"]
      label: "AWS Production"
    k8s-dev:
      pattern: ["k8s-dev-*"]
    k8s-prod:
      pattern: ["k8s-prod-*"]
```

```bash
# Check staging before deploying
bash-pilot ssh ping "staging-*"

# Check production after deploying
bash-pilot ssh ping "prod-*"
```
