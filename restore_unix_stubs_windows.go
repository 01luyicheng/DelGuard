//go:build windows
// +build windows

package main

// Windows 平台下为 macOS/Linux 的恢复相关函数提供存根，以便在 Windows 上编译通过。

func restoreFromTrashMacOSImpl(pattern string, opts RestoreOptions) error {
	return ErrUnsupportedPlatform
}

func restoreFromTrashLinuxImpl(pattern string, opts RestoreOptions) error {
	return ErrUnsupportedPlatform
}

func listMacOSTrashItems() ([]RecycleBinItem, error) {
	return nil, ErrUnsupportedPlatform
}

func listLinuxTrashItems() ([]RecycleBinItem, error) {
	return nil, ErrUnsupportedPlatform
}
