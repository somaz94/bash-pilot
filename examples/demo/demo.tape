# Runs the real bash-pilot inside the vhs container against the synthetic SSH
# config in this directory, installed into a throwaway HOME. The key files are
# empty and created here, one with loose permissions for the audit.
# To regenerate, from the repo root:
#   GOOS=linux go build -o bin/linux/bash-pilot ./cmd
#   docker run --rm -v "$PWD:/vhs" ghcr.io/charmbracelet/vhs examples/demo/demo.tape
Output examples/demo/demo.gif

Set Shell bash
Set FontSize 16
Set Width 1280
Set Height 720
Set Padding 16
Set TypingSpeed 35ms
Set Theme "Catppuccin Mocha"

Hide
Type "export PATH=/vhs/bin/linux:$PATH HOME=/tmp/home && mkdir -p ~/.ssh ~/.config/bash-pilot && cp examples/demo/ssh_config ~/.ssh/config && cp examples/demo/config.yaml ~/.config/bash-pilot/"
Enter
Type "cd ~/.ssh && touch id_ed25519_personal id_ed25519_work id_ed25519_deploy id_ed25519_old && chmod 600 id_ed25519_* && chmod 644 id_ed25519_old && cd ~"
Enter
Type "PS1='$ ' && clear"
Enter
Sleep 500ms
Show

Type "bash-pilot ssh list"
Enter
Sleep 3s
Type "bash-pilot ssh audit"
Enter
Sleep 5s
