package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

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
)

// TargetInfo 用于日志记录
type TargetInfo struct {
	Path string
}

// logOperation 记录操作日志
func logOperation(operation string, targets []TargetInfo, successCount, failCount int) {
	logFile := filepath.Join(os.TempDir(), "delguard.log")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return // 静默失败，不影响主程序
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
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
	// 检查是否为cp命令模式
	for _, arg := range os.Args[1:] {
		if arg == "--cp" {
			cpMode = true
			handleCopyCommand()
			return
		}
	}

	// 解析命令行参数
	flag.BoolVar(&verbose, "v", false, "详细模式")
	flag.BoolVar(&quiet, "q", false, "安静模式")
	flag.BoolVar(&recursive, "r", false, "递归删除目录")
	flag.BoolVar(&dryRun, "n", false, "试运行，不实际删除")
	flag.BoolVar(&force, "force", false, "强制彻底删除，不经过回收站")
	flag.BoolVar(&interactive, "i", false, "交互模式")
	flag.BoolVar(&interactive, "interactive", false, "交互模式") // 支持长参数形式
	flag.BoolVar(&installDefaultInteractive, "default-interactive", false, "安装时将 del/rm 默认指向交互删除")
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

	flag.Parse()

	// 加载配置（返回值用于初始化配置）
	config, err := LoadConfig()
	if err != nil {
		fmt.Printf("配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		fmt.Printf("配置验证失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	InitGlobalLogger(config.LogLevel)

	// 显示版本
	if showVersion {
		fmt.Printf("DelGuard v%s\n", version)
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

	// 安装别名
	if installDefaultInteractive {
		if err := installAliases(installDefaultInteractive); err != nil {
			fmt.Printf(T("参数解析失败: %v\n"), err)
			os.Exit(1)
		}
		fmt.Println(T("请新开一个 PowerShell 或 CMD 窗口使用："))
		fmt.Println(T("  del -i file.txt   # 交互删除"))
		fmt.Println(T("  del -rf folder    # 递归强制删除"))
		fmt.Println(T("  cp file.txt backup.txt  # 安全复制"))
		fmt.Println(T("  cp -r folder/ backup/   # 递归复制目录"))
		fmt.Println(T("  delguard --help   # 查看帮助"))
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
			fmt.Printf("恢复参数解析失败: %v\n", err)
			os.Exit(1)
		}

		// 列出模式
		if *listOnly {
			if err := listRecoverableFiles(pattern); err != nil {
				fmt.Printf("列出文件失败: %v\n", err)
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
			fmt.Printf("恢复失败: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// 正常删除模式
	files := flag.Args()
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
			fmt.Printf(T("错误：%v\n"), FormatErrorForDisplay(dgErr))
			preErrCount++
			continue
		}

		// 通配符展开
		expanded, err := filepath.Glob(file)
		if err != nil {
			dgErr := E(KindInvalidArgs, "expandGlob", "", err, "通配符展开失败")
			fmt.Printf(T("错误：%v\n"), dgErr)
			preErrCount++
			continue
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
				fmt.Printf(T("错误：无法解析路径 %s: %s\n"), file, dgErr.Error())
				preErrCount++
				continue
			}

			// 路径验证已在sanitizeFileName中完成

			// 文件存在性检查
			fileInfo, err := os.Stat(abs)
			if err != nil {
				if os.IsNotExist(err) {
					dgErr := E(KindNotFound, "accessFile", file, err, "文件不存在，请检查路径是否正确")
					fmt.Printf(T("错误：文件不存在 %s: %s\n"), file, FormatErrorForDisplay(dgErr))
				} else if os.IsPermission(err) {
					dgErr := E(KindPermission, "accessFile", file, err, "权限不足，请检查文件权限")
					fmt.Printf(T("错误：权限不足 %s: %s\n"), file, FormatErrorForDisplay(dgErr))
				} else {
					dgErr := WrapE("accessFile", file, err)
					fmt.Printf(T("错误：无法访问 %s: %s\n"), file, FormatErrorForDisplay(dgErr))
				}
				preErrCount++
				continue
			}

			// 检查文件类型
			if err := checkFileType(abs); err != nil {
				dgErr := E(KindProtected, "checkFileType", abs, err, "不支持删除特殊文件类型")
				fmt.Printf(T("错误：%s\n"), FormatErrorForDisplay(dgErr))
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

			targets = append(targets, target{
				arg: filepath.Base(expFile), // 只存储文件名，避免泄露完整路径
				abs: abs,
			})
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
		fmt.Println("🔍 执行文件验证...")
		validator := NewFileValidator()
		results, err := validator.ValidateBatch(getTargetPaths(targets))
		if err != nil {
			fmt.Printf("验证过程中出错: %v\n", err)
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
			fmt.Println("⚠️  一些文件未通过验证，请检查以上错误")
			os.Exit(1)
		} else {
			fmt.Println("✅ 所有文件都通过了验证")
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
				sizeStr := formatBytes(info.Size())
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
					fmt.Print(" [隐藏]")
				}
				if isCritical {
					fmt.Print(" [系统路径]")
				}
				fmt.Println()
			} else {
				fmt.Printf("  %d. %s (无法获取信息)\n", i+1, target.abs)
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

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToUpper(input))

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
	fmt.Println("DelGuard - 安全删除工具")
	fmt.Println("用法:")
	fmt.Println("  delguard [选项] <文件路径...>")
	fmt.Println("  delguard restore [选项] [模式>")
	fmt.Println("  delguard --cp [选项] <源文件> <目标文件>")
	fmt.Println()
	fmt.Println("主要选项:")
	fmt.Println("  -f, --force        强制删除，跳过确认")
	fmt.Println("  -i, --interactive  交互模式，逐个确认")
	fmt.Println("  -r, --recursive    递归删除目录")
	fmt.Println("  -v, --verbose      详细输出")
	fmt.Println("  --dry-run          仅验证，不实际删除")
	fmt.Println("  --protect          启用文件覆盖保护")
	fmt.Println("  --disable-protect  禁用文件覆盖保护")
	fmt.Println()
	fmt.Println("智能删除选项:")
	fmt.Println("  --smart-search     启用智能搜索（默认开启）")
	fmt.Println("  --search-content   搜索文件内容")
	fmt.Println("  --search-parent    搜索父目录")
	fmt.Println("  --similarity=N     相似度阈值（0-100，默认60）")
	fmt.Println("  --max-results=N    最大搜索结果数（默认10）")
	fmt.Println("  --force-confirm    跳过二次确认")
	fmt.Println()
	fmt.Println("恢复选项:")
	fmt.Println("  -l, --list         仅列出可恢复文件")
	fmt.Println("  -i, --interactive  交互式选择恢复")
	fmt.Println("  --max <数量>      最大恢复文件数")
	fmt.Println()
	fmt.Println("复制选项:")
	fmt.Println("  -r, --recursive    递归复制目录")
	fmt.Println("  -i, --interactive  交互模式")
	fmt.Println("  -f, --force        强制覆盖")
	fmt.Println("  -v, --verbose      详细输出")
	fmt.Println()
	fmt.Println("其他选项:")
	fmt.Println("  -h, --help         显示帮助信息")
	fmt.Println("  -V, --version      显示版本信息")
	fmt.Println("  --install          安装别名（rm/del/cp）")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  delguard file.txt             # 删除文件到回收站")
	fmt.Println("  delguard -f *.tmp             # 强制删除所有.tmp文件")
	fmt.Println("  delguard -i folder/           # 交互式删除目录")
	fmt.Println("  delguard test_fil             # 智能搜索相似文件名")
	fmt.Println("  delguard *.txt --force-confirm # 批量删除跳过确认")
	fmt.Println("  delguard --search-content doc  # 搜索文件内容")
	fmt.Println("  delguard restore file.txt     # 恢复指定文件")
	fmt.Println("  delguard restore -l           # 列出所有可恢复文件")
	fmt.Println("  cp file.txt backup.txt        # 安全复制文件")
	fmt.Println("  cp -r folder/ backup/         # 递归复制目录")
	fmt.Println()
	fmt.Println("注意: DelGuard会将文件移动到系统回收站，不会直接删除。")
	fmt.Println("      cp命令会安全处理文件覆盖，将原文件移入回收站。")
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
		fmt.Printf("参数解析失败: %v\n", err)
		os.Exit(1)
	}

	// 获取剩余参数（文件路径）
	files := cpFlag.Args()
	if len(files) < 2 {
		fmt.Println("用法: cp [选项] 源文件 目标文件")
		fmt.Println("       cp [选项] 源文件... 目标目录")
		fmt.Println("\n选项:")
		fmt.Println("  -r, --recursive    递归复制目录")
		fmt.Println("  -i, --interactive  交互模式")
		fmt.Println("  -f, --force        强制覆盖")
		fmt.Println("  -v, --verbose      详细输出")
		fmt.Println("  -p, --preserve     保留文件属性")
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
		fmt.Printf("错误: 目标 '%s' 不是目录\n", dest)
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
			fmt.Printf("处理 %d/%d: %s -> %s\n", i+1, len(sources), src, targetPath)
		}

		// 执行安全复制
		if err := SafeCopy(src, targetPath, opts); err != nil {
			fmt.Printf("复制失败: %s\n", err)
			failCount++
		} else {
			if verbose {
				fmt.Printf("✅ 成功复制: %s -> %s\n", src, targetPath)
			}
			successCount++
		}
	}

	// 显示结果总结
	if verbose || failCount > 0 {
		fmt.Printf("\n复制完成: 成功 %d 个，失败 %d 个\n", successCount, failCount)
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

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
