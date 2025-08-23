package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"unicode/utf8"
)

// sanitizeFileName 验证和清理文件名，防止路径遍历攻击
func sanitizeFileName(filename string) (string, error) {
	// 基本验证
	if filename == "" {
		return "", fmt.Errorf("文件名不能为空")
	}

	// 检查路径遍历
	if strings.Contains(filename, "..") {
		return "", fmt.Errorf("不允许路径遍历")
	}

	// 检查非法字符
	if runtime.GOOS == "windows" {
		// Windows非法字符
		if matched, _ := regexp.MatchString(`[<>:"|?*]`, filename); matched {
			return "", fmt.Errorf("包含Windows非法字符")
		}

		// Windows保留名称
		reservedNames := []string{
			"CON", "PRN", "AUX", "NUL",
			"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
			"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
		}
		name := strings.ToUpper(strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename)))
		for _, reserved := range reservedNames {
			if name == reserved {
				return "", fmt.Errorf("使用了Windows保留名称: %s", reserved)
			}
		}
	}

	// 检查控制字符
	for _, r := range filename {
		if r < 32 {
			return "", fmt.Errorf("包含控制字符")
		}
	}

	// 检查UTF-8编码
	if !utf8.ValidString(filename) {
		return "", fmt.Errorf("文件名编码无效")
	}

	// 检查路径长度
	if len(filename) > 260 {
		// Windows MAX_PATH 限制
		if runtime.GOOS == "windows" {
			return "", fmt.Errorf("文件名过长，超过Windows MAX_PATH限制")
		}
	}

	return filename, nil
}

// isSpecialFile 检查是否为特殊文件类型
func isSpecialFile(fileInfo os.FileInfo, path string) bool {
	mode := fileInfo.Mode()

	// 检查特殊文件类型
	if mode&os.ModeSymlink != 0 || // 符号链接
		mode&os.ModeDevice != 0 || // 设备文件
		mode&os.ModeSocket != 0 || // 套接字文件
		mode&os.ModeNamedPipe != 0 || // 命名管道
		mode&os.ModeCharDevice != 0 || // 字符设备
		mode&os.ModeIrregular != 0 { // 不规则文件
		return true
	}

	// 检查是否为挂载点或根目录
	if fileInfo.IsDir() {
		if isMountPoint(path) || isRootDirectory(path) {
			return true
		}
	}

	// 检查Windows特殊文件
	if runtime.GOOS == "windows" {
		return isWindowsSpecialFile(path)
	}

	return false
}

// isRootDirectory 检查是否为根目录
func isRootDirectory(path string) bool {
	cleanPath := filepath.Clean(path)

	// Unix系统根目录
	if runtime.GOOS != "windows" && cleanPath == "/" {
		return true
	}

	// Windows系统根目录
	if runtime.GOOS == "windows" {
		// 检查驱动器根目录，如 C:\
		if len(cleanPath) == 3 && cleanPath[1] == ':' && (cleanPath[2] == '\\' || cleanPath[2] == '/') {
			return true
		}
		// 检查UNC路径根目录
		if strings.HasPrefix(cleanPath, `\\`) {
			parts := strings.Split(cleanPath[2:], `\`)
			if len(parts) <= 2 {
				return true
			}
		}
	}

	return false
}

// isWindowsSpecialFile 检查Windows特殊文件
func isWindowsSpecialFile(path string) bool {
	// 检查Windows系统关键文件
	criticalPaths := []string{
		`C:\Windows`,
		`C:\Program Files`,
		`C:\Program Files (x86)`,
		`C:\ProgramData`,
		`C:\Users`,
	}

	cleanPath := filepath.Clean(strings.ToLower(path))
	for _, critical := range criticalPaths {
		// 更准确地检查路径前缀
		criticalClean := filepath.Clean(strings.ToLower(critical))
		if cleanPath == criticalClean || strings.HasPrefix(cleanPath, criticalClean+string(filepath.Separator)) {
			return true
		}
	}

	return false
}

// isMountPoint 检查是否为挂载点
func isMountPoint(path string) bool {
	// 在Unix系统上检查是否为挂载点
	if runtime.GOOS != "windows" {
		// 简化实现，实际应该检查 /proc/mounts 或使用系统调用
		// 这里仅检查一些常见的挂载点
		mountPoints := []string{"/", "/proc", "/sys", "/dev"}
		for _, mp := range mountPoints {
			if filepath.Clean(path) == filepath.Clean(mp) {
				return true
			}
		}
	} else {
		// Windows上检查驱动器根目录
		if len(path) == 3 && path[1] == ':' && (path[2] == '\\' || path[2] == '/') {
			return true
		}
	}
	return false
}

// checkFileSize 检查文件大小是否在允许范围内
func checkFileSize(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	// 检查是否为目录
	if info.IsDir() {
		return nil // 目录大小检查跳过
	}

	// 从配置获取最大文件大小限制
	config, _ := LoadConfig()
	maxFileSize := config.GetMaxFileSize()

	if info.Size() > maxFileSize {
		return fmt.Errorf("文件过大，超过限制 %s", formatBytes(maxFileSize))
	}

	return nil
}

// checkFilePermissions 检查文件权限
func checkFilePermissions(path string, info os.FileInfo) error {
	// 检查是否具有读取权限（至少需要读取权限才能安全删除）
	_, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("无读取权限: %v", err)
	}

	// 检查写入权限
	if runtime.GOOS != "windows" {
		// Unix系统检查写权限
		if info.Mode()&0200 == 0 {
			// 文件所有者没有写权限，简化处理
			return fmt.Errorf("文件所有者无写权限")
		}
	} else {
		// Windows系统权限检查
		return checkWindowsFilePermissions(path)
	}

	return nil
}

// checkWindowsFilePermissions 检查Windows文件权限
func checkWindowsFilePermissions(path string) error {
	// 简化实现，Windows文件权限检查复杂
	// 这里使用基本的文件访问检查
	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("Windows文件权限不足: %v", err)
		}
		return fmt.Errorf("无法访问文件: %v", err)
	}
	file.Close()
	return nil
}

// confirmHiddenFileDeletion 确认删除隐藏文件
func confirmHiddenFileDeletion(path string) bool {
	// 如果配置允许删除隐藏文件，则不需要确认
	config, _ := LoadConfig()
	if config.EnableHiddenCheck {
		fmt.Printf("⚠️  检测到隐藏文件: %s\n", path)
		fmt.Print("是否确认删除? [y/N]: ")

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		return input == "y" || input == "yes"
	}

	return true
}

// IsCriticalPath 检查是否为关键系统路径
func IsCriticalPath(path string) bool {
	cleanPath := filepath.Clean(path)

	// Windows关键路径
	if runtime.GOOS == "windows" {
		criticalPaths := []string{
			"C:\\Windows",
			"C:\\Program Files",
			"C:\\Program Files (x86)",
			os.Getenv("SYSTEMROOT"),
			os.Getenv("PROGRAMFILES"),
			os.Getenv("PROGRAMFILES(X86)"),
		}

		for _, critical := range criticalPaths {
			if critical != "" && strings.HasPrefix(cleanPath, filepath.Clean(critical)) {
				return true
			}
		}
	} else {
		// Unix系统关键路径
		criticalPaths := []string{
			"/bin",
			"/sbin",
			"/usr/bin",
			"/usr/sbin",
			"/etc",
			"/lib",
			"/lib64",
			"/usr/lib",
			"/usr/lib64",
			"/System",       // macOS
			"/Applications", // macOS
		}

		for _, critical := range criticalPaths {
			if strings.HasPrefix(cleanPath, critical) {
				return true
			}
		}
	}

	// 检查是否包含当前可执行文件
	if exe, err := os.Executable(); err == nil {
		if strings.HasPrefix(cleanPath, filepath.Dir(exe)) {
			return true
		}
	}

	return false
}

// ConfirmCritical 确认删除关键路径
func ConfirmCritical(path string) bool {
	fmt.Printf("🚨 警告: 检测到关键系统路径: %s\n", path)
	fmt.Print("删除关键系统文件可能导致系统不稳定或无法启动!\n是否确认删除? 输入 'DELETE' 确认: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	return input == "DELETE"
}

// CheckDeletePermission 检查删除权限
func CheckDeletePermission(path string) error {
	// 检查是否为关键路径
	if IsCriticalPath(path) {
		if !ConfirmCritical(path) {
			return fmt.Errorf("用户取消删除关键路径")
		}
	}

	// 检查文件权限
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("无法获取文件信息: %v", err)
	}

	if err := checkFilePermissions(path, info); err != nil {
		return fmt.Errorf("权限检查失败: %v", err)
	}

	return nil
}
