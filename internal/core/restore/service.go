package restore

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/01luyicheng/DelGuard/internal/config"
)

// Service 恢复服务
type Service struct {
	config          *config.Config
	metadataManager *MetadataManager
	strategy        RestoreStrategy
}

// RestoreItem 恢复项目
type RestoreItem struct {
	ID           string    `json:"id"`
	OriginalPath string    `json:"originalPath"`
	TrashPath    string    `json:"trashPath"`
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	DeletedTime  time.Time `json:"deletedTime"`
	DeletedBy    string    `json:"deletedBy"`
	Type         string    `json:"type"`
	Checksum     string    `json:"checksum"`
}

// RestoreResult 恢复结果
type RestoreResult struct {
	Success      bool   `json:"success"`
	OriginalPath string `json:"originalPath"`
	RestoredPath string `json:"restoredPath"`
	Error        string `json:"error,omitempty"`
}

// RestoreOptions 恢复选项
type RestoreOptions struct {
	Pattern           string
	Interactive       bool
	ListOnly          bool
	MaxResults        int
	VerifyIntegrity   bool
	CreateBackup      bool
	OverwriteExisting bool
	TargetDirectory   string
}

// NewService 创建恢复服务
func NewService(cfg *config.Config) *Service {
	// 获取回收站路径
	trashPath, _ := getTrashPath(cfg)

	// 创建元数据管理器
	metadataManager := NewMetadataManager(trashPath)
	metadataManager.Load()

	// 创建恢复策略
	strategy := NewSmartRestoreStrategy(metadataManager)

	return &Service{
		config:          cfg,
		metadataManager: metadataManager,
		strategy:        strategy,
	}
}

// getTrashPath 获取回收站路径（辅助函数）
func getTrashPath(cfg *config.Config) (string, error) {
	if cfg.Delete.TrashPath != "" {
		return cfg.Delete.TrashPath, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".delguard", "trash"), nil
}

// Execute 执行恢复命令
func (s *Service) Execute(ctx context.Context, args []string) error {
	// 解析恢复选项
	options, err := s.parseRestoreArgs(args)
	if err != nil {
		return err
	}

	// 获取可恢复的项目
	items, err := s.ListRestorableItems(ctx, options)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		fmt.Println("没有找到可恢复的文件")
		return nil
	}

	// 如果只是列出文件
	if options.ListOnly {
		s.displayRestorableItems(items)
		return nil
	}

	// 执行恢复
	if options.Interactive {
		return s.interactiveRestore(ctx, items, options)
	} else {
		return s.batchRestore(ctx, items, options)
	}
}

// parseRestoreArgs 解析恢复参数
func (s *Service) parseRestoreArgs(args []string) (*RestoreOptions, error) {
	options := &RestoreOptions{
		MaxResults:        s.config.Restore.MaxConcurrency * 10,
		VerifyIntegrity:   s.config.Restore.VerifyIntegrity,
		CreateBackup:      s.config.Restore.CreateBackup,
		OverwriteExisting: s.config.Restore.OverwriteExisting,
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-l", "--list":
			options.ListOnly = true
		case "-i", "--interactive":
			options.Interactive = true
		case "--max":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--max 需要指定数量")
			}
			max, err := strconv.Atoi(args[i+1])
			if err != nil {
				return nil, fmt.Errorf("无效的最大结果数: %s", args[i+1])
			}
			options.MaxResults = max
			i++
		case "--target":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--target 需要指定目标目录")
			}
			options.TargetDirectory = args[i+1]
			i++
		case "--verify":
			options.VerifyIntegrity = true
		case "--no-verify":
			options.VerifyIntegrity = false
		case "--backup":
			options.CreateBackup = true
		case "--no-backup":
			options.CreateBackup = false
		case "--overwrite":
			options.OverwriteExisting = true
		case "--no-overwrite":
			options.OverwriteExisting = false
		default:
			if !strings.HasPrefix(arg, "-") {
				options.Pattern = arg
			} else {
				return nil, fmt.Errorf("未知参数: %s", arg)
			}
		}
	}

	return options, nil
}

// ListRestorableItems 列出可恢复的项目
func (s *Service) ListRestorableItems(ctx context.Context, options *RestoreOptions) ([]*RestoreItem, error) {
	trashPath, err := s.getTrashPath()
	if err != nil {
		return nil, err
	}

	var items []*RestoreItem

	err = filepath.Walk(trashPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续遍历
		}

		// 检查上下文是否被取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 跳过目录本身
		if path == trashPath {
			return nil
		}

		// 创建恢复项目
		item := &RestoreItem{
			ID:          s.generateItemID(path),
			TrashPath:   path,
			Name:        info.Name(),
			Size:        info.Size(),
			DeletedTime: info.ModTime(), // 简化处理，使用修改时间
			Type:        s.getFileType(path, info),
		}

		// 尝试恢复原始路径
		item.OriginalPath = s.guessOriginalPath(path, trashPath)

		// 检查是否匹配模式
		if s.matchesPattern(item, options) {
			items = append(items, item)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 限制结果数量
	if len(items) > options.MaxResults {
		items = items[:options.MaxResults]
	}

	return items, nil
}

// matchesPattern 检查项目是否匹配模式
func (s *Service) matchesPattern(item *RestoreItem, options *RestoreOptions) bool {
	if options.Pattern == "" {
		return true
	}

	// 检查文件名匹配
	matched, _ := filepath.Match(options.Pattern, item.Name)
	if matched {
		return true
	}

	// 检查路径匹配
	if strings.Contains(strings.ToLower(item.OriginalPath), strings.ToLower(options.Pattern)) {
		return true
	}

	return false
}

// displayRestorableItems 显示可恢复的项目
func (s *Service) displayRestorableItems(items []*RestoreItem) {
	fmt.Printf("找到 %d 个可恢复的文件:\n\n", len(items))

	for i, item := range items {
		fmt.Printf("%d. %s\n", i+1, item.Name)
		fmt.Printf("   原始路径: %s\n", item.OriginalPath)
		fmt.Printf("   回收站路径: %s\n", item.TrashPath)
		fmt.Printf("   大小: %s | 删除时间: %s\n",
			s.formatSize(item.Size),
			item.DeletedTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("   类型: %s | ID: %s\n", item.Type, item.ID)
		fmt.Println()
	}
}

// interactiveRestore 交互式恢复
func (s *Service) interactiveRestore(ctx context.Context, items []*RestoreItem, options *RestoreOptions) error {
	fmt.Println("交互式恢复模式")
	fmt.Println("输入文件编号进行恢复，输入 'q' 退出，输入 'a' 恢复所有文件")
	fmt.Println()

	s.displayRestorableItems(items)

	for {
		fmt.Print("请选择要恢复的文件 (编号/a/q): ")
		var input string
		fmt.Scanln(&input)

		switch strings.ToLower(input) {
		case "q", "quit", "exit":
			fmt.Println("退出恢复操作")
			return nil
		case "a", "all":
			fmt.Println("恢复所有文件...")
			return s.batchRestore(ctx, items, options)
		default:
			// 解析编号
			index, err := strconv.Atoi(input)
			if err != nil || index < 1 || index > len(items) {
				fmt.Printf("无效的编号: %s\n", input)
				continue
			}

			item := items[index-1]
			result := s.restoreItem(ctx, item, options)
			if result.Success {
				fmt.Printf("✅ 成功恢复: %s -> %s\n", item.Name, result.RestoredPath)
			} else {
				fmt.Printf("❌ 恢复失败: %s (%s)\n", item.Name, result.Error)
			}
		}
	}
}

// batchRestore 批量恢复
func (s *Service) batchRestore(ctx context.Context, items []*RestoreItem, options *RestoreOptions) error {
	fmt.Printf("开始批量恢复 %d 个文件...\n", len(items))

	successCount := 0
	failCount := 0

	for i, item := range items {
		// 检查上下文是否被取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		fmt.Printf("恢复 %d/%d: %s\n", i+1, len(items), item.Name)

		result := s.restoreItem(ctx, item, options)
		if result.Success {
			successCount++
			fmt.Printf("  ✅ 成功: %s\n", result.RestoredPath)
		} else {
			failCount++
			fmt.Printf("  ❌ 失败: %s\n", result.Error)
		}
	}

	fmt.Printf("\n恢复完成: 成功 %d 个，失败 %d 个\n", successCount, failCount)
	return nil
}

// restoreItem 恢复单个项目
func (s *Service) restoreItem(ctx context.Context, item *RestoreItem, options *RestoreOptions) *RestoreResult {
	result := &RestoreResult{
		OriginalPath: item.OriginalPath,
	}

	// 确定目标路径
	targetPath := item.OriginalPath
	if options.TargetDirectory != "" {
		targetPath = filepath.Join(options.TargetDirectory, item.Name)
	}

	// 检查目标文件是否已存在
	if _, err := os.Stat(targetPath); err == nil {
		if !options.OverwriteExisting {
			result.Error = "目标文件已存在"
			return result
		}

		// 创建备份
		if options.CreateBackup {
			backupPath := targetPath + ".backup." + time.Now().Format("20060102150405")
			if err := os.Rename(targetPath, backupPath); err != nil {
				result.Error = fmt.Sprintf("创建备份失败: %v", err)
				return result
			}
		}
	}

	// 确保目标目录存在
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		result.Error = fmt.Sprintf("创建目标目录失败: %v", err)
		return result
	}

	// 执行恢复（移动文件）
	if err := os.Rename(item.TrashPath, targetPath); err != nil {
		result.Error = fmt.Sprintf("恢复文件失败: %v", err)
		return result
	}

	// 验证完整性
	if options.VerifyIntegrity {
		if err := s.verifyFileIntegrity(targetPath, item); err != nil {
			result.Error = fmt.Sprintf("文件完整性验证失败: %v", err)
			return result
		}
	}

	result.Success = true
	result.RestoredPath = targetPath
	return result
}

// verifyFileIntegrity 验证文件完整性
func (s *Service) verifyFileIntegrity(path string, item *RestoreItem) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	// 检查文件大小
	if info.Size() != item.Size {
		return fmt.Errorf("文件大小不匹配: 期望 %d，实际 %d", item.Size, info.Size())
	}

	// 这里可以添加更多的完整性检查，如校验和验证
	return nil
}

// generateItemID 生成项目ID
func (s *Service) generateItemID(path string) string {
	return fmt.Sprintf("%x", path)
}

// guessOriginalPath 猜测原始路径
func (s *Service) guessOriginalPath(trashPath, trashRoot string) string {
	// 简化实现：假设文件名没有改变
	relPath, _ := filepath.Rel(trashRoot, trashPath)

	// 尝试从当前工作目录恢复
	wd, _ := os.Getwd()
	return filepath.Join(wd, filepath.Base(relPath))
}

// getTrashPath 获取回收站路径
func (s *Service) getTrashPath() (string, error) {
	// 如果配置了自定义回收站路径
	if s.config.Delete.TrashPath != "" {
		return s.config.Delete.TrashPath, nil
	}

	// 根据操作系统获取默认回收站路径
	switch runtime.GOOS {
	case "windows":
		return s.getWindowsTrashPath()
	case "darwin":
		return s.getMacOSTrashPath()
	case "linux":
		return s.getLinuxTrashPath()
	default:
		return "", fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// getWindowsTrashPath 获取Windows回收站路径
func (s *Service) getWindowsTrashPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".delguard", "trash"), nil
}

// getMacOSTrashPath 获取macOS废纸篓路径
func (s *Service) getMacOSTrashPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".Trash"), nil
}

// getLinuxTrashPath 获取Linux回收站路径
func (s *Service) getLinuxTrashPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "Trash", "files"), nil
}

// getFileType 获取文件类型
func (s *Service) getFileType(path string, info os.FileInfo) string {
	if info.IsDir() {
		return "directory"
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".txt", ".md", ".rst":
		return "text"
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg":
		return "image"
	case ".mp4", ".avi", ".mkv", ".mov", ".wmv":
		return "video"
	case ".mp3", ".wav", ".flac", ".aac", ".ogg":
		return "audio"
	default:
		return "file"
	}
}

// formatSize 格式化文件大小
func (s *Service) formatSize(bytes int64) string {
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
