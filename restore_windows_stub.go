//go:build !windows
// +build !windows

package main

// 非 Windows 平台的占位实现，避免未定义符号导致的编译错误。
func restoreFromTrashWindows(pattern string, opts RestoreOptions) error {
	return ErrUnsupportedPlatform
}

// 非 Windows 平台的占位实现，列出回收站项目（Windows 专用接口桩）
func listRecycleBinItemsWindows() ([]RecycleBinItem, error) {
	return nil, ErrUnsupportedPlatform
}
