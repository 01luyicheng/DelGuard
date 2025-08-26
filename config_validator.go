package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ConfigValidatorImpl 配置验证器实现
type ConfigValidatorImpl struct {
	strictMode bool
}

// NewConfigValidatorImpl 创建配置验证器实现
func NewConfigValidatorImpl(strictMode bool) *ConfigValidatorImpl {
	return &ConfigValidatorImpl{
		strictMode: strictMode,
	}
}

// ValidateConfig 验证配置的完整性和安全性
func (cv *ConfigValidatorImpl) ValidateConfig(config *Config) []ConfigValidationError {
	var errors []ConfigValidationError

	// 基础字段验证
	errors = append(errors, cv.validateBasicFields(config)...)

	// 数值范围验证
	errors = append(errors, cv.validateNumericRanges(config)...)

	// 路径安全验证
	errors = append(errors, cv.validatePathSecurity(config)...)

	// 网络配置验证
	errors = append(errors, cv.validateNetworkConfig(config)...)

	// 平台特定验证
	errors = append(errors, cv.validatePlatformSpecific(config)...)

	// 安全设置验证
	errors = append(errors, cv.validateSecuritySettings(config)...)

	return errors
}

// validateBasicFields 验证基础字段
func (cv *ConfigValidatorImpl) validateBasicFields(config *Config) []ConfigValidationError {
	var errors []ConfigValidationError

	// 验证版本信息
	if config.Version == "" {
		errors = append(errors, ConfigValidationError{
			Field:   "version",
			Message: "配置版本不能为空",
			Level:   "error",
		})
	}

	// 验证语言代码
	if config.Language != "" && !cv.isValidLanguageCode(config.Language) {
		errors = append(errors, ConfigValidationError{
			Field:   "language",
			Message: fmt.Sprintf("无效的语言代码: %s", config.Language),
			Level:   "warning",
		})
	}

	// 验证日志级别
	validLogLevels := []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
	if config.LogLevel != "" && !cv.contains(validLogLevels, strings.ToUpper(config.LogLevel)) {
		errors = append(errors, ConfigValidationError{
			Field:   "log_level",
			Message: fmt.Sprintf("无效的日志级别: %s，有效值: %v", config.LogLevel, validLogLevels),
			Level:   "error",
		})
	}

	// 验证安全模式
	validSafeModes := []string{"strict", "normal", "relaxed"}
	if config.SafeMode != "" && !cv.contains(validSafeModes, config.SafeMode) {
		errors = append(errors, ConfigValidationError{
			Field:   "safe_mode",
			Message: fmt.Sprintf("无效的安全模式: %s，有效值: %v", config.SafeMode, validSafeModes),
			Level:   "error",
		})
	}

	// 验证交互模式
	validInteractiveModes := []string{"always", "never", "confirm"}
	if config.InteractiveMode != "" && !cv.contains(validInteractiveModes, config.InteractiveMode) {
		errors = append(errors, ConfigValidationError{
			Field:   "interactive_mode",
			Message: fmt.Sprintf("无效的交互模式: %s，有效值: %v", config.InteractiveMode, validInteractiveModes),
			Level:   "error",
		})
	}

	return errors
}

// validateNumericRanges 验证数值范围
func (cv *ConfigValidator) validateNumericRanges(config *Config) []ConfigValidationError {
	var errors []ConfigValidationError

	// 验证相似度阈值
	if config.SimilarityThreshold < 0 || config.SimilarityThreshold > 100 {
		errors = append(errors, ConfigValidationError{
			Field:   "similarity_threshold",
			Message: fmt.Sprintf("相似度阈值必须在0-100之间，当前值: %.2f", config.SimilarityThreshold),
			Level:   "error",
		})
	}

	// 验证最大文件大小
	if config.MaxFileSize < 0 {
		errors = append(errors, ConfigValidationError{
			Field:   "max_file_size",
			Message: "最大文件大小不能为负数",
			Level:   "error",
		})
	}

	if config.MaxFileSize > 100*GB {
		errors = append(errors, ConfigValidationError{
			Field:   "max_file_size",
			Message: fmt.Sprintf("最大文件大小过大: %d bytes，建议不超过100GB", config.MaxFileSize),
			Level:   "warning",
		})
	}

	// 验证回收站最大大小
	if config.TrashMaxSize < 0 {
		errors = append(errors, ConfigValidationError{
			Field:   "trash_max_size",
			Message: "回收站最大大小不能为负数",
			Level:   "error",
		})
	}

	// 验证最大备份文件数
	if config.MaxBackupFiles < 0 {
		errors = append(errors, ConfigValidationError{
			Field:   "max_backup_files",
			Message: "最大备份文件数不能为负数",
			Level:   "error",
		})
	}

	if config.MaxBackupFiles > 10000 {
		errors = append(errors, ConfigValidationError{
			Field:   "max_backup_files",
			Message: fmt.Sprintf("最大备份文件数过大: %d，建议不超过10000", config.MaxBackupFiles),
			Level:   "warning",
		})
	}

	// 验证路径长度限制
	if config.MaxPathLength < 256 || config.MaxPathLength > 32767 {
		errors = append(errors, ConfigValidationError{
			Field:   "max_path_length",
			Message: fmt.Sprintf("路径长度限制应在256-32767之间，当前值: %d", config.MaxPathLength),
			Level:   "warning",
		})
	}

	// 验证并发操作数
	if config.MaxConcurrentOps < 1 || config.MaxConcurrentOps > 1000 {
		errors = append(errors, ConfigValidationError{
			Field:   "max_concurrent_ops",
			Message: fmt.Sprintf("并发操作数应在1-1000之间，当前值: %d", config.MaxConcurrentOps),
			Level:   "warning",
		})
	}

	return errors
}

// validatePathSecurity 验证路径安全性
func (cv *ConfigValidator) validatePathSecurity(config *Config) []ConfigValidationError {
	var errors []ConfigValidationError

	// 这里可以添加路径验证逻辑
	// 例如检查是否包含危险的路径遍历字符

	return errors
}

// validateNetworkConfig 验证网络配置
func (cv *ConfigValidator) validateNetworkConfig(config *Config) []ConfigValidationError {
	var errors []ConfigValidationError

	// 这里可以添加网络配置验证逻辑
	// 例如验证URL格式、端口范围等

	return errors
}

// validatePlatformSpecific 验证平台特定配置
func (cv *ConfigValidator) validatePlatformSpecific(config *Config) []ConfigValidationError {
	var errors []ConfigValidationError

	// Windows特定验证
	if config.Windows != nil {
		if config.Windows.UseUAC && !cv.isWindowsPlatform() {
			errors = append(errors, ConfigValidationError{
				Field:   "windows.use_uac",
				Message: "UAC设置只在Windows平台有效",
				Level:   "warning",
			})
		}
	}

	// Linux特定验证
	if config.Linux != nil {
		if config.Linux.UseSystemTrash && !cv.isLinuxPlatform() {
			errors = append(errors, ConfigValidationError{
				Field:   "linux.use_system_trash",
				Message: "系统回收站设置只在Linux平台有效",
				Level:   "warning",
			})
		}
	}

	// macOS特定验证
	if config.Darwin != nil {
		if config.Darwin.UseFinderTrash && !cv.isDarwinPlatform() {
			errors = append(errors, ConfigValidationError{
				Field:   "darwin.use_finder_trash",
				Message: "Finder回收站设置只在macOS平台有效",
				Level:   "warning",
			})
		}
	}

	return errors
}

// validateSecuritySettings 验证安全设置
func (cv *ConfigValidator) validateSecuritySettings(config *Config) []ConfigValidationError {
	var errors []ConfigValidationError

	// 在严格模式下，某些安全功能必须启用
	if cv.strictMode {
		if !config.EnableSecurityChecks {
			errors = append(errors, ConfigValidationError{
				Field:   "enable_security_checks",
				Message: "严格模式下必须启用安全检查",
				Level:   "error",
			})
		}

		if !config.EnablePathValidation {
			errors = append(errors, ConfigValidationError{
				Field:   "enable_path_validation",
				Message: "严格模式下必须启用路径验证",
				Level:   "error",
			})
		}
	}

	// 检查恶意软件扫描设置
	if config.EnableMalwareScan {
		errors = append(errors, ConfigValidationError{
			Field:   "enable_malware_scan",
			Message: "注意：恶意软件扫描功能已被移除，此设置将被忽略",
			Level:   "warning",
		})
	}

	return errors
}

// 辅助方法
func (cv *ConfigValidator) isValidLanguageCode(code string) bool {
	// 简单的语言代码验证
	validCodes := []string{
		"en", "zh-cn", "zh-tw", "ja", "ko", "fr", "de", "es", "it", "ru",
		"pt", "ar", "hi", "th", "vi", "nl", "sv", "no", "fi",
	}
	return cv.contains(validCodes, strings.ToLower(code))
}

func (cv *ConfigValidator) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (cv *ConfigValidator) isWindowsPlatform() bool {
	return strings.Contains(strings.ToLower(os.Getenv("OS")), "windows")
}

func (cv *ConfigValidator) isLinuxPlatform() bool {
	return strings.Contains(strings.ToLower(os.Getenv("OSTYPE")), "linux")
}

func (cv *ConfigValidator) isDarwinPlatform() bool {
	return strings.Contains(strings.ToLower(os.Getenv("OSTYPE")), "darwin")
}

// validateURL 验证URL格式
func (cv *ConfigValidator) validateURL(urlStr string) error {
	if urlStr == "" {
		return nil
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("无效的URL格式: %v", err)
	}

	// 检查协议
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("不支持的协议: %s", parsedURL.Scheme)
	}

	// 检查主机名
	if parsedURL.Host == "" {
		return fmt.Errorf("缺少主机名")
	}

	return nil
}

// validatePath 验证路径格式和安全性
func (cv *ConfigValidator) validatePath(path string) error {
	if path == "" {
		return nil
	}

	// 检查路径遍历
	if strings.Contains(path, "..") {
		return fmt.Errorf("路径包含危险的遍历字符: %s", path)
	}

	// 检查绝对路径
	if !filepath.IsAbs(path) {
		return fmt.Errorf("必须使用绝对路径: %s", path)
	}

	// 检查路径长度
	if len(path) > MaxPathLength {
		return fmt.Errorf("路径过长: %d 字符，最大允许: %d", len(path), MaxPathLength)
	}

	// 检查非法字符
	illegalChars := []string{"|", "<", ">", "\"", "*", "?"}
	for _, char := range illegalChars {
		if strings.Contains(path, char) {
			return fmt.Errorf("路径包含非法字符 '%s': %s", char, path)
		}
	}

	return nil
}

// validateRegex 验证正则表达式
func (cv *ConfigValidator) validateRegex(pattern string) error {
	if pattern == "" {
		return nil
	}

	_, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("无效的正则表达式: %v", err)
	}

	return nil
}
