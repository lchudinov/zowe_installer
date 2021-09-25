package launcher

import "syscall"

func kill(pid int) error {
	return nil
}

func getSysProcAttr() *syscall.SysProcAttr {
	var attr syscall.SysProcAttr
	attr.CreationFlags = syscall.CREATE_NEW_PROCESS_GROUP
	return &attr
}
