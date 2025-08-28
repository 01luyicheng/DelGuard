package delete

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDeleteService_Integration(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()
	
	// 创建配置
	config := &Config{
		MaxConcurrency: 5,
		ProtectedPaths: []string{
			"C:\\Windows",
			"/usr/bin",
		},
		EnableLogging: true,
	}
	
	// 创建服务
	service := NewService(config)
	
	// 创建测试文件
	testFiles := []string{
		filepath.Join(tempDir, "file1.txt"),
		filepath.Join(tempDir, "file2.txt"),
		filepath.Join(tempDir, "file3.txt"),
		filepath.Join(tempDir, "subdir", "file4.txt"),
	}
	
	// 确保子目录存在
	if err := os.MkdirAll(filepath.Join(tempDir, "subdir"), 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	
	// 创建测试文件
	for _, file := range testFiles {
		content := []byte("test content for integration test")
		if err := os.WriteFile(file, content, 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}
	
	// 测试单个文件删除
	t.Run("single file delete", func(t *testing.T) {
		err := service.SafeDelete(testFiles[0])
		if err != nil {
			t.Errorf("SafeDelete() error = %v", err)
		}
		
		// 验证文件被删除
		if _, err := os.Stat(testFiles[0]); !os.IsNotExist(err) {
			t.Errorf("File should be deleted but still exists")
		}
	})
	
	// 测试批量删除
	t.Run("batch delete", func(t *testing.T) {
		remainingFiles := testFiles[1:]
		results := service.BatchDelete(remainingFiles)
		
		if len(results) != len(remainingFiles) {
			t.Errorf("Expected %d results, got %d", len(remainingFiles), len(results))
		}
		
		for _, result := range results {
			if !result.Success {
				t.Errorf("BatchDelete failed for file %s: %v", result.Path, result.Error)
			}
		}
	})
	
	// 测试统计信息
	t.Run("metrics", func(t *testing.T) {
		metrics := service.GetMetrics()
		
		if metrics.TotalOperations == 0 {
			t.Error("Expected some operations to be recorded")
		}
		
		if metrics.SuccessfulDeletes == 0 {
			t.Error("Expected some successful deletes to be recorded")
		}
		
		t.Logf("Metrics: %s", metrics.String())
	})
}

func TestDeleteService_ErrorHandling(t *testing.T) {
	service := NewService()
	
	tests := []struct {
		name     string
		filePath string
		wantErr  bool
		errType  ErrorCode
	}{
		{
			name:     "non-existent file",
			filePath: "/non/existent/file.txt",
			wantErr:  true,
			errType:  ErrFileNotFound,
		},
		{
			name:     "protected path",
			filePath: "C:\\Windows\\system32\\kernel32.dll",
			wantErr:  true,
			errType:  ErrProtectedPath,
		},
		{
			name:     "empty path",
			filePath: "",
			wantErr:  true,
			errType:  ErrInvalidPath,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.SafeDelete(tt.filePath)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeDelete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if tt.wantErr {
				if deleteErr, ok := err.(*DeleteError); ok {
					if deleteErr.Code != tt.errType {
						t.Errorf("Expected error type %v, got %v", tt.errType, deleteErr.Code)
					}
				} else {
					t.Errorf("Expected DeleteError, got %T", err)
				}
			}
		})
	}
}

func TestDeleteService_ConcurrentOperations(t *testing.T) {
	tempDir := t.TempDir()
	config := &Config{
		MaxConcurrency: 3,
		ProtectedPaths: []string{},
		EnableLogging:  false,
	}
	
	service := NewService(config)
	
	// 创建多个测试文件
	var testFiles []string
	for i := 0; i < 10; i++ {
		file := filepath.Join(tempDir, fmt.Sprintf("concurrent_test_%d.txt", i))
		testFiles = append(testFiles, file)
		
		if err := os.WriteFile(file, []byte("concurrent test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}
	
	// 测试并发批量删除
	ctx := context.WithTimeout(context.Background(), 10*time.Second)
	results := service.BatchDeleteWithContext(ctx, testFiles)
	
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			t.Errorf("Concurrent delete failed for %s: %v", result.Path, result.Error)
		}
	}
	
	if successCount != len(testFiles) {
		t.Errorf("Expected %d successful deletes, got %d", len(testFiles), successCount)
	}
	
	// 检查统计信息
	metrics := service.GetMetrics()
	if metrics.MaxConcurrency == 0 {
		t.Error("Expected max concurrency to be recorded")
	}
	
	t.Logf("Max concurrency recorded: %d", metrics.MaxConcurrency)
}

func TestDeleteService_ConfigManagement(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.json")
	
	// 创建配置管理器
	cm := NewConfigManager(configPath)
	
	// 测试保存配置
	originalConfig := &Config{
		MaxConcurrency: 8,
		ProtectedPaths: []string{"/test/path"},
		EnableLogging:  true,
	}
	
	if err := cm.SaveConfig(originalConfig); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
	
	// 测试加载配置
	loadedConfig, err := cm.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	if loadedConfig.MaxConcurrency != originalConfig.MaxConcurrency {
		t.Errorf("Expected MaxConcurrency %d, got %d", 
			originalConfig.MaxConcurrency, loadedConfig.MaxConcurrency)
	}
	
	if len(loadedConfig.ProtectedPaths) != len(originalConfig.ProtectedPaths) {
		t.Errorf("Expected %d protected paths, got %d", 
			len(originalConfig.ProtectedPaths), len(loadedConfig.ProtectedPaths))
	}
	
	if loadedConfig.EnableLogging != originalConfig.EnableLogging {
		t.Errorf("Expected EnableLogging %v, got %v", 
			originalConfig.EnableLogging, loadedConfig.EnableLogging)
	}
}

func TestDeleteService_LoggingIntegration(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")
	
	// 创建文件日志记录器
	logger, err := NewFileLogger(LogLevelDebug, logFile)
	if err != nil {
		t.Fatalf("Failed to create file logger: %v", err)
	}
	defer logger.Close()
	
	// 创建带有自定义日志记录器的服务
	config := DefaultConfig()
	service := NewServiceWithLogger(config, logger)
	
	// 创建测试文件
	testFile := filepath.Join(tempDir, "log_test.txt")
	if err := os.WriteFile(testFile, []byte("log test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// 执行删除操作
	if err := service.SafeDelete(testFile); err != nil {
		t.Errorf("SafeDelete() error = %v", err)
	}
	
	// 检查日志文件是否存在且有内容
	if stat, err := os.Stat(logFile); err != nil {
		t.Errorf("Log file should exist: %v", err)
	} else if stat.Size() == 0 {
		t.Error("Log file should not be empty")
	}
}