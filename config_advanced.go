package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// AdvancedConfig 高级配置结构
type AdvancedConfig struct {
	// 基础配置继承
	*Config

	// 高级操作配置
	Performance PerformanceConfig `json:"performance"`
	UI          UIConfig          `json:"ui"`
	Backup      BackupConfig      `json:"backup"`
	Security    SecurityConfig    `json:"security"`
	Integration IntegrationConfig `json:"integration"`
	Filters     FilterConfig      `json:"filters"`
	Logging     LoggingConfig     `json:"logging"`
	Hooks       HooksConfig       `json:"hooks"`
}

// PerformanceConfig 性能配置
type PerformanceConfig struct {
	ShowProgress  bool          `json:"show_progress"`
	BatchSize     int           `json:"batch_size"`
	Parallel      bool          `json:"parallel"`
	MaxWorkers    int           `json:"max_workers"`
	Timeout       time.Duration `json:"timeout"`
	EagerMode     bool          `json:"eager_mode"`
	SmartCleanup  bool          `json:"smart_cleanup"`
	MemoryLimit   int64         `json:"memory_limit_mb"`
	CacheSize     int           `json:"cache_size"`
	OptimizeSpeed bool          `json:"optimize_speed"`
	LazyLoad      bool          `json:"lazy_load"`
	PreloadData   bool          `json:"preload_data"`
}

// UIConfig 用户界面配置
type UIConfig struct {
	ColorOutput    bool   `json:"color_output"`
	ShowStats      bool   `json:"show_stats"`
	Notifications  bool   `json:"notifications"`
	ProgressStyle  string `json:"progress_style"`
	Theme          string `json:"theme"`
	IconSet        string `json:"icon_set"`
	AnimatedOutput bool   `json:"animated_output"`
	DetailLevel    string `json:"detail_level"`
	ConfirmStyle   string `json:"confirm_style"`
	ErrorDisplay   string `json:"error_display"`
	SuccessDisplay string `json:"success_display"`
	CompactMode    bool   `json:"compact_mode"`
	FullScreen     bool   `json:"full_screen"`
	AutoResize     bool   `json:"auto_resize"`
}

// BackupConfig 备份配置
type BackupConfig struct {
	AutoBackup        bool   `json:"auto_backup"`
	BackupDir         string `json:"backup_dir"`
	CompressionLevel  int    `json:"compression_level"`
	VerifyIntegrity   bool   `json:"verify_integrity"`
	PreserveTimes     bool   `json:"preserve_times"`
	BackupFormat      string `json:"backup_format"`
	MaxBackups        int    `json:"max_backups"`
	BackupRotation    string `json:"backup_rotation"`
	EncryptBackups    bool   `json:"encrypt_backups"`
	BackupSchedule    string `json:"backup_schedule"`
	IncrementalBackup bool   `json:"incremental_backup"`
	RemoteBackup      bool   `json:"remote_backup"`
	BackupLocation    string `json:"backup_location"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	SecureDelete      bool     `json:"secure_delete"`
	OverwritePasses   int      `json:"overwrite_passes"`
	SecureRandom      bool     `json:"secure_random"`
	VirusScan         bool     `json:"virus_scan"`
	HashVerification  bool     `json:"hash_verification"`
	DigitalSignature  bool     `json:"digital_signature"`
	AccessControl     bool     `json:"access_control"`
	AuditTrail        bool     `json:"audit_trail"`
	EncryptionEnabled bool     `json:"encryption_enabled"`
	KeyManagement     string   `json:"key_management"`
	TrustedSources    []string `json:"trusted_sources"`
	BlacklistedPaths  []string `json:"blacklisted_paths"`
	WhitelistedUsers  []string `json:"whitelisted_users"`
	SecurityLevel     string   `json:"security_level"`
}

// IntegrationConfig 集成配置
type IntegrationConfig struct {
	HooksEnabled       bool              `json:"hooks_enabled"`
	CustomScript       string            `json:"custom_script"`
	WebhookURL         string            `json:"webhook_url"`
	NotificationApps   []string          `json:"notification_apps"`
	CloudSync          bool              `json:"cloud_sync"`
	DatabaseLog        bool              `json:"database_log"`
	APIEndpoints       map[string]string `json:"api_endpoints"`
	PluginDirectory    string            `json:"plugin_directory"`
	ExternalCommands   map[string]string `json:"external_commands"`
	EnvironmentVars    map[string]string `json:"environment_vars"`
	ServiceIntegration bool              `json:"service_integration"`
	MessageQueues      []string          `json:"message_queues"`
}

// FilterConfig 过滤配置
type FilterConfig struct {
	SkipHidden         bool          `json:"skip_hidden"`
	FileSizeLimit      int64         `json:"file_size_limit"`
	IncludePattern     string        `json:"include_pattern"`
	ExcludePattern     string        `json:"exclude_pattern"`
	RegexMode          bool          `json:"regex_mode"`
	CaseSensitive      bool          `json:"case_sensitive"`
	FollowSymlinks     bool          `json:"follow_symlinks"`
	FileTypeFilters    []string      `json:"file_type_filters"`
	AgeFilter          time.Duration `json:"age_filter"`
	SizeFilter         string        `json:"size_filter"`
	NameFilter         string        `json:"name_filter"`
	ContentFilter      string        `json:"content_filter"`
	PermissionFilter   string        `json:"permission_filter"`
	OwnerFilter        string        `json:"owner_filter"`
	ModifiedFilter     string        `json:"modified_filter"`
	AccessedFilter     string        `json:"accessed_filter"`
	CreatedFilter      string        `json:"created_filter"`
	ConflictResolution string        `json:"conflict_resolution"`
	CustomFilters      []string      `json:"custom_filters"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	LogFormat       string `json:"log_format"`
	LogLevel        string `json:"log_level"`
	LogFile         string `json:"log_file"`
	LogRotation     bool   `json:"log_rotation"`
	LogMaxSize      int64  `json:"log_max_size"`
	LogMaxFiles     int    `json:"log_max_files"`
	LogCompression  bool   `json:"log_compression"`
	RemoteLogging   bool   `json:"remote_logging"`
	StructuredLogs  bool   `json:"structured_logs"`
	DebugMode       bool   `json:"debug_mode"`
	TraceEnabled    bool   `json:"trace_enabled"`
	MetricsEnabled  bool   `json:"metrics_enabled"`
	PerformanceLogs bool   `json:"performance_logs"`
}

// HooksConfig 钩子配置
type HooksConfig struct {
	PreDelete   []string            `json:"pre_delete"`
	PostDelete  []string            `json:"post_delete"`
	OnError     []string            `json:"on_error"`
	OnSuccess   []string            `json:"on_success"`
	PreBackup   []string            `json:"pre_backup"`
	PostBackup  []string            `json:"post_backup"`
	PreRestore  []string            `json:"pre_restore"`
	PostRestore []string            `json:"post_restore"`
	OnStart     []string            `json:"on_start"`
	OnExit      []string            `json:"on_exit"`
	CustomHooks map[string][]string `json:"custom_hooks"`
	HookTimeout time.Duration       `json:"hook_timeout"`
	HookRetries int                 `json:"hook_retries"`
	FailOnHook  bool                `json:"fail_on_hook"`
}

// ConfigFormats 支持的配置文件格式
var SupportedConfigFormats = []string{
	".json", ".jsonc", ".json5",
	".yaml", ".yml",
	".toml",
	".ini", ".cfg", ".conf",
	".properties", ".prop",
	".env",
	".xml",
	".hcl", ".tf",
}

// AdvancedConfigManager 高级配置管理器
type AdvancedConfigManager struct {
	config     *AdvancedConfig
	formatters map[string]ConfigFormatter
	validators []ConfigValidator
}

// ConfigFormatter 配置文件格式化器接口
type ConfigFormatter interface {
	Parse(content []byte) (*AdvancedConfig, error)
	Format(config *AdvancedConfig) ([]byte, error)
	Validate(content []byte) error
}

// ConfigValidator 配置验证器接口
type ConfigValidator interface {
	Validate(config *AdvancedConfig) error
	GetValidationErrors() []string
}

// NewAdvancedConfigManager 创建高级配置管理器
func NewAdvancedConfigManager() *AdvancedConfigManager {
	return &AdvancedConfigManager{
		config:     NewDefaultAdvancedConfig(),
		formatters: make(map[string]ConfigFormatter),
		validators: make([]ConfigValidator, 0),
	}
}

// NewDefaultAdvancedConfig 创建默认高级配置
func NewDefaultAdvancedConfig() *AdvancedConfig {
	baseConfig := &Config{}
	baseConfig.setDefaults()

	return &AdvancedConfig{
		Config: baseConfig,
		Performance: PerformanceConfig{
			ShowProgress:  true,
			BatchSize:     100,
			Parallel:      false,
			MaxWorkers:    4,
			Timeout:       30 * time.Second,
			EagerMode:     false,
			SmartCleanup:  true,
			MemoryLimit:   512, // 512MB
			CacheSize:     1000,
			OptimizeSpeed: false,
			LazyLoad:      true,
			PreloadData:   false,
		},
		UI: UIConfig{
			ColorOutput:    true,
			ShowStats:      true,
			Notifications:  false,
			ProgressStyle:  "bar",
			Theme:          "default",
			IconSet:        "unicode",
			AnimatedOutput: false,
			DetailLevel:    "normal",
			ConfirmStyle:   "interactive",
			ErrorDisplay:   "detailed",
			SuccessDisplay: "summary",
			CompactMode:    false,
			FullScreen:     false,
			AutoResize:     true,
		},
		Backup: BackupConfig{
			AutoBackup:        false,
			BackupDir:         "",
			CompressionLevel:  6,
			VerifyIntegrity:   true,
			PreserveTimes:     true,
			BackupFormat:      "zip",
			MaxBackups:        10,
			BackupRotation:    "size",
			EncryptBackups:    false,
			BackupSchedule:    "",
			IncrementalBackup: false,
			RemoteBackup:      false,
			BackupLocation:    "local",
		},
		Security: SecurityConfig{
			SecureDelete:      false,
			OverwritePasses:   3,
			SecureRandom:      true,
			VirusScan:         false,
			HashVerification:  true,
			DigitalSignature:  false,
			AccessControl:     true,
			AuditTrail:        true,
			EncryptionEnabled: false,
			KeyManagement:     "local",
			TrustedSources:    []string{},
			BlacklistedPaths:  []string{},
			WhitelistedUsers:  []string{},
			SecurityLevel:     "normal",
		},
		Integration: IntegrationConfig{
			HooksEnabled:       false,
			CustomScript:       "",
			WebhookURL:         "",
			NotificationApps:   []string{},
			CloudSync:          false,
			DatabaseLog:        false,
			APIEndpoints:       make(map[string]string),
			PluginDirectory:    "",
			ExternalCommands:   make(map[string]string),
			EnvironmentVars:    make(map[string]string),
			ServiceIntegration: false,
			MessageQueues:      []string{},
		},
		Filters: FilterConfig{
			SkipHidden:         false,
			FileSizeLimit:      10 * 1024 * 1024 * 1024, // 10GB
			IncludePattern:     "",
			ExcludePattern:     "",
			RegexMode:          false,
			CaseSensitive:      false,
			FollowSymlinks:     false,
			FileTypeFilters:    []string{},
			AgeFilter:          0,
			SizeFilter:         "",
			NameFilter:         "",
			ContentFilter:      "",
			PermissionFilter:   "",
			OwnerFilter:        "",
			ModifiedFilter:     "",
			AccessedFilter:     "",
			CreatedFilter:      "",
			ConflictResolution: "ask",
			CustomFilters:      []string{},
		},
		Logging: LoggingConfig{
			LogFormat:       "text",
			LogLevel:        LogLevelInfoStr,
			LogFile:         "",
			LogRotation:     false,
			LogMaxSize:      100 * 1024 * 1024, // 100MB
			LogMaxFiles:     5,
			LogCompression:  false,
			RemoteLogging:   false,
			StructuredLogs:  false,
			DebugMode:       false,
			TraceEnabled:    false,
			MetricsEnabled:  false,
			PerformanceLogs: false,
		},
		Hooks: HooksConfig{
			PreDelete:   []string{},
			PostDelete:  []string{},
			OnError:     []string{},
			OnSuccess:   []string{},
			PreBackup:   []string{},
			PostBackup:  []string{},
			PreRestore:  []string{},
			PostRestore: []string{},
			OnStart:     []string{},
			OnExit:      []string{},
			CustomHooks: make(map[string][]string),
			HookTimeout: 10 * time.Second,
			HookRetries: 2,
			FailOnHook:  false,
		},
	}
}

// LoadAdvancedConfig 加载高级配置
func (acm *AdvancedConfigManager) LoadAdvancedConfig(paths ...string) error {
	var lastErr error

	for _, path := range paths {
		if path == "" {
			continue
		}

		content, err := os.ReadFile(path)
		if err != nil {
			lastErr = err
			continue
		}

		ext := strings.ToLower(filepath.Ext(path))
		formatter, exists := acm.formatters[ext]
		if !exists {
			// 尝试自动检测格式
			formatter = acm.detectFormat(content)
			if formatter == nil {
				lastErr = fmt.Errorf("不支持的配置文件格式: %s", ext)
				continue
			}
		}

		config, err := formatter.Parse(content)
		if err != nil {
			lastErr = err
			continue
		}

		// 验证配置
		for _, validator := range acm.validators {
			if err := validator.Validate(config); err != nil {
				lastErr = err
				continue
			}
		}

		acm.config = config
		return nil
	}

	if lastErr != nil {
		return fmt.Errorf("无法加载配置文件: %v", lastErr)
	}

	return fmt.Errorf("未找到有效的配置文件")
}

// detectFormat 自动检测配置文件格式
func (acm *AdvancedConfigManager) detectFormat(content []byte) ConfigFormatter {
	s := string(content)

	// JSON格式检测
	if (strings.HasPrefix(s, "{") && strings.HasSuffix(strings.TrimSpace(s), "}")) ||
		(strings.HasPrefix(s, "[") && strings.HasSuffix(strings.TrimSpace(s), "]")) {
		return &JSONFormatter{}
	}

	// YAML格式检测
	if strings.Contains(s, ":") && (strings.Contains(s, "\n") || strings.Contains(s, "---")) {
		yamlIndicators := []string{"---", "...", "- ", ": "}
		for _, indicator := range yamlIndicators {
			if strings.Contains(s, indicator) {
				return &YAMLFormatter{}
			}
		}
	}

	// INI格式检测
	if strings.Contains(s, "[") && strings.Contains(s, "]") && strings.Contains(s, "=") {
		return &INIFormatter{}
	}

	// Properties格式检测
	if strings.Contains(s, "=") && !strings.Contains(s, "{") {
		return &PropertiesFormatter{}
	}

	// 默认使用JSON格式
	return &JSONFormatter{}
}

// SaveAdvancedConfig 保存高级配置
func (acm *AdvancedConfigManager) SaveAdvancedConfig(path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	formatter, exists := acm.formatters[ext]
	if !exists {
		formatter = &JSONFormatter{} // 默认使用JSON格式
	}

	content, err := formatter.Format(acm.config)
	if err != nil {
		return fmt.Errorf("格式化配置失败: %v", err)
	}

	// 创建目录
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

// GetConfig 获取当前配置
func (acm *AdvancedConfigManager) GetConfig() *AdvancedConfig {
	return acm.config
}

// UpdateConfig 更新配置
func (acm *AdvancedConfigManager) UpdateConfig(updater func(*AdvancedConfig)) error {
	updater(acm.config)

	// 验证更新后的配置
	for _, validator := range acm.validators {
		if err := validator.Validate(acm.config); err != nil {
			return err
		}
	}

	return nil
}

// RegisterFormatter 注册格式化器
func (acm *AdvancedConfigManager) RegisterFormatter(ext string, formatter ConfigFormatter) {
	acm.formatters[ext] = formatter
}

// RegisterValidator 注册验证器
func (acm *AdvancedConfigManager) RegisterValidator(validator ConfigValidator) {
	acm.validators = append(acm.validators, validator)
}

// JSONFormatter JSON格式化器
type JSONFormatter struct{}

func (jf *JSONFormatter) Parse(content []byte) (*AdvancedConfig, error) {
	// 去除注释
	cleaned := stripJSONComments(string(content))

	var config AdvancedConfig
	if err := json.Unmarshal([]byte(cleaned), &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (jf *JSONFormatter) Format(config *AdvancedConfig) ([]byte, error) {
	return json.MarshalIndent(config, "", "  ")
}

func (jf *JSONFormatter) Validate(content []byte) error {
	cleaned := stripJSONComments(string(content))
	var temp interface{}
	return json.Unmarshal([]byte(cleaned), &temp)
}

// 其他格式化器占位符（可以根据需要实现）
type YAMLFormatter struct{}
type TOMLFormatter struct{}
type INIFormatter struct{}
type PropertiesFormatter struct{}

// Parse 解析YAML格式的配置文件
// 当前版本暂不支持YAML格式，建议使用JSON格式
func (yf *YAMLFormatter) Parse(content []byte) (*AdvancedConfig, error) {
	return nil, fmt.Errorf("YAML格式暂未实现，请使用JSON格式")
}

// Format 将配置对象格式化为YAML格式
// 当前版本暂不支持YAML格式，建议使用JSON格式
func (yf *YAMLFormatter) Format(config *AdvancedConfig) ([]byte, error) {
	return nil, fmt.Errorf("YAML格式化暂未实现，请使用JSON格式")
}

// Validate 验证YAML格式的配置内容
// 当前版本暂不支持YAML格式，建议使用JSON格式
func (yf *YAMLFormatter) Validate(content []byte) error {
	return fmt.Errorf("YAML验证暂未实现，请使用JSON格式")
}

func (tf *TOMLFormatter) Parse(content []byte) (*AdvancedConfig, error) {
	return nil, fmt.Errorf("TOML格式暂未实现，请使用JSON格式")
}

func (tf *TOMLFormatter) Format(config *AdvancedConfig) ([]byte, error) {
	return nil, fmt.Errorf("TOML格式化暂未实现，请使用JSON格式")
}

func (tf *TOMLFormatter) Validate(content []byte) error {
	return fmt.Errorf("TOML验证暂未实现，请使用JSON格式")
}

func (inf *INIFormatter) Parse(content []byte) (*AdvancedConfig, error) {
	return nil, fmt.Errorf("INI格式暂未实现，请使用JSON格式")
}

func (inf *INIFormatter) Format(config *AdvancedConfig) ([]byte, error) {
	return nil, fmt.Errorf("INI格式化暂未实现，请使用JSON格式")
}

func (inf *INIFormatter) Validate(content []byte) error {
	return fmt.Errorf("INI验证暂未实现，请使用JSON格式")
}

func (pf *PropertiesFormatter) Parse(content []byte) (*AdvancedConfig, error) {
	return nil, fmt.Errorf("Properties格式暂未实现，请使用JSON格式")
}

func (pf *PropertiesFormatter) Format(config *AdvancedConfig) ([]byte, error) {
	return nil, fmt.Errorf("Properties格式化暂未实现，请使用JSON格式")
}

func (pf *PropertiesFormatter) Validate(content []byte) error {
	return fmt.Errorf("Properties验证暂未实现，请使用JSON格式")
}
