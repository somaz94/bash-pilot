# macOS: Switch Default Shell from Zsh to Bash

Guide to switching the macOS default shell from Zsh to Bash 5.x (Homebrew) and setting up a productive development environment.

<br/>

## Table of Contents

- [Why Switch?](#why-switch)
- [Step 1: Install Homebrew Bash](#step-1-install-homebrew-bash)
- [Step 2: Register as Login Shell](#step-2-register-as-login-shell)
- [Step 3: Change Default Shell](#step-3-change-default-shell)
- [Step 4: Configure .bash_profile](#step-4-configure-bash_profile)
- [Step 5: Verify Setup](#step-5-verify-setup)
- [Optional: Useful Add-ons](#optional-useful-add-ons)
- [Notes](#notes)

<br/>

## Why Switch?

macOS ships with Bash 3.2 (2007, GPLv2) as a legacy shell. Apple defaults to Zsh since Catalina, but many DevOps tools and scripts still target Bash. Homebrew provides Bash 5.x with modern features:

| Feature | Bash 3.2 (macOS built-in) | Bash 5.x (Homebrew) |
|---------|---------------------------|----------------------|
| Associative arrays | No | Yes |
| `mapfile` / `readarray` | No | Yes |
| `${var,,}` case conversion | No | Yes |
| Programmable completion v2 | No | Yes |
| `nameref` variables | No | Yes |

<br/>

## Step 1: Install Homebrew Bash

```bash
# Install Homebrew if not already installed
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install latest Bash
brew install bash
```

Check version:

```bash
/opt/homebrew/bin/bash --version
# GNU bash, version 5.x.x ...
```

> **Note:** On Apple Silicon (M1/M2/M3/M4), Homebrew prefix is `/opt/homebrew`. On Intel Macs, it's `/usr/local`.

<br/>

## Step 2: Register as Login Shell

Add Homebrew Bash to the allowed login shells:

```bash
# Check if already registered
grep '/opt/homebrew/bin/bash' /etc/shells

# If not listed, add it
echo '/opt/homebrew/bin/bash' | sudo tee -a /etc/shells
```

Verify:

```bash
cat /etc/shells
# Should include: /opt/homebrew/bin/bash
```

<br/>

## Step 3: Change Default Shell

```bash
chsh -s /opt/homebrew/bin/bash
```

Close and reopen your terminal, then verify:

```bash
echo $BASH_VERSION
# 5.x.x(1)-release

echo $SHELL
# /opt/homebrew/bin/bash
```

<br/>

## Step 4: Configure .bash_profile

On macOS, Bash reads `~/.bash_profile` for login shells (not `~/.bashrc`). Create or edit `~/.bash_profile`:

```bash
vi ~/.bash_profile
```

<br/>

### Homebrew Shell Environment

This must come first so Homebrew tools are available to subsequent configs:

```bash
eval "$(/opt/homebrew/bin/brew shellenv)"
```

<br/>

### Bash Completion

Install bash-completion v2 (required for modern CLI tools like kubectl, Cobra-based CLIs):

```bash
brew install bash-completion@2
```

> **Important:** bash-completion@2 is required. Version 1 (`bash-completion`) does not support Cobra-based tools like `kubectl`, `bash-pilot`, etc.

Add to `.bash_profile`:

```bash
if type brew &>/dev/null; then
  HOMEBREW_PREFIX=$(brew --prefix)
  if [[ -r "${HOMEBREW_PREFIX}/etc/profile.d/bash_completion.sh" ]]; then
    source "${HOMEBREW_PREFIX}/etc/profile.d/bash_completion.sh"
  fi
fi
```

<br/>

### Custom Prompt (PS1)

```bash
export PS1="\[\033[36m\]\u \[\033[32m\]\w \$ \[\033[0m\]"
```

<br/>

### Git Prompt (optional)

Show git branch info in your prompt:

```bash
brew install bash-git-prompt
```

```bash
if [ -f "$(brew --prefix)/opt/bash-git-prompt/share/gitprompt.sh" ]; then
  GIT_PROMPT_ONLY_IN_REPO=1
  source "$(brew --prefix)/opt/bash-git-prompt/share/gitprompt.sh"
fi
```

<br/>

### SSH Keys Auto-Load

```bash
ssh-add ~/.ssh/id_rsa_personal
ssh-add ~/.ssh/id_rsa_work
```

<br/>

### Kubectl Completion and Alias

```bash
source <(kubectl completion bash)
alias k=kubectl
complete -o default -F __start_kubectl k
```

<br/>

### Krew Plugin Manager

```bash
export PATH="${KREW_ROOT:-$HOME/.krew}/bin:$PATH"
```

<br/>

### Terraform Completion

```bash
complete -C /opt/homebrew/bin/terraform terraform
```

<br/>

### PATH Exports

```bash
# OpenJDK
export PATH="/opt/homebrew/opt/openjdk/bin:$PATH"

# Ruby (Homebrew)
export PATH="/opt/homebrew/opt/ruby/bin:$PATH"
export GEM_HOME="$HOME/.gem"
export PATH="$GEM_HOME/bin:$PATH"

# MySQL client
export PATH="/opt/homebrew/opt/mysql-client/bin:$PATH"

# libpq (PostgreSQL client)
export PATH="/opt/homebrew/opt/libpq/bin:$PATH"

# Android SDK
export PATH="$PATH:~/Library/Android/sdk/platform-tools"
```

<br/>

### Full Example

<details>
<summary>Complete .bash_profile example</summary>

```bash
# Homebrew
eval "$(/opt/homebrew/bin/brew shellenv)"

# Bash Completion (v2)
if type brew &>/dev/null; then
  HOMEBREW_PREFIX=$(brew --prefix)
  if [[ -r "${HOMEBREW_PREFIX}/etc/profile.d/bash_completion.sh" ]]; then
    source "${HOMEBREW_PREFIX}/etc/profile.d/bash_completion.sh"
  fi
fi

# Kubectl
source <(kubectl completion bash)
alias k=kubectl
complete -o default -F __start_kubectl k

# Krew
export PATH="${KREW_ROOT:-$HOME/.krew}/bin:$PATH"

# OpenJDK
export PATH="/opt/homebrew/opt/openjdk/bin:$PATH"

# Git Prompt
if [ -f "$(brew --prefix)/opt/bash-git-prompt/share/gitprompt.sh" ]; then
  GIT_PROMPT_ONLY_IN_REPO=1
  source "$(brew --prefix)/opt/bash-git-prompt/share/gitprompt.sh"
fi

# PS1
export PS1="\[\033[36m\]\u \[\033[32m\]\w \$ \[\033[0m\]"

# SSH Keys
ssh-add ~/.ssh/id_rsa_personal
ssh-add ~/.ssh/id_rsa_work

# PATH
export PATH="/opt/homebrew/opt/ruby/bin:$PATH"
export GEM_HOME="$HOME/.gem"
export PATH="$GEM_HOME/bin:$PATH"
export PATH="/opt/homebrew/opt/mysql-client/bin:$PATH"
export PATH="/opt/homebrew/opt/libpq/bin:$PATH"
export PATH="$PATH:~/Library/Android/sdk/platform-tools"

# Terraform
complete -C /opt/homebrew/bin/terraform terraform
```

</details>

<br/>

## Step 5: Verify Setup

Restart terminal and run:

```bash
# Check shell
echo $SHELL             # /opt/homebrew/bin/bash
echo $BASH_VERSION      # 5.x.x(1)-release

# Check Homebrew
brew --version

# Check completion
type _init_completion   # Should show "is a function"

# Check tools
kubectl version --client
terraform version
```

<br/>

## Optional: Useful Add-ons

<br/>

### OrbStack (Docker alternative)

If using OrbStack for containers/Linux VMs:

```bash
source ~/.orbstack/shell/init.bash 2>/dev/null || :
```

<br/>

### bash-pilot

Auto-manage SSH hosts and config:

```bash
brew install somaz94/tap/bash-pilot
bash-pilot init
bash-pilot completion bash > "$(brew --prefix)/etc/bash_completion.d/bash-pilot"
```

<br/>

## Notes

<br/>

### .bash_profile vs .bashrc

| File | When loaded |
|------|-------------|
| `~/.bash_profile` | Login shells (macOS Terminal, iTerm2 default) |
| `~/.bashrc` | Non-login interactive shells |

macOS Terminal and iTerm2 open login shells by default, so `~/.bash_profile` is the right place. If you also need `.bashrc`, source it from `.bash_profile`:

```bash
# At end of .bash_profile
if [ -f ~/.bashrc ]; then
  source ~/.bashrc
fi
```

<br/>

### Reverting to Zsh

```bash
chsh -s /bin/zsh
```

<br/>

### Updating Bash

```bash
brew upgrade bash
```
