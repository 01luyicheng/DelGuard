package main

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

func VerifyCrossplatform() {
	fmt.Println("=== 跨平台路径修复验证 ===")
	fmt.Printf("操作系统: %s\n", runtime.GOOS)
	fmt.Printf("路径分隔符: %s\n", string(filepath.Separator))

	// 验证路径构建
	fmt.Println("\n=== 路径构建测试 ===")

	// 测试filepath.Join
	windowsPath := filepath.Join("C:", "Windows", "System32")
	fmt.Printf("Windows路径: %s\n", windowsPath)

	unixPath := filepath.Join("/", "usr", "local", "bin")
	fmt.Printf("Unix路径: %s\n", unixPath)

	// 测试危险路径检测
	fmt.Println("\n=== 危险路径检测测试 ===")
	dangerousPaths := []string{
		"C:\\Windows\\System32",
		"/usr/bin",
		"C:\\Users\\test\\Documents",
		"/home/user",
	}

	for _, path := range dangerousPaths {
		isDangerous := isDangerousPath(path)
		fmt.Printf("路径 %s 危险状态: %v\n", path, isDangerous)
	}

	fmt.Println("\n=== 配置路径测试 ===")
	// 模拟配置路径
	configPaths := []string{
		"%USERPROFILE%/bin",
		"%ProgramFiles%/DelGuard",
		"%APPDATA%/DelGuard",
	}

	for _, path := range configPaths {
		expanded := expandEnvVars(path)
		fmt.Printf("原始: %s -> 展开: %s\n", path, expanded)
	}
}

// 模拟危险路径检测
func isDangerousPath(path string) bool {
	systemPaths := map[string]bool{
		"C:\\Windows\\System32": true,
		"/usr/bin":              true,
		"/bin":                  true,
		"C:\\":                  true,
		"/":                     true,
	}

	return systemPaths[path]
}

// 模拟环境变量展开
func expandEnvVars(path string) string {
	if runtime.GOOS == "windows" {
		path = strings.ReplaceAll(path, "%USERPROFILE%", "C:\\Users\\test")
		path = strings.ReplaceAll(path, "%ProgramFiles%", "C:\\Program Files")
		path = strings.ReplaceAll(path, "%APPDATA%", "C:\\Users\\test\\AppData\\Roaming")
	} else {
		path = strings.ReplaceAll(path, "$HOME", "/home/test")
		path = strings.ReplaceAll(path, "$USER", "test")
	}
	return path
}

// 添加缺失的strings import
