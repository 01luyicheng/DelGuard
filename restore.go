package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// RestoreOptions 恢复选项
type RestoreOptions struct {
	// 文件名模式，支持通配符
	Pattern string
	// 最大恢复文件数
	MaxFiles int
	// 是否交互式确认
	Interactive bool
}

// 恢复文件的错误类型
var (
	ErrRestoreNotSupported = errors.New("当前平台不支持恢复功能")
	ErrNoFilesToRestore    = errors.New("没有找到可恢复的文件")
	ErrRestoreFailed       = errors.New("恢复文件失败")
)

// restoreFromTrash 从回收站恢复文件
func restoreFromTrash(pattern string, opts RestoreOptions) error {
	switch runtime.GOOS {
	case "windows":
		return restoreFromTrashWindows(pattern, opts)
	case "darwin":
		return restoreFromTrashMacOS(pattern, opts)
	case "linux":
		return restoreFromTrashLinux(pattern, opts)
	default:
		return ErrRestoreNotSupported
	}
}

// -------------------- Windows --------------------

func restoreFromTrashWindows(pattern string, opts RestoreOptions) error {
	// Windows 回收站恢复需要 Shell/COM，当前版本提示用户使用系统回收站
	fmt.Println("Windows平台暂不支持命令行恢复功能")
	fmt.Println("请使用 Windows 资源管理器打开回收站，手动恢复文件")
	return ErrRestoreNotSupported
}

// -------------------- macOS --------------------

func restoreFromTrashMacOS(pattern string, opts RestoreOptions) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	trashDir := filepath.Join(homeDir, ".Trash")
	if _, err := os.Stat(trashDir); os.IsNotExist(err) {
		return ErrNoFilesToRestore
	}

	matches, err := filepath.Glob(filepath.Join(trashDir, pattern))
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return ErrNoFilesToRestore
	}
	if opts.MaxFiles > 0 && len(matches) > opts.MaxFiles {
		matches = matches[:opts.MaxFiles]
	}

	fmt.Printf("找到 %d 个匹配的文件:\n", len(matches))
	for i, file := range matches {
		fmt.Printf("%d. %s\n", i+1, filepath.Base(file))
	}

	if opts.Interactive {
		fmt.Print("是否恢复这些文件? (y/n): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			return nil
		}
	}

	restoredCount := 0
	for _, file := range matches {
		destPath := filepath.Join(".", filepath.Base(file))
		if _, err := os.Stat(destPath); err == nil {
			ext := filepath.Ext(destPath)
			baseName := strings.TrimSuffix(filepath.Base(destPath), ext)
			timestamp := time.Now().Format("20060102_150405")
			destPath = filepath.Join(".", fmt.Sprintf("%s_restored_%s%s", baseName, timestamp, ext))
		}
		if err := os.Rename(file, destPath); err != nil {
			fmt.Printf("无法恢复文件 %s: %s\n", filepath.Base(file), err.Error())
			continue
		}
		fmt.Printf("已恢复: %s\n", destPath)
		restoredCount++
	}

	if restoredCount == 0 {
		return ErrRestoreFailed
	}
	fmt.Printf("成功恢复 %d 个文件\n", restoredCount)
	return nil
}

// restoreFile 恢复单个文件
func restoreFile(trashPath, originalPath string) error {
	// 验证原始路径
	if err := validateRestorePath(originalPath); err != nil {
		return fmt.Errorf("路径验证失败: %v", err)
	}

	// 检查目标位置是否已存在文件
	if _, err := os.Stat(originalPath); err == nil {
		return fmt.Errorf("目标位置已存在文件: %s", originalPath)
	}

	// 检查回收站文件是否存在
	if _, err := os.Stat(trashPath); os.IsNotExist(err) {
		return fmt.Errorf("回收站中找不到该文件: %s", trashPath)
	}

	// 检查回收站文件权限
	if err := checkFilePermissions(trashPath, nil); err != nil {
		return fmt.Errorf("回收站文件权限检查失败: %v", err)
	}

	// 创建目标目录
	destDir := filepath.Dir(originalPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("无法创建目标目录: %v", err)
	}

	// 检查目标目录权限
	if err := checkDirectoryPermissions(destDir); err != nil {
		return fmt.Errorf("目标目录权限不足: %v", err)
	}

	// 移动文件
	if err := os.Rename(trashPath, originalPath); err != nil {
		return fmt.Errorf("无法移动文件: %v", err)
	}

	return nil
}

// validateRestorePath 验证恢复路径的有效性
func validateRestorePath(path string) error {
	// 检查空路径
	if path == "" {
		return fmt.Errorf("恢复路径不能为空")
	}

	// 检查路径遍历
	if strings.Contains(path, "..") {
		return fmt.Errorf("路径包含非法字符")
	}

	// 检查路径长度
	if len(path) > 260 {
		return fmt.Errorf("路径过长")
	}

	// 检查系统目录
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("无法获取绝对路径")
	}

	// 禁止恢复到系统关键目录
	protectedPaths := []string{
		"C:\\Windows",
		"C:\\Program Files",
		"C:\\Program Files (x86)",
		"/usr",
		"/bin",
		"/sbin",
		"/etc",
		"/var",
	}

	for _, protected := range protectedPaths {
		if strings.HasPrefix(strings.ToLower(absPath), strings.ToLower(protected)) {
			return fmt.Errorf("禁止恢复到系统目录: %s", protected)
		}
	}

	return nil
}

// checkDirectoryPermissions 检查目录权限
func checkDirectoryPermissions(dir string) error {
	// 检查目录是否存在
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// 如果目录不存在，检查父目录权限
		parent := filepath.Dir(dir)
		return checkDirectoryPermissions(parent)
	}

	// 检查目录是否可写
	testFile := filepath.Join(dir, ".delguard_test")
	if f, err := os.Create(testFile); err == nil {
		f.Close()
		os.Remove(testFile)
		return nil
	}

	return fmt.Errorf("目录无写权限: %s", dir)
}

// checkFilePermissions 检查文件权限
func checkFilePermissions(path string, fi os.FileInfo) error {
	if fi == nil {
		var err error
		fi, err = os.Stat(path)
		if err != nil {
			return err
		}
	}
	// 简单检查文件是否可读
	if _, err := os.OpenFile(path, os.O_RDONLY, 0); err != nil {
		return fmt.Errorf("无法读取文件: %v", err)
	}
	return nil
}

// -------------------- Linux --------------------

func restoreFromTrashLinux(pattern string, opts RestoreOptions) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	trashFilesDir := filepath.Join(homeDir, ".local", "share", "Trash", "files")
	trashInfoDir := filepath.Join(homeDir, ".local", "share", "Trash", "info")
	if _, err := os.Stat(trashFilesDir); os.IsNotExist(err) {
		return ErrNoFilesToRestore
	}

	matches, err := filepath.Glob(filepath.Join(trashFilesDir, pattern))
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return ErrNoFilesToRestore
	}
	if opts.MaxFiles > 0 && len(matches) > opts.MaxFiles {
		matches = matches[:opts.MaxFiles]
	}

	fmt.Printf("找到 %d 个匹配的项:\n", len(matches))
	dirCount := 0
	for i, file := range matches {
		fi, _ := os.Lstat(file)
		isDir := fi != nil && fi.IsDir()
		if isDir {
			dirCount++
		}
		if isDir {
			fmt.Printf("%d. %s [目录]\n", i+1, filepath.Base(file))
		} else {
			fmt.Printf("%d. %s\n", i+1, filepath.Base(file))
		}
	}

	if opts.Interactive {
		fmt.Printf("将恢复 %d 个项（其中目录 %d 个）。是否继续? (y/n): ", len(matches), dirCount)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			return nil
		}
	}

	restoredCount := 0
	for _, file := range matches {
		baseName := filepath.Base(file)

		// 读取 .trashinfo 的原始路径
		infoFile := filepath.Join(trashInfoDir, baseName+".trashinfo")
		originalPath := ""
		if infoContent, err := os.ReadFile(infoFile); err == nil {
			lines := strings.Split(string(infoContent), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "Path=") {
					raw := strings.TrimPrefix(line, "Path=")
					originalPath = decodeTrashInfoPath(raw)
					break
				}
			}
		}

		// 目标路径：优先原路径，否则当前目录
		destPath := filepath.Join(".", baseName)
		if originalPath != "" {
			destPath = originalPath
		}

		// 确保目标父目录存在（若无权限创建，回退当前目录）
		destDir := filepath.Dir(destPath)
		if originalPath != "" {
			if _, err := os.Stat(destDir); os.IsNotExist(err) {
				_ = os.MkdirAll(destDir, 0o755)
				if _, err2 := os.Stat(destDir); err2 != nil {
					destPath = filepath.Join(".", baseName)
				}
			}
		}

		// 目标存在则加时间戳
		if _, err := os.Stat(destPath); err == nil {
			ext := filepath.Ext(destPath)
			baseNameWithoutExt := strings.TrimSuffix(filepath.Base(destPath), ext)
			timestamp := time.Now().Format("20060102_150405")
			destPath = filepath.Join(filepath.Dir(destPath),
				fmt.Sprintf("%s_restored_%s%s", baseNameWithoutExt, timestamp, ext))
		}

		// 先尝试重命名；跨设备时在 linux.go 中实现的工具函数辅助复制回退
		if err := os.Rename(file, destPath); err != nil {
			if isEXDEV(err) {
				if err := copyTree(file, destPath); err != nil {
					fmt.Printf("无法恢复 %s（复制失败）: %v\n", baseName, err)
					continue
				}
				if rmErr := removeOriginal(file); rmErr != nil {
					_ = os.RemoveAll(destPath)
					fmt.Printf("无法恢复 %s（清理源失败）: %v\n", baseName, rmErr)
					continue
				}
			} else {
				fmt.Printf("无法恢复 %s: %v\n", baseName, err)
				continue
			}
		}

		// 清理 .trashinfo
		_ = os.Remove(infoFile)

		fmt.Printf("已恢复: %s\n", destPath)
		restoredCount++
	}

	if restoredCount == 0 {
		return ErrRestoreFailed
	}
	fmt.Printf("成功恢复 %d 个文件\n", restoredCount)
	return nil
}
