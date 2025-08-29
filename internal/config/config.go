package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

// Config 全局配置结构
type Config struct {
	Trash   TrashConfig   `yaml:"trash" mapstructure:"trash"`
	Logging LoggingConfig `yaml:"logging" mapstructure:"logging"`
	UI      UIConfig      `yaml:"ui" mapstructure:"ui"`
	Install InstallConfig `yaml:"install" mapstructure:"install"`
}

// TrashConfig 回收站配置
type TrashConfig struct {
	AutoClean     bool   `yaml:"auto_clean" mapstructure:"auto_clean"`
	MaxDays       int    `yaml:"max_days" mapstructure:"max_days"`
	ConfirmDelete bool   `yaml:"confirm_delete" mapstructure:"confirm_delete"`
	MaxSize       string `yaml:"max_size" mapstructure:"max_size"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level    string `yaml:"level" mapstructure:"level"`
	File     string `yaml:"file" mapstructure:"file"`
	MaxSize  int    `yaml:"max_size" mapstructure:"max_size"`
	MaxAge   int    `yaml:"max_age" mapstructure:"max_age"`
	Compress bool   `yaml:"compress" mapstructure:"compress"`
}

// UIConfig 界面配置
type UIConfig struct {
	Language string `yaml:"language" mapstructure:"language"`
	Color    bool   `yaml:"color" mapstructure:"color"`
	Unicode  bool   `yaml:"unicode" mapstructure:"unicode"`
}

// InstallConfig 安装配置
type InstallConfig struct {
	SystemWide  bool   `yaml:"system_wide" mapstructure:"system_wide"`
	InstallDir  string `yaml:"install_dir" mapstructure:"install_dir"`
	CreateAlias bool   `yaml:"create_alias" mapstructure:"create_alias"`
}

// GlobalConfig 全局配置实例
var GlobalConfig *Config

// Init 初始化配置
func Init() error {
	// 设置配置文件名和路径
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// 添加配置文件搜索路径
	configDir := getConfigDir()
	viper.AddConfigPath(configDir)
	viper.AddConfigPath(".")

	// 设置默认值
	setDefaults()

	// 尝试读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件不存在，创建默认配置
			if err := createDefaultConfig(configDir); err != nil {
				return fmt.Errorf("创建默认配置失败: %v", err)
			}
		} else {
			return fmt.Errorf("读取配置文件失败: %v", err)
		}
	}

	// 解析配置到结构体
	GlobalConfig = &Config{}
	if err := viper.Unmarshal(GlobalConfig); err != nil {
		return fmt.Errorf("解析配置失败: %v", err)
	}

	return nil
}

// setDefaults 设置默认配置值
func setDefaults() {
	// 回收站配置默认值
	viper.SetDefault("trash.auto_clean", false)
	viper.SetDefault("trash.max_days", 30)
	viper.SetDefault("trash.confirm_delete", true)
	viper.SetDefault("trash.max_size", "1GB")

	// 日志配置默认值
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.file", getDefaultLogPath())
	viper.SetDefault("logging.max_size", 10)
	viper.SetDefault("logging.max_age", 7)
	viper.SetDefault("logging.compress", true)

	// UI配置默认值
	viper.SetDefault("ui.language", "zh")
	viper.SetDefault("ui.color", true)
	viper.SetDefault("ui.unicode", true)

	// 安装配置默认值
	viper.SetDefault("install.system_wide", true)
	viper.SetDefault("install.install_dir", getDefaultInstallDir())
	viper.SetDefault("install.create_alias", true)
}

// createDefaultConfig 创建默认配置文件
func createDefaultConfig(configDir string) error {
	// 确保配置目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configFile := filepath.Join(configDir, "config.yaml")
	return viper.WriteConfigAs(configFile)
}

// getConfigDir 获取配置目录
func getConfigDir() string {
	var configDir string
	switch runtime.GOOS {
	case "windows":
		configDir = filepath.Join(os.Getenv("APPDATA"), "DelGuard")
	case "darwin":
		configDir = filepath.Join(os.Getenv("HOME"), ".config", "delguard")
	default:
		configDir = filepath.Join(os.Getenv("HOME"), ".config", "delguard")
	}
	return configDir
}

// getDefaultInstallDir 获取默认安装目录
func getDefaultInstallDir() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("ProgramFiles"), "DelGuard")
	default:
		return "/usr/local/bin"
	}
}

// getDefaultLogPath 获取默认日志路径
func getDefaultLogPath() string {
	var logDir string
	switch runtime.GOOS {
	case "windows":
		logDir = filepath.Join(os.Getenv("APPDATA"), "DelGuard", "logs")
	case "darwin":
		logDir = filepath.Join(os.Getenv("HOME"), "Library", "Logs", "DelGuard")
	default:
		logDir = filepath.Join(os.Getenv("HOME"), ".local", "share", "delguard", "logs")
	}

	os.MkdirAll(logDir, 0755)
	return filepath.Join(logDir, "delguard.log")
}
