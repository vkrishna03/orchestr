//go:build !windows

package runner

import (
	"os/exec"
	"syscall"
)

// setProcGroup sets process group for clean shutdown on Unix.
func setProcGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}

// killProcessGroup sends a signal to the process group.
func killProcessGroup(cmd *exec.Cmd, sig syscall.Signal) error {
	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err == nil {
		return syscall.Kill(-pgid, sig)
	}
	return cmd.Process.Signal(sig)
}

// interruptProcess sends SIGINT to the process.
func interruptProcess(cmd *exec.Cmd) error {
	return killProcessGroup(cmd, syscall.SIGINT)
}

// forceKillProcess forcefully kills the process.
func forceKillProcess(cmd *exec.Cmd) error {
	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err == nil {
		return syscall.Kill(-pgid, syscall.SIGKILL)
	}
	return cmd.Process.Kill()
}
