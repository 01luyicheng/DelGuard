package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// MacOSTrashManager macOS废纸篓管理器
type MacOSTrashManager struct {
	trashPath string
}

// NewMacOSTrashManager 创建macOS废纸篓管理器
func NewMacOSTrashManager() *MacOSTrashManager {
	homeDir, _ := os.UserHomeDir()
	trashPath := filepath.Join(homeDir, ".Trash")
	return &MacOSTrashManager{
		trashPath: trashPath,
	}
}

// MoveToTrash 将文件移动到macOS废纸篓
func (m *MacOSTrashManager) MoveToTrash(filePath string) error {
	// 转换为绝对路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("路径转换失败: %v", err)
	}

	// 检查文件是否存在
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", absPath)
	}

	// 确保废纸篓目录存在
	if err := os.MkdirAll(m.trashPath, 0755); err != nil {
		return fmt.Errorf("创建废纸篓目录失败: %v", err)
	}

	// 生成目标文件名（避免重名）
	fileName := filepath.Base(absPath)
	targetPath := filepath.Join(m.trashPath, fileName)

	// 如果目标文件已存在，添加时间戳
	if _, err := os.Stat(targetPath); err == nil {
		timestamp := time.Now().Format("20060102_150405")
		ext := filepath.Ext(fileName)
		nameWithoutExt := fileName[:len(fileName)-len(ext)]
		fileName = fmt.Sprintf("%s_%s%s", nameWithoutExt, timestamp, ext)
		targetPath = filepath.Join(m.trashPath, fileName)
	}

	// 移动文件到废纸篓
	err = os.Rename(absPath, targetPath)
	if err != nil {
		return fmt.Errorf("移动到废纸篓失败: %v", err)
	}

	return nil
}

// GetTrashPath 获取macOS废纸篓路径
func (m *MacOSTrashManager) GetTrashPath() (string, error) {
	if m.trashPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("获取用户目录失败: %v", err)
		}
		m.trashPath = filepath.Join(homeDir, ".Trash")
	}
	return m.trashPath, nil
}

// ListTrashFiles 列出macOS废纸篓中的文件
func (m *MacOSTrashManager) ListTrashFiles() ([]TrashFile, error) {
	trashPath, err := m.GetTrashPath()
	if err != nil {
		return nil, err
	}

	// 检查废纸篓目录是否存在
	if _, err := os.Stat(trashPath); os.IsNotExist(err) {
		return []TrashFile{}, nil // 返回空列表
	}

	entries, err := os.ReadDir(trashPath)
	if err != nil {
		return nil, fmt.Errorf("读取废纸篓失败: %v", err)
	}

	var trashFiles []TrashFile
	for _, entry := range entries {
		// 跳过隐藏文件
		if entry.Name()[0] == '.' {
			continue
		}

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

// RestoreFile 从macOS废纸篓恢复文件
func (m *MacOSTrashManager) RestoreFile(trashFile TrashFile, targetPath string) error {
	// 确保目标目录存在
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	// 检查目标文件是否已存在
	if _, err := os.Stat(targetPath); err == nil {
		return fmt.Errorf("目标文件已存在: %s", targetPath)
	}

	// 移动文件从废纸篓到目标位置
	err := os.Rename(trashFile.TrashPath, targetPath)
	if err != nil {
		return fmt.Errorf("恢复文件失败: %v", err)
	}

	return nil
}

// EmptyTrash 清空macOS废纸篓
func (m *MacOSTrashManager) EmptyTrash() error {
	trashPath, err := m.GetTrashPath()
	if err != nil {
		return err
	}

	// 检查废纸篓目录是否存在
	if _, err := os.Stat(trashPath); os.IsNotExist(err) {
		return nil // 废纸篓已经是空的
	}

	entries, err := os.ReadDir(trashPath)
	if err != nil {
		return fmt.Errorf("读取废纸篓失败: %v", err)
	}

	// 删除所有文件和目录
	for _, entry := range entries {
		// 跳过隐藏文件
		if entry.Name()[0] == '.' {
			continue
		}

		fullPath := filepath.Join(trashPath, entry.Name())
		err := os.RemoveAll(fullPath)
		if err != nil {
			return fmt.Errorf("删除文件失败 %s: %v", fullPath, err)
		}
	}

	return nil
}
