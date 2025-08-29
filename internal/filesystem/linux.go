package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LinuxTrashManager Linux Trash管理器
type LinuxTrashManager struct {
	trashPath string
	infoPath  string
}

// NewLinuxTrashManager 创建Linux Trash管理器
func NewLinuxTrashManager() *LinuxTrashManager {
	homeDir, _ := os.UserHomeDir()

	// 遵循XDG Trash规范
	trashPath := filepath.Join(homeDir, ".local", "share", "Trash", "files")
	infoPath := filepath.Join(homeDir, ".local", "share", "Trash", "info")

	return &LinuxTrashManager{
		trashPath: trashPath,
		infoPath:  infoPath,
	}
}

// MoveToTrash 将文件移动到Linux Trash
func (l *LinuxTrashManager) MoveToTrash(filePath string) error {
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
	if err := os.MkdirAll(l.trashPath, 0755); err != nil {
		return fmt.Errorf("创建Trash目录失败: %v", err)
	}
	if err := os.MkdirAll(l.infoPath, 0755); err != nil {
		return fmt.Errorf("创建Trash info目录失败: %v", err)
	}

	// 生成唯一的文件名
	fileName := filepath.Base(absPath)
	targetPath := filepath.Join(l.trashPath, fileName)
	infoFilePath := filepath.Join(l.infoPath, fileName+".trashinfo")

	// 如果目标文件已存在，添加时间戳
	counter := 1
	originalFileName := fileName
	for {
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			if _, err := os.Stat(infoFilePath); os.IsNotExist(err) {
				break
			}
		}

		ext := filepath.Ext(originalFileName)
		nameWithoutExt := originalFileName[:len(originalFileName)-len(ext)]
		fileName = fmt.Sprintf("%s_%d%s", nameWithoutExt, counter, ext)
		targetPath = filepath.Join(l.trashPath, fileName)
		infoFilePath = filepath.Join(l.infoPath, fileName+".trashinfo")
		counter++
	}

	// 移动文件到Trash
	err = os.Rename(absPath, targetPath)
	if err != nil {
		return fmt.Errorf("移动到Trash失败: %v", err)
	}

	// 创建.trashinfo文件
	err = l.createTrashInfo(infoFilePath, absPath)
	if err != nil {
		// 如果创建info文件失败，尝试恢复原文件
		os.Rename(targetPath, absPath)
		return fmt.Errorf("创建Trash信息文件失败: %v", err)
	}

	return nil
}

// ListTrashContents 列出回收站内容（接口实现）
func (l *LinuxTrashManager) ListTrashContents() ([]TrashItem, error) {
	files, err := l.ListTrashFiles()
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
func (l *LinuxTrashManager) RestoreFromTrash(fileName string, originalPath string) error {
	files, err := l.ListTrashFiles()
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.Name == fileName {
			targetPath := originalPath
			if targetPath == "" {
				targetPath = file.OriginalPath
			}
			return l.RestoreFile(file, targetPath)
		}
	}

	return fmt.Errorf("文件未找到: %s", fileName)
}

// GetStats 获取回收站统计信息
func (l *LinuxTrashManager) GetStats() (*TrashStats, error) {
	files, err := l.ListTrashContents()
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

// GetTrashStats 获取回收站统计信息（原有方法）
func (l *LinuxTrashManager) GetTrashStats() (*TrashStats, error) {
	return l.GetStats()
}

// CleanOldFiles 清理过期文件
func (l *LinuxTrashManager) CleanOldFiles(maxDays int) error {
	files, err := l.ListTrashFiles()
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
func (l *LinuxTrashManager) ValidateTrash() error {
	// 检查回收站目录是否存在，不存在则创建
	if _, err := os.Stat(l.trashPath); os.IsNotExist(err) {
		if err := os.MkdirAll(l.trashPath, 0755); err != nil {
			return fmt.Errorf("创建回收站目录失败: %v", err)
		}
	}

	// 创建必要的子目录
	filesDir := filepath.Join(l.trashPath, "files")
	infoDir := filepath.Join(l.trashPath, "info")

	if err := os.MkdirAll(filesDir, 0755); err != nil {
		return fmt.Errorf("创建files目录失败: %v", err)
	}

	if err := os.MkdirAll(infoDir, 0755); err != nil {
		return fmt.Errorf("创建info目录失败: %v", err)
	}

	return nil
}

// Clear 清空回收站
func (l *LinuxTrashManager) Clear() error {
	trashPath, err := l.GetTrashPath()
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

// IsEmpty 检查回收站是否为空
func (l *LinuxTrashManager) IsEmpty() bool {
	trashPath, err := l.GetTrashPath()
	if err != nil {
		return true
	}

	entries, err := os.ReadDir(trashPath)
	if err != nil {
		return true
	}

	return len(entries) == 0
}

// createTrashInfo 创建Trash信息文件
func (l *LinuxTrashManager) createTrashInfo(infoPath, originalPath string) error {
	// 创建符合XDG Trash规范的.trashinfo文件
	content := fmt.Sprintf("[Trash Info]\nPath=%s\nDeletionDate=%s\n",
		originalPath,
		time.Now().Format("2006-01-02T15:04:05"))

	return os.WriteFile(infoPath, []byte(content), 0644)
}

// GetTrashPath 获取Linux Trash路径
func (l *LinuxTrashManager) GetTrashPath() (string, error) {
	return l.trashPath, nil
}

// ListTrashFiles 列出Linux Trash中的文件
func (l *LinuxTrashManager) ListTrashFiles() ([]TrashFile, error) {
	// 检查Trash目录是否存在
	if _, err := os.Stat(l.trashPath); os.IsNotExist(err) {
		return []TrashFile{}, nil // 返回空列表
	}

	entries, err := os.ReadDir(l.trashPath)
	if err != nil {
		return nil, fmt.Errorf("读取Trash失败: %v", err)
	}

	var trashFiles []TrashFile
	for _, entry := range entries {
		fullPath := filepath.Join(l.trashPath, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue // 跳过无法获取信息的文件
		}

		// 尝试读取对应的.trashinfo文件获取原始路径
		infoFilePath := filepath.Join(l.infoPath, entry.Name()+".trashinfo")
		originalPath, deletionTime := l.readTrashInfo(infoFilePath)

		trashFile := TrashFile{
			Name:         entry.Name(),
			OriginalPath: originalPath,
			TrashPath:    fullPath,
			Size:         info.Size(),
			DeletedTime:  deletionTime,
			IsDirectory:  entry.IsDir(),
		}

		trashFiles = append(trashFiles, trashFile)
	}

	return trashFiles, nil
}

// readTrashInfo 读取.trashinfo文件信息
func (l *LinuxTrashManager) readTrashInfo(infoPath string) (string, time.Time) {
	content, err := os.ReadFile(infoPath)
	if err != nil {
		return "", time.Time{}
	}

	lines := strings.Split(string(content), "\n")
	var originalPath string
	var deletionTime time.Time

	// 简单解析.trashinfo文件
	for _, line := range lines {
		if len(line) > 5 && line[:5] == "Path=" {
			originalPath = line[5:]
		} else if len(line) > 13 && line[:13] == "DeletionDate=" {
			if t, err := time.Parse("2006-01-02T15:04:05", line[13:]); err == nil {
				deletionTime = t
			}
		}
	}

	return originalPath, deletionTime
}

// RestoreFile 从Linux Trash恢复文件
func (l *LinuxTrashManager) RestoreFile(trashFile TrashFile, targetPath string) error {
	// 如果没有指定目标路径，使用原始路径
	if targetPath == "" {
		if trashFile.OriginalPath != "" {
			targetPath = trashFile.OriginalPath
		} else {
			return fmt.Errorf("无法确定恢复路径")
		}
	}

	// 确保目标目录存在
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	// 检查目标文件是否已存在
	if _, err := os.Stat(targetPath); err == nil {
		return fmt.Errorf("目标文件已存在: %s", targetPath)
	}

	// 移动文件从Trash到目标位置
	err := os.Rename(trashFile.TrashPath, targetPath)
	if err != nil {
		return fmt.Errorf("恢复文件失败: %v", err)
	}

	// 删除对应的.trashinfo文件
	infoFilePath := filepath.Join(l.infoPath, filepath.Base(trashFile.TrashPath)+".trashinfo")
	os.Remove(infoFilePath) // 忽略删除错误

	return nil
}

// EmptyTrash 清空Linux Trash
func (l *LinuxTrashManager) EmptyTrash() error {
	// 清空files目录
	if err := l.emptyDirectory(l.trashPath); err != nil {
		return fmt.Errorf("清空Trash文件失败: %v", err)
	}

	// 清空info目录
	if err := l.emptyDirectory(l.infoPath); err != nil {
		return fmt.Errorf("清空Trash信息失败: %v", err)
	}

	return nil
}

// emptyDirectory 清空指定目录
func (l *LinuxTrashManager) emptyDirectory(dirPath string) error {
	// 检查目录是否存在
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return nil // 目录不存在，认为已经是空的
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	// 删除所有文件和目录
	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())
		err := os.RemoveAll(fullPath)
		if err != nil {
			return fmt.Errorf("删除文件失败 %s: %v", fullPath, err)
		}
	}

	return nil
}
