package main

import (
	"crypto/sha256"
	"delguard/utils"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SafeCopyOptions 安全复制选项
type SafeCopyOptions struct {
	Interactive bool
	Force       bool
	Verbose     bool
	Recursive   bool // 新增：递归复制目录
	Preserve    bool // 新增：保留文件属性
}

// SafeCopy 安全复制文件
// 如果目标文件已存在，会计算两个文件的哈希值进行比较，并在文件不同时提示用户确认
func SafeCopy(src, dst string, opts SafeCopyOptions) error {

	// 检查源文件是否存在
	srcInfo, err := os.Stat(src)
	if err != nil {
		log.Printf("[ERROR] 无法访问源文件 %s: %v", src, err)
		return fmt.Errorf("无法访问源文件 %s: %v", src, err)
	}

	// 处理目录复制
	if srcInfo.IsDir() {
		if !opts.Recursive {
			log.Printf("[WARN] 源路径 %s 是目录，未指定递归参数 -r", src)
			return fmt.Errorf("源路径 %s 是目录，使用 -r 参数进行递归复制", src)
		}
		log.Printf("[INFO] 递归复制目录: %s -> %s", src, dst)
		return copyDirectory(src, dst, opts)
	}

	// 正确处理目标为目录的情况：若 dst 是目录，拼接文件名
	dstInfo, dstErr := os.Stat(dst)
	if dstErr == nil && dstInfo.IsDir() {
		dst = filepath.Join(dst, filepath.Base(src))
		// 重新获取组合后的目标信息
		dstInfo, dstErr = os.Stat(dst)
	}

	if dstErr == nil {
		log.Printf("[INFO] 目标文件已存在: %s，执行安全覆盖检查", dst)
		return handleExistingFile(src, dst, opts)
	}
	if dstErr != nil && !os.IsNotExist(dstErr) {
		log.Printf("[ERROR] 检查目标文件时出错 %s: %v", dst, dstErr)
		return fmt.Errorf("检查目标文件时出错 %s: %v", dst, dstErr)
	}

	log.Printf("[INFO] 复制文件: %s -> %s", src, dst)
	return copyFileSafe(src, dst, opts)
}

// handleExistingFile 处理目标文件已存在的情况
func handleExistingFile(src, dst string, opts SafeCopyOptions) error {
	if opts.Force {
		// 强制模式下直接覆盖，但仍将原文件移入回收站
		if err := backupExistingFile(dst); err != nil {
			log.Printf("[WARN] 无法备份现有文件 %s: %v", dst, err)
		}
		log.Printf("[INFO] 强制覆盖文件: %s -> %s", src, dst)
		return copyFileInternal(src, dst)
	}

	// 计算两个文件的哈希值
	srcHash, err := calculateFileHash(src)
	if err != nil {
		return fmt.Errorf("计算源文件哈希时出错 %s: %v", src, err)
	}

	dstHash, err := calculateFileHash(dst)
	if err != nil {
		return fmt.Errorf("计算目标文件哈希时出错 %s: %v", dst, err)
	}

	// 比较哈希值
	if srcHash == dstHash {
		if opts.Verbose {
			log.Printf("[INFO] 源文件和目标文件内容相同，跳过复制：%s", src)
		}
		return nil
	}

	// 文件内容不同，提示用户
	log.Printf("[INFO] 目标文件已存在且内容不同: 源文件: %s (SHA256: %s), 目标文件: %s (SHA256: %s)", src, srcHash[:16], dst, dstHash[:16])

	if opts.Interactive {
		log.Printf("[PROMPT] 目标文件已存在且内容不同，是否覆盖？[y/N]")
		var input string
		if isStdinInteractive() {
			if s, ok := readLineWithTimeout(30 * time.Second); ok {
				input = strings.TrimSpace(strings.ToLower(s))
			} else {
				input = ""
			}
		} else {
			input = ""
		}
		if input != "y" && input != "yes" {
			log.Printf("[INFO] 安全复制已取消：%s -> %s", src, dst)
			return nil
		}
	} else {
		log.Printf("[INFO] 使用 -i 参数以交互模式运行可选择是否覆盖")
		return nil
	}

	// 用户确认覆盖，先备份现有文件
	if err := backupExistingFile(dst); err != nil {
		log.Printf("[ERROR] 备份现有文件失败 %s: %v", dst, err)
		return fmt.Errorf("备份现有文件失败 %s: %v", dst, err)
	}

	log.Printf("[INFO] 覆盖并复制文件: %s -> %s", src, dst)
	return copyFileInternal(src, dst)
}

// backupExistingFile 将现有文件备份到回收站
func backupExistingFile(filePath string) error {
	log.Printf("[INFO] 正在将现有文件 %s 移动到回收站...", filePath)

	// 使用现有的 moveToTrashPlatform 函数
	err := moveToTrashPlatform(filePath)
	if err != nil {
		log.Printf("[ERROR] 移动文件到回收站失败: %v", err)
		return fmt.Errorf("移动文件到回收站失败: %v", err)
	}

	log.Printf("[INFO] 已将现有文件 %s 移动到回收站", filePath)
	return nil
}

// copyFileInternal 执行实际的文件复制操作
func copyFileInternal(src, dst string) error {
	// 确保目标目录存在
	destDir := filepath.Dir(dst)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		log.Printf("[ERROR] 无法创建目标目录 %s: %v", destDir, err)
		return fmt.Errorf("无法创建目标目录 %s: %v", destDir, err)
	}

	// 使用统一的文件复制函数
	if err := utils.CopyFile(src, dst); err != nil {
		log.Printf("[ERROR] 安全复制失败: %s -> %s: %v", src, dst, err)
		return fmt.Errorf("安全复制失败: %s -> %s: %v", src, dst, err)
	}

	log.Printf("[INFO] 文件复制完成: %s -> %s", src, dst)
	return nil
}

// calculateFileHash 计算文件的SHA256哈希值
func calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// copyFileSafe 执行安全的文件复制
func copyFileSafe(src, dst string, opts SafeCopyOptions) error {
	// 确保目标目录存在
	destDir := filepath.Dir(dst)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("无法创建目标目录 %s: %v", destDir, err)
	}

	// 检查目标文件是否存在
	if _, err := os.Stat(dst); err == nil {
		// 目标文件存在，进行安全检查
		return handleExistingFile(src, dst, opts)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("检查目标文件时出错 %s: %v", dst, err)
	}

	// 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("无法打开源文件 %s: %v", src, err)
	}
	defer srcFile.Close()

	// 创建目标文件
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("无法创建目标文件 %s: %v", dst, err)
	}
	defer destFile.Close()

	// 执行复制
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("复制失败: %s -> %s: %v", src, dst, err)
	}

	// 同步确保数据写入磁盘
	err = destFile.Sync()
	if err != nil {
		return fmt.Errorf("同步文件时出错 %s: %v", dst, err)
	}

	// 保留文件属性（如果指定）
	if opts.Preserve {
		if err := preserveFileAttributes(src, dst); err != nil && opts.Verbose {
			fmt.Printf("警告：无法保留文件属性 %s: %v\n", src, err)
		}
	}

	if opts.Verbose {
		fmt.Printf("安全复制完成：%s -> %s\n", src, dst)
	}

	return nil
}

// copyDirectory 递归复制目录
func copyDirectory(src, dst string, opts SafeCopyOptions) error {
	// 获取源目录信息
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("无法访问源目录 %s: %v", src, err)
	}

	// 检查目标是否存在
	if _, err := os.Stat(dst); err == nil {
		// 目标已存在，检查是否为目录
		if dstInfo, err := os.Stat(dst); err == nil && !dstInfo.IsDir() {
			return fmt.Errorf("目标路径 %s 已存在且不是目录", dst)
		}
	}

	// 创建目标目录
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("无法创建目标目录 %s: %v", dst, err)
	}

	// 遍历源目录
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("无法读取源目录 %s: %v", src, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// 递归复制子目录
			if err := copyDirectory(srcPath, dstPath, opts); err != nil {
				return err
			}
		} else {
			// 复制文件
			if err := copyFileSafe(srcPath, dstPath, opts); err != nil {
				return err
			}
		}
	}

	// 保留目录属性（如果指定）
	if opts.Preserve {
		if err := preserveFileAttributes(src, dst); err != nil && opts.Verbose {
			fmt.Printf("警告：无法保留目录属性 %s: %v\n", src, err)
		}
	}

	return nil
}

// preserveFileAttributes 保留文件/目录的属性
func preserveFileAttributes(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// 设置权限
	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		return err
	}

	return nil
}
