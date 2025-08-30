package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd 根命令
var rootCmd = &cobra.Command{
	Use:   "delguard",
	Short: "DelGuard - 跨平台安全删除工具",
	Long: `DelGuard 是一款跨平台的安全删除工具，可以将删除的文件移动到回收站而不是直接删除。

支持的功能：
• 安全删除文件和目录
• 查看回收站内容
• 恢复已删除的文件
• 清空回收站
• 跨平台支持 (Windows/macOS/Linux)`,
	Version: "1.5.3",
}

// Execute 执行根命令
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// 全局标志
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径 (默认: $HOME/.delguard.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "详细输出")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "静默模式")

	// 绑定标志到viper
	if err := viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose")); err != nil {
		log.Printf("绑定verbose标志失败: %v", err)
	}
	if err := viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet")); err != nil {
		log.Printf("绑定quiet标志失败: %v", err)
	}
}

// initConfig 初始化配置
func initConfig() {
	if cfgFile != "" {
		// 使用指定的配置文件
		viper.SetConfigFile(cfgFile)
	} else {
		// 查找home目录
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// 在home目录中查找".delguard"配置文件
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".delguard")
	}

	// 读取环境变量
	viper.AutomaticEnv()

	// 如果找到配置文件，则读取它
	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Fprintln(os.Stderr, "使用配置文件:", viper.ConfigFileUsed())
		}
	}
}
