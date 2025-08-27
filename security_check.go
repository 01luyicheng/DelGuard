package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
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
			fmt.Println(T("  delguard --security-check  执行安全检查"))
			fmt.Println(T("  delguard --help           显示帮助"))
		default:
			log.Printf("[WARN] 未知选项: %s", args[1])
			fmt.Printf(T("未知选项: %s\n"), args[1])
			fmt.Println(T("使用 --help 查看用法信息"))
		}
		return
	}

	fmt.Println(T("DelGuard 安全检查工具"))
	fmt.Println(T("使用 --security-check 执行安全检查"))
	fmt.Println(T("使用 --help 显示帮助信息"))
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

	// 运行简单安全检查
	log.Println("[INFO] === 文件安全检查 ===")
	fmt.Println(T("\n=== 文件安全检查 ==="))

	// 扫描当前目录
	currentDir, _ := os.Getwd()
	fmt.Printf(T("正在检查当前目录: %s\n"), currentDir)

	result := RunSimpleSecurityCheck(currentDir)

	fmt.Printf(T("检查完成，耗时: %v\n"), result.Duration)
	fmt.Printf(T("检查文件: %d\n"), result.FilesScanned)
	fmt.Printf(T("安全提醒: %d\n"), result.Warnings)

	if result.Warnings > 0 {
		fmt.Println(T("⚠️  发现安全提醒:"))
		for _, msg := range result.Messages {
			fmt.Printf("  %s\n", msg)
		}
		fmt.Println(T("提示: 这些是安全提醒，帮助您识别潜在风险文件"))
	} else {
		fmt.Println(T("✅ 文件安全检查通过"))
	}

	// 生成安全建议
	log.Println("[INFO] === 安全建议 ===")
	fmt.Println(T("\n=== 安全建议 ==="))
	fmt.Println(T("• 定期更新系统补丁"))
	fmt.Println(T("• 定期备份重要数据"))
	fmt.Println(T("• 使用强密码策略"))
	fmt.Println(T("• 检查文件权限设置"))

	log.Println("[INFO] 提示: 安全检查完成")
	fmt.Println(T("\n提示: 安全检查完成"))
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

// SimpleSecurityChecker 简单的安全检查器
type SimpleSecurityChecker struct {
	Name string
}

// BasicSecurityResult 基础安全检查结果
type BasicSecurityResult struct {
	FilesScanned int
	Warnings     int
	Messages     []string
	ScanPath     string
	Duration     time.Duration
}

// CheckFile 检查单个文件的安全性
func (s *SimpleSecurityChecker) CheckFile(filePath string) *BasicSecurityResult {
	result := &BasicSecurityResult{
		FilesScanned: 1,
		ScanPath:     filePath,
		Messages:     []string{},
	}

	// 基础安全检查：文件扩展名检查
	ext := strings.ToLower(filepath.Ext(filePath))
	suspiciousExts := map[string]bool{
		".exe": true, ".bat": true, ".cmd": true, ".scr": true, ".pif": true,
		".com": true, ".vbs": true, ".js": true, ".jar": true,
	}

	if suspiciousExts[ext] {
		result.Warnings++
		result.Messages = append(result.Messages, fmt.Sprintf("⚠️  可执行文件: %s", filePath))
	}

	// 检查文件名是否包含可疑关键词
	name := strings.ToLower(filepath.Base(filePath))
	suspiciousKeywords := []string{"malware", "virus", "trojan", "hack", "crack", "keygen"}
	for _, keyword := range suspiciousKeywords {
		if strings.Contains(name, keyword) {
			result.Warnings++
			result.Messages = append(result.Messages, fmt.Sprintf("⚠️  可疑文件名: %s (包含 '%s')", filePath, keyword))
			break
		}
	}

	return result
}

// CheckDirectory 检查目录的安全性
func (s *SimpleSecurityChecker) CheckDirectory(dirPath string) *BasicSecurityResult {
	result := &BasicSecurityResult{
		ScanPath: dirPath,
		Messages: []string{},
	}

	startTime := time.Now()

	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		fileResult := s.CheckFile(path)
		result.FilesScanned += fileResult.FilesScanned
		result.Warnings += fileResult.Warnings
		result.Messages = append(result.Messages, fileResult.Messages...)

		return nil
	})

	result.Duration = time.Since(startTime)
	return result
}

// RunSimpleSecurityCheck 运行简单的安全检查
func RunSimpleSecurityCheck(targetPath string) *BasicSecurityResult {
	checker := &SimpleSecurityChecker{Name: "DelGuard安全助手"}

	info, err := os.Stat(targetPath)
	if err != nil {
		return &BasicSecurityResult{
			ScanPath: targetPath,
			Messages: []string{fmt.Sprintf("❌ 无法访问路径: %v", err)},
		}
	}

	if info.IsDir() {
		return checker.CheckDirectory(targetPath)
	} else {
		return checker.CheckFile(targetPath)
	}
}

// createDefaultSecurityConfig 创建默认安全配置
func createDefaultSecurityConfig(filePath string) error {
	defaultConfig := `{
  "security": {
    "enabled": true,
    "check_interval": "24h",
    "max_file_size": "100MB",
    "suspicious_extensions": [".exe", ".bat", ".cmd", ".scr", ".pif"],
    "suspicious_keywords": ["malware", "virus", "trojan", "hack"]
  },
  "alerts": {
    "enabled": true,
    "email": "",
    "webhook": ""
  }
}`

	return os.WriteFile(filePath, []byte(defaultConfig), 0644)
}
