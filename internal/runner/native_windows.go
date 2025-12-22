//go:build windows

package runner

import (
	"os/exec"
)

// setProcGroup is a no-op on Windows.
func setProcGroup(cmd *exec.Cmd) {
	// Windows doesn't support process groups in the same way
}

// interruptProcess sends interrupt to the process on Windows.
func interruptProcess(cmd *exec.Cmd) error {
	// Windows doesn't have SIGINT, just kill the process
	return cmd.Process.Kill()
}

// forceKillProcess forcefully kills the process on Windows.
func forceKillProcess(cmd *exec.Cmd) error {
	return cmd.Process.Kill()
}
