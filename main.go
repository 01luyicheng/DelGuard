package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	Run(os.Args)
}

func Run(args []string) {
	ctx := context.Background()
	
	// 加载配置
	config, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}
	
	// 处理命令
	if err := handleCommand(ctx, config); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}

func runMainApp(args []string) error {
	// 这里可以添加主应用程序逻辑
	// 暂时直接返回nil
	return nil
}
