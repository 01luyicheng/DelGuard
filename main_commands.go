package main

import (
	"context"
	"fmt"
	"os"
)

// handleCommand 处理命令行参数
func handleCommand(ctx context.Context, config *Config) error {
	// 简化命令处理
	if len(os.Args) < 2 {
		fmt.Println("DelGuard - 安全删除工具")
		fmt.Println("使用方法:")
		fmt.Println("  delguard <文件或目录> [选项]")
		fmt.Println("  delguard restore [模式] - 恢复文件")
		fmt.Println("  delguard cp <源> <目标> - 安全复制")
		fmt.Println("  delguard config init - 初始化配置")
		fmt.Println("  delguard config show - 显示当前配置")
		fmt.Println("  delguard version - 显示版本")
		return nil
	}

	cmd := os.Args[1]
	switch cmd {
	case "restore":
		return handleRestoreCommand(ctx, config, os.Args[2:])
	case "cp":
		if len(os.Args) < 4 {
			return fmt.Errorf("安全复制需要源和目标参数")
		}
		return handleCopyCommand(ctx, config, os.Args[2], os.Args[3])
	case "config":
		if len(os.Args) < 3 {
			return fmt.Errorf("请指定配置命令: init 或 show")
		}
		subCmd := os.Args[2]
		switch subCmd {
		case "init":
			return handleConfigInitCommand(config)
		case "show":
			return handleConfigShowCommand(config)
		default:
			return fmt.Errorf("未知配置命令: %s", subCmd)
		}
	case "version":
		fmt.Printf("DelGuard v%s\n", Version)
		return nil
	default:
		// 处理删除命令
		return handleDeleteCommand(ctx, config, os.Args[1:])
	}
}

// handleConfigShowCommand 处理配置显示命令
func handleConfigShowCommand(config *Config) error {
	fmt.Println("当前配置:")
	fmt.Printf("  版本: %s\n", config.Version)
	fmt.Printf("  架构版本: %s\n", config.SchemaVersion)
	fmt.Printf("  语言: %s\n", config.Language)
	fmt.Printf("  日志级别: %s\n", config.LogLevel)
	fmt.Printf("  交互模式: %s\n", config.InteractiveMode)
	fmt.Printf("  安全模式: %s\n", config.SafeMode)
	fmt.Printf("  使用回收站: %t\n", config.UseRecycleBin)
	fmt.Printf("  最大备份文件数: %d\n", config.MaxBackupFiles)
	fmt.Printf("  回收站最大容量: %d MB\n", config.TrashMaxSize)
	fmt.Printf("  最大文件大小: %d bytes\n", config.MaxFileSize)
	fmt.Printf("  路径长度限制: %d\n", config.MaxPathLength)
	fmt.Printf("  并发操作限制: %d\n", config.MaxConcurrentOps)
	fmt.Printf("  安全检查: %t\n", config.EnableSecurityChecks)
	fmt.Printf("  恶意软件扫描: %t\n", config.EnableMalwareScan)
	fmt.Printf("  路径验证: %t\n", config.EnablePathValidation)
	fmt.Printf("  隐藏文件检查: %t\n", config.EnableHiddenCheck)
	fmt.Printf("  覆盖保护: %t\n", config.EnableOverwriteProtection)
	fmt.Printf("  备份保留天数: %d\n", config.BackupRetentionDays)
	fmt.Printf("  日志保留天数: %d\n", config.LogRetentionDays)
	fmt.Printf("  遥测: %t\n", config.EnableTelemetry)

	if config.ConfigPath != "" {
		fmt.Printf("  配置文件路径: %s\n", config.ConfigPath)
	}

	return nil
}

// handleDeleteCommand 处理删除命令
func handleDeleteCommand(ctx context.Context, config *Config, args []string) error {
	// 创建删除器
	deleter := NewCoreDeleter(config)

	// 设置删除选项（默认使用回收站，交互式确认）
	deleter.SetOptions(false, true, false, true, true)

	// 执行删除
	results := deleter.Delete(args)

	// 显示结果
	return displayDeleteResults(results, true)
}

// handleRestoreCommand 处理恢复命令
func handleRestoreCommand(ctx context.Context, config *Config, args []string) error {
	pattern := ""
	if len(args) > 0 {
		pattern = args[0]
	}

	restorer := NewRestorer(config)
	return restorer.Restore(ctx, pattern, false, false, 0)
}

// handleCopyCommand 处理安全复制命令
func handleCopyCommand(ctx context.Context, config *Config, source, dest string) error {
	copier := NewSafeCopier(config)
	return copier.Copy(ctx, source, dest, false, false, false, true)
}

// handleConfigInitCommand 处理配置初始化命令
func handleConfigInitCommand(config *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("获取配置文件路径失败: %v", err)
	}
	if err := SaveDefaultConfig(configPath); err != nil {
		return fmt.Errorf("创建配置文件失败: %v", err)
	}
	fmt.Printf("配置文件已创建: %s\n", configPath)
	return nil
}

// displayDeleteResults 显示删除结果
func displayDeleteResults(results []DeleteResult, verbose bool) error {
	var success, failed, skipped int

	for _, result := range results {
		if result.Skipped {
			skipped++
			continue
		}
		if result.Error != nil {
			failed++
			fmt.Printf("删除失败: %s - %v\n", result.Path, result.Error)
		} else {
			success++
			fmt.Printf("已删除: %s\n", result.Path)
		}
	}

	fmt.Printf("\n删除完成: 成功%d个, 失败%d个, 跳过%d个\n", success, failed, skipped)
	return nil
}
