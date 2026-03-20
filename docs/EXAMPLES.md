# Examples

Hands-on examples for bash-pilot.

<br/>

## Table of Contents

- [Quick Demo](#quick-demo)
- [SSH List](#ssh-list)
- [SSH Ping](#ssh-ping)
- [SSH Audit](#ssh-audit)
- [Scripting with JSON](#scripting-with-json)

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
│   github.com-somaz94        github.com           somaz           id_rsa_somaz94
│   github.com-somaz940829    github.com           somaz-devops    id_rsa_somaz940829
└────────────────────────────────────────────────────

┌─ CLOUD ────────────────────────────────────────────
│   test-server               3.65.182.184         ec2-user        frankfurt-habby-1704.pem
│   jenkins                   18.159.54.27         ec2-user        frankfurt-habby-1704.pem
└────────────────────────────────────────────────────

┌─ K8S ──────────────────────────────────────────────
│   k8s-control-01            10.10.10.17          concrit         id_rsa_concrit
│   k8s-compute-01            10.10.10.18          concrit         id_rsa_concrit
│   k8s-compute-02            10.10.10.19          concrit         id_rsa_concrit
│   k8s-compute-03            10.10.10.22          concrit         id_rsa_concrit
└────────────────────────────────────────────────────

┌─ ON-PREM ──────────────────────────────────────────
│   nas                       10.10.10.5           somaz           id_rsa_concrit
│   server1                   10.10.10.10          concrit         id_rsa_concrit
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
✓ github.com-somaz94     0.12s
✓ nas                    0.02s
✗ test-server            timeout (3.65.182.184)
✓ k8s-control-01         0.01s
```

<br/>

### Filter by pattern

```bash
$ bash-pilot ssh ping "k8s-*"
✓ k8s-control-01         0.01s
✓ k8s-compute-01         0.01s
✓ k8s-compute-02         0.01s
✓ k8s-compute-03         0.01s
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
! id_rsa_concrit: used by 12 hosts (consider per-host keys)
! frankfurt-habby-1704.pem: permissions 0644 (should be 0600)
✓ id_rsa_somaz94: permissions OK (0600)
✓ id_rsa_somaz940829: permissions OK (0600)
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
