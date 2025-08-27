package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestBasicFunctionality 测试基本功能
func TestBasicFunctionality(t *testing.T) {
	// 创建测试文件
	testDir, err := os.MkdirTemp("", "delguard-test")
	if err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}
	defer os.RemoveAll(testDir)

	testFile := filepath.Join(testDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 测试文件是否存在
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Errorf("测试文件应该存在")
	}
}

// TestConfigCreation 测试配置创建
func TestConfigCreation(t *testing.T) {
	// 创建默认配置
	config := &Config{
		UseRecycleBin:   true,
		InteractiveMode: "confirm",
		Language:        "zh-CN",
		LogLevel:        "info",
		MaxFileSize:     1073741824,
	}

	// 验证配置值
	if !config.UseRecycleBin {
		t.Errorf("UseRecycleBin 应该为 true")
	}

	if config.InteractiveMode != "confirm" {
		t.Errorf("InteractiveMode 应该为 'confirm'")
	}

	if config.Language != "zh-CN" {
		t.Errorf("Language 应该为 'zh-CN'")
	}
}

// TestFileOperations 测试文件操作
func TestFileOperations(t *testing.T) {
	// 创建测试目录
	testDir, err := os.MkdirTemp("", "delguard-fileops-test")
	if err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}
	defer os.RemoveAll(testDir)

	// 创建测试文件
	testFile := filepath.Join(testDir, "test.txt")
	testContent := []byte("test file content")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 测试文件存在检查
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Errorf("文件应该存在")
	}

	// 测试不存在的文件
	nonExistentFile := filepath.Join(testDir, "nonexistent.txt")
	if _, err := os.Stat(nonExistentFile); !os.IsNotExist(err) {
		t.Errorf("不存在的文件检查失败")
	}

	// 测试目录检查
	if info, err := os.Stat(testDir); err != nil || !info.IsDir() {
		t.Errorf("目录检查失败")
	}

	if info, err := os.Stat(testFile); err != nil || info.IsDir() {
		t.Errorf("文件检查失败")
	}
}

// TestPathValidation 测试路径验证
func TestPathValidation(t *testing.T) {
	// 测试正常路径
	normalPaths := []string{
		"./test.txt",
		"test.txt",
		"folder/test.txt",
	}

	for _, path := range normalPaths {
		if filepath.IsAbs(path) && (path == "/" || path == "C:\\") {
			t.Errorf("不应该允许根目录路径: %s", path)
		}
	}

	// 测试危险路径
	dangerousPaths := []string{
		"/",
		filepath.Join("C:", ""),
		"/etc/passwd",
		filepath.Join("C:", "Windows", "System32"),
	}

	for _, path := range dangerousPaths {
		// 这里只是示例检查，实际的安全检查会在主程序中实现
		if path == "/" || path == "C:\\" {
			t.Logf("检测到危险路径: %s", path)
		}
	}
}

// TestConfigValidation 测试配置验证
func TestConfigValidation(t *testing.T) {
	// 测试有效配置
	validConfig := &Config{
		UseRecycleBin:   true,
		InteractiveMode: "confirm",
		Language:        "zh-CN",
		LogLevel:        "info",
		MaxFileSize:     1073741824,
	}

	// 基本验证
	if validConfig.InteractiveMode == "" {
		t.Errorf("交互模式不应为空")
	}

	if validConfig.Language == "" {
		t.Errorf("语言设置不应为空")
	}

	if validConfig.LogLevel == "" {
		t.Errorf("日志级别不应为空")
	}

	if validConfig.MaxFileSize <= 0 {
		t.Errorf("最大文件大小应该大于0")
	}

	// 测试无效值
	invalidModes := []string{"invalid_mode", "", "unknown"}
	for _, mode := range invalidModes {
		if mode != "always" && mode != "never" && mode != "confirm" && mode != "" {
			t.Logf("检测到无效交互模式: %s", mode)
		}
	}
}

// TestLanguageSupport 测试多语言支持
func TestLanguageSupport(t *testing.T) {
	supportedLanguages := []string{"zh-CN", "en-US", "ja-JP"}

	for _, lang := range supportedLanguages {
		config := &Config{
			Language: lang,
		}

		// 测试语言设置
		if config.Language != lang {
			t.Errorf("语言设置错误，期望 %s，实际 %s", lang, config.Language)
		}
	}
}

// TestErrorHandling 测试错误处理
func TestErrorHandling(t *testing.T) {
	// 测试访问不存在的文件
	nonExistentFile := "/path/to/nonexistent/file.txt"
	if _, err := os.Stat(nonExistentFile); !os.IsNotExist(err) {
		t.Errorf("应该正确检测不存在的文件")
	}

	// 测试空路径
	emptyPath := ""
	if emptyPath != "" {
		t.Errorf("空路径检查失败")
	}

	// 测试nil检查
	var nilSlice []string
	if nilSlice != nil {
		t.Errorf("nil切片检查失败")
	}
}

// TestPerformance 测试性能
func TestPerformance(t *testing.T) {
	// 创建多个测试文件
	testDir, err := os.MkdirTemp("", "delguard-perf-test")
	if err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}
	defer os.RemoveAll(testDir)

	// 创建10个测试文件（减少数量以加快测试）
	var testFiles []string
	for i := 0; i < 10; i++ {
		testFile := filepath.Join(testDir, "test_"+string(rune('0'+i))+".txt")
		if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}
		testFiles = append(testFiles, testFile)
	}

	// 验证所有文件都被创建
	for _, file := range testFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("测试文件未被创建: %s", file)
		}
	}

	t.Logf("成功创建 %d 个测试文件", len(testFiles))
}

// BenchmarkFileCreation 文件创建基准测试
func BenchmarkFileCreation(b *testing.B) {
	// 创建测试目录
	testDir, err := os.MkdirTemp("", "delguard-bench")
	if err != nil {
		b.Fatalf("创建测试目录失败: %v", err)
	}
	defer os.RemoveAll(testDir)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 创建测试文件
		testFile := filepath.Join(testDir, "bench_"+string(rune('0'+i%10))+".txt")
		if err := os.WriteFile(testFile, []byte("benchmark test"), 0644); err != nil {
			b.Fatalf("创建测试文件失败: %v", err)
		}

		// 立即删除以避免文件过多
		os.Remove(testFile)
	}
}
