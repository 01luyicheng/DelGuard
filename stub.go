//go:build windows
// +build windows

package main

import (
	"fmt"
	"os"
)

// moveToTrashPlatform 平台特定的移动到回收站实现
func moveToTrashPlatform(filePath string) error {
	return moveToTrashWindows(filePath)
}

// isEXDEV 检查是否为跨设备错误 (Windows平台无此错误)
func isEXDEV(err error) bool {
	return false
}

// removeOriginal 删除原始文件
func removeOriginal(path string) error {
	return os.RemoveAll(path)
}

// copyTree 复制目录树
func copyTree(src, dst string) error {
	// Windows平台使用系统API，不需要手动复制
	return fmt.Errorf("Windows平台不支持此操作")
}