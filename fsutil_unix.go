//go:build !windows
// +build !windows

package main

import (
	"os"
	"runtime"
)

// moveToTrashPlatform 平台特定的移动到回收站实现
func moveToTrashPlatform(filePath string) error {
	switch runtime.GOOS {
	case "darwin":
		// macOS平台实现
		return moveToTrashMacOS(filePath)
	case "linux":
		// Linux平台实现
		return moveToTrashLinux(filePath)
	default:
		// 不支持的平台，直接删除文件
		return os.RemoveAll(filePath)
	}
}

// isWindowsHiddenFile 检查是否为Windows隐藏文件（非Windows平台实现）
func isWindowsHiddenFile(path string) bool {
	// 非Windows平台，检查文件名是否以点开头
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return len(info.Name()) > 0 && info.Name()[0] == '.'
}

// isWindowsSystemFile 检查是否为Windows系统文件（非Windows平台实现）
func isWindowsSystemFile(path string) bool {
	// 非Windows平台，返回false
	return false
}

// checkDiskSpace 检查磁盘空间（非Windows平台实现）
func checkDiskSpace(path string, requiredBytes int64) error {
	// 非Windows平台暂不检查磁盘空间，直接返回成功
	return nil
}
