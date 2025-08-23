package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

func main() {
	// 解析命令行参数
	flag.BoolVar(&verbose, "v", false, "详细模式")
	flag.BoolVar(&quiet, "q", false, "安静模式")
	flag.BoolVar(&recursive, "r", false, "递归删除目录")
	flag.BoolVar(&dryRun, "n", false, "试运行，不实际删除")
	flag.BoolVar(&force, "force", false, "强制彻底删除，不经过回收站")
	flag.BoolVar(&interactive, "i", false, "交互模式")
	flag.BoolVar(&installDefaultInteractive, "install", false, "安装别名（默认启用交互模式）")
	flag.BoolVar(&showVersion, "version", false, "显示版本")
	flag.BoolVar(&showHelp, "help", false, "显示帮助")

	flag.Parse()

	// 加载配置（返回值用于初始化配置）
	_ = LoadConfig()

	// 显示版本
	if showVersion {
		fmt.Printf("DelGuard v%s\n", version)
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
		return
	}

	// 恢复文件模式
	if flag.NArg() > 0 && flag.Arg(0) == "restore" {
		pattern := ""
		if flag.NArg() > 1 {
			pattern = flag.Arg(1)
		}
		opts := RestoreOptions{
			Pattern: pattern,
		}
		if err := restoreFromTrash(pattern, opts); err != nil {
			fmt.Printf(T("恢复失败: %v\n"), err)
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

	// 预处理：解析所有文件/通配符
	var targets []target
	preErrCount := 0
	processedFiles := make(map[string]bool) // 防止重复处理相同文件

	for _, file := range files {
		// 安全检查：防止路径遍历和注入攻击
		if !validateInputPath(file) {
			dgErr := E(KindInvalidArgs, "validateInput", file, nil, "输入路径包含非法字符或格式")
			fmt.Printf(T("错误：%v\n"), dgErr)
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

			// 路径验证
			if !validatePath(abs) {
				dgErr := E(KindInvalidArgs, "validatePath", file, nil, "路径包含无效字符或格式")
				fmt.Printf(T("错误：路径无效 %s: %s\n"), file, dgErr.Error())
				preErrCount++
				continue
			}

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

			// 检查特殊文件类型（符号链接、设备文件等）
			if isSpecialFile(fileInfo) {
				fileType := "特殊文件"
				if fileInfo.Mode()&os.ModeSymlink != 0 {
					fileType = "符号链接"
				} else if fileInfo.Mode()&os.ModeDevice != 0 {
					fileType = "设备文件"
				} else if fileInfo.Mode()&os.ModeSocket != 0 {
					fileType = "套接字文件"
				} else if fileInfo.Mode()&os.ModeNamedPipe != 0 {
					fileType = "命名管道"
				}
				dgErr := E(KindProtected, "checkFileType", abs, nil, fmt.Sprintf("不支持删除%s类型", fileType))
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
			if err := checkDiskSpace(abs); err != nil {
				dgErr := E(KindIO, "checkDiskSpace", abs, err, "磁盘空间检查失败")
				fmt.Printf(T("错误：%s\n"), FormatErrorForDisplay(dgErr))
				preErrCount++
				continue
			}

			// 检查隐藏文件（需要用户确认）
			if isHiddenFile(fileInfo, abs) && !confirmHiddenFileDeletion(abs) {
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

	// 执行增强的安全检查
	// 安全检查（已集成到前面的预处理中）
	// 所有安全检查都在预处理阶段完成

	// 批量操作确认
	if len(targets) > 10 {
		fmt.Printf(T("警告：将删除%d个文件。是否继续？(y/N): "), len(targets))
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println(T("操作已取消"))
			return
		}
	} else if len(targets) > 0 {
		// 显示将要删除的文件列表
		fmt.Printf(T("将要删除以下%d个文件：\n"), len(targets))
		for i, target := range targets {
			info, err := os.Stat(target.abs)
			fileType := T("文件")
			if err == nil && info.IsDir() {
				fileType = T("目录")
			}
			fmt.Printf("  %d. %s (%s)\n", i+1, target.arg, fileType)
		}
		fmt.Printf(T("是否确认删除？(y/N): "))
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
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
}

type target struct {
	arg string
	abs string
}

func processTargets(tgs []target) (successCount, failCount int) {
	// 统计目录数量
	dirCount := 0
	for _, tg := range tgs {
		info, err := os.Stat(tg.abs)
		if err == nil && info.IsDir() {
			dirCount++
		}
	}

	// 安全检查：检查关键路径
	for _, tg := range tgs {
		if IsCriticalPath(tg.abs) {
			if !ConfirmCritical(tg.abs) {
				fmt.Printf(T("已取消关键路径 %s 的删除\n"), tg.abs)
				return
			}
		}
	}

	// 权限检查
	for _, tg := range tgs {
		if !CheckDeletePermission(tg.abs) {
			fmt.Printf(T("已跳过 %s\n"), tg.arg)
			return
		}
	}

	// 交互模式处理
	if interactive || GetInteractiveDefault() {
		reader := bufio.NewReader(os.Stdin)
		mode := "i" // 默认逐项确认

		// 批量模式选择
		if len(tgs) > 1 {
			fmt.Printf(T("即将删除 %d 个目标（其中目录 %d 个）。选择模式 [a]全部同意/[n]全部拒绝/[i]逐项/[q]退出 (默认 i): "), len(tgs), dirCount)
			line, _ := reader.ReadString('\n')
			line = strings.TrimSpace(strings.ToLower(line))
			if line == "a" {
				mode = "a"
			} else if line == "n" {
				fmt.Println(T("已取消所有删除。"))
				return
			} else if line == "q" {
				return
			} else if line != "" {
				mode = line
			}
		}

		// 处理删除
		for i, tg := range tgs {
			remaining := len(tgs) - i - 1

			if mode == "a" {
				// 全部同意，直接删除
				if err := deleteTarget(tg); err != nil {
					fmt.Printf(T("错误：无法删除 %s: %v\n"), tg.arg, err)
				}
			} else if mode == "i" {
				// 逐项确认
				info, err := os.Stat(tg.abs)
				fileType := T("文件")
				if err == nil && info.IsDir() {
					fileType = T("目录")
				}
				fmt.Printf(T("计划删除：%s (绝对路径: %s, 类型: %s)\n"), tg.arg, tg.abs, fileType)

				if remaining > 0 {
					fmt.Printf(T("删除 %s ? [y/N/a/q]: "), tg.arg)
				} else {
					// 最后一个文件或单文件，不显示a选项
					fmt.Printf(T("删除 %s ? [y/N/q]: "), tg.arg)
				}

				line, _ := reader.ReadString('\n')
				line = strings.TrimSpace(strings.ToLower(line))

				switch line {
				case "y", "yes":
					if err := deleteTarget(tg); err != nil {
						fmt.Printf(T("错误：无法删除 %s: %v\n"), tg.arg, err)
					}
				case "a":
					mode = "a" // 切换到全部同意模式
					if err := deleteTarget(tg); err != nil {
						fmt.Printf(T("错误：无法删除 %s: %v\n"), tg.arg, err)
					}
					// 继续处理剩余文件
					continue
				case "q":
					return
				default:
					fmt.Printf(T("已跳过 %s\n"), tg.arg)
				}
			}
		}
	} else {
		// 非交互模式，直接删除
		for _, tg := range tgs {
			if err := deleteTarget(tg); err != nil {
				fmt.Printf(T("错误：无法删除 %s: %v\n"), tg.arg, err)
			}
		}
	}
	return
}

func deleteTarget(tg target) error {
	if dryRun {
		fmt.Printf(T("[DRY-RUN] 将把 %s 移动到回收站\n"), tg.arg)
		return nil
	}

	// 安全检查
	if IsCriticalPath(tg.abs) && !ConfirmCritical(tg.abs) {
		return fmt.Errorf(T("用户取消了关键路径删除"))
	}

	// 权限检查
	if !CheckDeletePermission(tg.abs) {
		return fmt.Errorf(T("权限检查失败"))
	}

	// 执行删除
	err := moveToTrash(tg.abs)
	if err != nil {
		return err
	}

	if !quiet {
		fmt.Printf(T("已将 %s 移动到回收站\n"), tg.arg)
	}
	return
}

// checkSystemResources 检查系统资源
func checkSystemResources() error {
	// 检查内存使用情况
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// 如果内存使用超过80%，警告用户
	const maxMemoryUsage = 0.8
	if float64(m.Alloc)/float64(m.Sys) > maxMemoryUsage {
		return fmt.Errorf("系统内存使用过高（%.1f%%），可能影响操作性能", 
			float64(m.Alloc)/float64(m.Sys)*100)
	}
	
	// 检查CPU使用率（简化实现）
	// 这里我们检查是否有过多的goroutine
	if runtime.NumGoroutine() > 1000 {
		return fmt.Errorf("系统并发过高（%d个goroutine），可能影响稳定性", runtime.NumGoroutine())
	}
	
	return nil
}

// monitorOperation 监控操作过程
func monitorOperation(targets []target) {
	startTime := time.Now()
	fmt.Printf(T("开始监控操作，共%d个目标...\n"), len(targets))
	
	// 创建监控goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				
				// 检查内存使用
				memUsage := float64(m.Alloc) / 1024 / 1024 // MB
				fmt.Printf(T("[监控] 内存使用: %.1f MB, Goroutines: %d\n"), 
					memUsage, runtime.NumGoroutine())
				
				// 如果操作时间过长，发出警告
				if time.Since(startTime) > 5*time.Minute {
					fmt.Println(T("[警告] 操作执行时间过长，请检查系统状态"))
				}
			}
		}
	}()
}

// moveToTrash 将文件移动到回收站的主函数
func moveToTrash(filePath string) error {
	// 如果启用了强制删除模式，或者配置禁用了回收站功能，则直接彻底删除文件
	if force || !GetUseTrash() {
		return deletePermanently(filePath)
	}
	// 否则使用回收站功能
	return moveToTrashPlatform(filePath)
}

// deletePermanently 直接彻底删除文件或目录
func deletePermanently(filePath string) error {
	// 获取文件信息
	info, err := os.Stat(filePath)
	if err != nil {
		return WrapE("statFile", filePath, err)
	}

	if info.IsDir() {
		// 递归删除目录
		return os.RemoveAll(filePath)
	} else {
		// 删除单个文件
		return os.Remove(filePath)
	}
}

func printUsage() {
	cmd := "delguard"
	if runtime.GOOS == "windows" {
		cmd = "delguard.exe"
	}

	block := T(`DelGuard v%s - 跨平台安全删除工具

用法:
  %s [选项] <文件或目录...>
  %s restore [<通配符>]

选项:
  -v, --verbose           详细模式
  -q, --quiet             安静模式，减少输出
  -r, --recursive         递归删除目录
  -n, --dry-run           试运行，不实际删除
  --force                 强制彻底删除，不经过回收站
  -i, --interactive       交互模式，逐项确认
  --install               安装shell别名（默认启用交互模式）
  --version               显示版本信息
  --help                  显示此帮助信息

示例:
  %s file.txt             删除单个文件
  %s -r directory        递归删除目录
  %s -i *.txt           交互式删除所有txt文件
  %s --force old_file    彻底删除文件（不经过回收站）
  %s restore *.txt      恢复删除的txt文件

安全特性:
  - 防止删除系统关键目录
  - 防止删除回收站/废纸篓
  - 防止删除包含DelGuard程序的目录
  - 高权限操作需要额外确认
  - 可配置的交互确认机制

配置:
  用户配置存储在 ~/.delguard/config.toml
  可配置默认交互模式、语言选项等`)

	fmt.Printf(block, version, cmd, cmd, cmd, cmd, cmd, cmd, cmd, cmd)
}
