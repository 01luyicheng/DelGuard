package main

import (
	"delguard/utils"
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
			utils.FormatBytes(info.Size()), utils.FormatBytes(fv.MaxFileSize)))
		result.IsValid = false
	}

	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(cleanPath))
	if len(fv.BlockedExtensions) > 0 {
		for _, blockedExt := range fv.BlockedExtensions {
			if ext == strings.ToLower(blockedExt) {
				result.Errors = append(result.Errors, fmt.Sprintf("不允许的文件扩展名: %s", ext))
				result.IsValid = false
				break
			}
		}
	}

	// 检查允许的扩展名
	if len(fv.AllowedExtensions) > 0 {
		allowed := false
		for _, allowedExt := range fv.AllowedExtensions {
			if ext == strings.ToLower(allowedExt) {
				allowed = true
				break
			}
		}
		if !allowed {
			result.Warnings = append(result.Warnings, fmt.Sprintf("文件扩展名 %s 不在推荐列表中", ext))
		}
	}

	// 检查文件名模式
	filename := filepath.Base(cleanPath)
	for _, pattern := range fv.BlockedPatterns {
		matched, err := regexp.MatchString(pattern, filename)
		if err != nil {
			return nil, fmt.Errorf("正则表达式错误: %v", err)
		}
		if matched {
			result.Errors = append(result.Errors, fmt.Sprintf("文件名包含非法模式: %s", pattern))
			result.IsValid = false
		}
	}

	// 检查被阻止的文件名
	for _, blockedName := range fv.BlockedFilenames {
		if strings.ToLower(filename) == strings.ToLower(blockedName) {
			result.Errors = append(result.Errors, fmt.Sprintf("不允许的文件名: %s", filename))
			result.IsValid = false
		}
	}

	// 检查隐藏文件
	result.IsHidden, _ = isHiddenFile(info, cleanPath)
	if result.IsHidden && !fv.AllowHiddenFiles {
		result.Warnings = append(result.Warnings, "文件是隐藏文件")
	}

	// 检查系统文件
	result.IsSystem = isSystemFile(info, cleanPath)
	if result.IsSystem && !fv.AllowSystemFiles {
		result.Warnings = append(result.Warnings, "文件是系统文件")
	}

	// 检查可执行文件
	result.IsExecutable = isExecutableFile(info, cleanPath)
	if result.IsExecutable {
		result.Warnings = append(result.Warnings, "文件是可执行文件")
	}

	// 检查文件权限
	if err := checkFilePermissions(cleanPath, info); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("文件权限检查失败: %v", err))
		result.IsValid = false
	}

	// 检查特殊文件类型
	if isSpecialFile(info, cleanPath) {
		result.Errors = append(result.Errors, "不支持的特殊文件类型")
		result.IsValid = false
	}

	// 检查是否为关键路径
	if IsCriticalPath(cleanPath) {
		result.Errors = append(result.Errors, "不允许操作关键系统路径")
		result.IsValid = false
	}

	return result, nil
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

// GetValidationSummary 获取验证结果摘要
func (fv *FileValidator) GetValidationSummary(results []*FileValidationResult) string {
	total := len(results)
	valid := 0
	invalid := 0

	for _, result := range results {
		if result.IsValid {
			valid++
		} else {
			invalid++
		}
	}

	return fmt.Sprintf("验证完成: 总计 %d 个文件，%d 个有效，%d 个无效", total, valid, invalid)
}

// PrintValidationResult 打印验证结果
func PrintValidationResult(result *FileValidationResult) {
    status := "✅"
    if !result.IsValid {
        status = "❌"
    }
    fmt.Printf("%s %s\n", status, result.FileName)

    if len(result.Errors) > 0 {
        // 汇总错误信息并记录
        logger.Error("文件验证失败", result.FileName, fmt.Errorf("验证错误"), "验证失败")
        for _, e := range result.Errors {
            logger.Error("验证详情", result.FileName, fmt.Errorf(e), "验证错误详情")
        }
    }

	if len(result.Warnings) > 0 {
		fmt.Println("  警告:")
		for _, warn := range result.Warnings {
			fmt.Printf("    - %s\n", warn)
		}
	}

	if len(result.Suggestions) > 0 {
		fmt.Println("  建议:")
		for _, suggestion := range result.Suggestions {
			fmt.Printf("    - %s\n", suggestion)
		}
	}
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
	case mode&os.ModeNamedPipe != 0:
		return "命名管道"
	case mode&os.ModeSocket != 0:
		return "套接字"
	default:
		return "未知类型"
	}
}

// isExecutableFile 检查文件是否为可执行文件
func isExecutableFile(info os.FileInfo, path string) bool {
	// Unix系统检查执行权限
	if runtime.GOOS != "windows" {
		return info.Mode()&0111 != 0
	}

	// Windows系统检查扩展名
	ext := strings.ToLower(filepath.Ext(path))
	executableExtensions := []string{".exe", ".bat", ".cmd", ".com", ".msi", ".scr"}
	for _, execExt := range executableExtensions {
		if ext == execExt {
			return true
		}
	}
	return false
}

// isHiddenFile 检查文件是否为隐藏文件
func isHiddenFile(info os.FileInfo, path string) (bool, error) {
	if runtime.GOOS == "windows" {
		return isWindowsHiddenFile(path), nil
	}

	// Unix系统检查文件名是否以点开头
	filename := filepath.Base(path)
	return strings.HasPrefix(filename, "."), nil
}

// isSystemFile 检查文件是否为系统文件
func isSystemFile(info os.FileInfo, path string) bool {
	if runtime.GOOS == "windows" {
		// Windows系统检查文件属性
		return isWindowsSystemFile(path)
	}

	// Unix系统检查路径
	systemPaths := []string{"/bin", "/sbin", "/usr/bin", "/usr/sbin", "/etc", "/lib", "/lib64"}
	for _, sysPath := range systemPaths {
		if strings.HasPrefix(path, sysPath) {
			return true
		}
	}
	return false
}
