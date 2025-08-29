package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config 配置结构
type Config struct {
	// 删除配置
	Delete DeleteConfig `json:"delete"`

	// 搜索配置
	Search SearchConfig `json:"search"`

	// 恢复配置
	Restore RestoreConfig `json:"restore"`

	// 安全配置
	Security SecurityConfig `json:"security"`

	// 日志配置
	Logging LoggingConfig `json:"logging"`
}

// DeleteConfig 删除配置
type DeleteConfig struct {
	SafeMode     bool   `json:"safeMode"`     // 安全模式
	UseTrash     bool   `json:"useTrash"`     // 使用回收站
	BackupBefore bool   `json:"backupBefore"` // 删除前备份
	MaxFileSize  int64  `json:"maxFileSize"`  // 最大文件大小限制
	TrashPath    string `json:"trashPath"`    // 自定义回收站路径
}

// SearchConfig 搜索配置
type SearchConfig struct {
	MaxResults    int    `json:"maxResults"`    // 最大结果数
	CaseSensitive bool   `json:"caseSensitive"` // 大小写敏感
	UseRegex      bool   `json:"useRegex"`      // 使用正则表达式
	IndexEnabled  bool   `json:"indexEnabled"`  // 启用索引
	IndexPath     string `json:"indexPath"`     // 索引路径
}

// RestoreConfig 恢复配置
type RestoreConfig struct {
	VerifyIntegrity   bool `json:"verifyIntegrity"`   // 验证完整性
	CreateBackup      bool `json:"createBackup"`      // 创建备份
	OverwriteExisting bool `json:"overwriteExisting"` // 覆盖现有文件
	MaxConcurrency    int  `json:"maxConcurrency"`    // 最大并发数
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	RequireConfirmation bool     `json:"requireConfirmation"` // 需要确认
	ProtectedPaths      []string `json:"protectedPaths"`      // 受保护路径
	AllowedExtensions   []string `json:"allowedExtensions"`   // 允许的扩展名
	BlockedExtensions   []string `json:"blockedExtensions"`   // 阻止的扩展名
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level    string `json:"level"`    // 日志级别
	FilePath string `json:"filePath"` // 日志文件路径
	MaxSize  int    `json:"maxSize"`  // 最大文件大小(MB)
	MaxAge   int    `json:"maxAge"`   // 最大保存天数
}

// NewConfig 创建默认配置
func NewConfig() *Config {
	return &Config{
		Delete: DeleteConfig{
			SafeMode:     true,
			UseTrash:     true,
			BackupBefore: false,
			MaxFileSize:  1024 * 1024 * 1024, // 1GB
			TrashPath:    "",
		},
		Search: SearchConfig{
			MaxResults:    1000,
			CaseSensitive: false,
			UseRegex:      false,
			IndexEnabled:  true,
			IndexPath:     ".delguard/index",
		},
		Restore: RestoreConfig{
			VerifyIntegrity:   true,
			CreateBackup:      true,
			OverwriteExisting: false,
			MaxConcurrency:    4,
		},
		Security: SecurityConfig{
			RequireConfirmation: true,
			ProtectedPaths: []string{
				"/",
				"/bin",
				"/sbin",
				"/usr",
				"/etc",
				"C:\\Windows",
				"C:\\Program Files",
			},
			AllowedExtensions: []string{},
			BlockedExtensions: []string{
				".exe",
				".bat",
				".cmd",
				".ps1",
				".sh",
			},
		},
		Logging: LoggingConfig{
			Level:    "info",
			FilePath: ".delguard/logs/delguard.log",
			MaxSize:  10,
			MaxAge:   30,
		},
	}
}

// Load 加载配置
func Load() (*Config, error) {
	cfg := NewConfig()

	// 尝试从默认位置加载配置文件
	configPath := getDefaultConfigPath()
	if _, err := os.Stat(configPath); err == nil {
		if err := cfg.LoadFromFile(configPath); err != nil {
			return nil, fmt.Errorf("加载配置文件失败: %v", err)
		}
	}

	return cfg, nil
}

// LoadFromFile 从文件加载配置
func (c *Config) LoadFromFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, c)
}

// SaveToFile 保存配置到文件
func (c *Config) SaveToFile(filePath string) error {
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// Save 保存配置到默认位置
func (c *Config) Save() error {
	return c.SaveToFile(getDefaultConfigPath())
}

// Show 显示当前配置
func (c *Config) Show() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println("当前配置:")
	fmt.Println(string(data))
	return nil
}

// Set 设置配置项
func (c *Config) Set(key, value string) error {
	// 这里可以实现更复杂的配置设置逻辑
	fmt.Printf("设置配置项 %s = %s\n", key, value)
	return c.Save()
}

// Reset 重置配置为默认值
func (c *Config) Reset() error {
	*c = *NewConfig()
	return c.Save()
}

// getDefaultConfigPath 获取默认配置文件路径
func getDefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".delguard/config.json"
	}
	return filepath.Join(home, ".delguard", "config.json")
}
