//go:build !darwin && !windows
// +build !darwin,!windows

package main

// 该文件为非 macOS 平台提供 macOS 特定函数的桩实现，
// 以避免在非 darwin 构建时出现未定义符号的编译错误。

// restoreFromTrashMacOSImpl 非 macOS 平台桩实现
func restoreFromTrashMacOSImpl(pattern string, opts RestoreOptions) error {
	return ErrUnsupportedPlatform
}

// listMacOSTrashItems 非 macOS 平台桩实现
func listMacOSTrashItems() ([]RecycleBinItem, error) {
	return nil, ErrUnsupportedPlatform
}

// moveToTrashMacOS 非 macOS 平台桩实现
func moveToTrashMacOS(filePath string) error {
	return ErrUnsupportedPlatform
}
