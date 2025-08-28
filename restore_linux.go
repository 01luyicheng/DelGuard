//go:build linux

package main

import (
	"bufio"
	"delguard/utils"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// restoreFromTrashLinuxImpl Linux平台恢复实现
func restoreFromTrashLinuxImpl(pattern string, opts RestoreOptions) error {
	// 获取回收站中的所有文件
	items, err := listLinuxTrashItems()
	if err != nil {
		return fmt.Errorf(T("无法访问回收站: %w"), err)
	}

	if len(items) == 0 {
		return ErrNoFilesToRestore
	}

	// 根据模式筛选文件
	var matchedItems []RecycleBinItem
	if pattern == "" {
		matchedItems = items
	} else {
		// 支持通配符匹配
		regex := wildcardToRegex(pattern)
		for _, item := range items {
			matchedName := strings.ToLower(item.Name)
			if regexp.MustCompile(regex).MatchString(matchedName) {
				matchedItems = append(matchedItems, item)
			}
		}
	}

	if len(matchedItems) == 0 {
		return fmt.Errorf(T("没有找到匹配的文件: %s"), pattern)
	}

	// 限制最大文件数
	if opts.MaxFiles > 0 && len(matchedItems) > opts.MaxFiles {
		matchedItems = matchedItems[:opts.MaxFiles]
	}

	// 交互模式确认
	if opts.Interactive {
		fmt.Printf(T("找到 %d 个匹配文件:\n"), len(matchedItems))
		for i, item := range matchedItems {
			fmt.Printf("%d. %s (%s) - 删除时间: %s\n",
				i+1, item.Name, utils.FormatBytes(item.Size),
				item.DeletedTime.Format(TimeFormatStandard))
		}

		fmt.Print(T("确认恢复这些文件吗? (y/N): "))
		var response string
		if isStdinInteractive() {
			if s, ok := readLineWithTimeout(20 * time.Second); ok {
				response = strings.TrimSpace(strings.ToLower(s))
			} else {
				response = ""
			}
		} else {
			response = ""
		}
		if response != "y" && response != "yes" {
			return fmt.Errorf(T("用户取消操作"))
		}
	}

	// 恢复文件
	successCount := 0
	for _, item := range matchedItems {
		// 验证恢复路径安全性
		if err := validateRestorePath(item.OriginalPath); err != nil {
			fmt.Printf(T("恢复路径验证失败 %s: %v\n"), item.Name, err)
			continue
		}

		if err := restoreSingleFileLinux(item); err != nil {
			fmt.Printf(T("恢复文件失败 %s: %v\n"), item.Name, err)
		} else {
			fmt.Printf(T("成功恢复: %s -> %s\n"), item.Name, item.OriginalPath)
			successCount++
		}
	}

	if successCount == 0 {
		return fmt.Errorf(T("所有文件恢复失败"))
	}

	fmt.Printf(T("成功恢复 %d 个文件\n"), successCount)
	return nil
}

// listLinuxTrashItems 获取Linux回收站中的所有项目
func listLinuxTrashItems() ([]RecycleBinItem, error) {
	var items []RecycleBinItem

	// Linux回收站路径
	trashPath := filepath.Join(os.Getenv("HOME"), ".local/share/Trash")
	filesPath := filepath.Join(trashPath, "files")
	infoPath := filepath.Join(trashPath, "info")

	// 检查回收站目录是否存在
	if _, err := os.Stat(filesPath); os.IsNotExist(err) {
		return items, nil
	}

	// 遍历回收站文件
	files, err := ioutil.ReadDir(filesPath)
	if err != nil {
		return nil, fmt.Errorf("无法读取回收站: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(filesPath, file.Name())
		infoFile := filepath.Join(infoPath, file.Name()+".trashinfo")

		// 解析信息文件
		originalPath, deletedTime, err := parseLinuxTrashInfo(infoFile)
		if err != nil {
			// 如果无法解析信息文件，使用默认值
			originalPath = filePath
			deletedTime = file.ModTime()
		}

		item := RecycleBinItem{
			OriginalPath: originalPath,
			CurrentPath:  filePath,
			Name:         filepath.Base(originalPath),
			Size:         file.Size(),
			DeletedTime:  deletedTime,
			FileType:     getFileTypeByPath(originalPath),
		}

		items = append(items, item)
	}

	return items, nil
}

// parseLinuxTrashInfo 解析Linux回收站信息文件
func parseLinuxTrashInfo(infoFile string) (string, time.Time, error) {
	if _, err := os.Stat(infoFile); os.IsNotExist(err) {
		return "", time.Time{}, fmt.Errorf("信息文件不存在")
	}

	file, err := os.Open(infoFile)
	if err != nil {
		return "", time.Time{}, err
	}
	defer file.Close()

	var originalPath string
	var deletedTime time.Time

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "Path=") {
			path := strings.TrimPrefix(line, "Path=")
			// URL解码路径
			path = strings.ReplaceAll(path, "%20", " ")
			path = strings.ReplaceAll(path, "%21", "!")
			path = strings.ReplaceAll(path, "%23", "#")
			path = strings.ReplaceAll(path, "%24", "$")
			path = strings.ReplaceAll(path, "%25", "%")
			path = strings.ReplaceAll(path, "%26", "&")
			originalPath = path
		}

		if strings.HasPrefix(line, "DeletionDate=") {
			dateStr := strings.TrimPrefix(line, "DeletionDate=")
			deletedTime, _ = time.Parse("2006-01-02T15:04:05", dateStr)
		}
	}

	if originalPath == "" {
		return "", time.Time{}, fmt.Errorf("无法解析原始路径")
	}

	return originalPath, deletedTime, nil
}

// restoreSingleFileLinux 恢复单个Linux文件
func restoreSingleFileLinux(item RecycleBinItem) error {
	// 确保目标目录存在
	targetDir := filepath.Dir(item.OriginalPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("无法创建目标目录: %v", err)
	}

	// 直接重命名/移动文件
	err := os.Rename(item.CurrentPath, item.OriginalPath)
	if err != nil {
		// 如果重命名失败，尝试复制
		return copyFileLinux(item.CurrentPath, item.OriginalPath)
	}

	// 删除对应的信息文件
	infoFile := filepath.Join(os.Getenv("HOME"), ".local/share/Trash/info", filepath.Base(item.CurrentPath)+".trashinfo")
	os.Remove(infoFile)

	return nil
}

// copyFileLinux 复制文件作为恢复备选方案
func copyFileLinux(src, dst string) error {
	if err := utils.CopyFile(src, dst); err != nil {
		return err
	}

	// 删除源文件
	return os.Remove(src)
}

// getFileTypeByPath 根据文件路径获取文件类型
func getFileTypeByPath(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".txt":
		return "文本文件"
	case ".doc", ".docx":
		return "Word文档"
	case ".xls", ".xlsx":
		return "Excel表格"
	case ".pdf":
		return "PDF文档"
	case ".jpg", ".jpeg", ".png", ".gif":
		return "图片文件"
	case ".mp4", ".avi", ".mkv":
		return "视频文件"
	case ".mp3", ".wav", ".flac":
		return "音频文件"
	default:
		return "其他文件"
	}
}
