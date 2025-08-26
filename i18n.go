package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var (
	currentLocale = "en-US"
	i18nMu        sync.RWMutex
)

// 初始化：加载语言包并自动设置语言
func init() {
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
	case "auto", "":
		detected := DetectSystemLocale()
		// 如果未提供对应语言包，则回退到 en-US（zh-CN 为源语言，不需要包）
		if detected != "zh-CN" && !hasLanguage(detected) {
			currentLocale = "en-US"
		} else {
			currentLocale = detected
		}
	case "zh", "zh-cn", "zh_cn", "zh-hans":
		currentLocale = "zh-CN"
	case "en", "en-us", "en_us":
		currentLocale = "en-US"
	default:
		// 优先使用外部或内置语言包
		norm := normalizeLangCode(l)
		if norm == "zh-CN" || hasLanguage(norm) {
			currentLocale = norm
		} else {
			// fallback en-US
			currentLocale = "en-US"
		}
	}
}

// hasLanguage 检查是否存在语言包
func hasLanguage(code string) bool {
	_, ok := translations[code]
	return ok
}

// normalizeLangCode 规范化语言代码
func normalizeLangCode(code string) string {
	c := strings.TrimSpace(code)
	c = strings.ReplaceAll(c, "_", "-")
	c = strings.ToLower(c)
	switch c {
	case "zh", "zh-cn", "zh-hans":
		return "zh-CN"
	case "en", "en-us":
		return "en-US"
	case "ja", "ja-jp":
		return "ja"
	default:
		// 首字母区域部分大写，例如 en-gb -> en-GB
		parts := strings.Split(c, "-")
		if len(parts) == 2 {
			return strings.ToLower(parts[0]) + "-" + strings.ToUpper(parts[1])
		}
		return c
	}
}

// guessLocaleFromRaw 根据原始字符串猜测语言（用于环境变量/系统输出）
func guessLocaleFromRaw(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	if strings.Contains(v, "zh") {
		return "zh-CN"
	}
	if strings.Contains(v, "ja") {
		return "ja"
	}
	if strings.Contains(v, "en") {
		return "en-US"
	}
	// 通用模式：如 fr-FR, de-DE 等，若有外部包则使用
	// 尝试提取前两段
	v = strings.ReplaceAll(v, "_", "-")
	parts := strings.Split(v, ".") // 去除编码如 UTF-8
	base := parts[0]
	if base != "" {
		norm := normalizeLangCode(base)
		if hasLanguage(norm) {
			return norm
		}
	}
	return "en-US"
}

// shouldPrefix 判断是否需要添加前缀（避免多行或已包含前缀的内容）
func shouldPrefix(s string) bool {
	if s == "" {
		return false
	}
	if strings.HasPrefix(s, "DelGuard:") {
		return false
	}
	if strings.Contains(s, "\n") {
		return false
	}
	return true
}

// LoadLanguagePacks 从目录加载外部语言包（外部覆盖内置）
// 支持文件命名：<lang>.(json|jsonc|ini|cfg|conf|env|properties)
// 统一语义：均解析为 map[string]string，其中 key 为中文原文，value 为目标语言译文
func LoadLanguagePacks(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		// 目录不存在不视为错误
		return nil
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := strings.ToLower(filepath.Ext(name))
		switch ext {
		case ".json", ".jsonc", ".ini", ".cfg", ".conf", ".env", ".properties":
			// supported
		default:
			continue
		}
		lang := strings.TrimSuffix(name, filepath.Ext(name))
		lang = normalizeLangCode(lang)
		// 读取文件
		content, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			fmt.Fprintf(os.Stderr, "LoadLanguagePacks: 读取失败 %s: %v\n", name, err)
			continue
		}
		var mp map[string]string
		switch ext {
		case ".json":
			if err := json.Unmarshal(content, &mp); err != nil {
				fmt.Fprintf(os.Stderr, "LoadLanguagePacks: JSON解析失败 %s: %v\n", name, err)
				continue
			}
		case ".jsonc":
			cleaned := stripJSONComments(string(content))
			if err := json.Unmarshal([]byte(cleaned), &mp); err != nil {
				fmt.Fprintf(os.Stderr, "LoadLanguagePacks: JSONC解析失败 %s: %v\n", name, err)
				continue
			}
		case ".ini", ".cfg", ".conf":
			mp = parseINIToMap(string(content))
		case ".env", ".properties":
			mp = parseKVToMap(string(content))
		default:
			continue
		}
		if translations[lang] == nil {
			translations[lang] = map[string]string{}
		}
		// 合并（外部覆盖内置）
		for k, v := range mp {
			translations[lang][k] = v
		}
	}
	return nil
}

// parseINIToMap 将 INI/CFG/CONF 文本解析为 map[string]string
// 支持 key=value 或 key: value，忽略 [section]，# 和 ; 注释
func parseINIToMap(s string) map[string]string {
	out := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		raw := scanner.Text()
		if strings.TrimSpace(raw) == "" {
			continue
		}
		// 保留行尾空格，仅用于注释判断做左侧裁剪
		ltrim := strings.TrimLeft(raw, " \t")
		if strings.HasPrefix(ltrim, "#") || strings.HasPrefix(ltrim, ";") {
			continue
		}
		if strings.HasPrefix(ltrim, "[") && strings.HasSuffix(ltrim, "]") {
			// section ignored
			continue
		}
		// allow inline comments preceded by # or ; if not in quotes
		line := stripInlineComment(raw)
		var key, val string
		if idx := strings.Index(line, "="); idx >= 0 {
			// 优先使用 '=' 作为分隔符
			key = strings.TrimLeft(line[:idx], " \t")
			val = strings.TrimLeft(line[idx+1:], " \t")
		} else if idx := strings.Index(line, ":"); idx >= 0 {
			// 回退支持 ':'
			key = strings.TrimLeft(line[:idx], " \t")
			val = strings.TrimLeft(line[idx+1:], " \t")
		} else {
			continue
		}
		val = trimQuotes(val)
		if strings.TrimSpace(key) != "" {
			out[key] = val
		}
	}
	return out
}

// parseKVToMap 解析 .env/.properties 样式的 key=value
func parseKVToMap(s string) map[string]string {
	out := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		raw := scanner.Text()
		if strings.TrimSpace(raw) == "" {
			continue
		}
		ltrim := strings.TrimLeft(raw, " \t")
		if strings.HasPrefix(ltrim, "#") || strings.HasPrefix(ltrim, ";") {
			continue
		}
		line := stripInlineComment(raw)
		var key, val string
		if idx := strings.Index(line, "="); idx >= 0 {
			key = strings.TrimLeft(line[:idx], " \t")
			val = strings.TrimLeft(line[idx+1:], " \t")
		} else if idx := strings.Index(line, ":"); idx >= 0 { // tolerate ':'
			key = strings.TrimLeft(line[:idx], " \t")
			val = strings.TrimLeft(line[idx+1:], " \t")
		} else {
			continue
		}
		val = trimQuotes(val)
		if strings.TrimSpace(key) != "" {
			out[key] = val
		}
	}
	return out
}

// stripInlineComment 移除行尾注释（# 或 ;），不处理引号内
func stripInlineComment(line string) string {
	inStr := false
	esc := false
	var b strings.Builder
	for i := 0; i < len(line); i++ {
		c := line[i]
		if inStr {
			b.WriteByte(c)
			if esc {
				esc = false
				continue
			}
			if c == '\\' {
				esc = true
			} else if c == '"' {
				inStr = false
			}
			continue
		}
		if c == '"' {
			inStr = true
			b.WriteByte(c)
			continue
		}
		if c == '#' || c == ';' {
			break
		}
		b.WriteByte(c)
	}
	// 保留行尾空格，去除可能的前导空格
	return strings.TrimLeft(b.String(), " \t")
}

func trimQuotes(v string) string {
	if len(v) >= 2 && ((strings.HasPrefix(v, "\"") && strings.HasSuffix(v, "\"")) || (strings.HasPrefix(v, "'") && strings.HasSuffix(v, "'"))) {
		return v[1 : len(v)-1]
	}
	return v
}

// Locale returns current locale
func Locale() string {
	i18nMu.RLock()
	defer i18nMu.RUnlock()
	return currentLocale
}

// DetectSystemLocale returns best-guess system locale ("zh-CN" or "en-US")
func DetectSystemLocale() string {
	// Prefer standard envs
	for _, k := range []string{"LANGUAGE", "LC_ALL", "LANG"} {
		if v := os.Getenv(k); v != "" {
			return guessLocaleFromRaw(v)
		}
	}

	// Windows-specific language detection
	if runtime.GOOS == "windows" {
		// Check Windows UI language environment variables
		for _, k := range []string{"LANG", "LC_CTYPE", "LC_MESSAGES"} {
			if v := os.Getenv(k); v != "" {
				return guessLocaleFromRaw(v)
			}
		}

		// Check common Windows language environment variables
		if v := os.Getenv("USERPROFILE"); v != "" {
			// Check if user profile path contains Chinese characters
			if containsChinese(v) {
				return "zh-CN"
			}
		}

		// Check Windows system language via registry or environment variables
		// Use PowerShell to get the actual system UI language
		if lang := getWindowsSystemLanguage(); lang != "" {
			return lang
		}

		// Fallback: check if running in Chinese locale environment
		if isChineseWindowsEnvironment() {
			return "zh-CN"
		}

		// Default to English if cannot determine
		return "en-US"
	}

	// Unix-like systems
	// Windows often lacks LANG; try UI language via environment variables commonly set by shells
	for _, k := range []string{"MSYSTEM_CHOST", "MSYSTEM"} {
		if v := os.Getenv(k); v != "" {
			if strings.Contains(strings.ToLower(v), "mingw") {
				// cannot infer, default en-US
				return "en-US"
			}
		}
	}
	// Default
	return "en-US"
}

// getWindowsSystemLanguage uses PowerShell to detect the actual Windows system UI language
func getWindowsSystemLanguage() string {
	cmd := exec.Command("powershell", "-Command", "Get-Culture | Select-Object -ExpandProperty Name")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lang := strings.TrimSpace(string(output))
	return guessLocaleFromRaw(lang)
}

// isChineseWindowsEnvironment checks various indicators of Chinese Windows environment
func isChineseWindowsEnvironment() bool {
	// Check system directory names which often contain language indicators
	systemRoot := os.Getenv("SystemRoot")
	if systemRoot != "" && containsChinese(systemRoot) {
		return true
	}

	// Check user profile directory
	userProfile := os.Getenv("USERPROFILE")
	if userProfile != "" && containsChinese(userProfile) {
		return true
	}

	// Check program files directory
	programFiles := os.Getenv("ProgramFiles")
	if programFiles != "" && containsChinese(programFiles) {
		return true
	}

	// Check common Windows environment variables that might indicate Chinese locale
	for _, envVar := range []string{"USERNAME", "COMPUTERNAME", "USERDOMAIN"} {
		if value := os.Getenv(envVar); value != "" && containsChinese(value) {
			return true
		}
	}

	return false
}

// containsChinese checks if a string contains Chinese characters
func containsChinese(s string) bool {
	for _, r := range s {
		if r >= 0x4E00 && r <= 0x9FFF {
			return true
		}
	}
	return false
}
// Convention: Source strings in code are zh-CN; for en-US we map to English.
// If no mapping exists, returns s itself.
func T(s string) string {
    i18nMu.RLock()
    defer i18nMu.RUnlock()
    // 翻译
    var translated string
	if currentLocale == "zh-CN" {
		translated = s
	} else if m, ok := translations[currentLocale]; ok {
		if tr, ok2 := m[s]; ok2 {
			translated = tr
		} else {
			// 语言包缺失键时回退到英文
			if em, ok3 := translations["en-US"]; ok3 {
				if tr2, ok4 := em[s]; ok4 {
					translated = tr2
				} else {
					translated = s
				}
			} else {
				translated = s
			}
		}
	} else if m, ok := translations["en-US"]; ok {
		if tr, ok2 := m[s]; ok2 {
			translated = tr
		} else {
			translated = s
		}
	} else {
		translated = s
	}

    // 直接返回翻译结果，不添加任何前缀
    return translated
}

var translations = map[string]map[string]string{
    "ja": {
        "safe_copy_skip_same": "ソースファイルと宛先ファイルは同一のため、コピーをスキップします：%s",
        "safe_copy_confirm":   "宛先ファイルが存在し、内容が異なります。上書きしますか？[y/N] ",
        "safe_copy_cancelled": "セーフコピーがキャンセルされました：%s",
        "safe_copy_backup":    "既存のファイル %s をゴミ箱に移動しました",
        "safe_copy_success":   "セーフコピーが完了しました：%s -> %s",
        "safe_copy_failed":    "セーフコピーに失敗しました：%s -> %s: %v",
    },
    "en-US": {
        "safe_copy_skip_same": "Skipping copy as source and destination are identical: %s",
        "safe_copy_confirm":   "Destination file exists and differs. Overwrite? [y/N] ",
        "safe_copy_cancelled": "Safe copy cancelled: %s",
        "safe_copy_backup":    "Moved existing file %s to Trash",
        "safe_copy_success":   "Safe copy completed: %s -> %s",
        "safe_copy_failed":    "Safe copy failed: %s -> %s: %v",
        "参数解析失败: %v\n":        "Failed to parse arguments: %v\n",
        "别名安装成功！请重启终端或运行 'source ~/.bashrc' 使别名生效。": "Aliases installed. Restart your terminal or run 'source ~/.bashrc' to take effect.",
        "卸载功能尚未实现":                           "Uninstall is not implemented yet",
        "恢复文件失败: %v\n":                       "Failed to restore files: %v\n",
        "错误：无法解析路径 %s: %v\n":                 "Error: cannot resolve path %s: %v\n",
        "错误：无法访问 %s: %v\n":                   "Error: cannot access %s: %v\n",
        "提示：%s 是目录，删除目录需使用 -r/--recursive\n": "Tip: %s is a directory; use -r/--recursive to delete directories\n",
        "即将删除 %d 个目标（其中目录 %d 个）。选择模式 [a]全部同意/[n]全部拒绝/[i]逐项/[q]退出 (默认 i): ": "About to delete %d target(s) (%d directorie(s)). Choose [a] accept all / [n] reject all / [i] item-by-item / [q] quit (default i): ",
        "已取消所有删除。":                     "All deletions cancelled.",
        "计划删除：%s (绝对路径: %s, 类型: %s)\n": "Plan to delete: %s (abs: %s, type: %s)\n",
        "删除 %s ? [y/N/a/q]: ":          "Delete %s ? [y/N/a/q]: ",
        "已跳过 %s\n":                     "Skipped %s\n",
        "检测到关键路径，要求双重确认：%s\n":          "Critical path detected, double confirmation required: %s\n",
        "已取消关键路径 %s 的删除\n":             "Deletion of critical path %s cancelled\n",
        "[DRY-RUN] 将把 %s 移动到回收站\n":     "[DRY-RUN] Would move %s to Trash\n",
        "错误：无法删除 %s: %v\n":             "Error: failed to delete %s: %v\n",
        "已将 %s 移动到回收站\n":               "Moved %s to Trash\n",
        "警告：当前以高权限运行（root/管理员），已强制启用交互确认。\n": "Warning: running with elevated privileges (root/Administrator). Interactive confirmation enforced.\n",
        "当前平台 %s 暂不支持回收站功能\n":                "Platform %s does not support trash functionality\n",
        // 新增的安全警告和确认消息
        "警告：即将删除关键路径: %s\n":             "Warning: about to delete critical path: %s\n",
        "为确认风险，请输入完整路径继续（或直接回车取消）：":     "To confirm the risk, enter the full path to continue (or press Enter to cancel): ",
        "警告：当前以管理员/root权限运行，即将删除: %s\n": "Warning: running as administrator/root, about to delete: %s\n",
        "确认删除？[y/N]: ":                  "Confirm deletion? [y/N]: ",
        "警告：文件 %s 为只读文件\n":              "Warning: file %s is read-only\n",
        "确认删除只读文件？[y/N]: ":              "Confirm deletion of read-only file? [y/N]: ",
        "警告：检测到回收站/废纸篓目录: %s\n":         "Warning: detected trash/recycle bin directory: %s\n",
        "确认删除回收站目录？[y/N]: ":             "Confirm deletion of trash directory? [y/N]: ",
        "警告：检测到DelGuard程序目录: %s\n":      "Warning: detected DelGuard program directory: %s\n",
        "确认删除程序目录？[y/N]: ":              "Confirm deletion of program directory? [y/N]: ",
        "错误：权限不足，无法删除 %s\n":             "Error: insufficient permissions to delete %s\n",
        "错误：系统保护文件，无法删除 %s\n":           "Error: system protected file, cannot delete %s\n",
        "错误：路径包含非法字符: %s\n":             "Error: path contains invalid characters: %s\n",
        "错误：路径过长: %s\n":                 "Error: path too long: %s\n",
        "错误：磁盘空间不足，无法删除 %s\n":           "Error: insufficient disk space to delete %s\n",
        "错误：文件正在被使用: %s\n":              "Error: file is in use: %s\n",
        "错误：网络路径不可访问: %s\n":             "Error: network path not accessible: %s\n",
        "错误：符号链接目标不存在: %s\n":            "Error: symlink target does not exist: %s\n",
        "错误：硬链接计数异常: %s\n":              "Error: hard link count异常: %s\n",
        "错误：文件系统只读: %s\n":               "Error: filesystem is read-only: %s\n",
        "错误：磁盘错误: %s\n":                 "Error: disk error: %s\n",
        "错误：内存不足: %s\n":                 "Error: out of memory: %s\n",
        "错误：操作超时: %s\n":                 "Error: operation timed out: %s\n",
        "错误：系统调用失败: %s\n":               "Error: system call failed: %s\n",
        "错误：未知错误: %s\n":                 "Error: unknown error: %s\n",
        // Usage block translation
        `DelGuard v%s - 跨平台安全删除工具

用法:
  delguard [选项] <文件或目录>

选项:
  -r, --recursive              递归删除目录
  -f, --force                  强制删除（忽略警告）
  -i, --interactive            删除前确认（可通过环境变量/配置默认启用）
  -y, --yes                    跳过确认，默认全部同意（关键路径仍需二次确认）
  -n, --dry-run                干跑：只显示将要执行的操作，不实际删除
      --verbose                输出详细信息
      --quiet                  安静模式：仅输出错误
      --lang value             设置语言 (auto/zh-CN/en-US)，默认随系统/配置/环境变量
      --config path            指定配置文件路径
      --install                安装shell别名（Windows: del/rm；Unix: rm/del）
      --uninstall              卸载shell别名
      --restore                从回收站恢复文件
      --default-interactive    安装时将 del/rm 默认指向交互删除（等同于 -i）
  -max int                     最大恢复文件数 (默认: 10)
  -v, --version                显示版本信息
  -h, --help                   显示此帮助信息

说明:
  - 支持短参组合，例如: -rf 等价于 -r -f
  - 交互删除优先级: CLI(-i) > 环境变量(DELGUARD_INTERACTIVE/DELGUARD_DEFAULT_INTERACTIVE) > 配置文件
  - 对关键路径（系统根、系统目录、用户主目录等）启用双重确认
  - 交互批处理模式：在多目标下可选择"全部同意/全部拒绝/逐项/退出"
  - Windows 回收站允许同名文件共存（系统自动处理）
  - Windows/CMD & PowerShell 支持 rm 与 del 命令；Linux/macOS 同时提供 del 命令（通过别名）
`: `DelGuard v%s - Cross-platform safe deletion

Usage:
  delguard [options] <file-or-directory>

Options:
  -r, --recursive              Recursively delete directories
  -f, --force                  Force deletion (suppress warnings)
  -i, --interactive            Confirm before deleting (can be default via env/config)
  -y, --yes                    Skip confirmations (critical paths still require double confirm)
  -n, --dry-run                Dry run: show actions without deleting
      --verbose                Verbose output
      --quiet                  Quiet mode: only errors
      --lang value             Set language (auto/zh-CN/en-US), defaults to system/config/env
      --config path            Specify config file path
      --install                Install shell aliases (Windows: del/rm; Unix: rm/del)
      --uninstall              Uninstall aliases
      --restore                Restore from Trash
      --default-interactive    Make del/rm default to interactive (-i) on install
  -max int                     Max files to restore (default: 10)
  -v, --version                Show version
  -h, --help                   Show this help

Notes:
  - Combined short flags supported, e.g. -rf == -r -f
  - Interactive priority: CLI(-i) > ENV(DELGUARD_INTERACTIVE/DELGUARD_DEFAULT_INTERACTIVE) > config
  - Double-confirm for critical paths (system root, system dirs, user home, etc.)
  - Batch interactive mode: accept all / reject all / item-by-item / quit when multiple targets
      - Windows Recycle Bin allows duplicate names (handled by system)
      - Windows CMD & PowerShell support rm and del; Linux/macOS also provide del via aliases
`,
	},
}
