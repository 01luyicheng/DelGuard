package main

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// HelpSystem 增强的帮助系统
type HelpSystem struct {
	mode        CommandMode
	feedbackMgr *FeedbackManager
	currentLang string
	verboseHelp bool
	// showExamples bool // 移除重复字段
}

// NewHelpSystem 创建帮助系统
func NewHelpSystem(mode CommandMode, feedbackMgr *FeedbackManager) *HelpSystem {
	return &HelpSystem{
		mode:        mode,
		feedbackMgr: feedbackMgr,
		currentLang: currentLocale,
		verboseHelp: false,
		// showExamples: true, // 使用方法名而不是字段名
	}
}

// ShowHelp 显示帮助信息
func (hs *HelpSystem) ShowHelp() {
	hs.showHeader()
	hs.showUsage()
	hs.showOptions()

	// 直接调用showExamples方法而不是检查字段
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

	if hs.feedbackMgr.colorEnabled {
		fmt.Printf("%s%s%s %sv%s%s - %s%s%s\n\n",
			Colors.Bold, Colors.Blue, modeName, Colors.Reset,
			Colors.Cyan, version, Colors.Yellow, T("跨平台安全删除工具"), Colors.Reset)
	} else {
		fmt.Printf("%s v%s - %s\n\n", modeName, version, T("跨平台安全删除工具"))
	}

	// 显示当前模式说明
	modeDesc := hs.getModeDescription()
	if modeDesc != "" {
		fmt.Printf(T("当前模式: %s\n\n"), modeDesc)
	}
}

// getModeDisplayName 获取模式显示名称
func (hs *HelpSystem) getModeDisplayName() string {
	switch hs.mode {
	case ModeDel:
		return "DelGuard (del 模式)"
	case ModeRM:
		return "DelGuard (rm 模式)"
	case ModeCP:
		return "DelGuard (cp 模式)"
	default:
		return "DelGuard"
	}
}

// getModeDescription 获取模式描述
func (hs *HelpSystem) getModeDescription() string {
	switch hs.mode {
	case ModeDel:
		return "Windows风格删除命令，默认启用交互确认和智能搜索"
	case ModeRM:
		return "Unix风格删除命令，支持智能搜索和安全删除"
	case ModeCP:
		return "安全复制命令，支持文件覆盖保护和完整性验证"
	default:
		return "全功能安全删除工具，提供最佳用户体验"
	}
}

// showUsage 显示用法
func (hs *HelpSystem) showUsage() {
	fmt.Printf("%s%s%s\n", Colors.Bold, T("用法:"), Colors.Reset)

	switch hs.mode {
	case ModeDel:
		fmt.Printf(T("  del [选项] <文件或目录>\n"))
		fmt.Printf(T("  del [选项] <通配符模式>\n"))
	case ModeRM:
		fmt.Printf(T("  rm [选项] <文件或目录>\n"))
		fmt.Printf(T("  rm [选项] <通配符模式>\n"))
	case ModeCP:
		fmt.Printf(T("  cp [选项] <源文件> <目标文件>\n"))
		fmt.Printf(T("  cp [选项] <源文件> <目标目录>\n"))
		fmt.Printf(T("  cp [选项] <多个源文件> <目标目录>\n"))
	default:
		fmt.Printf(T("  delguard [选项] <文件或目录>\n"))
		fmt.Printf(T("  delguard [选项] <通配符模式>\n"))
		fmt.Printf(T("  delguard --install    # 安装系统别名\n"))
		fmt.Printf(T("  delguard --restore    # 从回收站恢复文件\n"))
	}
	fmt.Println()
}

// showOptions 显示选项
func (hs *HelpSystem) showOptions() {
	fmt.Printf("%s%s%s\n", Colors.Bold, T("选项:"), Colors.Reset)

	// 基础选项
	hs.showBasicOptions()

	// 模式特定选项
	switch hs.mode {
	case ModeDel:
		hs.showDelOptions()
	case ModeRM:
		hs.showRmOptions()
	case ModeCP:
		hs.showCpOptions()
	default:
		hs.showDelGuardOptions()
	}

	fmt.Println()
}

// showBasicOptions 显示基础选项
func (hs *HelpSystem) showBasicOptions() {
	options := [][]string{
		{"-h, --help", "显示此帮助信息"},
		{"-v, --version", "显示版本信息"},
		{"-q, --quiet", "安静模式：仅输出错误信息"},
		{"--verbose", "详细模式：输出详细操作信息"},
		{"-n, --dry-run", "试运行：显示将要执行的操作但不实际执行"},
		{"-i, --interactive", "交互模式：删除前确认"},
		{"-f, --force", "强制模式：忽略警告直接执行"},
		{"-y, --yes", "跳过确认：对所有询问默认回答'是'"},
	}

	hs.printOptions(options)
}

// showDelOptions 显示del模式选项
func (hs *HelpSystem) showDelOptions() {
	fmt.Printf("\n%s%s%s\n", Colors.Cyan, T("del模式特有选项:"), Colors.Reset)
	options := [][]string{
		{"-r, --recursive", "递归删除目录"},
		{"--smart-search", "启用智能搜索（默认开启）"},
		{"--search-content", "搜索文件内容"},
	}
	hs.printOptions(options)
}

// showRmOptions 显示rm模式选项
func (hs *HelpSystem) showRmOptions() {
	fmt.Printf("\n%s%s%s\n", Colors.Cyan, T("rm模式特有选项:"), Colors.Reset)
	options := [][]string{
		{"-r, -R, --recursive", "递归删除目录"},
		{"--smart-search", "启用智能搜索（默认开启）"},
		{"--preserve-root", "保护根目录（默认开启）"},
	}
	hs.printOptions(options)
}

// showCpOptions 显示cp模式选项
func (hs *HelpSystem) showCpOptions() {
	fmt.Printf("\n%s%s%s\n", Colors.Cyan, T("cp模式特有选项:"), Colors.Reset)
	options := [][]string{
		{"-r, --recursive", "递归复制目录"},
		{"-p, --preserve", "保持文件属性"},
		{"-u, --update", "仅复制更新的文件"},
		{"--safe-copy", "安全复制模式（默认开启）"},
		{"--verify-integrity", "验证文件完整性"},
		{"--protect", "启用文件覆盖保护"},
	}
	hs.printOptions(options)
}

// showDelGuardOptions 显示DelGuard完整选项
func (hs *HelpSystem) showDelGuardOptions() {
	fmt.Printf("\n%s%s%s\n", Colors.Cyan, T("DelGuard完整选项:"), Colors.Reset)

	// 智能删除选项
	smartOptions := [][]string{
		{"--smart-search", "启用智能搜索（推荐）"},
		{"--search-content", "搜索文件内容"},
		{"--search-parent", "搜索父目录"},
		{"--similarity <值>", "相似度阈值 (0.0-1.0)"},
		{"--max-results <数量>", "最大搜索结果数"},
	}
	fmt.Printf("\n  %s%s%s\n", Colors.Yellow, T("智能搜索:"), Colors.Reset)
	hs.printOptions(smartOptions)

	// 安全选项
	securityOptions := [][]string{
		{"--protect", "启用文件覆盖保护"},
		{"--secure-delete", "安全删除（多次覆写）"},
		{"--verify-integrity", "验证文件完整性"},
		{"--backup-dir <目录>", "指定备份目录"},
		{"--auto-backup", "自动备份重要文件"},
	}
	fmt.Printf("\n  %s%s%s\n", Colors.Yellow, T("安全选项:"), Colors.Reset)
	hs.printOptions(securityOptions)

	// 性能选项
	performanceOptions := [][]string{
		{"--parallel", "启用并行处理"},
		{"--max-workers <数量>", "最大工作线程数"},
		{"--batch-size <大小>", "批处理大小"},
		{"--timeout <时长>", "操作超时时间"},
		{"--eager-mode", "积极模式（更快处理）"},
	}
	fmt.Printf("\n  %s%s%s\n", Colors.Yellow, T("性能选项:"), Colors.Reset)
	hs.printOptions(performanceOptions)

	// 界面选项
	uiOptions := [][]string{
		{"--color-output", "彩色输出（默认开启）"},
		{"--show-progress", "显示详细进度"},
		{"--show-stats", "显示统计信息"},
		{"--notifications", "桌面通知"},
		{"--lang <语言>", "设置界面语言"},
	}
	fmt.Printf("\n  %s%s%s\n", Colors.Yellow, T("界面选项:"), Colors.Reset)
	hs.printOptions(uiOptions)

	// 过滤选项
	filterOptions := [][]string{
		{"--include-pattern <模式>", "包含文件模式"},
		{"--exclude-pattern <模式>", "排除文件模式"},
		{"--file-size-limit <大小>", "文件大小限制"},
		{"--age-filter <时长>", "文件年龄过滤器"},
		{"--skip-hidden", "跳过隐藏文件"},
		{"--follow-symlinks", "跟随符号链接"},
	}
	fmt.Printf("\n  %s%s%s\n", Colors.Yellow, T("过滤选项:"), Colors.Reset)
	hs.printOptions(filterOptions)

	// 配置选项
	configOptions := [][]string{
		{"--config <文件>", "指定配置文件路径"},
		{"--install", "安装系统别名"},
		{"--uninstall", "卸载系统别名"},
		{"--restore", "从回收站恢复文件"},
	}
	fmt.Printf("\n  %s%s%s\n", Colors.Yellow, T("配置管理:"), Colors.Reset)
	hs.printOptions(configOptions)
}

// printOptions 打印选项列表
func (hs *HelpSystem) printOptions(options [][]string) {
	for _, option := range options {
		flag := option[0]
		desc := option[1]

		if hs.feedbackMgr.colorEnabled {
			fmt.Printf("  %s%-25s%s %s\n", Colors.Green, flag, Colors.Reset, T(desc))
		} else {
			fmt.Printf("  %-25s %s\n", flag, T(desc))
		}
	}
}

// showExamples 显示使用示例
func (hs *HelpSystem) showExamples() {
	fmt.Printf("%s%s%s\n", Colors.Bold, T("使用示例:"), Colors.Reset)

	switch hs.mode {
	case ModeDel:
		hs.showDelExamples()
	case ModeRM:
		hs.showRmExamples()
	case ModeCP:
		hs.showCpExamples()
	default:
		hs.showDelGuardExamples()
	}

	fmt.Println()
}

// showDelExamples 显示del模式示例
func (hs *HelpSystem) showDelExamples() {
	examples := [][]string{
		{"del file.txt", "删除单个文件"},
		{"del -r folder/", "递归删除目录"},
		{"del *.tmp", "删除所有.tmp文件"},
		{"del -i important.doc", "交互式删除重要文件"},
		{"del -n folder/", "试运行：显示将要删除的文件"},
	}
	hs.printExamples(examples)
}

// showRmExamples 显示rm模式示例
func (hs *HelpSystem) showRmExamples() {
	examples := [][]string{
		{"rm file.txt", "删除单个文件"},
		{"rm -rf folder/", "强制递归删除目录"},
		{"rm -i *.log", "交互式删除所有日志文件"},
		{"rm --smart-search myfile", "智能搜索并删除相似文件"},
	}
	hs.printExamples(examples)
}

// showCpExamples 显示cp模式示例
func (hs *HelpSystem) showCpExamples() {
	examples := [][]string{
		{"cp file.txt backup.txt", "复制文件"},
		{"cp -r folder/ backup/", "递归复制目录"},
		{"cp -p file.txt dest/", "保持属性复制"},
		{"cp --verify-integrity data.zip dest/", "验证完整性复制"},
		{"cp --protect file.txt existing.txt", "保护模式复制"},
	}
	hs.printExamples(examples)
}

// showDelGuardExamples 显示DelGuard示例
func (hs *HelpSystem) showDelGuardExamples() {
	examples := [][]string{
		{"delguard file.txt", "安全删除文件（移动到回收站）"},
		{"delguard -r --smart-search myfolder", "智能搜索并递归删除目录"},
		{"delguard --auto-backup important.doc", "自动备份后删除"},
		{"delguard --parallel *.log", "并行删除所有日志文件"},
		{"delguard --secure-delete secret.txt", "安全删除（多次覆写）"},
		{"delguard --include-pattern '*.tmp' --exclude-pattern '*important*' .", "按模式删除文件"},
		{"delguard --config /path/to/config.json file.txt", "使用自定义配置"},
		{"delguard --restore", "从回收站恢复文件"},
		{"delguard --install", "安装系统别名"},
	}
	hs.printExamples(examples)
}

// printExamples 打印示例列表
func (hs *HelpSystem) printExamples(examples [][]string) {
	for _, example := range examples {
		command := example[0]
		desc := example[1]

		if hs.feedbackMgr.colorEnabled {
			fmt.Printf("  %s$%s %s%-40s%s # %s\n",
				Colors.Gray, Colors.Reset, Colors.Cyan, command, Colors.Reset, T(desc))
		} else {
			fmt.Printf("  $ %-40s # %s\n", command, T(desc))
		}
	}
}

// showAdvancedOptions 显示高级选项
func (hs *HelpSystem) showAdvancedOptions() {
	fmt.Printf("%s%s%s\n", Colors.Bold, T("高级配置:"), Colors.Reset)

	fmt.Printf("\n  %s%s%s\n", Colors.Yellow, T("环境变量:"), Colors.Reset)
	envVars := [][]string{
		{"DELGUARD_CONFIG", "配置文件路径"},
		{"DELGUARD_LANGUAGE", "界面语言"},
		{"DELGUARD_INTERACTIVE", "默认交互模式"},
		{"DELGUARD_USE_RECYCLE_BIN", "是否使用回收站"},
		{"DELGUARD_LOG_LEVEL", "日志级别"},
		{"DELGUARD_MAX_WORKERS", "最大工作线程数"},
	}
	hs.printOptions(envVars)

	fmt.Printf("\n  %s%s%s\n", Colors.Yellow, T("配置文件支持的格式:"), Colors.Reset)
	formats := [][]string{
		{"JSON (.json)", "标准JSON格式"},
		{"JSONC (.jsonc)", "支持注释的JSON"},
		{"YAML (.yaml, .yml)", "YAML格式"},
		{"TOML (.toml)", "TOML格式"},
		{"INI (.ini, .cfg)", "INI配置格式"},
		{"Properties (.properties)", "Java Properties格式"},
	}
	hs.printOptions(formats)

	fmt.Printf("\n  %s%s%s\n", Colors.Yellow, T("支持的语言:"), Colors.Reset)
	languages := [][]string{
		{"zh-CN", "简体中文"},
		{"zh-TW", "繁体中文"},
		{"en-US", "英语"},
		{"ja", "日语"},
		{"ko-KR", "韩语"},
		{"es-ES", "西班牙语"},
		{"fr-FR", "法语"},
		{"de-DE", "德语"},
		{"ru-RU", "俄语"},
		{"ar-SA", "阿拉伯语"},
		{"auto", "自动检测"},
	}
	hs.printOptions(languages)

	fmt.Println()
}

// showConfiguration 显示配置信息
func (hs *HelpSystem) showConfiguration() {
	fmt.Printf("%s%s%s\n", Colors.Bold, T("配置文件位置:"), Colors.Reset)

	// 配置文件搜索路径
	paths := []string{
		"~/.delguard/config.json",
		"/etc/delguard/config.json",
		"./config.json",
	}

	if runtime.GOOS == "windows" {
		paths = []string{
			"%USERPROFILE%\\.delguard\\config.json",
			"%SystemRoot%\\delguard\\config.json",
			".\\config.json",
		}
	}

	fmt.Printf(T("  DelGuard按以下优先级搜索配置文件:\n"))
	for i, path := range paths {
		fmt.Printf("  %d. %s\n", i+1, path)
	}

	fmt.Printf("\n  %s%s%s\n", Colors.Yellow, T("配置示例:"), Colors.Reset)
	fmt.Printf(`  {
    "use_recycle_bin": true,
    "interactive_mode": "confirm",
    "language": "auto",
    "performance": {
      "parallel": true,
      "max_workers": 4
    },
    "ui": {
      "color_output": true,
      "show_progress": true
    }
  }
`)

	fmt.Println()
}

// showTroubleshooting 显示故障排除
func (hs *HelpSystem) showTroubleshooting() {
	fmt.Printf("%s%s%s\n", Colors.Bold, T("故障排除:"), Colors.Reset)

	troubleshooting := [][]string{
		{"权限不足", "尝试以管理员身份运行程序"},
		{"文件不存在", "使用 --smart-search 选项进行智能搜索"},
		{"文件被占用", "关闭使用该文件的程序，或使用 --force 选项"},
		{"回收站不支持", "检查当前平台是否支持回收站功能"},
		{"配置文件错误", "使用 --validate-only 选项验证配置文件"},
		{"语言包缺失", "检查语言包文件是否完整"},
		{"性能问题", "调整 --max-workers 和 --batch-size 参数"},
		{"内存不足", "使用 --lazy-load 选项减少内存使用"},
	}

	for _, item := range troubleshooting {
		problem := item[0]
		solution := item[1]
		fmt.Printf("  %s%s%s %s\n", Colors.Red, T("问题:"), Colors.Reset, T(problem))
		fmt.Printf("  %s%s%s %s\n\n", Colors.Green, T("解决:"), Colors.Reset, T(solution))
	}
}

// showFooter 显示页脚信息
func (hs *HelpSystem) showFooter() {
	fmt.Printf("%s%s%s\n", Colors.Bold, T("更多信息:"), Colors.Reset)
	fmt.Printf("  %s https://github.com/delguard/delguard\n", T("项目主页:"))
	fmt.Printf("  %s https://github.com/delguard/delguard/issues\n", T("问题报告:"))
	fmt.Printf("  %s https://delguard.readthedocs.io\n", T("文档地址:"))

	fmt.Printf("\n%s%s:%s\n", Colors.Bold, T("版本信息"), Colors.Reset)
	fmt.Printf("  %s %s\n", T("版本:"), version)
	fmt.Printf("  %s %s\n", T("构建时间:"), time.Now().Format("2006-01-02"))
	fmt.Printf("  %s %s\n", T("Go版本:"), runtime.Version())
	fmt.Printf("  %s %s/%s\n", T("系统:"), runtime.GOOS, runtime.GOARCH)

	fmt.Printf("\n%s%s%s MIT License\n", Colors.Bold, T("许可证:"), Colors.Reset)
	fmt.Printf("Copyright (c) 2024 DelGuard Team\n")

	if hs.feedbackMgr.colorEnabled {
		fmt.Printf("\n%s%s%s %s\n",
			Colors.Bold+Colors.Green, T("感谢使用 DelGuard！"), Colors.Reset, Icons.Success)
	} else {
		fmt.Printf("\n%s\n", T("感谢使用 DelGuard！"))
	}
}

// ShowQuickHelp 显示快速帮助
func (hs *HelpSystem) ShowQuickHelp() {
	modeName := strings.ToLower(hs.getModeDisplayName())

	fmt.Printf(T("快速帮助 - %s\n\n"), modeName)

	switch hs.mode {
	case ModeDel:
		fmt.Printf(T("常用命令:\n"))
		fmt.Printf("  del file.txt          # %s\n", T("删除文件"))
		fmt.Printf("  del -r folder/        # %s\n", T("删除目录"))
		fmt.Printf("  del -i important.*    # %s\n", T("交互式删除"))
	case ModeRM:
		fmt.Printf(T("常用命令:\n"))
		fmt.Printf("  rm file.txt           # %s\n", T("删除文件"))
		fmt.Printf("  rm -rf folder/        # %s\n", T("强制删除目录"))
		fmt.Printf("  rm -i *.log          # %s\n", T("交互式删除"))
	case ModeCP:
		fmt.Printf(T("常用命令:\n"))
		fmt.Printf("  cp file.txt dest.txt  # %s\n", T("复制文件"))
		fmt.Printf("  cp -r src/ dest/      # %s\n", T("复制目录"))
		fmt.Printf("  cp -p file.txt dest/  # %s\n", T("保持属性"))
	default:
		fmt.Printf(T("常用命令:\n"))
		fmt.Printf("  delguard file.txt     # %s\n", T("安全删除"))
		fmt.Printf("  delguard --install    # %s\n", T("安装别名"))
		fmt.Printf("  delguard --restore    # %s\n", T("恢复文件"))
	}

	fmt.Printf(T("\n使用 %s --help 查看完整帮助\n"), modeName)
}

// ShowVersionInfo 显示版本信息
func (hs *HelpSystem) ShowVersionInfo() {
	if hs.feedbackMgr.colorEnabled {
		fmt.Printf("%sDelGuard%s v%s%s%s\n",
			Colors.Bold+Colors.Blue, Colors.Reset,
			Colors.Cyan, version, Colors.Reset)
	} else {
		fmt.Printf("DelGuard v%s\n", version)
	}

	fmt.Printf("%s\n", T("跨平台安全删除工具"))
	fmt.Printf("%s %s %s\n", T("构建信息:"), runtime.GOOS, runtime.GOARCH)
	fmt.Printf("%s %s\n", T("Go版本:"), runtime.Version())
	fmt.Printf("%s MIT\n", T("许可证:"))
}
