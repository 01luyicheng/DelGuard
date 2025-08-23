package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// FileValidationResult 文件验证验证结果
type FileValidationResult struct {
	FileName     string
	IsValid      bool
	Errors       []string
	Warnings     []string
	Suggestions  []string
	FileSize     int64
	FileType     string
	IsHidden     bool
	IsSystem     bool
	IsExecutable bool
	IsSymlink    bool
}

// FileValidator 文件验证器
type FileValidator struct {
	MaxFileSize       int64
	AllowedExtensions []string
	BlockedExtensions []string
	BlockedPatterns   []string
	BlockedFilenames  []string
	AllowHiddenFiles  bool
	AllowSystemFiles  bool
	AllowSymlinks     bool
}

// NewFileValidator 创建新的文件验证器
func NewFileValidator() *FileValidator {
	return &FileValidator{
		MaxFileSize:       1024 * 1024 * 1024, // 1GB 默认最大文件大小
		AllowedExtensions: []string{".txt", ".doc", ".docx", ".pdf", ".jpg", ".png", ".gif", ".zip", ".rar"},
		BlockedExtensions: []string{".exe", ".bat", ".cmd", ".scr", ".com", ".pif", ".app", ".msi", ".jar", ".js", ".vbs", ".wsf"},
		BlockedPatterns:   []string{`\.\./`, `\.\.\\`, `^\s*$`}, // 路径遍历模式和空文件名
		BlockedFilenames:  []string{"desktop.ini", "thumbs.db", ".ds_store", "icon\r", "icon\n"},
		AllowHiddenFiles:  false,
		AllowSystemFiles:  false,
		AllowSymlinks:     true,
	}
}

// ValidateFile 验证单个文件
func (fv *FileValidator) ValidateFile(filePath string) (*FileValidationResult, error) {
	result := &FileValidationResult{
		FileName:    filePath,
		IsValid:     true,
		Errors:      make([]string, 0),
		Warnings:    make([]string, 0),
		Suggestions: make([]string, 0),
	}

	// 清理路径
	cleanPath := filepath.Clean(filePath)
	if cleanPath != filePath {
		result.Warnings = append(result.Warnings, "路径包含冗余部分，已清理")
	}

	// 获取文件信息（包括符号链接本身的信息）
	info, err := os.Lstat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			result.Errors = append(result.Errors, "文件不存在")
			result.Suggestions = append(result.Suggestions, "请检查文件路径是否正确")
			result.IsValid = false
			return result, nil
		}
		return nil, fmt.Errorf("无法获取文件信息: %v", err)
	}

	// 检查符号链接
	result.IsSymlink = info.Mode()&os.ModeSymlink != 0
	if result.IsSymlink && !fv.AllowSymlinks {
		result.Warnings = append(result.Warnings, "文件是符号链接")
		if !fv.AllowSymlinks {
			result.Errors = append(result.Errors, "不允许操作符号链接")
			result.IsValid = false
		}
	}

	result.FileSize = info.Size()
	result.FileType = getFileType(info)

	// 检查文件大小
	if info.Size() > fv.MaxFileSize {
		result.Errors = append(result.Errors, fmt.Sprintf("文件大小超过限制 (%s > %s)",
			formatBytes(info.Size()), formatBytes(fv.MaxFileSize)))
		result.IsValid = false
	}

	// 检查扩展名
	ext := strings.ToLower(filepath.Ext(filePath))

	// 检查是否在阻止列表中
	for _, blockedExt := range fv.BlockedExtensions {
		if ext == strings.ToLower(blockedExt) {
			result.Errors = append(result.Errors, fmt.Sprintf("不支持的文件类型: %s", ext))
			result.IsValid = false
			break
		}
	}

	// 如果有允许列表，检查是否在允许列表中
	if len(fv.AllowedExtensions) > 0 {
		allowed := false
		for _, allowedExt := range fv.AllowedExtensions {
			if ext == strings.ToLower(allowedExt) {
				allowed = true
				break
			}
		}
		if !allowed {
			result.Errors = append(result.Errors, fmt.Sprintf("文件类型不在允许列表中: %s", ext))
			result.IsValid = false
		}
	}

	// 检查文件名
	filename := strings.ToLower(filepath.Base(filePath))
	for _, blockedName := range fv.BlockedFilenames {
		if filename == strings.ToLower(blockedName) {
			result.Errors = append(result.Errors, fmt.Sprintf("不允许操作系统文件: %s", filename))
			result.IsValid = false
			break
		}
	}

	// 检查隐藏文件
	isHidden, err := isHiddenFile(info, filePath)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("检查隐藏文件属性时出错: %v", err))
	} else {
		result.IsHidden = isHidden
		if isHidden && !fv.AllowHiddenFiles {
			result.Warnings = append(result.Warnings, "文件是隐藏文件")
			// 隐藏文件默认不阻止，除非配置明确禁止
		}
	}

	// 检查可执行文件
	result.IsExecutable = isExecutableFile(info, filePath)
	if result.IsExecutable {
		result.Warnings = append(result.Warnings, "文件是可执行文件")
	}

	// 检查系统文件（仅Windows）
	isSystem, err := isSystemFile(info, filePath)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("检查系统文件属性时出错: %v", err))
	} else if isSystem {
		result.IsSystem = isSystem
		result.Warnings = append(result.Warnings, "文件是系统文件")
		if !fv.AllowSystemFiles {
			result.Errors = append(result.Errors, "不允许操作系统文件")
			result.IsValid = false
		}
	}

	// 检查路径遍历模式
	for _, pattern := range fv.BlockedPatterns {
		matched, err := regexp.MatchString(pattern, filePath)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("检查路径模式时出错: %v", err))
			continue
		}
		if matched {
			result.Errors = append(result.Errors, fmt.Sprintf("文件路径包含非法模式: %s", pattern))
			result.IsValid = false
		}
	}

	// 检查路径长度
	if len(filePath) > 260 { // Windows MAX_PATH 限制
		result.Warnings = append(result.Warnings, "文件路径较长，可能导致兼容性问题")
	}

	return result, nil
}

// checkPathTraversal 检查路径遍历攻击
func (fv *FileValidator) checkPathTraversal(filePath string) error {
	// 检查相对路径遍历
	cleanPath := filepath.Clean(filePath)
	if strings.Contains(cleanPath, ".."+string(filepath.Separator)) {
		return fmt.Errorf("检测到路径遍历攻击模式")
	}

	// 检查正则表达式模式
	for _, pattern := range fv.BlockedPatterns {
		matched, err := regexp.MatchString(pattern, filePath)
		if err != nil {
			continue
		}
		if matched {
			return fmt.Errorf("检测到非法路径模式: %s", pattern)
		}
	}

	return nil
}

// checkMaliciousPatterns 检查恶意模式
func (fv *FileValidator) checkMaliciousPatterns(filePath string) error {
	// 检查文件名中的恶意模式
	filename := filepath.Base(filePath)
	maliciousPatterns := []string{
		"dropbox", "mega", "gdrive", // 云存储相关
		"password", "passwd", "credential", // 凭据相关
		"wallet", "crypto", "bitcoin", // 加密货币相关
		"key", "private", "secret", // 秘密相关
	}

	lowerFilename := strings.ToLower(filename)
	for _, pattern := range maliciousPatterns {
		if strings.Contains(lowerFilename, pattern) {
			return fmt.Errorf("检测到可疑文件名模式: %s", pattern)
		}
	}

	return nil
}

// ValidateBatch 批量验证文件
func (fv *FileValidator) ValidateBatch(filePaths []string) ([]*FileValidationResult, error) {
	results := make([]*FileValidationResult, 0, len(filePaths))

	for _, filePath := range filePaths {
		result, err := fv.ValidateFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("验证文件 %s 时出错: %v", filePath, err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetValidationSummary 获取验证摘要
func (fv *FileValidator) GetValidationSummary(results []*FileValidationResult) string {
	total := len(results)
	valid := 0
	invalid := 0
	warnings := 0

	for _, result := range results {
		if result.IsValid {
			valid++
		} else {
			invalid++
		}
		if len(result.Warnings) > 0 {
			warnings++
		}
	}

	return fmt.Sprintf("验证完成: 总计 %d 个文件, 有效 %d 个, 无效 %d 个, %d 个有警告",
		total, valid, invalid, warnings)
}

// getFileType 获取文件类型描述
func getFileType(info os.FileInfo) string {
	mode := info.Mode()
	switch {
	case mode.IsDir():
		return "目录"
	case mode.IsRegular():
		return "普通文件"
	case mode&os.ModeSymlink != 0:
		return "符号链接"
	case mode&os.ModeDevice != 0:
		return "设备文件"
	case mode&os.ModeSocket != 0:
		return "套接字"
	case mode&os.ModeNamedPipe != 0:
		return "命名管道"
	default:
		return "未知类型"
	}
}

// isExecutableFile 检查文件是否可执行
func isExecutableFile(info os.FileInfo, filePath string) bool {
	if info == nil {
		return false
	}

	if info.IsDir() {
		return false
	}

	mode := info.Mode()
	if runtime.GOOS == "windows" {
		// Windows: 检查文件扩展名
		ext := strings.ToLower(filepath.Ext(filePath))
		executableExts := []string{".exe", ".bat", ".cmd", ".com", ".scr", ".msi", ".ps1"}
		for _, executableExt := range executableExts {
			if ext == executableExt {
				return true
			}
		}
		return false
	}

	// Unix: 检查执行权限位
	return mode&0111 != 0
}

// isHiddenFile 检查文件是否为隐藏文件
func isHiddenFile(info os.FileInfo, filePath string) (bool, error) {
	if runtime.GOOS == "windows" {
		// Windows: 简化实现，检查文件名
		filename := filepath.Base(filePath)
		return strings.HasPrefix(filename, "."), nil
	}

	// Unix: 检查文件名是否以点开头
	filename := filepath.Base(filePath)
	return strings.HasPrefix(filename, "."), nil
}

// isSystemFile 检查文件是否为系统文件（仅Windows）
func isSystemFile(info os.FileInfo, filePath string) (bool, error) {
	if runtime.GOOS != "windows" {
		return false, nil
	}

	// Windows: 简化实现，检查特定系统目录
	systemPaths := []string{
		"C:\\Windows",
		"C:\\Program Files",
		"C:\\System",
	}

	cleanPath := filepath.Clean(strings.ToLower(filePath))
	for _, sysPath := range systemPaths {
		if strings.HasPrefix(cleanPath, strings.ToLower(sysPath)) {
			return true, nil
		}
	}

	return false, nil
}

// PrintValidationResult 打印验证结果
func PrintValidationResult(result *FileValidationResult) {
	status := "✅"
	if !result.IsValid {
		status = "❌"
	} else if len(result.Warnings) > 0 {
		status = "⚠️"
	}

	fmt.Printf("%s %s\n", status, result.FileName)

	if result.FileSize > 0 {
		fmt.Printf("   大小: %s\n", formatBytes(result.FileSize))
	}

	if result.IsHidden {
		fmt.Printf("   属性: 隐藏文件\n")
	}

	if result.IsSystem {
		fmt.Printf("   属性: 系统文件\n")
	}

	if result.IsExecutable {
		fmt.Printf("   属性: 可执行文件\n")
	}

	if result.IsSymlink {
		fmt.Printf("   属性: 符号链接\n")
	}

	for _, warning := range result.Warnings {
		fmt.Printf("   ⚠️  警告: %s\n", warning)
	}

	for _, err := range result.Errors {
		fmt.Printf("   ❌ 错误: %s\n", err)
	}

	if len(result.Suggestions) > 0 {
		for _, suggestion := range result.Suggestions {
			fmt.Printf("   💡 建议: %s\n", suggestion)
		}
	}

	fmt.Println()
}

// formatBytes 格式化字节数为人类可读格式
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
