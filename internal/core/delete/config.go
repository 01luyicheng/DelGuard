package delete

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	configPath string
}

// NewConfigManager 创建配置管理器
func NewConfigManager(configPath string) *ConfigManager {
	return &ConfigManager{
		configPath: configPath,
	}
}

// LoadConfig 加载配置
func (cm *ConfigManager) LoadConfig() (*Config, error) {
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		// 配置文件不存在，返回默认配置
		return DefaultConfig, nil
	}

	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 验证配置
	if err := cm.validateConfig(&config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %v", err)
	}

	return &config, nil
}

// SaveConfig 保存配置
func (cm *ConfigManager) SaveConfig(config *Config) error {
	if err := cm.validateConfig(config); err != nil {
		return fmt.Errorf("配置验证失败: %v", err)
	}

	// 确保配置目录存在
	if err := os.MkdirAll(filepath.Dir(cm.configPath), 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	if err := os.WriteFile(cm.configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

// validateConfig 验证配置
func (cm *ConfigManager) validateConfig(config *Config) error {
	if config.MaxConcurrency <= 0 {
		return fmt.Errorf("最大并发数必须大于0")
	}

	if config.MaxConcurrency > 100 {
		return fmt.Errorf("最大并发数不能超过100")
	}

	// 验证受保护路径
	for _, path := range config.ProtectedPaths {
		if path == "" {
			return fmt.Errorf("受保护路径不能为空")
		}
	}

	return nil
}

// GetDefaultConfigPath 获取默认配置文件路径
func GetDefaultConfigPath() string {
	var configDir string

	switch runtime.GOOS {
	case "windows":
		configDir = os.Getenv("APPDATA")
		if configDir == "" {
			configDir = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
	case "darwin":
		configDir = filepath.Join(os.Getenv("HOME"), "Library", "Application Support")
	default:
		configDir = os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			configDir = filepath.Join(os.Getenv("HOME"), ".config")
		}
	}

	return filepath.Join(configDir, "delguard", "config.json")
}

// CreateDefaultConfig 创建默认配置文件
func CreateDefaultConfig() error {
	configPath := GetDefaultConfigPath()
	cm := NewConfigManager(configPath)

	// 检查配置文件是否已存在
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("配置文件已存在: %s", configPath)
	}

	return cm.SaveConfig(DefaultConfig)
}
