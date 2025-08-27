//go:build !windows

package delete

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// moveToWindowsRecycleBin 在非Windows系统上不可用
func (s *Service) moveToWindowsRecycleBin(filePath string) error {
	return fmt.Errorf("Windows回收站功能在此平台不可用")
}

// moveToUnixTrash Unix/Linux回收站实现
func (s *Service) moveToUnixTrash(filePath string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	trashDir := filepath.Join(homeDir, ".local/share/Trash/files")
	if err := os.MkdirAll(trashDir, 0755); err != nil {
		return err
	}

	fileName := filepath.Base(filePath)
	trashPath := filepath.Join(trashDir, fileName)

	// 如果目标文件已存在，添加时间戳
	if _, err := os.Stat(trashPath); err == nil {
		ext := filepath.Ext(fileName)
		name := strings.TrimSuffix(fileName, ext)
		trashPath = filepath.Join(trashDir, fmt.Sprintf("%s_%d%s", name, os.Getpid(), ext))
	}

	return os.Rename(filePath, trashPath)
}
