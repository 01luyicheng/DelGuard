//go:build windows
// +build windows

package main

import (
	"delguard/utils"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// restoreFromTrashWindows Windows平台恢复实现
func restoreFromTrashWindows(pattern string, opts RestoreOptions) error {
	// 获取回收站中的所有文件
	items, err := listRecycleBinItemsWindows()
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
		// 支持通配符匹配（预编译一次正则）
		regex := wildcardToRegex(pattern)
		re := regexp.MustCompile(regex)
		for _, item := range items {
			matchedName := strings.ToLower(item.Name)
			if re.MatchString(matchedName) {
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
		
		if err := restoreSingleFileWindows(item); err != nil {
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

// listRecycleBinItemsWindows 列出Windows回收站中的所有项目
func listRecycleBinItemsWindows() ([]RecycleBinItem, error) {
	var items []RecycleBinItem

	// 遍历可能的盘符 A: - Z:
	for d := 'A'; d <= 'Z'; d++ {
		driveRoot := fmt.Sprintf("%c:\\", d)
		if _, err := os.Stat(driveRoot); err != nil {
			continue
		}

		// 优先尝试新式路径: <Drive>:\$Recycle.Bin
		baseRecycle := filepath.Join(driveRoot, "$Recycle.Bin")
		if info, err := os.Stat(baseRecycle); err == nil && info.IsDir() {
			// 遍历所有 SID 子目录
			entries, err := os.ReadDir(baseRecycle)
			if err == nil {
				for _, e := range entries {
					if !e.IsDir() {
						continue
					}
					sidPath := filepath.Join(baseRecycle, e.Name())
					userItems := scanUserRecycleBinWindows(sidPath)
					items = append(items, userItems...)
				}
			}
		}

		// 兼容旧式路径: <Drive>:\RECYCLED 或 RECYCLER
		legacyPaths := []string{
			filepath.Join(driveRoot, "RECYCLED"),
			filepath.Join(driveRoot, "RECYCLER"),
		}
		for _, legacy := range legacyPaths {
			if info, err := os.Stat(legacy); err == nil && info.IsDir() {
				userItems := scanUserRecycleBinWindows(legacy)
				items = append(items, userItems...)
			}
		}
	}

	return items, nil
}

// scanUserRecycleBinWindows 扫描用户的回收站目录
func scanUserRecycleBinWindows(userRecyclePath string) []RecycleBinItem {
	var items []RecycleBinItem

	entries, err := os.ReadDir(userRecyclePath)
	if err != nil {
		return items
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileName := entry.Name()

		// 跳过非信息文件
		upper := strings.ToUpper(fileName)
		if !strings.HasPrefix(upper, "$I") {
			continue
		}

		infoFile := filepath.Join(userRecyclePath, fileName)
		dataFile := filepath.Join(userRecyclePath, strings.Replace(upper, "$I", "$R", 1))

		if info, err := parseRecycleInfoFileWindows(infoFile, dataFile); err == nil {
			items = append(items, *info)
		}
	}

	return items
}

// parseRecycleInfoFileWindows 解析Windows回收站信息文件
func parseRecycleInfoFileWindows(infoPath, dataPath string) (*RecycleBinItem, error) {
	data, err := os.ReadFile(infoPath)
	if err != nil {
		return nil, err
	}

	if len(data) < 24 {
		return nil, fmt.Errorf(T("信息文件格式无效"))
	}

	// 跳过文件头，读取原始路径（偏移量24开始）
	pathStart := 24
	if len(data) <= pathStart {
		return nil, fmt.Errorf(T("信息文件格式无效"))
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

	// 尝试重命名（带简单重试，处理文件占用）
	const maxRetries = 3
	const retryDelay = 100 * time.Millisecond
	for i := 0; i < maxRetries; i++ {
		if err := os.Rename(item.CurrentPath, item.OriginalPath); err != nil {
			if i < maxRetries-1 && isFileInUse(err) {
				time.Sleep(retryDelay)
				continue
			}
			// 重命名仍失败则回退到复制
			if err := copyFileWindows(item.CurrentPath, item.OriginalPath); err != nil {
				return err
			}
			// 复制恢复成功后清理 $I 信息文件
			cleanupRecycleInfoWindows(item.CurrentPath)
			return nil
		}
		// 重命名成功后清理 $I 信息文件
		cleanupRecycleInfoWindows(item.CurrentPath)
		return nil
	}

	// 理论不可达
	return nil
}

// cleanupRecycleInfoWindows 删除与 $R 数据文件对应的 $I 信息文件
func cleanupRecycleInfoWindows(dataFilePath string) {
	base := filepath.Base(dataFilePath)
	dir := filepath.Dir(dataFilePath)
	upper := strings.ToUpper(base)
	infoName := upper
	if strings.HasPrefix(upper, "$R") {
		infoName = strings.Replace(upper, "$R", "$I", 1)
	}
	infoPath := filepath.Join(dir, infoName)
	_ = os.Remove(infoPath)
}

// copyFileWindows 复制文件
func copyFileWindows(src, dst string) error {
	if err := utils.CopyFile(src, dst); err != nil {
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
