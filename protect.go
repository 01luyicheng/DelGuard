package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// IsCriticalPath checks if path is critical protected path
func IsCriticalPath(absPath string) bool {
	p := filepath.Clean(absPath)

	// Check if it's trash/recycle bin directory
	if IsTrashDirectory(p) {
		return true
	}

	// Check if contains DelGuard program directory
	if ContainsDelGuardDirectory(p) {
		return true
	}

	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "windows":
		// Windows drive roots
		if isWindowsDriveRoot(p) {
			return true
		}
		crit := []string{
			`C:\Windows`,
			`C:\Windows\System32`,
			`C:\Program Files`,
			`C:\Program Files (x86)`,
			`C:\ProgramData`,
		}
		if home != "" {
			crit = append(crit, filepath.Clean(home))
		}
		for _, c := range crit {
			if strings.EqualFold(p, filepath.Clean(c)) {
				return true
			}
		}
		return false
	case "darwin":
		crit := []string{
			"/", "/System", "/Library", "/Applications",
			"/usr", "/bin", "/sbin", "/opt",
		}
		if home != "" {
			crit = append(crit, filepath.Clean(home))
		}
		for _, c := range crit {
			if p == filepath.Clean(c) {
				return true
			}
		}
		return false
	default:
		crit := []string{
			"/", "/bin", "/sbin", "/usr", "/etc", "/root",
			"/lib", "/lib32", "/lib64", "/libx32", "/opt",
			"/var", "/boot", "/dev", "/proc", "/sys",
		}
		if home != "" {
			crit = append(crit, filepath.Clean(home))
		}
		for _, c := range crit {
			if p == filepath.Clean(c) {
				return true
			}
		}
		return false
	}
}

// IsTrashDirectory checks if path is trash/recycle bin directory
func IsTrashDirectory(path string) bool {
	cleanPath := filepath.Clean(path)

	switch runtime.GOOS {
	case "windows":
		trashPaths := []string{
			`C:\$Recycle.Bin`,
			`C:\Recycler`,
			`C:\RECYCLER`,
		}
		for _, trashPath := range trashPaths {
			if strings.EqualFold(cleanPath, filepath.Clean(trashPath)) {
				return true
			}
		}

		if userProfile, err := os.UserHomeDir(); err == nil {
			userTrash := filepath.Join(userProfile, "AppData", "Local", "Microsoft", "Windows", "Recycle")
			if strings.EqualFold(cleanPath, filepath.Clean(userTrash)) {
				return true
			}
		}

	case "darwin":
		if home, err := os.UserHomeDir(); err == nil {
			trashPaths := []string{
				filepath.Join(home, ".Trash"),
				filepath.Join(home, "Library", "Mobile Documents", "com~apple~CloudDocs", "Deleted Items"),
			}
			for _, trashPath := range trashPaths {
				if cleanPath == filepath.Clean(trashPath) {
					return true
				}
			}
		}

	default:
		if home, err := os.UserHomeDir(); err == nil {
			trashPaths := []string{
				filepath.Join(home, ".local", "share", "Trash"),
				filepath.Join(home, ".Trash"),
				filepath.Join(home, ".local", "trash"),
			}
			for _, trashPath := range trashPaths {
				if cleanPath == filepath.Clean(trashPath) {
					return true
				}
			}
		}
	}

	return false
}

// ContainsDelGuardDirectory checks if path contains DelGuard program directory
func ContainsDelGuardDirectory(path string) bool {
	cleanPath := filepath.Clean(path)

	exePath, err := os.Executable()
	if err != nil {
		return false
	}
	exeDir := filepath.Dir(filepath.Clean(exePath))

	if strings.HasPrefix(exeDir, cleanPath+string(filepath.Separator)) ||
		exeDir == cleanPath {
		return true
	}

	commonInstallDirs := []string{
		"/usr/local/bin",
		"/usr/bin",
		"/opt/delguard",
		"C:\\Program Files\\DelGuard",
		"C:\\Program Files (x86)\\DelGuard",
	}

	for _, installDir := range commonInstallDirs {
		if strings.HasPrefix(installDir, cleanPath+string(filepath.Separator)) ||
			installDir == cleanPath {
			return true
		}
	}

	return false
}

func isWindowsDriveRoot(p string) bool {
	if len(p) == 3 && p[1] == ':' && (p[2] == '\\' || p[2] == '/') &&
		((p[0] >= 'A' && p[0] <= 'Z') || (p[0] >= 'a' && p[0] <= 'z')) {
		return true
	}
	return false
}

// ConfirmCritical confirms critical path deletion
func ConfirmCritical(absPath string) bool {
	fmt.Printf("Warning: About to delete critical path: %s\n", absPath)
	fmt.Print("To confirm risk, enter full path to continue (or press Enter to cancel): ")
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(filepath.Clean(line), filepath.Clean(absPath))
	}
	return filepath.Clean(line) == filepath.Clean(absPath)
}

// CheckDeletePermission checks delete permissions
func CheckDeletePermission(filePath string) bool {
	// 首先检查文件是否存在和可访问
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	// 检查只读文件
	if info.Mode().Perm()&0222 == 0 {
		fmt.Printf("Warning: File %s is read-only\\n", filePath)
		fmt.Print("Confirm deletion of read-only file? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(strings.ToLower(line))
		if line != "y" && line != "yes" {
			return false
		}
	}

	// 检查管理员权限（需要额外确认）
	if IsElevated() {
		fmt.Printf("Warning: Running with admin/root privileges, about to delete: %s\n", filePath)
		fmt.Print("Confirm deletion? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(strings.ToLower(line))
		if line != "y" && line != "yes" {
			return false
		}
	}

	return true
}

// validatePath validates path with comprehensive security checks
func validatePath(path string) bool {
	if path == "" {
		return false
	}

	// 检查路径长度限制
	if len(path) > 32767 {
		return false
	}

	// 检查空字节注入攻击
	if strings.Contains(path, "\x00") {
		return false
	}

	// 检查路径遍历攻击
	if strings.Contains(path, "..") || strings.Contains(path, "~") {
		return false
	}

	// 检查控制字符
	for _, char := range path {
		if char < 32 && char != '\t' && char != '\n' && char != '\r' {
			return false
		}
	}

	switch runtime.GOOS {
	case "windows":
		// Windows 非法字符
		invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
		for _, char := range invalidChars {
			if strings.Contains(path, char) {
				return false
			}
		}

		// Windows 保留文件名
		reservedNames := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
		baseName := strings.ToUpper(filepath.Base(path))
		for _, reserved := range reservedNames {
			if baseName == reserved || strings.HasPrefix(baseName, reserved+".") {
				return false
			}
		}

		// 检查 Windows 设备路径
		if strings.HasPrefix(path, "\\\\?\\") || strings.HasPrefix(path, "\\\\.\\") {
			return false
		}

	default:
		// Unix/Linux/macOS 非法字符
		if strings.Contains(path, "\\") {
			return false
		}
	}

	// 检查隐藏文件/系统文件模式（防止误删）
	baseName := filepath.Base(path)
	if strings.HasPrefix(baseName, ".") && !isUserIntendedHiddenFile(path) {
		return false
	}

	// 检查路径是否为绝对路径（相对路径更安全）
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	// 最终清理检查
	cleanPath := filepath.Clean(absPath)
	if cleanPath != absPath {
		return false
	}

	return true
}

// isUserIntendedHiddenFile 检查是否为用户有意操作隐藏文件
func isUserIntendedHiddenFile(path string) bool {
	// 用户明确指定隐藏文件时才允许操作
	// 在实际使用中，这个函数应该结合用户交互确认
	return true // 暂时返回true，后续结合交互模式
}
