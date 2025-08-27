package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// DeleteOperation 删除操作结构体
type DeleteOperation struct {
	Path        string
	IsDirectory bool
	Size        int64
	Force       bool
	Recursive   bool
	Verbose     bool
}

// DeleteResult 删除结果
type DeleteResult struct {
	Path        string
	Success     bool
	Error       error
	Size        int64
	Duration    time.Duration
	Skipped     bool
	Reason      string
	IsDirectory bool
}

// CoreDeleter 核心删除器
type CoreDeleter struct {
	config       *Config
	smartParser  *SmartParser
	dryRun       bool
	interactive  bool
	preserveRoot bool
	force        bool
	recursive    bool
	verbose      bool
	stats        DeleteStats
}

// DeleteStats 删除统计信息
type DeleteStats struct {
	TotalFiles   int64
	TotalDirs    int64
	DeletedFiles int64
	DeletedDirs  int64
	SkippedFiles int64
	SkippedDirs  int64
	TotalSize    int64
	DeletedSize  int64
	Errors       int64
	StartTime    time.Time
	EndTime      time.Time
}

// NewCoreDeleter 创建核心删除器
func NewCoreDeleter(config *Config) *CoreDeleter {
	return &CoreDeleter{
		config:       config,
		smartParser:  NewSmartParser(),
		preserveRoot: true,
		stats:        DeleteStats{StartTime: time.Now()},
	}
}

// SetOptions 设置删除选项
func (cd *CoreDeleter) SetOptions(dryRun, interactive, force, recursive, verbose bool) {
	cd.dryRun = dryRun
	cd.interactive = interactive
	cd.force = force
	cd.recursive = recursive
	cd.verbose = verbose
}

// Delete 执行删除操作
func (cd *CoreDeleter) Delete(paths []string) []DeleteResult {
	var results []DeleteResult

	// 解析和验证路径
	parsedPaths, _ := cd.smartParser.ParseArguments(paths)

	for _, parsed := range parsedPaths {
		if parsed.Type == ArgTypeFile || parsed.Type == ArgTypeDirectory {
			result := cd.deleteSingle(parsed.NormalizedPath)
			results = append(results, result)
		}
	}

	cd.stats.EndTime = time.Now()
	return results
}

// deleteSingle 删除单个文件或目录
func (cd *CoreDeleter) deleteSingle(path string) DeleteResult {
	startTime := time.Now()
	result := DeleteResult{
		Path: path,
	}

	// 基本安全检查
	if err := cd.basicSafetyCheck(path); err != nil {
		result.Error = err
		result.Reason = "安全检查失败"
		cd.stats.Errors++
		return result
	}

	// 获取文件信息
	info, err := os.Stat(path)
	if err != nil {
		result.Error = fmt.Errorf("无法访问路径: %v", err)
		cd.stats.Errors++
		return result
	}

	result.Size = info.Size()
	result.IsDirectory = info.IsDir()

	// 交互式确认
	if cd.interactive && !cd.confirmDeletion(path, info) {
		result.Skipped = true
		result.Reason = "用户取消"
		if info.IsDir() {
			cd.stats.SkippedDirs++
		} else {
			cd.stats.SkippedFiles++
		}
		return result
	}

	// 干运行模式
	if cd.dryRun {
		result.Success = true
		result.Reason = "干运行模式"
		result.Duration = time.Since(startTime)
		return result
	}

	// 执行删除
	if info.IsDir() {
		err = cd.deleteDirectory(path)
		if err == nil {
			cd.stats.DeletedDirs++
		}
	} else {
		err = cd.deleteFile(path)
		if err == nil {
			cd.stats.DeletedFiles++
		}
	}

	result.Success = err == nil
	result.Error = err
	result.Duration = time.Since(startTime)

	if err != nil {
		cd.stats.Errors++
	} else {
		cd.stats.DeletedSize += result.Size
	}

	return result
}

// basicSafetyCheck 基本安全检查（简化版）
func (cd *CoreDeleter) basicSafetyCheck(path string) error {
	cleanPath := filepath.Clean(path)

	// 1. 检查是否为根目录
	if cd.isRootPath(cleanPath) {
		return fmt.Errorf("不允许删除根目录")
	}

	// 2. 检查是否为当前程序
	if cd.isSelfExecutable(cleanPath) {
		return fmt.Errorf("不允许删除程序自身")
	}

	// 3. 检查是否为重要系统目录
	if cd.isCriticalSystemPath(cleanPath) && !cd.force {
		return fmt.Errorf("检测到关键系统路径: %s", cleanPath)
	}

	// 4. 检查路径长度限制
	if len(cleanPath) > 4096 {
		return fmt.Errorf("路径过长")
	}

	// 5. 检查路径中的非法字符（Windows文件名非法字符，但路径中的冒号是合法的）
	if strings.ContainsAny(cleanPath, "<>\"|?*") {
		return fmt.Errorf("路径包含非法字符")
	}

	return nil
}

// isRootPath 检查是否为根路径
func (cd *CoreDeleter) isRootPath(path string) bool {
	cleanPath := filepath.Clean(path)

	if runtime.GOOS == "windows" {
		// Windows驱动器根目录 (C:\, D:\ 等)
		if len(cleanPath) == 3 && cleanPath[1] == ':' &&
			(cleanPath[2] == '\\' || cleanPath[2] == '/') {
			return true
		}
	} else {
		// Unix根目录
		if cleanPath == "/" {
			return true
		}
	}

	return false
}

// isSelfExecutable 检查是否为程序自身
func (cd *CoreDeleter) isSelfExecutable(path string) bool {
	if exe, err := os.Executable(); err == nil {
		return strings.EqualFold(filepath.Clean(path), filepath.Clean(exe))
	}
	return false
}

// isCriticalSystemPath 检查是否为关键系统路径（简化版）
func (cd *CoreDeleter) isCriticalSystemPath(path string) bool {
	cleanPath := strings.ToLower(filepath.Clean(path))

	var criticalPaths []string

	switch runtime.GOOS {
	case "windows":
		systemDrive := strings.ToLower(os.Getenv("SYSTEMDRIVE"))
		if systemDrive == "" {
			systemDrive = "c:"
		}
		criticalPaths = []string{
			filepath.Join(systemDrive, "windows", "system32"),
			filepath.Join(systemDrive, "windows", "syswow64"),
			filepath.Join(systemDrive, "windows", "boot"),
		}
	case "linux", "darwin":
		criticalPaths = []string{
			"/bin", "/sbin", "/usr/bin", "/usr/sbin",
			"/etc", "/lib", "/usr/lib", "/boot",
		}
	}

	for _, critical := range criticalPaths {
		if strings.HasPrefix(cleanPath, critical) {
			return true
		}
	}

	return false
}

// confirmDeletion 确认删除操作
func (cd *CoreDeleter) confirmDeletion(path string, info os.FileInfo) bool {
	fileType := "文件"
	if info.IsDir() {
		fileType = "目录"
	}

	fmt.Printf("确认删除%s: %s", fileType, path)
	if info.IsDir() {
		fmt.Printf(" (可能包含子项)")
	}
	fmt.Printf(" [y/N]: ")

	var input string
	if isStdinInteractive() {
		if s, ok := readLineWithTimeout(30 * time.Second); ok {
			input = strings.ToLower(strings.TrimSpace(s))
		}
	}

	return input == "y" || input == "yes"
}

// deleteFile 删除文件
func (cd *CoreDeleter) deleteFile(path string) error {
	if cd.verbose {
		fmt.Printf("删除文件: %s\n", path)
	}

	// 获取文件名用于提示
	filename := filepath.Base(path)

	// 根据配置选择删除方式
	if cd.config.UseRecycleBin {
		// 使用回收站删除
		if err := cd.moveToTrash(path); err != nil {
			// 提供友好的错误提示
			cd.showFriendlyError(err, path)
			return err
		}
		// 成功删除后显示提示
		fmt.Printf("DelGuard: [%s] 已被移动到回收站\n", filename)
	} else {
		// 永久删除
		err := os.Remove(path)
		if err != nil {
			cd.showFriendlyError(err, path)
			return fmt.Errorf("删除文件失败: %v", err)
		}
		// 永久删除提示
		fmt.Printf("DelGuard: [%s] 已被永久删除\n", filename)
	}

	return nil
}

// deleteDirectory 删除目录
func (cd *CoreDeleter) deleteDirectory(path string) error {
	if cd.verbose {
		fmt.Printf("删除目录: %s\n", path)
	}

	// 获取目录名用于提示
	dirname := filepath.Base(path)

	// 根据配置选择删除方式
	if cd.config.UseRecycleBin {
		// 使用回收站删除
		if err := cd.moveToTrash(path); err != nil {
			cd.showFriendlyError(err, path)
			return err
		}
		fmt.Printf("DelGuard: [%s] 目录已被移动到回收站\n", dirname)
	} else {
		// 永久删除
		if cd.recursive {
			err := os.RemoveAll(path)
			if err != nil {
				cd.showFriendlyError(err, path)
				return fmt.Errorf("递归删除目录失败: %v", err)
			}
		} else {
			// 非递归删除，只删除空目录
			err := os.Remove(path)
			if err != nil {
				cd.showFriendlyError(err, path)
				return fmt.Errorf("删除空目录失败: %v", err)
			}
		}
		fmt.Printf("DelGuard: [%s] 目录已被永久删除\n", dirname)
	}

	return nil
}

// moveToTrash 跨平台回收站删除
func (cd *CoreDeleter) moveToTrash(path string) error {
	switch runtime.GOOS {
	case "windows":
		return cd.moveToTrashWindows(path)
	case "darwin":
		return cd.moveToTrashMacOS(path)
	case "linux":
		return cd.moveToTrashLinux(path)
	default:
		// 不支持的平台，使用永久删除
		return os.Remove(path)
	}
}

// GetStats 获取删除统计信息
func (cd *CoreDeleter) GetStats() DeleteStats {
	return cd.stats
}

// PrintStats 打印统计信息
func (cd *CoreDeleter) PrintStats() {
	duration := cd.stats.EndTime.Sub(cd.stats.StartTime)

	fmt.Println("\n📊 删除操作统计:")
	fmt.Printf("⏱️  总耗时: %v\n", duration)
	fmt.Printf("📁 目录: 删除 %d, 跳过 %d\n", cd.stats.DeletedDirs, cd.stats.SkippedDirs)
	fmt.Printf("📄 文件: 删除 %d, 跳过 %d\n", cd.stats.DeletedFiles, cd.stats.SkippedFiles)

	if cd.stats.DeletedSize > 0 {
		fmt.Printf("💾 释放空间: %s\n", formatBytes(cd.stats.DeletedSize))
	}

	if cd.stats.Errors > 0 {
		fmt.Printf("❌ 错误数量: %d\n", cd.stats.Errors)
	}
}

// formatBytes 格式化字节数
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// showFriendlyError 显示友好的错误信息
func (cd *CoreDeleter) showFriendlyError(err error, path string) {
	filename := filepath.Base(path)
	
	// 根据错误类型提供不同的提示
	errMsg := err.Error()
	switch {
	case strings.Contains(errMsg, "permission") || strings.Contains(errMsg, "权限"):
		fmt.Printf("DelGuard: 无法删除 [%s] - 权限不足\n", filename)
		fmt.Println("建议：")
		fmt.Println("  1. 以管理员身份重新运行程序")
		fmt.Println("  2. 检查文件是否被其他程序占用")
		fmt.Println("  3. 确认您对该文件有删除权限")
	case strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "不存在"):
		fmt.Printf("DelGuard: 无法删除 [%s] - 文件不存在\n", filename)
		fmt.Println("建议：")
		fmt.Println("  1. 检查文件路径是否正确")
		fmt.Println("  2. 使用绝对路径或相对路径")
		fmt.Println("  3. 确认文件没有被移动或重命名")
	case strings.Contains(errMsg, "in use") || strings.Contains(errMsg, "被使用"):
		fmt.Printf("DelGuard: 无法删除 [%s] - 文件正在被使用\n", filename)
		fmt.Println("建议：")
		fmt.Println("  1. 关闭正在使用该文件的程序")
		fmt.Println("  2. 等待文件操作完成后重试")
		fmt.Println("  3. 重启电脑后再次尝试")
	case strings.Contains(errMsg, "path") || strings.Contains(errMsg, "路径"):
		fmt.Printf("DelGuard: 无法删除 [%s] - 路径问题\n", filename)
		fmt.Println("建议：")
		fmt.Println("  1. 检查路径中是否包含特殊字符")
		fmt.Println("  2. 确保路径长度不超过260个字符")
		fmt.Println("  3. 使用引号包围包含空格的路径")
	default:
		fmt.Printf("DelGuard: 无法删除 [%s] - %s\n", filename, errMsg)
		fmt.Println("建议：")
		fmt.Println("  1. 检查磁盘空间是否充足")
		fmt.Println("  2. 确认文件系统没有错误")
		fmt.Println("  3. 联系技术支持获取帮助")
	}
}
