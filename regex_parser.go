package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// RegexParser 正则表达式解析器
type RegexParser struct {
	pattern string
	regex   *regexp.Regexp
}

// NewRegexParser 创建新的正则表达式解析器
func NewRegexParser(pattern string) (*RegexParser, error) {
	// 检查是否为通配符模式
	if isWildcardPattern(pattern) {
		regexPattern := convertWildcardToRegex(pattern)
		regex, err := regexp.Compile(regexPattern)
		if err != nil {
			return nil, fmt.Errorf("通配符模式编译失败: %v", err)
		}
		return &RegexParser{
			pattern: pattern,
			regex:   regex,
		}, nil
	}

	// 尝试编译为正则表达式
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("正则表达式编译失败: %v", err)
	}

	return &RegexParser{
		pattern: pattern,
		regex:   regex,
	}, nil
}

// isWildcardPattern 检查是否为通配符模式
func isWildcardPattern(pattern string) bool {
	// 简单的通配符检测
	return strings.Contains(pattern, "*") || strings.Contains(pattern, "?")
}

// convertWildcardToRegex 将通配符转换为正则表达式
func convertWildcardToRegex(pattern string) string {
	// 转义正则表达式特殊字符，但保留通配符
	result := ""
	for i, char := range pattern {
		switch char {
		case '*':
			result += ".*"
		case '?':
			result += "."
		case '.', '^', '$', '+', '(', ')', '[', ']', '{', '}', '|', '\\':
			result += "\\" + string(char)
		default:
			result += string(char)
		}
		_ = i // 避免未使用变量警告
	}
	return "^" + result + "$"
}

// Match 检查文件名是否匹配模式
func (rp *RegexParser) Match(filename string) bool {
	return rp.regex.MatchString(filename)
}

// FindMatches 在目录中查找匹配的文件
func (rp *RegexParser) FindMatches(searchDir string, recursive bool) ([]string, error) {
	var matches []string

	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续搜索
		}

		// 跳过目录（除非递归搜索）
		if info.IsDir() {
			if !recursive && path != searchDir {
				return filepath.SkipDir
			}
			return nil
		}

		filename := filepath.Base(path)
		if rp.Match(filename) {
			matches = append(matches, path)
		}

		return nil
	})

	return matches, err
}

// GetPattern 获取原始模式
func (rp *RegexParser) GetPattern() string {
	return rp.pattern
}

// IsWildcard 检查是否为通配符模式
func (rp *RegexParser) IsWildcard() bool {
	return isWildcardPattern(rp.pattern)
}

// BatchOperationConfirm 批量操作确认
type BatchOperationConfirm struct {
	files     []string
	operation string
	force     bool
	pattern   string // 新增：原始模式
}

// NewBatchOperationConfirm 创建批量操作确认
func NewBatchOperationConfirm(files []string, operation string, force bool, pattern string) *BatchOperationConfirm {
	return &BatchOperationConfirm{
		files:     files,
		operation: operation,
		force:     force,
		pattern:   pattern,
	}
}

// Confirm 确认批量操作
func (boc *BatchOperationConfirm) Confirm() (bool, error) {
	if boc.force {
		return true, nil
	}

	if len(boc.files) == 0 {
		fmt.Println("❌ 没有找到匹配的文件")
		return false, nil
	}

	// 显示模式信息
	if boc.pattern != "" {
		fmt.Printf("🎯 模式: %s\n", boc.pattern)
	}
	fmt.Printf("⚠️  准备%s %d 个文件：\n\n", boc.operation, len(boc.files))

	// 显示文件列表（分页显示）
	pageSize := 10
	totalPages := (len(boc.files) + pageSize - 1) / pageSize
	currentPage := 1

	for {
		// 显示当前页的文件
		start := (currentPage - 1) * pageSize
		end := start + pageSize
		if end > len(boc.files) {
			end = len(boc.files)
		}

		fmt.Printf("📄 第 %d/%d 页：\n", currentPage, totalPages)
		totalSize := int64(0)
		for i := start; i < end; i++ {
			fileInfo, err := os.Stat(boc.files[i])
			var sizeStr string
			if err == nil {
				size := fileInfo.Size()
				totalSize += size
				sizeStr = formatFileSize(size)
			} else {
				sizeStr = "<unknown>"
			}
			fmt.Printf("  %d. %s (%s)\n", i+1, boc.files[i], sizeStr)
		}

		// 显示当前页总大小
		if totalSize > 0 {
			fmt.Printf("\n📊 当前页总大小: %s\n", formatFileSize(totalSize))
		}

		// 显示操作选项
		fmt.Printf("\n🎯 选项：\n")
		fmt.Printf("  y - 确认%s所有文件\n", boc.operation)
		fmt.Printf("  n - 取消操作\n")
		if totalPages > 1 {
			if currentPage < totalPages {
				fmt.Printf("  > - 下一页\n")
			}
			if currentPage > 1 {
				fmt.Printf("  < - 上一页\n")
			}
		}
		fmt.Printf("  s - 跳过确认（强制执行）\n")
		fmt.Printf("  i - 显示所有文件详细信息\n")
		fmt.Print("\n请选择: ")

		var input string
		if isStdinInteractive() {
			if s, ok := readLineWithTimeout(30 * time.Second); ok {
				input = strings.ToLower(strings.TrimSpace(s))
			} else {
				input = ""
			}
		} else {
			input = ""
		}

		switch input {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		case "s", "skip":
			boc.force = true
			return true, nil
		case "i", "info":
			boc.showDetailedInfo()
		case ">", "next":
			if currentPage < totalPages {
				currentPage++
			}
		case "<", "prev":
			if currentPage > 1 {
				currentPage--
			}
		default:
			fmt.Println("❌ 无效的选择，请重新输入")
		}

		fmt.Println() // 空行分隔
	}
}

// showDetailedInfo 显示详细信息
func (boc *BatchOperationConfirm) showDetailedInfo() {
	fmt.Printf("\n📊 所有文件详细信息：\n")
	fmt.Printf("──────────────────────────────\n")

	totalSize := int64(0)
	totalFiles := len(boc.files)
	fileTypes := make(map[string]int)

	for i, file := range boc.files {
		fileInfo, err := os.Stat(file)
		if err == nil {
			size := fileInfo.Size()
			totalSize += size
			modTime := fileInfo.ModTime().Format(TimeFormatStandard)
			ext := strings.ToLower(filepath.Ext(file))
			if ext == "" {
				ext = "<no ext>"
			}
			fileTypes[ext]++

			fmt.Printf("%3d. %-50s %10s %s\n", i+1,
				truncateString(file, 50),
				formatFileSize(size),
				modTime)
		} else {
			fmt.Printf("%3d. %-50s %10s %s\n", i+1,
				truncateString(file, 50),
				"<error>",
				"<unknown>")
		}
	}

	fmt.Printf("──────────────────────────────\n")
	fmt.Printf("📁 总文件数: %d\n", totalFiles)
	fmt.Printf("📊 总大小: %s\n", formatFileSize(totalSize))
	fmt.Printf("📄 文件类型分布：\n")
	for ext, count := range fileTypes {
		fmt.Printf("  %s: %d 个\n", ext, count)
	}
	fmt.Printf("\n按回车键继续...")
	if isStdinInteractive() {
		_, _ = readLineWithTimeout(20 * time.Second)
	}
}

// formatFileSize 格式化文件大小
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// GetFiles 获取文件列表
func (boc *BatchOperationConfirm) GetFiles() []string {
	return boc.files
}

// SetForce 设置强制模式
func (boc *BatchOperationConfirm) SetForce(force bool) {
	boc.force = force
}

// IsForce 检查是否为强制模式
func (boc *BatchOperationConfirm) IsForce() bool {
	return boc.force
}
