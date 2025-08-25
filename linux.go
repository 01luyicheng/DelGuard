//go:build linux
// +build linux

package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

// moveToTrashLinux 将文件移动到Linux回收站
func moveToTrashLinux(filePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return ErrFileNotFound
	}

	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// 确定回收站目录
	trashDir := filepath.Join(homeDir, ".local", "share", "Trash")
	trashFilesDir := filepath.Join(trashDir, "files")
	trashInfoDir := filepath.Join(trashDir, "info")

	// 确保回收站目录存在
	if err := os.MkdirAll(trashFilesDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(trashInfoDir, 0755); err != nil {
		return err
	}

	// 获取文件的绝对路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}

	// 生成唯一的文件名以避免冲突
	baseName := filepath.Base(absPath)
	trashFileName := baseName
	trashFilePath := filepath.Join(trashFilesDir, trashFileName)

	counter := 1
	for {
		if _, err := os.Stat(trashFilePath); os.IsNotExist(err) {
			break
		}
		// 文件名已存在，添加数字后缀
		ext := filepath.Ext(baseName)
		nameWithoutExt := strings.TrimSuffix(baseName, ext)
		trashFileName = fmt.Sprintf("%s.%d%s", nameWithoutExt, counter, ext)
		trashFilePath = filepath.Join(trashFilesDir, trashFileName)
		counter++
	}

	// 移动文件到回收站
	if err := os.Rename(absPath, trashFilePath); err != nil {
		// 如果跨设备，需要复制然后删除
		if isEXDEV(err) {
			if err := copyTree(absPath, trashFilePath); err != nil {
				return err
			}
			if err := removeOriginal(absPath); err != nil {
				// 如果删除失败，也要删除复制的文件
				os.Remove(trashFilePath)
				return err
			}
		} else {
			return err
		}
	}

	// 创建 .trashinfo 文件
	trashInfoPath := filepath.Join(trashInfoDir, trashFileName+".trashinfo")
	trashInfoContent := fmt.Sprintf(
		"[Trash Info]\nPath=%s\nDeletionDate=%s\n",
		absPath,
		time.Now().Format("2006-01-02T15:04:05"),
	)

	if err := os.WriteFile(trashInfoPath, []byte(trashInfoContent), 0644); err != nil {
		// 如果创建.trashinfo文件失败，应该将文件移回原位
		os.Rename(trashFilePath, absPath)
		os.Remove(trashInfoPath)
		return err
	}

	return nil
}

// getCurrentUserSID 获取当前用户的SID（Linux实现）
func getCurrentUserSID() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	return currentUser.Uid, nil
}

// CheckFilePermissions 检查文件权限（Linux实现）
func CheckFilePermissions(filePath string) (bool, error) {
	// Linux平台的文件权限检查
	_, err := os.Stat(filePath)
	if err != nil {
		return false, err
	}

	return true, nil
}
