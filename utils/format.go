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

// CopyFile 统一的文件复制函数
//
// 参数:
//   - src: 源文件路径
//   - dst: 目标文件路径
//
// 返回值:
//   - error: 复制过程中的错误
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// 同步文件系统，确保数据写入磁盘
	return destFile.Sync()
}
