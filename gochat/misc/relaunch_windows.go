//go:build windows
// +build windows

package misc

import (
	"os/exec"
	"syscall"
)

func SetBackgroundAttributes(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
}
