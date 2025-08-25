package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

// SpecialFileHandler 特殊文件处理器
type SpecialFileHandler struct {
	config *Config
	cache  *FileInfoCache // 文件信息缓存
	mu     sync.RWMutex   // 保护并发访问
}

// FileInfoCache 文件信息缓存
type FileInfoCache struct {
	cache map[string]*CachedFileInfo
	mu    sync.RWMutex
	ttl   time.Duration
}

// CachedFileInfo 缓存的文件信息
type CachedFileInfo struct {
	Info      os.FileInfo
	Issues    []FileIssue
	Timestamp time.Time
	Checksum  string
}

const (
	// DefaultCacheTTL 默认缓存生存时间
	DefaultCacheTTL = 5 * time.Minute
)

// NewSpecialFileHandler 创建特殊文件处理器
func NewSpecialFileHandler(config *Config) *SpecialFileHandler {
	return &SpecialFileHandler{
		config: config,
		cache: &FileInfoCache{
			cache: make(map[string]*CachedFileInfo),
			ttl:   DefaultCacheTTL, // 默认5分钟缓存
		},
	}
}

// FileIssue 文件问题类型
type FileIssue struct {
	Type        string            // 问题类型
	Description string            // 问题描述
	Severity    string            // 严重程度: info, warning, error, critical
	Suggestion  string            // 建议
	Recoverable bool              // 是否可恢复
	Metadata    map[string]string // 额外元数据
	Timestamp   time.Time         // 检测时间
}

// AnalyzeFile 分析文件的特殊情况
func (h *SpecialFileHandler) AnalyzeFile(path string) ([]FileIssue, error) {
	// 检查缓存
	if cached := h.getCachedInfo(path); cached != nil {
		return cached.Issues, nil
	}

	var issues []FileIssue

	// 获取文件信息
	info, err := os.Lstat(path) // 使用Lstat避免跟随符号链接
	if err != nil {
		return nil, WrapE("获取文件信息", path, err)
	}

	// 检查文件访问权限
	if accessIssue := h.checkFileAccess(path); accessIssue != nil {
		issues = append(issues, *accessIssue)
	}

	// 检查隐藏文件
	if h.isHiddenFile(info, path) {
		issues = append(issues, FileIssue{
			Type:        "hidden_file",
			Description: "检测到隐藏文件",
			Severity:    "warning",
			Suggestion:  "隐藏文件通常包含重要配置或系统数据，请谨慎删除",
			Recoverable: true,
			Metadata:    map[string]string{"type": "hidden"},
			Timestamp:   time.Now(),
		})
	}

	// 检查文件名长度
	if h.isLongFileName(path) {
		issues = append(issues, FileIssue{
			Type:        "long_filename",
			Description: "文件名过长",
			Severity:    "warning",
			Suggestion:  "超长文件名可能导致跨平台兼容性问题",
			Recoverable: true,
			Metadata:    map[string]string{"length": strconv.Itoa(len(filepath.Base(path)))},
			Timestamp:   time.Now(),
		})
	}

	// 检查特殊字符
	if chars := h.getSpecialCharacters(path); len(chars) > 0 {
		issues = append(issues, FileIssue{
			Type:        "special_chars",
			Description: "文件名包含特殊字符",
			Severity:    "info",
			Suggestion:  "包含特殊字符的文件名可能在某些系统上不被支持",
			Recoverable: true,
			Metadata:    map[string]string{"chars": strings.Join(chars, ",")},
			Timestamp:   time.Now(),
		})
	}

	// 检查Unicode问题
	if unicodeIssues := h.getUnicodeIssues(path); len(unicodeIssues) > 0 {
		issues = append(issues, FileIssue{
			Type:        "unicode_issue",
			Description: "文件名包含Unicode特殊字符",
			Severity:    "warning",
			Suggestion:  "某些Unicode字符可能导致显示或处理问题",
			Recoverable: true,
			Metadata:    map[string]string{"issues": strings.Join(unicodeIssues, ",")},
			Timestamp:   time.Now(),
		})
	}

	// 检查空格问题
	if spaceIssues := h.getSpaceIssues(path); len(spaceIssues) > 0 {
		issues = append(issues, FileIssue{
			Type:        "space_issue",
			Description: "文件名包含空格问题",
			Severity:    "warning",
			Suggestion:  "空格问题可能导致命令行操作困难",
			Recoverable: true,
			Metadata:    map[string]string{"issues": strings.Join(spaceIssues, ",")},
			Timestamp:   time.Now(),
		})
	}

	// 检查只读属性
	if h.isReadOnlyFile(info) {
		issues = append(issues, FileIssue{
			Type:        "readonly",
			Description: "文件为只读属性",
			Severity:    "warning",
			Suggestion:  "只读文件通常包含重要数据，删除前请确认",
			Recoverable: false,
			Metadata:    map[string]string{"mode": info.Mode().String()},
			Timestamp:   time.Now(),
		})
	}

	// 检查系统文件
	if h.isSystemFile(info, path) {
		issues = append(issues, FileIssue{
			Type:        "system_file",
			Description: "检测到系统文件",
			Severity:    "error",
			Suggestion:  "系统文件对系统运行至关重要，强烈建议不要删除",
			Recoverable: false,
			Metadata:    map[string]string{"path": path},
			Timestamp:   time.Now(),
		})
	}

	// 检查符号链接
	if h.isSymbolicLink(info) {
		target, _ := os.Readlink(path)
		issues = append(issues, FileIssue{
			Type:        "symlink",
			Description: "文件为符号链接",
			Severity:    "info",
			Suggestion:  "删除符号链接不会影响原始文件",
			Recoverable: true,
			Metadata:    map[string]string{"target": target},
			Timestamp:   time.Now(),
		})
	}

	// 检查硬链接
	if h.isHardLink(info) {
		issues = append(issues, FileIssue{
			Type:        "hardlink",
			Description: "文件可能是硬链接",
			Severity:    "warning",
			Suggestion:  "硬链接文件有多个引用，删除不会直接影响其他引用",
			Recoverable: true,
			Metadata:    map[string]string{"nlink": "multiple"},
			Timestamp:   time.Now(),
		})
	}

	// 棆查文件大小
	if sizeIssue := h.checkFileSize(info); sizeIssue != nil {
		issues = append(issues, *sizeIssue)
	}

	// 检查文件内容类型
	if contentIssue := h.checkFileContent(path, info); contentIssue != nil {
		issues = append(issues, *contentIssue)
	}

	// 检查文件时间戛
	if timeIssue := h.checkFileTimestamp(info); timeIssue != nil {
		issues = append(issues, *timeIssue)
	}

	// 检查文件名安全性
	if securityIssue := h.checkFilenameSecurity(path); securityIssue != nil {
		issues = append(issues, *securityIssue)
	}

	// 缓存结果
	h.cacheFileInfo(path, info, issues)

	return issues, nil
}

// isHiddenFile 检查是否为隐藏文件
func (h *SpecialFileHandler) isHiddenFile(info os.FileInfo, path string) bool {
	basename := filepath.Base(path)

	if runtime.GOOS == "windows" {
		// Windows系统：检查隐藏属性（简化实现）
		// 注意：os.ModeHidden 在一些 Go 版本中可能不可用
		// 这里使用简化的检查方法：以点开头的文件
		return strings.HasPrefix(basename, ".")
	} else {
		// Unix系统：以点开头的文件为隐藏文件
		return strings.HasPrefix(basename, ".")
	}
}

// isLongFileName 检查文件名是否过长
func (h *SpecialFileHandler) isLongFileName(path string) bool {
	basename := filepath.Base(path)

	// Windows文件名限制为255字符
	if runtime.GOOS == "windows" && len(basename) > 255 {
		return true
	}

	// Unix系统一般也是255字节限制
	if len(basename) > 255 {
		return true
	}

	// 路径总长度检查
	if runtime.GOOS == "windows" && len(path) > 260 {
		return true
	}

	// Unix系统路径长度限制通常是4096
	if runtime.GOOS != "windows" && len(path) > 4096 {
		return true
	}

	return false
}

// hasSpecialCharacters 检查是否包含特殊字符
func (h *SpecialFileHandler) hasSpecialCharacters(path string) bool {
	basename := filepath.Base(path)

	// 检查控制字符
	for _, r := range basename {
		if r < 32 || r == 127 {
			return true
		}
	}

	// 检查平台特定的非法字符
	if runtime.GOOS == "windows" {
		// Windows非法字符
		illegalChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
		for _, char := range illegalChars {
			if strings.Contains(basename, char) {
				return true
			}
		}
	}

	return false
}

// hasUnicodeIssues 检查Unicode问题
func (h *SpecialFileHandler) hasUnicodeIssues(path string) bool {
	basename := filepath.Base(path)

	// 检查UTF-8编码有效性
	if !utf8.ValidString(basename) {
		return true
	}

	// 检查Unicode控制字符
	for _, r := range basename {
		if unicode.IsControl(r) && r != '\t' && r != '\n' && r != '\r' {
			return true
		}

		// 检查Unicode方向字符（可能用于攻击）
		if r >= 0x202A && r <= 0x202E {
			return true
		}

		// 检查零宽字符
		if r == 0x200B || r == 0x200C || r == 0x200D || r == 0xFEFF {
			return true
		}
	}

	return false
}

// hasSpaceIssues 检查空格问题
func (h *SpecialFileHandler) hasSpaceIssues(path string) bool {
	basename := filepath.Base(path)

	// 检查前后空格
	if strings.TrimSpace(basename) != basename {
		return true
	}

	// 检查多个连续空格
	if strings.Contains(basename, "  ") {
		return true
	}

	// 检查其他空白字符
	for _, r := range basename {
		if unicode.IsSpace(r) && r != ' ' {
			return true
		}
	}

	return false
}

// isReadOnlyFile 检查是否为只读文件
func (h *SpecialFileHandler) isReadOnlyFile(info os.FileInfo) bool {
	mode := info.Mode()

	// 检查写权限
	if mode&0200 == 0 {
		return true
	}

	return false
}

// isSystemFile 检查是否为系统文件
func (h *SpecialFileHandler) isSystemFile(info os.FileInfo, path string) bool {
	// 使用现有的isSpecialFile函数
	return isSpecialFile(info, path)
}

// isSymbolicLink 检查是否为符号链接
func (h *SpecialFileHandler) isSymbolicLink(info os.FileInfo) bool {
	return info.Mode()&os.ModeSymlink != 0
}

// HandleSpecialFile 处理特殊文件，返回是否应该继续删除
func (h *SpecialFileHandler) HandleSpecialFile(path string, force bool) (bool, error) {
	issues, err := h.AnalyzeFile(path)
	if err != nil {
		return false, fmt.Errorf("分析文件失败: %v", err)
	}

	if len(issues) == 0 {
		return true, nil // 没有问题，可以继续
	}

	// 显示问题列表
	fmt.Printf(T("\n文件 '%s' 检测到以下问题:\n"), path)
	hasErrors := false
	hasWarnings := false

	for i, issue := range issues {
		symbol := "ℹ️"
		switch issue.Severity {
		case "warning":
			symbol = "⚠️"
			hasWarnings = true
		case "error":
			symbol = "❌"
			hasErrors = true
		}

		fmt.Printf(T("  %d. %s %s: %s\n"), i+1, symbol, issue.Description, issue.Suggestion)
	}

	// 如果强制模式，直接返回
	if force {
		if hasErrors {
			fmt.Printf(T("⚠️ 警告: 检测到严重问题，但使用了强制模式\n"))
		}
		return true, nil
	}

	// 如果有错误级别的问题，需要特别确认
	if hasErrors {
		fmt.Printf(T("\n❌ 检测到严重问题！\n"))
		if !h.confirmDangerousAction("检测到严重问题，是否仍要继续删除") {
			return false, nil
		}
	}

	// 如果有警告，询问用户
	if hasWarnings {
		if !h.confirmAction("检测到警告，是否继续删除") {
			return false, nil
		}
	}

	return true, nil
}

// confirmAction 确认操作
func (h *SpecialFileHandler) confirmAction(message string) bool {
	fmt.Printf(T("%s? (y/N): "), message)
	// 15秒超时，非交互/超时默认否定
	text, ok := readLineWithTimeout(15 * time.Second)
	if !ok {
		return false
	}
	input := strings.ToLower(strings.TrimSpace(text))
	return input == "y" || input == "yes"
}

// confirmDangerousAction 确认危险操作
func (h *SpecialFileHandler) confirmDangerousAction(message string) bool {
	fmt.Printf(T("⚠️  %s? \n"), message)
	fmt.Printf(T("请输入 'YES' 确认继续 (其他任何输入将取消): "))
	// 30秒超时，非交互/超时默认否定
	text, ok := readLineWithTimeout(30 * time.Second)
	if !ok {
		return false
	}
	input := strings.TrimSpace(text)
	return input == "YES"
}

// NormalizeFileName 标准化文件名，修复一些常见问题
func (h *SpecialFileHandler) NormalizeFileName(filename string) string {
	// 移除前后空格
	filename = strings.TrimSpace(filename)

	// 替换多个连续空格为单个空格
	spaceRegex := regexp.MustCompile(`\s+`)
	filename = spaceRegex.ReplaceAllString(filename, " ")

	// 移除控制字符
	var result strings.Builder
	for _, r := range filename {
		if !unicode.IsControl(r) || r == '\t' || r == '\n' || r == '\r' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// SuggestSaferFileName 建议更安全的文件名
func (h *SpecialFileHandler) SuggestSaferFileName(filename string) string {
	// 标准化文件名
	normalized := h.NormalizeFileName(filename)

	// 替换问题字符
	if runtime.GOOS == "windows" {
		// Windows非法字符替换
		illegalChars := map[string]string{
			"<":  "(",
			">":  ")",
			":":  "-",
			"\"": "'",
			"|":  "-",
			"?":  "",
			"*":  "",
		}

		for old, new := range illegalChars {
			normalized = strings.ReplaceAll(normalized, old, new)
		}
	}

	// 限制长度
	if len(normalized) > 200 {
		ext := filepath.Ext(normalized)
		base := strings.TrimSuffix(normalized, ext)
		if len(base) > 200-len(ext) {
			base = base[:200-len(ext)]
		}
		normalized = base + ext
	}

	return normalized
}

// getCachedInfo 获取缓存的文件信息
func (h *SpecialFileHandler) getCachedInfo(path string) *CachedFileInfo {
	h.cache.mu.RLock()
	defer h.cache.mu.RUnlock()

	if cached, exists := h.cache.cache[path]; exists {
		if time.Since(cached.Timestamp) < h.cache.ttl {
			// 检查文件是否被修改
			if info, err := os.Stat(path); err == nil {
				currentChecksum := h.calculateChecksum(path)
				if currentChecksum == cached.Checksum &&
					info.ModTime() == cached.Info.ModTime() {
					return cached
				}
			}
		}
		// 缓存过期或文件变更，删除旧缓存
		delete(h.cache.cache, path)
	}
	return nil
}

// cacheFileInfo 缓存文件信息
func (h *SpecialFileHandler) cacheFileInfo(path string, info os.FileInfo, issues []FileIssue) {
	h.cache.mu.Lock()
	defer h.cache.mu.Unlock()

	checksum := h.calculateChecksum(path)
	h.cache.cache[path] = &CachedFileInfo{
		Info:      info,
		Issues:    issues,
		Timestamp: time.Now(),
		Checksum:  checksum,
	}
}

// calculateChecksum 计算文件校验和
func (h *SpecialFileHandler) calculateChecksum(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	hash := md5.New()
	// 只读取前1KB计算校验和，提高性能
	buffer := make([]byte, 1024)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return ""
	}
	hash.Write(buffer[:n])
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// checkFileAccess 检查文件访问权限
func (h *SpecialFileHandler) checkFileAccess(path string) *FileIssue {
    // 检查读权限
    f, err := os.Open(path)
    if err != nil {
        return &FileIssue{
            Type:        "access_denied",
            Description: "无法访问文件",
            Severity:    "error",
            Suggestion:  "请检查文件权限或以管理员身份运行",
            Recoverable: false,
            Metadata:    map[string]string{"error": err.Error()},
            Timestamp:   time.Now(),
        }
    }
    // 成功打开后立即关闭，避免持有句柄导致后续操作（移动/删除）失败
    _ = f.Close()
    return nil
}

// getSpecialCharacters 获取特殊字符列表
func (h *SpecialFileHandler) getSpecialCharacters(path string) []string {
	basename := filepath.Base(path)
	var chars []string

	// 检查控制字符
	for _, r := range basename {
		if r < 32 || r == 127 {
			chars = append(chars, fmt.Sprintf("\\x%02x", r))
		}
	}

	// 检查平台特定的非法字符
	if runtime.GOOS == "windows" {
		illegalChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
		for _, char := range illegalChars {
			if strings.Contains(basename, char) {
				chars = append(chars, char)
			}
		}
	}

	return chars
}

// getUnicodeIssues 获取Unicode问题列表
func (h *SpecialFileHandler) getUnicodeIssues(path string) []string {
	basename := filepath.Base(path)
	var issues []string

	// 检查UTF-8编码有效性
	if !utf8.ValidString(basename) {
		issues = append(issues, "invalid_utf8")
	}

	// 检查Unicode控制字符
	for _, r := range basename {
		if unicode.IsControl(r) && r != '\t' && r != '\n' && r != '\r' {
			issues = append(issues, fmt.Sprintf("control_char_U+%04X", r))
		}

		// 检查Unicode方向字符（可能用于攻击）
		if r >= 0x202A && r <= 0x202E {
			issues = append(issues, fmt.Sprintf("direction_char_U+%04X", r))
		}

		// 检查零宽字符
		if r == 0x200B || r == 0x200C || r == 0x200D || r == 0xFEFF {
			issues = append(issues, fmt.Sprintf("zero_width_char_U+%04X", r))
		}
	}

	return issues
}

// getSpaceIssues 获取空格问题列表
func (h *SpecialFileHandler) getSpaceIssues(path string) []string {
	basename := filepath.Base(path)
	var issues []string

	// 检查前后空格
	if strings.TrimSpace(basename) != basename {
		issues = append(issues, "leading_trailing_spaces")
	}

	// 检查多个连续空格
	if strings.Contains(basename, "  ") {
		issues = append(issues, "multiple_spaces")
	}

	// 检查其他空白字符
	for _, r := range basename {
		if unicode.IsSpace(r) && r != ' ' {
			issues = append(issues, fmt.Sprintf("unusual_space_U+%04X", r))
		}
	}

	return issues
}

// isHardLink 检查是否为硬链接
func (h *SpecialFileHandler) isHardLink(info os.FileInfo) bool {
	// 在Unix系统中，如果链接数大于1，则可能是硬链接
	// 这里使用一个简化的检测方法
	// 实际实现中可能需要平台特定的代码
	return false // 简化实现
}

// checkFileSize 检查文件大小
func (h *SpecialFileHandler) checkFileSize(info os.FileInfo) *FileIssue {
	size := info.Size()
	if size == 0 {
		return &FileIssue{
			Type:        "empty_file",
			Description: "空文件",
			Severity:    "info",
			Suggestion:  "空文件可能是占位符或临时文件",
			Recoverable: true,
			Metadata:    map[string]string{"size": "0"},
			Timestamp:   time.Now(),
		}
	}

	// 检查大文件（超过1GB）
	if size > 1024*1024*1024 {
		return &FileIssue{
			Type:        "large_file",
			Description: "大文件",
			Severity:    "warning",
			Suggestion:  "大文件删除操作可能耗时较长",
			Recoverable: true,
			Metadata:    map[string]string{"size": strconv.FormatInt(size, 10)},
			Timestamp:   time.Now(),
		}
	}

	return nil
}

// checkFileContent 检查文件内容类型
func (h *SpecialFileHandler) checkFileContent(path string, info os.FileInfo) *FileIssue {
	if info.IsDir() {
		return nil
	}

	// 检查可执行文件
	if h.isExecutableFile(info) {
		return &FileIssue{
			Type:        "executable",
			Description: "可执行文件",
			Severity:    "warning",
			Suggestion:  "删除可执行文件可能影响系统功能",
			Recoverable: true,
			Metadata:    map[string]string{"type": "executable"},
			Timestamp:   time.Now(),
		}
	}

	// 检查脚本文件
	if h.isScriptFile(path) {
		return &FileIssue{
			Type:        "script",
			Description: "脚本文件",
			Severity:    "info",
			Suggestion:  "脚本文件可能包含重要的自动化逻辑",
			Recoverable: true,
			Metadata:    map[string]string{"type": "script"},
			Timestamp:   time.Now(),
		}
	}

	return nil
}

// checkFileTimestamp 检查文件时间戛
func (h *SpecialFileHandler) checkFileTimestamp(info os.FileInfo) *FileIssue {
	modTime := info.ModTime()
	now := time.Now()

	// 检查未来时间
	if modTime.After(now) {
		return &FileIssue{
			Type:        "future_timestamp",
			Description: "文件修改时间在未来",
			Severity:    "warning",
			Suggestion:  "文件时间戛异常，可能是系统时间问题",
			Recoverable: true,
			Metadata:    map[string]string{"modtime": modTime.Format(time.RFC3339)},
			Timestamp:   time.Now(),
		}
	}

	// 检查非常旧的文件（超过10年）
	if now.Sub(modTime) > 10*365*24*time.Hour {
		return &FileIssue{
			Type:        "very_old_file",
			Description: "非常古老的文件",
			Severity:    "info",
			Suggestion:  "这个文件可能已经不再需要",
			Recoverable: true,
			Metadata:    map[string]string{"age": now.Sub(modTime).String()},
			Timestamp:   time.Now(),
		}
	}

	return nil
}

// checkFilenameSecurity 检查文件名安全性
func (h *SpecialFileHandler) checkFilenameSecurity(path string) *FileIssue {
	basename := filepath.Base(path)

	// 检查可疑的文件名模式
	suspiciousPatterns := []string{
		`\.exe\.`, // 双扩展名
		`\.scr\.`,
		`\.bat\.`,
		`\.(js|vbs|cmd)\.`,
	}

	for _, pattern := range suspiciousPatterns {
		if matched, _ := regexp.MatchString(pattern, strings.ToLower(basename)); matched {
			return &FileIssue{
				Type:        "suspicious_filename",
				Description: "可疑的文件名模式",
				Severity:    "warning",
				Suggestion:  "文件名可能被设计用于欺骗或隐藏真实类型",
				Recoverable: true,
				Metadata:    map[string]string{"pattern": pattern},
				Timestamp:   time.Now(),
			}
		}
	}

	return nil
}

// isExecutableFile 检查是否为可执行文件
func (h *SpecialFileHandler) isExecutableFile(info os.FileInfo) bool {
	mode := info.Mode()
	return mode&0111 != 0 // 检查执行权限
}

// isScriptFile 检查是否为脚本文件
func (h *SpecialFileHandler) isScriptFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	scriptExts := []string{".sh", ".bat", ".cmd", ".ps1", ".py", ".pl", ".rb", ".js", ".vbs"}
	for _, scriptExt := range scriptExts {
		if ext == scriptExt {
			return true
		}
	}
	return false
}
