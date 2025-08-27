//go:build !windows
// +build !windows

package main

import (
	"os"
)

// moveToTrashWindows Windows平台回收站删除（存根）
func (cd *CoreDeleter) moveToTrashWindows(path string) error {
	// 非Windows平台不支持
	return os.Remove(path)
}

// moveToTrashMacOS macOS平台回收站删除
func (cd *CoreDeleter) moveToTrashMacOS(path string) error {
	// 暂时使用永久删除
	return os.Remove(path)
}

// moveToTrashLinux Linux平台回收站删除
func (cd *CoreDeleter) moveToTrashLinux(path string) error {
	// 暂时使用永久删除
	return os.Remove(path)
}
