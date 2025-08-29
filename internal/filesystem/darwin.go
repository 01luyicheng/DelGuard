package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DarwinTrashManager macOS Trash管理器
type DarwinTrashManager struct {
	trashPath string
}

// NewDarwinTrashManager 创建macOS Trash管理器
func NewDarwinTrashManager() *DarwinTrashManager {
	homeDir, _ := os.UserHomeDir()
	trashPath := filepath.Join(homeDir, ".Trash")

	return &DarwinTrashManager{
		trashPath: trashPath,
	}
}

// MoveToTrash 将文件移动到macOS Trash
func (d *DarwinTrashManager) MoveToTrash(filePath string) error {
	// 转换为绝对路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("路径转换失败: %v", err)
	}

	// 检查文件是否存在
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", absPath)
	}

	// 确保Trash目录存在
	if err := os.MkdirAll(d.trashPath, 0755); err != nil {
		return fmt.Errorf("创建Trash目录失败: %v", err)
	}

	// 生成唯一的文件名
	fileName := filepath.Base(absPath)
	targetPath := filepath.Join(d.trashPath, fileName)

	// 如果目标文件已存在，添加时间戳
	counter := 1
	originalFileName := fileName
	for {
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			break
		}

		ext := filepath.Ext(originalFileName)
		nameWithoutExt := originalFileName[:len(originalFileName)-len(ext)]
		fileName = fmt.Sprintf("%s_%d%s", nameWithoutExt, counter, ext)
		targetPath = filepath.Join(d.trashPath, fileName)
		counter++
	}

	// 移动文件到Trash
	err = os.Rename(absPath, targetPath)
	if err != nil {
		return fmt.Errorf("移动到Trash失败: %v", err)
	}

	return nil
}

// GetTrashPath 获取macOS Trash路径
func (d *DarwinTrashManager) GetTrashPath() (string, error) {
	return d.trashPath, nil
}

// ListTrashContents 列出回收站内容
func (d *DarwinTrashManager) ListTrashContents() ([]TrashItem, error) {
	// 检查Trash目录是否存在
	if _, err := os.Stat(d.trashPath); os.IsNotExist(err) {
		return []TrashItem{}, nil // 返回空列表
	}

	entries, err := os.ReadDir(d.trashPath)
	if err != nil {
		return nil, fmt.Errorf("读取Trash失败: %v", err)
	}

	var trashItems []TrashItem
	for _, entry := range entries {
		fullPath := filepath.Join(d.trashPath, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue // 跳过无法获取信息的文件
		}

		trashItem := TrashItem{
			Name:         entry.Name(),
			OriginalPath: "", // macOS Trash不保存原始路径信息
			Path:         fullPath,
			Size:         info.Size(),
			DeletedTime:  info.ModTime(), // 使用修改时间作为删除时间
			IsDirectory:  entry.IsDir(),
		}

		trashItems = append(trashItems, trashItem)
	}

	return trashItems, nil
}

// RestoreFromTrash 从回收站恢复文件
func (d *DarwinTrashManager) RestoreFromTrash(fileName string, originalPath string) error {
	if originalPath == "" {
		return fmt.Errorf("macOS需要指定恢复路径")
	}

	trashFilePath := filepath.Join(d.trashPath, fileName)

	// 检查文件是否存在于回收站
	if _, err := os.Stat(trashFilePath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在于回收站: %s", fileName)
	}

	// 确保目标目录存在
	targetDir := filepath.Dir(originalPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	// 检查目标文件是否已存在
	if _, err := os.Stat(originalPath); err == nil {
		return fmt.Errorf("目标文件已存在: %s", originalPath)
	}

	// 移动文件从Trash到目标位置
	err := os.Rename(trashFilePath, originalPath)
	if err != nil {
		return fmt.Errorf("恢复文件失败: %v", err)
	}

	return nil
}

// GetStats 获取回收站统计信息
func (d *DarwinTrashManager) GetStats() (*TrashStats, error) {
	files, err := d.ListTrashContents()
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

// Clear 清空回收站
func (d *DarwinTrashManager) Clear() error {
	// 检查回收站目录是否存在
	if _, err := os.Stat(d.trashPath); os.IsNotExist(err) {
		return nil // 回收站已经是空的
	}

	entries, err := os.ReadDir(d.trashPath)
	if err != nil {
		return fmt.Errorf("读取回收站失败: %v", err)
	}

	// 删除所有文件和目录
	for _, entry := range entries {
		fullPath := filepath.Join(d.trashPath, entry.Name())
		err := os.RemoveAll(fullPath)
		if err != nil {
			return fmt.Errorf("删除文件失败 %s: %v", fullPath, err)
		}
	}

	return nil
}

// IsEmpty 检查回收站是否为空
func (d *DarwinTrashManager) IsEmpty() bool {
	entries, err := os.ReadDir(d.trashPath)
	if err != nil {
		return true
	}

	return len(entries) == 0
}

// GetTrashStats 获取回收站统计信息（原有方法）
func (d *DarwinTrashManager) GetTrashStats() (*TrashStats, error) {
	return d.GetStats()
}

// EmptyTrash 清空回收站（原有方法）
func (d *DarwinTrashManager) EmptyTrash() error {
	return d.Clear()
}

// CleanOldFiles 清理过期文件
func (d *DarwinTrashManager) CleanOldFiles(maxDays int) error {
	files, err := d.ListTrashContents()
	if err != nil {
		return err
	}

	cutoffTime := time.Now().AddDate(0, 0, -maxDays)

	for _, file := range files {
		if file.DeletedTime.Before(cutoffTime) {
			fullPath := filepath.Join(d.trashPath, file.Name)
			if err := os.RemoveAll(fullPath); err != nil {
				return fmt.Errorf("清理过期文件失败 %s: %v", fullPath, err)
			}
		}
	}

	return nil
}

// ValidateTrash 验证回收站完整性
func (d *DarwinTrashManager) ValidateTrash() error {
	// 检查回收站目录是否存在，不存在则创建
	if _, err := os.Stat(d.trashPath); os.IsNotExist(err) {
		if err := os.MkdirAll(d.trashPath, 0755); err != nil {
			return fmt.Errorf("创建回收站目录失败: %v", err)
		}
	}

	// 检查目录权限
	testFile := filepath.Join(d.trashPath, ".delguard_test")
	file, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("回收站目录无写权限: %s", d.trashPath)
	}
	file.Close()
	os.Remove(testFile)

	return nil
}

// ListTrashFiles 列出回收站中的文件（兼容原有接口）
func (d *DarwinTrashManager) ListTrashFiles() ([]TrashFile, error) {
	items, err := d.ListTrashContents()
	if err != nil {
		return nil, err
	}

	files := make([]TrashFile, len(items))
	for i, item := range items {
		files[i] = TrashFile{
			Name:         item.Name,
			OriginalPath: item.OriginalPath,
			TrashPath:    item.Path,
			Size:         item.Size,
			DeletedTime:  item.DeletedTime,
			IsDirectory:  item.IsDirectory,
		}
	}

	return files, nil
}

// RestoreFile 从回收站恢复文件（兼容原有接口）
func (d *DarwinTrashManager) RestoreFile(trashFile TrashFile, targetPath string) error {
	return d.RestoreFromTrash(trashFile.Name, targetPath)
}
