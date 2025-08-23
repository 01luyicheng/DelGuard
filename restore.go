package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// RestoreOptions 恢复选项
type RestoreOptions struct {
	Pattern     string
	MaxFiles    int
	Interactive bool
}

// EnhancedRestoreOptions 增强的恢复选项
type EnhancedRestoreOptions struct {
	SkipIntegrityCheck bool
	BackupExisting     bool
	MaxFileSize        int64
	AllowedExtensions  []string
	ScanForMalware     bool
}

// RecycleBinItem 回收站项目信息（跨平台定义）
type RecycleBinItem struct {
	OriginalPath string    `json:"original_path"`
	CurrentPath  string    `json:"current_path"`
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	DeletedTime  time.Time `json:"deleted_time"`
	FileType     string    `json:"file_type"`
}

// 恢复文件的错误类型
var (
	ErrRestoreNotSupported = errors.New("当前平台不支持恢复功能")
	ErrNoFilesToRestore    = errors.New("没有可恢复的文件")
	ErrRestoreFailed       = errors.New("恢复文件失败")
)

// restoreFromTrash 从回收站恢复文件
func restoreFromTrash(pattern string, opts RestoreOptions) error {
	switch runtime.GOOS {
	case "windows":
		return restoreFromTrashWindows(pattern, opts)
	case "darwin":
		return restoreFromTrashMacOS(pattern, opts)
	default:
		return restoreFromTrashLinux(pattern, opts)
	}
}

// restoreFromTrashWindows Windows平台恢复实现
func restoreFromTrashWindows(pattern string, opts RestoreOptions) error {
	return restoreFromTrashWindowsImpl(pattern, opts)
}

// restoreFromTrashMacOS macOS平台恢复实现
func restoreFromTrashMacOS(pattern string, opts RestoreOptions) error {
	return restoreFromTrashMacOSImpl(pattern, opts)
}

// restoreFromTrashMacOSImpl macOS平台恢复实现函数
func restoreFromTrashMacOSImpl(pattern string, opts RestoreOptions) error {
	// 对于macOS，使用restore_darwin.go中的实现
	return restoreFromTrashMacOSImpl(pattern, opts)
}

// restoreFromTrashLinux Linux平台恢复实现
func restoreFromTrashLinux(pattern string, opts RestoreOptions) error {
	return restoreFromTrashLinuxImpl(pattern, opts)
}

// restoreFromTrashLinuxImpl Linux平台恢复实现函数
func restoreFromTrashLinuxImpl(pattern string, opts RestoreOptions) error {
	// 对于Linux，使用restore_linux.go中的实现
	return restoreFromTrashLinuxImpl(pattern, opts)
}

// listRecoverableFiles 列出可恢复的文件
func listRecoverableFiles(pattern string) error {
	items, err := listRecycleBinItems()
	if err != nil {
		return err
	}

	if len(items) == 0 {
		fmt.Println("回收站中没有可恢复的文件")
		return nil
	}

	fmt.Printf("找到 %d 个可恢复的文件:\n\n", len(items))

	// 根据模式筛选
	var filtered []RecycleBinItem
	if pattern == "" {
		filtered = items
	} else {
		// 简单的字符串匹配
		for _, item := range items {
			if strings.Contains(strings.ToLower(item.Name), strings.ToLower(pattern)) {
				filtered = append(filtered, item)
			}
		}
	}

	if len(filtered) == 0 {
		fmt.Printf("没有找到匹配 '%s' 的文件\n", pattern)
		return nil
	}

	// 显示文件信息
	for i, item := range filtered {
		fmt.Printf("%d. %s\n", i+1, item.Name)
		fmt.Printf("   原始路径: %s\n", item.OriginalPath)
		fmt.Printf("   文件大小: %s\n", formatBytes(item.Size))
		fmt.Printf("   删除时间: %s\n", item.DeletedTime.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}

	return nil
}

// listRecycleBinItems 跨平台的回收站项目列表
func listRecycleBinItems() ([]RecycleBinItem, error) {
	switch runtime.GOOS {
	case "windows":
		return listRecycleBinItemsWindows()
	case "darwin":
		return listRecycleBinItemsMacOS()
	case "linux":
		return listRecycleBinItemsLinux()
	default:
		return nil, ErrRestoreNotSupported
	}
}

// listRecycleBinItemsMacOS 获取macOS回收站项目
func listRecycleBinItemsMacOS() ([]RecycleBinItem, error) {
	if runtime.GOOS != "darwin" {
		return nil, ErrUnsupportedPlatform
	}
	// 由于平台特定文件有构建标签，这里使用通用实现
	return []RecycleBinItem{}, nil
}

// listRecycleBinItemsLinux 获取Linux回收站项目
func listRecycleBinItemsLinux() ([]RecycleBinItem, error) {
	if runtime.GOOS != "linux" {
		return nil, ErrUnsupportedPlatform
	}
	// 由于平台特定文件有构建标签，这里使用通用实现
	return []RecycleBinItem{}, nil
}

// EnhancedRestore 执行增强恢复操作
func EnhancedRestore(pattern string, opts RestoreOptions, enhancedOpts EnhancedRestoreOptions) error {
	// 获取可恢复文件列表
	items, err := listRecycleBinItems()
	if err != nil {
		return fmt.Errorf("无法获取回收站文件列表: %v", err)
	}

	if len(items) == 0 {
		return ErrNoFilesToRestore
	}

	// 根据模式筛选
	var filtered []RecycleBinItem
	if pattern == "" {
		filtered = items
	} else {
		for _, item := range items {
			match, err := filepath.Match(pattern, item.Name)
			if err != nil {
				// 如果模式无效，则使用包含匹配
				if strings.Contains(item.Name, pattern) {
					filtered = append(filtered, item)
				}
			} else if match {
				filtered = append(filtered, item)
			}
		}
	}

	// 应用最大文件数限制
	if opts.MaxFiles > 0 && len(filtered) > opts.MaxFiles {
		filtered = filtered[:opts.MaxFiles]
		fmt.Printf("限制恢复文件数为 %d 个\n", opts.MaxFiles)
	}

	// 应用增强选项筛选
	var enhancedFiltered []RecycleBinItem
	for _, item := range filtered {
		// 检查文件大小限制
		if enhancedOpts.MaxFileSize > 0 && item.Size > enhancedOpts.MaxFileSize {
			fmt.Printf("跳过文件 %s (大小 %s 超过限制 %s)\n",
				item.Name, formatBytes(item.Size), formatBytes(enhancedOpts.MaxFileSize))
			continue
		}

		// 检查文件扩展名
		if len(enhancedOpts.AllowedExtensions) > 0 {
			allowed := false
			for _, ext := range enhancedOpts.AllowedExtensions {
				if strings.HasSuffix(strings.ToLower(item.Name), strings.ToLower(ext)) {
					allowed = true
					break
				}
			}
			if !allowed {
				fmt.Printf("跳过文件 %s (扩展名不在允许列表中)\n", item.Name)
				continue
			}
		}

		enhancedFiltered = append(enhancedFiltered, item)
	}

	if len(enhancedFiltered) == 0 {
		return ErrNoFilesToRestore
	}

	// 交互确认
	if opts.Interactive {
		fmt.Printf("找到 %d 个符合条件的文件:\n", len(enhancedFiltered))
		for i, item := range enhancedFiltered {
			fmt.Printf("%d. %s (%s)\n", i+1, item.Name, formatBytes(item.Size))
		}

		fmt.Print("确认恢复这些文件? [y/N]: ")
		var input string
		fmt.Scanln(&input)
		if strings.ToLower(input) != "y" && strings.ToLower(input) != "yes" {
			fmt.Println("操作已取消")
			return nil
		}
	}

	// 执行恢复
	successCount := 0
	for _, item := range enhancedFiltered {
		// 检查目标路径是否已存在文件
		if enhancedOpts.BackupExisting {
			if _, err := os.Stat(item.OriginalPath); err == nil {
				// 文件存在，创建备份
				backupPath := item.OriginalPath + ".backup." + time.Now().Format("20060102150405")
				if err := os.Rename(item.OriginalPath, backupPath); err != nil {
					fmt.Printf("无法备份现有文件 %s: %v\n", item.OriginalPath, err)
					continue
				}
				fmt.Printf("已备份现有文件到 %s\n", backupPath)
			}
		}

		// 执行恢复
		if err := restoreSingleItem(item, enhancedOpts); err != nil {
			fmt.Printf("恢复文件 %s 失败: %v\n", item.Name, err)
			continue
		}

		fmt.Printf("已恢复文件: %s\n", item.Name)
		successCount++
	}

	fmt.Printf("成功恢复 %d 个文件\n", successCount)
	if successCount < len(enhancedFiltered) {
		return fmt.Errorf("部分文件恢复失败")
	}

	return nil
}

// restoreSingleItem 恢复单个文件
func restoreSingleItem(item RecycleBinItem, opts EnhancedRestoreOptions) error {
	// 检查源文件是否存在
	if _, err := os.Stat(item.CurrentPath); os.IsNotExist(err) {
		return fmt.Errorf("源文件不存在: %s", item.CurrentPath)
	} else if err != nil {
		return fmt.Errorf("检查源文件时出错: %v", err)
	}

	// 检查目标路径是否存在
	if _, err := os.Stat(item.OriginalPath); err == nil {
		// 目标已存在
		if opts.BackupExisting {
			backupPath := item.OriginalPath + ".backup." + time.Now().Format("20060102150405")
			if err := os.Rename(item.OriginalPath, backupPath); err != nil {
				return fmt.Errorf("备份现有文件失败: %v", err)
			}
		} else {
			return fmt.Errorf("目标文件已存在: %s", item.OriginalPath)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("检查目标文件时出错: %v", err)
	}

	// 确保目标目录存在
	targetDir := filepath.Dir(item.OriginalPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("无法创建目标目录 %s: %v", targetDir, err)
	}

	// 执行文件移动（带重试机制）
	maxRetries := 3
	retryDelay := 100 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		err := os.Rename(item.CurrentPath, item.OriginalPath)
		if err == nil {
			break
		}

		if i == maxRetries-1 {
			return fmt.Errorf("移动文件失败: %v", err)
		}

		// 处理文件被占用的情况
		if isFileInUse(err) {
			time.Sleep(retryDelay)
			continue
		}

		return fmt.Errorf("移动文件失败: %v", err)
	}

	return nil
}

// isFileInUse 检查文件是否被占用
func isFileInUse(err error) bool {
	if runtime.GOOS == "windows" {
		return strings.Contains(err.Error(), "The process cannot access the file")
	}
	return strings.Contains(err.Error(), "text file busy")
}

// ResolveLanguage 解析语言设置
func ResolveLanguage(lang string) string {
	if lang != "" {
		return lang
	}
	return "en-US"
}

// wildcardToRegex 将通配符转换为正则表达式
func wildcardToRegex(pattern string) string {
	pattern = strings.ToLower(pattern)
	pattern = strings.ReplaceAll(pattern, ".", "\\.")
	pattern = strings.ReplaceAll(pattern, "*", ".*")
	pattern = strings.ReplaceAll(pattern, "?", ".")
	return "(?i)^" + pattern + "$"
}
