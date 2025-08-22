//go:build windows
// +build windows

package main

import (
	"syscall"
)

var (
	shell32IsUserAnAdmin = syscall.NewLazyDLL("shell32.dll").NewProc("IsUserAnAdmin")
)

// IsElevated 返回是否以管理员权限运行（Windows）
func IsElevated() bool {
	// 使用 shell32!IsUserAnAdmin 简单判断
	if shell32IsUserAnAdmin.Find() != nil {
		return false
	}
	ret, _, _ := shell32IsUserAnAdmin.Call()
	return ret != 0
}
