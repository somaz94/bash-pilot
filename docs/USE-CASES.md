# Use Cases

Real-world scenarios for bash-pilot.

<br/>

## Table of Contents

- [SSH Host Inventory](#ssh-host-inventory)
- [Infrastructure Health Check](#infrastructure-health-check)
- [Security Compliance Audit](#security-compliance-audit)
- [Onboarding New Team Members](#onboarding-new-team-members)
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
