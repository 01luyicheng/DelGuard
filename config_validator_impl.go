package main

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// ProductionConfigValidator 生产环境配置验证器
type ProductionConfigValidator struct {
	strictMode bool
}

// NewProductionConfigValidator 创建生产环境配置验证器
func NewProductionConfigValidator(strictMode bool) *ProductionConfigValidator {
	return &ProductionConfigValidator{
		strictMode: strictMode,
	}
}

// ValidateProductionConfig 验证生产环境配置
func (pcv *ProductionConfigValidator) ValidateProductionConfig(config *Config) []string {
	var errors []string

	// 基础字段验证
	errors = append(errors, pcv.validateBasicFields(config)...)

	// 数值范围验证
	errors = append(errors, pcv.validateNumericRanges(config)...)

	// 安全设置验证
	errors = append(errors, pcv.validateSecuritySettings(config)...)

	return errors
}

// validateBasicFields 验证基础字段
func (pcv *ProductionConfigValidator) validateBasicFields(config *Config) []string {
	var errors []string

	// 验证版本信息
	if config.Version == "" {
		errors = append(errors, "配置版本不能为空")
	}

	// 验证语言代码
	if config.Language != "" && !pcv.isValidLanguageCode(config.Language) {
		errors = append(errors, fmt.Sprintf("无效的语言代码: %s", config.Language))
	}

	// 验证日志级别
	validLogLevels := []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
	if config.LogLevel != "" && !pcv.contains(validLogLevels, strings.ToUpper(config.LogLevel)) {
		errors = append(errors, fmt.Sprintf("无效的日志级别: %s", config.LogLevel))
	}

	// 验证安全模式
	validSafeModes := []string{"strict", "normal", "relaxed"}
	if config.SafeMode != "" && !pcv.contains(validSafeModes, config.SafeMode) {
		errors = append(errors, fmt.Sprintf("无效的安全模式: %s", config.SafeMode))
	}

	// 验证交互模式
	validInteractiveModes := []string{"always", "never", "confirm"}
	if config.InteractiveMode != "" && !pcv.contains(validInteractiveModes, config.InteractiveMode) {
		errors = append(errors, fmt.Sprintf("无效的交互模式: %s", config.InteractiveMode))
	}

	return errors
}

// validateNumericRanges 验证数值范围
func (pcv *ProductionConfigValidator) validateNumericRanges(config *Config) []string {
	var errors []string

	// 验证相似度阈值
	if config.SimilarityThreshold < 0 || config.SimilarityThreshold > 100 {
		errors = append(errors, fmt.Sprintf("相似度阈值必须在0-100之间，当前值: %.2f", config.SimilarityThreshold))
	}

	// 验证最大文件大小
	if config.MaxFileSize < 0 {
		errors = append(errors, "最大文件大小不能为负数")
	}

	if config.MaxFileSize > 100*GB {
		errors = append(errors, fmt.Sprintf("最大文件大小过大: %d bytes，建议不超过100GB", config.MaxFileSize))
	}

	// 验证回收站最大大小
	if config.TrashMaxSize < 0 {
		errors = append(errors, "回收站最大大小不能为负数")
	}

	// 验证最大备份文件数
	if config.MaxBackupFiles < 0 {
		errors = append(errors, "最大备份文件数不能为负数")
	}

	if config.MaxBackupFiles > 10000 {
		errors = append(errors, fmt.Sprintf("最大备份文件数过大: %d，建议不超过10000", config.MaxBackupFiles))
	}

	// 验证路径长度限制
	if config.MaxPathLength < 256 || config.MaxPathLength > 32767 {
		errors = append(errors, fmt.Sprintf("路径长度限制应在256-32767之间，当前值: %d", config.MaxPathLength))
	}

	// 验证并发操作数
	if config.MaxConcurrentOps < 1 || config.MaxConcurrentOps > 1000 {
		errors = append(errors, fmt.Sprintf("并发操作数应在1-1000之间，当前值: %d", config.MaxConcurrentOps))
	}

	return errors
}

// validateSecuritySettings 验证安全设置
func (pcv *ProductionConfigValidator) validateSecuritySettings(config *Config) []string {
	var errors []string

	// 在严格模式下，某些安全功能必须启用
	if pcv.strictMode {
		if !config.EnableSecurityChecks {
			errors = append(errors, "严格模式下必须启用安全检查")
		}

		if !config.EnablePathValidation {
			errors = append(errors, "严格模式下必须启用路径验证")
		}
	}

	// 检查恶意软件扫描设置
	if config.EnableMalwareScan {
		errors = append(errors, "注意：恶意软件扫描功能已被移除，此设置将被忽略")
	}

	return errors
}

// 辅助方法
func (pcv *ProductionConfigValidator) isValidLanguageCode(code string) bool {
	validCodes := []string{
		"en", "zh-cn", "zh-tw", "ja", "ko", "fr", "de", "es", "it", "ru",
		"pt", "ar", "hi", "th", "vi", "nl", "sv", "no", "fi",
	}
	return pcv.contains(validCodes, strings.ToLower(code))
}

func (pcv *ProductionConfigValidator) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (pcv *ProductionConfigValidator) isWindowsPlatform() bool {
	return runtime.GOOS == "windows"
}

func (pcv *ProductionConfigValidator) isLinuxPlatform() bool {
	return runtime.GOOS == "linux"
}

func (pcv *ProductionConfigValidator) isDarwinPlatform() bool {
	return runtime.GOOS == "darwin"
}

// ValidateURL 验证URL格式
func (pcv *ProductionConfigValidator) ValidateURL(urlStr string) error {
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

// ValidatePath 验证路径格式和安全性
func (pcv *ProductionConfigValidator) ValidatePath(path string) error {
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

// ValidateRegex 验证正则表达式
func (pcv *ProductionConfigValidator) ValidateRegex(pattern string) error {
	if pattern == "" {
		return nil
	}

	_, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("无效的正则表达式: %v", err)
	}

	return nil
}

// GetValidationSummary 获取验证摘要
func (pcv *ProductionConfigValidator) GetValidationSummary(errors []string) string {
	if len(errors) == 0 {
		return "✅ 配置验证通过，所有设置都符合生产环境要求"
	}

	summary := fmt.Sprintf("❌ 发现 %d 个配置问题：\n", len(errors))
	for i, err := range errors {
		summary += fmt.Sprintf("  %d. %s\n", i+1, err)
	}

	return summary
}
