package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Config struct {
	// 默认交互删除（当 CLI 未显式指定 -i 时生效）
	DefaultInteractive bool `json:"defaultInteractive"`
	// 默认语言（例如 "zh-CN" / "en-US"）
	Language string `json:"language"`
	// 输出详细程度：quiet/normal/verbose
	Verbosity string `json:"verbosity"`
	// 是否启用安全模式（保护系统关键文件）
	SafeMode bool `json:"safeMode"`
	// 是否启用回收站功能（false时直接删除）
	UseTrash bool `json:"useTrash"`
	// 是否跳过只读文件确认
	SkipReadOnlyConfirm bool `json:"skipReadOnlyConfirm"`
	// 是否跳过隐藏文件确认
	SkipHiddenConfirm bool `json:"skipHiddenConfirm"`
	// 是否跳过系统文件确认
	SkipSystemConfirm bool `json:"skipSystemConfirm"`
	// 是否启用管理员权限警告
	AdminWarning bool `json:"adminWarning"`
	// 最大备份文件数量
	MaxBackupFiles int `json:"maxBackupFiles"`
	// 日志文件路径
	LogFile string `json:"logFile"`
	// 回收站最大容量（MB）
	TrashMaxSize int `json:"trashMaxSize"`
	// 自动清理回收站旧文件（天）
	TrashAutoCleanDays int `json:"trashAutoCleanDays"`
	// 启用文件操作日志
	EnableOperationLog bool `json:"enableOperationLog"`
	// 启用删除确认声音提示
	EnableSoundAlert bool `json:"enableSoundAlert"`
	// 颜色输出模式：auto/always/never
	ColorMode string `json:"colorMode"`
}

var overrideConfigPath = ""

// SetConfigOverride 通过 --config 指定配置文件路径
func SetConfigOverride(p string) {
	overrideConfigPath = strings.TrimSpace(p)
}

// GetDefaultInteractive 计算默认交互删除（不含 CLI -i），优先级：
// 环境变量 DELGUARD_INTERACTIVE / DELGUARD_DEFAULT_INTERACTIVE > 配置文件
func GetDefaultInteractive() bool {
	if v, ok := readEnvBool("DELGUARD_INTERACTIVE"); ok {
		return v
	}
	if v, ok := readEnvBool("DELGUARD_DEFAULT_INTERACTIVE"); ok {
		return v
	}
	cfg := LoadConfig()
	return cfg.DefaultInteractive
}

// ResolveLanguage 确定语言，优先级：CLI --lang > ENV DELGUARD_LANG > 配置文件 > 系统语言
func ResolveLanguage(cliLang string) string {
	if strings.TrimSpace(cliLang) != "" {
		return cliLang
	}
	if v := strings.TrimSpace(os.Getenv("DELGUARD_LANG")); v != "" {
		return v
	}
	cfg := LoadConfig()
	if strings.TrimSpace(cfg.Language) != "" {
		return cfg.Language
	}
	return DetectSystemLocale()
}

// ResolveVerbosity 根据 CLI 和配置/环境确定 verbose/quiet
// 返回 (verbose, quiet)
func ResolveVerbosity(cliVerbose, cliQuiet bool) (bool, bool) {
	if cliVerbose && cliQuiet {
		// CLI 同时设置时以 verbose 优先，quiet 关闭
		return true, false
	}
	if cliVerbose || cliQuiet {
		return cliVerbose, cliQuiet
	}
	// 环境变量
	if v := strings.TrimSpace(strings.ToLower(os.Getenv("DELGUARD_VERBOSITY"))); v != "" {
		switch v {
		case "verbose", "debug", "2":
			return true, false
		case "quiet", "silent", "0":
			return false, true
		case "normal", "info", "1":
			return false, false
		}
	}
	// 配置文件
	cfg := LoadConfig()
	switch strings.ToLower(strings.TrimSpace(cfg.Verbosity)) {
	case "verbose", "debug":
		return true, false
	case "quiet", "silent":
		return false, true
	default:
		return false, false
	}
}

func readEnvBool(key string) (bool, bool) {
	val, exists := os.LookupEnv(key)
	if !exists {
		return false, false
	}
	val = strings.TrimSpace(strings.ToLower(val))
	switch val {
	case "1", "true", "yes", "on":
		return true, true
	case "0", "false", "no", "off", "":
		return false, true
	default:
		// 非法值时，认为存在但为 false
		return false, true
	}
}

func LoadConfigPath() string {
	if overrideConfigPath != "" {
		return overrideConfigPath
	}
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return ""
		}
		return filepath.Join(appData, "DelGuard", "config.json")
	}

	xdg := os.Getenv("XDG_CONFIG_HOME")
	base := xdg
	if base == "" {
		home, _ := os.UserHomeDir()
		if home == "" {
			return ""
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "delguard", "config.json")
}

func LoadConfig() Config {
	cfg := Config{
		DefaultInteractive:  false,
		Language:            "",
		Verbosity:           "normal",
		SafeMode:            true,
		UseTrash:            true,
		SkipReadOnlyConfirm: false,
		SkipHiddenConfirm:   false,
		SkipSystemConfirm:   false,
		AdminWarning:        true,
		MaxBackupFiles:      10,
		LogFile:             "",
		TrashMaxSize:        1024,
		TrashAutoCleanDays:  30,
		EnableOperationLog:  false,
		EnableSoundAlert:    true,
		ColorMode:           "auto",
	}

	path := LoadConfigPath()
	if path == "" {
		return cfg
	}
	b, err := os.ReadFile(path)
	if err != nil || len(b) == 0 {
		return cfg
	}
	_ = json.Unmarshal(b, &cfg) // 解析失败则使用默认值
	return cfg
}

// SaveConfig 保存配置到文件
func SaveConfig(cfg Config) error {
	path := LoadConfigPath()
	if path == "" {
		return os.ErrInvalid
	}

	// 创建配置目录
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// GetSafeMode 获取安全模式设置
func GetSafeMode() bool {
	if v, ok := readEnvBool("DELGUARD_SAFE_MODE"); ok {
		return v
	}
	cfg := LoadConfig()
	return cfg.SafeMode
}

// GetUseTrash 获取是否使用回收站
func GetUseTrash() bool {
	if v, ok := readEnvBool("DELGUARD_USE_TRASH"); ok {
		return v // 环境变量为true时表示使用回收站
	}
	cfg := LoadConfig()
	return cfg.UseTrash
}

// ShouldSkipReadOnlyConfirm 是否跳过只读文件确认
func ShouldSkipReadOnlyConfirm() bool {
	if v, ok := readEnvBool("DELGUARD_SKIP_READONLY_CONFIRM"); ok {
		return v
	}
	cfg := LoadConfig()
	return cfg.SkipReadOnlyConfirm
}

// ShouldSkipHiddenConfirm 是否跳过隐藏文件确认
func ShouldSkipHiddenConfirm() bool {
	if v, ok := readEnvBool("DELGUARD_SKIP_HIDDEN_CONFIRM"); ok {
		return v
	}
	cfg := LoadConfig()
	return cfg.SkipHiddenConfirm
}

// ShouldSkipSystemConfirm 是否跳过系统文件确认
func ShouldSkipSystemConfirm() bool {
	if v, ok := readEnvBool("DELGUARD_SKIP_SYSTEM_CONFIRM"); ok {
		return v
	}
	cfg := LoadConfig()
	return cfg.SkipSystemConfirm
}

// ShouldShowAdminWarning 是否显示管理员权限警告
func ShouldShowAdminWarning() bool {
	if v, ok := readEnvBool("DELGUARD_ADMIN_WARNING"); ok {
		return v
	}
	cfg := LoadConfig()
	return cfg.AdminWarning
}

// GetTrashMaxSize 获取回收站最大容量（MB）
func GetTrashMaxSize() int {
	if val := os.Getenv("DELGUARD_TRASH_MAX_SIZE"); val != "" {
		if size, err := parseInt(val); err == nil && size > 0 {
			return size
		}
	}
	cfg := LoadConfig()
	return cfg.TrashMaxSize
}

// GetTrashAutoCleanDays 获取回收站自动清理天数
func GetTrashAutoCleanDays() int {
	if val := os.Getenv("DELGUARD_TRASH_AUTO_CLEAN_DAYS"); val != "" {
		if days, err := parseInt(val); err == nil && days > 0 {
			return days
		}
	}
	cfg := LoadConfig()
	return cfg.TrashAutoCleanDays
}

// ShouldEnableOperationLog 是否启用操作日志
func ShouldEnableOperationLog() bool {
	if v, ok := readEnvBool("DELGUARD_ENABLE_OPERATION_LOG"); ok {
		return v
	}
	cfg := LoadConfig()
	return cfg.EnableOperationLog
}

// ShouldEnableSoundAlert 是否启用声音提示
func ShouldEnableSoundAlert() bool {
	if v, ok := readEnvBool("DELGUARD_ENABLE_SOUND_ALERT"); ok {
		return v
	}
	cfg := LoadConfig()
	return cfg.EnableSoundAlert
}

// GetColorMode 获取颜色输出模式
func GetColorMode() string {
	if val := os.Getenv("DELGUARD_COLOR_MODE"); val != "" {
		val = strings.ToLower(strings.TrimSpace(val))
		if val == "always" || val == "auto" || val == "never" {
			return val
		}
	}
	cfg := LoadConfig()
	return cfg.ColorMode
}

// parseInt 解析整数字符串
func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

// GetInteractiveDefault 获取默认交互模式设置
func GetInteractiveDefault() bool {
	cfg := LoadConfig()
	return cfg.DefaultInteractive
}
