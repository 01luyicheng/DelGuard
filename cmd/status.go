package cmd

import (
	"fmt"
	"runtime"

	"delguard/internal/filesystem"

	"github.com/spf13/cobra"
)

// statusCmd 状态命令
var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"info", "stat"},
	Short:   "显示DelGuard状态信息",
	Long: `显示DelGuard的运行状态和系统信息。

包括:
• 系统信息
• 回收站路径和状态
• 文件统计信息
• 版本信息

示例:
  delguard status
  delguard info    # 别名
  delguard stat    # 别名`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)

	// 添加标志
	statusCmd.Flags().BoolP("detailed", "d", false, "显示详细信息")
}

func runStatus(cmd *cobra.Command, args []string) error {
	detailed, _ := cmd.Flags().GetBool("detailed")

	fmt.Println("🛡️  DelGuard 状态信息")
	fmt.Println("=" + string(make([]rune, 50)))

	// 系统信息
	fmt.Printf("🖥️  操作系统: %s %s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("🏗️  Go版本: %s\n", runtime.Version())
	fmt.Printf("📦 DelGuard版本: %s\n", rootCmd.Version)

	// 获取回收站管理器
	manager, err := filesystem.GetTrashManager()
	if err != nil {
		fmt.Printf("❌ 回收站管理器初始化失败: %v\n", err)
		return nil
	}

	// 回收站路径
	trashPath, err := manager.GetTrashPath()
	if err != nil {
		fmt.Printf("❌ 获取回收站路径失败: %v\n", err)
	} else {
		fmt.Printf("🗑️  回收站路径: %s\n", trashPath)
	}

	// 回收站统计信息
	trashFiles, err := manager.ListTrashFiles()
	if err != nil {
		fmt.Printf("❌ 获取回收站文件列表失败: %v\n", err)
	} else {
		fileCount := 0
		dirCount := 0
		totalSize := int64(0)

		for _, file := range trashFiles {
			if file.IsDirectory {
				dirCount++
			} else {
				fileCount++
			}
			totalSize += file.Size
		}

		fmt.Printf("📊 回收站统计:\n")
		fmt.Printf("   • 文件数量: %d\n", fileCount)
		fmt.Printf("   • 目录数量: %d\n", dirCount)
		fmt.Printf("   • 总计大小: %s\n", filesystem.FormatFileSize(totalSize))

		if detailed && len(trashFiles) > 0 {
			fmt.Printf("\n📋 最近删除的文件:\n")
			displayCount := len(trashFiles)
			if displayCount > 5 {
				displayCount = 5
			}

			for i := 0; i < displayCount; i++ {
				file := trashFiles[i]
				typeIcon := "📄"
				if file.IsDirectory {
					typeIcon = "📁"
				}
				fmt.Printf("   %s %s (%s, %s)\n",
					typeIcon, file.Name,
					filesystem.FormatFileSize(file.Size),
					file.DeletedTime.Format("2006-01-02 15:04"))
			}

			if len(trashFiles) > 5 {
				fmt.Printf("   ... 还有 %d 个项目\n", len(trashFiles)-5)
			}
		}
	}

	// 系统集成状态
	fmt.Printf("\n🔧 系统集成状态:\n")

	// 检查命令别名状态（这里先显示占位信息，后续会在安装功能中实现）
	switch runtime.GOOS {
	case "windows":
		fmt.Printf("   • del命令替换: 未安装\n")
		fmt.Printf("   • PowerShell集成: 未安装\n")
	case "darwin", "linux":
		fmt.Printf("   • rm命令替换: 未安装\n")
		fmt.Printf("   • Shell集成: 未安装\n")
	}

	fmt.Printf("\n💡 提示: 使用 'delguard install' 安装系统集成\n")

	return nil
}
