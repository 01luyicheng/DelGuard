//go:build windows
// +build windows

package main

import (
	"syscall"
	"unsafe"
)

// DiskUsage 磁盘使用情况
type DiskUsage struct {
	Total uint64 // 总空间（字节）
	Free  uint64 // 可用空间（字节）
	Used  uint64 // 已用空间（字节）
}

func getDiskUsageWindows(path string) (*DiskUsage, error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpaceEx := kernel32.NewProc("GetDiskFreeSpaceExW")

	var freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes uint64

	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}

	ret, _, err := getDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalNumberOfBytes)),
		uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)),
	)

	if ret == 0 {
		return nil, err
	}

	return &DiskUsage{
		Total: totalNumberOfBytes,
		Free:  totalNumberOfFreeBytes,
		Used:  totalNumberOfBytes - totalNumberOfFreeBytes,
	}, nil
}

func getDiskUsage(path string) (*DiskUsage, error) {
	return getDiskUsageWindows(path)
}
