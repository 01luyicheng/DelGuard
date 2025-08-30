package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

// Config 全局配置结构
type Config struct {
	Trash       TrashConfig       `yaml:"trash" mapstructure:"trash"`
	Logging     LoggingConfig     `yaml:"logging" mapstructure:"logging"`
	UI          UIConfig          `yaml:"ui" mapstructure:"ui"`
	Install     InstallConfig     `yaml:"install" mapstructure:"install"`
	Security    SecurityConfig    `yaml:"security" mapstructure:"security"`
	Performance PerformanceConfig `yaml:"performance" mapstructure:"performance"`
}

// TrashConfig 回收站配置
type TrashConfig struct {
	AutoClean     bool   `yaml:"auto_clean" mapstructure:"auto_clean"`
	MaxDays       int    `yaml:"max_days" mapstructure:"max_days"`
	ConfirmDelete bool   `yaml:"confirm_delete" mapstructure:"confirm_delete"`
	MaxSize       string `yaml:"max_size" mapstructure:"max_size"`
	UseSystemTrash bool   `yaml:"use_system_trash" mapstructure:"use_system_trash"`
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
	Language    string `yaml:"language" mapstructure:"language"`
	Color       bool   `yaml:"color" mapstructure:"color"`
	Unicode     bool   `yaml:"unicode" mapstructure:"unicode"`
	ProgressBar bool   `yaml:"progress_bar" mapstructure:"progress_bar"`
}

// InstallConfig 安装配置
type InstallConfig struct {
	SystemWide    bool   `yaml:"system_wide" mapstructure:"system_wide"`
	InstallDir    string `yaml:"install_dir" mapstructure:"install_dir"`
	CreateAlias   bool   `yaml:"create_alias" mapstructure:"create_alias"`
	BackupOriginal bool   `yaml:"backup_original" mapstructure:"backup_original"`
}

// SecurityConfig 安全设置
type SecurityConfig struct {
	StrictMode       bool     `yaml:"strict_mode" mapstructure:"strict_mode"`
	MaxPathLength    int      `yaml:"max_path_length" mapstructure:"max_path_length"`
	AllowedExtensions []string `yaml:"allowed_extensions" mapstructure:"allowed_extensions"`
	BlockedExtensions []string `yaml:"blocked_extensions" mapstructure:"blocked_extensions"`
}

// PerformanceConfig 性能设置
type PerformanceConfig struct {
	BatchSize     int `yaml:"batch_size" mapstructure:"batch_size"`
	BufferSize    int `yaml:"buffer_size" mapstructure:"buffer_size"`
	MaxConcurrent int `yaml:"max_concurrent" mapstructure:"max_concurrent"`
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
	viper.SetDefault("trash.use_system_trash", true)

	// 日志配置默认值
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.file", getDefaultLogPath())
	viper.SetDefault("logging.max_size", 10)
	viper.SetDefault("logging.max_age", 7)
	viper.SetDefault("logging.compress", true)

	// UI配置默认值
	viper.SetDefault("ui.language", "zh-CN")
	viper.SetDefault("ui.color", true)
	viper.SetDefault("ui.unicode", true)
	viper.SetDefault("ui.progress_bar", true)

	// 安装配置默认值
	viper.SetDefault("install.system_wide", true)
	viper.SetDefault("install.install_dir", getDefaultInstallDir())
	viper.SetDefault("install.create_alias", true)
	viper.SetDefault("install.backup_original", true)

	// 安全设置默认值
	viper.SetDefault("security.strict_mode", false)
	viper.SetDefault("security.max_path_length", 4096)
	viper.SetDefault("security.allowed_extensions", []string{"*"})
	viper.SetDefault("security.blocked_extensions", []string{".sys", ".dll", ".exe", ".msi"})

	// 性能设置默认值
	viper.SetDefault("performance.batch_size", 10)
	viper.SetDefault("performance.buffer_size", 8192)
	viper.SetDefault("performance.max_concurrent", 5)
	
	// 其他全局配置
	viper.SetDefault("verbose", false)
	viper.SetDefault("force", false)
	viper.SetDefault("quiet", false)
}

// createDefaultConfig 创建默认配置文件
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

	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("创建日志目录失败: %v", err)
		// 回退到临时目录
		logDir = os.TempDir()
	}
	return filepath.Join(logDir, "delguard.log")
}

// GetDefaultLogPath 获取默认日志路径的公共函数
func GetDefaultLogPath() string {
	return getDefaultLogPath()
}
