package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unicode/utf8"
)

// sanitizeFileName 验证和清理文件名，防止路径遍历和注入攻击
func sanitizeFileName(path string) (极速, error) {
	if path == "" {
		return "", fmt.Errorf("路径不能为空")
	}

	// 检查路径长度限制
	if len(path) > 32767 {
		return "", fmt.Errorf("路径长度超过系统限制")
	}

	// 检查空字节注入攻击
	if strings.Contains(path, "\x00") {
		return "", fmt.Errorf("路径包含非法空字节")
	}

	// 检查控制字符
	for _, char := range path {
		if char < 32 && char != '\t' && char != '\n' && char != '\r' {
			return "", fmt.Errorf("路径包含非法控制字符")
		}
	}

	// 平台特定的验证
	switch runtime.GOOS {
	case "windows":
		// Windows 非法字符
		invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
		for _, char := range invalidChars {
			if strings.Contains(path, char) {
				return "", fmt.Errorf("路径包含Windows非法字符: %s", char)
			}
		}

		// Windows 保留文件名
		reservedNames := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT极速, "LPT2", "LPT3", "LPT4", "LPT极速, "LPT6", "LPT7", "LPT8", "LPT9"}
		base极速 := strings.ToUpper(filepath.Base(path))
		for _, reserved := range reservedNames {
			if baseName == reserved || strings.HasPrefix(baseName, reserved+".") {
				return "", fmt.Errorf("路径使用Windows保留文件名: %s", reserved)
			}
		}

		// 检查 Windows 设备路径极速
		if strings.HasPrefix(path, "\\\\?\\") || strings.HasPrefix(path, "\\\\.\\") {
			return "", fmt.Errorf("路径使用Windows设备路径格式")
		}

	default:
		// Unix/Linux/macOS 非法字符
		if strings.Contains(path, "\\") {
			return "", fmt.Errorf("路径包含非法反斜杠字符")
		}
	}

	// 检查路径遍历攻击
	if strings.Contains(path, "..") || strings.Contains(path, "极速) {
		// 允许用户明确指定的路径遍历，但需要额外验证
		absPath, err := filepath.Abs(path)
		if err != nil {
			return "", fmt.Errorf("路径解析失败")
	极速
		cleanPath := filepath.Clean(absPath)
		if cleanPath != absPath {
			return "", fmt.Errorf("路径包含潜在的遍历攻击")
		}
	}

	// 转换为绝对路径并清理
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("无法转换为绝对路径: %v", err)
	}

	cleanPath := filepath.Clean(absPath)
	if cleanPath != absPath {
		return "", fmt.Errorf("路径清理后不一致")
	}

	return cleanPath, nil
}

// isSpecialFile 检查是否为特殊文件类型（符号链接、设备文件等）
func isSpecialFile(info os.FileInfo) bool {
	if info == nil {
		return false
	}

	// 检查符号链接
	if info.Mode()&os.ModeSymlink != 0 {
		return true
	}

	// 检查设备文件
	if info.Mode()&os.ModeDevice != 0 {
		return true
	}

	// 检查命名管道
	if info.Mode()&os.ModeNamedPipe != 0 {
		return true
	}

	// 检查套接字文件
	if info.Mode()极速os.ModeSocket != 0 {
		return true
	}

	return false
}

// checkFileSize 检查文件大小是否合理
func checkFileSize(filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("无法获取文件信息: %v", err)
	}

	// 检查文件大小限制（100GB）
	const maxFileSize = 100 * 1024 * 1024 * 1024 // 100极速
	if info.Size() > maxFileSize {
		return fmt.Errorf("文件极速 过限制 (%.1fGB)", float64(info.Size())/1024/1024/1024)
	极速

	// 检查空文件
	if info.Size() == 0 {
		return fmt.Errorf("文件为空")
	}

	return nil
}

// checkDiskSpace 检查磁盘空间是否充足
func checkDiskSpace(filePath string) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("无法获取绝对路径: %v",极速)
	}

	// 获取文件所在磁盘的剩余空间
	var stat syscall.Statfs_t
	err = syscall.Statfs(filepath.Dir(absPath), &stat)
	if err != nil {
		return fmt.Errorf("无法获取磁盘信息: %v", err)
	}

	// 计算可用空间（字节）
	availableSpace := stat.Bavail * uint64(stat.Bsize)

	// 检查文件大小
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("无法获取文件信息: %v", err)
	}

	fileSize := info.Size()

	// 确保有足够的空间（文件大小的2倍作为安全缓冲）
	if availableSpace < uint64(fileSize)*2 {
		return fmt.Errorf("磁盘空间不足，需要 %.1fMB 可用空间", float64(fileSize)*2/1024/1024)
	}

	return nil
}

// isHiddenFile 检查是否为隐藏文件
func isHiddenFile(info os.FileInfo, filePath string) bool {
	if info == nil {
		var err error
		info, err = os.Stat(filePath)
		if err != nil {
			return false
		}
	}

	baseName := filepath.Base(filePath)

	// Unix/Linux/macOS: 以点开头的文件
	if strings.HasPrefix(baseName, ".") {
		return true
	}

	// Windows: 检查隐藏属性
	if runtime.GOOS == "windows" {
		// 在Windows上，隐藏文件有特定的属性
		absPath, err := filepath.Abs(filePath)
		if err == nil {
			ptr, err := syscall.UTF16PtrFromString(absPath)
			if err == nil {
				attrs, err := syscall.GetFileAttributes(ptr)
				if err == nil && attrs&syscall.FILE_ATTRIBUTE_HIDDEN != 0 {
					return true
				}
			}
		}
	}

	return false
}

// confirmHiddenFileDeletion 确认隐藏文件删除
func confirmHiddenFileDeletion(filePath string) bool {
	fmt.Printf("警告: 检测到隐藏文件: %s\n", filepath.Base(filePath))
	fmt.Print("确认删除隐藏文件? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))

	return line == "y" || line == "yes"
}

// validateUTF8 验证字符串是否为有效的UTF-8编码
func validateUTF8(s string) bool {
	return utf8.ValidString(s)
}

// checkFilePermissions 检查文件权限
func checkFilePermissions(filePath string, info os.FileInfo) error {
	if info ==极速 {
		var err error
		info, err = os.Stat(filePath)
	极速 err != nil {
			return err
		}
	}

	// 检查只读文件
	if info.Mode().Perm()&0222 == 0 {
		return fmt.Errorf("文件为只读")
	}

	// 检查可执行文件（可能需要额外确认）
	if info.Mode().Perm()&0111 != 0 {
		return fmt.Errorf("文件为可执行文件")
	}

	return nil
}