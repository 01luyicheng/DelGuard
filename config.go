package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Config 配置结构体
type Config struct {
	// 基本设置
	UseRecycleBin   bool   `json:"use_recycle_bin"`
	InteractiveMode string `json:"interactive_mode"` // "always", "never", "confirm"
	Language        string `json:"language"`
	LogLevel        string `json:"log_level"`
	SafeMode        string `json:"safe_mode"` // "strict", "normal", "relaxed"

	// 限制设置
	MaxBackupFiles   int64 `json:"max_backup_files"`
	TrashMaxSize     int64 `json:"trash_max_size"` // MB
	MaxFileSize      int64 `json:"max_file_size"`  // bytes
	MaxPathLength    int   `json:"max_path_length"`
	MaxConcurrentOps int   `json:"max_concurrent_ops"`

	// 安全设置
	EnableSecurityChecks bool `json:"enable_security_checks"`
	EnableMalwareScan    bool `json:"enable_malware_scan"`
	EnablePathValidation bool `json:"enable_path_validation"`
	EnableHiddenCheck    bool `json:"enable_hidden_check"`

	// 高级设置
	BackupRetentionDays int    `json:"backup_retention_days"`
	LogRetentionDays    int    `json:"log_retention_days"`
	EnableTelemetry     bool   `json:"enable_telemetry"`
	TelemetryEndpoint   string `json:"telemetry_endpoint"`

	// 平台特定设置
	Windows WindowsConfig `json:"windows"`
	Linux   LinuxConfig   `json:"linux"`
	Darwin  DarwinConfig  `json:"darwin"`

	// 运行时状态（不保存到配置文件）
	ConfigPath string    `json:"-"`
	LastLoaded time.Time `json:"-"`
}

// WindowsConfig Windows平台特定配置
type WindowsConfig struct {
	RecycleBinPath     string `json:"recycle_bin_path"`
	UseSystemTrash     bool   `json:"use_system_trash"`
	EnableUACPrompt    bool   `json:"enable_uac_prompt"`
	CheckFileOwnership bool   `json:"check_file_ownership"`
}

// LinuxConfig Linux平台特定配置
type LinuxConfig struct {
	TrashDir      string `json:"trash_dir"`
	UseXDGTrash   bool   `json:"use_xdg_trash"`
	CheckSELinux  bool   `json:"check_selinux"`
	CheckAppArmor bool   `json:"check_apparmor"`
}

// DarwinConfig macOS平台特定配置
type DarwinConfig struct {
	TrashDir        string `json:"trash_dir"`
	UseSystemTrash  bool   `json:"use_system_trash"`
	CheckFileVault  bool   `json:"check_filevault"`
	CheckGatekeeper bool   `json:"check_gatekeeper"`
}

var defaultConfig *Config

// LoadConfig 加载配置
func LoadConfig() (*Config, error) {
	if defaultConfig != nil {
		return defaultConfig, nil
	}

	config := &Config{}

	// 加载默认配置
	config.setDefaults()

	// 查找配置文件路径
	configPaths := config.findConfigPaths()

	// 尝试加载配置文件
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			if err := config.loadFromFile(path); err != nil {
				// 记录错误但不停止加载
				LogWarn("config", path, fmt.Sprintf("配置文件加载失败: %v", err))
				continue
			}
			config.ConfigPath = path
			break
		}
	}

	// 从环境变量加载配置
	config.loadFromEnv()

	defaultConfig = config
	return config, nil
}

// setDefaults 设置默认配置值
func (c *Config) setDefaults() {
	c.UseRecycleBin = true
	c.InteractiveMode = "confirm"
	c.Language = "auto"
	c.LogLevel = "info"
	c.SafeMode = "normal"

	c.MaxBackupFiles = 1000
	c.TrashMaxSize = 10000      // 10GB
	c.MaxFileSize = 10737418240 // 10GB
	c.MaxPathLength = 4096
	c.MaxConcurrentOps = 100

	c.EnableSecurityChecks = true
	c.EnableMalwareScan = false // 默认关闭，提高性能
	c.EnablePathValidation = true
	c.EnableHiddenCheck = true

	c.BackupRetentionDays = 30
	c.LogRetentionDays = 7
	c.EnableTelemetry = false
	c.TelemetryEndpoint = "https://telemetry.delguard.io/v1/report"

	// 平台特定默认值
	switch runtime.GOOS {
	case "windows":
		c.Windows.RecycleBinPath = ""
		c.Windows.UseSystemTrash = true
		c.Windows.EnableUACPrompt = true
		c.Windows.CheckFileOwnership = true
	case "linux":
		c.Linux.TrashDir = ""
		c.Linux.UseXDGTrash = true
		c.Linux.CheckSELinux = false
		c.Linux.CheckAppArmor = false
	case "darwin":
		c.Darwin.TrashDir = ""
		c.Darwin.UseSystemTrash = true
		c.Darwin.CheckFileVault = false
		c.Darwin.CheckGatekeeper = false
	}
}

// findConfigPaths 查找配置文件路径
func (c *Config) findConfigPaths() []string {
	var paths []string

	// 用户配置目录
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".delguard", "config.json"))
	}

	// 系统配置目录
	switch runtime.GOOS {
	case "windows":
		if systemRoot := os.Getenv("SystemRoot"); systemRoot != "" {
			paths = append(paths, filepath.Join(systemRoot, "delguard", "config.json"))
		}
	default:
		paths = append(paths, "/etc/delguard/config.json")
	}

	// 当前目录
	paths = append(paths, "config.json")

	return paths
}

// loadFromFile 从文件加载配置
func (c *Config) loadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, c)
}

// loadFromEnv 从环境变量加载配置
func (c *Config) loadFromEnv() {
	// 基本设置
	if v := os.Getenv("DELGUARD_USE_RECYCLE_BIN"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			c.UseRecycleBin = b
		}
	}

	if v := os.Getenv("DELGUARD_INTERACTIVE_MODE"); v != "" {
		c.InteractiveMode = v
	}

	if v := os.Getenv("DELGUARD_LANGUAGE"); v != "" {
		c.Language = v
	}

	if v := os.Getenv("DELGUARD_LOG_LEVEL"); v != "" {
		c.LogLevel = v
	}

	if v := os.Getenv("DELGUARD_SAFE_MODE"); v != "" {
		c.SafeMode = v
	}

	// 限制设置
	if v := os.Getenv("DELGUARD_MAX_BACKUP_FILES"); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			c.MaxBackupFiles = i
		}
	}

	if v := os.Getenv("DELGUARD_TRASH_MAX_SIZE"); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			c.TrashMaxSize = i
		}
	}

	if v := os.Getenv("DELGUARD_MAX_FILE_SIZE"); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			c.MaxFileSize = i
		}
	}

	// 安全设置
	if v := os.Getenv("DELGUARD_ENABLE_SECURITY_CHECKS"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			c.EnableSecurityChecks = b
		}
	}

	if v := os.Getenv("DELGUARD_ENABLE_MALWARE_SCAN"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			c.EnableMalwareScan = b
		}
	}
}

// GetInteractiveDefault 获取交互模式默认值
func (c *Config) GetInteractiveDefault() bool {
	return c.InteractiveMode == "always" || c.InteractiveMode == "confirm"
}

// GetUseTrash 获取是否使用回收站
func (c *Config) GetUseTrash() bool {
	return c.UseRecycleBin
}

// GetMaxFileSize 获取最大文件大小限制
func (c *Config) GetMaxFileSize() int64 {
	return c.MaxFileSize
}

// GetSafeMode 获取安全模式
func (c *Config) GetSafeMode() string {
	return c.SafeMode
}

// Validate 验证配置有效性
func (c *Config) Validate() error {
	var errs []error

	// 验证日志级别
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}
	if !validLevels[c.LogLevel] {
		err := fmt.Errorf("无效的日志级别: %s (有效值: debug, info, warn, error, fatal)", c.LogLevel)
		errs = append(errs, err)
	}

	// 验证交互模式
	switch c.InteractiveMode {
	case "always", "never", "confirm":
	default:
		err := fmt.Errorf("无效的交互模式: %s (有效值: always, never, confirm)", c.InteractiveMode)
		errs = append(errs, err)
	}

	// 验证安全模式
	switch c.SafeMode {
	case "strict", "normal", "relaxed":
	default:
		err := fmt.Errorf("无效的安全模式: %s (有效值: strict, normal, relaxed)", c.SafeMode)
		errs = append(errs, err)
	}

	// 验证数值范围
	if c.MaxFileSize < 0 {
		err := fmt.Errorf("最大文件大小不能为负数: %d", c.MaxFileSize)
		errs = append(errs, err)
	}

	if c.MaxPathLength <= 0 {
		err := fmt.Errorf("最大路径长度必须为正数: %d", c.MaxPathLength)
		errs = append(errs, err)
	}

	if c.MaxConcurrentOps <= 0 {
		err := fmt.Errorf("最大并发操作数必须为正数: %d", c.MaxConcurrentOps)
		errs = append(errs, err)
	}
	if c.MaxConcurrentOps > 100 {
		err := fmt.Errorf("并发操作数过大: %d (最大100)", c.MaxConcurrentOps)
		errs = append(errs, err)
	}

	if c.MaxBackupFiles < 0 {
		err := fmt.Errorf("最大备份文件数不能为负数: %d", c.MaxBackupFiles)
		errs = append(errs, err)
	}

	if c.TrashMaxSize < 0 {
		err := fmt.Errorf("回收站最大容量不能为负数: %d MB", c.TrashMaxSize)
		errs = append(errs, err)
	}

	if c.BackupRetentionDays < 0 {
		err := fmt.Errorf("备份保留天数不能为负数: %d", c.BackupRetentionDays)
		errs = append(errs, err)
	}
	if c.BackupRetentionDays > 365 {
		err := fmt.Errorf("备份保留天数过长: %d (最大365天)", c.BackupRetentionDays)
		errs = append(errs, err)
	}

	if c.LogRetentionDays < 0 {
		err := fmt.Errorf("日志保留天数不能为负数: %d", c.LogRetentionDays)
		errs = append(errs, err)
	}
	if c.LogRetentionDays > 30 {
		err := fmt.Errorf("日志保留天数过长: %d (最大30天)", c.LogRetentionDays)
		errs = append(errs, err)
	}

	// 平台特定路径验证
	if runtime.GOOS == "windows" {
		if c.Windows.RecycleBinPath != "" {
			if _, err := os.Stat(c.Windows.RecycleBinPath); err != nil {
				if os.IsNotExist(err) {
					err = fmt.Errorf("Windows回收站路径不存在: %s", c.Windows.RecycleBinPath)
				} else {
					err = fmt.Errorf("Windows回收站路径访问失败: %v", err)
				}
				errs = append(errs, err)
			}
		}
	} else {
		if runtime.GOOS == "linux" {
			if c.Linux.TrashDir != "" {
				if _, err := os.Stat(c.Linux.TrashDir); err != nil {
					if os.IsNotExist(err) {
						err = fmt.Errorf("Linux回收站路径不存在: %s", c.Linux.TrashDir)
					} else {
						err = fmt.Errorf("Linux回收站路径访问失败: %v", err)
					}
					errs = append(errs, err)
				}
			}
		} else if runtime.GOOS == "darwin" {
			if c.Darwin.TrashDir != "" {
				if _, err := os.Stat(c.Darwin.TrashDir); err != nil {
					if os.IsNotExist(err) {
						err = fmt.Errorf("macOS回收站路径不存在: %s", c.Darwin.TrashDir)
					} else {
						err = fmt.Errorf("macOS回收站路径访问失败: %v", err)
					}
					errs = append(errs, err)
				}
			}
		}
	}

	// 返回所有错误
	if len(errs) > 0 {
		var errorMsgs []string
		for _, err := range errs {
			errorMsgs = append(errorMsgs, err.Error())
		}
		return fmt.Errorf("配置验证失败:\n%s", strings.Join(errorMsgs, "\n"))
	}

	return nil
}

// Save 保存配置到文件
func (c *Config) Save(path string) error {
	// 创建目录
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	// 序列化配置
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	// 写入文件
	return os.WriteFile(path, data, 0644)
}
