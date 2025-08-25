package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"delguard/utils"
)

// 命令模式枚举
type CommandMode int

const (
	ModeDelGuard CommandMode = iota // DelGuard 默认模式
	ModeDel                         // del 命令模式（Windows风格删除）
	ModeRM                          // rm 命令模式（Unix风格删除）
	ModeCP                          // cp 命令模式（复制）
)

// CommandDetector 命令检测器
type CommandDetector struct {
	programName string
	mode        CommandMode
}

// 全局变量
var (
	version                   = "1.0.0"
	verbose                   bool
	quiet                     bool
	recursive                 bool
	dryRun                    bool
	force                     bool
	interactive               bool
	installDefaultInteractive bool
	installAliasOnly          bool // 新增：仅安装别名
	uninstallAliasOnly        bool // 新增：仅卸载别名
	showVersion               bool
	showHelp                  bool
	validateOnly              bool          // 新增：仅验证模式
	timeout                   time.Duration // 新增：操作超时时间
	safeCopy                  bool          // 新增：安全复制模式
	protect                   bool          // 启用文件覆盖保护
	disableProtect            bool          // 禁用文件覆盖保护
	cpMode                    bool          // 新增：cp命令模式
	// 智能删除相关参数
	smartSearch         bool    // 启用智能搜索
	searchContent       bool    // 搜索文件内容
	searchParent        bool    // 搜索父目录
	similarityThreshold float64 // 相似度阈值
	maxResults          int     // 最大搜索结果数
	forceConfirm        bool    // 强制跳过确认
	// 配置文件路径覆盖
	configPath string
	// 新增的高级选项
	showProgress       bool          // 显示详细进度
	batchSize          int           // 批处理大小
	parallel           bool          // 启用并行处理
	maxWorkers         int           // 最大工作线程数
	autoBacKup         bool          // 自动备份
	backupDir          string        // 备份目录
	compressionLevel   int           // 压缩级别
	verifyIntegrity    bool          // 验证文件完整性
	secureDelete       bool          // 安全删除（多次覆写）
	showStats          bool          // 显示统计信息
	colorOutput        bool          // 彩色输出
	logFormat          string        // 日志格式
	notifications      bool          // 桌面通知
	preserveTimes      bool          // 保持时间戳
	skipHidden         bool          // 跳过隐藏文件
	fileSizeLimit      int64         // 文件大小限制
	includePattern     string        // 包含模式
	excludePattern     string        // 排除模式
	regexMode          bool          // 正则表达式模式
	caseSensitive      bool          // 大小写敏感
	followSymlinks     bool          // 跟随符号链接
	eagerMode          bool          // 积极模式（更快的操作）
	smartCleanup       bool          // 智能清理（自动清理空目录）
	conflictResolution string        // 冲突解决策略
	fileTypeFilters    []string      // 文件类型过滤器
	ageFilter          time.Duration // 文件年龄过滤器
	sizeFilter         string        // 文件大小过滤器
	customScript       string        // 自定义脚本路径
	hooksEnabled       bool          // 启用钩子系统
	// 命令检测相关
	cmdDetector *CommandDetector // 命令检测器
	currentMode CommandMode      // 当前命令模式
	// 安全组件
	inputValidator *InputValidator     // 输入验证器
	concurrencyMgr *ConcurrencyManager // 并发管理器
	resourceMgr    *ResourceManager    // 资源管理器
)

// TargetInfo 用于日志记录
type TargetInfo struct {
	Path string
}

// ArgumentResult 参数解析结果
type ArgumentResult struct {
	Targets []string // 目标文件/目录列表
	Flags   []string // 标志参数列表
}

// SmartArgumentParser 智能参数解析器
type SmartArgumentParser struct {
	args []string
	mode CommandMode // 当前命令模式
}

// NewCommandDetector 创建命令检测器
//
// 返回值:
//   - *CommandDetector: 命令检测器实例指针
func NewCommandDetector() *CommandDetector {
	return &CommandDetector{}
}

// DetectCommand 检测当前程序被调用的命令名称
func (cd *CommandDetector) DetectCommand() CommandMode {
	if len(os.Args) == 0 {
		return ModeDelGuard
	}

	// 获取程序名称（去除路径和扩展名）
	programPath := os.Args[0]
	programName := filepath.Base(programPath)

	// 去除Windows的.exe扩展名
	if runtime.GOOS == "windows" {
		programName = strings.TrimSuffix(programName, ".exe")
	}

	// 转换为小写进行匹配
	programName = strings.ToLower(programName)
	cd.programName = programName

	// 根据程序名称确定模式
	switch programName {
	case "del":
		cd.mode = ModeDel
		return ModeDel
	case "rm":
		cd.mode = ModeRM
		return ModeRM
	case "cp", "copy":
		cd.mode = ModeCP
		return ModeCP
	default:
		cd.mode = ModeDelGuard
		return ModeDelGuard
	}
}

// GetModeName 获取模式的友好名称
func (cd *CommandDetector) GetModeName() string {
	switch cd.mode {
	case ModeDel:
		return "del"
	case ModeRM:
		return "rm"
	case ModeCP:
		return "cp"
	default:
		return "delguard"
	}
}

// ApplyModeDefaults 根据检测到的命令模式应用默认设置
func (cd *CommandDetector) ApplyModeDefaults() {
	switch cd.mode {
	case ModeDel:
		// del命令默认设置（Windows风格）
		if !interactive {
			interactive = true // del命令默认开启交互模式
		}
		smartSearch = true // 启用智能搜索
	case ModeRM:
		// rm命令默认设置（Unix风格）
		// rm命令通常不默认开启交互模式，除非明确指定-i
		smartSearch = true // 启用智能搜索
	case ModeCP:
		// cp命令模式
		cpMode = true
		safeCopy = true // 启用安全复制
	default:
		// DelGuard默认模式
		smartSearch = true
		interactive = true // DelGuard默认开启交互模式
	}
}

// NewSmartArgumentParser 创建智能参数解析器
func NewSmartArgumentParser(args []string) *SmartArgumentParser {
	return &SmartArgumentParser{
		args: args,
		mode: currentMode,
	}
}

// ParseArguments 智能解析命令行参数，特别处理以'-'开头的文件
func (p *SmartArgumentParser) ParseArguments() (*ArgumentResult, error) {
	result := &ArgumentResult{
		Targets: make([]string, 0),
		Flags:   make([]string, 0),
	}

	// 如果是cp模式，使用不同的解析逻辑
	if p.mode == ModeCP {
		return p.parseCopyArguments()
	}

	forceFileMode := false // 强制文件模式（遇到--后）

	for i := 0; i < len(p.args); i++ {
		arg := p.args[i]

		// 验证输入参数安全性
		if inputValidator != nil {
			validationResult := inputValidator.ValidateArgument(arg)
			if !validationResult.IsValid {
				return nil, fmt.Errorf("参数验证失败: %s - %v", arg, validationResult.Errors)
			}
			// 如果参数被清理，使用清理后的版本
			if validationResult.Sanitized != arg {
				arg = validationResult.Sanitized
			}
		}

		// 如果遇到 "--"，后面的都当作文件处理
		if arg == "--" {
			forceFileMode = true
			for j := i + 1; j < len(p.args); j++ {
				// 验证路径参数
				fileArg := p.args[j]
				if inputValidator != nil {
					pathResult := inputValidator.ValidatePath(fileArg)
					if !pathResult.IsValid {
						return nil, fmt.Errorf("路径验证失败: %s - %v", fileArg, pathResult.Errors)
					}
					if pathResult.Sanitized != fileArg {
						fileArg = pathResult.Sanitized
					}
				}
				result.Targets = append(result.Targets, fileArg)
			}
			break
		}

		// 强制文件模式下，所有参数都当作文件
		if forceFileMode {
			result.Targets = append(result.Targets, arg)
			continue
		}

		// 如果是标志参数（以'-'开头但不是单独的'-'）
		if strings.HasPrefix(arg, "-") && arg != "-" {
			// 检查这个参数是否为已知的标志
			if p.isKnownFlag(arg) {
				result.Flags = append(result.Flags, arg)
				// 如果这个标志需要参数值，跳过下一个参数
				if p.flagNeedsValue(arg) && i+1 < len(p.args) {
					i++
					// 验证标志值
					flagValue := p.args[i]
					if inputValidator != nil {
						argResult := inputValidator.ValidateArgument(flagValue)
						if !argResult.IsValid {
							return nil, fmt.Errorf("标志值验证失败: %s=%s - %v", arg, flagValue, argResult.Errors)
						}
						if argResult.Sanitized != flagValue {
							flagValue = argResult.Sanitized
						}
					}
					result.Flags = append(result.Flags, flagValue)
				}
			} else {
				// 可能是以'-'开头的文件名，使用增强的检测逻辑
				isFile, suggestion := p.smartDetectFile(arg)
				if isFile {
					targetFile := arg
					if suggestion != "" {
						targetFile = suggestion
					}
					// 验证文件路径
					if inputValidator != nil {
						pathResult := inputValidator.ValidatePath(targetFile)
						if !pathResult.IsValid {
							return nil, fmt.Errorf("文件路径验证失败: %s - %v", targetFile, pathResult.Errors)
						}
						if pathResult.Sanitized != targetFile {
							targetFile = pathResult.Sanitized
						}
					}
					result.Targets = append(result.Targets, targetFile)
				} else {
					// 既不是已知标志，也不是文件
					return nil, p.createUnknownFlagError(arg)
				}
			}
		} else {
			// 普通文件参数（不以'-'开头）
			// 验证文件路径
			if inputValidator != nil {
				pathResult := inputValidator.ValidatePath(arg)
				if !pathResult.IsValid {
					return nil, fmt.Errorf("文件路径验证失败: %s - %v", arg, pathResult.Errors)
				}
				if pathResult.Sanitized != arg {
					arg = pathResult.Sanitized
				}
			}
			result.Targets = append(result.Targets, arg)
		}
	}

	return result, nil
}

// isKnownFlag 检查是否为已知的标志参数
func (p *SmartArgumentParser) isKnownFlag(arg string) bool {
	// 基本标志（所有模式都支持）
	commonFlags := map[string]bool{
		"-v": true, "-q": true, "-r": true, "-n": true, "-i": true, "-h": true,
		"--verbose": true, "--quiet": true, "--recursive": true, "--dry-run": true,
		"--force": true, "--interactive": true, "--help": true, "--version": true,
	}

	// 检查基本标志
	if commonFlags[arg] {
		return true
	}

	// DelGuard特定标志
	delguardFlags := map[string]bool{
		"--validate-only": true, "--safe-copy": true, "--protect": true,
		"--disable-protect": true, "--timeout": true, "--cp": true,
		"--smart-search": true, "--search-content": true, "--search-parent": true,
		"--similarity": true, "--max-results": true, "--force-confirm": true,
		"--default-interactive": true, "--install": true,
		"--uninstall": true, // 新增：卸载别名
		// 新增的高级选项
		"--show-progress": true, "--batch-size": true, "--parallel": true,
		"--max-workers": true, "--auto-backup": true, "--backup-dir": true,
		"--compression-level": true, "--verify-integrity": true, "--secure-delete": true,
		"--show-stats": true, "--color-output": true, "--log-format": true,
		"--notifications": true, "--preserve-times": true, "--skip-hidden": true,
		"--file-size-limit": true, "--include-pattern": true, "--exclude-pattern": true,
		"--regex-mode": true, "--case-sensitive": true, "--follow-symlinks": true,
		"--eager-mode": true, "--smart-cleanup": true, "--conflict-resolution": true,
		"--file-type-filters": true, "--age-filter": true, "--size-filter": true,
		"--custom-script": true, "--hooks-enabled": true, "--lang": true,
		"--config": true, "--restore": true,
		// 短参数支持
		"-f": true, "-y": true, "-p": true, "-s": true, "-c": true, "-b": true,
		"-w": true, "-e": true, "-a": true, "-t": true, "-l": true, "-o": true,
	}

	// 根据模式检查特定标志
	switch p.mode {
	case ModeCP:
		// cp模式只支持特定标志
		return p.isCopyFlag(arg)
	case ModeDel, ModeRM:
		// del/rm模式支持部分DelGuard标志
		supportedFlags := map[string]bool{
			"--smart-search": true, "--search-content": true, "--search-parent": true,
			"--similarity": true, "--max-results": true, "--force-confirm": true,
			"--timeout": true,
		}
		return supportedFlags[arg]
	default:
		// DelGuard默认模式支持所有标志
		return delguardFlags[arg]
	}
}

// flagNeedsValue 检查标志是否需要参数值
func (p *SmartArgumentParser) flagNeedsValue(flag string) bool {
	valueFlags := map[string]bool{
		"--timeout": true, "--similarity": true, "--max-results": true,
		"--batch-size": true, "--max-workers": true, "--backup-dir": true,
		"--compression-level": true, "--log-format": true, "--file-size-limit": true,
		"--include-pattern": true, "--exclude-pattern": true, "--conflict-resolution": true,
		"--file-type-filters": true, "--age-filter": true, "--size-filter": true,
		"--custom-script": true, "--lang": true, "--config": true,
		// 短参数
		"-t": true, "-s": true, "-b": true, "-w": true, "-c": true,
		"-l": true, "-o": true, "-a": true, "-e": true,
	}
	return valueFlags[flag]
}

// checkFileExists 检查文件是否存在
func (p *SmartArgumentParser) checkFileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// smartSearchDashFile 对以'-'开头的文件进行智能搜索
func (p *SmartArgumentParser) smartSearchDashFile(target string) (string, error) {
	// 去除开头的'-'符号进行搜索
	cleanTarget := strings.TrimPrefix(target, "-")
	if cleanTarget == target {
		// 如果没有'-'开头，直接返回
		return "", nil
	}

	// 在当前目录搜索类似的文件
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// 使用智能搜索引擎
	config := DefaultSmartSearchConfig
	config.SimilarityThreshold = similarityThreshold
	config.MaxResults = maxResults
	searcher := NewSmartFileSearch(config)

	// 搜索清理后的文件名
	results, err := searcher.SearchFiles(cleanTarget, cwd)
	if err != nil {
		return "", err
	}

	// 如果没有找到结果，也尝试搜索原始名称（包含'-'）
	if len(results) == 0 {
		results, err = searcher.SearchFiles(target, cwd)
		if err != nil {
			return "", err
		}
	}

	if len(results) == 0 {
		return "", nil
	}

	// 显示搜索结果并让用户选择
	fmt.Printf(T("\n未找到文件 '%s'，但发现以下相似文件：\n"), target)
	for i, result := range results {
		fmt.Printf(T("  %d. %s (相似度: %.1f%%, 匹配方式: %s)\n"),
			i+1, result.Path, result.Similarity, result.MatchType)
		if result.Context != "" {
			fmt.Printf(T("     内容匹配: %s\n"), result.Context)
		}
	}
	fmt.Printf(T("  0. 取消操作\n"))

	// 读取用户选择（带交互检测与超时，避免在无TTY环境阻塞）
	var choice int
	fmt.Printf(T("请选择要删除的文件 (0-%d): "), len(results))
	var input string
	if isStdinInteractive() {
		if s, ok := readLineWithTimeout(20 * time.Second); ok {
			input = strings.TrimSpace(s)
		} else {
			input = ""
		}
	} else {
		input = ""
	}
	if input == "" {
		choice = 0
	} else {
		var err error
		choice, err = strconv.Atoi(input)
		if err != nil {
			fmt.Printf(T("无效输入，取消操作\n"))
			return "", nil
		}
	}

	if choice <= 0 || choice > len(results) {
		fmt.Printf(T("取消操作\n"))
		return "", nil
	}

	selected := results[choice-1]
	fmt.Printf(T("选择了文件: %s\n"), selected.Path)
	return selected.Path, nil
}

// parseCopyArguments 解析复制命令参数
func (p *SmartArgumentParser) parseCopyArguments() (*ArgumentResult, error) {
	result := &ArgumentResult{
		Targets: make([]string, 0),
		Flags:   make([]string, 0),
	}

	// cp命令的参数解析相对简单，只需要区分标志和文件
	for i := 0; i < len(p.args); i++ {
		arg := p.args[i]
		if strings.HasPrefix(arg, "-") && arg != "-" {
			if p.isCopyFlag(arg) {
				result.Flags = append(result.Flags, arg)
			} else {
				// cp命令中的未知标志，当作文件处理
				result.Targets = append(result.Targets, arg)
			}
		} else {
			result.Targets = append(result.Targets, arg)
		}
	}

	return result, nil
}

// isCopyFlag 检查是否为cp命令的有效标志
func (p *SmartArgumentParser) isCopyFlag(arg string) bool {
	copyFlags := map[string]bool{
		"-r": true, "--recursive": true,
		"-i": true, "--interactive": true,
		"-f": true, "--force": true,
		"-v": true, "--verbose": true,
		"-p": true, "--preserve": true,
		"-a": true, "--archive": true,
		"-u": true, "--update": true,
		"-n": true, "--no-clobber": true,
	}
	return copyFlags[arg]
}

// smartDetectFile 智能检测是否为文件（增强版）
func (p *SmartArgumentParser) smartDetectFile(arg string) (bool, string) {
	// 1. 直接检查文件是否存在
	if p.checkFileExists(arg) {
		return true, ""
	}

	// 2. 如果启用了智能搜索，尝试智能匹配
	if smartSearch {
		suggestion, err := p.smartSearchDashFile(arg)
		if err == nil && suggestion != "" {
			return true, suggestion
		}
	}

	// 3. 检查是否为常见的文件名模式
	if p.looksLikeFileName(arg) {
		return true, ""
	}

	return false, ""
}

// looksLikeFileName 检查字符串是否看起来像文件名
func (p *SmartArgumentParser) looksLikeFileName(arg string) bool {
	// 包含文件扩展名
	if strings.Contains(arg, ".") && len(strings.Split(arg, ".")) > 1 {
		ext := filepath.Ext(arg)
		if len(ext) > 1 && len(ext) <= 5 { // 合理的扩展名长度
			return true
		}
	}

	// 包含路径分隔符
	if strings.Contains(arg, "/") || strings.Contains(arg, "\\") {
		return true
	}

	// 以 '-' 开头的参数只有在包含扩展名或路径分隔符时才可能是文件
	// 避免将 "-unknown" 等误判为文件名，应返回未知标志错误
	// 若确为文件，智能搜索 smartSearch 会返回建议并在其它路径处理中接管

	return false
}

// isShortFlagCombination 检查是否为短标志组合（如-rf、-la等）
func (p *SmartArgumentParser) isShortFlagCombination(arg string) bool {
	if len(arg) < 3 || arg[0] != '-' {
		return false
	}

	// 常见的短标志组合
	commonCombinations := []string{
		"-rf", "-la", "-al", "-lt", "-lh", "-ls",
		"-iv", "-vf", "-rv", "-ri", "-fi",
	}

	for _, combo := range commonCombinations {
		if arg == combo {
			return true
		}
	}

	// 检查是否为已知短标志的组合
	for i := 1; i < len(arg); i++ {
		flagChar := "-" + string(arg[i])
		if !p.isKnownFlag(flagChar) {
			return false
		}
	}

	return true
}

// createUnknownFlagError 创建未知标志错误
func (p *SmartArgumentParser) createUnknownFlagError(arg string) error {
	msg := fmt.Sprintf("未知标志: %s", arg)

	// 根据模式提供不同的建议
	switch p.mode {
	case ModeDel:
		msg += "\n提示：如果这是文件名，请使用 del -- " + arg
	case ModeRM:
		msg += "\n提示：如果这是文件名，请使用 rm -- " + arg
	case ModeCP:
		msg += "\n提示：cp命令不支持此标志"
	default:
		msg += "\n提示：如果这是文件名，请使用 delguard -- " + arg
	}

	return fmt.Errorf(msg)
}

// humanizedFileProcessor 人性化文件处理器
func humanizedFileProcessor(filePath string, specialHandler *SpecialFileHandler, trashMonitor *TrashOperationMonitor, forceMode bool) error {
	// 1. 检查关键文件保护
	if err := checkCriticalProtection(filePath, forceMode); err != nil {
		return err
	}

	// 2. 检查回收站操作
	trashOp, err := trashMonitor.DetectTrashOperation(filePath)
	if err != nil {
		return fmt.Errorf("回收站检测失败: %v", err)
	}

	if trashOp != nil {
		allowed, err := trashMonitor.WarnTrashOperation(trashOp, forceMode)
		if err != nil {
			return err
		}
		if !allowed {
			return fmt.Errorf("用户取消回收站操作")
		}
		trashMonitor.LogTrashOperation(trashOp, "允许执行")
	}

	// 3. 检查特殊文件问题
	allowed, err := specialHandler.HandleSpecialFile(filePath, forceMode)
	if err != nil {
		return fmt.Errorf("特殊文件检查失败: %v", err)
	}
	if !allowed {
		return fmt.Errorf("用户取消特殊文件操作")
	}

	return nil
}

func logOperation(operation string, targets []TargetInfo, successCount, failCount int) {
	logFile := filepath.Join(os.TempDir(), "delguard.log")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return // 静默失败，不影响主程序
	}
	defer f.Close()

	timestamp := time.Now().Format(TimeFormatStandard)
	logEntry := fmt.Sprintf("[%s] %s: 成功%d个, 失败%d个, 总计%d个\n",
		timestamp, operation, successCount, failCount, len(targets))

	for _, target := range targets {
		status := "成功"
		if failCount > 0 {
			status = "失败"
		}
		logEntry += fmt.Sprintf("  %s: %s\n", status, target.Path)
	}

	f.WriteString(logEntry)
}

// ContextWithTimeout 创建带超时的上下文
func ContextWithTimeout() (context.Context, context.CancelFunc) {
	if timeout > 0 {
		return context.WithTimeout(context.Background(), timeout)
	}
	return context.WithCancel(context.Background())
}

func main() {
	// 1. 初始化命令检测器并检测当前命令模式
	cmdDetector = NewCommandDetector()
	currentMode = cmdDetector.DetectCommand()

	// 2. 根据检测到的命令模式应用默认设置
	cmdDetector.ApplyModeDefaults()

	// 3. 如果是cp模式，直接处理复制命令
	if currentMode == ModeCP {
		handleCopyCommand()
		return
	}

	// 4. 检查是否为cp命令模式（兼容旧的--cp参数）
	for _, arg := range os.Args[1:] {
		if arg == "--cp" {
			cpMode = true
			currentMode = ModeCP
			handleCopyCommand()
			return
		}
	}

	// 5. 解析命令行参数
	// 首先进行预解析，分离标志参数和文件参数
	preParser := NewSmartArgumentParser(os.Args[1:])
	preResult, err := preParser.ParseArguments()
	if err != nil {
		// 如果智能解析失败，回退到标准解析（但会给出更好的错误信息）
		fmt.Printf(T("参数解析失败: %v\n"), err)
		os.Exit(3)
	}

	// 使用标准flag包解析标志参数
	var configPath string
	flag.BoolVar(&verbose, "v", false, "详细模式")
	flag.BoolVar(&quiet, "q", false, "安静模式")
	flag.BoolVar(&recursive, "r", false, "递归删除目录")
	flag.BoolVar(&dryRun, "n", false, "试运行，不实际删除")
	flag.BoolVar(&force, "force", false, "强制彻底删除，不经过回收站")
	flag.BoolVar(&interactive, "i", false, "交互模式")
	flag.BoolVar(&interactive, "interactive", false, "交互模式") // 支持长参数形式
	flag.BoolVar(&installDefaultInteractive, "default-interactive", false, "安装时将 del/rm 默认指向交互删除")
	flag.BoolVar(&installAliasOnly, "install", false, "安装shell别名（del/rm/cp）")
	flag.BoolVar(&uninstallAliasOnly, "uninstall", false, "卸载已安装的shell别名")
	flag.BoolVar(&showVersion, "version", false, "显示版本")
	flag.BoolVar(&showHelp, "h", false, "显示帮助")
	flag.BoolVar(&showHelp, "help", false, "显示帮助")
	flag.BoolVar(&validateOnly, "validate-only", false, "仅验证文件，不执行删除操作")
	flag.BoolVar(&safeCopy, "safe-copy", false, "安全复制模式") // 新增：安全复制模式
	flag.BoolVar(&protect, "protect", false, "启用文件覆盖保护")
	flag.BoolVar(&disableProtect, "disable-protect", false, "禁用文件覆盖保护")
	flag.DurationVar(&timeout, "timeout", 30*time.Second, "操作超时时间")
	flag.BoolVar(&cpMode, "cp", false, "启用cp命令模式")
	// 智能删除参数
	flag.BoolVar(&smartSearch, "smart-search", true, "启用智能搜索（默认开启）")
	flag.BoolVar(&searchContent, "search-content", false, "搜索文件内容")
	flag.BoolVar(&searchParent, "search-parent", false, "搜索父目录")
	flag.Float64Var(&similarityThreshold, "similarity", 60.0, "相似度阈值（0-100）")
	flag.IntVar(&maxResults, "max-results", 10, "最大搜索结果数量")
	flag.BoolVar(&forceConfirm, "force-confirm", false, "跳过二次确认")
	flag.StringVar(&configPath, "config", "", "指定配置文件路径（支持 .json/.jsonc/.ini/.cfg/.conf/.env/.properties）")

	// 解析标志参数
	flag.CommandLine.Parse(preResult.Flags)

	// 6. 根据检测到的命令模式调整默认参数（用户参数不会被覆盖）
	if !verbose && !quiet { // 只在用户未明确设置时才应用默认值
		switch currentMode {
		case ModeDel:
			// del命令默认显示详细信息
			if !quiet {
				verbose = true
			}
		case ModeRM:
			// rm命令默认安静模式（除非指定-v）
			if !verbose {
				quiet = true
			}
		}
	}

	// 7. 显示命令模式信息（在详细模式下）
	if verbose {
		fmt.Printf(T("DelGuard v%s - 当前模式: %s\n"), version, cmdDetector.GetModeName())
	}

	// 加载配置（支持 --config 覆盖，支持多格式）
	config, err := LoadConfigWithOverride(configPath)
	if err != nil {
		fmt.Printf(T("配置加载失败: %v\n"), err)
		os.Exit(1)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		fmt.Printf(T("配置验证失败: %v\n"), err)
		os.Exit(1)
	}

	// 初始化输入验证器
	inputValidator = NewInputValidator(config)
	// 设置安全级别
	if config.SafeMode == "strict" {
		inputValidator.SetSecurityLevel(SecurityStrict)
	} else if config.SafeMode == "relaxed" {
		inputValidator.SetSecurityLevel(SecurityLow)
	} else {
		inputValidator.SetSecurityLevel(SecurityMedium)
	}

	// 初始化并发管理器
	concurrencyMgr = NewConcurrencyManager(config.MaxConcurrentOps)
	defer func() {
		if err := concurrencyMgr.Close(); err != nil {
			fmt.Printf(T("关闭并发管理器失败: %v\n"), err)
		}
	}()

	// 初始化资源管理器
	resourceMgr = NewResourceManager()
	defer func() {
		if err := resourceMgr.Close(); err != nil {
			fmt.Printf(T("关闭资源管理器失败: %v\n"), err)
		}
	}()

	// 初始化日志
	InitGlobalLogger(config.LogLevel)

	// 初始化特殊文件处理器
	specialHandler := NewSpecialFileHandler(config)

	// 初始化回收站监控器
	trashMonitor := NewTrashOperationMonitor(config)

	// 显示版本
	if showVersion {
		fmt.Printf(T("DelGuard v%s\n"), version)
		return
	}

	// 处理覆盖保护开关
	if protect {
		if err := EnableOverwriteProtection(); err != nil {
			fmt.Printf(T("启用覆盖保护失败: %v\n"), err)
			os.Exit(1)
		}
		fmt.Println(T("✅ 文件覆盖保护已启用"))
		return
	}

	if disableProtect {
		if err := DisableOverwriteProtection(); err != nil {
			fmt.Printf(T("禁用覆盖保护失败: %v\n"), err)
			os.Exit(1)
		}
		fmt.Println(T("⚠️ 文件覆盖保护已禁用"))
		return
	}

	// 显示帮助
	if showHelp {
		printUsage()
		return
	}

	// 卸载别名（优先于安装）
	if uninstallAliasOnly {
		if err := uninstallAliases(); err != nil {
			fmt.Printf(T("别名卸载失败: %v\n"), err)
			os.Exit(1)
		}
		fmt.Println(T("已尝试从当前终端环境卸载 DelGuard 别名。若仍有残留，请重启终端或手动检查配置文件。"))
		return
	}

	// 安装别名
	if installDefaultInteractive || installAliasOnly {
		opts := ParseInstallOptions()
		// 设置语言
		SetLocale(opts.Language)
		// 安装别名
		if err := installAliases(opts.Interactive, opts.Overwrite); err != nil {
			fmt.Printf(T("别名安装失败: %v\n"), err)
			os.Exit(1)
		}
		if opts.Silent {
			fmt.Println(T("已静默安装别名。"))
		} else {
			fmt.Println(T("请新开一个 PowerShell 或 CMD 窗口使用："))
			fmt.Println(T("  del file.txt      # 安全删除文件"))
			fmt.Println(T("  del -i file.txt   # 交互删除"))
			fmt.Println(T("  rm -rf folder     # 递归删除目录"))
			fmt.Println(T("  cp file.txt backup.txt  # 安全复制"))
			fmt.Println(T("  cp -r folder/ backup/   # 递归复制目录"))
			fmt.Println(T("  delguard --help   # 查看帮助"))
		}
		return
	}

	// 恢复文件模式
	if flag.NArg() > 0 && flag.Arg(0) == "restore" {
		pattern := ""
		if flag.NArg() > 1 {
			pattern = flag.Arg(1)
		}

		// 创建恢复子命令的flag
		restoreFlagSet := flag.NewFlagSet("restore", flag.ExitOnError)
		maxFiles := restoreFlagSet.Int("max", 0, "最大恢复文件数")
		interactiveRestore := restoreFlagSet.Bool("i", false, "交互模式确认")
		listOnly := restoreFlagSet.Bool("l", false, "仅列出可恢复文件")

		// 解析恢复参数
		if err := restoreFlagSet.Parse(flag.Args()[1:]); err != nil {
			fmt.Printf(T("恢复参数解析失败: %v\n"), err)
			os.Exit(1)
		}

		// 列出模式
		if *listOnly {
			if err := listRecoverableFiles(pattern); err != nil {
				fmt.Printf(T("列出文件失败: %v\n"), err)
				os.Exit(1)
			}
			return
		}

		opts := RestoreOptions{
			Pattern:     pattern,
			MaxFiles:    *maxFiles,
			Interactive: *interactiveRestore || interactive,
		}

		if err := restoreFromTrash(pattern, opts); err != nil {
			fmt.Printf(T("恢复失败: %v\n"), err)
			os.Exit(1)
		}
		return
	}

	// 正常删除模式
	// 使用预解析的文件列表
	files := preResult.Targets
	if len(files) == 0 {
		printUsage()

		// 显示最近操作日志
		logFile := filepath.Join(os.TempDir(), "delguard.log")
		if data, err := os.ReadFile(logFile); err == nil {
			fmt.Println(T("\n最近操作日志:"))
			lines := strings.Split(string(data), "\n")
			for i, line := range lines {
				if i >= 5 { // 最多显示5条
					break
				}
				if line != "" {
					fmt.Println("  " + line)
				}
			}
		}

		os.Exit(1)
	}

	// 创建带超时的上下文
	ctx, cancel := ContextWithTimeout()
	defer cancel()

	// 启动资源监控
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		monitorResources(ctx)
	}()

	// 预处理：解析所有文件/通配符
	var targets []target
	preErrCount := 0
	processedFiles := make(map[string]bool) // 防止重复处理相同文件

	for _, file := range files {
		// 安全检查：防止路径遍历和注入攻击
		if _, err := sanitizeFileName(file); err != nil {
			dgErr := E(KindInvalidArgs, "validateInput", file, err, "输入路径包含非法字符或格式")
			logger.Error("文件通配符展开失败", "error", dgErr, "pattern")
			preErrCount++
			continue
		}

		// 通配符展开
		expanded, err := filepath.Glob(file)
		if err != nil {
			dgErr := E(KindInvalidArgs, "expandGlob", "", err, "通配符展开失败")
			logger.Error("文件通配符展开失败", "error", dgErr, "pattern")
			preErrCount++
			continue
		}

		// 如果没有通配符匹配但启用了智能搜索，尝试智能搜索
		if len(expanded) == 0 && smartSearch {
			smartFiles, smartErr := enhancedFileResolver(file)
			if smartErr == nil && len(smartFiles) > 0 {
				// 智能搜索成功，使用找到的文件
				for _, smartFile := range smartFiles {
					if processedFiles[smartFile] {
						continue
					}
					processedFiles[smartFile] = true

					smartAbs, absErr := filepath.Abs(smartFile)
					if absErr != nil {
						continue
					}

					_, statErr := os.Stat(smartAbs)
					if statErr == nil {
						// 添加到扩展列表
						expanded = append(expanded, smartFile)
					}
				}
			}
		}

		for _, expFile := range expanded {
			// 检查重复文件
			if processedFiles[expFile] {
				continue
			}
			processedFiles[expFile] = true

			// 解析绝对路径
			abs, err := filepath.Abs(expFile)
			if err != nil {
				dgErr := WrapE("resolveAbsPath", file, err)
				logger.Error("无法解析文件路径", "error", dgErr, "file")
				preErrCount++
				continue
			}

			// 路径验证已在sanitizeFileName中完成

			// 文件存在性检查
			fileInfo, err := os.Stat(abs)
			if err != nil {
				if os.IsNotExist(err) {
					// 尝试智能搜索
					if smartSearch {
						smartFiles, smartErr := enhancedFileResolver(file)
						if smartErr == nil && len(smartFiles) > 0 {
							// 智能搜索成功，使用找到的文件
							for _, smartFile := range smartFiles {
								if processedFiles[smartFile] {
									continue
								}
								processedFiles[smartFile] = true

								smartAbs, absErr := filepath.Abs(smartFile)
								if absErr != nil {
									continue
								}

								smartInfo, statErr := os.Stat(smartAbs)
								if statErr == nil {
									// 添加到目标列表
									targets = append(targets, target{
										arg: filepath.Base(smartFile),
										abs: smartAbs,
									})
									// 需要重新验证智能搜索找到的文件
									fileInfo = smartInfo
									abs = smartAbs
									goto validateSmartFile
								}
							}
							continue
						} else {
							// 智能搜索失败，显示友好提示
							fmt.Printf(T("❌ 文件不存在: %s\n"), file)
							fmt.Printf(T("💡 建议:\n"))
							fmt.Printf(T("   1. 检查文件路径是否正确\n"))
							fmt.Printf(T("   2. 使用 --search-content 搜索文件内容\n"))
							fmt.Printf(T("   3. 使用 --search-parent 搜索父目录\n"))
							fmt.Printf(T("   4. 尝试使用通配符如 *.txt\n"))
							preErrCount++
							continue
						}
					} else {
						// 未启用智能搜索
						fmt.Printf(T("❌ 文件不存在: %s\n"), file)
						fmt.Printf(T("💡 建议: 启用智能搜索 --smart-search 来查找相似文件\n"))
						preErrCount++
						continue
					}
				} else if os.IsPermission(err) {
					fmt.Printf(T("❌ 权限不足: %s\n"), file)
					fmt.Printf(T("💡 建议: 以管理员身份运行或检查文件权限\n"))
					preErrCount++
					continue
				} else {
					logger.Error("无法访问文件", "error", err, "file")
					preErrCount++
					continue
				}
			}

		validateSmartFile:

			// 检查文件类型
			if err := checkFileType(abs); err != nil {
				dgErr := E(KindProtected, "checkFileType", abs, err, "不支持删除特殊文件类型")
				logger.Error("不支持删除特殊文件类型", "error", dgErr, "file")
				preErrCount++
				continue
			}

			// 检查文件权限
			if err := checkFilePermissions(abs, fileInfo); err != nil {
				dgErr := E(KindPermission, "checkPermissions", abs, err, "文件权限检查失败")
				fmt.Printf(T("错误：%s\n"), FormatErrorForDisplay(dgErr))
				preErrCount++
				continue
			}

			// 检查文件大小
			if err := checkFileSize(abs); err != nil {
				dgErr := E(KindInvalidArgs, "checkFileSize", abs, err, "文件大小检查失败")
				fmt.Printf(T("错误：%s\n"), FormatErrorForDisplay(dgErr))
				preErrCount++
				continue
			}

			// 检查磁盘空间
			if !force {
				info, err := os.Stat(abs)
				if err == nil {
					// 只在Windows平台上调用checkDiskSpace
					if runtime.GOOS == "windows" {
						err = checkDiskSpace(abs, info.Size())
						if err != nil {
							dgErr := E(KindIO, "checkDiskSpace", abs, err, "磁盘空间不足")
							fmt.Printf(T("错误：%s\n"), FormatErrorForDisplay(dgErr))
							preErrCount++
							continue
						}
					}
					// 其他平台不检查磁盘空间
				}
			}

			// 人性化文件处理检查
			if err := humanizedFileProcessor(abs, specialHandler, trashMonitor, force); err != nil {
				dgErr := E(KindProtected, "humanizedCheck", abs, err, "人性化检查失败")
				fmt.Printf(T("错误：%s\n"), FormatErrorForDisplay(dgErr))
				preErrCount++
				continue
			}

			// 检查隐藏文件（需要用户确认）
			isHidden, err := isHiddenFile(fileInfo, abs)
			if err != nil {
				dgErr := E(KindIO, "checkHiddenFile", abs, err, "检查隐藏文件失败")
				fmt.Printf(T("错误：%s\n"), FormatErrorForDisplay(dgErr))
				preErrCount++
				continue
			}
			if isHidden && !confirmHiddenFileDeletion(abs) {
				fmt.Printf(T("已跳过隐藏文件: %s\n"), filepath.Base(abs))
				continue
			}

			// 只有当前面没有通过智能搜索添加时才添加
			if !processedFiles[abs] {
				targets = append(targets, target{
					arg: filepath.Base(expFile), // 只存储文件名，避免泄露完整路径
					abs: abs,
				})
			}
		}
	}

	if preErrCount > 0 {
		os.Exit(1)
	}

	if len(targets) == 0 {
		fmt.Println(T("没有找到匹配的文件"))
		return
	}

	// 如果是仅验证模式，则只验证文件不执行删除
	if validateOnly {
		fmt.Println(T("🔍 执行文件验证..."))
		validator := NewFileValidator()
		results, err := validator.ValidateBatch(getTargetPaths(targets))
		if err != nil {
			fmt.Printf(T("验证过程中出错: %v\n"), err)
			os.Exit(1)
		}

		validCount := 0
		for _, result := range results {
			PrintValidationResult(result)
			if result.IsValid {
				validCount++
			}
		}

		fmt.Println(validator.GetValidationSummary(results))
		if validCount != len(results) {
			fmt.Println(T("⚠️  一些文件未通过验证，请检查以上错误"))
			os.Exit(1)
		} else {
			fmt.Println(T("✅ 所有文件都通过了验证"))
		}
		return
	}

	// 执行增强的安全检查
	// 安全检查（已集成到前面的预处理中）
	// 所有安全检查都在预处理阶段完成

	// 最终确认 - 加强安全检查
	if len(targets) > 0 {
		fmt.Printf(T("⚠️  准备删除 %d 个文件/目录:\n"), len(targets))

		// 显示详细信息
		criticalCount := 0
		hiddenCount := 0
		largeCount := 0

		for i, target := range targets {
			info, err := os.Stat(target.abs)
			if err == nil {
				sizeStr := utils.FormatBytes(info.Size())
				isHidden, _ := isHiddenFile(info, target.abs)
				isCritical := IsCriticalPath(target.abs)

				prefix := "  "
				if isCritical {
					prefix = "🔴 "
					criticalCount++
				} else if isHidden {
					prefix = "👁️  "
					hiddenCount++
				} else if info.Size() > 100*1024*1024 { // 100MB
					prefix = "📁 "
					largeCount++
				}

				fmt.Printf("%s%d. %s (%s)", prefix, i+1, target.abs, sizeStr)
				if isHidden {
					fmt.Print(T(" [隐藏]"))
				}
				if isCritical {
					fmt.Print(T(" [系统路径]"))
				}
				fmt.Println()
			} else {
				fmt.Printf(T("  %d. %s (无法获取信息)\n"), i+1, target.abs)
			}
		}

		// 显示警告信息
		if criticalCount > 0 {
			fmt.Printf(T("🚨 警告: 包含 %d 个系统关键路径！\n"), criticalCount)
		}
		if hiddenCount > 0 {
			fmt.Printf(T("👁️  警告: 包含 %d 个隐藏文件！\n"), hiddenCount)
		}
		if largeCount > 0 {
			fmt.Printf(T("📁 警告: 包含 %d 个大文件！\n"), largeCount)
		}

		// 要求用户输入完整确认
		fmt.Printf(T("\n⚠️  此操作将永久删除以上文件！\n"))
		fmt.Print(T("请输入 'YES' 确认删除: "))

		var input string
		if isStdinInteractive() {
			if s, ok := readLineWithTimeout(20 * time.Second); ok {
				input = strings.TrimSpace(strings.ToUpper(s))
			} else {
				input = ""
			}
		} else {
			input = ""
		}

		if input != "YES" {
			fmt.Println(T("操作已取消"))
			return
		}
	}

	// 交互确认与删除
	successCount, failCount := processTargets(targets)

	// 记录操作日志
	var targetInfos []TargetInfo
	for _, t := range targets {
		targetInfos = append(targetInfos, TargetInfo{Path: t.abs})
	}
	logOperation("删除", targetInfos, successCount, failCount)

	// 等待资源监控完成
	cancel()
	wg.Wait()
}

type target struct {
	arg string
	abs string
}

// getTargetPaths 获取目标路径列表
func getTargetPaths(targets []target) []string {
	paths := make([]string, len(targets))
	for i, t := range targets {
		paths[i] = t.abs
	}
	return paths
}

// monitorResources 监控系统资源使用情况
func monitorResources(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			// 检查内存使用
			memUsage := float64(m.Alloc) / 1024 / 1024 // MB
			if memUsage > 512 {                        // 超过512MB
				LogWarn("resource", "memory", fmt.Sprintf("内存使用较高: %.1f MB", memUsage))
			}

			// 检查goroutine数量
			if runtime.NumGoroutine() > 1000 {
				LogWarn("resource", "goroutine", fmt.Sprintf("Goroutine数量较多: %d", runtime.NumGoroutine()))
			}
		}
	}
}

// printUsage 显示使用帮助
func printUsage() {
	fmt.Println(T("DelGuard - 安全删除工具"))
	fmt.Println(T("用法:"))
	fmt.Println(T("  delguard [选项] <文件路径...>"))
	fmt.Println(T("  delguard restore [选项] [模式>"))
	fmt.Println(T("  delguard --cp [选项] <源文件> <目标文件>"))
	fmt.Println()
	fmt.Println(T("主要选项:"))
	fmt.Println(T("  -f, --force        强制删除，跳过确认"))
	fmt.Println(T("  -i, --interactive  交互模式，逐个确认"))
	fmt.Println(T("  -r, --recursive    递归删除目录"))
	fmt.Println(T("  -v, --verbose      详细输出"))
	fmt.Println(T("  --dry-run          仅验证，不实际删除"))
	fmt.Println(T("  --protect          启用文件覆盖保护"))
	fmt.Println(T("  --disable-protect  禁用文件覆盖保护"))
	fmt.Println()
	fmt.Println(T("智能删除选项:"))
	fmt.Println(T("  --smart-search     启用智能搜索（默认开启）"))
	fmt.Println(T("  --search-content   搜索文件内容"))
	fmt.Println(T("  --search-parent    搜索父目录"))
	fmt.Println(T("  --similarity=N     相似度阈值（0-100，默认60）"))
	fmt.Println(T("  --max-results=N    最大搜索结果数（默认10）"))
	fmt.Println(T("  --force-confirm    跳过二次确认"))
	fmt.Println()
	fmt.Println(T("恢复选项:"))
	fmt.Println(T("  -l, --list         仅列出可恢复文件"))
	fmt.Println(T("  -i, --interactive  交互式选择恢复"))
	fmt.Println(T("  --max <数量>      最大恢复文件数"))
	fmt.Println()
	fmt.Println(T("复制选项:"))
	fmt.Println(T("  -r, --recursive    递归复制目录"))
	fmt.Println(T("  -i, --interactive  交互模式"))
	fmt.Println(T("  -f, --force        强制覆盖"))
	fmt.Println(T("  -v, --verbose      详细输出"))
	fmt.Println()
	fmt.Println(T("其他选项:"))
	fmt.Println(T("  -h, --help         显示帮助信息"))
	fmt.Println(T("  -V, --version      显示版本信息"))
	fmt.Println(T("  --install          安装别名（rm/del/cp）"))
	fmt.Println()
	fmt.Println(T("示例:"))
	fmt.Println(T("  delguard file.txt             # 删除文件到回收站"))
	fmt.Println(T("  delguard -f *.tmp             # 强制删除所有.tmp文件"))
	fmt.Println(T("  delguard -i folder/           # 交互式删除目录"))
	fmt.Println(T("  delguard test_fil             # 智能搜索相似文件名"))
	fmt.Println(T("  delguard *.txt --force-confirm # 批量删除跳过确认"))
	fmt.Println(T("  delguard --search-content doc  # 搜索文件内容"))
	fmt.Println(T("  delguard restore file.txt     # 恢复指定文件"))
	fmt.Println(T("  delguard restore -l           # 列出所有可恢复文件"))
	fmt.Println(T("  cp file.txt backup.txt        # 安全复制文件"))
	fmt.Println(T("  cp -r folder/ backup/         # 递归复制目录"))
	fmt.Println()
	fmt.Println(T("注意: DelGuard会将文件移动到系统回收站，不会直接删除。"))
	fmt.Println(T("      cp命令会安全处理文件覆盖，将原文件移入回收站。"))
}

// handleCopyCommand 处理cp命令
func handleCopyCommand() {
	// 创建新的flag set用于cp命令参数解析
	cpFlag := flag.NewFlagSet("cp", flag.ExitOnError)
	var (
		recursive   bool
		interactive bool
		force       bool
		verbose     bool
		preserve    bool
	)

	cpFlag.BoolVar(&recursive, "r", false, "递归复制目录")
	cpFlag.BoolVar(&recursive, "recursive", false, "递归复制目录")
	cpFlag.BoolVar(&interactive, "i", false, "交互模式")
	cpFlag.BoolVar(&interactive, "interactive", false, "交互模式")
	cpFlag.BoolVar(&force, "f", false, "强制覆盖")
	cpFlag.BoolVar(&force, "force", false, "强制覆盖")
	cpFlag.BoolVar(&verbose, "v", false, "详细输出")
	cpFlag.BoolVar(&verbose, "verbose", false, "详细输出")
	cpFlag.BoolVar(&preserve, "p", false, "保留文件属性")
	cpFlag.BoolVar(&preserve, "preserve", false, "保留文件属性")

	// 解析参数
	// 手动解析参数，跳过全局flag
	var cpArgs []string
	foundCp := false

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--cp" {
			foundCp = true
			continue
		}
		if foundCp {
			cpArgs = append(cpArgs, arg)
		}
	}

	// 如果没有找到--cp，检查是否是第一个参数
	if !foundCp && len(os.Args) > 1 {
		if os.Args[1] == "--cp" {
			if len(os.Args) > 2 {
				cpArgs = os.Args[2:]
			}
		}
	}

	// 使用flag包解析cp参数
	if err := cpFlag.Parse(cpArgs); err != nil {
		fmt.Printf(T("参数解析失败: %v\n"), err)
		os.Exit(1)
	}

	// 获取剩余参数（文件路径）
	files := cpFlag.Args()
	if len(files) < 2 {
		fmt.Println(T("用法: cp [选项] 源文件 目标文件"))
		fmt.Println(T("       cp [选项] 源文件... 目标目录"))
		fmt.Println(T("\n选项:"))
		fmt.Println(T("  -r, --recursive    递归复制目录"))
		fmt.Println(T("  -i, --interactive  交互模式"))
		fmt.Println(T("  -f, --force        强制覆盖"))
		fmt.Println(T("  -v, --verbose      详细输出"))
		fmt.Println(T("  -p, --preserve     保留文件属性"))
		os.Exit(1)
	}

	// 创建复制选项
	opts := SafeCopyOptions{
		Interactive: interactive,
		Force:       force,
		Verbose:     verbose,
		Recursive:   recursive,
		Preserve:    preserve,
	}

	// 判断是复制到文件还是目录
	var sources []string
	var dest string

	if len(files) >= 2 {
		dest = files[len(files)-1]
		sources = files[:len(files)-1]
	}

	// 检查目标是否为目录
	destInfo, err := os.Stat(dest)
	isDestDir := err == nil && destInfo.IsDir()

	// 处理多个源文件
	if len(sources) > 1 && !isDestDir {
		fmt.Printf(T("错误: 目标 '%s' 不是目录\n"), dest)
		os.Exit(1)
	}

	successCount := 0
	failCount := 0

	for i, src := range sources {
		var targetPath string
		if isDestDir {
			targetPath = filepath.Join(dest, filepath.Base(src))
		} else {
			targetPath = dest
		}

		if verbose {
			fmt.Printf(T("处理 %d/%d: %s -> %s\n"), i+1, len(sources), src, targetPath)
		}

		// 执行安全复制
		if err := SafeCopy(src, targetPath, opts); err != nil {
			fmt.Printf(T("复制失败: %s\n"), err)
			failCount++
		} else {
			if verbose {
				fmt.Printf(T("✅ 成功复制: %s -> %s\n"), src, targetPath)
			}
			successCount++
		}
	}

	// 显示结果总结
	if verbose || failCount > 0 {
		fmt.Printf(T("\n复制完成: 成功 %d 个，失败 %d 个\n"), successCount, failCount)
	}

	if failCount > 0 {
		os.Exit(1)
	}
}

// processTargets 处理目标文件删除
func processTargets(targets []target) (int, int) {
	successCount := 0
	failCount := 0

	for i, target := range targets {
		fmt.Printf(T("处理 %d/%d: %s\n"), i+1, len(targets), target.abs)

		// 执行删除操作
		if err := moveToTrashPlatform(target.abs); err != nil {
			dgErr := WrapE("moveToTrash", target.abs, err)
			fmt.Printf(T("删除失败: %s\n"), FormatErrorForDisplay(dgErr))
			failCount++
		} else {
			fmt.Printf(T("✅ 成功删除: %s\n"), target.abs)
			successCount++
		}
	}

	// 显示结果总结
	fmt.Printf(T("\n操作完成: 成功 %d 个，失败 %d 个\n"), successCount, failCount)

	return successCount, failCount
}

// 添加缺失的辅助函数
func checkFileType(abs string) error {
	// 简单实现，实际项目中应该根据文件类型进行检查
	info, err := os.Stat(abs)
	if err != nil {
		return err
	}

	// 检查是否为特殊文件类型
	if isSpecialFile(info, abs) {
		return fmt.Errorf("不支持删除特殊文件类型")
	}

	return nil
}
