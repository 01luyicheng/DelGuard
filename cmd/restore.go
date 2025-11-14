package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"delguard/internal/filesystem"
	"delguard/internal/security"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// restoreCmd 恢复命令
var restoreCmd = &cobra.Command{
	Use:     "restore [文件名或索引...]",
	Aliases: []string{"recover", "undelete"},
	Short:   "从回收站恢复文件",
	Long: `从回收站恢复已删除的文件到原始位置或指定位置。

可以通过文件名或索引号来指定要恢复的文件。
使用 'delguard list' 查看回收站中的文件和对应的索引。

示例:
  delguard restore file.txt
  delguard restore 1 2 3          # 按索引恢复
  delguard restore --all           # 恢复所有文件
  delguard restore file.txt --to=/path/to/restore
  delguard recover file.txt        # 别名`,
	RunE: runRestore,
}

func init() {
	rootCmd.AddCommand(restoreCmd)

	// 添加标志
	restoreCmd.Flags().StringP("to", "t", "", "恢复到指定目录（默认恢复到原始位置）")
	restoreCmd.Flags().BoolP("all", "a", false, "恢复所有文件")
	restoreCmd.Flags().BoolP("force", "f", false, "强制恢复，覆盖已存在的文件")
	restoreCmd.Flags().BoolP("interactive", "i", false, "交互式恢复，每个文件都询问")
	restoreCmd.Flags().StringP("filter", "F", "", "按模式过滤要恢复的文件")
	restoreCmd.Flags().BoolP("dry-run", "n", false, "预览模式，显示将要恢复的文件但不实际恢复")
}

func runRestore(cmd *cobra.Command, args []string) error {
	// 获取标志值
	targetDir, _ := cmd.Flags().GetString("to")
	restoreAll, _ := cmd.Flags().GetBool("all")
	force, _ := cmd.Flags().GetBool("force")
	interactive, _ := cmd.Flags().GetBool("interactive")
	filter, _ := cmd.Flags().GetString("filter")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	verbose := viper.GetBool("verbose")
	quiet := viper.GetBool("quiet")

	// 获取回收站管理器
	manager, err := filesystem.GetTrashManager()
	if err != nil {
		return fmt.Errorf("初始化回收站管理器失败: %v", err)
	}

	// 获取回收站文件列表
	trashFiles, err := manager.ListTrashFiles()
	if err != nil {
		return fmt.Errorf("获取回收站文件列表失败: %v", err)
	}

	if len(trashFiles) == 0 {
		if !quiet {
			fmt.Println("🗑️  回收站是空的")
		}
		return nil
	}

	var filesToRestore []filesystem.TrashFile

	if restoreAll {
		// 恢复所有文件
		filesToRestore = trashFiles
	} else if len(args) == 0 && filter == "" {
		return fmt.Errorf("请指定要恢复的文件名、索引或使用 --all 恢复所有文件")
	} else {
		// 根据参数选择文件
		filesToRestore, err = selectFilesToRestore(trashFiles, args, filter)
		if err != nil {
			return err
		}
	}

	if len(filesToRestore) == 0 {
		return fmt.Errorf("没有找到匹配的文件")
	}

	// 预览模式
	if dryRun {
		fmt.Println("🔍 预览模式 - 以下文件将被恢复:")
		for i, file := range filesToRestore {
			restorePath := getRestorePath(file, targetDir)
			fmt.Printf("  %d. 📄 %s -> %s\n", i+1, file.Name, restorePath)
		}
		return nil
	}

	// 确认恢复
	if !force && !interactive && len(filesToRestore) > 1 {
		fmt.Printf("🔄 将要恢复 %d 个文件，确认吗? [y/N]: ", len(filesToRestore))
		var response string
		_, err := fmt.Scanln(&response)
		if err != nil {
			// 处理输入错误
			fmt.Println("❌ 读取输入失败，操作已取消")
			return nil
		}
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("❌ 操作已取消")
			return nil
		}
	}

	// 创建路径验证器
	validator := security.NewPathValidator()

	// 执行恢复
	successCount := 0
	errorCount := 0

	// 批量处理优化
	batchSize := 10
	if len(filesToRestore) > batchSize {
		fmt.Printf("🔄 正在批量恢复 %d 个文件...\n", len(filesToRestore))
	}

	for i, file := range filesToRestore {
		// 显示进度
		if len(filesToRestore) > batchSize && !quiet {
			fmt.Printf("进度: %d/%d\r", i+1, len(filesToRestore))
		}

		// 确定恢复路径
		restorePath := getRestorePath(file, targetDir)

		// 验证恢复路径安全性
		if err := validator.ValidateRestorePath(restorePath); err != nil {
			if !quiet {
				fmt.Fprintf(os.Stderr, "⚠️  安全警告: %s - %v\n", file.Name, err)
			}
			continue
		}

		// 交互式确认
		if interactive {
			fmt.Printf("恢复 '%s' 到 '%s'? [y/N]: ", file.Name, restorePath)
			var response string
			_, err := fmt.Scanln(&response)
			if err != nil {
				if verbose {
					fmt.Printf("⏭️  跳过: %s (输入错误)\n", file.Name)
				}
				continue
			}
			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				if verbose {
					fmt.Printf("⏭️  跳过: %s\n", file.Name)
				}
				continue
			}
		}

		// 检查目标文件是否已存在
		if !force {
			if _, err := os.Stat(restorePath); err == nil {
				// 如果文件已存在，添加后缀
				ext := filepath.Ext(restorePath)
				base := restorePath[:len(restorePath)-len(ext)]
				counter := 1
				for {
					newPath := fmt.Sprintf("%s_%d%s", base, counter, ext)
					if _, err := os.Stat(newPath); os.IsNotExist(err) {
						restorePath = newPath
						break
					}
					counter++
				}
				if !quiet {
					fmt.Fprintf(os.Stderr, "⚠️  文件已存在，重命名为: %s\n", filepath.Base(restorePath))
				}
			}
		}

		// 执行恢复
		err := manager.RestoreFile(file, restorePath)
		if err != nil {
			errorCount++
			if !quiet {
				fmt.Fprintf(os.Stderr, "❌ 恢复失败 '%s': %v\n", file.Name, err)
			}
		} else {
			successCount++
			if verbose {
				fmt.Printf("✅ 已恢复: %s -> %s\n", file.Name, restorePath)
			} else if !quiet {
				fmt.Printf("✅ 已恢复: %s\n", file.Name)
			}
		}
	}

	// 显示结果摘要
	if !quiet {
		if len(filesToRestore) > 10 {
			fmt.Println() // 换行
		}
		fmt.Println() // 换行
		if successCount > 0 {
			fmt.Printf("✅ 成功恢复 %d 个文件\n", successCount)
		}
		if errorCount > 0 {
			fmt.Printf("❌ %d 个文件恢复失败\n", errorCount)
		}
	}

	if errorCount > 0 {
		return fmt.Errorf("部分文件恢复失败")
	}

	return nil
}

// selectFilesToRestore 选择要恢复的文件
func selectFilesToRestore(trashFiles []filesystem.TrashFile, args []string, filter string) ([]filesystem.TrashFile, error) {
	var selected []filesystem.TrashFile

	// 应用过滤器
	if filter != "" {
		for _, file := range trashFiles {
			matched, err := matchPattern(file.Name, filter)
			if err != nil {
				return nil, fmt.Errorf("过滤器模式错误: %v", err)
			}
			if matched {
				selected = append(selected, file)
			}
		}
		return selected, nil
	}

	// 根据参数选择文件
	for _, arg := range args {
		// 尝试作为索引解析
		if idx := parseIndex(arg, len(trashFiles)); idx >= 0 {
			selected = append(selected, trashFiles[idx])
			continue
		}

		// 作为文件名匹配
		found := false
		for _, file := range trashFiles {
			if strings.EqualFold(file.Name, arg) {
				selected = append(selected, file)
				found = true
				break
			}
		}

		if !found {
			// 尝试部分匹配
			for _, file := range trashFiles {
				if strings.Contains(strings.ToLower(file.Name), strings.ToLower(arg)) {
					selected = append(selected, file)
					found = true
					break
				}
			}
		}

		if !found {
			return nil, fmt.Errorf("未找到文件: %s", arg)
		}
	}

	return selected, nil
}

// parseIndex 解析索引号
func parseIndex(s string, maxIndex int) int {
	var idx int
	if _, err := fmt.Sscanf(s, "%d", &idx); err != nil {
		return -1
	}

	// 转换为0基索引
	idx--
	if idx < 0 || idx >= maxIndex {
		return -1
	}

	return idx
}

// getRestorePath 获取恢复路径
func getRestorePath(file filesystem.TrashFile, targetDir string) string {
	if targetDir != "" {
		// 恢复到指定目录
		return filepath.Join(targetDir, file.Name)
	}

	// 恢复到原始位置
	if file.OriginalPath != "" {
		return file.OriginalPath
	}

	// 如果没有原始路径信息，恢复到当前目录
	currentDir, _ := os.Getwd()
	return filepath.Join(currentDir, file.Name)
}
