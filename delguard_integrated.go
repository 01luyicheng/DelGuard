package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// DelGuardApp 主应用程序结构
type DelGuardApp struct {
	config      *SmartConfig
	deleter     *CoreDeleter
	healthCheck *HealthChecker
	verbose     bool
	dryRun      bool
	interactive bool
	force       bool
	recursive   bool
}

// NewDelGuardApp 创建新的应用程序实例
func NewDelGuardApp() *DelGuardApp {
	return &DelGuardApp{}
}

// Initialize 初始化应用程序
func (app *DelGuardApp) Initialize(configPath string) error {
	// 初始化全局错误处理器
	InitGlobalErrorHandler(app.verbose, true, app.verbose)

	// 确定配置文件路径
	if configPath == "" {
		homeDir, _ := os.UserHomeDir()
		configPath = filepath.Join(homeDir, ".delguard", "config.json")
	}

	// 创建智能配置管理器
	app.config = NewSmartConfig(configPath)

	// 加载配置
	if err := app.config.LoadConfig(); err != nil {
		if os.IsNotExist(err) {
			LogInfo("配置文件不存在，使用默认配置")
			LogInfo("运行 'delguard config generate' 创建配置文件")

			// 使用默认配置
			app.config.config = &Config{
				Language:    "zh-cn",
				Verbose:     app.verbose,
				Interactive: app.interactive,
			}
		} else {
			return fmt.Errorf("加载配置失败: %v", err)
		}
	}

	// 创建核心删除器
	app.deleter = NewCoreDeleter(app.config.config)

	// 创建健康检查器
	app.healthCheck = &HealthChecker{}

	LogDebug("应用程序初始化完成", "")
	return nil
}

// SetOptions 设置应用程序选项
func (app *DelGuardApp) SetOptions(verbose, dryRun, interactive, force, recursive bool) {
	app.verbose = verbose
	app.dryRun = dryRun
	app.interactive = interactive
	app.force = force
	app.recursive = recursive

	if app.deleter != nil {
		app.deleter.SetOptions(dryRun, interactive, force, recursive, verbose)
	}
}

// RunDelete 执行删除操作
func (app *DelGuardApp) RunDelete(paths []string) error {
	if len(paths) == 0 {
		return fmt.Errorf("没有指定要删除的文件或目录")
	}

	LogInfo(fmt.Sprintf("开始删除操作，目标数量: %d", len(paths)))

	// 执行删除
	results := app.deleter.Delete(paths)

	// 显示结果
	app.printResults(results)

	// 显示统计信息
	if app.verbose {
		app.deleter.PrintStats()
	}

	return nil
}

// RunHealthCheck 执行健康检查
func (app *DelGuardApp) RunHealthCheck() error {
	LogInfo("开始系统健康检查")
	return CheckSystemHealth()
}

// RunConfigGeneration 运行配置生成
func (app *DelGuardApp) RunConfigGeneration() error {
	LogInfo("启动交互式配置生成器")
	return RunInteractiveConfigGenerator()
}

// printResults 打印删除结果
func (app *DelGuardApp) printResults(results []DeleteResult) {
	if len(results) == 0 {
		fmt.Println("没有文件被处理")
		return
	}

	fmt.Println("\n📋 删除结果:")
	fmt.Println(strings.Repeat("─", 60))

	successCount := 0
	errorCount := 0
	skipCount := 0

	for _, result := range results {
		status := "❌"
		if result.Success {
			status = "✅"
			successCount++
		} else if result.Skipped {
			status = "⏭️ "
			skipCount++
		} else {
			errorCount++
		}

		fileType := "📄"
		if result.IsDirectory {
			fileType = "📁"
		}

		fmt.Printf("%s %s %s", status, fileType, result.Path)

		if result.Error != nil {
			fmt.Printf(" - 错误: %v", result.Error)
		} else if result.Skipped {
			fmt.Printf(" - 跳过: %s", result.Reason)
		} else if result.Success && app.verbose {
			if result.Duration > 0 {
				fmt.Printf(" - 耗时: %v", result.Duration)
			}
		}

		fmt.Println()
	}

	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("📊 总计: 成功 %d, 跳过 %d, 错误 %d\n", successCount, skipCount, errorCount)
}

// Cleanup 清理资源
func (app *DelGuardApp) Cleanup() {
	if GlobalErrorHandler != nil {
		GlobalErrorHandler.Close()
	}
	LogDebug("应用程序清理完成", "")
}

// 全局应用程序实例
var globalApp *DelGuardApp

// 命令行变量
var (
	configFile  string
	verbose     bool
	dryRun      bool
	interactive bool
	force       bool
	recursive   bool
)

// 根命令
var rootCmd = &cobra.Command{
	Use:   "delguard [files/directories...]",
	Short: "DelGuard - 智能文件删除工具",
	Long: `DelGuard 是一个智能的文件删除工具，具有以下特性：

🔍 智能路径识别和参数解析
⚙️  交互式配置生成  
🛡️  基本安全保护（非过度设计）
📊 详细的操作统计
🔧 系统健康检查

使用示例:
  delguard file.txt                    # 删除单个文件
  delguard -r directory/               # 递归删除目录
  delguard -i *.tmp                    # 交互式删除临时文件
  delguard --dry-run file.txt          # 干运行模式
  delguard config generate             # 生成配置文件
  delguard health                      # 系统健康检查`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}

		// 初始化应用程序
		globalApp = NewDelGuardApp()
		globalApp.SetOptions(verbose, dryRun, interactive, force, recursive)

		if err := globalApp.Initialize(configFile); err != nil {
			return fmt.Errorf("初始化失败: %v", err)
		}

		defer globalApp.Cleanup()

		// 执行删除操作
		return globalApp.RunDelete(args)
	},
}

// 配置命令
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "配置管理",
	Long:  "管理DelGuard的配置文件",
}

var generateConfigCmd = &cobra.Command{
	Use:   "generate",
	Short: "生成配置文件",
	Long:  "交互式生成DelGuard配置文件",
	RunE: func(cmd *cobra.Command, args []string) error {
		globalApp = NewDelGuardApp()
		globalApp.SetOptions(verbose, false, false, false, false)

		if err := globalApp.Initialize(configFile); err != nil {
			return fmt.Errorf("初始化失败: %v", err)
		}

		defer globalApp.Cleanup()

		return globalApp.RunConfigGeneration()
	},
}

// 健康检查命令
var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "系统健康检查",
	Long:  "检查DelGuard组件和配置文件的健康状态",
	RunE: func(cmd *cobra.Command, args []string) error {
		globalApp = NewDelGuardApp()
		globalApp.SetOptions(verbose, false, false, false, false)

		if err := globalApp.Initialize(configFile); err != nil {
			return fmt.Errorf("初始化失败: %v", err)
		}

		defer globalApp.Cleanup()

		return globalApp.RunHealthCheck()
	},
}

// 版本命令
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Long:  "显示DelGuard的版本信息",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("DelGuard v2.0.0")
		fmt.Println("智能文件删除工具")
		fmt.Println("构建时间:", time.Now().Format("2006-01-02"))
		fmt.Println("Go版本: 1.19+")
	},
}

func init() {
	// 根命令标志
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "配置文件路径")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "详细输出")
	rootCmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "干运行模式（不实际删除）")
	rootCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "交互式确认")
	rootCmd.Flags().BoolVarP(&force, "force", "f", false, "强制删除（跳过安全检查）")
	rootCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "递归删除目录")

	// 添加子命令
	configCmd.AddCommand(generateConfigCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(versionCmd)
}

// 主函数入口
func RunDelGuardApp() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ 错误: %v\n", err)
		os.Exit(1)
	}
}
