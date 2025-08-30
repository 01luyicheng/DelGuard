package main

import (
	"fmt"
	"log"
	"os"

	"delguard/cmd"
	"delguard/internal/config"
	"delguard/internal/logger"
)

func main() {
	// 初始化配置
	if err := config.Init(); err != nil {
		log.Printf("初始化配置失败: %v", err)
	}

	// 初始化日志
	defer func() {
		if err := logger.Close(); err != nil {
			log.Printf("关闭日志失败: %v", err)
		}
	}()

	if config.GlobalConfig != nil && config.GlobalConfig.Logging.File != "" {
		cfg := config.GlobalConfig.Logging
		if err := logger.Init(cfg.File, cfg.Level, cfg.MaxSize, cfg.MaxAge, cfg.Compress); err != nil {
			log.Printf("初始化日志失败: %v", err)
			// 回退到默认日志配置
			if err := logger.Init(config.GetDefaultLogPath(), "info", 10, 7, true); err != nil {
				log.Printf("使用默认配置初始化日志失败: %v", err)
			}
		}
	} else {
		// 使用默认配置初始化日志
		if err := logger.Init(config.GetDefaultLogPath(), "info", 10, 7, true); err != nil {
			log.Printf("使用默认配置初始化日志失败: %v", err)
		}
	}

	// 设置优雅退出处理
	defer func() {
		// 确保日志文件被正确关闭
		if err := logger.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "关闭日志文件失败: %v\n", err)
		}
		
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "程序发生严重错误: %v\n", r)
			os.Exit(1)
		}
	}()

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}