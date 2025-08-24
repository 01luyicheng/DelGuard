//go:build darwin
// +build darwin

package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// moveToTrashMacOS 将文件移动到macOS废纸篓
func moveToTrashMacOS(filePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return ErrFileNotFound
	}

	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// 确定废纸篓目录
	trashDir := filepath.Join(homeDir, ".Trash")

	// 确保废纸篓目录存在
	if err := os.MkdirAll(trashDir, 0755); err != nil {
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
	trashFilePath := filepath.Join(trashDir, trashFileName)

	counter := 1
	for {
		if _, err := os.Stat(trashFilePath); os.IsNotExist(err) {
			break
		}
		// 文件名已存在，添加时间戳后缀
		ext := filepath.Ext(baseName)
		nameWithoutExt := strings.TrimSuffix(baseName, ext)
		trashFileName = fmt.Sprintf("%s_%d%s", nameWithoutExt, counter, ext)
		trashFilePath = filepath.Join(trashDir, trashFileName)
		counter++
	}

	// 移动文件到废纸篓
	return os.Rename(absPath, trashFilePath)
}

// getCurrentUserSID 获取当前用户的SID（macOS实现）
func getCurrentUserSID() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	return currentUser.Uid, nil
}

// CheckFilePermissions 检查文件权限（macOS实现）
func CheckFilePermissions(filePath string) (bool, error) {
	// macOS平台的文件权限检查
	_, err := os.Stat(filePath)
	if err != nil {
		return false, err
	}

	return true, nil
}

// checkDiskSpace macOS平台检查磁盘空间 (stub实现)
func checkDiskSpace(path string, requiredBytes int64) error {
	// macOS平台暂不检查磁盘空间，直接返回成功
	// 可以在未来实现使用syscall.Statfs检查磁盘空间
	return nil
}
