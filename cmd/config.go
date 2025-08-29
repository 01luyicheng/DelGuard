package cmd

import (
	"fmt"

	"delguard/internal/config"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "配置管理",
	Long:  "管理DelGuard的配置选项",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "显示当前配置",
	Run: func(cmd *cobra.Command, args []string) {
		showConfig()
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "设置配置项",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		setConfig(args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
}

func showConfig() {
	if config.GlobalConfig == nil {
		fmt.Println("❌ 配置未初始化")
		return
	}

	cfg := config.GlobalConfig
	fmt.Println("📋 DelGuard 当前配置:")
	fmt.Println()
	fmt.Println("🗑️  回收站设置:")
	fmt.Printf("   自动清理: %v\n", cfg.Trash.AutoClean)
	fmt.Printf("   保留天数: %d 天\n", cfg.Trash.MaxDays)
	fmt.Println()
	fmt.Println("📝 日志设置:")
	fmt.Printf("   日志级别: %s\n", cfg.Logging.Level)
	fmt.Printf("   日志文件: %s\n", cfg.Logging.File)
	fmt.Printf("   文件大小: %d MB\n", cfg.Logging.MaxSize)
	fmt.Printf("   保留天数: %d 天\n", cfg.Logging.MaxAge)
	fmt.Printf("   压缩存储: %v\n", cfg.Logging.Compress)
	fmt.Println()
	fmt.Println("🎨 界面设置:")
	fmt.Printf("   语言: %s\n", cfg.UI.Language)
	fmt.Printf("   彩色输出: %v\n", cfg.UI.Color)
}

func setConfig(key, value string) {
	if config.GlobalConfig == nil {
		fmt.Println("❌ 配置未初始化")
		return
	}

	switch key {
	case "trash.auto_clean":
		if value == "true" {
			config.GlobalConfig.Trash.AutoClean = true
		} else {
			config.GlobalConfig.Trash.AutoClean = false
		}
		fmt.Printf("✅ 已设置 %s = %s\n", key, value)
	case "ui.language":
		config.GlobalConfig.UI.Language = value
		fmt.Printf("✅ 已设置 %s = %s\n", key, value)
	case "ui.color":
		if value == "true" {
			config.GlobalConfig.UI.Color = true
		} else {
			config.GlobalConfig.UI.Color = false
		}
		fmt.Printf("✅ 已设置 %s = %s\n", key, value)
	default:
		fmt.Printf("❌ 未知的配置项: %s\n", key)
		fmt.Println("支持的配置项:")
		fmt.Println("  trash.auto_clean  - 自动清理回收站 (true/false)")
		fmt.Println("  ui.language       - 界面语言 (zh/en)")
		fmt.Println("  ui.color          - 彩色输出 (true/false)")
	}
}
