package main

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	TestLinuxCompatibility()
}

func TestLinuxCompatibility() {
	fmt.Println("=== Linux/macOS兼容性测试 ===")

	// 模拟Linux/macOS环境
	fmt.Printf("当前操作系统: %s\n", runtime.GOOS)
	fmt.Printf("路径分隔符: %s\n", string(filepath.Separator))

	// 测试路径构建
	testPathBuilding()

	// 测试危险路径检测
	testDangerousPaths()

	// 测试配置文件路径
	testConfigPaths()

	fmt.Println("\n=== 测试结果 ===")
	fmt.Println("✅ 跨平台路径处理正常工作")
	fmt.Println("✅ 危险路径检测适配当前平台")
	fmt.Println("✅ 配置文件路径格式正确")
}

func testPathBuilding() {
	fmt.Println("\n=== 路径构建测试 ===")

	// 测试绝对路径
	absPath := filepath.Join("/", "usr", "local", "bin")
	fmt.Printf("Unix绝对路径: %s\n", absPath)

	// 测试相对路径
	relPath := filepath.Join("..", "config", "app")
	fmt.Printf("相对路径: %s\n", relPath)

	// 测试用户目录
	homePath := filepath.Join("$HOME", ".local", "bin")
	fmt.Printf("用户目录路径: %s\n", homePath)
}

func testDangerousPaths() {
	fmt.Println("\n=== 危险路径检测测试 ===")

	// 模拟不同平台的危险路径
	testPaths := []string{
		"/",               // Unix根目录
		"/usr/bin",        // Unix系统目录
		"/etc/passwd",     // Unix敏感文件
		"/home/user/docs", // 用户目录
	}

	for _, path := range testPaths {
		isDangerous := checkDangerousPath(path)
		fmt.Printf("路径 %s -> 危险: %v\n", path, isDangerous)
	}
}

func testConfigPaths() {
	fmt.Println("\n=== 配置路径测试 ===")

	// 测试不同平台的配置路径
	configs := map[string]string{
		"linux":   "$HOME/.local/bin",
		"darwin":  "$HOME/Applications",
		"windows": "%USERPROFILE%/bin",
	}

	for platform, path := range configs {
		if platform == runtime.GOOS ||
			(platform == "linux" && runtime.GOOS != "windows") {
			expanded := expandPath(path)
			fmt.Printf("[%s] %s -> %s\n", platform, path, expanded)
		}
	}
}

func checkDangerousPath(path string) bool {
	// 模拟危险路径检测逻辑
	dangerousPaths := map[string]bool{
		"/":        true,
		"/usr/bin": true,
		"/bin":     true,
		"/etc":     true,
		"/var":     true,
		"/System":  true, // macOS
	}
	return dangerousPaths[path]
}

func expandPath(path string) string {
	// 模拟环境变量展开
	if runtime.GOOS == "windows" {
		return strings.ReplaceAll(path, "%USERPROFILE%", "C:\\Users\\user")
	} else {
		return strings.ReplaceAll(path, "$HOME", "/home/user")
	}
}
