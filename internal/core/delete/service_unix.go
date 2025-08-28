//go:build !windows

package delete

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// moveToWindowsRecycleBin 在非Windows系统上不可用
func (s *Service) moveToWindowsRecycleBin(filePath string) error {
	return fmt.Errorf("Windows回收站功能在此平台不可用")
}

// moveToUnixTrash Unix系统回收站实现
func (s *Service) moveToUnixTrash(filePath string) error {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户主目录失败: %v", err)
	}

	// 创建回收站目录结构
	trashDir := filepath.Join(homeDir, ".local", "share", "Trash")
	filesDir := filepath.Join(trashDir, "files")
	infoDir := filepath.Join(trashDir, "info")

	// 确保回收站目录存在
	if err := os.MkdirAll(filesDir, 0755); err != nil {
		return fmt.Errorf("创建回收站目录失败: %v", err)
	}
	if err := os.MkdirAll(infoDir, 0755); err != nil {
		return fmt.Errorf("创建回收站信息目录失败: %v", err)
	}

	// 获取文件名和扩展名
	fileName := filepath.Base(filePath)
	
	// 生成唯一的回收站文件名
	trashFileName := fileName
	trashFilePath := filepath.Join(filesDir, trashFileName)
	counter := 1
	
	// 如果文件名冲突，添加数字后缀
	for {
		if _, err := os.Stat(trashFilePath); os.IsNotExist(err) {
			break
		}
		ext := filepath.Ext(fileName)
		nameWithoutExt := fileName[:len(fileName)-len(ext)]
		trashFileName = fmt.Sprintf("%s_%d%s", nameWithoutExt, counter, ext)
		trashFilePath = filepath.Join(filesDir, trashFileName)
		counter++
	}

	// 移动文件到回收站
	if err := os.Rename(filePath, trashFilePath); err != nil {
		return fmt.Errorf("移动文件到回收站失败: %v", err)
	}

	// 创建.trashinfo文件
	infoFileName := trashFileName + ".trashinfo"
	infoFilePath := filepath.Join(infoDir, infoFileName)
	
	// 获取绝对路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}

	// 创建info文件内容
	infoContent := fmt.Sprintf("[Trash Info]\nPath=%s\nDeletionDate=%s\n",
		absPath, time.Now().Format("2006-01-02T15:04:05"))

	if err := os.WriteFile(infoFilePath, []byte(infoContent), 0644); err != nil {
		// 如果创建info文件失败，尝试恢复原文件
		os.Rename(trashFilePath, filePath)
		return fmt.Errorf("创建回收站信息文件失败: %v", err)
	}

	return nil
}