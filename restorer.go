package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// getOS 获取当前操作系统类型
func getOS() string {
	return runtime.GOOS
}

// Restorer 文件恢复器
type Restorer struct {
	config *Config
}

// NewRestorer 创建新的文件恢复器
func NewRestorer(config *Config) *Restorer {
	return &Restorer{config: config}
}

// Restore 执行文件恢复
func (r *Restorer) Restore(ctx context.Context, pattern string, listOnly, interactive bool, maxFiles int) error {
	// 获取回收站路径
	trashPath, err := r.getTrashPath()
	if err != nil {
		return NewDGError(ErrSystem, "无法获取回收站路径", err)
	}

	// 列出可恢复的文件
	files, err := r.listRecoverableFiles(trashPath, pattern)
	if err != nil {
		return NewDGError(ErrIOFailure, "列出可恢复文件失败", err)
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
		r.displayRecoverableFiles(files)
		return nil
	}

	// 交互模式
	if interactive {
		return r.interactiveRestore(files)
	}

	// 批量恢复
	return r.batchRestore(files)
}

// getTrashPath 获取回收站路径
func (r *Restorer) getTrashPath() (string, error) {
	switch getOS() {
	case "windows":
		return r.getWindowsTrashPath()
	case "darwin":
		return r.getMacOSTrashPath()
	case "linux":
		return r.getLinuxTrashPath()
	default:
		return "", fmt.Errorf("不支持的操作系统")
	}
}

// getWindowsTrashPath 获取Windows回收站路径
func (r *Restorer) getWindowsTrashPath() (string, error) {
	appdata := os.Getenv("APPDATA")
	if appdata == "" {
		return "", fmt.Errorf("无法获取APPDATA环境变量")
	}
	return filepath.Join(appdata, "Microsoft", "Windows", "Recent"), nil
}

// getMacOSTrashPath 获取macOS废纸篓路径
func (r *Restorer) getMacOSTrashPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".Trash"), nil
}

// getLinuxTrashPath 获取Linux回收站路径
func (r *Restorer) getLinuxTrashPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "Trash", "files"), nil
}

// RecoverableFile 可恢复文件信息
type RecoverableFile struct {
	OriginalPath string
	CurrentPath  string
	Size         int64
	DeletedTime  time.Time
	IsDirectory  bool
}

// listRecoverableFiles 列出可恢复的文件
func (r *Restorer) listRecoverableFiles(trashPath, pattern string) ([]*RecoverableFile, error) {
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
				OriginalPath: r.getOriginalPath(path),
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
func (r *Restorer) getOriginalPath(currentPath string) string {
	// 这里简化处理，实际应该从回收站元数据获取
	return strings.TrimPrefix(currentPath, r.getTrashDir())
}

// getTrashDir 获取回收站目录前缀
func (r *Restorer) getTrashDir() string {
	trashPath, _ := r.getTrashPath()
	return trashPath
}

// displayRecoverableFiles 显示可恢复的文件列表
func (r *Restorer) displayRecoverableFiles(files []*RecoverableFile) {
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
			trimPath(file.OriginalPath, 25),
			formatSize(file.Size),
			file.DeletedTime.Format("2006-01-02"))
	}

	fmt.Println("└─────────────────────────────────────────────────────────────────────────────┘")
}

// trimPath 截断路径显示
func trimPath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	return "..." + path[len(path)-maxLen+3:]
}

// interactiveRestore 交互式恢复
func (r *Restorer) interactiveRestore(files []*RecoverableFile) error {
	r.displayRecoverableFiles(files)

	fmt.Print("\n输入要恢复的文件序号（多个用逗号分隔，如：1,3,5）或 'all' 恢复全部: ")
	var input string
	fmt.Scanln(&input)

	input = strings.TrimSpace(strings.ToLower(input))
	if input == "all" {
		return r.batchRestore(files)
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

	return r.batchRestore(selectedFiles)
}

// batchRestore 批量恢复文件
func (r *Restorer) batchRestore(files []*RecoverableFile) error {
	if len(files) == 0 {
		return nil
	}

	fmt.Printf("正在恢复 %d 个文件...\n", len(files))
	
	var success, failed int
	
	for _, file := range files {
		if err := r.restoreSingleFile(file); err != nil {
			fmt.Printf("❌ 恢复失败: %s - %v\n", file.OriginalPath, err)
			failed++
		} else {
			fmt.Printf("✅ 恢复成功: %s\n", file.OriginalPath)
			success++
		}
	}

	fmt.Printf("\n恢复完成: 成功 %d 个，失败 %d 个\n", success, failed)
	
	if failed > 0 {
		return NewDGError(ErrIOFailure, fmt.Sprintf("部分文件恢复失败，失败数量: %d", failed), nil)
	}
	
	return nil
}

// restoreSingleFile 恢复单个文件
func (r *Restorer) restoreSingleFile(file *RecoverableFile) error {
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
func formatSize(bytes int64) string {
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