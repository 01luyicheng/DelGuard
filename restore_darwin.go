//go:build darwin
// +build darwin

package main

import (
	"delguard/utils"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// restoreFromTrashMacOSImpl macOS平台恢复实现
func restoreFromTrashMacOSImpl(pattern string, opts RestoreOptions) error {
	// 获取废纸篓中的所有文件
	items, err := listMacOSTrashItems()
	if err != nil {
		return fmt.Errorf(T("无法访问废纸篓: %w"), err)
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
		
		if err := restoreSingleFileMacOS(item); err != nil {
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

// listMacOSTrashItems 获取macOS废纸篓中的所有项目
func listMacOSTrashItems() ([]RecycleBinItem, error) {
	var items []RecycleBinItem

	// macOS废纸篓路径
	trashPath := filepath.Join(os.Getenv("HOME"), ".Trash")

	// 检查废纸篓目录是否存在
	if _, err := os.Stat(trashPath); os.IsNotExist(err) {
		return items, nil
	}

	// 遍历废纸篓目录
	files, err := ioutil.ReadDir(trashPath)
	if err != nil {
		return nil, fmt.Errorf("无法读取废纸篓: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		item := RecycleBinItem{
			OriginalPath: filepath.Join(trashPath, file.Name()),
			CurrentPath:  filepath.Join(trashPath, file.Name()),
			Name:         file.Name(),
			Size:         file.Size(),
			DeletedTime:  file.ModTime(),
			FileType:     getFileTypeByPath(file.Name()),
		}

		// 尝试解析.info文件获取原始路径
		infoFile := filepath.Join(trashPath, "."+file.Name()+".info")
		if info, err := parseMacOSInfoFile(infoFile); err == nil && info != "" {
			item.OriginalPath = info
		}

		items = append(items, item)
	}

	return items, nil
}

// parseMacOSInfoFile 解析macOS废纸篓信息文件
func parseMacOSInfoFile(infoPath string) (string, error) {
	if _, err := os.Stat(infoPath); os.IsNotExist(err) {
		return "", nil
	}

	data, err := ioutil.ReadFile(infoPath)
	if err != nil {
		return "", err
	}

	// macOS .info文件是二进制格式，这里简化处理
	// 实际应该解析plist格式
	content := string(data)
	if idx := strings.Index(content, "file://"); idx != -1 {
		// 提取URL编码的路径
		start := idx + 7
		end := strings.Index(content[start:], "\x00")
		if end == -1 {
			end = len(content) - start
		}
		path := content[start : start+end]
		path = strings.ReplaceAll(path, "%20", " ")
		return path, nil
	}

	return "", nil
}

// restoreSingleFileMacOS 恢复单个macOS文件
func restoreSingleFileMacOS(item RecycleBinItem) error {
	// 确保目标目录存在
	targetDir := filepath.Dir(item.OriginalPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("无法创建目标目录: %v", err)
	}

	// 直接重命名/移动文件
	err := os.Rename(item.CurrentPath, item.OriginalPath)
	if err != nil {
		// 如果重命名失败，尝试复制
		return copyFileMacOS(item.CurrentPath, item.OriginalPath)
	}

	return nil
}

// copyFileMacOS 复制文件作为恢复备选方案
func copyFileMacOS(src, dst string) error {
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
		return T("文本文件")
	case ".doc", ".docx":
		return T("Word文档")
	case ".xls", ".xlsx":
		return T("Excel表格")
	case ".pdf":
		return T("PDF文档")
	case ".jpg", ".jpeg", ".png", ".gif":
		return T("图片文件")
	case ".mp4", ".avi", ".mkv":
		return T("视频文件")
	case ".mp3", ".wav", ".flac":
		return T("音频文件")
	default:
		return T("其他文件")
	}
}
