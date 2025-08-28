package restore

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/01luyicheng/DelGuard/internal/config"
)

// Service 恢复服务
type Service struct {
	config         *config.Config
	enhancedService *EnhancedRestoreService
	historyManager  *RestoreHistoryManager
}

// NewService 创建新的恢复服务
func NewService(cfg *config.Config) *Service {
	// 默认恢复选项
	options := RestoreOptions{
		VerifyIntegrity:   true,
		CreateBackup:      true,
		OverwriteExisting: false,
		PreserveMetadata:  true,
		MaxConcurrency:    4,
		ChunkSize:         1024 * 1024, // 1MB
		EnableResume:      true,
	}

	enhancedService := NewEnhancedRestoreService(options)
	
	// 创建历史管理器
	historyManager, _ := NewRestoreHistoryManager("./restore_history")

	return &Service{
		config:          cfg,
		enhancedService: enhancedService,
		historyManager:  historyManager,
	}
}

// Execute 执行恢复命令
func (s *Service) Execute(ctx context.Context, args []string) error {
	pattern := ""
	listOnly := false
	interactive := false
	maxFiles := 0

	// 解析参数
	for i, arg := range args {
		switch arg {
		case "--list", "-l":
			listOnly = true
		case "--interactive", "-i":
			interactive = true
		case "--max":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &maxFiles)
			}
		default:
			if pattern == "" && !strings.HasPrefix(arg, "-") {
				pattern = arg
			}
		}
	}

	return s.restore(ctx, pattern, listOnly, interactive, maxFiles)
}

// restore 执行恢复操作
func (s *Service) restore(ctx context.Context, pattern string, listOnly, interactive bool, maxFiles int) error {
	// 获取回收站路径
	trashPath, err := s.getTrashPath()
	if err != nil {
		return fmt.Errorf("无法获取回收站路径: %v", err)
	}

	// 列出可恢复的文件
	files, err := s.listRecoverableFiles(trashPath, pattern)
	if err != nil {
		return fmt.Errorf("列出可恢复文件失败: %v", err)
	}

	if len(files) == 0 {
		fmt.Println("回收站中没有可恢复的文件")
		return nil
	}

	// 限制文件数量
	if maxFiles > 0 && len(files) > maxFiles {
		files = files[:maxFiles]
	}

	// 仅列出模式
	if listOnly {
		s.displayRecoverableFiles(files)
		return nil
	}

	// 交互模式
	if interactive {
		return s.interactiveRestore(files)
	}

	// 批量恢复
	return s.batchRestore(files)
}

// RecoverableFile 可恢复文件信息
type RecoverableFile struct {
	OriginalPath string
	CurrentPath  string
	Size         int64
	DeletedTime  time.Time
	IsDirectory  bool
}

// getTrashPath 获取回收站路径
func (s *Service) getTrashPath() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return s.getWindowsTrashPath()
	case "darwin":
		return s.getMacOSTrashPath()
	case "linux":
		return s.getLinuxTrashPath()
	default:
		return "", fmt.Errorf("不支持的操作系统")
	}
}

// getWindowsTrashPath 获取Windows回收站路径
func (s *Service) getWindowsTrashPath() (string, error) {
	appdata := os.Getenv("APPDATA")
	if appdata == "" {
		return "", fmt.Errorf("无法获取APPDATA环境变量")
	}
	return filepath.Join(appdata, "Microsoft", "Windows", "Recent"), nil
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

// listRecoverableFiles 列出可恢复的文件
func (s *Service) listRecoverableFiles(trashPath, pattern string) ([]*RecoverableFile, error) {
	var files []*RecoverableFile

	// 检查回收站是否存在
	if _, err := os.Stat(trashPath); os.IsNotExist(err) {
		return files, nil
	}

	// 遍历回收站
	err := filepath.Walk(trashPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 跳过错误文件
		}

		if info.IsDir() && path != trashPath {
			return filepath.SkipDir
		}

		if !info.IsDir() {
			// 检查文件名是否匹配模式
			if pattern != "" && !strings.Contains(strings.ToLower(info.Name()), strings.ToLower(pattern)) {
				return nil
			}

			file := &RecoverableFile{
				OriginalPath: s.getOriginalPath(path),
				CurrentPath:  path,
				Size:         info.Size(),
				DeletedTime:  info.ModTime(),
				IsDirectory:  false,
			}
			files = append(files, file)
		}

		return nil
	})

	return files, err
}

// getOriginalPath 获取文件原始路径
func (s *Service) getOriginalPath(currentPath string) string {
	// 这里简化处理，实际应该从回收站元数据获取
	trashPath, _ := s.getTrashPath()
	return strings.TrimPrefix(currentPath, trashPath)
}

// displayRecoverableFiles 显示可恢复的文件列表
func (s *Service) displayRecoverableFiles(files []*RecoverableFile) {
	if len(files) == 0 {
		fmt.Println("没有可恢复的文件")
		return
	}

	fmt.Printf("可恢复的文件列表 (%d个):\n", len(files))
	fmt.Println("┌─────────────────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ 序号 │ 文件路径                    │ 大小    │ 删除时间           │")
	fmt.Println("├─────────────────────────────────────────────────────────────────────────────┤")

	for i, file := range files {
		fmt.Printf("│ %4d │ %-25s │ %7s │ %s │\n",
			i+1,
			s.trimPath(file.OriginalPath, 25),
			s.formatSize(file.Size),
			file.DeletedTime.Format("2006-01-02"))
	}

	fmt.Println("└─────────────────────────────────────────────────────────────────────────────┘")
}

// trimPath 截断路径显示
func (s *Service) trimPath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	return "..." + path[len(path)-maxLen+3:]
}

// interactiveRestore 交互式恢复
func (s *Service) interactiveRestore(files []*RecoverableFile) error {
	s.displayRecoverableFiles(files)

	fmt.Print("\n输入要恢复的文件序号（多个用逗号分隔，如：1,3,5）或 'all' 恢复全部: ")
	var input string
	fmt.Scanln(&input)

	input = strings.TrimSpace(strings.ToLower(input))
	if input == "all" {
		return s.batchRestore(files)
	}

	// 解析序号
	indices := strings.Split(input, ",")
	var selectedFiles []*RecoverableFile

	for _, idxStr := range indices {
		var idx int
		if _, err := fmt.Sscanf(strings.TrimSpace(idxStr), "%d", &idx); err != nil {
			continue
		}
		if idx > 0 && idx <= len(files) {
			selectedFiles = append(selectedFiles, files[idx-1])
		}
	}

	if len(selectedFiles) == 0 {
		fmt.Println("未选择任何文件")
		return nil
	}

	return s.batchRestore(selectedFiles)
}

// batchRestore 批量恢复文件
func (s *Service) batchRestore(files []*RecoverableFile) error {
	if len(files) == 0 {
		return nil
	}

	fmt.Printf("正在恢复 %d 个文件...\n", len(files))

	var success, failed int

	for _, file := range files {
		if err := s.restoreSingleFile(file); err != nil {
			fmt.Printf("❌ 恢复失败: %s - %v\n", file.OriginalPath, err)
			failed++
		} else {
			fmt.Printf("✅ 恢复成功: %s\n", file.OriginalPath)
			success++
		}
	}

	fmt.Printf("\n恢复完成: 成功 %d 个，失败 %d 个\n", success, failed)

	if failed > 0 {
		return fmt.Errorf("部分文件恢复失败，失败数量: %d", failed)
	}

	return nil
}

// restoreSingleFile 恢复单个文件
func (s *Service) restoreSingleFile(file *RecoverableFile) error {
	// 确保目标目录存在
	destDir := filepath.Dir(file.OriginalPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	// 检查目标文件是否已存在
	if _, err := os.Stat(file.OriginalPath); err == nil {
		// 文件已存在，创建备份
		backupPath := file.OriginalPath + ".restored_backup." + time.Now().Format("20060102150405")
		if err := os.Rename(file.OriginalPath, backupPath); err != nil {
			return fmt.Errorf("备份现有文件失败: %v", err)
		}
	}

	// 执行恢复
	return os.Rename(file.CurrentPath, file.OriginalPath)
}

// formatSize 格式化文件大小显示
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

// EnhancedRestore 增强恢复功能
func (s *Service) EnhancedRestore(ctx context.Context, files []*RecoverableFile, options RestoreOptions) ([]RestoreResult, error) {
	// 更新恢复选项
	s.enhancedService.options = options
	
	// 开始恢复会话
	sessionID := s.historyManager.StartSession(options)
	defer s.historyManager.EndSession(sessionID)
	
	// 执行恢复
	results, err := s.enhancedService.RestoreFiles(ctx, files)
	
	// 记录恢复历史
	for _, result := range results {
		record := RestoreRecord{
			SourcePath:    "", // 需要从文件信息获取
			DestPath:      result.FilePath,
			FileSize:      result.BytesRestored,
			Checksum:      result.Checksum,
			Success:       result.Success,
			Duration:      result.Duration.Milliseconds(),
			RestoreMethod: "enhanced",
		}
		
		if result.Error != nil {
			record.Error = result.Error.Error()
		}
		
		s.historyManager.AddRecord(sessionID, record)
	}
	
	return results, err
}

// GetRestoreProgress 获取恢复进度
func (s *Service) GetRestoreProgress() RestoreProgress {
	return s.enhancedService.GetProgress()
}

// CancelRestore 取消恢复操作
func (s *Service) CancelRestore() {
	s.enhancedService.Cancel()
}

// GetRestoreHistory 获取恢复历史
func (s *Service) GetRestoreHistory(limit int) []*RestoreSession {
	return s.historyManager.GetRecentSessions(limit)
}

// RollbackRestore 回滚恢复操作
func (s *Service) RollbackRestore(sessionID string) error {
	return s.historyManager.RollbackSession(sessionID)
}

// GetRestoreStatistics 获取恢复统计
func (s *Service) GetRestoreStatistics() map[string]interface{} {
	return s.historyManager.GetStatistics()
}

// CleanupRestoreHistory 清理恢复历史
func (s *Service) CleanupRestoreHistory(maxAge time.Duration) error {
	return s.historyManager.CleanupOldSessions(maxAge)
}
