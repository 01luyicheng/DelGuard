package main

import (
	"flag"
	"fmt"
	"os"
)

// SecurityTool 安全工具入口函数，由主程序调用
func SecurityTool(args []string) {
	// 定义命令行参数
	help := flag.Bool("help", false, "显示帮助信息")
	version := flag.Bool("version", false, "显示版本信息")
	check := flag.Bool("check", false, "运行安全检查")
	verify := flag.Bool("verify", false, "验证安全配置")
	report := flag.Bool("report", false, "生成安全报告")

	flag.Parse()

	// 显示帮助
	if *help {
		printHelp()
		return
	}

	// 显示版本
	if *version {
		fmt.Printf("DelGuard Security Tool v%s\n", "1.0.0")
		return
	}

	// 运行安全检查
	if *check {
		fmt.Println("运行基本安全检查...")
		RunSecurityVerification() // 使用现有的验证函数
		return
	}

	// 验证安全配置
	if *verify {
		fmt.Println("验证安全配置...")
		RunSecurityVerification()
		return
	}

	// 生成安全报告
	if *report {
		fmt.Println("生成安全报告...")
		RunSecurityVerification()
		return
	}

	// 如果没有指定参数，运行完整检查
	if flag.NArg() == 0 {
		fmt.Println("运行完整安全检查...")
		RunSecurityVerification()
		return
	}

	// 处理子命令
	if flag.NArg() > 0 {
		switch flag.Arg(0) {
		case "check":
			RunSecurityVerification()
		case "verify":
			RunSecurityVerification()
		case "report":
			RunSecurityVerification()
		default:
			fmt.Printf("未知命令: %s\n", flag.Arg(0))
			printHelp()
			os.Exit(1)
		}
	}
}

func printHelp() {
	fmt.Println("DelGuard Security Tool")
	fmt.Println("用法:")
	fmt.Println("  security_tool [选项]")
	fmt.Println("  security_tool [命令]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -help     显示帮助信息")
	fmt.Println("  -version  显示版本信息")
	fmt.Println("  -check    运行基本安全检查")
	fmt.Println("  -verify   验证安全配置")
	fmt.Println("  -report   生成安全报告")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  check     运行基本安全检查")
	fmt.Println("  verify    验证安全配置")
	fmt.Println("  report    生成安全报告")
	fmt.Println()
	fmt.Println("如果没有指定命令或选项，则运行所有检查。")
}
