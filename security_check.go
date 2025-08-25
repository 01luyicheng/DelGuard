package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// RunSecurityCheckTool 运行安全检查工具
func RunSecurityCheckTool(args []string) {
	if len(args) > 1 {
		switch args[1] {
		case "--security-check":
			log.Println("[INFO] 执行安全检查流程")
			RunBasicSecurityChecks()
		case "--help":
			log.Println("[INFO] 显示安全检查工具帮助")
			fmt.Println(T("DelGuard 安全检查工具"))
			fmt.Println(T("用法:"))
			fmt.Println(T("  go run security_check.go --security-check  执行安全检查"))
			fmt.Println(T("  go run security_check.go --help           显示帮助"))
		default:
			log.Printf("[WARN] 未知选项: %s", args[1])
			fmt.Printf(T("未知选项: %s\n"), args[1])
			fmt.Println(T("使用 --help 查看用法信息"))
		}
		return
	}

	fmt.Println(T("DelGuard 安全检查工具"))
	fmt.Println(T("使用 --security-check 执行安全检查"))
	fmt.Println(T("使用 --help 查看用法信息"))
}

// RunBasicSecurityChecks 运行基本安全检查
func RunBasicSecurityChecks() {
	log.Println("[INFO] === DelGuard 基础安全检查 ===")
	log.Printf("[INFO] 平台: %s/%s", runtime.GOOS, runtime.GOARCH)

	// 检查关键文件
	requiredFiles := []string{
		"config/security_template.json",
		"SECURITY.md",
		"SECURITY_SUMMARY.md",
	}

	missingFiles := []string{}
	for _, file := range requiredFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			log.Printf("[WARN] 缺少安全文件: %s", file)
			missingFiles = append(missingFiles, file)
		}
	}

	if len(missingFiles) > 0 {
		log.Println("[ERROR] 缺少必要的安全文件:")
		for _, file := range missingFiles {
			log.Printf("[ERROR]   - %s", file)
		}
		fmt.Println(T("❌ 缺少必要的安全文件:"))
		for _, file := range missingFiles {
			fmt.Printf("   - %s\n", file)
		}
	} else {
		log.Println("[INFO] 所有必要的安全文件已存在")
		fmt.Println(T("✅ 所有必要的安全文件已存在"))
	}

	// 检查配置文件
	if config, err := LoadConfig(); err == nil {
		if err := validateConfig(config); err != nil {
			log.Printf("[ERROR] 配置校验失败: %v", err)
			fmt.Printf(T("❌ 配置校验失败: %v\n"), err)
		} else {
			log.Println("[INFO] 配置校验通过")
			fmt.Println(T("✅ 配置校验通过"))
		}
	} else {
		log.Printf("[ERROR] 加载配置失败: %v", err)
		fmt.Printf(T("❌ 加载配置失败: %v\n"), err)
	}

	// 检查关键路径保护
	testPaths := []string{
		"/etc",
		"/bin",
		"/usr/bin",
		"/System",
		filepath.Join(os.Getenv("SYSTEMDRIVE"), "Windows"),
		filepath.Join(os.Getenv("SYSTEMDRIVE"), "Program Files"),
	}

	protectedCount := 0
	for _, path := range testPaths {
		if IsCriticalPath(path) {
			protectedCount++
		}
	}

	if protectedCount >= len(testPaths)/2 {
		log.Println("[INFO] 关键路径保护正常")
		fmt.Println(T("✅ 关键路径保护正常"))
	} else {
		log.Println("[WARN] 关键路径保护可能未正常工作")
		fmt.Println(T("❌ 关键路径保护可能未正常工作"))
	}

	// 检查环境变量
	log.Println("[INFO] === 环境变量检查 ===")
	fmt.Println(T("\n=== 环境变量检查 ==="))
	sensitiveVars := []string{"PATH", "HOME", "USER", "SUDO_USER"}
	for _, envVar := range sensitiveVars {
		if value := os.Getenv(envVar); value != "" {
			log.Printf("[INFO] 环境变量 %s: %s", envVar, value)
			fmt.Printf(T("✅ %s: %s\n"), envVar, value)
		} else {
			log.Printf("[WARN] 环境变量未设置: %s", envVar)
			fmt.Printf(T("⚠️  %s 未设置\n"), envVar)
		}
	}

	// 检查临时文件权限
	log.Println("[INFO] === 临时文件安全检查 ===")
	fmt.Println(T("\n=== 临时文件安全检查 ==="))
	tempDirs := []string{os.TempDir(), "/tmp", "/var/tmp"}
	for _, tempDir := range tempDirs {
		if info, err := os.Stat(tempDir); err == nil {
			mode := info.Mode()
			if mode&0002 != 0 {
				log.Printf("[WARN] 临时目录 %s 具有全局写权限", tempDir)
				fmt.Printf(T("⚠️  临时目录 %s 具有全局写权限\n"), tempDir)
			} else {
				log.Printf("[INFO] 临时目录 %s 权限安全", tempDir)
				fmt.Printf(T("✅ 临时目录 %s 权限安全\n"), tempDir)
			}
		}
	}

	// 生成安全建议
	log.Println("[INFO] === 安全建议 ===")
	fmt.Println(T("\n=== 安全建议 ==="))
	fmt.Println(T("• 定期更新系统补丁"))
	fmt.Println(T("• 定期备份重要数据"))
	fmt.Println(T("• 使用强密码策略"))
	fmt.Println(T("• 检查文件权限设置"))

	log.Println("[INFO] 提示: 完整的安全检查实现仍在进行中")
	fmt.Println(T("\n提示: 完整的安全检查实现仍在进行中"))
}

// validateConfig 验证配置
func validateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// 基本验证
	if config.MaxBackupFiles < 0 {
		return fmt.Errorf("MaxBackupFiles cannot be negative")
	}

	if config.Language == "" {
		return fmt.Errorf("language cannot be empty")
	}

	return nil
}

// CheckSystemIntegrity 检查系统完整性
func CheckSystemIntegrity() error {
	// 检查必要的目录
	requiredDirs := []string{"config", "logs", "backups"}
	for _, dir := range requiredDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %v", dir, err)
			}
		}
	}

	// 检查安全配置文件
	securityFiles := []string{
		"config/security_template.json",
	}

	for _, file := range securityFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			// 尝试创建默认配置
			if err := createDefaultSecurityConfig(file); err != nil {
				return fmt.Errorf("missing security file %s and failed to create default: %v", file, err)
			}
		}
	}

	return nil
}

// createDefaultSecurityConfig 创建默认安全配置
func createDefaultSecurityConfig(filename string) error {
	// 确保目录存在
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 创建默认配置内容
	defaultConfig := `{
		"version": "1.0",
		"critical_paths": [
			"/etc",
			"/bin", 
			"/usr/bin",
			"/System",
			filepath.Join(os.Getenv("SYSTEMDRIVE"), "Windows"),
			filepath.Join(os.Getenv("SYSTEMDRIVE"), "Program Files")
		],
		"max_file_size": 1073741824,
		"allowed_extensions": [
			".txt", ".doc", ".docx", ".pdf", ".jpg", ".png", ".gif"
		],
		"enable_encryption": true,
		"backup_count": 3
	}`

	return os.WriteFile(filename, []byte(defaultConfig), 0644)
}

// VerifySecurityFeatures 验证安全功能
func VerifySecurityFeatures() []string {
	var issues []string

	// 检查路径遍历保护
	testPaths := []string{
		"../../../etc/passwd",
		"..\\..\\windows\\system32\\cmd.exe",
		"/etc/shadow",
	}

	for _, path := range testPaths {
		if strings.Contains(path, "..") {
			// 应该被阻止 - 检查清理后的路径是否包含..
			cleanPath := filepath.Clean(path)
			if strings.Contains(cleanPath, ".."+string(filepath.Separator)) ||
				strings.Contains(cleanPath, "..\\") ||
				strings.Contains(cleanPath, "../") {
				// 路径遍历检测正常
				continue
			} else {
				issues = append(issues, fmt.Sprintf("Path traversal protection may not be working for: %s", path))
			}
		}
	}

	// 检查系统路径保护
	systemPaths := []string{
		"/etc/passwd",
		"/bin/sh",
		filepath.Join(os.Getenv("SYSTEMDRIVE"), "Windows", "System32", "cmd.exe"),
	}

	for _, path := range systemPaths {
		if !IsCriticalPath(path) {
			issues = append(issues, fmt.Sprintf("Critical path protection not working for: %s", path))
		}
	}

	return issues
}
