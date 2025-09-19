package main

import (
	"os/exec"
	"syscall"
)

func humanizeBoolean(b bool) string {
	if b {
		return "✅ Yes"
	}
	return "❌ No"
}

func executeCommand(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd
}
