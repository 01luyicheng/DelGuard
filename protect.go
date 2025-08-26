package main

import (
	"delguard/utils"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"
)

// sanitizeFileName 验证和清理文件名，防止路径遍历攻击
func sanitizeFileName(filename string) (string, error) {
	// 基本验证
	if filename == "" {
		return "", fmt.Errorf("文件名不能为空")
	}

	// Unicode 标准化防止绕过
	filename = normalizeUnicode(filename)

	// URL 解码防止编码绕过
	if decoded, err := url.QueryUnescape(filename); err == nil {
		filename = decoded
	}

	// 检查路径遍历（多种模式）
	if containsPathTraversal(filename) {
		return "", fmt.Errorf("检测到路径遍历攻击")
	}

	// 检查非法字符（但允许通配符和驱动器路径）
	if runtime.GOOS == "windows" {
		// Windows非法字符（不包括 * 和 ?，它们是合法的通配符）
		// 也不包括驱动器路径中的冒号（如 C:）
		if matched, _ := regexp.MatchString(`[<>"|]`, filename); matched {
			return "", fmt.Errorf("包含Windows非法字符")
		}

		// 检查是否有多个冒号（驱动器路径只能有一个冒号）
		colonCount := strings.Count(filename, ":")
		if colonCount > 1 {
			return "", fmt.Errorf("包含多个冒号")
		}

		// 如果有冒号，检查是否为有效的驱动器路径格式
		if colonCount == 1 {
			if !regexp.MustCompile(`^[a-zA-Z]:`).MatchString(filename) {
				return "", fmt.Errorf("无效的驱动器路径格式")
			}
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

// isDelGuardProject 检查路径是否为DelGuard项目目录
func isDelGuardProject(path string) bool {
	cleanPath := filepath.Clean(path)

	// 获取当前可执行文件的目录与路径（保留对自身的保护）
	if execPath, err := os.Executable(); err == nil {
		// 保护可执行文件本身
		if strings.EqualFold(cleanPath, filepath.Clean(execPath)) {
			return true
		}
	}

	// 定义核心源文件集合
	coreFiles := []string{"main.go", "config.go", "protect.go"}

	// helper：判断某个目录是否为DelGuard项目目录（包含核心文件）
	isProjectDir := func(dir string) bool {
		for _, f := range coreFiles {
			if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
				return false
			}
		}
		return true
	}

	// 如果传入的是目录：判断该目录是否为项目目录
	if info, err := os.Stat(cleanPath); err == nil && info.IsDir() {
		if isProjectDir(cleanPath) {
			return true
		}
	}

	// 如果传入的是文件：判断父目录是否为项目目录
	parent := filepath.Dir(cleanPath)
	if parent != "" && parent != "." {
		if isProjectDir(parent) {
			return true
		}
	}

	// 核心可执行名保护（当传入的是文件名本身时）
	basename := filepath.Base(cleanPath)
	if strings.EqualFold(basename, "delguard.exe") ||
		strings.EqualFold(basename, "delguard") ||
		strings.EqualFold(basename, "DelGuard.exe") ||
		strings.EqualFold(basename, "DelGuard") {
		return true
	}

	return false
}

// isTrashDirectory 检查是否为回收站目录
func isTrashDirectory(path string) bool {
	cleanPath := filepath.Clean(strings.ToLower(path))

	// 常见的回收站目录名称
	trashNames := []string{
		"recycle", "recycled", "recycler", "$recycle.bin",
		"trash", "trashes", ".trash", ".trashes",
		"wastebasket", "bin", ".bin",
	}

	baseName := strings.ToLower(filepath.Base(cleanPath))
	for _, trashName := range trashNames {
		if baseName == trashName {
			return true
		}
	}

	// 检查路径中是否包含回收站关键词
	for _, trashName := range trashNames {
		if strings.Contains(cleanPath, trashName) {
			return true
		}
	}

	// 平台特定的回收站检查
	switch runtime.GOOS {
	case "windows":
		// Windows回收站路径
		if strings.Contains(cleanPath, "$recycle.bin") ||
			strings.Contains(cleanPath, "recycler") ||
			strings.Contains(cleanPath, "recycled") {
			return true
		}
	case "darwin":
		// macOS回收站
		if strings.Contains(cleanPath, ".trash") ||
			strings.Contains(cleanPath, "/.trashes") {
			return true
		}
	case "linux":
		// Linux回收站
		if strings.Contains(cleanPath, ".local/share/trash") ||
			strings.Contains(cleanPath, "/.trash") {
			return true
		}
	}

	return false
}

// checkCriticalProtection 检查关键文件保护
func checkCriticalProtection(path string, force bool) error {
	cleanPath := filepath.Clean(path)

	// 1. 检查DelGuard项目保护
	if isDelGuardProject(cleanPath) {
		if !force {
			return fmt.Errorf("检测到DelGuard项目文件: %s\n为了安全，默认不允许删除DelGuard项目文件\n如果确实需要删除，请使用 --force 参数", cleanPath)
		}
		// 强制模式下给出警告
		fmt.Printf(T("⚠️  警告：正在删除DelGuard项目文件: %s\n"), cleanPath)
		if !confirmDangerousOperation("确定要删除DelGuard项目文件吗") {
			return fmt.Errorf("用户取消删除DelGuard项目文件")
		}
	}

	// 2. 检查回收站目录保护
	if isTrashDirectory(cleanPath) {
		if !force {
			return fmt.Errorf("检测到回收站/废纸篓目录: %s\n为了防止数据丢失，默认不允许直接删除回收站目录\n如果需要清空回收站，请使用系统自带的清空功能\n如果确实需要删除，请使用 --force 参数", cleanPath)
		}
		// 强制模式下给出警告
		fmt.Printf(T("⚠️  警告：正在删除回收站/废纸篓目录: %s\n"), cleanPath)
		fmt.Printf(T("警告：这将永久性删除回收站中的所有文件！\n"))
		if !confirmDangerousOperation("确定要删除回收站目录吗") {
			return fmt.Errorf("用户取消删除回收站目录")
		}
	}

	// 3. 检查系统关键文件（使用现有的函数）
	info, err := os.Stat(cleanPath)
	if err == nil {
		if isSpecialFile(info, cleanPath) {
			if !force {
				return fmt.Errorf("检测到系统关键文件: %s\n为了防止系统损坏，默认不允许删除系统关键文件\n如果确实需要删除，请使用 --force 参数", cleanPath)
			}
			// 强制模式下给出警告
			fmt.Printf(T("⚠️  警告：正在删除系统关键文件: %s\n"), cleanPath)
			if !confirmDangerousOperation("确定要删除系统关键文件吗") {
				return fmt.Errorf("用户取消删除系统关键文件")
			}
		}
	}

	return nil
}

// confirmDangerousOperation 危险操作确认
func confirmDangerousOperation(message string) bool {
	fmt.Printf(T("%s (y/N): "), message)
	var input string
	if isStdinInteractive() {
		if s, ok := readLineWithTimeout(20 * time.Second); ok {
			input = strings.ToLower(strings.TrimSpace(s))
		}
	}
	return input == "y" || input == "yes"
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
	// 检查Windows系统关键文件（仅保护真正重要的系统目录）
	systemDrive := os.Getenv("SYSTEMDRIVE")
	if systemDrive == "" {
		systemDrive = "C:"
	}
	criticalPaths := []string{
		filepath.Join(systemDrive, "Windows", "System32"),
		filepath.Join(systemDrive, "Windows", "SysWOW64"),
		filepath.Join(systemDrive, "Windows", "Boot"),
		filepath.Join(systemDrive, "Windows", "Fonts"),
		filepath.Join(systemDrive, "Program Files", "Windows NT"),
		filepath.Join(systemDrive, "ProgramData", "Microsoft", "Windows"),
	}

	cleanPath := filepath.Clean(strings.ToLower(path))
	for _, critical := range criticalPaths {
		// 更准确地检查路径前缀
		criticalClean := filepath.Clean(strings.ToLower(critical))
		if cleanPath == criticalClean || strings.HasPrefix(cleanPath, criticalClean+string(filepath.Separator)) {
			return true
		}
	}

	// 检查是否为系统启动文件
	if strings.HasSuffix(strings.ToLower(path), "bootmgr") ||
		strings.HasSuffix(strings.ToLower(path), "ntldr") ||
		strings.HasSuffix(strings.ToLower(path), "boot.ini") {
		return true
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
		return fmt.Errorf("文件过大，超过限制 %s", utils.FormatBytes(maxFileSize))
	}

	return nil
}

// checkFilePermissions 检查文件权限
func checkFilePermissions(path string, info os.FileInfo) error {
	// 对于目录，检查是否可以访问
	if info.IsDir() {
		// 检查目录访问权限
		entries, err := os.ReadDir(path)
		if err != nil {
			return fmt.Errorf("无法访问目录: %v", err)
		}
		_ = entries // 避免未使用变量警告
	} else {
		// 对于文件，检查是否可以打开
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("无法访问文件: %v", err)
		}
		file.Close()
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
	// 检查文件/目录是否存在和可访问
	info, err := os.Stat(path)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("Windows文件权限不足: %v", err)
		}
		return fmt.Errorf("无法访问文件: %v", err)
	}

	// 对于目录，检查是否可以列出内容
	if info.IsDir() {
		_, err := os.ReadDir(path)
		if err != nil {
			if os.IsPermission(err) {
				return fmt.Errorf("Windows目录权限不足: %v", err)
			}
			return fmt.Errorf("无法访问目录: %v", err)
		}
	} else {
		// 对于文件，尝试以只读模式打开
		file, err := os.Open(path)
		if err != nil {
			if os.IsPermission(err) {
				return fmt.Errorf("Windows文件权限不足: %v", err)
			}
			return fmt.Errorf("无法访问文件: %v", err)
		}
		file.Close()
	}
	return nil
}

// confirmHiddenFileDeletion 确认删除隐藏文件
func confirmHiddenFileDeletion(path string) bool {
	// 如果配置允许删除隐藏文件，则不需要确认
	config, _ := LoadConfig()
	if config.EnableHiddenCheck {
		fmt.Printf("⚠️  检测到隐藏文件: %s\n", path)
		fmt.Print("是否确认删除? [y/N]: ")
		var input string
		if isStdinInteractive() {
			if s, ok := readLineWithTimeout(20 * time.Second); ok {
				input = strings.TrimSpace(strings.ToLower(s))
			}
		}

		return input == "y" || input == "yes"
	}

	return true
}

// IsCriticalPath 检查是否为关键系统路径
func IsCriticalPath(path string) bool {
	cleanPath := filepath.Clean(path)

	// Windows关键路径
	if runtime.GOOS == "windows" {
		systemDrive := os.Getenv("SYSTEMDRIVE")
		if systemDrive == "" {
			systemDrive = "C:"
		}
		criticalPaths := []string{
			filepath.Join(systemDrive, "Windows"),
			filepath.Join(systemDrive, "Program Files"),
			filepath.Join(systemDrive, "Program Files (x86)"),
			filepath.Join(systemDrive, "ProgramData"),
			filepath.Join(systemDrive, "System Volume Information"),
			filepath.Join(systemDrive, "Recovery"),
			filepath.Join(systemDrive, "$Recycle.Bin"),
			os.Getenv("SYSTEMROOT"),
			os.Getenv("PROGRAMFILES"),
			os.Getenv("PROGRAMFILES(X86)"),
			os.Getenv("PROGRAMDATA"),
			os.Getenv("WINDIR"),
			filepath.Join(systemDrive, "Windows"),
		}

		// 添加用户特定的关键目录
		if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
			criticalPaths = append(criticalPaths,
				filepath.Join(userProfile, "AppData", "Roaming", "Microsoft", "Windows", "Start Menu"),
				filepath.Join(userProfile, "AppData", "Local", "Microsoft", "Windows"),
				filepath.Join(userProfile, "NTUSER.DAT"),
			)
		}

		for _, critical := range criticalPaths {
			if critical != "" && strings.HasPrefix(strings.ToLower(cleanPath), strings.ToLower(filepath.Clean(critical))) {
				return true
			}
		}
	} else if runtime.GOOS == "linux" {
		// Linux系统关键路径（包括现代应用路径）
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
			"/boot",
			"/dev",
			"/proc",
			"/sys",
			"/run",
			"/var/lib/dpkg",
			"/var/lib/rpm",
			"/var/lib/pacman",
			// 现代应用路径
			"/snap",
			"/var/lib/snapd",
			"/var/lib/flatpak",
			"/usr/share/applications",
			"/usr/local/share/applications",
			"/opt",
			// AppImage目录
			"/usr/bin/appimaged",
			// 容器目录
			"/var/lib/docker",
			"/var/lib/containerd",
			"/var/lib/podman",
		}

		// 添加用户目录中的关键路径
		if homeDir := os.Getenv("HOME"); homeDir != "" {
			criticalPaths = append(criticalPaths,
				homeDir+"/.local/share/flatpak",
				homeDir+"/.local/share/applications",
				homeDir+"/.config",
				homeDir+"/.ssh",
				homeDir+"/.gnupg",
			)
		}

		for _, critical := range criticalPaths {
			if strings.HasPrefix(cleanPath, critical) {
				return true
			}
		}
	} else if runtime.GOOS == "darwin" {
		// macOS系统关键路径（包括现代系统目录）
		criticalPaths := []string{
			"/System",
			"/Applications",
			"/bin",
			"/sbin",
			"/usr/bin",
			"/usr/sbin",
			"/etc",
			"/lib",
			"/usr/lib",
			"/private",
			"/var",
			"/tmp",
			// 现代macOS系统目录
			"/System/Library",
			"/System/Applications",
			"/System/DriverKit",
			"/System/iOSSupport",
			"/System/Volumes",
			"/Library/Application Support",
			"/Library/LaunchAgents",
			"/Library/LaunchDaemons",
			"/Library/Preferences",
			"/Library/Security",
			"/Library/SystemMigration",
			// Homebrew目录
			"/usr/local/Cellar",
			"/usr/local/Homebrew",
			"/opt/homebrew",
			// MacPorts目录
			"/opt/local",
		}

		// 添加用户目录中的关键路径
		if homeDir := os.Getenv("HOME"); homeDir != "" {
			criticalPaths = append(criticalPaths,
				homeDir+"/Library/Preferences",
				homeDir+"/Library/Application Support",
				homeDir+"/Library/Keychains",
				homeDir+"/.ssh",
				homeDir+"/.gnupg",
			)
		}

		for _, critical := range criticalPaths {
			if strings.HasPrefix(cleanPath, critical) {
				return true
			}
		}
	}

	// 检查是否为当前可执行文件本身（不包括整个目录）
	if exe, err := os.Executable(); err == nil {
		if strings.EqualFold(cleanPath, filepath.Clean(exe)) {
			return true
		}
	}

	return false
}

// ConfirmCritical 确认删除关键路径
func ConfirmCritical(path string) bool {
	fmt.Printf("🚨 警告: 检测到关键系统路径: %s\n", path)
	fmt.Print("删除关键系统文件可能导致系统不稳定或无法启动!\n是否确认删除? 输入 'DELETE' 确认: ")
	var input string
	if isStdinInteractive() {
		if s, ok := readLineWithTimeout(30 * time.Second); ok {
			input = strings.TrimSpace(s)
		}
	}

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

// normalizeUnicode 标准化Unicode字符串防止绕过
func normalizeUnicode(s string) string {
	// 简单的Unicode标准化，去除不可见字符
	var result strings.Builder
	for _, r := range s {
		// 过滤控制字符和不可见字符
		if r >= 32 && r != 127 {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// containsPathTraversal 检测路径遍历攻击模式
func containsPathTraversal(path string) bool {
	// 检查各种路径遍历模式
	traversalPatterns := []string{
		"..",
		".." + string(filepath.Separator),
		"%2e%2e",         // URL编码
		"%252e%252e",     // 双重编码
		"\\u002e\\u002e", // Unicode编码
		"\\x2e\\x2e",     // 十六进制编码
	}

	lowerPath := strings.ToLower(path)
	for _, pattern := range traversalPatterns {
		if strings.Contains(lowerPath, pattern) {
			return true
		}
	}

	// 检查路径组件
	parts := strings.Split(filepath.Clean(path), string(filepath.Separator))
	for _, part := range parts {
		if part == ".." {
			return true
		}
	}

	return false
}
