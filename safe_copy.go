package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
		return fmt.Errorf("无法访问源文件 %s: %v", src, err)
	}

	// 处理目录复制
	if srcInfo.IsDir() {
		if !opts.Recursive {
			return fmt.Errorf("源路径 %s 是目录，使用 -r 参数进行递归复制", src)
		}
		return copyDirectory(src, dst, opts)
	}

	// 检查目标文件是否存在
	if _, err := os.Stat(dst); err == nil {
		// 目标文件存在，进行安全检查
		return handleExistingFile(src, dst, opts)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("检查目标文件时出错 %s: %v", dst, err)
	}

	// 如果目标文件存在
	if err == nil {
		// 检查目标是否为目录
		if info, err := os.Stat(dst); err != nil {
			return fmt.Errorf("检查目标文件时出错 %s: %v", dst, err)
		} else if info.IsDir() {
			// 如果目标是目录，则将源文件复制到该目录下
			dst = filepath.Join(dst, filepath.Base(src))
			// 再次检查这个组合路径
			_, err = os.Stat(dst)
			if err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("检查目标文件时出错 %s: %v", dst, err)
			}
			// 如果仍然存在，则继续处理
			if err == nil {
				// 目标文件存在，继续下面的逻辑
				// 如果目标文件存在且不是目录，进行安全检查
				return handleExistingFile(src, dst, opts)
			} else {
				// 目标文件不存在，直接复制
				return copyFileSafe(src, dst, opts)
			}
		}

		// 如果目标文件存在且不是目录，进行安全检查
		return handleExistingFile(src, dst, opts)
	}
	
	// 目标文件不存在，直接复制
	return copyFileSafe(src, dst, opts)
}

// handleExistingFile 处理目标文件已存在的情况
func handleExistingFile(src, dst string, opts SafeCopyOptions) error {
	if opts.Force {
		// 强制模式下直接覆盖，但仍将原文件移入回收站
		if err := backupExistingFile(dst); err != nil {
			fmt.Printf("警告：无法备份现有文件 %s: %v\n", dst, err)
		}
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
			fmt.Printf("源文件和目标文件内容相同，跳过复制：%s\n", src)
		}
		return nil
	}

	// 文件内容不同，提示用户
	fmt.Printf("目标文件已存在且内容不同:\n")
	fmt.Printf("  源文件: %s (SHA256: %s)\n", src, srcHash[:16])
	fmt.Printf("  目标文件: %s (SHA256: %s)\n", dst, dstHash[:16])

	if opts.Interactive {
		fmt.Print("目标文件已存在且内容不同，是否覆盖？[y/N] ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input != "y" && input != "yes" {
			fmt.Printf("安全复制已取消：%s\n", src+" -> "+dst)
			return nil
		}
	} else {
		fmt.Println("使用 -i 参数以交互模式运行可选择是否覆盖")
		return nil
	}

	// 用户确认覆盖，先备份现有文件
	if err := backupExistingFile(dst); err != nil {
		return fmt.Errorf("备份现有文件失败 %s: %v", dst, err)
	}

	// 执行复制
	return copyFileInternal(src, dst)
}

// backupExistingFile 将现有文件备份到回收站
func backupExistingFile(filePath string) error {
	fmt.Printf("正在将现有文件 %s 移动到回收站...\n", filePath)

	// 使用现有的 moveToTrashPlatform 函数
	err := moveToTrashPlatform(filePath)
	if err != nil {
		return fmt.Errorf("移动文件到回收站失败: %v", err)
	}

	fmt.Printf("已将现有文件 %s 移动到回收站\n", filePath)
	return nil
}

// copyFileInternal 执行实际的文件复制操作
func copyFileInternal(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("无法打开源文件 %s: %v", src, err)
	}
	defer srcFile.Close()

	// 确保目标目录存在
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("无法创建目标目录 %s: %v", dstDir, err)
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("无法创建目标文件 %s: %v", dst, err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("安全复制失败: %s -> %s: %v", src, dst, err)
	}

	// 同步确保数据写入磁盘
	err = dstFile.Sync()
	if err != nil {
		return fmt.Errorf("同步文件时出错 %s: %v", dst, err)
	}

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