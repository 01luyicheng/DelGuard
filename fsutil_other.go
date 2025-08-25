//go:build !windows
// +build !windows

package main

import (
	"os"
	"syscall"
)

// isEXDEV 检查是否为跨设备错误
func isEXDEV(err error) bool {
	if err == nil {
		return false
	}
	pathErr, ok := err.(*os.LinkError)
	if !ok {
		return false
	}
	errno, ok := pathErr.Err.(syscall.Errno)
	return ok && errno == syscall.EXDEV
}

// removeOriginal 删除原始文件
func removeOriginal(path string) error {
	return os.RemoveAll(path)
}

// copyTree 复制目录树
func copyTree(src, dst string) error {
	// 实现目录树复制逻辑
	return nil
}
