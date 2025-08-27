//go:build !linux && !windows
// +build !linux,!windows

package main

// moveToTrashLinux Linux 平台的回收站实现存根
func moveToTrashLinux(filePath string) error {
	// 非 Linux 平台不应该调用此函数
	return ErrUnsupportedPlatform
}

// restoreFromTrashLinuxImpl Linux 平台的恢复实现存根
func restoreFromTrashLinuxImpl(pattern string, opts RestoreOptions) error {
	return ErrUnsupportedPlatform
}

// listLinuxTrashItems 列出 Linux 回收站项目存根
func listLinuxTrashItems() ([]RecycleBinItem, error) {
	return []RecycleBinItem{}, ErrUnsupportedPlatform
}
