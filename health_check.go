package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	Component   string    `json:"component"`
	Status      string    `json:"status"` // "ok", "warning", "error"
	Message     string    `json:"message"`
	Details     []string  `json:"details,omitempty"`
	CheckedAt   time.Time `json:"checked_at"`
	Suggestions []string  `json:"suggestions,omitempty"`
}

// HealthChecker 系统健康检查器
type HealthChecker struct {
	results []HealthCheckResult
	verbose bool
}

// NewHealthChecker 创建新的健康检查器
func NewHealthChecker(verbose bool) *HealthChecker {
	return &HealthChecker{
		results: make([]HealthCheckResult, 0),
		verbose: verbose,
	}
}

// RunFullCheck 运行完整的系统健康检查
func (hc *HealthChecker) RunFullCheck() error {
	fmt.Println("🔍 开始系统健康检查...")
	fmt.Println()

	// 检查核心组件
	hc.checkCoreComponents()

	// 检查配置文件
	hc.checkConfigFiles()

	// 检查依赖项
	hc.checkDependencies()

	// 检查权限
	hc.checkPermissions()

	// 检查磁盘空间
	hc.checkDiskSpace()

	// 检查语言文件
	hc.checkLanguageFiles()

	// 显示检查结果
	hc.displayResults()

	return nil
}

// checkCoreComponents 检查核心组件
func (hc *HealthChecker) checkCoreComponents() {
	fmt.Print("📦 检查核心组件... ")

	coreFiles := []string{
		"main.go",
		"config.go",
		"protect.go",
		"file_operations.go",
		"types.go",
	}

	var missingFiles []string
	var existingFiles []string

	for _, file := range coreFiles {
		if _, err := os.Stat(file); err != nil {
			missingFiles = append(missingFiles, file)
		} else {
			existingFiles = append(existingFiles, file)
		}
	}

	if len(missingFiles) == 0 {
		hc.addResult("核心组件", "ok", "所有核心文件完整", existingFiles, nil)
		fmt.Println("✅")
	} else if len(missingFiles) < len(coreFiles)/2 {
		suggestions := []string{
			"检查是否在正确的项目目录中",
			"从备份或版本控制恢复缺失文件",
		}
		hc.addResult("核心组件", "warning", fmt.Sprintf("缺少 %d 个核心文件", len(missingFiles)), missingFiles, suggestions)
		fmt.Println("⚠️")
	} else {
		suggestions := []string{
			"重新下载或克隆完整项目",
			"检查文件权限和磁盘空间",
		}
		hc.addResult("核心组件", "error", "缺少大量核心文件", missingFiles, suggestions)
		fmt.Println("❌")
	}
}

// checkConfigFiles 检查配置文件
func (hc *HealthChecker) checkConfigFiles() {
	fmt.Print("⚙️  检查配置文件... ")

	configPaths := []string{
		"config",
		"config/languages",
	}

	var issues []string
	var validConfigs []string

	// 检查配置目录
	for _, path := range configPaths {
		if info, err := os.Stat(path); err != nil {
			issues = append(issues, fmt.Sprintf("目录不存在: %s", path))
		} else if !info.IsDir() {
			issues = append(issues, fmt.Sprintf("不是目录: %s", path))
		} else {
			validConfigs = append(validConfigs, path)
		}
	}

	// 检查默认配置文件
	if _, err := LoadConfig(); err != nil {
		issues = append(issues, fmt.Sprintf("配置加载失败: %v", err))
	} else {
		validConfigs = append(validConfigs, "默认配置")
	}

	// 检查语言文件
	langDir := "config/languages"
	if entries, err := os.ReadDir(langDir); err == nil {
		langCount := 0
		for _, entry := range entries {
			if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".json") ||
				strings.HasSuffix(entry.Name(), ".jsonc") ||
				strings.HasSuffix(entry.Name(), ".yaml") ||
				strings.HasSuffix(entry.Name(), ".yml") ||
				strings.HasSuffix(entry.Name(), ".toml") ||
				strings.HasSuffix(entry.Name(), ".ini") ||
				strings.HasSuffix(entry.Name(), ".properties")) {
				langCount++
			}
		}
		validConfigs = append(validConfigs, fmt.Sprintf("%d 个语言文件", langCount))
	}

	if len(issues) == 0 {
		hc.addResult("配置文件", "ok", "配置文件完整且有效", validConfigs, nil)
		fmt.Println("✅")
	} else if len(issues) <= 2 {
		suggestions := []string{
			"运行交互式配置生成器重新创建配置",
			"检查配置文件语法和格式",
		}
		hc.addResult("配置文件", "warning", "配置存在轻微问题", issues, suggestions)
		fmt.Println("⚠️")
	} else {
		suggestions := []string{
			"运行 delguard --init-config 重新初始化配置",
			"从备份恢复配置文件",
		}
		hc.addResult("配置文件", "error", "配置文件存在严重问题", issues, suggestions)
		fmt.Println("❌")
	}
}

// checkDependencies 检查依赖项
func (hc *HealthChecker) checkDependencies() {
	fmt.Print("📚 检查依赖项... ")

	var issues []string
	var validDeps []string

	// 检查 go.mod 文件
	if _, err := os.Stat("go.mod"); err != nil {
		issues = append(issues, "go.mod 文件不存在")
	} else {
		validDeps = append(validDeps, "go.mod")

		// 尝试检查模块依赖
		if content, err := os.ReadFile("go.mod"); err == nil {
			contentStr := string(content)
			if strings.Contains(contentStr, "module") {
				validDeps = append(validDeps, "模块定义正常")
			} else {
				issues = append(issues, "go.mod 缺少模块定义")
			}
		}
	}

	// 检查关键目录
	dirs := []string{"utils", "config"}
	for _, dir := range dirs {
		if info, err := os.Stat(dir); err != nil {
			issues = append(issues, fmt.Sprintf("目录不存在: %s", dir))
		} else if info.IsDir() {
			validDeps = append(validDeps, dir)
		}
	}

	if len(issues) == 0 {
		hc.addResult("依赖项", "ok", "所有依赖项正常", validDeps, nil)
		fmt.Println("✅")
	} else {
		suggestions := []string{
			"运行 go mod tidy 整理依赖",
			"运行 go mod download 下载依赖",
		}
		hc.addResult("依赖项", "warning", "依赖项存在问题", issues, suggestions)
		fmt.Println("⚠️")
	}
}

// checkPermissions 检查权限
func (hc *HealthChecker) checkPermissions() {
	fmt.Print("🔐 检查权限... ")

	var issues []string
	var validPerms []string

	// 检查当前目录权限
	if info, err := os.Stat("."); err != nil {
		issues = append(issues, "无法访问当前目录")
	} else {
		mode := info.Mode()
		if runtime.GOOS == "windows" {
			validPerms = append(validPerms, "Windows目录访问正常")
		} else {
			if mode&0200 != 0 {
				validPerms = append(validPerms, "目录写权限正常")
			} else {
				issues = append(issues, "目录缺少写权限")
			}
		}
	}

	// 检查配置目录权限
	configDir := "config"
	if info, err := os.Stat(configDir); err == nil {
		if info.IsDir() {
			if entries, err := os.ReadDir(configDir); err == nil {
				validPerms = append(validPerms, "配置目录可读")
				_ = entries
			} else {
				issues = append(issues, "配置目录无法读取")
			}
		}
	}

	// 检查临时目录权限
	tempDir := os.TempDir()
	if tempFile, err := os.CreateTemp(tempDir, "delguard_test_*"); err == nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		validPerms = append(validPerms, "临时目录写权限正常")
	} else {
		issues = append(issues, "临时目录无写权限")
	}

	if len(issues) == 0 {
		hc.addResult("权限检查", "ok", "权限配置正常", validPerms, nil)
		fmt.Println("✅")
	} else {
		suggestions := []string{
			"以管理员权限运行程序",
			"检查文件系统权限设置",
		}
		hc.addResult("权限检查", "warning", "权限存在问题", issues, suggestions)
		fmt.Println("⚠️")
	}
}

// checkDiskSpace 检查磁盘空间
func (hc *HealthChecker) checkDiskSpace() {
	fmt.Print("💾 检查磁盘空间... ")

	var issues []string
	var validSpace []string

	// 获取当前目录磁盘使用情况
	if usage, err := getDiskUsage("."); err == nil {
		freeGB := float64(usage.Free) / (1024 * 1024 * 1024)
		totalGB := float64(usage.Total) / (1024 * 1024 * 1024)
		usedPercent := float64(usage.Used) / float64(usage.Total) * 100

		validSpace = append(validSpace, fmt.Sprintf("可用空间: %.1f GB", freeGB))
		validSpace = append(validSpace, fmt.Sprintf("总空间: %.1f GB", totalGB))
		validSpace = append(validSpace, fmt.Sprintf("使用率: %.1f%%", usedPercent))

		if freeGB < 1.0 {
			issues = append(issues, "磁盘空间不足 1GB")
		} else if usedPercent > 95 {
			issues = append(issues, "磁盘使用率超过 95%")
		}
	} else {
		issues = append(issues, "无法获取磁盘使用信息")
	}

	if len(issues) == 0 {
		hc.addResult("磁盘空间", "ok", "磁盘空间充足", validSpace, nil)
		fmt.Println("✅")
	} else {
		suggestions := []string{
			"清理临时文件释放空间",
			"移动大文件到其他磁盘",
		}
		hc.addResult("磁盘空间", "warning", "磁盘空间紧张", issues, suggestions)
		fmt.Println("⚠️")
	}
}

// checkLanguageFiles 检查语言文件
func (hc *HealthChecker) checkLanguageFiles() {
	fmt.Print("🌐 检查语言文件... ")

	var issues []string
	var validLangs []string

	langDir := "config/languages"
	if entries, err := os.ReadDir(langDir); err != nil {
		issues = append(issues, "无法读取语言目录")
	} else {
		validCount := 0
		invalidCount := 0

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			filename := entry.Name()
			if strings.HasSuffix(filename, ".md") || strings.HasSuffix(filename, ".txt") {
				continue // 跳过文档文件
			}

			filePath := filepath.Join(langDir, filename)

			// 检查文件格式
			if strings.HasSuffix(filename, ".json") || strings.HasSuffix(filename, ".jsonc") {
				if hc.validateJSONFile(filePath) {
					validCount++
				} else {
					invalidCount++
					issues = append(issues, fmt.Sprintf("JSON格式错误: %s", filename))
				}
			} else if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
				validCount++ // 简化检查
			} else if strings.HasSuffix(filename, ".toml") {
				validCount++ // 简化检查
			} else if strings.HasSuffix(filename, ".ini") || strings.HasSuffix(filename, ".properties") {
				validCount++ // 简化检查
			} else {
				issues = append(issues, fmt.Sprintf("未知格式: %s", filename))
			}
		}

		validLangs = append(validLangs, fmt.Sprintf("有效语言文件: %d", validCount))
		if invalidCount > 0 {
			validLangs = append(validLangs, fmt.Sprintf("无效文件: %d", invalidCount))
		}
	}

	if len(issues) == 0 {
		hc.addResult("语言文件", "ok", "语言文件完整有效", validLangs, nil)
		fmt.Println("✅")
	} else if len(issues) <= 2 {
		suggestions := []string{
			"检查语言文件语法",
			"使用配置生成器重新创建语言文件",
		}
		hc.addResult("语言文件", "warning", "语言文件存在问题", issues, suggestions)
		fmt.Println("⚠️")
	} else {
		suggestions := []string{
			"重新初始化语言配置",
			"从备份恢复语言文件",
		}
		hc.addResult("语言文件", "error", "语言文件存在严重问题", issues, suggestions)
		fmt.Println("❌")
	}
}

// validateJSONFile 验证JSON文件格式
func (hc *HealthChecker) validateJSONFile(filePath string) bool {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}

	// 处理JSONC格式（移除注释）
	if strings.HasSuffix(filePath, ".jsonc") {
		content = removeJSONComments(content)
	}

	var data interface{}
	return json.Unmarshal(content, &data) == nil
}

// removeJSONComments 移除JSON注释（简化实现）
func removeJSONComments(content []byte) []byte {
	lines := strings.Split(string(content), "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "/*") {
			result = append(result, line)
		}
	}

	return []byte(strings.Join(result, "\n"))
}

// addResult 添加检查结果
func (hc *HealthChecker) addResult(component, status, message string, details, suggestions []string) {
	result := HealthCheckResult{
		Component:   component,
		Status:      status,
		Message:     message,
		Details:     details,
		CheckedAt:   time.Now(),
		Suggestions: suggestions,
	}
	hc.results = append(hc.results, result)
}

// displayResults 显示检查结果
func (hc *HealthChecker) displayResults() {
	fmt.Println()
	fmt.Println("📋 健康检查报告")
	fmt.Println("=" + strings.Repeat("=", 50))

	okCount := 0
	warningCount := 0
	errorCount := 0

	for _, result := range hc.results {
		switch result.Status {
		case "ok":
			okCount++
			fmt.Printf("✅ %s: %s\n", result.Component, result.Message)
		case "warning":
			warningCount++
			fmt.Printf("⚠️  %s: %s\n", result.Component, result.Message)
		case "error":
			errorCount++
			fmt.Printf("❌ %s: %s\n", result.Component, result.Message)
		}

		if hc.verbose && len(result.Details) > 0 {
			for _, detail := range result.Details {
				fmt.Printf("   • %s\n", detail)
			}
		}

		if result.Status != "ok" && len(result.Suggestions) > 0 {
			fmt.Printf("   💡 建议:\n")
			for _, suggestion := range result.Suggestions {
				fmt.Printf("      - %s\n", suggestion)
			}
		}
		fmt.Println()
	}

	// 总结
	fmt.Println("📊 检查总结")
	fmt.Println("-" + strings.Repeat("-", 30))
	fmt.Printf("✅ 正常: %d\n", okCount)
	fmt.Printf("⚠️  警告: %d\n", warningCount)
	fmt.Printf("❌ 错误: %d\n", errorCount)

	if errorCount > 0 {
		fmt.Println("\n🚨 发现严重问题，建议立即修复！")
	} else if warningCount > 0 {
		fmt.Println("\n⚠️  发现一些问题，建议尽快处理。")
	} else {
		fmt.Println("\n🎉 系统状态良好！")
	}
}

// GetResults 获取检查结果
func (hc *HealthChecker) GetResults() []HealthCheckResult {
	return hc.results
}

// ExportResults 导出检查结果到JSON文件
func (hc *HealthChecker) ExportResults(filename string) error {
	data, err := json.MarshalIndent(hc.results, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
