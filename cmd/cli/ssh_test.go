package cli

import "testing"

func TestSSHPingArgs(t *testing.T) {
	for _, args := range [][]string{nil, {"web-*"}} {
		if err := sshPingCmd.Args(sshPingCmd, args); err != nil {
			t.Errorf("Args(%q) = %v, want nil", args, err)
		}
	}
	// Only args[0] is used as the pattern, so a second one must not be dropped silently.
	if err := sshPingCmd.Args(sshPingCmd, []string{"web-1", "web-2"}); err == nil {
		t.Error("Args with two patterns = nil, want an error")
	}
}
