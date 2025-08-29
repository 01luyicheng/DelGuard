package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"delguard/internal/filesystem"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [files...]",
	Short: "安全删除文件到回收站",
	Long: `将指定的文件或目录安全地移动到系统回收站。
支持多个文件同时删除，支持通配符模式。`,
	Aliases: []string{"del", "rm"},
	Args:    cobra.MinimumNArgs(1),
	RunE:    runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().BoolP("force", "f", false, "强制删除，不显示确认提示")
	deleteCmd.Flags().BoolP("recursive", "r", false, "递归删除目录")
	deleteCmd.Flags().BoolP("verbose", "v", false, "显示详细信息")
	deleteCmd.Flags().BoolP("interactive", "i", false, "交互式删除，每个文件都询问")
	deleteCmd.Flags().BoolP("dry-run", "n", false, "预览模式，显示将要删除的文件但不实际删除")
}

func runDelete(cmd *cobra.Command, args []string) error {
	// 获取标志值
	force, _ := cmd.Flags().GetBool("force")
	recursive, _ := cmd.Flags().GetBool("recursive")
	interactive, _ := cmd.Flags().GetBool("interactive")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	verbose := viper.GetBool("verbose")
	quiet := viper.GetBool("quiet")

	// 获取回收站管理器
	manager, err := filesystem.GetTrashManager()
	if err != nil {
		return fmt.Errorf("初始化回收站管理器失败: %v", err)
	}

	// 展开所有文件路径（处理通配符）
	var filesToDelete []string
	for _, arg := range args {
		matches, err := filepath.Glob(arg)
		if err != nil {
			if !quiet {
				fmt.Fprintf(os.Stderr, "⚠️  警告: 无法处理路径 '%s': %v\n", arg, err)
			}
			continue
		}

		if len(matches) == 0 {
			// 没有匹配的文件，检查是否是直接路径
			if _, err := os.Stat(arg); err == nil {
				filesToDelete = append(filesToDelete, arg)
			} else {
				if !quiet {
					fmt.Fprintf(os.Stderr, "⚠️  警告: 文件不存在 '%s'\n", arg)
				}
			}
		} else {
			filesToDelete = append(filesToDelete, matches...)
		}
	}

	if len(filesToDelete) == 0 {
		return fmt.Errorf("没有找到要删除的文件")
	}

	// 验证文件并过滤
	var validFiles []string
	for _, file := range filesToDelete {
		absPath, err := filepath.Abs(file)
		if err != nil {
			if !quiet {
				fmt.Fprintf(os.Stderr, "⚠️  警告: 无法获取绝对路径 '%s': %v\n", file, err)
			}
			continue
		}

		info, err := os.Stat(absPath)
		if err != nil {
			if !quiet {
				fmt.Fprintf(os.Stderr, "⚠️  警告: 无法访问文件 '%s': %v\n", file, err)
			}
			continue
		}

		// 检查是否为目录且未指定递归删除
		if info.IsDir() && !recursive {
			if !quiet {
				fmt.Fprintf(os.Stderr, "⚠️  警告: '%s' 是目录，使用 -r 选项递归删除\n", file)
			}
			continue
		}

		validFiles = append(validFiles, absPath)
	}

	if len(validFiles) == 0 {
		return fmt.Errorf("没有有效的文件可以删除")
	}

	// 预览模式
	if dryRun {
		fmt.Println("🔍 预览模式 - 以下文件将被移动到回收站:")
		for _, file := range validFiles {
			info, _ := os.Stat(file)
			fileType := "文件"
			if info.IsDir() {
				fileType = "目录"
			}
			fmt.Printf("  📄 %s (%s)\n", file, fileType)
		}
		return nil
	}

	// 确认删除
	if !force && !interactive {
		fmt.Printf("🗑️  将要删除 %d 个项目到回收站，确认吗? [y/N]: ", len(validFiles))
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			log.Printf("读取输入时出错: %v", err)
			fmt.Println("❌ 读取输入失败，操作已取消")
			return nil
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("❌ 操作已取消")
			return nil
		}
	}

	// 执行删除
	successCount := 0
	errorCount := 0

	for _, file := range validFiles {
		// 交互式确认
		if interactive {
			fmt.Printf("删除 '%s'? [y/N]: ", file)
			var response string
			if _, err := fmt.Scanln(&response); err != nil {
				log.Printf("读取输入时出错: %v", err)
				fmt.Println("❌ 读取输入失败，跳过此文件")
				continue
			}

			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				if verbose {
					fmt.Printf("⏭️  跳过: %s\n", file)
				}
				continue
			}
		}

		// 执行删除
		err := manager.MoveToTrash(file)
		if err != nil {
			errorCount++
			if !quiet {
				fmt.Fprintf(os.Stderr, "❌ 删除失败 '%s': %v\n", file, err)
			}
		} else {
			successCount++
			if verbose {
				fmt.Printf("✅ 已移动到回收站: %s\n", file)
			}
		}
	}

	// 显示结果摘要
	if !quiet {
		if successCount > 0 {
			fmt.Printf("✅ 成功删除 %d 个项目到回收站\n", successCount)
		}
		if errorCount > 0 {
			fmt.Printf("❌ %d 个项目删除失败\n", errorCount)
		}
	}

	if errorCount > 0 {
		return fmt.Errorf("部分文件删除失败")
	}

	return nil
}
