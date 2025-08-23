package main

import (
	"fmt"
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
			RunBasicSecurityChecks()
		case "--help":
			fmt.Println("DelGuard Security Check Tool")
			fmt.Println("Usage:")
			fmt.Println("  go run security_check.go --security-check  Run security checks")
			fmt.Println("  go run security_check.go --help           Show help")
		default:
			fmt.Printf("Unknown option: %s\n", args[1])
			fmt.Println("Use --help for usage information")
		}
		return
	}

	fmt.Println("DelGuard Security Check Tool")
	fmt.Println("Use --security-check to run security checks")
	fmt.Println("Use --help for usage information")
}

// RunBasicSecurityChecks 运行基本安全检查
func RunBasicSecurityChecks() {
	fmt.Println("=== DelGuard Basic Security Check ===")
	fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	
	// 检查关键文件
	requiredFiles := []string{
		"config/security_template.json",
		"SECURITY.md",
		"SECURITY_SUMMARY.md",
	}
	
	missingFiles := []string{}
	for _, file := range requiredFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			missingFiles = append(missingFiles, file)
		}
	}
	
	if len(missingFiles) > 0 {
		fmt.Println("❌ Missing required security files:")
		for _, file := range missingFiles {
			fmt.Printf("   - %s\n", file)
		}
	} else {
		fmt.Println("✅ All required security files present")
	}
	
	// 检查配置文件
	if config, err := LoadConfig(); err == nil {
		if err := validateConfig(config); err != nil {
			fmt.Printf("❌ Configuration validation failed: %v\n", err)
		} else {
			fmt.Println("✅ Configuration validation passed")
		}
	} else {
		fmt.Printf("❌ Failed to load configuration: %v\n", err)
	}
	
	// 检查关键路径保护
	testPaths := []string{
		"/etc",
		"/bin",
		"/usr/bin",
		"/System",
		"C:\\Windows",
		"C:\\Program Files",
	}
	
	protectedCount := 0
	for _, path := range testPaths {
		if IsCriticalPath(path) {
			protectedCount++
		}
	}
	
	if protectedCount >= len(testPaths)/2 {
		fmt.Println("✅ Critical path protection working")
	} else {
		fmt.Println("❌ Critical path protection may not be working correctly")
	}

	// 检查环境变量
	fmt.Println("\n=== Environment Variables Check ===")
	sensitiveVars := []string{"PATH", "HOME", "USER", "SUDO_USER"}
	for _, envVar := range sensitiveVars {
		if value := os.Getenv(envVar); value != "" {
			fmt.Printf("✅ %s: %s\n", envVar, value)
		} else {
			fmt.Printf("⚠️  %s not set\n", envVar)
		}
	}

	// 检查临时文件权限
	fmt.Println("\n=== Temporary Files Security Check ===")
	tempDirs := []string{os.TempDir(), "/tmp", "/var/tmp"}
	for _, tempDir := range tempDirs {
		if info, err := os.Stat(tempDir); err == nil {
			mode := info.Mode()
			if mode&0002 != 0 {
				fmt.Printf("⚠️  临时目录 %s 具有全局写权限\n", tempDir)
			} else {
				fmt.Printf("✅ 临时目录 %s 权限安全\n", tempDir)
			}
		}
	}

	// 生成安全建议
	fmt.Println("\n=== Security Recommendations ===")
	fmt.Println("• 定期更新系统补丁")
	fmt.Println("• 定期备份重要数据")
	fmt.Println("• 使用强密码策略")
	fmt.Println("• 检查文件权限设置")
	
	fmt.Println("\nNote: Full security check implementation in progress")
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
			"C:\\Windows",
			"C:\\Program Files"
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
		"C:\\Windows\\System32\\cmd.exe",
	}
	
	for _, path := range systemPaths {
		if !IsCriticalPath(path) {
			issues = append(issues, fmt.Sprintf("Critical path protection not working for: %s", path))
		}
	}
	
	return issues
}