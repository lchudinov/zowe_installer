// +build linux zos !windows

package launcher

import "syscall"

func kill(pid int) error {
	return syscall.Kill(-pid, syscall.SIGTERM)
}

func getSysProcAttr() *syscall.SysProcAttr {
	var attr syscall.SysProcAttr
	attr.Setsid = true
	attr.Setpgid = true
	attr.Pgid = 0
	return &attr
}
