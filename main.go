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
	if config.GlobalConfig != nil {
		cfg := config.GlobalConfig.Logging
		if err := logger.Init(cfg.File, cfg.Level, cfg.MaxSize, cfg.MaxAge, cfg.Compress); err != nil {
			log.Printf("初始化日志失败: %v", err)
		}
	}

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}