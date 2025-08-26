package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SmartConfigManager 智能配置管理器
type SmartConfigManager struct {
	configPath   string
	backupDir    string
	config       *Config
	lastModified time.Time
	watchers     []ConfigWatcher
}

// ConfigWatcher 配置监听器接口
type ConfigWatcher interface {
	OnConfigChanged(oldConfig, newConfig *Config) error
}

// ConfigValidationError 配置验证错误
type ConfigValidationError struct {
	Field   string `json:"field"`
	Value   string `json:"value"`
	Message string `json:"message"`
	Level   string `json:"level"` // "error", "warning", "info"
}

// ConfigValidationResult 配置验证结果
type ConfigValidationResult struct {
	Valid   bool                    `json:"valid"`
	Errors  []ConfigValidationError `json:"errors"`
	Fixed   []string                `json:"fixed,omitempty"`
	Backups []string                `json:"backups,omitempty"`
}

// NewSmartConfigManager 创建智能配置管理器
func NewSmartConfigManager(configPath string) *SmartConfigManager {
	return &SmartConfigManager{
		configPath: configPath,
		backupDir:  filepath.Join(filepath.Dir(configPath), "backups"),
		watchers:   make([]ConfigWatcher, 0),
	}
}

// LoadConfigWithFallback 加载配置，支持容错和回退
func (scm *SmartConfigManager) LoadConfigWithFallback() (*Config, error) {
	fmt.Println("🔄 加载配置文件...")

	// 尝试加载主配置文件
	config, err := scm.loadConfigFile(scm.configPath)
	if err == nil {
		// 验证配置
		if result := scm.ValidateConfig(config); result.Valid {
			scm.config = config
			scm.updateLastModified()
			fmt.Println("✅ 配置文件加载成功")
			return config, nil
		} else {
			fmt.Printf("⚠️  配置文件存在问题，尝试自动修复...\n")
			// 尝试自动修复
			if fixedConfig, fixed := scm.autoFixConfig(config, result.Errors); fixed {
				scm.config = fixedConfig
				// 备份原配置并保存修复后的配置
				scm.backupConfig("auto-fix")
				scm.saveConfig(fixedConfig)
				fmt.Println("✅ 配置已自动修复并保存")
				return fixedConfig, nil
			}
		}
	}

	fmt.Printf("❌ 主配置文件加载失败: %v\n", err)

	// 尝试从备份恢复
	if backupConfig, err := scm.loadFromBackup(); err == nil {
		fmt.Println("✅ 从备份恢复配置成功")
		scm.config = backupConfig
		return backupConfig, nil
	}

	// 生成默认配置
	fmt.Println("🔧 生成默认配置...")
	defaultConfig := scm.generateDefaultConfig()
	scm.config = defaultConfig

	// 保存默认配置
	if err := scm.saveConfig(defaultConfig); err != nil {
		fmt.Printf("⚠️  保存默认配置失败: %v\n", err)
	} else {
		fmt.Println("✅ 默认配置已生成并保存")
	}

	return defaultConfig, nil
}

// loadConfigFile 加载配置文件
func (scm *SmartConfigManager) loadConfigFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 处理不同格式的配置文件
	var config Config

	// 尝试JSON格式
	if strings.HasSuffix(path, ".json") || strings.HasSuffix(path, ".jsonc") {
		// 处理JSONC注释
		if strings.HasSuffix(path, ".jsonc") {
			data = scm.removeComments(data)
		}

		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("JSON格式错误: %v", err)
		}
	} else {
		// 默认尝试JSON
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("配置格式错误: %v", err)
		}
	}

	return &config, nil
}

// ValidateConfig 验证配置文件
func (scm *SmartConfigManager) ValidateConfig(config *Config) ConfigValidationResult {
	result := ConfigValidationResult{
		Valid:  true,
		Errors: make([]ConfigValidationError, 0),
	}

	// 验证基本字段
	scm.validateBasicFields(config, &result)

	// 验证数值范围
	scm.validateNumericRanges(config, &result)

	// 验证路径和文件
	scm.validatePaths(config, &result)

	// 验证平台特定设置
	scm.validatePlatformSettings(config, &result)

	// 检查是否有错误
	for _, err := range result.Errors {
		if err.Level == "error" {
			result.Valid = false
			break
		}
	}

	return result
}

// validateBasicFields 验证基本字段
func (scm *SmartConfigManager) validateBasicFields(config *Config, result *ConfigValidationResult) {
	// 验证语言设置
	validLanguages := []string{"zh-cn", "en-us", "ja-jp", "ko-kr", "fr-fr", "de-de", "es-es"}
	if !scm.contains(validLanguages, config.Language) {
		result.Errors = append(result.Errors, ConfigValidationError{
			Field:   "language",
			Value:   config.Language,
			Message: "不支持的语言设置，将使用默认语言 zh-cn",
			Level:   "warning",
		})
	}

	// 验证交互模式
	validModes := []string{"always", "confirm", "never"}
	if !scm.contains(validModes, config.InteractiveMode) {
		result.Errors = append(result.Errors, ConfigValidationError{
			Field:   "interactive_mode",
			Value:   config.InteractiveMode,
			Message: "无效的交互模式，将使用默认模式 confirm",
			Level:   "warning",
		})
	}

	// 验证日志级别
	validLogLevels := []string{"debug", "info", "warn", "error"}
	if !scm.contains(validLogLevels, config.LogLevel) {
		result.Errors = append(result.Errors, ConfigValidationError{
			Field:   "log_level",
			Value:   config.LogLevel,
			Message: "无效的日志级别，将使用默认级别 info",
			Level:   "warning",
		})
	}

	// 验证安全模式
	validSafeModes := []string{"strict", "normal", "relaxed"}
	if !scm.contains(validSafeModes, config.SafeMode) {
		result.Errors = append(result.Errors, ConfigValidationError{
			Field:   "safe_mode",
			Value:   config.SafeMode,
			Message: "无效的安全模式，将使用默认模式 normal",
			Level:   "warning",
		})
	}
}

// validateNumericRanges 验证数值范围
func (scm *SmartConfigManager) validateNumericRanges(config *Config, result *ConfigValidationResult) {
	// 验证文件大小限制
	if config.MaxFileSize < 0 {
		result.Errors = append(result.Errors, ConfigValidationError{
			Field:   "max_file_size",
			Value:   fmt.Sprintf("%d", config.MaxFileSize),
			Message: "文件大小限制不能为负数",
			Level:   "error",
		})
	} else if config.MaxFileSize > 10*1024*1024*1024 { // 10GB
		result.Errors = append(result.Errors, ConfigValidationError{
			Field:   "max_file_size",
			Value:   fmt.Sprintf("%d", config.MaxFileSize),
			Message: "文件大小限制过大，建议不超过10GB",
			Level:   "warning",
		})
	}

	// 验证并发操作数
	if config.MaxConcurrentOps <= 0 {
		result.Errors = append(result.Errors, ConfigValidationError{
			Field:   "max_concurrent_ops",
			Value:   fmt.Sprintf("%d", config.MaxConcurrentOps),
			Message: "并发操作数必须大于0",
			Level:   "error",
		})
	} else if config.MaxConcurrentOps > 100 {
		result.Errors = append(result.Errors, ConfigValidationError{
			Field:   "max_concurrent_ops",
			Value:   fmt.Sprintf("%d", config.MaxConcurrentOps),
			Message: "并发操作数过大，可能影响系统性能",
			Level:   "warning",
		})
	}

	// 验证备份保留天数
	if config.BackupRetentionDays < 0 {
		result.Errors = append(result.Errors, ConfigValidationError{
			Field:   "backup_retention_days",
			Value:   fmt.Sprintf("%d", config.BackupRetentionDays),
			Message: "备份保留天数不能为负数",
			Level:   "error",
		})
	} else if config.BackupRetentionDays > 365 {
		result.Errors = append(result.Errors, ConfigValidationError{
			Field:   "backup_retention_days",
			Value:   fmt.Sprintf("%d", config.BackupRetentionDays),
			Message: "备份保留时间过长，可能占用大量磁盘空间",
			Level:   "warning",
		})
	}
}

// validatePaths 验证路径设置
func (scm *SmartConfigManager) validatePaths(config *Config, result *ConfigValidationResult) {
	// 验证Linux回收站路径
	if config.Linux.TrashDir != "" {
		if strings.HasPrefix(config.Linux.TrashDir, "~") {
			// 展开用户目录
			homeDir, err := os.UserHomeDir()
			if err != nil {
				result.Errors = append(result.Errors, ConfigValidationError{
					Field:   "linux.trash_dir",
					Value:   config.Linux.TrashDir,
					Message: "无法展开用户目录路径",
					Level:   "warning",
				})
			} else {
				expandedPath := strings.Replace(config.Linux.TrashDir, "~", homeDir, 1)
				if _, err := os.Stat(filepath.Dir(expandedPath)); err != nil {
					result.Errors = append(result.Errors, ConfigValidationError{
						Field:   "linux.trash_dir",
						Value:   config.Linux.TrashDir,
						Message: "回收站目录的父目录不存在",
						Level:   "warning",
					})
				}
			}
		}
	}

	// 验证macOS废纸篓路径
	if config.Darwin.TrashDir != "" {
		if strings.HasPrefix(config.Darwin.TrashDir, "~") {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				result.Errors = append(result.Errors, ConfigValidationError{
					Field:   "darwin.trash_dir",
					Value:   config.Darwin.TrashDir,
					Message: "无法展开用户目录路径",
					Level:   "warning",
				})
			} else {
				expandedPath := strings.Replace(config.Darwin.TrashDir, "~", homeDir, 1)
				if _, err := os.Stat(filepath.Dir(expandedPath)); err != nil {
					result.Errors = append(result.Errors, ConfigValidationError{
						Field:   "darwin.trash_dir",
						Value:   config.Darwin.TrashDir,
						Message: "废纸篓目录的父目录不存在",
						Level:   "warning",
					})
				}
			}
		}
	}
}

// validatePlatformSettings 验证平台特定设置
func (scm *SmartConfigManager) validatePlatformSettings(config *Config, result *ConfigValidationResult) {
	// 这里可以添加更多平台特定的验证逻辑
	// 例如检查Windows UAC设置、Linux SELinux设置等
}

// autoFixConfig 自动修复配置
func (scm *SmartConfigManager) autoFixConfig(config *Config, errors []ConfigValidationError) (*Config, bool) {
	fixedConfig := *config // 复制配置
	fixed := false

	for _, err := range errors {
		if err.Level == "error" || err.Level == "warning" {
			switch err.Field {
			case "language":
				fixedConfig.Language = "zh-cn"
				fixed = true
			case "interactive_mode":
				fixedConfig.InteractiveMode = "confirm"
				fixed = true
			case "log_level":
				fixedConfig.LogLevel = "info"
				fixed = true
			case "safe_mode":
				fixedConfig.SafeMode = "normal"
				fixed = true
			case "max_file_size":
				if fixedConfig.MaxFileSize < 0 {
					fixedConfig.MaxFileSize = 100 * 1024 * 1024 // 100MB
					fixed = true
				}
			case "max_concurrent_ops":
				if fixedConfig.MaxConcurrentOps <= 0 {
					fixedConfig.MaxConcurrentOps = 10
					fixed = true
				}
			case "backup_retention_days":
				if fixedConfig.BackupRetentionDays < 0 {
					fixedConfig.BackupRetentionDays = 30
					fixed = true
				}
			}
		}
	}

	return &fixedConfig, fixed
}

// backupConfig 备份配置文件
func (scm *SmartConfigManager) backupConfig(suffix string) error {
	// 创建备份目录
	if err := os.MkdirAll(scm.backupDir, 0755); err != nil {
		return fmt.Errorf("创建备份目录失败: %v", err)
	}

	// 生成备份文件名
	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("delguard-%s-%s.json", timestamp, suffix)
	backupPath := filepath.Join(scm.backupDir, backupName)

	// 读取当前配置文件
	data, err := os.ReadFile(scm.configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 写入备份文件
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("写入备份文件失败: %v", err)
	}

	fmt.Printf("📦 配置已备份到: %s\n", backupPath)
	return nil
}

// loadFromBackup 从备份恢复配置
func (scm *SmartConfigManager) loadFromBackup() (*Config, error) {
	// 列出备份文件
	entries, err := os.ReadDir(scm.backupDir)
	if err != nil {
		return nil, fmt.Errorf("读取备份目录失败: %v", err)
	}

	// 找到最新的备份文件
	var latestBackup string
	var latestTime time.Time

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().After(latestTime) {
			latestTime = info.ModTime()
			latestBackup = entry.Name()
		}
	}

	if latestBackup == "" {
		return nil, fmt.Errorf("未找到可用的备份文件")
	}

	// 加载备份配置
	backupPath := filepath.Join(scm.backupDir, latestBackup)
	fmt.Printf("🔄 从备份恢复: %s\n", latestBackup)

	return scm.loadConfigFile(backupPath)
}

// generateDefaultConfig 生成默认配置
func (scm *SmartConfigManager) generateDefaultConfig() *Config {
	return &Config{
		Version:                   "1.0.0",
		SchemaVersion:             "1.0",
		Language:                  "zh-cn",
		InteractiveMode:           "confirm",
		LogLevel:                  "info",
		SafeMode:                  "normal",
		UseRecycleBin:             true,
		MaxFileSize:               100 * 1024 * 1024, // 100MB
		MaxConcurrentOps:          10,
		BackupRetentionDays:       30,
		LogRetentionDays:          30,
		EnableSecurityChecks:      true,
		EnablePathValidation:      true,
		EnableHiddenCheck:         true,
		EnableOverwriteProtection: true,
		OutputPrefixEnabled:       true,
		OutputPrefix:              "DelGuard: ",
		MaxBackupFiles:            DefaultMaxBackupFiles,
		TrashMaxSize:              DefaultTrashMaxSize,
		MaxPathLength:             DefaultMaxPathLength,
		SimilarityThreshold:       0.8,
		LogMaxSize:                10,
		LogMaxBackups:             7,
		LogRotateDaily:            false,
		LogCompress:               true,
	}
}

// saveConfig 保存配置文件
func (scm *SmartConfigManager) saveConfig(config *Config) error {
	// 创建配置目录
	configDir := filepath.Dir(scm.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 序列化配置
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(scm.configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	scm.updateLastModified()
	return nil
}

// ReloadConfig 重新加载配置
func (scm *SmartConfigManager) ReloadConfig() (*Config, error) {
	fmt.Println("🔄 重新加载配置...")
	return scm.LoadConfigWithFallback()
}

// WatchConfig 监控配置文件变化
func (scm *SmartConfigManager) WatchConfig() error {
	// 简化实现：定期检查文件修改时间
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if scm.isConfigModified() {
				fmt.Println("📝 检测到配置文件变化，重新加载...")
				if newConfig, err := scm.LoadConfigWithFallback(); err == nil {
					scm.notifyWatchers(scm.config, newConfig)
					scm.config = newConfig
				}
			}
		}
	}()

	return nil
}

// AddWatcher 添加配置监听器
func (scm *SmartConfigManager) AddWatcher(watcher ConfigWatcher) {
	scm.watchers = append(scm.watchers, watcher)
}

// notifyWatchers 通知所有监听器
func (scm *SmartConfigManager) notifyWatchers(oldConfig, newConfig *Config) {
	for _, watcher := range scm.watchers {
		if err := watcher.OnConfigChanged(oldConfig, newConfig); err != nil {
			fmt.Printf("⚠️  配置监听器通知失败: %v\n", err)
		}
	}
}

// isConfigModified 检查配置文件是否被修改
func (scm *SmartConfigManager) isConfigModified() bool {
	info, err := os.Stat(scm.configPath)
	if err != nil {
		return false
	}

	return info.ModTime().After(scm.lastModified)
}

// updateLastModified 更新最后修改时间
func (scm *SmartConfigManager) updateLastModified() {
	if info, err := os.Stat(scm.configPath); err == nil {
		scm.lastModified = info.ModTime()
	}
}

// removeComments 移除JSONC注释
func (scm *SmartConfigManager) removeComments(data []byte) []byte {
	lines := strings.Split(string(data), "\n")
	var result []string

	inBlockComment := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 处理块注释
		if strings.Contains(trimmed, "/*") {
			inBlockComment = true
		}
		if strings.Contains(trimmed, "*/") {
			inBlockComment = false
			continue
		}
		if inBlockComment {
			continue
		}

		// 处理行注释
		if strings.HasPrefix(trimmed, "//") {
			continue
		}

		// 移除行尾注释
		if idx := strings.Index(line, "//"); idx != -1 {
			line = line[:idx]
		}

		result = append(result, line)
	}

	return []byte(strings.Join(result, "\n"))
}

// contains 检查切片是否包含指定元素
func (scm *SmartConfigManager) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetConfig 获取当前配置
func (scm *SmartConfigManager) GetConfig() *Config {
	return scm.config
}

// ShowValidationReport 显示配置验证报告
func (scm *SmartConfigManager) ShowValidationReport(result ConfigValidationResult) {
	fmt.Println("📋 配置验证报告")
	fmt.Println("=" + strings.Repeat("=", 50))

	if result.Valid {
		fmt.Println("✅ 配置文件验证通过")
	} else {
		fmt.Println("❌ 配置文件存在问题")
	}

	if len(result.Errors) > 0 {
		fmt.Println("\n问题详情:")
		for _, err := range result.Errors {
			switch err.Level {
			case "error":
				fmt.Printf("❌ [错误] %s: %s (值: %s)\n", err.Field, err.Message, err.Value)
			case "warning":
				fmt.Printf("⚠️  [警告] %s: %s (值: %s)\n", err.Field, err.Message, err.Value)
			case "info":
				fmt.Printf("ℹ️  [信息] %s: %s (值: %s)\n", err.Field, err.Message, err.Value)
			}
		}
	}

	if len(result.Fixed) > 0 {
		fmt.Println("\n🔧 已自动修复:")
		for _, fix := range result.Fixed {
			fmt.Printf("  • %s\n", fix)
		}
	}

	if len(result.Backups) > 0 {
		fmt.Println("\n📦 相关备份:")
		for _, backup := range result.Backups {
			fmt.Printf("  • %s\n", backup)
		}
	}
}
