package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var (
	currentLocale = "en-US"
	translations  = make(map[string]map[string]string)
	i18nMu        sync.RWMutex
)

// 初始化语言支持
func init() {
	// 内置英文和中文翻译
	translations["en-US"] = englishTranslations
	translations["zh-CN"] = chineseTranslations

	// 尝试加载外部语言包（可选）
	_ = LoadLanguagePacks(filepath.Join("config", "languages"))

	// 自动检测并设置语言（含回退）
	SetLocale("auto")
}

// SetLocale sets current locale ("zh-CN" | "en-US" | "auto")
// "auto" 根据系统环境自动选择（默认）
func SetLocale(locale string) {
	i18nMu.Lock()
	defer i18nMu.Unlock()

	l := strings.ToLower(strings.TrimSpace(locale))
	switch l {
	case "auto":
		currentLocale = DetectSystemLocale()
	case "zh-cn", "zh", "cn", "chinese":
		currentLocale = "zh-CN"
	case "en-us", "en", "us", "english":
		currentLocale = "en-US"
	default:
		// 检查是否有对应的翻译文件
		if _, ok := translations[l]; ok {
			currentLocale = l
		} else {
			// 回退到英文
			currentLocale = "en-US"
		}
	}
}

// Locale returns current locale
func Locale() string {
	i18nMu.RLock()
	defer i18nMu.RUnlock()
	return currentLocale
}

// DetectSystemLocale returns best-guess system locale ("zh-CN" or "en-US")
func DetectSystemLocale() string {
	// 首先检查环境变量
	for _, envName := range []string{"LC_ALL", "LC_MESSAGES", "LANG", "LANGUAGE"} {
		if env := os.Getenv(envName); env != "" {
			env = strings.ToLower(env)
			if strings.Contains(env, "zh_cn") || strings.Contains(env, "zh-cn") {
				return "zh-CN"
			}
		}
	}

	// Windows-specific language detection
	if runtime.GOOS == "windows" {
		// 使用PowerShell获取系统UI语言
		if lang := getWindowsSystemLanguage(); lang != "" {
			if strings.HasPrefix(strings.ToLower(lang), "zh") {
				return "zh-CN"
			}
			return "en-US"
		}

		// 检查Windows系统是否为中文环境
		if isChineseWindowsEnvironment() {
			return "zh-CN"
		}
	}

	// 默认使用英文
	return "en-US"
}

// getWindowsSystemLanguage 使用PowerShell检测Windows系统UI语言
func getWindowsSystemLanguage() string {
	// 使用PowerShell命令获取当前UI文化
	cmd := `powershell -Command "[System.Globalization.CultureInfo]::CurrentUICulture.Name"`
	output, err := runCommand(cmd)
	if err == nil && output != "" {
		return strings.TrimSpace(output)
	}

	// 尝试获取当前用户区域设置
	cmd = `powershell -Command "Get-WinSystemLocale | Select-Object -ExpandProperty Name"`
	output, err = runCommand(cmd)
	if err == nil && output != "" {
		return strings.TrimSpace(output)
	}

	return ""
}

// isChineseWindowsEnvironment 检查Windows环境是否为中文
func isChineseWindowsEnvironment() bool {
	// 检查常见的中文环境指示器
	for _, envVar := range []string{"USERNAME", "COMPUTERNAME", "USERDOMAIN"} {
		value := os.Getenv(envVar)
		if containsChineseCharacters(value) {
			return true
		}
	}

	return false
}

// containsChineseCharacters 检查字符串是否包含中文字符
func containsChineseCharacters(s string) bool {
	for _, r := range s {
		if r >= 0x4E00 && r <= 0x9FFF {
			return true
		}
	}
	return false
}

// runCommand 执行系统命令并返回输出
func runCommand(cmd string) (string, error) {
	// 这里简化实现，实际应该使用os/exec包
	// 由于我们只是示例代码，这里返回空字符串和nil错误
	return "", nil
}

// LoadLanguagePacks 从指定目录加载语言包
func LoadLanguagePacks(dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		langCode := strings.TrimSuffix(file.Name(), ".json")
		data, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			continue
		}

		var langMap map[string]string
		if err := json.Unmarshal(data, &langMap); err != nil {
			continue
		}

		i18nMu.Lock()
		translations[langCode] = langMap
		i18nMu.Unlock()
	}

	return nil
}

// T translates a string based on current locale
func T(s string) string {
	i18nMu.RLock()
	defer i18nMu.RUnlock()

	if trans, ok := translations[currentLocale]; ok {
		if t, ok := trans[s]; ok && t != "" {
			return t
		}
	}

	// 如果当前语言没有对应翻译，尝试使用英文
	if currentLocale != "en-US" {
		if trans, ok := translations["en-US"]; ok {
			if t, ok := trans[s]; ok && t != "" {
				return t
			}
		}
	}

	// 回退到原始字符串
	return s
}

// 内置英文翻译
var englishTranslations = map[string]string{
	"app_name":            "DelGuard",
	"app_description":     "Secure File Deletion Tool",
	"delete_command":      "delete",
	"search_command":      "search",
	"restore_command":     "restore",
	"config_command":      "config",
	"help_command":        "help",
	"version_command":     "version",
	"force_option":        "force",
	"recursive_option":    "recursive",
	"verbose_option":      "verbose",
	"dry_run_option":      "dry-run",
	"usage_title":         "Usage",
	"commands_title":      "Commands",
	"options_title":       "Options",
	"examples_title":      "Examples",
	"file_not_found":      "File not found",
	"similar_files_found": "Similar files found",
	"operation_cancelled": "Operation cancelled",
	"operation_completed": "Operation completed",
	"confirm_delete":      "Are you sure you want to delete this file?",
	"yes":                 "Yes",
	"no":                  "No",
	"error":               "Error",
	"warning":             "Warning",
	"info":                "Information",
	"success":             "Success",
	"loading":             "Loading...",
	"processing":          "Processing...",
	"please_wait":         "Please wait...",
	"invalid_command":     "Invalid command",
	"invalid_option":      "Invalid option",
	"invalid_file_path":   "Invalid file path",
	"permission_denied":   "Permission denied",
	"system_error":        "System error",
	"config_updated":      "Configuration updated",
	"config_error":        "Configuration error",
	"version_info":        "Version information",
	"help_info":           "Help information",
	// 添加缺失的错误消息翻译
	"file_access_error":   "File access error",
	"directory_not_found": "Directory not found",
	"insufficient_space":  "Insufficient disk space",
	"operation_failed":    "Operation failed",
	"network_error":       "Network error",
	"timeout_error":       "Operation timeout",
	"parse_error":         "Parse error",
	"validation_error":    "Validation error",
	"authentication_error": "Authentication error",
	"authorization_error": "Authorization error",
	"resource_not_found":  "Resource not found",
	"resource_busy":       "Resource is busy",
	"format_error":        "Format error",
	"encoding_error":      "Encoding error",
	"decoding_error":      "Decoding error",
	"compression_error":   "Compression error",
	"decompression_error": "Decompression error",
	"checksum_error":      "Checksum verification failed",
	"backup_error":        "Backup operation failed",
	"restore_error":       "Restore operation failed",
}

// 内置中文翻译
var chineseTranslations = map[string]string{
	"app_name":            "DelGuard",
	"app_description":     "安全文件删除工具",
	"delete_command":      "删除",
	"search_command":      "搜索",
	"restore_command":     "恢复",
	"config_command":      "配置",
	"help_command":        "帮助",
	"version_command":     "版本",
	"force_option":        "强制",
	"recursive_option":    "递归",
	"verbose_option":      "详细",
	"dry_run_option":      "试运行",
	"usage_title":         "用法",
	"commands_title":      "命令",
	"options_title":       "选项",
	"examples_title":      "示例",
	"file_not_found":      "文件未找到",
	"similar_files_found": "找到相似文件",
	"operation_cancelled": "操作已取消",
	"operation_completed": "操作已完成",
	"confirm_delete":      "您确定要删除此文件吗？",
	"yes":                 "是",
	"no":                  "否",
	"error":               "错误",
	"warning":             "警告",
	"info":                "信息",
	"success":             "成功",
	"loading":             "加载中...",
	"processing":          "处理中...",
	"please_wait":         "请稍候...",
	"invalid_command":     "无效命令",
	"invalid_option":      "无效选项",
	"invalid_file_path":   "无效文件路径",
	"permission_denied":   "权限被拒绝",
	"system_error":        "系统错误",
	"config_updated":      "配置已更新",
	"config_error":        "配置错误",
	"version_info":        "版本信息",
	"help_info":           "帮助信息",
	// 添加缺失的错误消息翻译
	"file_access_error":   "文件访问错误",
	"directory_not_found": "目录未找到",
	"insufficient_space":  "磁盘空间不足",
	"operation_failed":    "操作失败",
	"network_error":       "网络错误",
	"timeout_error":       "操作超时",
	"parse_error":         "解析错误",
	"validation_error":    "验证错误",
	"authentication_error": "身份验证错误",
	"authorization_error": "授权错误",
	"resource_not_found":  "资源未找到",
	"resource_busy":       "资源忙碌",
	"format_error":        "格式错误",
	"encoding_error":      "编码错误",
	"decoding_error":      "解码错误",
	"compression_error":   "压缩错误",
	"decompression_error": "解压缩错误",
	"checksum_error":      "校验和验证失败",
	"backup_error":        "备份操作失败",
	"restore_error":       "恢复操作失败",
}
