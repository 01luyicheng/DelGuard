package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestInputValidator_ValidatePath 测试路径验证
func TestInputValidator_ValidatePath(t *testing.T) {
	validator := NewInputValidator(&Config{SafeMode: "normal"})

	tests := []struct {
		name        string
		path        string
		expectValid bool
		expectError string
	}{
		{
			name:        "Valid normal path",
			path:        "/home/user/file.txt",
			expectValid: true,
		},
		{
			name:        "Empty path",
			path:        "",
			expectValid: false,
			expectError: "路径不能为空",
		},
		{
			name:        "Path traversal attack",
			path:        "../../../etc/passwd",
			expectValid: false,
			expectError: "路径遍历攻击",
		},
		{
			name:        "Null byte injection",
			path:        "file.txt\x00",
			expectValid: false,
			expectError: "空字节注入攻击",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidatePath(tt.path)

			if tt.expectValid && !result.IsValid {
				t.Errorf("Expected path to be valid, but got errors: %v", result.Errors)
			}

			if !tt.expectValid && result.IsValid {
				t.Errorf("Expected path to be invalid, but validation passed")
			}

			if !tt.expectValid && tt.expectError != "" {
				found := false
				for _, err := range result.Errors {
					if strings.Contains(err, tt.expectError) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error containing '%s', but got: %v", tt.expectError, result.Errors)
				}
			}
		})
	}
}

// TestSpecialFileHandler_AnalyzeFile 测试特殊文件分析
func TestSpecialFileHandler_AnalyzeFile(t *testing.T) {
	config := &Config{}
	handler := NewSpecialFileHandler(config)

	// 创建测试文件
	tempDir := t.TempDir()

	// 测试隐藏文件
	hiddenFile := filepath.Join(tempDir, ".hidden")
	if err := os.WriteFile(hiddenFile, []byte("hidden content"), 0644); err != nil {
		t.Fatal(err)
	}

	issues, err := handler.AnalyzeFile(hiddenFile)
	if err != nil {
		t.Fatalf("Failed to analyze hidden file: %v", err)
	}

	// 检查是否检测到隐藏文件
	found := false
	for _, issue := range issues {
		if issue.Type == "hidden_file" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Failed to detect hidden file")
	}
}

// TestConcurrencyManager 测试并发管理器
func TestConcurrencyManager(t *testing.T) {
	maxConcurrent := 3
	cm := NewConcurrencyManager(maxConcurrent)
	defer cm.Close()

	// 测试基本操作获取和释放
	op, err := cm.AcquireOperation("test", "/tmp/test")
	if err != nil {
		t.Fatalf("Failed to acquire operation: %v", err)
	}

	cm.ReleaseOperation(op)

	// 测试并发限制
	var ops []*Operation
	for i := 0; i < maxConcurrent; i++ {
		op, err := cm.AcquireOperation("test", fmt.Sprintf("/tmp/test%d", i))
		if err != nil {
			t.Fatalf("Failed to acquire operation %d: %v", i, err)
		}
		ops = append(ops, op)
	}

	// 释放操作
	for _, op := range ops {
		cm.ReleaseOperation(op)
	}
}

// TestResourceManager 测试资源管理器
func TestResourceManager(t *testing.T) {
	rm := NewResourceManager()
	defer rm.Close()

	// 测试文件打开和关闭
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	// 创建测试文件
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	// 通过资源管理器打开文件
	managedFile, err := rm.OpenFile(testFile, os.O_RDONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to open file through resource manager: %v", err)
	}

	// 检查文件是否正确打开
	if managedFile.Path != testFile {
		t.Errorf("Expected path %s, got %s", testFile, managedFile.Path)
	}

	// 关闭文件
	err = rm.CloseFile(testFile)
	if err != nil {
		t.Errorf("Failed to close file: %v", err)
	}
}

// TestConfigValidation 测试配置验证
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectValid bool
		expectError string
	}{
		{
			name: "Valid config",
			config: &Config{
				Version:             "1.0.0",
				InteractiveMode:     "confirm",
				SafeMode:            "normal",
				LogLevel:            "info",
				Language:            "auto",
				MaxFileSize:         1024 * 1024 * 1024,
				MaxPathLength:       4096,
				MaxConcurrentOps:    10,
				MaxBackupFiles:      1000,
				TrashMaxSize:        10000,
				BackupRetentionDays: 30,
				LogRetentionDays:    7,
			},
			expectValid: true,
		},
		{
			name: "Invalid log level",
			config: &Config{
				Version:         "1.0.0",
				InteractiveMode: "confirm",
				SafeMode:        "normal",
				LogLevel:        "invalid",
				Language:        "auto",
			},
			expectValid: false,
			expectError: "无效的日志级别",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectValid && err != nil {
				t.Errorf("Expected config to be valid, but got error: %v", err)
			}

			if !tt.expectValid && err == nil {
				t.Error("Expected config to be invalid, but validation passed")
			}

			if !tt.expectValid && tt.expectError != "" && err != nil {
				if !strings.Contains(err.Error(), tt.expectError) {
					t.Errorf("Expected error containing '%s', but got: %v", tt.expectError, err)
				}
			}
		})
	}
}

// TestConcurrentOperations 测试并发操作
func TestConcurrentOperations(t *testing.T) {
	cm := NewConcurrencyManager(5)
	defer cm.Close()

	var wg sync.WaitGroup
	successCount := 0
	mu := sync.Mutex{}

	// 启动多个并发操作
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			err := cm.SafeExecute("test", fmt.Sprintf("/tmp/test%d", id), func(ctx context.Context) error {
				// 模拟一些工作
				select {
				case <-time.After(50 * time.Millisecond):
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			})

			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// 检查并发限制是否正常工作
	if successCount < 5 {
		t.Errorf("Expected at least 5 successful operations, got %d", successCount)
	}
}

// TestEdgeCases 测试边缘情况
func TestEdgeCases(t *testing.T) {
	// 测试空输入
	validator := NewInputValidator(&Config{SafeMode: "strict"})

	result := validator.ValidatePath("")
	if result.IsValid {
		t.Error("Empty path should be invalid")
	}

	// 测试非常长的输入
	longPath := strings.Repeat("a/", 2000)
	result = validator.ValidatePath(longPath)
	if result.IsValid {
		t.Error("Very long path should be invalid")
	}

	// 测试特殊字符
	specialPath := "file\x00name"
	result = validator.ValidatePath(specialPath)
	if result.IsValid {
		t.Error("Path with null byte should be invalid")
	}
}

// BenchmarkInputValidation 性能测试
func BenchmarkInputValidation(b *testing.B) {
	validator := NewInputValidator(&Config{SafeMode: "normal"})
	testPath := "/home/user/documents/test.txt"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidatePath(testPath)
	}
}
