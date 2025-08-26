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

// ConfigGenerator 交互式配置生成器
type ConfigGenerator struct {
	scanner *bufio.Scanner
	config  *Config
}

// NewConfigGenerator 创建新的配置生成器
func NewConfigGenerator() *ConfigGenerator {
	return &ConfigGenerator{
		scanner: bufio.NewScanner(os.Stdin),
		config:  &Config{},
	}
}

// GenerateInteractiveConfig 交互式生成配置文件
func (cg *ConfigGenerator) GenerateInteractiveConfig() error {
	cg.showWelcome()

	// 基本设置
	if err := cg.configureBasicSettings(); err != nil {
		return err
	}

	// 安全设置
	if err := cg.configureSecuritySettings(); err != nil {
		return err
	}

	// 高级设置
	if err := cg.configureAdvancedSettings(); err != nil {
		return err
	}

	// 平台特定设置
	if err := cg.configurePlatformSettings(); err != nil {
		return err
	}

	// 显示配置预览
	cg.showConfigPreview()

	// 确认并保存
	return cg.confirmAndSave()
}

// showWelcome 显示欢迎界面
func (cg *ConfigGenerator) showWelcome() {
	fmt.Println()
	fmt.Println("🎯 DelGuard 交互式配置生成器")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println()
	fmt.Println("欢迎使用 DelGuard 配置生成器！")
	fmt.Println("我将引导您创建个性化的配置文件。")
	fmt.Println()
	fmt.Println("💡 提示：")
	fmt.Println("  • 直接按回车使用默认值（显示在括号中）")
	fmt.Println("  • 输入 'help' 获取选项说明")
	fmt.Println("  • 输入 'skip' 跳过当前设置")
	fmt.Println()
	cg.waitForEnter("按回车键开始配置...")
}

// configureBasicSettings 配置基本设置
func (cg *ConfigGenerator) configureBasicSettings() error {
	fmt.Println("📋 基本设置")
	fmt.Println("-" + strings.Repeat("-", 30))

	// 语言设置
	cg.config.Language = cg.askChoice(
		"🌐 选择界面语言",
		[]string{"zh-cn", "en-us", "ja-jp", "ko-kr", "fr-fr", "de-de", "es-es"},
		"zh-cn",
		map[string]string{
			"zh-cn": "简体中文",
			"en-us": "English",
			"ja-jp": "日本語",
			"ko-kr": "한국어",
			"fr-fr": "Français",
			"de-de": "Deutsch",
			"es-es": "Español",
		},
	)

	// 交互模式
	cg.config.InteractiveMode = cg.askChoice(
		"🤝 交互确认模式",
		[]string{"always", "confirm", "never"},
		"confirm",
		map[string]string{
			"always":  "总是询问确认",
			"confirm": "危险操作时确认",
			"never":   "从不询问（谨慎使用）",
		},
	)

	// 日志级别
	cg.config.LogLevel = cg.askChoice(
		"📝 日志详细程度",
		[]string{"debug", "info", "warn", "error"},
		"info",
		map[string]string{
			"debug": "调试级别（最详细）",
			"info":  "信息级别（推荐）",
			"warn":  "警告级别",
			"error": "错误级别（最简洁）",
		},
	)

	// 安全模式
	cg.config.SafeMode = cg.askChoice(
		"🛡️  安全模式",
		[]string{"strict", "normal", "relaxed"},
		"normal",
		map[string]string{
			"strict":  "严格模式（最安全，限制较多）",
			"normal":  "标准模式（推荐）",
			"relaxed": "宽松模式（限制较少）",
		},
	)

	// 回收站设置
	cg.config.UseRecycleBin = cg.askYesNo(
		"🗑️  是否使用系统回收站",
		true,
		"启用后删除的文件会进入回收站，可以恢复",
	)

	fmt.Println()
	return nil
}

// configureSecuritySettings 配置安全设置
func (cg *ConfigGenerator) configureSecuritySettings() error {
	fmt.Println("🔒 安全设置")
	fmt.Println("-" + strings.Repeat("-", 30))

	cg.config.EnableSecurityChecks = cg.askYesNo(
		"🔍 启用安全检查",
		true,
		"检查文件权限、路径安全等",
	)

	cg.config.EnablePathValidation = cg.askYesNo(
		"🛣️  启用路径验证",
		true,
		"防止路径遍历攻击和非法路径",
	)

	cg.config.EnableHiddenCheck = cg.askYesNo(
		"👁️  删除隐藏文件时确认",
		true,
		"删除隐藏文件前会询问确认",
	)

	cg.config.EnableOverwriteProtection = cg.askYesNo(
		"🛡️  启用覆盖保护",
		true,
		"防止意外覆盖重要文件",
	)

	// 文件大小限制
	maxSizeStr := cg.askString(
		"📏 单个文件大小限制 (MB)",
		"100",
		"超过此大小的文件需要额外确认",
	)
	if maxSize, err := strconv.ParseInt(maxSizeStr, 10, 64); err == nil {
		cg.config.MaxFileSize = maxSize * 1024 * 1024 // 转换为字节
	} else {
		cg.config.MaxFileSize = 100 * 1024 * 1024 // 默认100MB
	}

	fmt.Println()
	return nil
}

// configureAdvancedSettings 配置高级设置
func (cg *ConfigGenerator) configureAdvancedSettings() error {
	fmt.Println("⚙️  高级设置")
	fmt.Println("-" + strings.Repeat("-", 30))

	// 备份保留天数
	retentionStr := cg.askString(
		"📅 备份文件保留天数",
		"30",
		"超过此天数的备份文件会被自动清理",
	)
	if retention, err := strconv.Atoi(retentionStr); err == nil {
		cg.config.BackupRetentionDays = retention
	} else {
		cg.config.BackupRetentionDays = 30
	}

	// 最大并发操作数
	concurrentStr := cg.askString(
		"🔄 最大并发操作数",
		"10",
		"同时处理的文件数量，影响性能和资源使用",
	)
	if concurrent, err := strconv.Atoi(concurrentStr); err == nil {
		cg.config.MaxConcurrentOps = concurrent
	} else {
		cg.config.MaxConcurrentOps = 10
	}

	// 输出前缀设置
	cg.config.OutputPrefixEnabled = cg.askYesNo(
		"🏷️  启用输出前缀",
		true,
		"在消息前添加 'DelGuard:' 前缀",
	)

	if cg.config.OutputPrefixEnabled {
		cg.config.OutputPrefix = cg.askString(
			"✏️  自定义输出前缀",
			"DelGuard: ",
			"自定义消息前缀文本",
		)
	}

	// 日志轮转设置
	cg.config.LogRotateDaily = cg.askYesNo(
		"📊 启用日志按日轮转",
		false,
		"每天创建新的日志文件",
	)

	if cg.config.LogRotateDaily {
		maxSizeStr := cg.askString(
			"📦 日志文件最大大小 (MB)",
			"10",
			"单个日志文件的最大大小",
		)
		if maxSize, err := strconv.ParseInt(maxSizeStr, 10, 64); err == nil {
			cg.config.LogMaxSize = maxSize
		} else {
			cg.config.LogMaxSize = 10
		}

		maxBackupsStr := cg.askString(
			"🗂️  保留日志文件数量",
			"7",
			"保留的历史日志文件数量",
		)
		if maxBackups, err := strconv.Atoi(maxBackupsStr); err == nil {
			cg.config.LogMaxBackups = maxBackups
		} else {
			cg.config.LogMaxBackups = 7
		}

		cg.config.LogCompress = cg.askYesNo(
			"🗜️  压缩历史日志",
			true,
			"压缩旧日志文件以节省空间",
		)
	}

	fmt.Println()
	return nil
}

// configurePlatformSettings 配置平台特定设置
func (cg *ConfigGenerator) configurePlatformSettings() error {
	fmt.Println("🖥️  平台特定设置")
	fmt.Println("-" + strings.Repeat("-", 30))

	switch runtime.GOOS {
	case "windows":
		return cg.configureWindowsSettings()
	case "linux":
		return cg.configureLinuxSettings()
	case "darwin":
		return cg.configureDarwinSettings()
	default:
		fmt.Println("当前平台无特殊配置项")
	}

	fmt.Println()
	return nil
}

// configureWindowsSettings 配置Windows特定设置
func (cg *ConfigGenerator) configureWindowsSettings() error {
	fmt.Println("🪟 Windows 设置")

	cg.config.Windows.UseSystemTrash = cg.askYesNo(
		"🗑️  使用系统回收站",
		true,
		"使用Windows系统回收站",
	)

	cg.config.Windows.EnableUACPrompt = cg.askYesNo(
		"🛡️  启用UAC提示",
		false,
		"需要管理员权限时显示UAC提示",
	)

	cg.config.Windows.CheckFileOwnership = cg.askYesNo(
		"👤 检查文件所有权",
		true,
		"删除前检查文件所有者权限",
	)

	return nil
}

// configureLinuxSettings 配置Linux特定设置
func (cg *ConfigGenerator) configureLinuxSettings() error {
	fmt.Println("🐧 Linux 设置")

	cg.config.Linux.UseXDGTrash = cg.askYesNo(
		"🗑️  使用XDG回收站",
		true,
		"使用符合XDG标准的回收站",
	)

	trashDir := cg.askString(
		"📁 回收站目录",
		"~/.local/share/Trash",
		"自定义回收站目录路径",
	)
	cg.config.Linux.TrashDir = trashDir

	cg.config.Linux.CheckSELinux = cg.askYesNo(
		"🔒 检查SELinux",
		false,
		"在SELinux环境中进行额外检查",
	)

	return nil
}

// configureDarwinSettings 配置macOS特定设置
func (cg *ConfigGenerator) configureDarwinSettings() error {
	fmt.Println("🍎 macOS 设置")

	cg.config.Darwin.UseSystemTrash = cg.askYesNo(
		"🗑️  使用系统废纸篓",
		true,
		"使用macOS系统废纸篓",
	)

	trashDir := cg.askString(
		"📁 废纸篓目录",
		"~/.Trash",
		"自定义废纸篓目录路径",
	)
	cg.config.Darwin.TrashDir = trashDir

	cg.config.Darwin.CheckFileVault = cg.askYesNo(
		"🔐 检查FileVault",
		false,
		"在FileVault环境中进行额外检查",
	)

	return nil
}

// showConfigPreview 显示配置预览
func (cg *ConfigGenerator) showConfigPreview() {
	fmt.Println("👀 配置预览")
	fmt.Println("=" + strings.Repeat("=", 50))

	fmt.Printf("🌐 语言: %s\n", cg.config.Language)
	fmt.Printf("🤝 交互模式: %s\n", cg.config.InteractiveMode)
	fmt.Printf("📝 日志级别: %s\n", cg.config.LogLevel)
	fmt.Printf("🛡️  安全模式: %s\n", cg.config.SafeMode)
	fmt.Printf("🗑️  使用回收站: %v\n", cg.config.UseRecycleBin)
	fmt.Printf("🔍 安全检查: %v\n", cg.config.EnableSecurityChecks)
	fmt.Printf("📏 文件大小限制: %d MB\n", cg.config.MaxFileSize/(1024*1024))
	fmt.Printf("📅 备份保留: %d 天\n", cg.config.BackupRetentionDays)
	fmt.Printf("🔄 最大并发: %d\n", cg.config.MaxConcurrentOps)

	if cg.config.OutputPrefixEnabled {
		fmt.Printf("🏷️  输出前缀: \"%s\"\n", cg.config.OutputPrefix)
	}

	fmt.Println()
}

// confirmAndSave 确认并保存配置
func (cg *ConfigGenerator) confirmAndSave() error {
	if !cg.askYesNo("💾 保存此配置", true, "将配置保存到文件") {
		fmt.Println("❌ 配置生成已取消")
		return nil
	}

	// 设置配置版本和时间戳
	cg.config.Version = "1.0.0"
	cg.config.SchemaVersion = "1.0"

	// 设置默认值
	cg.setDefaultValues()

	// 创建配置目录
	configDir := "config"
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 生成配置文件
	configPath := filepath.Join(configDir, "delguard.json")
	if err := cg.saveConfigWithComments(configPath); err != nil {
		return fmt.Errorf("保存配置文件失败: %v", err)
	}

	fmt.Printf("✅ 配置文件已保存到: %s\n", configPath)
	fmt.Println()
	fmt.Println("🎉 配置生成完成！")
	fmt.Println("💡 提示：")
	fmt.Println("  • 您可以随时编辑配置文件")
	fmt.Println("  • 运行 'delguard --init-config' 重新生成配置")
	fmt.Println("  • 运行 'delguard --health-check' 检查系统状态")

	return nil
}

// setDefaultValues 设置默认值
func (cg *ConfigGenerator) setDefaultValues() {
	if cg.config.MaxBackupFiles == 0 {
		cg.config.MaxBackupFiles = DefaultMaxBackupFiles
	}
	if cg.config.TrashMaxSize == 0 {
		cg.config.TrashMaxSize = DefaultTrashMaxSize
	}
	if cg.config.MaxPathLength == 0 {
		cg.config.MaxPathLength = DefaultMaxPathLength
	}
	if cg.config.SimilarityThreshold == 0 {
		cg.config.SimilarityThreshold = 0.8
	}
	if cg.config.LogRetentionDays == 0 {
		cg.config.LogRetentionDays = 30
	}
}

// saveConfigWithComments 保存带注释的配置文件
func (cg *ConfigGenerator) saveConfigWithComments(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// 写入文件头注释
	fmt.Fprintln(file, "// DelGuard 配置文件")
	fmt.Fprintln(file, "// 此文件由交互式配置生成器创建")
	fmt.Fprintf(file, "// 生成时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintln(file, "// 您可以手动编辑此文件，或运行 'delguard --init-config' 重新生成")
	fmt.Fprintln(file, "//")
	fmt.Fprintln(file, "// 配置说明:")
	fmt.Fprintln(file, "// - language: 界面语言 (zh-cn, en-us, ja-jp, ko-kr, fr-fr, de-de, es-es)")
	fmt.Fprintln(file, "// - interactive_mode: 交互模式 (always=总是确认, confirm=危险操作确认, never=从不确认)")
	fmt.Fprintln(file, "// - log_level: 日志级别 (debug, info, warn, error)")
	fmt.Fprintln(file, "// - safe_mode: 安全模式 (strict=严格, normal=标准, relaxed=宽松)")
	fmt.Fprintln(file, "// - use_recycle_bin: 是否使用系统回收站")
	fmt.Fprintln(file, "// - max_file_size: 单个文件大小限制 (字节)")
	fmt.Fprintln(file, "// - backup_retention_days: 备份文件保留天数")
	fmt.Fprintln(file, "// - max_concurrent_ops: 最大并发操作数")
	fmt.Fprintln(file, "//")
	fmt.Fprintln(file, "{")

	// 序列化配置为JSON
	data, err := json.MarshalIndent(cg.config, "  ", "  ")
	if err != nil {
		return err
	}

	// 写入JSON内容（去掉第一个和最后一个大括号）
	jsonStr := string(data)
	lines := strings.Split(jsonStr, "\n")
	for i := 1; i < len(lines)-1; i++ {
		fmt.Fprintln(file, "  "+lines[i])
	}

	fmt.Fprintln(file, "}")

	return nil
}

// askChoice 询问选择题
func (cg *ConfigGenerator) askChoice(question string, options []string, defaultValue string, descriptions map[string]string) string {
	fmt.Printf("\n%s:\n", question)

	for i, option := range options {
		desc := descriptions[option]
		if desc == "" {
			desc = option
		}
		marker := " "
		if option == defaultValue {
			marker = "✓"
		}
		fmt.Printf("  %s %d) %s - %s\n", marker, i+1, option, desc)
	}

	fmt.Printf("\n请选择 (1-%d) [默认: %s]: ", len(options), defaultValue)

	if cg.scanner.Scan() {
		input := strings.TrimSpace(cg.scanner.Text())
		if input == "" {
			return defaultValue
		}
		if input == "help" {
			fmt.Println("💡 选项说明:")
			for _, option := range options {
				fmt.Printf("  %s: %s\n", option, descriptions[option])
			}
			return cg.askChoice(question, options, defaultValue, descriptions)
		}
		if input == "skip" {
			return defaultValue
		}

		// 尝试解析数字选择
		if choice, err := strconv.Atoi(input); err == nil && choice >= 1 && choice <= len(options) {
			return options[choice-1]
		}

		// 尝试直接匹配选项
		for _, option := range options {
			if strings.EqualFold(input, option) {
				return option
			}
		}

		fmt.Printf("❌ 无效选择，请重新输入\n")
		return cg.askChoice(question, options, defaultValue, descriptions)
	}

	return defaultValue
}

// askYesNo 询问是否问题
func (cg *ConfigGenerator) askYesNo(question string, defaultValue bool, description string) bool {
	defaultStr := "n"
	if defaultValue {
		defaultStr = "y"
	}

	fmt.Printf("\n%s\n", question)
	if description != "" {
		fmt.Printf("  💡 %s\n", description)
	}
	fmt.Printf("请选择 (y/n) [默认: %s]: ", defaultStr)

	if cg.scanner.Scan() {
		input := strings.ToLower(strings.TrimSpace(cg.scanner.Text()))
		if input == "" {
			return defaultValue
		}
		if input == "help" {
			fmt.Printf("  y/yes: 是\n  n/no: 否\n")
			return cg.askYesNo(question, defaultValue, description)
		}
		if input == "skip" {
			return defaultValue
		}

		return input == "y" || input == "yes" || input == "true"
	}

	return defaultValue
}

// askString 询问字符串输入
func (cg *ConfigGenerator) askString(question, defaultValue, description string) string {
	fmt.Printf("\n%s\n", question)
	if description != "" {
		fmt.Printf("  💡 %s\n", description)
	}
	fmt.Printf("请输入 [默认: %s]: ", defaultValue)

	if cg.scanner.Scan() {
		input := strings.TrimSpace(cg.scanner.Text())
		if input == "" {
			return defaultValue
		}
		if input == "help" {
			fmt.Printf("  输入文本值，或按回车使用默认值\n")
			return cg.askString(question, defaultValue, description)
		}
		if input == "skip" {
			return defaultValue
		}

		return input
	}

	return defaultValue
}

// waitForEnter 等待用户按回车
func (cg *ConfigGenerator) waitForEnter(message string) {
	fmt.Print(message)
	cg.scanner.Scan()
}
