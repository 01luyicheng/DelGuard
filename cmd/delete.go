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
	"delguard/internal/security"
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
		// 清理路径，防止路径遍历攻击
		cleanArg := filepath.Clean(arg)
		
		// 验证路径长度
		if len(cleanArg) > 4096 {
			if !quiet {
				fmt.Fprintf(os.Stderr, "⚠️  警告: 路径过长 '%s'\n", cleanArg)
			}
			continue
		}
		
		// 检查路径是否包含空字符
		if strings.ContainsRune(cleanArg, 0) {
			if !quiet {
				fmt.Fprintf(os.Stderr, "⚠️  警告: 路径包含非法字符 '%s'\n", cleanArg)
			}
			continue
		}
		
		matches, err := filepath.Glob(cleanArg)
		if err != nil {
			if !quiet {
				fmt.Fprintf(os.Stderr, "⚠️  警告: 无法处理路径 '%s': %v\n", cleanArg, err)
			}
			continue
		}

		if len(matches) == 0 {
			// 没有匹配的文件，检查是否是直接路径
			if _, err := os.Stat(cleanArg); err == nil {
				filesToDelete = append(filesToDelete, cleanArg)
			} else {
				if !quiet {
					fmt.Fprintf(os.Stderr, "⚠️  警告: 文件不存在 '%s'\n", cleanArg)
				}
			}
		} else {
			filesToDelete = append(filesToDelete, matches...)
		}
	}

	if len(filesToDelete) == 0 {
		return fmt.Errorf("没有找到要删除的文件")
	}

	// 创建路径验证器
	validator := security.NewPathValidator()
	
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

		// 验证路径安全性
		if err := validator.ValidateDeletePath(absPath); err != nil {
			if !quiet {
				fmt.Fprintf(os.Stderr, "⚠️  安全警告: %v\n", err)
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

		// 检查是否为系统文件
		if isSystemFile(absPath) {
			if !quiet {
				fmt.Fprintf(os.Stderr, "⚠️  警告: '%s' 可能是系统文件，删除可能导致系统问题\n", file)
			}
			if !force {
				continue
			}
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

	// 批量处理优化
	batchSize := 10
	if len(validFiles) > batchSize {
		fmt.Printf("🔄 正在批量处理 %d 个文件...\n", len(validFiles))
	}

	for i, file := range validFiles {
		// 显示进度
		if len(validFiles) > batchSize && !quiet {
			fmt.Printf("进度: %d/%d\r", i+1, len(validFiles))
		}

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
		if len(validFiles) > 10 {
			fmt.Println() // 添加换行
		}
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

// isSystemFile 检查是否为系统文件
func isSystemFile(path string) bool {
	// 检查文件属性
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	
	// 检查文件名
	name := strings.ToLower(filepath.Base(path))
	systemFiles := []string{
		"ntldr", "boot.ini", "bootmgr", "pagefile.sys", "hiberfil.sys",
		"swapfile.sys", "desktop.ini", "thumbs.db", ".DS_Store",
	}
	
	for _, sysFile := range systemFiles {
		if name == sysFile {
			return true
		}
	}
	
	// 检查目录名 - 使用动态系统盘符检测
	dir := strings.ToLower(filepath.Dir(path))
	
	// 动态获取系统盘符
	systemDrive := os.Getenv("SystemDrive")
	if systemDrive == "" {
		systemDrive = "C:"
	}
	
	systemDirs := []string{
		"windows", "system32", "syswow64", "program files", "program files (x86)",
		"programdata", "recovery", "boot", "msocache", "perflogs",
	}

	// 检查相对路径和绝对路径
	for _, sysDir := range systemDirs {
		// 检查相对路径
		if strings.Contains(dir, sysDir) {
			return true
		}
		
		// 检查绝对路径
		absSysDir := filepath.Join(strings.ToLower(systemDrive), sysDir)
		if strings.Contains(dir, absSysDir) {
			return true
		}
	}
	
	return false
}
