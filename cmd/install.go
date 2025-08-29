package cmd

import (
	"fmt"
	"runtime"

	"delguard/internal/installer"

	"github.com/spf13/cobra"
)

var (
	systemWide   bool
	forceInstall bool
	installPath  string
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "安装DelGuard，替换系统删除命令",
	Long: `安装DelGuard到系统中，替换rm、del等删除命令。

安装后，当您使用rm、del等命令删除文件时，文件将被安全地移动到回收站而不是永久删除。

示例:
  delguard install                    # 用户级安装
  delguard install --system           # 系统级安装（需要管理员权限）
  delguard install --path /custom/path # 自定义安装路径
  delguard install --force            # 强制安装，覆盖现有安装`,
	RunE: runInstall,
}

func init() {
	rootCmd.AddCommand(installCmd)

	installCmd.Flags().BoolVarP(&systemWide, "system", "s", false, "系统级安装（需要管理员权限）")
	installCmd.Flags().BoolVarP(&forceInstall, "force", "f", false, "强制安装，覆盖现有安装")
	installCmd.Flags().StringVarP(&installPath, "path", "p", "", "自定义安装路径")
}

func runInstall(cmd *cobra.Command, args []string) error {
	fmt.Printf("🚀 DelGuard 安装程序\n")
	fmt.Printf("操作系统: %s %s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println()

	// 获取系统安装器
	systemInstaller, err := installer.GetSystemInstaller()
	if err != nil {
		return fmt.Errorf("获取系统安装器失败: %v", err)
	}

	// 检查是否已安装
	if systemInstaller.IsInstalled() && !forceInstall {
		fmt.Println("⚠️ DelGuard已经安装")
		fmt.Printf("安装路径: %s\n", systemInstaller.GetInstallPath())
		fmt.Println("如需重新安装，请使用 --force 参数")
		return nil
	}

	// 检查权限
	if systemWide && !installer.IsRunningAsAdmin() {
		return fmt.Errorf("❌ 系统级安装需要管理员权限\n" +
			"请以管理员身份运行此命令")
	}

	// 显示安装信息
	config := installer.GetDefaultInstallConfig()
	if installPath != "" {
		config.InstallPath = installPath
	}
	config.SystemWide = systemWide
	config.ForceInstall = forceInstall

	fmt.Println("📋 安装配置:")
	fmt.Printf("  安装类型: %s\n", getInstallType(systemWide))
	fmt.Printf("  安装路径: %s\n", config.InstallPath)
	fmt.Printf("  备份路径: %s\n", config.BackupPath)
	fmt.Printf("  目标命令: %v\n", installer.GetTargetCommands())
	fmt.Println()

	// 确认安装
	if !forceInstall {
		fmt.Print("是否继续安装？ (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" && response != "yes" && response != "YES" {
			fmt.Println("❌ 安装已取消")
			return nil
		}
	}

	// 执行安装
	fmt.Println("🔧 开始安装...")
	if err := systemInstaller.Install(); err != nil {
		return fmt.Errorf("安装失败: %v", err)
	}

	// 显示安装后说明
	showPostInstallInstructions()

	return nil
}

func getInstallType(systemWide bool) string {
	if systemWide {
		return "系统级安装"
	}
	return "用户级安装"
}

func showPostInstallInstructions() {
	fmt.Println()
	fmt.Println("🎉 安装完成！")
	fmt.Println()
	fmt.Println("📝 使用说明:")
	fmt.Println("  现在您可以使用以下命令安全删除文件:")

	switch runtime.GOOS {
	case "windows":
		fmt.Println("    del file.txt        # 删除文件到回收站")
		fmt.Println("    rmdir folder        # 删除目录到回收站")
		fmt.Println("    delguard list       # 查看回收站文件")
		fmt.Println("    delguard restore    # 恢复文件")
	case "darwin", "linux":
		fmt.Println("    rm file.txt         # 删除文件到回收站")
		fmt.Println("    rm -r folder        # 删除目录到回收站")
		fmt.Println("    delguard list       # 查看回收站文件")
		fmt.Println("    delguard restore    # 恢复文件")
	}

	fmt.Println()
	fmt.Println("⚠️ 重要提示:")
	fmt.Println("  - 请重新启动终端或重新加载配置文件")
	fmt.Println("  - 如需永久删除文件，请使用: delguard delete --permanent")
	fmt.Println("  - 如需卸载，请使用: delguard uninstall")

	if runtime.GOOS == "windows" {
		fmt.Println("  - PowerShell用户请重新启动PowerShell")
		fmt.Println("  - 或运行: . $PROFILE 重新加载配置")
	}
}
