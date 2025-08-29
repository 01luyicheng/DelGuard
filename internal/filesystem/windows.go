package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// WindowsTrashManager Windows回收站管理器
type WindowsTrashManager struct {
	trashPath string
}

// NewWindowsTrashManager 创建Windows回收站管理器
func NewWindowsTrashManager() *WindowsTrashManager {
	return &WindowsTrashManager{}
}

// MoveToTrash 将文件移动到Windows回收站
func (w *WindowsTrashManager) MoveToTrash(filePath string) error {
	// 转换为绝对路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("路径转换失败: %v", err)
	}

	// 检查文件是否存在
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", absPath)
	}

	// 使用Windows Shell API移动到回收站
	return w.moveToRecycleBin(absPath)
}

// moveToRecycleBin 使用Windows API移动文件到回收站
func (w *WindowsTrashManager) moveToRecycleBin(filePath string) error {
	// 简化实现：直接使用Go的os包移动到用户回收站目录
	// 这是一个跨平台兼容的实现方案

	// 获取用户回收站路径
	userProfile := os.Getenv("USERPROFILE")
	if userProfile == "" {
		return fmt.Errorf("无法获取用户配置目录")
	}

	// 创建DelGuard专用回收站目录
	delguardTrash := filepath.Join(userProfile, ".delguard", "trash")
	if err := os.MkdirAll(delguardTrash, 0755); err != nil {
		return fmt.Errorf("创建DelGuard回收站目录失败: %v", err)
	}

	// 生成唯一文件名
	fileName := filepath.Base(filePath)
	targetPath := filepath.Join(delguardTrash, fileName)

	// 如果文件已存在，添加时间戳
	if _, err := os.Stat(targetPath); err == nil {
		timestamp := fmt.Sprintf("_%d", time.Now().Unix())
		ext := filepath.Ext(fileName)
		nameWithoutExt := fileName[:len(fileName)-len(ext)]
		fileName = nameWithoutExt + timestamp + ext
		targetPath = filepath.Join(delguardTrash, fileName)
	}

	// 移动文件
	return os.Rename(filePath, targetPath)
}

// GetTrashPath 获取Windows回收站路径
func (w *WindowsTrashManager) GetTrashPath() (string, error) {
	// 使用DelGuard专用回收站目录
	userProfile := os.Getenv("USERPROFILE")
	if userProfile == "" {
		userProfile = "C:\\Users\\Default"
	}

	// 返回DelGuard专用回收站路径
	return filepath.Join(userProfile, ".delguard", "trash"), nil
}

// ListTrashFiles 列出Windows回收站中的文件
func (w *WindowsTrashManager) ListTrashFiles() ([]TrashFile, error) {
	trashPath, err := w.GetTrashPath()
	if err != nil {
		return nil, err
	}

	// 检查回收站目录是否存在
	if _, err := os.Stat(trashPath); os.IsNotExist(err) {
		return []TrashFile{}, nil // 返回空列表
	}

	entries, err := os.ReadDir(trashPath)
	if err != nil {
		return nil, fmt.Errorf("读取回收站失败: %v", err)
	}

	var trashFiles []TrashFile
	for _, entry := range entries {
		fullPath := filepath.Join(trashPath, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue // 跳过无法获取信息的文件
		}

		trashFile := TrashFile{
			Name:        entry.Name(),
			TrashPath:   fullPath,
			Size:        info.Size(),
			DeletedTime: info.ModTime(),
			IsDirectory: entry.IsDir(),
		}

		trashFiles = append(trashFiles, trashFile)
	}

	return trashFiles, nil
}

// RestoreFile 从Windows回收站恢复文件
func (w *WindowsTrashManager) RestoreFile(trashFile TrashFile, targetPath string) error {
	// 确保目标目录存在
	targetDir := filepath.Dir(targetPath)
	if err := CreateDirIfNotExists(targetDir); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	// 移动文件从回收站到目标位置
	err := os.Rename(trashFile.TrashPath, targetPath)
	if err != nil {
		return fmt.Errorf("恢复文件失败: %v", err)
	}

	return nil
}

// EmptyTrash 清空Windows回收站
func (w *WindowsTrashManager) EmptyTrash() error {
	trashPath, err := w.GetTrashPath()
	if err != nil {
		return err
	}

	// 检查回收站目录是否存在
	if _, err := os.Stat(trashPath); os.IsNotExist(err) {
		return nil // 回收站已经是空的
	}

	entries, err := os.ReadDir(trashPath)
	if err != nil {
		return fmt.Errorf("读取回收站失败: %v", err)
	}

	// 删除所有文件和目录
	for _, entry := range entries {
		fullPath := filepath.Join(trashPath, entry.Name())
		err := os.RemoveAll(fullPath)
		if err != nil {
			return fmt.Errorf("删除文件失败 %s: %v", fullPath, err)
		}
	}

	return nil
}

// ListTrashContents 列出回收站内容（接口实现）
func (w *WindowsTrashManager) ListTrashContents() ([]TrashItem, error) {
	files, err := w.ListTrashFiles()
	if err != nil {
		return nil, err
	}

	items := make([]TrashItem, len(files))
	for i, file := range files {
		items[i] = TrashItem{
			Name:         file.Name,
			OriginalPath: file.OriginalPath,
			Path:         file.TrashPath,
			Size:         file.Size,
			DeletedTime:  file.DeletedTime,
			IsDirectory:  file.IsDirectory,
		}
	}

	return items, nil
}

// RestoreFromTrash 从回收站恢复文件（接口实现）
func (w *WindowsTrashManager) RestoreFromTrash(fileName string, originalPath string) error {
	files, err := w.ListTrashFiles()
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.Name == fileName {
			targetPath := originalPath
			if targetPath == "" {
				targetPath = file.OriginalPath
			}
			return w.RestoreFile(file, targetPath)
		}
	}

	return fmt.Errorf("文件未找到: %s", fileName)
}

// GetStats 获取回收站统计信息（接口实现）
func (w *WindowsTrashManager) GetStats() (*TrashStats, error) {
	return w.GetTrashStats()
}

// Clear 清空回收站（接口实现）
func (w *WindowsTrashManager) Clear() error {
	return w.EmptyTrash()
}

// IsEmpty 检查回收站是否为空
func (w *WindowsTrashManager) IsEmpty() bool {
	files, err := w.ListTrashFiles()
	if err != nil {
		return true
	}
	return len(files) == 0
}

// GetTrashStats 获取回收站统计信息
func (w *WindowsTrashManager) GetTrashStats() (*TrashStats, error) {
	files, err := w.ListTrashFiles()
	if err != nil {
		return nil, err
	}

	stats := &TrashStats{
		TotalFiles: int64(len(files)),
		TotalSize:  0,
	}

	if len(files) > 0 {
		stats.OldestFile = files[0].DeletedTime
		for _, file := range files {
			stats.TotalSize += file.Size
			if file.DeletedTime.Before(stats.OldestFile) {
				stats.OldestFile = file.DeletedTime
			}
		}
	}

	return stats, nil
}

// CleanOldFiles 清理过期文件
func (w *WindowsTrashManager) CleanOldFiles(maxDays int) error {
	files, err := w.ListTrashFiles()
	if err != nil {
		return err
	}

	cutoffTime := time.Now().AddDate(0, 0, -maxDays)

	for _, file := range files {
		if file.DeletedTime.Before(cutoffTime) {
			if err := os.RemoveAll(file.TrashPath); err != nil {
				return fmt.Errorf("清理过期文件失败 %s: %v", file.TrashPath, err)
			}
		}
	}

	return nil
}

// ValidateTrash 验证回收站完整性
func (w *WindowsTrashManager) ValidateTrash() error {
	trashPath, err := w.GetTrashPath()
	if err != nil {
		return err
	}

	// 检查回收站目录是否存在，不存在则创建
	if _, err := os.Stat(trashPath); os.IsNotExist(err) {
		if err := os.MkdirAll(trashPath, 0755); err != nil {
			return fmt.Errorf("创建回收站目录失败: %v", err)
		}
	}

	// 检查目录权限
	if !w.hasWritePermission(trashPath) {
		return fmt.Errorf("回收站目录无写权限: %s", trashPath)
	}

	return nil
}

// hasWritePermission 检查写权限
func (w *WindowsTrashManager) hasWritePermission(path string) bool {
	testFile := filepath.Join(path, ".delguard_test")
	file, err := os.Create(testFile)
	if err != nil {
		return false
	}
	file.Close()
	os.Remove(testFile)
	return true
}
