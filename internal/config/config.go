package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// Config 全局配置结构
type Config struct {
	Trash   TrashConfig   `yaml:"trash"`
	Logging LoggingConfig `yaml:"logging"`
	UI      UIConfig      `yaml:"ui"`
}

// TrashConfig 回收站配置
type TrashConfig struct {
	AutoClean bool `yaml:"auto_clean"`
	MaxDays   int  `yaml:"max_days"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level    string `yaml:"level"`
	File     string `yaml:"file"`
	MaxSize  int    `yaml:"max_size"`
	MaxAge   int    `yaml:"max_age"`
	Compress bool   `yaml:"compress"`
}

// UIConfig 界面配置
type UIConfig struct {
	Language string `yaml:"language"`
	Color    bool   `yaml:"color"`
}

// GlobalConfig 全局配置实例
var GlobalConfig *Config

// Init 初始化配置
func Init() error {
	GlobalConfig = &Config{
		Trash: TrashConfig{
			AutoClean: false,
			MaxDays:   30,
		},
		Logging: LoggingConfig{
			Level:    "info",
			File:     getDefaultLogPath(),
			MaxSize:  10,
			MaxAge:   7,
			Compress: true,
		},
		UI: UIConfig{
			Language: "zh",
			Color:    true,
		},
	}
	return nil
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
