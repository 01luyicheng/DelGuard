package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func main() {
	// 设置运行时参数
	runtime.GOMAXPROCS(runtime.NumCPU())

	// 创建上下文用于优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 初始化全局日志记录器
	InitGlobalLogger("INFO")
	defer func() {
		if globalLogger != nil {
			globalLogger.Close()
		}
	}()

	// 启动信号监听协程
	go func() {
		sig := <-sigChan
		LogInfo("main", "", fmt.Sprintf("收到信号: %v，开始优雅关闭", sig))
		cancel()
	}()

	// 运行主应用程序
	exitCode := runMainApp(ctx)

	// 退出程序
	os.Exit(exitCode)
}

// runMainApp 运行主应用程序逻辑
func runMainApp(ctx context.Context) int {
	defer func() {
		if r := recover(); r != nil {
			LogError("main", "", fmt.Errorf("应用程序崩溃: %v", r), "程序发生未处理的异常")
			fmt.Fprintf(os.Stderr, "❌ 程序发生严重错误: %v\n", r)
		}
	}()

	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		LogInfo("main", "", "应用程序启动前收到取消信号")
		return ExitUserCancelled
	default:
	}

	// 运行DelGuard应用程序
	if err := RunDelGuardApp(ctx); err != nil {
		LogError("main", "", err, "DelGuard应用程序执行失败")

		// 根据错误类型返回适当的退出码
		if dgErr, ok := err.(*DGError); ok {
			return dgErr.Kind.ExitCode()
		}

		return ExitGeneralError
	}

	return ExitSuccess
}

// RunDelGuardApp 运行DelGuard应用程序的主要逻辑
func RunDelGuardApp(ctx context.Context) error {
	// 这里需要实现实际的应用程序逻辑
	// 目前返回一个占位符错误，表示需要实现
	return fmt.Errorf("DelGuard应用程序逻辑尚未实现")
}
