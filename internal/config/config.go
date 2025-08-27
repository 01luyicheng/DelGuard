package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config 配置结构
type Config struct {
	Language         string `json:"language"`
	MaxFileSize      int64  `json:"max_file_size"`
	MaxBackupFiles   int    `json:"max_backup_files"`
	EnableRecycleBin bool   `json:"enable_recycle_bin"`
	EnableLogging    bool   `json:"enable_logging"`
	LogLevel         string `json:"log_level"`
	ConfigPath       string `json:"-"`
}

// NewConfig 创建新的配置实例
func NewConfig() *Config {
	return &Config{
		Language:         "zh-cn",
		MaxFileSize:      1024 * 1024 * 1024, // 1GB
		MaxBackupFiles:   10,
		EnableRecycleBin: true,
		EnableLogging:    true,
		LogLevel:         "info",
	}
}

// Load 加载配置
func Load() (*Config, error) {
	cfg := NewConfig()

	// 获取配置文件路径
	configPath, err := getConfigPath()
	if err != nil {
		return cfg, nil // 使用默认配置
	}

	cfg.ConfigPath = configPath

	// 如果配置文件存在，加载它
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return cfg, nil // 使用默认配置
		}

		if err := json.Unmarshal(data, cfg); err != nil {
			return cfg, nil // 使用默认配置
		}
	}

	return cfg, nil
}

// LoadFromFile 从指定文件加载配置
func (c *Config) LoadFromFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	if err := json.Unmarshal(data, c); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	c.ConfigPath = filePath
	return nil
}

// SaveToFile 保存配置到指定文件
func (c *Config) SaveToFile(filePath string) error {
	// 确保配置目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	return os.WriteFile(filePath, data, 0644)
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Language == "" {
		return fmt.Errorf("语言设置不能为空")
	}

	if c.MaxFileSize < 0 {
		return fmt.Errorf("最大文件大小不能为负数")
	}

	if c.MaxBackupFiles < 0 {
		return fmt.Errorf("最大备份文件数不能为负数")
	}

	validLogLevels := []string{"debug", "info", "warn", "error"}
	validLevel := false
	for _, level := range validLogLevels {
		if c.LogLevel == level {
			validLevel = true
			break
		}
	}
	if !validLevel {
		return fmt.Errorf("无效的日志级别: %s", c.LogLevel)
	}

	return nil
}

// Save 保存配置
func (c *Config) Save() error {
	if c.ConfigPath == "" {
		return fmt.Errorf("配置路径未设置")
	}

	// 确保配置目录存在
	dir := filepath.Dir(c.ConfigPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	return os.WriteFile(c.ConfigPath, data, 0644)
}

// Show 显示配置
func (c *Config) Show() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

// Set 设置配置项
func (c *Config) Set(key, value string) error {
	switch key {
	case "language":
		c.Language = value
	case "log_level":
		c.LogLevel = value
	default:
		return fmt.Errorf("未知配置项: %s", key)
	}
	return c.Save()
}

// Reset 重置配置
func (c *Config) Reset() error {
	*c = Config{
		Language:         "zh-cn",
		MaxFileSize:      1024 * 1024 * 1024,
		MaxBackupFiles:   10,
		EnableRecycleBin: true,
		EnableLogging:    true,
		LogLevel:         "info",
		ConfigPath:       c.ConfigPath,
	}
	return c.Save()
}

// getConfigPath 获取配置文件路径
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".delguard", "config.json"), nil
}
