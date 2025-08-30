package filesystem

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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

	// 创建元数据目录
	metadataDir := filepath.Join(d.trashPath, ".delguard_metadata")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return fmt.Errorf("创建元数据目录失败: %v", err)
	}

	// 获取文件信息
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %v", err)
	}

	// 生成唯一的文件名
	fileName := filepath.Base(absPath)
	baseName := fileName[:len(fileName)-len(filepath.Ext(fileName))]
	ext := filepath.Ext(fileName)
	timestamp := time.Now().Format("20060102_150405")
	uniqueName := fmt.Sprintf("%s_%s%s", baseName, timestamp, ext)
	targetPath := filepath.Join(d.trashPath, uniqueName)

	// 如果目标文件已存在，添加更多随机字符
	counter := 1
	for {
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			break
		}
		uniqueName = fmt.Sprintf("%s_%s_%d%s", baseName, timestamp, counter, ext)
		targetPath = filepath.Join(d.trashPath, uniqueName)
		counter++
	}

	// 创建元数据
	metadata := TrashMetadata{
		OriginalPath: absPath,
		DeletedTime:  time.Now(),
		FileName:     fileName,
		Size:         fileInfo.Size(),
		IsDirectory:  fileInfo.IsDir(),
		Permissions:  fileInfo.Mode().String(),
		SystemTrash:  false,
	}
	
	metadataFile := filepath.Join(metadataDir, uniqueName+".json")
	if err := d.writeJSONMetadata(metadataFile, metadata); err != nil {
		return fmt.Errorf("创建元数据文件失败: %v", err)
	}

	// 移动文件到Trash
	err = os.Rename(absPath, targetPath)
	if err != nil {
		// 如果重命名失败，尝试复制后删除
		if copyErr := d.copyAndRemove(absPath, targetPath); copyErr != nil {
			// 清理元数据文件
			os.Remove(metadataFile)
			return fmt.Errorf("移动到Trash失败: %v", copyErr)
		}
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

	metadataDir := filepath.Join(d.trashPath, ".delguard_metadata")
	var trashItems []TrashItem
	
	for _, entry := range entries {
		// 跳过元数据目录和隐藏文件
		if entry.Name() == ".delguard_metadata" || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		
		fullPath := filepath.Join(d.trashPath, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue // 跳过无法获取信息的文件
		}

		// 尝试读取对应的元数据文件获取原始路径
		metadataFile := filepath.Join(metadataDir, entry.Name()+".json")
		originalPath, deletedTime := d.readJSONMetadata(metadataFile)
		
		if deletedTime.IsZero() {
			deletedTime = info.ModTime() // 使用修改时间作为回退
		}

		trashItem := TrashItem{
			Name:         entry.Name(),
			OriginalPath: originalPath,
			Path:         fullPath,
			Size:         info.Size(),
			DeletedTime:  deletedTime,
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

	// 创建元数据目录
	metadataDir := filepath.Join(d.trashPath, ".delguard_metadata")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return fmt.Errorf("创建元数据目录失败: %v", err)
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

// writeJSONMetadata 写入JSON格式的元数据文件
func (d *DarwinTrashManager) writeJSONMetadata(metadataFile string, metadata TrashMetadata) error {
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化元数据失败: %v", err)
	}
	
	return os.WriteFile(metadataFile, data, 0644)
}

// readJSONMetadata 读取JSON格式的元数据文件
func (d *DarwinTrashManager) readJSONMetadata(metadataFile string) (string, time.Time) {
	data, err := os.ReadFile(metadataFile)
	if err != nil {
		return "", time.Time{}
	}
	
	var metadata TrashMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return "", time.Time{}
	}
	
	return metadata.OriginalPath, metadata.DeletedTime
}

// copyAndRemove 复制文件后删除源文件（用于跨设备移动）
func (d *DarwinTrashManager) copyAndRemove(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// 检查源文件类型
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		// 复制目录
		return d.copyDirectory(src, dst)
	}

	// 复制文件
	return d.copyFile(src, dst)
}

// copyFile 复制单个文件
func (d *DarwinTrashManager) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// 确保数据写入磁盘
	dstFile.Sync()

	// 删除源文件
	return os.Remove(src)
}

// copyDirectory 递归复制目录
func (d *DarwinTrashManager) copyDirectory(src, dst string) error {
	// 获取源目录信息
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// 创建目标目录
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	// 读取源目录内容
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// 递归复制每个条目
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// 递归复制子目录
			if err := d.copyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// 复制文件
			if err := d.copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	// 删除源目录
	return os.RemoveAll(src)
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
