package system

import (
	"os/exec"
	"syscall"
)

// IsAdmin 检查是否以管理员权限运行
func IsAdmin() bool {
	cmd := exec.Command("net", "session")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err := cmd.Run()
	return err == nil
}
