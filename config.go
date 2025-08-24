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
	MaxBackupFiles   int64 `json:"max_backup_files"`
	TrashMaxSize     int64 `json:"trash_max_size"` // MB
	MaxFileSize      int64 `json:"max_file_size"`  // bytes
	MaxPathLength    int   `json:"max_path_length"`
	MaxConcurrentOps int   `json:"max_concurrent_ops"`

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
	// 设置版本信息
	c.Version = "1.0.0"
	c.SchemaVersion = "1.0"

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
	c.EnableOverwriteProtection = true // 默认启用覆盖保护

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
	// 验证文件路径，防止路径遍历
	if err := validateConfigPath(path); err != nil {
		return fmt.Errorf("配置文件路径验证失败: %v", err)
	}

	// 检查文件大小，防止加载过大文件
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Size() > 1024*1024 { // 1MB 限制
		return fmt.Errorf("配置文件过大: %d bytes", info.Size())
	}

	// 读取文件内容
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// 验证JSON内容，防止恶意输入
	if err := validateJSONContent(data); err != nil {
		return fmt.Errorf("JSON内容验证失败: %v", err)
	}

	// 使用安全的JSON解析
	return safeJSONUnmarshal(data, c)
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
		if validateEnvString(v, []string{"debug", "info", "warn", "error", "fatal"}) {
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
		if i, err := strconv.ParseInt(v, 10, 64); err == nil && validateEnvInt64(i, 0, 100*1024*1024*1024) {
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
	// 添加最大文件大小上限检查
	if c.MaxFileSize > 100*1024*1024*1024 { // 100GB
		err := fmt.Errorf("最大文件大小过大: %d (最大100GB)", c.MaxFileSize)
		errs = append(errs, err)
	}

	if c.MaxPathLength <= 0 {
		err := fmt.Errorf("最大路径长度必须为正数: %d", c.MaxPathLength)
		errs = append(errs, err)
	}
	// 添加路径长度上限检查
	if c.MaxPathLength > 32767 { // Windows系统限制
		err := fmt.Errorf("最大路径长度过长: %d (最大32767)", c.MaxPathLength)
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
	// 添加备份文件数上限检查
	if c.MaxBackupFiles > 100000 {
		err := fmt.Errorf("最大备份文件数过大: %d (最大100000)", c.MaxBackupFiles)
		errs = append(errs, err)
	}

	if c.TrashMaxSize < 0 {
		err := fmt.Errorf("回收站最大容量不能为负数: %d MB", c.TrashMaxSize)
		errs = append(errs, err)
	}
	// 添加回收站大小上限检查
	if c.TrashMaxSize > 1000000 { // 1TB
		err := fmt.Errorf("回收站最大容量过大: %d MB (最大1TB)", c.TrashMaxSize)
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

	// 验证远程配置地址（如果启用）
	if c.EnableTelemetry && c.TelemetryEndpoint != "" {
		if err := validateURL(c.TelemetryEndpoint); err != nil {
			err = fmt.Errorf("遥测端点URL无效: %v", err)
			errs = append(errs, err)
		}
	}

	// 验证语言设置
	if !validateLanguage(c.Language) {
		err := fmt.Errorf("不支持的语言: %s (支持: auto, zh, en)", c.Language)
		errs = append(errs, err)
	}

	// 平台特定路径验证
	if runtime.GOOS == "windows" {
		if c.Windows.RecycleBinPath != "" {
			// 验证路径安全性
			if err := validatePath(c.Windows.RecycleBinPath); err != nil {
				err = fmt.Errorf("Windows回收站路径不安全: %v", err)
				errs = append(errs, err)
			} else if _, err := os.Stat(c.Windows.RecycleBinPath); err != nil {
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
				// 验证路径安全性
				if err := validatePath(c.Linux.TrashDir); err != nil {
					err = fmt.Errorf("Linux回收站路径不安全: %v", err)
					errs = append(errs, err)
				} else if _, err := os.Stat(c.Linux.TrashDir); err != nil {
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
				// 验证路径安全性
				if err := validatePath(c.Darwin.TrashDir); err != nil {
					err = fmt.Errorf("macOS回收站路径不安全: %v", err)
					errs = append(errs, err)
				} else if _, err := os.Stat(c.Darwin.TrashDir); err != nil {
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

// validateConfigPath 验证配置文件路径安全性
func validateConfigPath(path string) error {
	// 清理路径
	cleanPath := filepath.Clean(path)

	// 检查路径遍历
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("检测到路径遍历攻击")
	}

	// 检查绝对路径
	if !filepath.IsAbs(cleanPath) {
		// 将相对路径转为绝对路径
		abs, err := filepath.Abs(cleanPath)
		if err != nil {
			return fmt.Errorf("无法获取绝对路径: %v", err)
		}
		cleanPath = abs
	}

	// 检查文件扩展名
	if filepath.Ext(cleanPath) != ".json" {
		return fmt.Errorf("仅允许JSON配置文件")
	}

	return nil
}

// validateJSONContent 验证JSON内容安全性
func validateJSONContent(data []byte) error {
	// 检查内容长度
	if len(data) == 0 {
		return fmt.Errorf("JSON内容为空")
	}

	// 检查内容是否为有效JSON（预验证）
	var temp interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("JSON格式无效: %v", err)
	}

	// 检查JSON嵌套深度，防止深度攻击
	if err := checkJSONDepth(temp, 0, 10); err != nil {
		return err
	}

	return nil
}

// checkJSONDepth 检查JSON嵌套深度
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

// safeJSONUnmarshal 安全的JSON解析（带版本管理）
func safeJSONUnmarshal(data []byte, c *Config) error {
	// 创建一个临时配置对象
	tempConfig := &Config{}

	// 设置默认值
	tempConfig.setDefaults()

	// 解析JSON
	if err := json.Unmarshal(data, tempConfig); err != nil {
		return fmt.Errorf("JSON解析失败: %v", err)
	}

	// 检查并升级配置版本
	if err := upgradeConfigVersion(tempConfig); err != nil {
		return fmt.Errorf("配置版本升级失败: %v", err)
	}

	// 验证解析后的配置
	if err := tempConfig.Validate(); err != nil {
		return fmt.Errorf("配置验证失败: %v", err)
	}

	// 只有在验证成功后才更新实际配置
	*c = *tempConfig

	return nil
}

// validateEnvString 验证环境变量字符串值
func validateEnvString(value string, allowedValues []string) bool {
	// 检查长度
	if len(value) > 100 {
		return false
	}

	// 检查是否在允许列表中
	for _, allowed := range allowedValues {
		if value == allowed {
			return true
		}
	}

	return false
}

// validateEnvBool 验证环境变量布尔值
func validateEnvBool(value string) bool {
	// 只允许特定的布尔值
	allowedValues := []string{"true", "false", "1", "0", "yes", "no"}
	lowerValue := strings.ToLower(value)

	for _, allowed := range allowedValues {
		if lowerValue == allowed {
			return true
		}
	}

	return false
}

// validateEnvInt64 验证环境变量整数值
func validateEnvInt64(value, min, max int64) bool {
	return value >= min && value <= max
}

// validateURL 验证URL格式
func validateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL不能为空")
	}

	// 检查URL长度
	if len(urlStr) > 2048 {
		return fmt.Errorf("URL太长")
	}

	// 检查URL是否包含非法字符
	for _, char := range urlStr {
		if char < 32 || char == 127 {
			return fmt.Errorf("URL包含非法字符")
		}
	}

	// 检查协议白名单
	if !strings.HasPrefix(urlStr, "https://") && !strings.HasPrefix(urlStr, "http://") {
		return fmt.Errorf("仅允许HTTP/HTTPS协议")
	}

	return nil
}

// validatePath 验证路径安全性
func validatePath(path string) error {
	if path == "" {
		return fmt.Errorf("路径不能为空")
	}

	// 检查路径长度
	if len(path) > 32767 {
		return fmt.Errorf("路径太长")
	}

	// 清理路径
	cleanPath := filepath.Clean(path)

	// 检查路径遍历
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("检测到路径遍历攻击")
	}

	// 检查是否为绝对路径
	if !filepath.IsAbs(cleanPath) {
		return fmt.Errorf("必须为绝对路径")
	}

	return nil
}

// validateLanguage 验证语言设置
func validateLanguage(lang string) bool {
	allowedLanguages := []string{"auto", "zh", "en", "zh-CN", "en-US"}
	for _, allowed := range allowedLanguages {
		if lang == allowed {
			return true
		}
	}
	return false
}

// upgradeConfigVersion 升级配置版本
func upgradeConfigVersion(config *Config) error {
	currentSchemaVersion := "1.0"

	// 如果没有版本信息，认为是旧版本
	if config.Version == "" {
		config.Version = "0.9.0" // 假设的旧版本
	}
	if config.SchemaVersion == "" {
		config.SchemaVersion = "0.9" // 假设的旧架构版本
	}

	// 检查是否需要升级
	if config.SchemaVersion == currentSchemaVersion {
		return nil // 无需升级
	}

	// 执行版本升级
	switch config.SchemaVersion {
	case "0.9":
		if err := upgradeFrom09To10(config); err != nil {
			return fmt.Errorf("从0.9升级到1.0失败: %v", err)
		}
		fallthrough // 继续升级到更高版本
	default:
		// 不支持的版本
		if config.SchemaVersion != currentSchemaVersion {
			return fmt.Errorf("不支持的配置架构版本: %s", config.SchemaVersion)
		}
	}

	// 更新版本信息
	config.Version = "1.0.0"
	config.SchemaVersion = currentSchemaVersion

	return nil
}

// upgradeFrom09To10 从0.9版本升级到1.0版本
func upgradeFrom09To10(config *Config) error {
	// 在这里添加具体的升级逻辑
	// 例如：新增字段的默认值设置、字段重命名等

	// 示例：设置新增的安全选项默认值
	if !hasField(config, "EnableOverwriteProtection") {
		config.EnableOverwriteProtection = true
	}

	// 示例：调整旧的配置值
	if config.MaxConcurrentOps > 100 {
		config.MaxConcurrentOps = 100 // 限制最大并发数
	}

	return nil
}

// hasField 检查配置中是否包含指定字段（简化实现）
func hasField(config *Config, fieldName string) bool {
	// 这里是一个简化的实现
	// 在实际使用中，可以使用反射或其他方法来检查
	switch fieldName {
	case "EnableOverwriteProtection":
		return true // 假设字段存在
	default:
		return false
	}
}

// Save 保存配置到指定文件
func (c *Config) Save(path string) error {
	// 验证配置
	if err := c.Validate(); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	// 确保配置目录存在
	configDir := filepath.Dir(path)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("无法创建配置目录: %w", err)
	}

	// 创建临时文件
	tempFile := path + ".tmp"
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("无法创建临时配置文件: %w", err)
	}
	defer file.Close()

	// 使用JSON编码配置
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(c); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("配置文件编码失败: %w", err)
	}

	// 原子性替换配置文件
	if err := os.Rename(tempFile, path); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("无法保存配置文件: %w", err)
	}

	return nil
}

// GetConfigVersion 获取配置版本信息
func (c *Config) GetConfigVersion() (string, string) {
	return c.Version, c.SchemaVersion
}

// IsConfigVersionSupported 检查配置版本是否支持
func (c *Config) IsConfigVersionSupported() bool {
	supportedVersions := []string{"1.0", "0.9"}
	for _, version := range supportedVersions {
		if c.SchemaVersion == version {
			return true
		}
	}
	return false
}

// SaveWithVersion 保存配置并指定版本
func (c *Config) SaveWithVersion(filePath, version string) error {
	// 设置配置版本
	c.Version = version

	// 验证配置
	if err := c.Validate(); err != nil {
		return fmt.Errorf("配置验证失败: %v", err)
	}

	// 创建备份
	if _, err := os.Stat(filePath); err == nil {
		backupPath := filePath + ".bak"
		if err := copyFile(filePath, backupPath); err != nil {
			LogWarn("config", filePath, fmt.Sprintf("无法创建配置备份: %v", err))
		}
	}

	// 保存配置
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 写入临时文件
	tempPath := filePath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("写入临时配置文件失败: %v", err)
	}

	// 原子性地替换原文件
	if err := os.Rename(tempPath, filePath); err != nil {
		return fmt.Errorf("替换配置文件失败: %v", err)
	}

	LogInfo("config", filePath, "配置已保存")
	return nil
}

// getConfigPath 获取配置文件路径
func getConfigPath() (string, error) {
	// 获取用户配置目录
	configDir, err := os.UserConfigDir()
	if err != nil {
		// 如果获取失败，尝试使用用户主目录
		homeDir, homeErr := os.UserHomeDir()
		if homeErr != nil {
			return "", fmt.Errorf("无法获取用户主目录: %v", homeErr)
		}
		configDir = filepath.Join(homeDir, ".config")
	}
	
	// 创建DelGuard配置目录路径
	delguardDir := filepath.Join(configDir, "delguard")
	configPath := filepath.Join(delguardDir, "config.json")
	
	// 检查路径是否有效
	if err := validateConfigPath(configPath); err != nil {
		return "", fmt.Errorf("配置路径无效: %v", err)
	}
	
	return configPath, nil
}
