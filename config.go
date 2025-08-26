package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// 配置默认常量
const (
	DefaultMaxBackupFiles    int64 = 1000
	DefaultTrashMaxSize      int64 = 10000   // MB
	DefaultMaxConfigFileSize int64 = 1048576 // 1MB 配置文件大小限制

	// 配置限制常量
	DefaultMaxPathLength    int = 4096 // 最大路径长度
	DefaultMaxConcurrentOps int = 100  // 最大并发操作数
)

// Config 配置结构体
type Config struct {
	// 版本管理
	Version       string `json:"version"`        // 配置文件版本
	SchemaVersion string `json:"schema_version"` // 配置架构版本

	// 基本设置
	UseRecycleBin   bool   `json:"use_recycle_bin"`
	InteractiveMode string `json:"interactive_mode"` // "always", "never", "confirm"
	Language        string `json:"language"`
	LogLevel        string `json:"log_level"`
	SafeMode        string `json:"safe_mode"` // "strict", "normal", "relaxed"

	// 限制设置
	MaxBackupFiles      int64   `json:"max_backup_files"`
	TrashMaxSize        int64   `json:"trash_max_size"` // MB
	MaxFileSize         int64   `json:"max_file_size"`  // bytes
	SimilarityThreshold float64 `json:"similarity_threshold"`
	MaxPathLength       int     `json:"max_path_length"`
	MaxConcurrentOps    int     `json:"max_concurrent_ops"`

	// 安全设置
	EnableSecurityChecks      bool `json:"enable_security_checks"`
	EnableMalwareScan         bool `json:"enable_malware_scan"`
	EnablePathValidation      bool `json:"enable_path_validation"`
	EnableHiddenCheck         bool `json:"enable_hidden_check"`
	EnableOverwriteProtection bool `json:"enable_overwrite_protection"`

	// 高级设置
	BackupRetentionDays int    `json:"backup_retention_days"`
	LogRetentionDays    int    `json:"log_retention_days"`
	EnableTelemetry     bool   `json:"enable_telemetry"`
	TelemetryEndpoint   string `json:"telemetry_endpoint"`

	// 输出前缀设置
	OutputPrefixEnabled bool   `json:"output_prefix_enabled"` // 是否为单行消息自动添加前缀
	OutputPrefix        string `json:"output_prefix"`         // 自定义前缀，默认 "DelGuard: "

	// 日志轮转设置
	LogMaxSize     int64 `json:"log_max_size"`     // 日志文件最大大小(MB)
	LogMaxBackups  int   `json:"log_max_backups"`  // 日志文件最大备份数量
	LogRotateDaily bool  `json:"log_rotate_daily"` // 是否按日轮转
	LogCompress    bool  `json:"log_compress"`     // 是否压缩旧日志

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

// LoadConfigWithOverride 允许通过外部传入的路径优先加载配置
func LoadConfigWithOverride(overridePath string) (*Config, error) {
	if defaultConfig != nil && defaultConfig.ConfigPath != "" {
		return defaultConfig, nil
	}

	cfg := &Config{}
	cfg.setDefaults()

	var paths []string
	if strings.TrimSpace(overridePath) != "" {
		paths = []string{overridePath}
	} else {
		paths = cfg.findConfigPaths()
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			if err := cfg.loadFromFile(p); err != nil {
				LogWarn("config", p, fmt.Sprintf("配置文件加载失败: %v", err))
				continue
			}
			cfg.ConfigPath = p
			break
		}
	}

	cfg.loadFromEnv()
	defaultConfig = cfg
	return cfg, nil
}

// setDefaults 设置默认配置值
func (c *Config) setDefaults() {
	// 设置版本信息
	c.Version = "1.0.0"
	c.SchemaVersion = "1.0"

	c.UseRecycleBin = true
	c.InteractiveMode = "confirm"
	c.Language = "auto"
	c.LogLevel = LogLevelInfoStr
	c.SafeMode = "normal"

	c.MaxBackupFiles = DefaultMaxBackupFiles
	c.TrashMaxSize = DefaultTrashMaxSize
	c.MaxFileSize = DefaultMaxFileSize
	c.SimilarityThreshold = DefaultSimilarityThreshold
	c.MaxPathLength = DefaultMaxPathLength
	c.MaxConcurrentOps = DefaultMaxConcurrentOps

	c.EnableSecurityChecks = true
	c.EnableMalwareScan = false // 默认关闭，提高性能
	c.EnablePathValidation = true
	c.EnableHiddenCheck = true
	c.EnableOverwriteProtection = true // 默认启用覆盖保护

	c.BackupRetentionDays = 30
	c.LogRetentionDays = 7
	c.EnableTelemetry = false
	c.TelemetryEndpoint = "https://telemetry.delguard.io/v1/report"

	// 输出前缀默认值
	c.OutputPrefixEnabled = false
	c.OutputPrefix = ""

	// 日志轮转默认设置
	c.LogMaxSize = 100       // 100MB
	c.LogMaxBackups = 10     // 10个备份文件
	c.LogRotateDaily = false // 默认不按日轮转
	c.LogCompress = true     // 默认压缩旧日志

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
		// 兼容多扩展名
		for _, name := range []string{"config.json", "config.jsonc", "config.ini", "config.cfg", "config.conf", ".env", "delguard.properties"} {
			paths = append(paths, filepath.Join(home, ".delguard", name))
		}
	}

	// 系统配置目录
	switch runtime.GOOS {
	case "windows":
		if systemRoot := os.Getenv("SystemRoot"); systemRoot != "" {
			for _, name := range []string{"config.json", "config.jsonc", "config.ini", "config.cfg", "config.conf", ".env", "delguard.properties"} {
				paths = append(paths, filepath.Join(systemRoot, "delguard", name))
			}
		}
	default:
		for _, name := range []string{"config.json", "config.jsonc", "config.ini", "config.cfg", "config.conf", ".env", "delguard.properties"} {
			paths = append(paths, filepath.Join("/etc/delguard", name))
		}
	}

	// 当前目录
	for _, name := range []string{"config.json", "config.jsonc", "config.ini", "config.cfg", "config.conf", ".env", "delguard.properties"} {
		paths = append(paths, name)
	}

	return paths
}

// loadFromFile 从文件加载配置
func (c *Config) loadFromFile(path string) error {
	// 验证文件路径，防止路径遍历
	if err := validateConfigPath(path); err != nil {
		return fmt.Errorf("配置文件路径验证失败: %v", err)
	}

	// 检查文件大小，防止加载过大文件
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Size() > DefaultMaxConfigFileSize { // 使用常量限制配置文件大小
		return fmt.Errorf("配置文件过大: %d bytes", info.Size())
	}

	// 读取文件内容
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// 根据扩展名解析
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		if err := validateJSONContent(data); err != nil {
			return fmt.Errorf("JSON内容验证失败: %v", err)
		}
		return safeJSONUnmarshal(data, c)
	case ".jsonc":
		cleaned := stripJSONComments(string(data))
		if err := validateJSONContent([]byte(cleaned)); err != nil {
			return fmt.Errorf("JSONC内容验证失败: %v", err)
		}
		return safeJSONUnmarshal([]byte(cleaned), c)
	case ".ini", ".cfg", ".conf":
		return parseINIIntoConfig(string(data), c)
	case ".env", ".properties":
		return parseEnvFileIntoConfig(string(data), c)
	default:
		return fmt.Errorf("不支持的配置文件类型: %s", filepath.Ext(path))
	}
}

// loadFromEnv 从环境变量加载配置
func (c *Config) loadFromEnv() {
	// 基本设置
	if v := os.Getenv("DELGUARD_USE_RECYCLE_BIN"); v != "" {
		if validateEnvBool(v) {
			if b, err := strconv.ParseBool(v); err == nil {
				c.UseRecycleBin = b
			}
		}
	}

	if v := os.Getenv("DELGUARD_INTERACTIVE_MODE"); v != "" {
		if validateEnvString(v, []string{"always", "never", "confirm"}) {
			c.InteractiveMode = v
		}
	}

	if v := os.Getenv("DELGUARD_LANGUAGE"); v != "" {
		if validateEnvString(v, []string{"auto", "zh", "en"}) {
			c.Language = v
		}
	}

	if v := os.Getenv("DELGUARD_LOG_LEVEL"); v != "" {
		if validateEnvString(v, []string{LogLevelDebugStr, LogLevelInfoStr, LogLevelWarnStr, LogLevelErrorStr, LogLevelFatalStr}) {
			c.LogLevel = v
		}
	}

	if v := os.Getenv("DELGUARD_SAFE_MODE"); v != "" {
		if validateEnvString(v, []string{"strict", "normal", "relaxed"}) {
			c.SafeMode = v
		}
	}

	// 限制设置
	if v := os.Getenv("DELGUARD_MAX_BACKUP_FILES"); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil && validateEnvInt64(i, 0, 10000) {
			c.MaxBackupFiles = i
		}
	}

	if v := os.Getenv("DELGUARD_TRASH_MAX_SIZE"); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil && validateEnvInt64(i, 0, 100000) {
			c.TrashMaxSize = i
		}
	}

	if v := os.Getenv("DELGUARD_MAX_FILE_SIZE"); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil && validateEnvInt64(i, 0, 100*GB) {
			c.MaxFileSize = i
		}
	}

	// 安全设置
	if v := os.Getenv("DELGUARD_ENABLE_SECURITY_CHECKS"); v != "" {
		if validateEnvBool(v) {
			if b, err := strconv.ParseBool(v); err == nil {
				c.EnableSecurityChecks = b
			}
		}
	}

	if v := os.Getenv("DELGUARD_ENABLE_MALWARE_SCAN"); v != "" {
		if validateEnvBool(v) {
			if b, err := strconv.ParseBool(v); err == nil {
				c.EnableMalwareScan = b
			}
		}
	}

	// 输出前缀相关
	if v := os.Getenv("DELGUARD_OUTPUT_PREFIX_ENABLED"); v != "" {
		if validateEnvBool(v) {
			if b, err := strconv.ParseBool(v); err == nil {
				c.OutputPrefixEnabled = b
			}
		}
	}
	if v := os.Getenv("DELGUARD_OUTPUT_PREFIX"); v != "" {
		// 允许空字符串以彻底移除前缀，但通常结合 ENABLED=false 更清晰
		// 这里直接赋值，长度限制由 T() 内部或调用处控制
		c.OutputPrefix = v
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

// Validate 验证配置有效性 - 增强版
func (c *Config) Validate() error {
	var errs []error

	// 基础配置验证
	if err := c.validateBasicSettings(); err != nil {
		errs = append(errs, err)
	}

	// 数值范围验证
	if err := c.validateNumericalRanges(); err != nil {
		errs = append(errs, err)
	}

	// 安全设置验证
	if err := c.validateSecuritySettings(); err != nil {
		errs = append(errs, err)
	}

	// 网络配置验证
	if err := c.validateNetworkSettings(); err != nil {
		errs = append(errs, err)
	}

	// 平台特定验证
	if err := c.validatePlatformSpecific(); err != nil {
		errs = append(errs, err)
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

// validateBasicSettings 验证基础设置
func (c *Config) validateBasicSettings() error {
	var errs []error

	// 验证日志级别
	validLevels := map[string]bool{
		LogLevelDebugStr: true, LogLevelInfoStr: true, LogLevelWarnStr: true, LogLevelErrorStr: true, LogLevelFatalStr: true,
	}
	if !validLevels[c.LogLevel] {
		errs = append(errs, fmt.Errorf("无效的日志级别: %s", c.LogLevel))
	}

	// 验证交互模式
	switch c.InteractiveMode {
	case "always", "never", "confirm":
	default:
		errs = append(errs, fmt.Errorf("无效的交互模式: %s", c.InteractiveMode))
	}

	// 验证安全模式
	switch c.SafeMode {
	case "strict", "normal", "relaxed":
	default:
		errs = append(errs, fmt.Errorf("无效的安全模式: %s", c.SafeMode))
	}

	// 验证语言设置
	if !validateLanguage(c.Language) {
		errs = append(errs, fmt.Errorf("不支持的语言: %s", c.Language))
	}

	if len(errs) > 0 {
		return fmt.Errorf("基础设置验证失败: %v", errs)
	}
	return nil
}

// validateNumericalRanges 验证数值范围
func (c *Config) validateNumericalRanges() error {
	var errs []error

	// 验证文件大小限制
	if c.MaxFileSize < 0 {
		errs = append(errs, fmt.Errorf("最大文件大小不能为负数: %d", c.MaxFileSize))
	} else if c.MaxFileSize > 100*GB { // 100GB
		errs = append(errs, fmt.Errorf("最大文件大小过大: %d (最大100GB)", c.MaxFileSize))
	}

	// 验证路径长度
	if c.MaxPathLength <= 0 {
		errs = append(errs, fmt.Errorf("最大路径长度必须为正数: %d", c.MaxPathLength))
	} else if c.MaxPathLength > 32767 {
		errs = append(errs, fmt.Errorf("最大路径长度过长: %d (最大32767)", c.MaxPathLength))
	}

	// 验证并发操作数
	if c.MaxConcurrentOps <= 0 {
		errs = append(errs, fmt.Errorf("最大并发操作数必须为正数: %d", c.MaxConcurrentOps))
	} else if c.MaxConcurrentOps > 1000 {
		errs = append(errs, fmt.Errorf("并发操作数过大: %d (最大1000)", c.MaxConcurrentOps))
	}

	// 验证其他数值
	if c.MaxBackupFiles < 0 {
		errs = append(errs, fmt.Errorf("最大备份文件数不能为负数: %d", c.MaxBackupFiles))
	} else if c.MaxBackupFiles > 1000000 {
		errs = append(errs, fmt.Errorf("最大备份文件数过大: %d (最大1000000)", c.MaxBackupFiles))
	}

	if c.TrashMaxSize < 0 {
		errs = append(errs, fmt.Errorf("回收站最大容量不能为负数: %d MB", c.TrashMaxSize))
	} else if c.TrashMaxSize > 10000000 { // 10TB
		errs = append(errs, fmt.Errorf("回收站最大容量过大: %d MB (最大10TB)", c.TrashMaxSize))
	}

	if c.BackupRetentionDays < 0 {
		errs = append(errs, fmt.Errorf("备份保留天数不能为负数: %d", c.BackupRetentionDays))
	} else if c.BackupRetentionDays > 3650 {
		errs = append(errs, fmt.Errorf("备份保留天数过长: %d (最大3650天)", c.BackupRetentionDays))
	}

	if c.LogRetentionDays < 0 {
		errs = append(errs, fmt.Errorf("日志保留天数不能为负数: %d", c.LogRetentionDays))
	} else if c.LogRetentionDays > 365 {
		errs = append(errs, fmt.Errorf("日志保留天数过长: %d (最大365天)", c.LogRetentionDays))
	}

	if len(errs) > 0 {
		return fmt.Errorf("数值范围验证失败: %v", errs)
	}
	return nil
}

// validateSecuritySettings 验证安全设置
func (c *Config) validateSecuritySettings() error {
	var errs []error

	// 在严格模式下，必须启用某些安全检查
	if c.SafeMode == "strict" {
		if !c.EnableSecurityChecks {
			errs = append(errs, fmt.Errorf("严格模式下必须启用安全检查"))
		}
		if !c.EnablePathValidation {
			errs = append(errs, fmt.Errorf("严格模式下必须启用路径验证"))
		}
	}

	// 检查不合理的组合
	if !c.EnableSecurityChecks && c.EnableMalwareScan {
		errs = append(errs, fmt.Errorf("在禁用安全检查的情况下不能启用恶意软件扫描"))
	}

	if len(errs) > 0 {
		return fmt.Errorf("安全设置验证失败: %v", errs)
	}
	return nil
}

// validateNetworkSettings 验证网络设置
func (c *Config) validateNetworkSettings() error {
	var errs []error

	if c.EnableTelemetry {
		if c.TelemetryEndpoint == "" {
			errs = append(errs, fmt.Errorf("启用遥测时必须设置遥测端点"))
		} else if err := validateURL(c.TelemetryEndpoint); err != nil {
			errs = append(errs, fmt.Errorf("遥测端点URL无效: %v", err))
		} else if err := validateURLSecurity(c.TelemetryEndpoint); err != nil {
			errs = append(errs, fmt.Errorf("遥测端点URL不安全: %v", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("网络设置验证失败: %v", errs)
	}
	return nil
}

// validatePlatformSpecific 验证平台特定设置
func (c *Config) validatePlatformSpecific() error {
	var errs []error

	switch runtime.GOOS {
	case "windows":
		if c.Windows.RecycleBinPath != "" {
			if err := validatePath(c.Windows.RecycleBinPath); err != nil {
				errs = append(errs, fmt.Errorf("Windows回收站路径不安全: %v", err))
			}
		}
	case "linux":
		if c.Linux.TrashDir != "" {
			if err := validatePath(c.Linux.TrashDir); err != nil {
				errs = append(errs, fmt.Errorf("Linux回收站路径不安全: %v", err))
			}
		}
	case "darwin":
		if c.Darwin.TrashDir != "" {
			if err := validatePath(c.Darwin.TrashDir); err != nil {
				errs = append(errs, fmt.Errorf("macOS回收站路径不安全: %v", err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("平台特定设置验证失败: %v", errs)
	}
	return nil
}

// SaveConfig 保存配置到文件
func (c *Config) SaveConfig() error {
	configPath, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("获取配置路径失败: %v", err)
	}

	// 创建配置目录
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 保存配置
	return c.SaveWithVersion(configPath, "1.0")
}

// 辅助函数
func stripJSONComments(s string) string {
	var b strings.Builder
	inLine := false
	inBlock := false
	for i := 0; i < len(s); i++ {
		if inLine {
			if s[i] == '\n' {
				inLine = false
				b.WriteByte(s[i])
			}
			continue
		}
		if inBlock {
			if i+1 < len(s) && s[i] == '*' && s[i+1] == '/' {
				inBlock = false
				i++
			}
			continue
		}
		if i+1 < len(s) {
			if s[i] == '/' && s[i+1] == '/' {
				inLine = true
				i++
				continue
			}
			if s[i] == '/' && s[i+1] == '*' {
				inBlock = true
				i++
				continue
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func parseEnvFileIntoConfig(content string, c *Config) error {
	mp := map[string]string{}
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		var key, val string
		if idx := strings.IndexAny(line, "=:"); idx >= 0 {
			key = strings.TrimSpace(line[:idx])
			val = strings.TrimSpace(line[idx+1:])
		} else {
			continue
		}
		val = strings.Trim(val, "\"'")
		if key != "" {
			mp[strings.ToUpper(strings.ReplaceAll(key, ".", "_"))] = val
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	applyKVToConfig(mp, c)
	return nil
}

func parseINIIntoConfig(content string, c *Config) error {
	section := ""
	kv := map[string]string{}
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(line[1 : len(line)-1])
			continue
		}
		var key, val string
		if idx := strings.IndexAny(line, "=:"); idx >= 0 {
			key = strings.TrimSpace(line[:idx])
			val = strings.TrimSpace(line[idx+1:])
		} else {
			continue
		}
		val = strings.Trim(val, "\"'")
		fullKey := key
		upKey := strings.ToUpper(key)
		if !strings.HasPrefix(upKey, "DELGUARD_") {
			if section != "" {
				fullKey = section + "." + key
			}
		}
		kv[strings.ToUpper(strings.ReplaceAll(fullKey, "-", "_"))] = val
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	applyKVToConfig(kv, c)
	return nil
}

func applyKVToConfig(kv map[string]string, c *Config) {
	// 布尔键
	boolKeys := map[string]*bool{
		"DELGUARD_USE_RECYCLE_BIN":              &c.UseRecycleBin,
		"DELGUARD_ENABLE_SECURITY_CHECKS":       &c.EnableSecurityChecks,
		"DELGUARD_ENABLE_MALWARE_SCAN":          &c.EnableMalwareScan,
		"DELGUARD_ENABLE_PATH_VALIDATION":       &c.EnablePathValidation,
		"DELGUARD_ENABLE_HIDDEN_CHECK":          &c.EnableHiddenCheck,
		"DELGUARD_ENABLE_OVERWRITE_PROTECTION":  &c.EnableOverwriteProtection,
		"DELGUARD_ENABLE_TELEMETRY":             &c.EnableTelemetry,
		"DELGUARD_WINDOWS_USE_SYSTEM_TRASH":     &c.Windows.UseSystemTrash,
		"DELGUARD_WINDOWS_ENABLE_UAC_PROMPT":    &c.Windows.EnableUACPrompt,
		"DELGUARD_WINDOWS_CHECK_FILE_OWNERSHIP": &c.Windows.CheckFileOwnership,
		"DELGUARD_LINUX_USE_XDG_TRASH":          &c.Linux.UseXDGTrash,
		"DELGUARD_LINUX_CHECK_SELINUX":          &c.Linux.CheckSELinux,
		"DELGUARD_LINUX_CHECK_APPARMOR":         &c.Linux.CheckAppArmor,
		"DELGUARD_DARWIN_USE_SYSTEM_TRASH":      &c.Darwin.UseSystemTrash,
		"DELGUARD_DARWIN_CHECK_FILEVAULT":       &c.Darwin.CheckFileVault,
		"DELGUARD_DARWIN_CHECK_GATEKEEPER":      &c.Darwin.CheckGatekeeper,
		// 输出前缀
		"DELGUARD_OUTPUT_PREFIX_ENABLED": &c.OutputPrefixEnabled,
	}

	for k, ptr := range boolKeys {
		if v, ok := kv[k]; ok && validateEnvBool(v) {
			if b, err := strconv.ParseBool(strings.ToLower(v)); err == nil {
				*ptr = b
			}
		}
	}

	// 字符串键
	strSet := map[string]*string{
		"DELGUARD_INTERACTIVE_MODE":         &c.InteractiveMode,
		"DELGUARD_LANGUAGE":                 &c.Language,
		"DELGUARD_LOG_LEVEL":                &c.LogLevel,
		"DELGUARD_SAFE_MODE":                &c.SafeMode,
		"DELGUARD_TELEMETRY_ENDPOINT":       &c.TelemetryEndpoint,
		"DELGUARD_WINDOWS_RECYCLE_BIN_PATH": &c.Windows.RecycleBinPath,
		"DELGUARD_LINUX_TRASH_DIR":          &c.Linux.TrashDir,
		"DELGUARD_DARWIN_TRASH_DIR":         &c.Darwin.TrashDir,
		// 输出前缀
		"DELGUARD_OUTPUT_PREFIX": &c.OutputPrefix,
	}
	for k, ptr := range strSet {
		if v, ok := kv[k]; ok {
			*ptr = v
		}
	}

	// 数值键
	int64Set := map[string]*int64{
		"DELGUARD_MAX_BACKUP_FILES": &c.MaxBackupFiles,
		"DELGUARD_TRASH_MAX_SIZE":   &c.TrashMaxSize,
		"DELGUARD_MAX_FILE_SIZE":    &c.MaxFileSize,
	}
	for k, ptr := range int64Set {
		if v, ok := kv[k]; ok {
			if i, err := strconv.ParseInt(v, 10, 64); err == nil {
				*ptr = i
			}
		}
	}
	intSet := map[string]*int{
		"DELGUARD_MAX_PATH_LENGTH":       &c.MaxPathLength,
		"DELGUARD_MAX_CONCURRENT_OPS":    &c.MaxConcurrentOps,
		"DELGUARD_BACKUP_RETENTION_DAYS": &c.BackupRetentionDays,
		"DELGUARD_LOG_RETENTION_DAYS":    &c.LogRetentionDays,
	}
	for k, ptr := range intSet {
		if v, ok := kv[k]; ok {
			if i, err := strconv.Atoi(v); err == nil {
				*ptr = i
			}
		}
	}
}

// 验证和安全函数
func validateConfigPath(path string) error {
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("检测到路径遍历攻击")
	}
	if !filepath.IsAbs(cleanPath) {
		abs, err := filepath.Abs(cleanPath)
		if err != nil {
			return fmt.Errorf("无法获取绝对路径: %v", err)
		}
		cleanPath = abs
	}
	allowed := map[string]bool{
		".json": true, ".jsonc": true, ".ini": true, ".cfg": true, ".conf": true, ".env": true, ".properties": true,
	}
	if !allowed[strings.ToLower(filepath.Ext(cleanPath))] {
		return fmt.Errorf("仅允许以下配置文件类型: .json/.jsonc/.ini/.cfg/.conf/.env/.properties")
	}
	return nil
}

func validateJSONContent(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("JSON内容为空")
	}
	var temp interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("JSON格式无效: %v", err)
	}
	if err := checkJSONDepth(temp, 0, 10); err != nil {
		return err
	}
	return nil
}

func checkJSONDepth(data interface{}, currentDepth, maxDepth int) error {
	if currentDepth > maxDepth {
		return fmt.Errorf("JSON嵌套深度过大: %d", currentDepth)
	}
	switch v := data.(type) {
	case map[string]interface{}:
		for _, value := range v {
			if err := checkJSONDepth(value, currentDepth+1, maxDepth); err != nil {
				return err
			}
		}
	case []interface{}:
		for _, value := range v {
			if err := checkJSONDepth(value, currentDepth+1, maxDepth); err != nil {
				return err
			}
		}
	}
	return nil
}

func safeJSONUnmarshal(data []byte, c *Config) error {
	tempConfig := &Config{}
	tempConfig.setDefaults()
	if err := json.Unmarshal(data, tempConfig); err != nil {
		return fmt.Errorf("JSON解析失败: %v", err)
	}
	if err := upgradeConfigVersion(tempConfig); err != nil {
		return fmt.Errorf("配置版本升级失败: %v", err)
	}
	if err := tempConfig.Validate(); err != nil {
		return fmt.Errorf("配置验证失败: %v", err)
	}
	*c = *tempConfig
	return nil
}

func validateEnvString(value string, allowedValues []string) bool {
	if len(value) > 100 {
		return false
	}
	for _, allowed := range allowedValues {
		if value == allowed {
			return true
		}
	}
	return false
}

func validateEnvBool(value string) bool {
	allowedValues := []string{"true", "false", "1", "0", "yes", "no"}
	lowerValue := strings.ToLower(value)
	for _, allowed := range allowedValues {
		if lowerValue == allowed {
			return true
		}
	}
	return false
}

func validateEnvInt64(value, min, max int64) bool {
	return value >= min && value <= max
}

func validateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL不能为空")
	}
	if len(urlStr) > 2048 {
		return fmt.Errorf("URL太长")
	}
	for _, char := range urlStr {
		if char < 32 || char == 127 {
			return fmt.Errorf("URL包含非法字符")
		}
	}
	if !strings.HasPrefix(urlStr, "https://") && !strings.HasPrefix(urlStr, "http://") {
		return fmt.Errorf("仅允许HTTP/HTTPS协议")
	}
	return nil
}

func validatePath(path string) error {
	if path == "" {
		return fmt.Errorf("路径不能为空")
	}
	if len(path) > 32767 {
		return fmt.Errorf("路径太长")
	}
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("检测到路径遍历攻击")
	}
	if !filepath.IsAbs(cleanPath) {
		return fmt.Errorf("必须为绝对路径")
	}
	return nil
}

func validateLanguage(lang string) bool {
	allowedLanguages := []string{"auto", "zh", "en", "zh-CN", "en-US"}
	for _, allowed := range allowedLanguages {
		if lang == allowed {
			return true
		}
	}
	return false
}

func upgradeConfigVersion(config *Config) error {
	currentSchemaVersion := "1.0"
	if config.Version == "" {
		config.Version = "0.9.0"
	}
	if config.SchemaVersion == "" {
		config.SchemaVersion = "0.9"
	}
	if config.SchemaVersion == currentSchemaVersion {
		return nil
	}
	switch config.SchemaVersion {
	case "0.9":
		if err := upgradeFrom09To10(config); err != nil {
			return fmt.Errorf("从0.9升级到1.0失败: %v", err)
		}
		fallthrough
	default:
		if config.SchemaVersion != currentSchemaVersion {
			return fmt.Errorf("不支持的配置架构版本: %s", config.SchemaVersion)
		}
	}
	config.Version = "1.0.0"
	config.SchemaVersion = currentSchemaVersion
	return nil
}

func upgradeFrom09To10(config *Config) error {
	if !hasField(config, "EnableOverwriteProtection") {
		config.EnableOverwriteProtection = true
	}
	if config.MaxConcurrentOps > 1000 {
		config.MaxConcurrentOps = 1000
	}
	return nil
}

func hasField(config *Config, fieldName string) bool {
	switch fieldName {
	case "EnableOverwriteProtection":
		return true
	default:
		return false
	}
}

func (c *Config) Save(path string) error {
	if err := c.Validate(); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}
	configDir := filepath.Dir(path)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("无法创建配置目录: %w", err)
	}
	tempFile := path + ".tmp"
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("无法创建临时配置文件: %w", err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(c); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("配置文件编码失败: %w", err)
	}
	if err := os.Rename(tempFile, path); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("无法保存配置文件: %w", err)
	}
	return nil
}

func (c *Config) GetConfigVersion() (string, string) {
	return c.Version, c.SchemaVersion
}

func (c *Config) IsConfigVersionSupported() bool {
	supportedVersions := []string{"1.0", "0.9"}
	for _, version := range supportedVersions {
		if c.SchemaVersion == version {
			return true
		}
	}
	return false
}

func (c *Config) SaveWithVersion(filePath, version string) error {
	c.Version = version
	if err := c.Validate(); err != nil {
		return fmt.Errorf("配置验证失败: %v", err)
	}
	if _, err := os.Stat(filePath); err == nil {
		backupPath := filePath + ".bak"
		if err := copyFile(filePath, backupPath); err != nil {
			LogWarn("config", filePath, fmt.Sprintf("无法创建配置备份: %v", err))
		}
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}
	tempPath := filePath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("写入临时配置文件失败: %v", err)
	}
	if err := os.Rename(tempPath, filePath); err != nil {
		return fmt.Errorf("替换配置文件失败: %v", err)
	}
	LogInfo("config", filePath, "配置已保存")
	return nil
}

func getConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		homeDir, homeErr := os.UserHomeDir()
		if homeErr != nil {
			return "", fmt.Errorf("无法获取用户主目录: %v", homeErr)
		}
		configDir = filepath.Join(homeDir, ".config")
	}
	delguardDir := filepath.Join(configDir, "delguard")
	configPath := filepath.Join(delguardDir, "config.json")
	if err := validateConfigPath(configPath); err != nil {
		return "", fmt.Errorf("配置路径无效: %v", err)
	}
	return configPath, nil
}

func validateURLSecurity(urlStr string) error {
	if !strings.HasPrefix(urlStr, "https://") {
		return fmt.Errorf("URL必须使用HTTPS协议")
	}
	if strings.Contains(urlStr, "localhost") || strings.Contains(urlStr, "127.0.0.1") {
		return fmt.Errorf("不允许使用本地地址")
	}
	dangerousPatterns := []string{
		"192.168.", "10.", "172.16.", "172.17.", "172.18.", "172.19.",
		"172.20.", "172.21.", "172.22.", "172.23.", "172.24.", "172.25.",
		"172.26.", "172.27.", "172.28.", "172.29.", "172.30.", "172.31.",
	}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(urlStr, pattern) {
			return fmt.Errorf("不允许使用内网地址")
		}
	}
	return nil
}
