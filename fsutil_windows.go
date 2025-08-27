//go:build windows
// +build windows

package main

// moveToTrashPlatform 平台特定的移动到回收站实现
func moveToTrashPlatform(filePath string) error {
	return moveToTrashWindows(filePath)
}
