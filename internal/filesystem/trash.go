package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// TrashManager 回收站管理器接口
type TrashManager interface {
	// MoveToTrash 将文件移动到回收站
	MoveToTrash(filePath string) error
	// GetTrashPath 获取回收站路径
	GetTrashPath() (string, error)
	// ListTrashFiles 列出回收站中的文件
	ListTrashFiles() ([]TrashFile, error)
	// ListTrashContents 列出回收站内容（统一接口）
	ListTrashContents() ([]TrashItem, error)
	// RestoreFile 从回收站恢复文件
	RestoreFile(trashFile TrashFile, targetPath string) error
	// RestoreFromTrash 从回收站恢复文件（统一接口）
	RestoreFromTrash(fileName string, originalPath string) error
	// EmptyTrash 清空回收站
	EmptyTrash() error
	// GetTrashStats 获取回收站统计信息
	GetTrashStats() (*TrashStats, error)
	// GetStats 获取回收站统计信息（统一接口）
	GetStats() (*TrashStats, error)
	// Clear 清空回收站（统一接口）
	Clear() error
	// IsEmpty 检查回收站是否为空
	IsEmpty() bool
	// CleanOldFiles 清理过期文件
	CleanOldFiles(maxDays int) error
	// ValidateTrash 验证回收站完整性
	ValidateTrash() error
}

// TrashFile 回收站文件信息
type TrashFile struct {
	ID           string    // 唯一标识符
	Name         string    // 文件名
	OriginalPath string    // 原始路径
	TrashPath    string    // 回收站中的路径
	Size         int64     // 文件大小
	DeletedTime  time.Time // 删除时间
	IsDirectory  bool      // 是否为目录
	Permissions  string    // 文件权限
}

// TrashItem 通用回收站项目信息（用于接口统一）
type TrashItem struct {
	Name         string    // 文件名
	OriginalPath string    // 原始路径
	Path         string    // 当前路径（回收站中的路径）
	Size         int64     // 文件大小
	DeletedTime  time.Time // 删除时间
	IsDirectory  bool      // 是否为目录
}

// TrashStats 回收站统计信息
type TrashStats struct {
	TotalFiles int64     // 总文件数
	TotalSize  int64     // 总大小
	OldestFile time.Time // 最旧文件时间
}

// GetTrashManager 根据操作系统获取对应的回收站管理器
func GetTrashManager() (TrashManager, error) {
	switch runtime.GOOS {
	case "windows":
		return NewWindowsTrashManager(), nil
	case "darwin":
		return NewDarwinTrashManager(), nil
	case "linux":
		return NewLinuxTrashManager(), nil
	default:
		return nil, fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// FormatFileSize 格式化文件大小显示
func FormatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// IsValidPath 检查路径是否有效
func IsValidPath(path string) bool {
	if path == "" {
		return false
	}

	// 检查路径是否存在
	_, err := os.Stat(path)
	return err == nil
}

// GetAbsolutePath 获取绝对路径
func GetAbsolutePath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("无法获取绝对路径: %v", err)
	}
	return absPath, nil
}

// CreateDirIfNotExists 如果目录不存在则创建
func CreateDirIfNotExists(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}
