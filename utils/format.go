// Package utils 提供通用的工具函数
package utils

import (
	"fmt"
	"io"
	"os"
)

// FormatBytes 格式化字节数为人类可读的字符串
//
// 参数:
//   - bytes: 字节数
//
// 返回值:
//   - string: 格式化后的字符串
func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// CopyFile 复制文件从源到目标
func CopyFile(src, dst string) error {
	// 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("无法打开源文件: %v", err)
	}
	defer srcFile.Close()

	// 创建目标文件
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("无法创建目标文件: %v", err)
	}
	defer dstFile.Close()

	// 执行复制
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("复制失败: %v", err)
	}

	// 确保数据写入磁盘
	err = dstFile.Sync()
	if err != nil {
		return fmt.Errorf("同步文件失败: %v", err)
	}

	return nil
}
