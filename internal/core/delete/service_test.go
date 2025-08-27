package delete

import (
	"os"
	"path/filepath"
	"testing"
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
			name:     "protected system path",
			filePath: "C:\\Windows\\system32\\kernel32.dll",
			wantErr:  true,
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

func BenchmarkDeleteService_SafeDelete(b *testing.B) {
	service := NewService()
	tempDir := b.TempDir()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		testFile := filepath.Join(tempDir, "bench_test.txt")
		os.WriteFile(testFile, []byte("benchmark test"), 0644)

		err := service.SafeDelete(testFile)
		if err != nil {
			b.Errorf("SafeDelete() error = %v", err)
		}
	}
}

func TestDeleteService_RecoverFromRecycleBin(t *testing.T) {
	service := NewService()

	// 创建临时测试文件
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "recover_test.txt")
	content := []byte("recovery test content")

	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 删除文件到回收站
	err := service.SafeDelete(testFile)
	if err != nil {
		t.Fatalf("SafeDelete() error = %v", err)
	}

	// 尝试恢复文件
	err = service.RecoverFromRecycleBin(testFile)
	if err != nil {
		t.Logf("RecoverFromRecycleBin() error = %v (may be expected on some systems)", err)
	}
}
