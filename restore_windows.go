//go:build windows
// +build windows

package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// restoreFromTrashWindowsImpl Windows平台恢复实现
func restoreFromTrashWindowsImpl(pattern string, opts RestoreOptions) error {
	// 获取回收站中的所有文件
	items, err := listRecycleBinItemsWindows()
	if err != nil {
		return fmt.Errorf("无法访问回收站: %w", err)
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
			if regexp.MustCompile(regex).MatchString(strings.ToLower(item.Name)) {
				matchedItems = append(matchedItems, item)
			}
		}
	}

	if len(matchedItems) == 0 {
		return fmt.Errorf("没有找到匹配的文件: %s", pattern)
	}

	// 限制最大文件数
	if opts.MaxFiles > 0 && len(matchedItems) > opts.MaxFiles {
		matchedItems = matchedItems[:opts.MaxFiles]
	}

	// 交互模式确认
	if opts.Interactive {
		fmt.Printf("找到 %d 个匹配文件:\n", len(matchedItems))
		for i, item := range matchedItems {
			fmt.Printf("%d. %s (%s) - 删除时间: %s\n",
				i+1, item.Name, formatBytes(item.Size),
				item.DeletedTime.Format("2006-01-02 15:04:05"))
		}

		fmt.Print("确认恢复这些文件吗? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			return fmt.Errorf("用户取消操作")
		}
	}

	// 恢复文件
	successCount := 0
	for _, item := range matchedItems {
		if err := restoreSingleFileWindows(item); err != nil {
			fmt.Printf("恢复文件失败 %s: %v\n", item.Name, err)
		} else {
			fmt.Printf("成功恢复: %s -> %s\n", item.Name, item.OriginalPath)
			successCount++
		}
	}

	if successCount == 0 {
		return fmt.Errorf("所有文件恢复失败")
	}

	return nil
}

// listRecycleBinItemsWindows 列出Windows回收站中的所有项目
func listRecycleBinItemsWindows() ([]RecycleBinItem, error) {
	var items []RecycleBinItem

	// 获取当前用户的SID
	userSID, err := getCurrentUserSIDWindows()
	if err != nil {
		userSID = "S-1-5-18" // 使用系统SID作为默认值
	}

	// 获取所有驱动器
	drives := []string{"C", "D", "E", "F", "G", "H", "I", "J"}
	for _, drive := range drives {
		recyclePath := fmt.Sprintf("%s:\\$Recycle.Bin\\%s", drive, userSID)
		if _, err := os.Stat(recyclePath); os.IsNotExist(err) {
			// 尝试通用路径
			recyclePath = fmt.Sprintf("%s:\\RECYCLED", drive)
			if _, err := os.Stat(recyclePath); os.IsNotExist(err) {
				continue
			}
		}

		// 扫描回收站文件
		userItems := scanUserRecycleBinWindows(recyclePath)
		items = append(items, userItems...)
	}

	return items, nil
}

// scanUserRecycleBinWindows 扫描用户的回收站目录
func scanUserRecycleBinWindows(userRecyclePath string) []RecycleBinItem {
	var items []RecycleBinItem

	files, err := ioutil.ReadDir(userRecyclePath)
	if err != nil {
		return items
	}

	for _, file := range files {
		fileName := file.Name()

		// 跳过非信息文件
		if !strings.HasPrefix(strings.ToUpper(fileName), "$I") {
			continue
		}

		infoFile := filepath.Join(userRecyclePath, fileName)
		dataFile := filepath.Join(userRecyclePath, strings.Replace(strings.ToUpper(fileName), "$I", "$R", 1))

		if info, err := parseRecycleInfoFileWindows(infoFile, dataFile); err == nil {
			items = append(items, *info)
		}
	}

	return items
}

// parseRecycleInfoFileWindows 解析Windows回收站信息文件
func parseRecycleInfoFileWindows(infoPath, dataPath string) (*RecycleBinItem, error) {
	data, err := ioutil.ReadFile(infoPath)
	if err != nil {
		return nil, err
	}

	if len(data) < 24 {
		return nil, fmt.Errorf("信息文件格式无效")
	}

	// 跳过文件头，读取原始路径（偏移量24开始）
	pathStart := 24
	if len(data) <= pathStart {
		return nil, fmt.Errorf("信息文件格式无效")
	}

	// 解析UTF-16路径
	pathData := data[pathStart:]
	var originalPath string

	// 将UTF-16转换为字符串
	for i := 0; i < len(pathData)-1; i += 2 {
		if pathData[i] == 0 && pathData[i+1] == 0 {
			break
		}
		char := binary.LittleEndian.Uint16(pathData[i : i+2])
		if char == 0 {
			break
		}
		originalPath += string(rune(char))
	}

	// 获取数据文件信息
	var size int64
	var deletedTime time.Time
	if stat, err := os.Stat(dataPath); err == nil {
		size = stat.Size()
		deletedTime = stat.ModTime()
	}

	return &RecycleBinItem{
		OriginalPath: originalPath,
		CurrentPath:  dataPath,
		Name:         filepath.Base(originalPath),
		Size:         size,
		DeletedTime:  deletedTime,
		FileType:     getFileTypeByPath(originalPath),
	}, nil
}

// restoreSingleFileWindows 恢复单个Windows文件
func restoreSingleFileWindows(item RecycleBinItem) error {
	// 确保目标目录存在
	targetDir := filepath.Dir(item.OriginalPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("无法创建目标目录: %v", err)
	}

	// 直接重命名/移动文件
	err := os.Rename(item.CurrentPath, item.OriginalPath)
	if err != nil {
		// 如果重命名失败，尝试复制
		return copyFileWindows(item.CurrentPath, item.OriginalPath)
	}

	return nil
}

// copyFileWindows 复制文件作为恢复备选方案
func copyFileWindows(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// 删除源文件
	return os.Remove(src)
}

// getCurrentUserSIDWindows 获取当前用户SID
func getCurrentUserSIDWindows() (string, error) {
	// 简化实现，实际应该使用Windows API获取
	return "S-1-5-21", nil
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
