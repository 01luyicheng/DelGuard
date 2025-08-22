package main

import (
	"runtime"
)

// detectPlatform 检测当前操作系统平台
func detectPlatform() string {
	return runtime.GOOS
}

// moveToTrashPlatform 根据平台调用相应的回收站移动函数
func moveToTrashPlatform(filePath string) error {
	switch detectPlatform() {
	case "windows":
		return moveToTrashWindows(filePath)
	case "darwin":
		return moveToTrashMacOS(filePath)
	case "linux":
		return moveToTrashLinux(filePath)
	default:
		return ErrUnsupportedPlatform
	}
}

// 以下函数在各平台特定文件中实现
var (
	// 这些变量仅用于编译时检查，确保各平台实现了所需函数
	_ = moveToTrashWindows
	_ = moveToTrashMacOS
	_ = moveToTrashLinux
)
