package main

import (
	"fmt"
)

// HelpSystem 增强的帮助系统
type HelpSystem struct {
	mode        CommandMode
	feedbackMgr *FeedbackManager
	currentLang string
	verboseHelp bool
}

// NewHelpSystem 创建帮助系统
func NewHelpSystem(mode CommandMode, feedbackMgr *FeedbackManager) *HelpSystem {
	return &HelpSystem{
		mode:        mode,
		feedbackMgr: feedbackMgr,
		currentLang: currentLocale,
		verboseHelp: false,
	}
}

// ShowHelp 显示帮助信息
func (hs *HelpSystem) ShowHelp() {
	hs.showHeader()
	hs.showUsage()
	hs.showOptions()
	hs.showExamples()

	if hs.verboseHelp {
		hs.showAdvancedOptions()
		hs.showConfiguration()
		hs.showTroubleshooting()
	}

	hs.showFooter()
}

// showHeader 显示标题头部
func (hs *HelpSystem) showHeader() {
	modeName := hs.getModeDisplayName()
	fmt.Printf("%s v%s - %s\n\n", modeName, Version, T("跨平台安全删除工具"))
}

// getModeDisplayName 获取模式显示名称
func (hs *HelpSystem) getModeDisplayName() string {
	switch hs.mode {
	case ModeCP:
		return "cp"
	case ModeDel:
		return "del"
	case ModeRM:
		return "rm"
	default:
		return "DelGuard"
	}
}

// showUsage 显示用法信息
func (hs *HelpSystem) showUsage() {
	switch hs.mode {
	case ModeCP:
		fmt.Printf("%s:\n", T("用法"))
		fmt.Printf("  cp [%s] <%s> <%s>\n", T("选项"), T("源文件"), T("目标文件"))
		fmt.Printf("  cp [%s] <%s...> <%s>\n\n", T("选项"), T("源文件"), T("目标目录"))
	case ModeDel:
		fmt.Printf("%s:\n", T("用法"))
		fmt.Printf("  del [%s] <%s...>\n\n", T("选项"), T("文件或目录"))
	case ModeRM:
		fmt.Printf("%s:\n", T("用法"))
		fmt.Printf("  rm [%s] <%s...>\n\n", T("选项"), T("文件或目录"))
	default:
		fmt.Printf("%s:\n", T("用法"))
		fmt.Printf("  delguard [%s] <%s...>\n", T("选项"), T("文件或目录"))
		fmt.Printf("  delguard restore [%s] [%s]\n\n", T("选项"), T("模式"))
	}
}

// showOptions 显示选项信息
func (hs *HelpSystem) showOptions() {
	fmt.Printf("%s:\n", T("选项"))

	// 基本选项（所有模式都支持）
	basicOptions := [][]string{
		{"-r, --recursive", T("递归处理目录")},
		{"-f, --force", T("强制执行，跳过确认")},
		{"-i, --interactive", T("交互模式，逐个确认")},
		{"-v, --verbose", T("详细输出")},
		{"-q, --quiet", T("静默模式")},
		{"-n, --dry-run", T("预览模式，不实际执行")},
		{"-h, --help", T("显示帮助信息")},
		{"--version", T("显示版本信息")},
	}

	for _, opt := range basicOptions {
		fmt.Printf("  %-20s %s\n", opt[0], opt[1])
	}

	// 根据模式显示特定选项
	switch hs.mode {
	case ModeCP:
		hs.showCopyOptions()
	case ModeDel, ModeRM:
		hs.showDeleteOptions()
	default:
		hs.showDelGuardOptions()
	}

	fmt.Println()
}

// showCopyOptions 显示复制模式特有选项
func (hs *HelpSystem) showCopyOptions() {
	fmt.Printf("\n%s:\n", T("复制选项"))
	copyOptions := [][]string{
		{"-p, --preserve", T("保留文件属性")},
		{"-a, --archive", T("归档模式（等同于 -rp）")},
		{"-u, --update", T("仅复制较新的文件")},
		{"--no-clobber", T("不覆盖现有文件")},
	}

	for _, opt := range copyOptions {
		fmt.Printf("  %-20s %s\n", opt[0], opt[1])
	}
}

// showDeleteOptions 显示删除模式特有选项
func (hs *HelpSystem) showDeleteOptions() {
	fmt.Printf("\n%s:\n", T("删除选项"))
	deleteOptions := [][]string{
		{"--smart-search", T("启用智能搜索")},
		{"--search-content", T("搜索文件内容")},
		{"--search-parent", T("搜索父目录")},
		{"--similarity N", T("设置相似度阈值 (0-100)")},
		{"--max-results N", T("限制搜索结果数量")},
	}

	for _, opt := range deleteOptions {
		fmt.Printf("  %-20s %s\n", opt[0], opt[1])
	}
}

// showDelGuardOptions 显示DelGuard模式特有选项
func (hs *HelpSystem) showDelGuardOptions() {
	fmt.Printf("\n%s:\n", T("高级选项"))
	advancedOptions := [][]string{
		{"--validate-only", T("仅验证文件，不执行删除")},
		{"--safe-copy", T("启用安全复制模式")},
		{"--timeout DURATION", T("设置操作超时时间")},
		{"--smart-search", T("启用智能搜索")},
		{"--search-content", T("搜索文件内容")},
		{"--search-parent", T("搜索父目录")},
		{"--similarity N", T("设置相似度阈值 (0-100)")},
		{"--max-results N", T("限制搜索结果数量")},
		{"--install", T("安装系统别名")},
		{"--uninstall", T("卸载系统别名")},
		{"--lang LANG", T("设置语言")},
		{"--config FILE", T("指定配置文件")},
		{"--security-scan", T("执行文件安全检查（识别潜在风险文件）")},
		{"--security-check", T("执行系统安全检查（全面安全评估）")},
	}

	for _, opt := range advancedOptions {
		fmt.Printf("  %-20s %s\n", opt[0], opt[1])
	}

	fmt.Printf("\n%s:\n", T("恢复选项"))
	restoreOptions := [][]string{
		{"restore", T("恢复已删除的文件")},
		{"restore --list", T("列出可恢复的文件")},
		{"restore --all", T("恢复所有文件")},
		{"restore PATTERN", T("恢复匹配模式的文件")},
	}

	for _, opt := range restoreOptions {
		fmt.Printf("  %-20s %s\n", opt[0], opt[1])
	}
}

// showExamples 显示使用示例
func (hs *HelpSystem) showExamples() {
	fmt.Printf("%s:\n", T("使用示例"))

	switch hs.mode {
	case ModeCP:
		examples := [][]string{
			{"cp file.txt backup.txt", T("复制文件")},
			{"cp -r folder/ backup/", T("递归复制目录")},
			{"cp -i *.txt backup/", T("交互式复制多个文件")},
			{"cp -p file.txt backup.txt", T("保留文件属性复制")},
		}
		for _, ex := range examples {
			fmt.Printf("  %-30s # %s\n", ex[0], ex[1])
		}
	case ModeDel:
		examples := [][]string{
			{"del file.txt", T("删除文件")},
			{"del -r folder", T("递归删除目录")},
			{"del -i *.tmp", T("交互式删除临时文件")},
			{"del --smart-search myfile", T("智能搜索并删除")},
		}
		for _, ex := range examples {
			fmt.Printf("  %-30s # %s\n", ex[0], ex[1])
		}
	case ModeRM:
		examples := [][]string{
			{"rm file.txt", T("删除文件")},
			{"rm -rf folder", T("强制递归删除目录")},
			{"rm -i *.log", T("交互式删除日志文件")},
			{"rm --smart-search oldfile", T("智能搜索并删除")},
		}
		for _, ex := range examples {
			fmt.Printf("  %-30s # %s\n", ex[0], ex[1])
		}
	default:
		examples := [][]string{
			{"delguard file.txt", T("安全删除文件")},
			{"delguard -r folder", T("递归删除目录")},
			{"delguard --smart-search myfile", T("智能搜索并删除")},
			{"delguard search pattern", T("独立搜索文件")},
			{"delguard restore", T("恢复已删除的文件")},
			{"delguard restore --list", T("列出可恢复的文件")},
			{"delguard --install", T("安装系统别名")},
			{"delguard --security-scan file.exe", T("检查文件安全风险（可执行文件等）")},
			{"delguard --security-scan /path/to/dir", T("检查目录中的潜在风险文件")},
		}
		for _, ex := range examples {
			fmt.Printf("  %-30s # %s\n", ex[0], ex[1])
		}
	}
	fmt.Println()
}

// showAdvancedOptions 显示高级选项（详细模式）
func (hs *HelpSystem) showAdvancedOptions() {
	fmt.Printf("%s:\n", T("高级配置选项"))

	performanceOptions := [][]string{
		{"--batch-size N", T("批处理大小")},
		{"--max-workers N", T("最大工作线程数")},
		{"--parallel", T("启用并行处理")},
		{"--show-progress", T("显示进度条")},
		{"--auto-backup", T("自动备份")},
		{"--backup-dir DIR", T("指定备份目录")},
		{"--compression-level N", T("压缩级别 (0-9)")},
		{"--verify-integrity", T("验证文件完整性")},
		{"--secure-delete", T("安全删除（多次覆写）")},
		{"--show-stats", T("显示统计信息")},
		{"--color-output", T("彩色输出")},
		{"--log-format FORMAT", T("日志格式")},
		{"--notifications", T("启用通知")},
		{"--preserve-times", T("保留时间戳")},
		{"--skip-hidden", T("跳过隐藏文件")},
		{"--file-size-limit SIZE", T("文件大小限制")},
		{"--include-pattern PATTERN", T("包含模式")},
		{"--exclude-pattern PATTERN", T("排除模式")},
		{"--regex-mode", T("正则表达式模式")},
		{"--case-sensitive", T("大小写敏感")},
		{"--follow-symlinks", T("跟随符号链接")},
		{"--eager-mode", T("积极模式")},
		{"--smart-cleanup", T("智能清理")},
		{"--conflict-resolution MODE", T("冲突解决模式")},
		{"--file-type-filters TYPE", T("文件类型过滤器")},
		{"--age-filter AGE", T("文件年龄过滤器")},
		{"--size-filter SIZE", T("文件大小过滤器")},
		{"--custom-script SCRIPT", T("自定义脚本")},
		{"--hooks-enabled", T("启用钩子")},
	}

	for _, opt := range performanceOptions {
		fmt.Printf("  %-25s %s\n", opt[0], opt[1])
	}
	fmt.Println()
}

// showConfiguration 显示配置信息
func (hs *HelpSystem) showConfiguration() {
	fmt.Printf("%s:\n", T("配置文件"))
	fmt.Printf("  %s: ~/.delguard/config.yaml\n", T("默认配置文件"))
	fmt.Printf("  %s: DELGUARD_CONFIG\n", T("环境变量"))
	fmt.Printf("  %s: --config FILE\n", T("命令行指定"))
	fmt.Println()
}

// showTroubleshooting 显示故障排除信息
func (hs *HelpSystem) showTroubleshooting() {
	fmt.Printf("%s:\n", T("故障排除"))
	fmt.Printf("  %s:\n", T("常见问题"))
	fmt.Printf("    • %s\n", T("权限不足：以管理员身份运行"))
	fmt.Printf("    • %s\n", T("文件被占用：关闭相关程序"))
	fmt.Printf("    • %s\n", T("路径不存在：检查路径拼写"))
	fmt.Printf("    • %s\n", T("回收站已满：清空回收站"))
	fmt.Printf("  %s:\n", T("获取帮助"))
	fmt.Printf("    • %s: https://github.com/user/delguard/issues\n", T("问题反馈"))
	fmt.Printf("    • %s: https://github.com/user/delguard/wiki\n", T("文档"))
	fmt.Println()
}

// showFooter 显示页脚信息
func (hs *HelpSystem) showFooter() {
	fmt.Printf("%s:\n", T("注意事项"))
	fmt.Printf("  • %s\n", T("DelGuard 会将文件移动到系统回收站，不会直接删除"))
	fmt.Printf("  • %s\n", T("cp 命令会安全处理文件覆盖，将原文件移入回收站"))
	fmt.Printf("  • %s\n", T("使用 --dry-run 可以预览操作而不实际执行"))
	fmt.Printf("  • %s\n", T("使用 restore 命令可以恢复已删除的文件"))
	fmt.Println()
}

// getModeDescription 获取模式描述
func (hs *HelpSystem) getModeDescription() string {
	switch hs.mode {
	case ModeCP:
		return T("安全复制模式 - 复制文件时自动备份目标文件")
	case ModeDel:
		return T("Windows风格删除模式 - 兼容del命令语法")
	case ModeRM:
		return T("Unix风格删除模式 - 兼容rm命令语法")
	default:
		return T("DelGuard默认模式 - 提供完整的安全删除功能")
	}
}
