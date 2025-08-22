//go:build !darwin && !linux && !windows
// +build !darwin,!linux,!windows

package main

// 在非主流平台上的存根函数
func moveToTrashMacOS(filePath string) error {
	return ErrUnsupportedPlatform
}

func moveToTrashLinux(filePath string) error {
	return ErrUnsupportedPlatform
}

func moveToTrashWindows(filePath string) error {
	return ErrUnsupportedPlatform
}
