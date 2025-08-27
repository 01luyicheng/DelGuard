//go:build !windows
// +build !windows

package main

import (
	"syscall"
)

// DiskUsage 磁盘使用情况
type DiskUsage struct {
	Total uint64 // 总空间（字节）
	Free  uint64 // 可用空间（字节）
	Used  uint64 // 已用空间（字节）
}

func getDiskUsageUnix(path string) (*DiskUsage, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return nil, err
	}

	total := uint64(stat.Blocks) * uint64(stat.Bsize)
	free := uint64(stat.Bavail) * uint64(stat.Bsize)
	used := total - free

	return &DiskUsage{
		Total: total,
		Free:  free,
		Used:  used,
	}, nil
}

func getDiskUsage(path string) (*DiskUsage, error) {
	return getDiskUsageUnix(path)
}
