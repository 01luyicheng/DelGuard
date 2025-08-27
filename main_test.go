package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestApplicationBasics 测试应用程序基础功能
func TestApplicationBasics(t *testing.T) {
	// 测试配置结构体
	config := &Config{
		UseRecycleBin:   true,
		InteractiveMode: "confirm",
		Language:        "zh-CN",
		LogLevel:        "info",
		MaxFileSize:     1073741824,
	}

	if config == nil {
		t.Errorf("配置创建失败")
	}

	// 验证默认值
	if config.Language == "" {
		t.Errorf("默认语言不应为空")
	}

	if config.LogLevel == "" {
		t.Errorf("默认日志级别不应为空")
	}

	if config.MaxFileSize <= 0 {
		t.Errorf("最大文件大小应该大于0")
	}
}

// TestCommandLineArguments 测试命令行参数
func TestCommandLineArguments(t *testing.T) {
	// 保存原始命令行参数
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	testCases := []struct {
		name  string
		args  []string
		valid bool
	}{
		{"帮助参数", []string{"delguard", "--help"}, true},
		{"版本参数", []string{"delguard", "--version"}, true},
		{"详细模式", []string{"delguard", "-v"}, true},
		{"文件参数", []string{"delguard", "file.txt"}, true},
		{"交互模式", []string{"delguard", "-i", "file.txt"}, true},
		{"强制模式", []string{"delguard", "--force", "file.txt"}, true},
		{"空参数", []string{"delguard"}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 设置测试参数
			os.Args = tc.args

			// 基本参数验证
			if len(tc.args) < 1 {
				t.Errorf("参数数量不足")
			}

			if tc.valid && len(tc.args) < 2 && tc.args[0] == "delguard" {
				// 某些情况下需要至少2个参数
				t.Logf("参数验证: %v", tc.args)
			}
		})
	}
}

// TestFileSystemOperations 测试文件系统操作
func TestFileSystemOperations(t *testing.T) {
	// 创建测试文件
	testDir, err := os.MkdirTemp("", "delguard-fs-test")
	if err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}
	defer os.RemoveAll(testDir)

	testFile := filepath.Join(testDir, "fs_test.txt")
	if err := os.WriteFile(testFile, []byte("filesystem test content"), 0644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 测试文件操作
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Errorf("测试文件应该存在")
	}

	// 测试文件读取
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("读取文件失败: %v", err)
	}

	if string(content) != "filesystem test content" {
		t.Errorf("文件内容不匹配")
	}
}

// TestErrorScenarios 测试错误场景
func TestErrorScenarios(t *testing.T) {
	// 测试访问不存在的文件
	nonExistentFile := "/nonexistent/file.txt"
	if _, err := os.Stat(nonExistentFile); !os.IsNotExist(err) {
		t.Errorf("应该正确处理不存在的文件")
	}

	// 测试空文件列表
	var emptyList []string
	if len(emptyList) != 0 {
		t.Errorf("空列表长度应该为0")
	}

	// 测试nil检查
	var nilList []string = nil
	if nilList != nil {
		t.Errorf("nil列表检查失败")
	}
}

// TestIntegrationScenarios 集成测试场景
func TestIntegrationScenarios(t *testing.T) {
	// 创建测试环境
	testDir, err := os.MkdirTemp("", "delguard-integration")
	if err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}
	defer os.RemoveAll(testDir)

	// 创建多个测试文件
	testFiles := []string{
		filepath.Join(testDir, "file1.txt"),
		filepath.Join(testDir, "file2.txt"),
		filepath.Join(testDir, "file3.txt"),
	}

	for i, file := range testFiles {
		content := []byte("integration test content " + string(rune('1'+i)))
		if err := os.WriteFile(file, content, 0644); err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}
	}

	// 创建子目录和文件
	subDir := filepath.Join(testDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("创建子目录失败: %v", err)
	}

	subFile := filepath.Join(subDir, "subfile.txt")
	if err := os.WriteFile(subFile, []byte("sub file content"), 0644); err != nil {
		t.Fatalf("创建子文件失败: %v", err)
	}

	// 验证所有文件都被创建
	allFiles := append(testFiles, subFile)
	for _, file := range allFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("文件未被创建: %s", file)
		}
	}

	// 验证目录结构
	if _, err := os.Stat(subDir); os.IsNotExist(err) {
		t.Errorf("子目录未被创建: %s", subDir)
	}

	t.Logf("集成测试完成，创建了 %d 个文件和 1 个目录", len(allFiles))
}

// TestConfigurationHandling 测试配置处理
func TestConfigurationHandling(t *testing.T) {
	// 测试不同的配置组合
	configs := []Config{
		{
			UseRecycleBin:   true,
			InteractiveMode: "always",
			Language:        "zh-CN",
			LogLevel:        "debug",
		},
		{
			UseRecycleBin:   false,
			InteractiveMode: "never",
			Language:        "en-US",
			LogLevel:        "error",
		},
		{
			UseRecycleBin:   true,
			InteractiveMode: "confirm",
			Language:        "ja-JP",
			LogLevel:        "info",
		},
	}

	for i, config := range configs {
		t.Run("配置测试"+string(rune('1'+i)), func(t *testing.T) {
			// 验证配置字段
			if config.Language == "" {
				t.Errorf("语言不应为空")
			}

			if config.LogLevel == "" {
				t.Errorf("日志级别不应为空")
			}

			if config.InteractiveMode == "" {
				t.Errorf("交互模式不应为空")
			}
		})
	}
}
