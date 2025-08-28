package delete

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestDeleteService_ValidateFile(t *testing.T) {
	service := NewService()

	tests := []struct {
		name     string
		filePath string
		wantErr  bool
	}{
		{
			name:     "valid file path",
			filePath: "test.txt",
			wantErr:  false,
		},
		{
			name:     "empty path",
			filePath: "",
			wantErr:  true,
		},
		{
			name:     "protected system path windows",
			filePath: "C:\\Windows\\system32\\kernel32.dll",
			wantErr:  true,
		},
		{
			name:     "protected system path unix",
			filePath: "/usr/bin/ls",
			wantErr:  true,
		},
		{
			name:     "relative path",
			filePath: "../test.txt",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateFile(tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteService_SafeDelete(t *testing.T) {
	// 创建临时测试文件
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_delete.txt")

	// 写入测试内容
	content := []byte("test content for deletion")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	service := NewService()

	// 测试安全删除
	err := service.SafeDelete(testFile)
	if err != nil {
		t.Errorf("SafeDelete() error = %v", err)
	}

	// 验证文件是否被删除
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Errorf("File should be deleted but still exists")
	}
}

func TestDeleteService_BatchDelete(t *testing.T) {
	// 创建临时测试目录和文件
	tempDir := t.TempDir()
	testFiles := []string{
		filepath.Join(tempDir, "file1.txt"),
		filepath.Join(tempDir, "file2.txt"),
		filepath.Join(tempDir, "file3.txt"),
	}

	// 创建测试文件
	for _, file := range testFiles {
		if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	service := NewService()

	// 测试批量删除
	results := service.BatchDelete(testFiles)

	// 验证结果
	if len(results) != len(testFiles) {
		t.Errorf("Expected %d results, got %d", len(testFiles), len(results))
	}

	for i, result := range results {
		if result.Error != nil {
			t.Errorf("BatchDelete failed for file %s: %v", testFiles[i], result.Error)
		}
		if !result.Success {
			t.Errorf("BatchDelete reported failure for file %s", testFiles[i])
		}
	}
}


func TestDeleteService_Execute(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		setup   func(tempDir string) []string
		wantErr bool
	}{
		{
			name: "execute with valid files",
			args: []string{"file1.txt", "file2.txt"},
			setup: func(tempDir string) []string {
				files := []string{
					filepath.Join(tempDir, "file1.txt"),
					filepath.Join(tempDir, "file2.txt"),
				}
				for _, file := range files {
					os.WriteFile(file, []byte("test"), 0644)
				}
				return files
			},
			wantErr: false,
		},
		{
			name:    "execute with no args",
			args:    []string{},
			setup:   func(tempDir string) []string { return nil },
			wantErr: true,
		},
		{
			name: "execute with verbose flag",
			args: []string{"-v", "file1.txt"},
			setup: func(tempDir string) []string {
				file := filepath.Join(tempDir, "file1.txt")
				os.WriteFile(file, []byte("test"), 0644)
				return []string{file}
			},
			wantErr: false,
		},
		{
			name: "execute with only flags",
			args: []string{"-v", "--verbose"},
			setup: func(tempDir string) []string { return nil },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			var actualFiles []string
			if tt.setup != nil {
				actualFiles = tt.setup(tempDir)
			}

			// 替换参数中的文件名为实际路径
			var args []string
			for _, arg := range tt.args {
				if !strings.HasPrefix(arg, "-") && len(actualFiles) > 0 {
					for _, file := range actualFiles {
						if strings.HasSuffix(file, arg) {
							args = append(args, file)
							break
						}
					}
				} else {
					args = append(args, arg)
				}
			}

			service := NewService()
			ctx := context.WithTimeout(context.Background(), 5*time.Second)
			err := service.Execute(ctx, args)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteService_RecoverFromRecycleBin(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping recycle bin test on Windows - requires COM interface")
	}

	service := NewService()
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "recover_test.txt")

	// 测试恢复不存在的文件
	err := service.RecoverFromRecycleBin(testFile)
	if err == nil {
		t.Error("Expected error for non-existing file recovery")
	}
}

// 基准测试
func BenchmarkDeleteService_SafeDelete(b *testing.B) {
	service := NewService()
	tempDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testFile := filepath.Join(tempDir, fmt.Sprintf("bench_file_%d.txt", i))
		os.WriteFile(testFile, []byte("benchmark test"), 0644)
		service.SafeDelete(testFile)
	}
}

func BenchmarkDeleteService_BatchDelete(b *testing.B) {
	service := NewService()
	tempDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var files []string
		for j := 0; j < 10; j++ {
			file := filepath.Join(tempDir, fmt.Sprintf("batch_%d_%d.txt", i, j))
			os.WriteFile(file, []byte("test"), 0644)
			files = append(files, file)
		}
		service.BatchDelete(files)
	}
}

// 测试辅助函数
func TestDeleteService_NewService(t *testing.T) {
	tests := []struct {
		name   string
		config []interface{}
	}{
		{
			name:   "new service without config",
			config: nil,
		},
		{
			name:   "new service with config",
			config: []interface{}{"test-config"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var service *Service
			if tt.config == nil {
				service = NewService()
			} else {
				service = NewService(tt.config...)
			}

			if service == nil {
				t.Error("NewService() returned nil")
			}
		})
	}
}
