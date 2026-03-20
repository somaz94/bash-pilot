#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BINARY="${PROJECT_DIR}/bin/bash-pilot"
DEMO_DIR="/tmp/bash-pilot-demo"
DEMO_SSH_CONFIG="${DEMO_DIR}/ssh_config"
DEMO_CONFIG="${DEMO_DIR}/config.yaml"

# Colors
GREEN='\033[32m'
YELLOW='\033[33m'
CYAN='\033[36m'
BOLD='\033[1m'
RESET='\033[0m'

header() {
  echo ""
  echo -e "${BOLD}${CYAN}=== Phase $1: $2 ===${RESET}"
  echo ""
}

run() {
  echo -e "${YELLOW}\$ $*${RESET}"
  eval "$@"
  echo ""
}

# Build if needed
if [ ! -f "$BINARY" ]; then
  echo "Building bash-pilot..."
  (cd "$PROJECT_DIR" && make build)
fi

# ============================================================
header 1 "Setup demo environment"
# ============================================================

mkdir -p "$DEMO_DIR/.ssh"

# Create fake SSH keys for demo
for key in id_rsa_personal id_rsa_work id_rsa_deploy staging.pem; do
  touch "$DEMO_DIR/.ssh/$key"
  chmod 0600 "$DEMO_DIR/.ssh/$key"
done

# Create a key with bad permissions for audit demo
touch "$DEMO_DIR/.ssh/id_rsa_insecure"
chmod 0644 "$DEMO_DIR/.ssh/id_rsa_insecure"

# Create demo SSH config
cat > "$DEMO_SSH_CONFIG" <<'EOF'
Host github.com-personal
  Hostname github.com
  User git
  IdentityFile DEMO_DIR/.ssh/id_rsa_personal

Host github.com-work
  Hostname github.com
  User git
  IdentityFile DEMO_DIR/.ssh/id_rsa_work

Host gitlab-internal
  Hostname 10.10.10.60
  User deploy
  IdentityFile DEMO_DIR/.ssh/id_rsa_work

Host staging-web
  Hostname 52.78.100.10
  User ec2-user
  IdentityFile DEMO_DIR/.ssh/staging.pem

Host staging-api
  Hostname 52.78.100.11
  User ec2-user
  IdentityFile DEMO_DIR/.ssh/staging.pem

Host staging-db
  Hostname 52.78.100.12
  User ec2-user
  IdentityFile DEMO_DIR/.ssh/staging.pem

Host nas
  Hostname 10.10.10.5
  User admin
  IdentityFile DEMO_DIR/.ssh/id_rsa_work

Host server1
  Hostname 10.10.10.10
  User deploy
  IdentityFile DEMO_DIR/.ssh/id_rsa_work

Host server2
  Hostname 10.10.10.12
  User deploy
  IdentityFile DEMO_DIR/.ssh/id_rsa_work

Host k8s-control-01
  Hostname 10.10.10.17
  User deploy
  IdentityFile DEMO_DIR/.ssh/id_rsa_deploy

Host k8s-compute-01
  Hostname 10.10.10.18
  User deploy
  IdentityFile DEMO_DIR/.ssh/id_rsa_deploy

Host k8s-compute-02
  Hostname 10.10.10.19
  User deploy
  IdentityFile DEMO_DIR/.ssh/id_rsa_deploy

Host k8s-compute-03
  Hostname 10.10.10.22
  User deploy
  IdentityFile DEMO_DIR/.ssh/id_rsa_deploy

Host jump-box
  Hostname 203.0.113.50
  User admin
  IdentityFile DEMO_DIR/.ssh/id_rsa_insecure
EOF

# Replace DEMO_DIR placeholder with actual path
sed -i '' "s|DEMO_DIR|${DEMO_DIR}|g" "$DEMO_SSH_CONFIG" 2>/dev/null || \
  sed -i "s|DEMO_DIR|${DEMO_DIR}|g" "$DEMO_SSH_CONFIG"

# Create demo config
cat > "$DEMO_CONFIG" <<EOF
ssh:
  config_file: ${DEMO_SSH_CONFIG}
  groups:
    git:
      pattern: ["github.com*", "gitlab*"]
    cloud:
      pattern: ["staging-*"]
      label: "AWS Seoul"
    k8s:
      pattern: ["k8s-*"]
    on-prem:
      pattern: ["server*", "nas*"]
  ping:
    timeout: 3s
    parallel: 10
EOF

echo -e "${GREEN}Demo environment created at ${DEMO_DIR}${RESET}"
echo ""

# ============================================================
header 2 "init — Auto-generate config from SSH config"
# ============================================================

export HOME="$DEMO_DIR"
mkdir -p "$DEMO_DIR/.config/bash-pilot"

run "$BINARY" init --config "$DEMO_CONFIG" --force

echo -e "${GREEN}Config auto-generated at ~/.config/bash-pilot/config.yaml${RESET}"
echo ""

# ============================================================
header 3 "ssh list — Host grouping"
# ============================================================

run "$BINARY" ssh list --config "$DEMO_CONFIG"

# ============================================================
header 4 "ssh list — JSON output"
# ============================================================

run "$BINARY" ssh list --config "$DEMO_CONFIG" -o json

# ============================================================
header 5 "ssh ping — Connectivity test (all hosts)"
# ============================================================

echo -e "${YELLOW}Note: Most hosts will timeout since they are demo IPs${RESET}"
echo ""
run "$BINARY" ssh ping --config "$DEMO_CONFIG" || true

# ============================================================
header 6 "ssh ping — Filter by pattern (k8s-* only)"
# ============================================================

run "$BINARY" ssh ping --config "$DEMO_CONFIG" 'k8s-*' || true

# ============================================================
header 7 "ssh audit — Security audit"
# ============================================================

run "$BINARY" ssh audit --config "$DEMO_CONFIG"

# ============================================================
header 8 "ssh audit — JSON output"
# ============================================================

run "$BINARY" ssh audit --config "$DEMO_CONFIG" -o json

# ============================================================
header 9 "git profiles — Git identity profiles"
# ============================================================

# Create a demo gitconfig with includeIf profiles
DEMO_GITCONFIG="${DEMO_DIR}/.gitconfig"
DEMO_GITCONFIG_WORK="${DEMO_DIR}/.gitconfig-work"
DEMO_GITCONFIG_PERSONAL="${DEMO_DIR}/.gitconfig-personal"

cat > "$DEMO_GITCONFIG_WORK" <<'EOF'
[user]
	email = developer@company.com
	signingkey = ABC123DEF456
EOF

cat > "$DEMO_GITCONFIG_PERSONAL" <<'EOF'
[user]
	email = user@gmail.com
EOF

cat > "$DEMO_GITCONFIG" <<EOF
[user]
	name = Demo User
	email = demo@example.com
[includeIf "gitdir:${DEMO_DIR}/work/"]
	path = ${DEMO_GITCONFIG_WORK}
[includeIf "gitdir:${DEMO_DIR}/personal/"]
	path = ${DEMO_GITCONFIG_PERSONAL}
[safe]
	directory = ${DEMO_DIR}/repo1
	directory = ${DEMO_DIR}/repo1
	directory = ${DEMO_DIR}/repo2
	directory = /nonexistent/old-project
EOF

run "$BINARY" git profiles --gitconfig "$DEMO_GITCONFIG"

# ============================================================
header 10 "git profiles — JSON output"
# ============================================================

run "$BINARY" git profiles --gitconfig "$DEMO_GITCONFIG" -o json

# ============================================================
header 11 "git doctor — Diagnose gitconfig issues"
# ============================================================

run "$BINARY" git doctor --gitconfig "$DEMO_GITCONFIG"

# ============================================================
header 12 "git clean — Preview stale/duplicate entries"
# ============================================================

run "$BINARY" git clean --gitconfig "$DEMO_GITCONFIG" --dry-run

# ============================================================
header 13 "git clean — Actually clean up"
# ============================================================

run "$BINARY" git clean --gitconfig "$DEMO_GITCONFIG"

echo -e "${GREEN}Cleaned gitconfig. Backup at ${DEMO_GITCONFIG}.bak${RESET}"
echo ""

# ============================================================
header 14 "env check — Shell environment health scan"
# ============================================================

run "$BINARY" env check

# ============================================================
header 15 "env check — JSON output"
# ============================================================

run "$BINARY" env check -o json

# ============================================================
header 16 "env path — PATH analysis"
# ============================================================

run "$BINARY" env path

# ============================================================
header 17 "env path — JSON output"
# ============================================================

run "$BINARY" env path -o json

# ============================================================
header 18 "version"
# ============================================================

run "$BINARY" version

# ============================================================
echo ""
echo -e "${GREEN}${BOLD}Demo complete!${RESET}"
echo -e "Run ${CYAN}make demo-clean${RESET} to remove demo resources."
