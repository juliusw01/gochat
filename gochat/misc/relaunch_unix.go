//go:build !windows
// +build !windows

package misc

import (
	"os/exec"
	"syscall"
)

func SetBackgroundAttributes(cmd *exec.Cmd) {
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
}