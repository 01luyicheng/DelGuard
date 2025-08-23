package main

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

var (
	currentLocale = "en-US"
	i18nMu        sync.RWMutex
)

// SetLocale sets current locale ("zh-CN" | "en-US" | "auto")
// "auto" 根据系统环境自动选择（默认）
func SetLocale(locale string) {
	i18nMu.Lock()
	defer i18nMu.Unlock()
	l := strings.ToLower(strings.TrimSpace(locale))
	switch l {
	case "auto", "":
		currentLocale = DetectSystemLocale()
	case "zh", "zh-cn", "zh_cn", "zh-hans":
		currentLocale = "zh-CN"
	case "en", "en-us", "en_us":
		currentLocale = "en-US"
	default:
		// fallback en-US
		currentLocale = "en-US"
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
	// Prefer standard envs
	for _, k := range []string{"LANGUAGE", "LC_ALL", "LANG"} {
		if v := os.Getenv(k); v != "" {
			v = strings.ToLower(v)
			if strings.Contains(v, "zh") {
				return "zh-CN"
			}
			if strings.Contains(v, "en") {
				return "en-US"
			}
			// 如果环境变量值不是中文也不是英文，继续检查其他环境变量
			continue
		}
	}

	// Windows-specific language detection
	if runtime.GOOS == "windows" {
		// Check Windows UI language environment variables
		for _, k := range []string{"LANG", "LC_CTYPE", "LC_MESSAGES"} {
			if v := os.Getenv(k); v != "" {
				v = strings.ToLower(v)
				if strings.Contains(v, "zh") {
					return "zh-CN"
				}
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
	if strings.Contains(strings.ToLower(lang), "zh") {
		return "zh-CN"
	}
	if strings.Contains(strings.ToLower(lang), "en") {
		return "en-US"
	}
	return ""
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

// T translates a zh-CN source string to current locale.
// Convention: Source strings in code are zh-CN; for en-US we map to English.
// If no mapping exists, returns s itself.
func T(s string) string {
	i18nMu.RLock()
	defer i18nMu.RUnlock()
	if currentLocale == "zh-CN" {
		return s
	}
	if m, ok := translations["en-US"]; ok {
		if tr, ok2 := m[s]; ok2 {
			return tr
		}
	}
	return s
}

var translations = map[string]map[string]string{
	"en-US": {
		"参数解析失败: %v\n": "Failed to parse arguments: %v\n",
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
  - 交互批处理模式：在多目标下可选择“全部同意/全部拒绝/逐项/退出”
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

// 错误信息国际化映射
var errorMessages = map[string]map[string]string{
	"zh": {
		"file_not_found":           "文件不存在：%s",
		"permission_denied":        "权限不足：%s",
		"critical_path_warning":    "⚠️ 警告：您正在尝试删除系统关键路径 %s",
		"confirm_delete":           "确认删除 %s 吗？",
		"confirm_critical_delete":  "⚠️ 您确定要删除系统关键路径 %s 吗？此操作可能导致系统不稳定！",
		"delete_success":           "✅ 成功删除：%s",
		"delete_failed":            "❌ 删除失败：%s",
		"trash_success":            "🗑️ 已将 %s 移至回收站",
		"permanent_delete":         "⚠️ 已永久删除：%s",
		"disk_space_warning":       "⚠️ 磁盘空间不足，无法完成操作",
		"file_locked":              "文件被锁定：%s",
		"path_too_long":            "路径过长：%s",
		"invalid_characters":       "路径包含非法字符：%s",
		"network_path":             "不支持删除网络路径：%s",
		"hidden_file_warning":      "👁️ 这是一个隐藏文件：%s",
		"system_file_warning":      "⚙️ 这是一个系统文件：%s",
		"readonly_file_warning":    "🔒 这是一个只读文件：%s",
		"large_file_warning":       "📁 这是一个大文件（%s），确认删除吗？",
		"batch_delete_warning":     "⚠️ 您即将删除 %d 个文件，确认继续吗？",
		"recursive_delete_warning": "⚠️ 您即将递归删除整个目录 %s，确认继续吗？",
		"empty_trash_confirm":      "🗑️ 您确定要清空回收站吗？此操作不可恢复！",
		"restore_success":          "✅ 成功恢复：%s",
		"restore_failed":           "❌ 恢复失败：%s",
		"backup_created":           "💾 已创建备份：%s",
		"config_error":             "配置错误：%s",
		"memory_warning":           "⚠️ 内存使用过高，可能影响性能",
		"timeout_warning":          "⏰ 操作超时，已自动取消",
		"invalid_config":           "配置无效：%s",
		"config_restored":          "配置已从备份恢复",
		"config_saved":             "✅ 配置已保存",
		"invalid_language":         "不支持的语言设置：%s",
		"invalid_range":            "参数超出有效范围：%s",
		"backup_failed":            "备份失败：%s",
		"validation_error":         "验证失败：%s",
		"security_error":           "安全检查失败：%s",
		"path_validation_error":    "路径验证失败：%s",
		"symlink_error":            "符号链接检查失败：%s",
		"unc_path_error":           "UNC路径不受支持：%s",
		"device_path_error":        "设备路径不受支持：%s",
		"reserved_name_error":      "Windows保留设备名不被允许：%s",
	},
	"en": {
		"file_not_found":           "File not found: %s",
		"permission_denied":        "Permission denied: %s",
		"critical_path_warning":    "⚠️ Warning: You are attempting to delete a critical system path %s",
		"confirm_delete":           "Confirm deletion of %s?",
		"confirm_critical_delete":  "⚠️ Are you sure you want to delete the critical system path %s? This may cause system instability!",
		"delete_success":           "✅ Successfully deleted: %s",
		"delete_failed":            "❌ Failed to delete: %s",
		"trash_success":            "🗑️ Moved %s to trash",
		"permanent_delete":         "⚠️ Permanently deleted: %s",
		"disk_space_warning":       "⚠️ Insufficient disk space to complete operation",
		"file_locked":              "File is locked: %s",
		"path_too_long":            "Path too long: %s",
		"invalid_characters":       "Path contains invalid characters: %s",
		"network_path":             "Network paths are not supported: %s",
		"hidden_file_warning":      "👁️ This is a hidden file: %s",
		"system_file_warning":      "⚙️ This is a system file: %s",
		"readonly_file_warning":    "🔒 This is a read-only file: %s",
		"large_file_warning":       "📁 This is a large file (%s), confirm deletion?",
		"batch_delete_warning":     "⚠️ You are about to delete %d files, continue?",
		"recursive_delete_warning": "⚠️ You are about to recursively delete the entire directory %s, continue?",
		"empty_trash_confirm":      "🗑️ Are you sure you want to empty the trash? This cannot be undone!",
		"restore_success":          "✅ Successfully restored: %s",
		"restore_failed":           "❌ Failed to restore: %s",
		"backup_created":           "💾 Backup created: %s",
		"config_error":             "Configuration error: %s",
		"memory_warning":           "⚠️ High memory usage may affect performance",
		"timeout_warning":          "⏰ Operation timeout, automatically cancelled",
		"invalid_config":           "Invalid configuration: %s",
		"config_restored":          "Configuration restored from backup",
		"config_saved":             "✅ Configuration saved",
		"invalid_language":         "Unsupported language setting: %s",
		"invalid_range":            "Parameter out of valid range: %s",
		"backup_failed":            "Backup failed: %s",
		"validation_error":         "Validation failed: %s",
		"security_error":           "Security check failed: %s",
		"path_validation_error":    "Path validation failed: %s",
		"symlink_error":            "Symbolic link check failed: %s",
		"unc_path_error":           "UNC paths are not supported: %s",
		"device_path_error":        "Device paths are not supported: %s",
		"reserved_name_error":      "Windows reserved device names are not allowed: %s",
	},
	"ja": {
		"file_not_found":           "ファイルが見つかりません: %s",
		"permission_denied":        "アクセス権限がありません: %s",
		"critical_path_warning":    "⚠️ 警告: システムの重要なパス %s を削除しようとしています",
		"confirm_delete":           "%s を削除してもよろしいですか？",
		"confirm_critical_delete":  "⚠️ 本当にシステムの重要なパス %s を削除してもよろしいですか？システムが不安定になる可能性があります！",
		"delete_success":           "✅ 正常に削除されました: %s",
		"delete_failed":            "❌ 削除に失敗しました: %s",
		"trash_success":            "🗑️ %s をゴミ箱に移動しました",
		"permanent_delete":         "⚠️ 完全に削除されました: %s",
		"disk_space_warning":       "⚠️ ディスク容量不足のため、操作を完了できません",
		"file_locked":              "ファイルがロックされています: %s",
		"path_too_long":            "パスが長すぎます: %s",
		"invalid_characters":       "パスに無効な文字が含まれています: %s",
		"network_path":             "ネットワークパスはサポートされていません: %s",
		"hidden_file_warning":      "👁️ これは隠しファイルです: %s",
		"system_file_warning":      "⚙️ これはシステムファイルです: %s",
		"readonly_file_warning":    "🔒 これは読み取り尰端ファイルです: %s",
		"large_file_warning":       "📁 これは大きなファイルです（%s）、削除を確認しますか？",
		"batch_delete_warning":     "⚠️ %d 個のファイルを削除しようとしています、続行してもよろしいですか？",
		"recursive_delete_warning": "⚠️ ディレクトリ %s を再帰的に削除しようとしています、続行してもよろしいですか？",
		"empty_trash_confirm":      "🗑️ 本当にゴミ箱を空にしてもよろしいですか？この操作は元に戻せません！",
		"restore_success":          "✅ 正常に復元されました: %s",
		"restore_failed":           "❌ 復元に失敗しました: %s",
		"backup_created":           "💾 バックアップが作成されました: %s",
		"config_error":             "設定エラー: %s",
		"memory_warning":           "⚠️ メモリ使用量が高すぎるとパフォーマンスに影響する可能性があります",
		"timeout_warning":          "⏰ 操作タイムアウト、自動的にキャンセルされました",
		"invalid_config":           "無効な設定: %s",
		"config_restored":          "バックアップから設定が復元されました",
		"config_saved":             "✅ 設定が保存されました",
		"invalid_language":         "サポートされていない言語設定: %s",
		"invalid_range":            "パラメータが有効範囲外です: %s",
		"backup_failed":            "バックアップに失敗しました: %s",
		"validation_error":         "検証に失敗しました: %s",
		"security_error":           "セキュリティチェックに失敗しました: %s",
		"path_validation_error":    "パス検証に失敗しました: %s",
		"symlink_error":            "シンボリックリンクのチェックに失敗しました: %s",
		"unc_path_error":           "UNCパスはサポートされていません: %s",
		"device_path_error":        "デバイスパスはサポートされていません: %s",
		"reserved_name_error":      "Windows予約デバイス名は許可されていません: %s",
	},
}
