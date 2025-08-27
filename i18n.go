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
		"无法访问回收站: %w":         "ごみ箱にアクセスできません: %w",
		"没有找到匹配的文件: %s":       "一致するファイルが見つかりません: %s",
		"找到 %d 个匹配文件:\n":      "%d 個の一致ファイルが見つかりました:\n",
		"确认恢复这些文件吗? (y/N): ":  "これらのファイルを復元しますか？ (y/N): ",
		"用户取消操作":              "ユーザーによって操作がキャンセルされました",
		"恢复路径验证失败 %s: %v\n":   "復元パスの検証に失敗しました %s: %v\n",
		"恢复文件失败 %s: %v\n":     "ファイルの復元に失敗しました %s: %v\n",
		"成功恢复: %s -> %s\n":    "復元成功: %s -> %s\n",
		"所有文件恢复失败":            "すべてのファイルの復元に失敗しました",
		"成功恢复 %d 个文件\n":       "%d 個のファイルを正常に復元しました\n",
		"信息文件格式无效":            "情報ファイルの形式が無効です",
		"文本文件":                "テキストファイル",
		"Word文档":              "Word文書",
		"Excel表格":             "Excelシート",
		"PDF文档":               "PDF文書",
		"图片文件":                "画像ファイル",
		"视频文件":                "ビデオファイル",
		"音频文件":                "オーディオファイル",
		"其他文件":                "その他のファイル",
		"警告：即将删除关键路径: %s\n":   "警告: 重要なパスを削除しようとしています: %s\n",
		"为确认风险，请输入完整路径继续（或直接回车取消）：":     "リスクを確認するため、続行する場合は完全なパスを入力してください（Enterでキャンセル）:",
		"警告：当前以管理员/root权限运行，即将删除: %s\n": "警告: 管理者/root権限で実行しています。削除しようとしています: %s\n",
		"确认删除？[y/N]: ":                 "削除を確認しますか？[y/N]:",
		"警告：文件 %s 为只读文件\n":             "警告: ファイル %s は読み取り専用です\n",
		"确认删除只读文件？[y/N]: ":             "読み取り専用ファイルを削除しますか？[y/N]:",
		"警告：检测到回收站/废纸篓目录: %s\n":        "警告: ゴミ箱/ごみ箱ディレクトリが検出されました: %s\n",
		"确认删除回收站目录？[y/N]: ":            "ゴミ箱ディレクトリを削除しますか？[y/N]:",
		"警告：检测到DelGuard程序目录: %s\n":     "警告: DelGuardプログラムディレクトリが検出されました: %s\n",
		"确认删除程序目录？[y/N]: ":             "プログラムディレクトリを削除しますか？[y/N]:",
		"错误：权限不足，无法删除 %s\n":            "エラー: 権限が不足しているため %s を削除できません\n",
		"错误：系统保护文件，无法删除 %s\n":          "エラー: システム保護ファイルのため %s を削除できません\n",
		"错误：路径包含非法字符: %s\n":            "エラー: パスに不正な文字が含まれています: %s\n",
		"错误：路径过长: %s\n":                "エラー: パスが長すぎます: %s\n",
		"错误：磁盘空间不足，无法删除 %s\n":          "エラー: ディスク容量が不足しているため %s を削除できません\n",
		"错误：文件正在被使用: %s\n":             "エラー: ファイルが使用中です: %s\n",
		"错误：网络路径不可访问: %s\n":            "エラー: ネットワークパスにアクセスできません: %s\n",
		"错误：符号链接目标不存在: %s\n":           "エラー: シンボリックリンクのターゲットが存在しません: %s\n",
		"错误：硬链接计数异常: %s\n":             "エラー: ハードリンクカウントが異常です: %s\n",
		"错误：文件系统只读: %s\n":              "エラー: ファイルシステムが読み取り専用です: %s\n",
		"错误：磁盘错误: %s\n":                "エラー: ディスクエラー: %s\n",
		"错误：内存不足: %s\n":                "エラー: メモリが不足しています: %s\n",
		"错误：操作超时: %s\n":                "エラー: 操作がタイムアウトしました: %s\n",
		"错误：系统调用失败: %s\n":              "エラー: システムコールが失敗しました: %s\n",
		"错误：未知错误: %s\n":                "エラー: 不明なエラー: %s\n",
		"通配符模式编译失败: %s":                "ワイルドカードパターンのコンパイルに失敗しました: %s",
		"正则表达式编译失败: %s":                "正規表現のコンパイルに失敗しました: %s",
		"⚠️  未找到文件 '%s'，正在进行智能搜索...\n": "⚠️  ファイル '%s' が見つからないため、スマート検索を実行しています...\n",
		"智能搜索失败: %v":                   "スマート検索に失敗しました: %v",
		"🔍 未找到文件名匹配，正在搜索文件内容...\n":     "🔍 ファイル名が一致しないため、ファイルコンテンツを検索しています...\n",
		"未找到与 '%s' 匹配的文件":              "'%s' に一致するファイルが見つかりません",
		"🔍 自动选择高相似度文件: %s (%.1f%%)\n":  "🔍 高類似度のファイルを自動選択: %s (%.1f%%)\n",
		"目录":              "ディレクトリ",
		"大文件 (%.1fGB)":    "大容量ファイル (%.1fGB)",
		"文件 (%.1fMB)":     "ファイル (%.1fMB)",
		"文件 (%.1fKB)":     "ファイル (%.1fKB)",
		"文件 (%d字节)":       "ファイル (%dバイト)",
		"子目录":             "サブディレクトリ",
		"DelGuard 安全检查工具": "DelGuard セキュリティチェックツール",
		"用法:":             "使用方法:",
		"  delguard --security-check  执行安全检查": "  delguard --security-check  セキュリティチェックを実行",
		"  delguard --help           显示帮助":    "  delguard --help           ヘルプを表示",
		"未知选项: %s\n":                          "未知のオプション: %s\n",
		"使用 --help 查看用法信息":                    "--helpを使用して使用方法を確認してください",
		"❌ 缺少必要的安全文件:":                        "❌ 必要なセキュリティファイルが不足しています:",
		"✅ 所有必要的安全文件已存在":                      "✅ すべての必要なセキュリティファイルが存在します",
		"❌ 配置校验失败: %v\n":                      "❌ 設定の検証に失敗しました: %v\n",
		"✅ 配置校验通过":                            "✅ 設定の検証に成功しました",
		"❌ 加载配置失败: %v\n":                      "❌ 設定の読み込みに失敗しました: %v\n",
		"✅ 关键路径保护正常":                          "✅ 重要なパスの保護が正常です",
		"❌ 关键路径保护可能未正常工作":                     "❌ 重要なパスの保護が正常に動作していない可能性があります",
		"=== 环境变量检查 ===":                      "=== 環境変数チェック ===",
		"✅ %s: %s\n":                          "✅ %s: %s\n",
		"⚠️  %s 未设置\n":                        "⚠️  %s が設定されていません\n",
		"=== 临时文件安全检查 ===":                    "=== 一時ファイルのセキュリティチェック ===",
		"⚠️  临时目录 %s 具有全局写权限\n":               "⚠️ 一時ディレクトリ %s にグローバルな書き込み権限があります\n",
		"✅ 临时目录 %s 权限安全\n":                    "✅ 一時ディレクトリ %s の権限は安全です\n",
		"=== 文件安全检查 ===":                      "=== ファイルのセキュリティチェック ===",
	},
	"en-US": {
		"safe_copy_skip_same": "Skipping copy as source and destination are identical: %s",
		"safe_copy_confirm":   "Destination file exists and differs. Overwrite? [y/N] ",
		"safe_copy_cancelled": "Safe copy cancelled: %s",
		"safe_copy_backup":    "Moved existing file %s to trash",
		"safe_copy_success":   "Safe copy completed: %s -> %s",
		"safe_copy_failed":    "Safe copy failed: %s -> %s: %v",
		"无法访问回收站: %w":         "Cannot access recycle bin: %w",
		"没有找到匹配的文件: %s":       "No matching files found: %s",
		"找到 %d 个匹配文件:\n":      "Found %d matching files:\n",
		"确认恢复这些文件吗? (y/N): ":  "Confirm recovery of these files? (y/N): ",
		"用户取消操作":              "Operation cancelled by user",
		"恢复路径验证失败 %s: %v\n":   "Recovery path validation failed %s: %v\n",
		"恢复文件失败 %s: %v\n":     "Failed to recover file %s: %v\n",
		"成功恢复: %s -> %s\n":    "Successfully recovered: %s -> %s\n",
		"所有文件恢复失败":            "All file recoveries failed",
		"成功恢复 %d 个文件\n":       "Successfully recovered %d files\n",
		"信息文件格式无效":            "Invalid information file format",
		"文本文件":                "Text file",
		"Word文档":              "Word document",
		"Excel表格":             "Excel spreadsheet",
		"PDF文档":               "PDF document",
		"图片文件":                "Image file",
		"视频文件":                "Video file",
		"音频文件":                "Audio file",
		"其他文件":                "Other file",
		"警告：即将删除关键路径: %s\n":   "Warning: About to delete critical path: %s\n",
		"为确认风险，请输入完整路径继续（或直接回车取消）：":     "To confirm risk, enter full path to continue (or press Enter to cancel):",
		"警告：当前以管理员/root权限运行，即将删除: %s\n": "Warning: Running as admin/root, about to delete: %s\n",
		"确认删除？[y/N]: ":                 "Confirm deletion? [y/N]: ",
		"警告：文件 %s 为只读文件\n":             "Warning: File %s is read-only\n",
		"确认删除只读文件？[y/N]: ":             "Confirm deletion of read-only file? [y/N]: ",
		"警告：检测到回收站/废纸篓目录: %s\n":        "Warning: Detected recycle bin directory: %s\n",
		"确认删除回收站目录？[y/N]: ":            "Confirm deletion of recycle bin directory? [y/N]: ",
		"警告：检测到DelGuard程序目录: %s\n":     "Warning: Detected DelGuard program directory: %s\n",
		"确认删除程序目录？[y/N]: ":             "Confirm deletion of program directory? [y/N]: ",
		"错误：权限不足，无法删除 %s\n":            "Error: Insufficient permissions to delete %s\n",
		"错误：系统保护文件，无法删除 %s\n":          "Error: System protected file, cannot delete %s\n",
		"错误：路径包含非法字符: %s\n":            "Error: Path contains invalid characters: %s\n",
		"错误：路径过长: %s\n":                "Error: Path too long: %s\n",
		"错误：磁盘空间不足，无法删除 %s\n":          "Error: Insufficient disk space to delete %s\n",
		"错误：文件正在被使用: %s\n":             "Error: File is in use: %s\n",
		"错误：网络路径不可访问: %s\n":            "Error: Network path not accessible: %s\n",
		"错误：符号链接目标不存在: %s\n":           "Error: Symbolic link target does not exist: %s\n",
		"错误：硬链接计数异常: %s\n":             "Error: Hard link count anomaly: %s\n",
		"错误：文件系统只读: %s\n":              "Error: File system is read-only: %s\n",
		"错误：磁盘错误: %s\n":                "Error: Disk error: %s\n",
		"错误：内存不足: %s\n":                "Error: Insufficient memory: %s\n",
		"错误：操作超时: %s\n":                "Error: Operation timeout: %s\n",
		"错误：系统调用失败: %s\n":              "Error: System call failed: %s\n",
		"错误：未知错误: %s\n":                "Error: Unknown error: %s\n",
		"通配符模式编译失败: %s":                "Wildcard pattern compilation failed: %s",
		"正则表达式编译失败: %s":                "Regular expression compilation failed: %s",
		"⚠️  未找到文件 '%s'，正在进行智能搜索...\n": "⚠️  File '%s' not found, performing smart search...\n",
		"智能搜索失败: %v":                   "Smart search failed: %v",
		"🔍 未找到文件名匹配，正在搜索文件内容...\n":     "🔍 No filename match found, searching file contents...\n",
		"未找到与 '%s' 匹配的文件":              "No files found matching '%s'",
		"🔍 自动选择高相似度文件: %s (%.1f%%)\n":  "🔍 Auto-selected high similarity file: %s (%.1f%%)\n",
		"目录":              "Directory",
		"大文件 (%.1fGB)":    "Large file (%.1fGB)",
		"文件 (%.1fMB)":     "File (%.1fMB)",
		"文件 (%.1fKB)":     "File (%.1fKB)",
		"文件 (%d字节)":       "File (%d bytes)",
		"子目录":             "Subdirectory",
		"DelGuard 安全检查工具": "DelGuard Security Check Tool",
		"用法:":             "Usage:",
		"  delguard --security-check  执行安全检查": "  delguard --security-check  Run security check",
		"  delguard --help           显示帮助":    "  delguard --help           Show help",
		"未知选项: %s\n":                          "Unknown option: %s\n",
		"使用 --help 查看用法信息":                    "Use --help to see usage information",
		"❌ 缺少必要的安全文件:":                        "❌ Missing required security files:",
		"✅ 所有必要的安全文件已存在":                      "✅ All required security files exist",
		"❌ 配置校验失败: %v\n":                      "❌ Configuration validation failed: %v\n",
		"✅ 配置校验通过":                            "✅ Configuration validation passed",
		"❌ 加载配置失败: %v\n":                      "❌ Failed to load configuration: %v\n",
		"✅ 关键路径保护正常":                          "✅ Critical path protection is normal",
		"❌ 关键路径保护可能未正常工作":                     "❌ Critical path protection may not be working properly",
		"=== 环境变量检查 ===":                      "=== Environment variable check ===",
		"✅ %s: %s\n":                          "✅ %s: %s\n",
		"⚠️  %s 未设置\n":                        "⚠️  %s is not set\n",
		"=== 临时文件安全检查 ===":                    "=== Temporary file security check ===",
		"⚠️  临时目录 %s 具有全局写权限\n":               "⚠️  Temporary directory %s has global write permissions\n",
		"✅ 临时目录 %s 权限安全\n":                    "✅ Temporary directory %s permissions are secure\n",
		"=== 文件安全检查 ===":                      "=== File security check ===",
	},
}
