package main

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

// InputValidator 输入验证器
type InputValidator struct {
	config          *Config
	mu              sync.RWMutex
	trustedPatterns []string // 受信任的路径模式
	blockedPatterns []string // 阻止的路径模式
}

// ValidationResult 验证结果
type ValidationResult struct {
	IsValid     bool                   // 是否有效
	Errors      []string               // 错误列表
	Warnings    []string               // 警告列表
	Suggestions []string               // 建议列表
	Sanitized   string                 // 清理后的输入
	Metadata    map[string]interface{} // 附加元数据
}

// SecurityLevel 安全级别
type SecurityLevel int

const (
	SecurityLow SecurityLevel = iota
	SecurityMedium
	SecurityHigh
	SecurityStrict
)

// NewInputValidator 创建一个新的输入验证器实例
//
// 参数:
//   - config: 配置对象，包含验证规则和安全设置
//
// 返回值:
//   - *InputValidator: 输入验证器实例指针
func NewInputValidator(config *Config) *InputValidator {
	validator := &InputValidator{
		config:          config,
		trustedPatterns: getDefaultTrustedPatterns(),
		blockedPatterns: getDefaultBlockedPatterns(),
	}
	return validator
}

// getDefaultTrustedPatterns 获取默认受信任的路径模式
func getDefaultTrustedPatterns() []string {
	patterns := []string{
		"^[a-zA-Z]:" + regexp.QuoteMeta(string(filepath.Separator)) + "[^<>:\"|?*]*$", // Windows路径
		"^/[^\\x00]*$",    // Unix路径
		"^\\./[^\\x00]*$", // 相对路径
	}

	if runtime.GOOS == "windows" {
		sep := regexp.QuoteMeta(string(filepath.Separator))
		patterns = append(patterns,
			"^[a-zA-Z]:"+sep+"(?:[^<>:\"|?*"+sep+"]+"+sep+")*[^<>:\"|?*]*$",
		)
	}

	return patterns
}

// getDefaultBlockedPatterns 获取默认阻止的路径模式
func getDefaultBlockedPatterns() []string {
	sep := regexp.QuoteMeta(string(filepath.Separator))
	return []string{
		"\\.\\./",                  // 路径遍历
		sep + sep + "\\.\\." + sep, // Windows路径遍历
		"\\x00",                    // 空字节注入
		"[\\x01-\\x08\\x0B\\x0C\\x0E-\\x1F\\x7F]", // 控制字符
		"<script",                  // 潜在的脚本注入
		"javascript:",              // JavaScript协议
		"data:",                    // Data协议
		"file:",                    // File协议
		"(?i)(rm\\s+-rf|del\\s+/)", // 危险命令模式
	}
}

// ValidatePath 验证文件路径
func (v *InputValidator) ValidatePath(path string) *ValidationResult {
	result := &ValidationResult{
		IsValid:     true,
		Errors:      []string{},
		Warnings:    []string{},
		Suggestions: []string{},
		Metadata:    make(map[string]interface{}),
	}

	// 基本检查
	if path == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "路径不能为空")
		return result
	}

	// 检查路径长度
	if len(path) > 4096 {
		result.IsValid = false
		result.Errors = append(result.Errors, "路径长度超过限制 (4096)")
		return result
	}

	// 检查UTF-8编码
	if !utf8.ValidString(path) {
		result.IsValid = false
		result.Errors = append(result.Errors, "路径包含无效的UTF-8编码")
		return result
	}

	// 检查空字节注入
	if strings.Contains(path, "\x00") {
		result.IsValid = false
		result.Errors = append(result.Errors, "检测到空字节注入攻击")
		return result
	}

	// 检查路径遍历攻击
	if v.hasPathTraversal(path) {
		result.IsValid = false
		result.Errors = append(result.Errors, "检测到路径遍历攻击")
		return result
	}

	// 检查危险字符
	if dangerousChars := v.getDangerousCharacters(path); len(dangerousChars) > 0 {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("路径包含危险字符: %s", strings.Join(dangerousChars, ", ")))
		result.Metadata["dangerous_chars"] = dangerousChars
	}

	// 检查控制字符
	if controlChars := v.getControlCharacters(path); len(controlChars) > 0 {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("路径包含控制字符: %s", strings.Join(controlChars, ", ")))
		result.Metadata["control_chars"] = controlChars
	}

	// 检查Unicode方向字符（可能用于攻击）
	if v.hasUnicodeDirectionChars(path) {
		result.IsValid = false
		result.Errors = append(result.Errors, "检测到Unicode方向字符攻击")
		return result
	}

	// 检查零宽字符
	if v.hasZeroWidthChars(path) {
		result.Warnings = append(result.Warnings, "路径包含零宽字符")
	}

	// 平台特定验证
	if platformErrors := v.validatePlatformSpecific(path); len(platformErrors) > 0 {
		result.Errors = append(result.Errors, platformErrors...)
		result.IsValid = false
	}

	// 清理路径
	sanitized := v.sanitizePath(path)
	result.Sanitized = sanitized

	if sanitized != path {
		result.Warnings = append(result.Warnings, "路径已被清理")
		result.Metadata["original_path"] = path
	}

	return result
}

// ValidateArgument 验证命令行参数
func (v *InputValidator) ValidateArgument(arg string) *ValidationResult {
	result := &ValidationResult{
		IsValid:     true,
		Errors:      []string{},
		Warnings:    []string{},
		Suggestions: []string{},
		Metadata:    make(map[string]interface{}),
	}

	// 基本检查
	if len(arg) > 1024 {
		result.IsValid = false
		result.Errors = append(result.Errors, "参数长度超过限制 (1024)")
		return result
	}

	// 检查命令注入
	if v.hasCommandInjection(arg) {
		result.IsValid = false
		result.Errors = append(result.Errors, "检测到命令注入攻击")
		return result
	}

	// 检查脚本注入
	if v.hasScriptInjection(arg) {
		result.IsValid = false
		result.Errors = append(result.Errors, "检测到脚本注入攻击")
		return result
	}

	// 检查URL注入
	if v.hasURLInjection(arg) {
		result.Warnings = append(result.Warnings, "参数包含URL，请确认安全性")
	}

	// 清理参数
	sanitized := v.sanitizeArgument(arg)
	result.Sanitized = sanitized

	return result
}

// ValidateConfig 验证配置值
func (v *InputValidator) ValidateConfig(key, value string) *ValidationResult {
	result := &ValidationResult{
		IsValid:     true,
		Errors:      []string{},
		Warnings:    []string{},
		Suggestions: []string{},
		Metadata:    make(map[string]interface{}),
	}

	// 检查配置键
	if !v.isValidConfigKey(key) {
		result.IsValid = false
		result.Errors = append(result.Errors, "无效的配置键")
		return result
	}

	// 根据配置类型验证值
	switch key {
	case "max_file_size", "max_backup_files", "trash_max_size":
		if _, err := strconv.ParseInt(value, 10, 64); err != nil {
			result.IsValid = false
			result.Errors = append(result.Errors, "配置值必须是数字")
		}
	case "language":
		if !v.isValidLanguageCode(value) {
			result.Warnings = append(result.Warnings, "未知的语言代码")
		}
	case "log_level":
		if !v.isValidLogLevel(value) {
			result.IsValid = false
			result.Errors = append(result.Errors, "无效的日志级别")
		}
	default:
		// 通用字符串验证
		if len(value) > 1024 {
			result.IsValid = false
			result.Errors = append(result.Errors, "配置值过长")
		}
	}

	return result
}

// hasPathTraversal 检查路径遍历攻击
func (v *InputValidator) hasPathTraversal(path string) bool {
	// 标准化路径
	cleaned := filepath.Clean(path)

	// 检查..模式
	traversalPatterns := []string{
		"../",
		"..\\",
		"%2e%2e%2f", // URL编码的../
		"%2e%2e%5c", // URL编码的..\\
		"\\.\\.",    // Windows模式
		"..",
		"...",
		"....",
		"%2e%2e",          // URL编码的..
		"%252e%252e",      // 双重URL编码的..
		"%c0%ae%c0%ae",    // UTF-8编码的..
		"%c1%9e%c1%9e",    // UTF-8编码的..
		"%252e%252e%252f", // 三重URL编码的../
		"%252e%252e%255c", // 三重URL编码的..\\
	}

	for _, pattern := range traversalPatterns {
		if strings.Contains(strings.ToLower(path), pattern) {
			return true
		}
	}

	// 检查路径是否向上遍历
	if strings.Contains(cleaned, "..") {
		return true
	}

	// 检查规范化攻击
	if cleaned != path {
		// 如果规范化后的路径与原始路径不同，可能存在攻击
		return true
	}

	// 检查符号链接攻击
	if v.hasSymlinkAttack(path) {
		return true
	}

	// 检查空字节注入
	if strings.Contains(path, "\x00") {
		return true
	}

	// 检查Unicode方向字符攻击
	unicodePatterns := []string{
		"\u202a", // Left-to-right embedding
		"\u202b", // Right-to-left embedding
		"\u202c", // Pop directional formatting
		"\u202d", // Left-to-right override
		"\u202e", // Right-to-left override
		"\u200e", // Left-to-right mark
		"\u200f", // Right-to-left mark
	}

	for _, pattern := range unicodePatterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}

	// 检查零宽字符攻击
	zeroWidthPatterns := []string{
		"\u200b", // Zero-width space
		"\u200c", // Zero-width non-joiner
		"\u200d", // Zero-width joiner
		"\ufeff", // Zero-width no-break space
	}

	for _, pattern := range zeroWidthPatterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}

	return false
}

// hasSymlinkAttack 检查符号链接攻击
func (v *InputValidator) hasSymlinkAttack(path string) bool {
	// 检查路径中是否包含符号链接的特殊字符
	if strings.Contains(path, "->") {
		return true
	}

	// 检查路径中是否包含转义序列
	if strings.Contains(path, "\x1b") {
		return true
	}

	// 检查路径中是否包含控制字符
	for _, char := range path {
		if unicode.IsControl(char) && char != '\t' && char != '\n' && char != '\r' {
			return true
		}
	}

	return false
}

// getDangerousCharacters 获取危险字符列表
func (v *InputValidator) getDangerousCharacters(input string) []string {
	var dangerous []string

	dangerousChars := map[rune]string{
		'|':  "管道符",
		'&':  "逻辑AND",
		';':  "命令分隔符",
		'$':  "变量引用",
		'`':  "命令替换",
		'>':  "重定向符",
		'<':  "输入重定向",
		'*':  "通配符",
		'?':  "单字符通配符",
		'[':  "字符类开始",
		']':  "字符类结束",
		'{':  "大括号展开开始",
		'}':  "大括号展开结束",
		'(':  "子shell开始",
		')':  "子shell结束",
		'\'': "单引号",
		'"':  "双引号",
	}

	for _, char := range input {
		if desc, exists := dangerousChars[char]; exists {
			dangerous = append(dangerous, fmt.Sprintf("%c(%s)", char, desc))
		}
	}

	return dangerous
}

// getControlCharacters 获取控制字符列表
func (v *InputValidator) getControlCharacters(input string) []string {
	var controls []string

	for _, char := range input {
		if unicode.IsControl(char) && char != '\t' && char != '\n' && char != '\r' {
			controls = append(controls, fmt.Sprintf("\\x%02x", char))
		}
	}

	return controls
}

// hasUnicodeDirectionChars 检查Unicode方向字符
func (v *InputValidator) hasUnicodeDirectionChars(input string) bool {
	for _, char := range input {
		// Unicode方向字符范围
		if char >= 0x202A && char <= 0x202E {
			return true
		}
		// 其他可能危险的Unicode字符
		if char == 0x200E || char == 0x200F {
			return true
		}
	}
	return false
}

// hasZeroWidthChars 检查零宽字符
func (v *InputValidator) hasZeroWidthChars(input string) bool {
	zeroWidthChars := []rune{
		0x200B, // 零宽度空格
		0x200C, // 零宽度非连接符
		0x200D, // 零宽度连接符
		0xFEFF, // 字节顺序标记
	}

	for _, char := range input {
		for _, zw := range zeroWidthChars {
			if char == zw {
				return true
			}
		}
	}
	return false
}

// validatePlatformSpecific 平台特定验证
func (v *InputValidator) validatePlatformSpecific(path string) []string {
	var errors []string

	if runtime.GOOS == "windows" {
		// Windows特定验证
		// 检查保留名称
		basename := filepath.Base(path)
		reservedNames := []string{
			"CON", "PRN", "AUX", "NUL",
			"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
			"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
		}

		baseWithoutExt := strings.TrimSuffix(basename, filepath.Ext(basename))
		for _, reserved := range reservedNames {
			if strings.EqualFold(baseWithoutExt, reserved) {
				errors = append(errors, fmt.Sprintf("'%s' 是Windows保留名称", reserved))
			}
		}

		// 检查非法字符
		illegalChars := []rune{'<', '>', ':', '"', '|', '?', '*'}
		for _, char := range path {
			for _, illegal := range illegalChars {
				if char == illegal {
					errors = append(errors, fmt.Sprintf("Windows路径不能包含字符 '%c'", char))
					break
				}
			}
		}

		// 检查路径长度限制
		if len(path) > 260 {
			errors = append(errors, "Windows路径长度不能超过260字符")
		}
	}

	return errors
}

// hasCommandInjection 检查命令注入
func (v *InputValidator) hasCommandInjection(input string) bool {
	dangerousPatterns := []string{
		"\\|",                 // 管道
		"&&",                  // 逻辑AND
		"\\|\\|",              // 逻辑OR
		";",                   // 命令分隔符
		"`",                   // 命令替换
		"\\$\\(",              // 命令替换
		"\\$\\{",              // 变量展开
		">/",                  // 重定向到文件
		">>",                  // 追加重定向
		"<",                   // 输入重定向
		"\\\\x[0-9a-fA-F]{2}", // 十六进制编码
		"eval\\s",             // eval命令
		"exec\\s",             // exec命令
		"system\\s",           // system调用
	}

	for _, pattern := range dangerousPatterns {
		if matched, _ := regexp.MatchString(pattern, input); matched {
			return true
		}
	}

	return false
}

// hasScriptInjection 检查脚本注入
func (v *InputValidator) hasScriptInjection(input string) bool {
	lowerInput := strings.ToLower(input)

	dangerousPatterns := []string{
		"<script",
		"javascript:",
		"vbscript:",
		"data:text/html",
		"data:application/",
		"onclick",
		"onload",
		"onerror",
		"eval\\s*\\(",
		"document\\.",
		"window\\.",
	}

	for _, pattern := range dangerousPatterns {
		if matched, _ := regexp.MatchString(pattern, lowerInput); matched {
			return true
		}
	}

	return false
}

// hasURLInjection 检查URL注入
func (v *InputValidator) hasURLInjection(input string) bool {
	// 检查是否包含URL
	if strings.Contains(input, "://") {
		// 尝试解析URL
		if u, err := url.Parse(input); err == nil && u.Scheme != "" {
			// 检查是否为危险协议
			dangerousSchemes := []string{"javascript", "data", "file", "ftp"}
			for _, scheme := range dangerousSchemes {
				if strings.EqualFold(u.Scheme, scheme) {
					return true
				}
			}
		}
	}
	return false
}

// sanitizePath 清理路径
func (v *InputValidator) sanitizePath(path string) string {
	// 移除控制字符
	var result strings.Builder
	for _, char := range path {
		if !unicode.IsControl(char) || char == '\t' || char == '\n' || char == '\r' {
			result.WriteRune(char)
		}
	}

	// 标准化路径
	cleaned := filepath.Clean(result.String())

	// 移除零宽字符
	zeroWidthChars := []rune{0x200B, 0x200C, 0x200D, 0xFEFF}
	for _, zw := range zeroWidthChars {
		cleaned = strings.ReplaceAll(cleaned, string(zw), "")
	}

	return cleaned
}

// sanitizeArgument 清理参数
func (v *InputValidator) sanitizeArgument(arg string) string {
	// 移除危险字符
	dangerousChars := []string{"|", "&", ";", "$", "`", ">", "<"}
	cleaned := arg
	for _, char := range dangerousChars {
		cleaned = strings.ReplaceAll(cleaned, char, "")
	}

	// 移除控制字符
	var result strings.Builder
	for _, char := range cleaned {
		if !unicode.IsControl(char) || char == '\t' || char == '\n' || char == '\r' {
			result.WriteRune(char)
		}
	}

	return result.String()
}

// isValidConfigKey 检查配置键是否有效
func (v *InputValidator) isValidConfigKey(key string) bool {
	validKeys := []string{
		"use_recycle_bin", "interactive_mode", "language", "log_level",
		"safe_mode", "max_backup_files", "trash_max_size", "max_file_size",
		"max_path_length", "max_concurrent_ops", "enable_security_checks",
		"enable_malware_scan", "enable_path_validation", "enable_hidden_check",
		"enable_overwrite_protection", "backup_retention_days", "log_retention_days",
		"enable_telemetry", "telemetry_endpoint",
	}

	for _, validKey := range validKeys {
		if key == validKey {
			return true
		}
	}
	return false
}

// isValidLanguageCode 检查语言代码是否有效
func (v *InputValidator) isValidLanguageCode(code string) bool {
	validCodes := []string{"auto", "en", "zh", "zh-CN", "zh-TW", "ja", "ko", "es", "fr", "de", "ru"}
	for _, validCode := range validCodes {
		if code == validCode {
			return true
		}
	}
	return false
}

// isValidLogLevel 检查日志级别是否有效
func (v *InputValidator) isValidLogLevel(level string) bool {
	validLevels := []string{LogLevelDebugStr, LogLevelInfoStr, LogLevelWarnStr, LogLevelErrorStr, LogLevelFatalStr}
	for _, validLevel := range validLevels {
		if level == validLevel {
			return true
		}
	}
	return false
}

// SetSecurityLevel 设置安全级别
func (v *InputValidator) SetSecurityLevel(level SecurityLevel) {
	v.mu.Lock()
	defer v.mu.Unlock()

	switch level {
	case SecurityStrict:
		// 最严格模式：阻止所有可疑输入
		v.blockedPatterns = append(v.blockedPatterns,
			"[^a-zA-Z0-9._/-]", // 只允许基本字符
		)
	case SecurityHigh:
		// 高安全模式：额外的检查
		v.blockedPatterns = append(v.blockedPatterns,
			"\\\\x[0-9a-fA-F]{2}", // 十六进制编码
			"%[0-9a-fA-F]{2}",     // URL编码
		)
	}
}

// ValidateWithRecovery 带恢复的验证
func (v *InputValidator) ValidateWithRecovery(input string, inputType string) (*ValidationResult, error) {
	defer func() {
		if r := recover(); r != nil {
			// 记录panic但不崩溃
			fmt.Printf("输入验证发生panic: %v, 输入类型: %s, 输入内容: %s\n", r, inputType, input)
		}
	}()

	switch inputType {
	case "path":
		return v.ValidatePath(input), nil
	case "argument":
		return v.ValidateArgument(input), nil
	default:
		return &ValidationResult{
			IsValid: false,
			Errors:  []string{"未知的输入类型"},
		}, fmt.Errorf("未知的输入类型: %s", inputType)
	}
}
