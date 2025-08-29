package cmd

import (
	"fmt"

	"delguard/internal/filesystem"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// emptyCmd 清空回收站命令
var emptyCmd = &cobra.Command{
	Use:     "empty",
	Aliases: []string{"clear", "purge"},
	Short:   "清空回收站",
	Long: `永久删除回收站中的所有文件和目录。

⚠️  警告: 此操作不可逆，清空后的文件无法恢复！

示例:
  delguard empty
  delguard empty --force    # 跳过确认提示
  delguard clear            # 别名
  delguard purge            # 别名`,
	RunE: runEmpty,
}

func init() {
	rootCmd.AddCommand(emptyCmd)

	// 添加标志
	emptyCmd.Flags().BoolP("force", "f", false, "强制清空，不显示确认提示")
	emptyCmd.Flags().BoolP("dry-run", "n", false, "预览模式，显示将要删除的文件但不实际删除")
}

func runEmpty(cmd *cobra.Command, args []string) error {
	// 获取标志值
	force, _ := cmd.Flags().GetBool("force")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	quiet := viper.GetBool("quiet")

	// 获取回收站管理器
	manager, err := filesystem.GetTrashManager()
	if err != nil {
		return fmt.Errorf("初始化回收站管理器失败: %v", err)
	}

	// 获取回收站文件列表（用于统计和预览）
	trashFiles, err := manager.ListTrashFiles()
	if err != nil {
		return fmt.Errorf("获取回收站文件列表失败: %v", err)
	}

	if len(trashFiles) == 0 {
		if !quiet {
			fmt.Println("🗑️  回收站已经是空的")
		}
		return nil
	}

	// 计算总大小
	totalSize := int64(0)
	for _, file := range trashFiles {
		totalSize += file.Size
	}

	// 预览模式
	if dryRun {
		fmt.Printf("🔍 预览模式 - 将要永久删除 %d 个项目 (%s):\n",
			len(trashFiles), filesystem.FormatFileSize(totalSize))

		// 显示前10个文件
		displayCount := len(trashFiles)
		if displayCount > 10 {
			displayCount = 10
		}

		for i := 0; i < displayCount; i++ {
			file := trashFiles[i]
			typeIcon := "📄"
			if file.IsDirectory {
				typeIcon = "📁"
			}
			fmt.Printf("  %s %s (%s)\n", typeIcon, file.Name, filesystem.FormatFileSize(file.Size))
		}

		if len(trashFiles) > 10 {
			fmt.Printf("  ... 还有 %d 个项目\n", len(trashFiles)-10)
		}

		return nil
	}

	// 显示警告信息
	if !quiet {
		fmt.Printf("⚠️  警告: 即将永久删除回收站中的 %d 个项目 (%s)\n",
			len(trashFiles), filesystem.FormatFileSize(totalSize))
		fmt.Println("⚠️  此操作不可逆，删除后无法恢复！")
	}

	// 确认操作
	if !force {
		fmt.Print("确认要清空回收站吗? 请输入 'yes' 确认: ")
		var response string
		fmt.Scanln(&response)
		if response != "yes" && response != "YES" {
			fmt.Println("❌ 操作已取消")
			return nil
		}
	}

	// 执行清空操作
	if !quiet {
		fmt.Println("🗑️  正在清空回收站...")
	}

	err = manager.EmptyTrash()
	if err != nil {
		return fmt.Errorf("清空回收站失败: %v", err)
	}

	// 显示成功信息
	if !quiet {
		fmt.Printf("✅ 成功清空回收站，删除了 %d 个项目 (%s)\n",
			len(trashFiles), filesystem.FormatFileSize(totalSize))
	}

	return nil
}
