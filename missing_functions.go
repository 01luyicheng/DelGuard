package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// ValidateFilePath 验证文件路径的有效性
func ValidateFilePath(path string) error {
	if path == "" {
		return fmt.Errorf("文件路径不能为空")
	}

	// 使用现有的PathUtils进行验证
	return PathUtils.ValidatePath(path)
}

// MoveToRecycleBin 将文件移动到回收站
func MoveToRecycleBin(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("文件路径不能为空")
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", filePath)
	}

	// 验证路径安全性
	if err := ValidateFilePath(filePath); err != nil {
		return fmt.Errorf("路径验证失败: %v", err)
	}

	// 使用现有的回收站功能
	return moveToTrashPlatform(filePath)
}

// GetConfigPath 获取配置文件路径
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("无法获取用户主目录: %v", err)
	}

	configDir := filepath.Join(homeDir, ".delguard")
	configFile := filepath.Join(configDir, "config.json")

	// 确保配置目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("无法创建配置目录: %v", err)
	}

	return configFile, nil
}

// IsValidConfigKey 检查配置键是否有效
func IsValidConfigKey(key string) bool {
	validKeys := map[string]bool{
		"interactive":    true,
		"use_trash":      true,
		"safe_mode":      true,
		"max_file_size":  true,
		"language":       true,
		"log_level":      true,
		"backup_enabled": true,
		"confirm_delete": true,
		"show_progress":  true,
		"color_output":   true,
	}

	return validKeys[key]
}

// IsValidLogLevel 检查日志级别是否有效
func IsValidLogLevel(level string) bool {
	validLevels := map[string]bool{
		"debug":  true,
		"info":   true,
		"warn":   true,
		"error":  true,
		"fatal":  true,
		"silent": true,
	}

	return validLevels[level]
}

// IsValidLanguageCode 检查语言代码是否有效
func IsValidLanguageCode(code string) bool {
	validCodes := map[string]bool{
		"en":    true,
		"zh-cn": true,
		"zh-tw": true,
		"ja":    true,
		"ko":    true,
		"fr":    true,
		"de":    true,
		"es":    true,
		"ru":    true,
	}

	return validCodes[code]
}
