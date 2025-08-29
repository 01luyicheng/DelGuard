package cmd

import (
	"fmt"
	"runtime"

	"delguard/internal/installer"

	"github.com/spf13/cobra"
)

var (
	forceUninstall bool
	keepConfig     bool
)

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "卸载DelGuard，恢复系统删除命令",
	Long: `卸载DelGuard，恢复系统原始的rm、del等删除命令。

卸载后，删除命令将恢复为系统默认行为（永久删除）。

示例:
  delguard uninstall                # 卸载DelGuard
  delguard uninstall --force        # 强制卸载，不显示确认提示
  delguard uninstall --keep-config  # 卸载但保留配置文件`,
	RunE: runUninstall,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)

	uninstallCmd.Flags().BoolVarP(&forceUninstall, "force", "f", false, "强制卸载，不显示确认提示")
	uninstallCmd.Flags().BoolVar(&keepConfig, "keep-config", false, "保留配置文件和日志")
}

func runUninstall(cmd *cobra.Command, args []string) error {
	fmt.Printf("🗑️ DelGuard 卸载程序\n")
	fmt.Printf("操作系统: %s %s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println()

	// 获取系统安装器
	systemInstaller, err := installer.GetSystemInstaller()
	if err != nil {
		return fmt.Errorf("获取系统安装器失败: %v", err)
	}

	// 检查是否已安装
	if !systemInstaller.IsInstalled() {
		fmt.Println("ℹ️ DelGuard未安装或已被卸载")
		return nil
	}

	// 显示卸载信息
	fmt.Println("📋 卸载信息:")
	fmt.Printf("  安装路径: %s\n", systemInstaller.GetInstallPath())
	fmt.Printf("  目标命令: %v\n", installer.GetTargetCommands())
	if keepConfig {
		fmt.Println("  配置文件: 将保留")
	} else {
		fmt.Println("  配置文件: 将删除")
	}
	fmt.Println()

	// 警告信息
	fmt.Println("⚠️ 警告:")
	fmt.Println("  卸载后，删除命令将恢复为系统默认行为（永久删除）")
	fmt.Println("  请确保您了解这一变化的影响")
	fmt.Println()

	// 确认卸载
	if !forceUninstall {
		fmt.Print("是否继续卸载？ (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" && response != "yes" && response != "YES" {
			fmt.Println("❌ 卸载已取消")
			return nil
		}
	}

	// 执行卸载
	fmt.Println("🔧 开始卸载...")
	if err := systemInstaller.Uninstall(); err != nil {
		return fmt.Errorf("卸载失败: %v", err)
	}

	// 显示卸载后说明
	showPostUninstallInstructions()

	return nil
}

func showPostUninstallInstructions() {
	fmt.Println()
	fmt.Println("✅ 卸载完成！")
	fmt.Println()
	fmt.Println("📝 重要提示:")

	switch runtime.GOOS {
	case "windows":
		fmt.Println("  - 删除命令已恢复为Windows默认行为")
		fmt.Println("  - del、rmdir命令现在将永久删除文件")
		fmt.Println("  - 请重新启动PowerShell以完全清除别名")
	case "darwin", "linux":
		fmt.Println("  - 删除命令已恢复为系统默认行为")
		fmt.Println("  - rm命令现在将永久删除文件")
		fmt.Println("  - 请重新启动终端或重新加载shell配置")
	}

	fmt.Println("  - 如需重新安装，请使用: delguard install")
	fmt.Println()
	fmt.Println("🙏 感谢使用DelGuard！")
}
