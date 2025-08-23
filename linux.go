//go:build linux
// +build linux

package main

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// moveToTrashLinux 将文件移动到Linux回收站
func moveToTrashLinux(filePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return E(KindNotFound, "moveToTrash", filePath, err, "文件不存在")
	}

	// 获取绝对路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return E(KindIO, "moveToTrash", filePath, err, "无法解析绝对路径")
	}

	// 安全检查：防止删除关键路径
	if IsCriticalPath(absPath) {
		return E(KindProtected, "moveToTrash", absPath, nil, "无法删除关键受保护路径")
	}

	// 检查路径是否可访问
	if _, err := os.Lstat(absPath); err != nil {
		return E(KindIO, "moveToTrash", absPath, err, "无法访问文件")
	}

	// 检查文件权限
	fileInfo, err := os.Lstat(absPath)
	if err != nil {
		return E(KindIO, "moveToTrash", absPath, err, "无法获取文件信息")
	}

	// 检查特殊文件类型
	if isSpecialFile(fileInfo) {
		return E(KindProtected, "moveToTrash", absPath, nil, "不支持删除特殊文件类型")
	}

	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return E(KindIO, "moveToTrash", filePath, err, "无法获取用户主目录")
	}

	// 支持多种回收站路径（遵循freedesktop.org规范）
	// 优先使用XDG标准路径，回退到传统路径
	trashDirs := []string{
		// 标准XDG路径
		filepath.Join(homeDir, ".local", "share", "Trash"),
		// 传统路径（兼容性）
		filepath.Join(homeDir, ".Trash"),
		// 某些发行版使用的路径
		filepath.Join(homeDir, ".trash"),
	}

	var trashDir string
	for _, dir := range trashDirs {
		if stat, err := os.Stat(dir); err == nil && stat.IsDir() {
			trashDir = dir
			break
		}
	}

	// 如果都不存在，创建标准XDG路径
	if trashDir == "" {
		trashDir = filepath.Join(homeDir, ".local", "share", "Trash")
	}

	filesDir := filepath.Join(trashDir, "files")
	infoDir := filepath.Join(trashDir, "info")

	// 确保回收站目录存在
	if err := os.MkdirAll(filesDir, 0755); err != nil {
		return E(KindPermission, "moveToTrash", filePath, err, "无法创建回收站files目录")
	}
	if err := os.MkdirAll(infoDir, 0755); err != nil {
		return E(KindPermission, "moveToTrash", filePath, err, "无法创建回收站info目录")
	}

	// 生成唯一的文件名（处理同名文件）
	baseName := filepath.Base(absPath)
	trashFileName := baseName
	counter := 1

	// 检查文件是否已存在于回收站
	for {
		trashFilePath := filepath.Join(filesDir, trashFileName)
		if _, err := os.Stat(trashFilePath); os.IsNotExist(err) {
			break
		}
		// 如果文件已存在，添加时间戳和计数器
		ext := filepath.Ext(baseName)
		nameWithoutExt := baseName[:len(baseName)-len(ext)]
		timestamp := time.Now().Format("20060102_150405")
		trashFileName = fmt.Sprintf("%s_%s_%d%s", nameWithoutExt, timestamp, counter, ext)
		counter++
	}

	trashFilePath := filepath.Join(filesDir, trashFileName)
	infoFilePath := filepath.Join(infoDir, trashFileName+".trashinfo")

	// 获取文件信息用于权限设置（注意：fileInfo已经在前面定义过）
	fileInfo, err = os.Lstat(absPath)
	if err != nil {
		return E(KindIO, "moveToTrash", absPath, err, "无法获取文件信息")
	}

	// 优先尝试 rename，若跨设备（EXDEV）则回退为复制后删除
	if err := os.Rename(absPath, trashFilePath); err != nil {
		if !isEXDEV(err) {
			return WrapE("moveToTrash", absPath, err)
		}
		// 跨设备回退：复制后删除源
		if err := copyTree(absPath, trashFilePath); err != nil {
			return E(KindIO, "moveToTrash", absPath, err, "跨设备复制失败")
		}
		// 删除源（根据类型）
		if rmErr := removeOriginal(absPath); rmErr != nil {
			_ = os.RemoveAll(trashFilePath) // 回滚已复制的目标
			return E(KindIO, "moveToTrash", absPath, rmErr, "删除源文件失败")
		}
	}

	// 确保目标文件权限与源文件一致
	if err := os.Chmod(trashFilePath, fileInfo.Mode()); err != nil {
		// 非致命错误，继续执行
	}

	// 创建.trashinfo文件（Path 字段按段 URL 编码，保留路径分隔符）
	infoContent := fmt.Sprintf("[Trash Info]\nPath=%s\nDeletionDate=%s\n",
		encodeTrashInfoPath(absPath), time.Now().Format("2006-01-02T15:04:05"))

	if err := os.WriteFile(infoFilePath, []byte(infoContent), 0644); err != nil {
		// 如果创建info文件失败，尝试恢复原文件
		_ = os.Rename(trashFilePath, absPath)
		return E(KindIO, "moveToTrash", absPath, err, "创建回收站信息文件失败")
	}

	return nil
}

func isEXDEV(err error) bool {
	var errno syscall.Errno
	if errors.As(err, &errno) {
		return errno == syscall.EXDEV
	}
	// 某些情况下 err 不是 *PathError 包装，保守判断
	return strings.Contains(strings.ToLower(err.Error()), "cross-device")
}

func removeOriginal(p string) error {
	fi, err := os.Lstat(p)
	if err != nil {
		return err
	}
	if fi.IsDir() {
		return os.RemoveAll(p)
	}
	return os.Remove(p)
}

func copyTree(src, dst string) error {
	fi, err := os.Lstat(src)
	if err != nil {
		return err
	}

	switch {
	case fi.Mode()&os.ModeSymlink != 0:
		return copySymlink(src, dst)
	case fi.IsDir():
		if err := os.MkdirAll(dst, fi.Mode().Perm()); err != nil {
			return err
		}
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		for _, e := range entries {
			sChild := filepath.Join(src, e.Name())
			dChild := filepath.Join(dst, e.Name())
			if err := copyTree(sChild, dChild); err != nil {
				return err
			}
		}
		return nil
	default:
		return copyFile(src, dst, fi)
	}
}

func copyFile(src, dst string, fi os.FileInfo) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	// 确保父目录存在
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fi.Mode())
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

func copySymlink(src, dst string) error {
	target, err := os.Readlink(src)
	if err != nil {
		return err
	}
	// 确保父目录存在
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	// 创建符号链接
	if err := os.Symlink(target, dst); err != nil {
		return err
	}

	// 尝试保持符号链接的权限（如果平台支持）
	// 注意：在Linux上，符号链接的权限通常由系统管理
	// 我们无法直接设置符号链接的权限，但可以设置目标文件的权限

	return nil
}

// encodeTrashInfoPath 对 .trashinfo 的 Path 字段进行按段 URL 编码，保留 /
func encodeTrashInfoPath(p string) string {
	if p == "" {
		return ""
	}
	// 处理绝对路径的前导/
	isAbsolute := strings.HasPrefix(p, "/")
	parts := strings.Split(p, "/")
	for i := range parts {
		// 跳过空段（绝对路径的前导/）
		if i == 0 && parts[i] == "" {
			continue
		}
		if parts[i] != "" {
			parts[i] = url.PathEscape(parts[i])
		}
	}
	result := strings.Join(parts, "/")
	if isAbsolute && !strings.HasPrefix(result, "/") {
		result = "/" + result
	}
	return result
}

// DecodeTrashInfoPath 解码.trashinfo中的Path字段，供其他平台使用
func DecodeTrashInfoPath(p string) string {
	return decodeTrashInfoPath(p)
}

// decodeTrashInfoPath 解码.trashinfo中的Path字段
func decodeTrashInfoPath(p string) string {
	if p == "" {
		return ""
	}
	parts := strings.Split(p, "/")
	for i := range parts {
		if parts[i] != "" {
			if decoded, err := url.PathUnescape(parts[i]); err == nil {
				parts[i] = decoded
			}
		}
	}
	return strings.Join(parts, "/")
}

// 为Linux平台提供其他平台函数的存根
func moveToTrashWindows(filePath string) error {
	return ErrUnsupportedPlatform
}

func moveToTrashMacOS(filePath string) error {
	return ErrUnsupportedPlatform
}

// checkFileOwnershipUnix 检查Unix文件所有权
func checkFileOwnershipUnix(filePath string) error {
	// 获取当前用户信息
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("无法获取当前用户信息: %v", err)
	}
	
	// 获取文件信息
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("无法获取文件信息: %v", err)
	}
	
	// 检查是否为符号链接，避免安全风险
	if info.Mode()&os.ModeSymlink != 0 {
		// 检查符号链接的目标
		target, err := os.Readlink(filePath)
		if err != nil {
			return fmt.Errorf("无法读取符号链接目标: %v", err)
		}
		
		// 检查目标文件的所有权
		targetInfo, err := os.Stat(target)
		if err != nil {
			return fmt.Errorf("无法获取符号链接目标信息: %v", err)
		}
		
		// 检查目标文件权限
		if targetInfo.Mode().Perm()&0222 == 0 {
			return fmt.Errorf("符号链接目标文件为只读")
		}
		
		return nil
	}
	
	// 检查文件权限
	if info.Mode().Perm()&0222 == 0 {
		return fmt.Errorf("文件为只读")
	}
	
	// 检查是否为系统文件
	if info.Mode()&os.ModeDevice != 0 || info.Mode()&os.ModeCharDevice != 0 {
		return fmt.Errorf("无法操作系统设备文件")
	}
	
	return nil
}
