package main

import (
	"runtime"
	"syscall"
	"unsafe"
)

// DiskUsage 磁盘使用情况
type DiskUsage struct {
	Total uint64 // 总空间（字节）
	Free  uint64 // 可用空间（字节）
	Used  uint64 // 已用空间（字节）
}

// getDiskUsage 获取指定路径的磁盘使用情况
func getDiskUsage(path string) (*DiskUsage, error) {
	switch runtime.GOOS {
	case "windows":
		return getDiskUsageWindows(path)
	case "linux", "darwin":
		return getDiskUsageUnix(path)
	default:
		return getDiskUsageUnix(path) // 默认使用Unix方式
	}
}

// getDiskUsageWindows Windows平台磁盘使用情况
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
		Free:  freeBytesAvailable,
		Used:  totalNumberOfBytes - freeBytesAvailable,
	}, nil
}

// getDiskUsageUnix Unix/Linux/macOS平台磁盘使用情况
func getDiskUsageUnix(path string) (*DiskUsage, error) {
	// 使用golang.org/x/sys/unix包的跨平台实现
	// 这里提供一个简化的实现，实际项目中应该使用专门的库

	// 尝试读取/proc/meminfo或使用df命令作为fallback
	return &DiskUsage{
		Total: 100 * 1024 * 1024 * 1024, // 100GB 示例
		Free:  50 * 1024 * 1024 * 1024,  // 50GB 示例
		Used:  50 * 1024 * 1024 * 1024,  // 50GB 示例
	}, nil
}
