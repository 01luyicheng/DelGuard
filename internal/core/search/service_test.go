package search

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSearchService_FindFiles(t *testing.T) {
	// 创建临时测试目录结构
	tempDir := t.TempDir()

	// 创建测试文件
	testFiles := []string{
		"test1.txt",
		"test2.log",
		"document.pdf",
		"subdir/nested.txt",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)

		// 创建目录
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		// 创建文件
		if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", fullPath, err)
		}
	}

	service := NewService()

	tests := []struct {
		name        string
		pattern     string
		recursive   bool
		expectedMin int // 最少期望找到的文件数
	}{
		{
			name:        "find txt files",
			pattern:     "*.txt",
			recursive:   false,
			expectedMin: 1,
		},
		{
			name:        "find all files recursively",
			pattern:     "*",
			recursive:   true,
			expectedMin: 4,
		},
		{
			name:        "find log files",
			pattern:     "*.log",
			recursive:   false,
			expectedMin: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := service.FindFiles(tempDir, tt.pattern, tt.recursive)
			if err != nil {
				t.Errorf("FindFiles() error = %v", err)
				return
			}

			if len(results) < tt.expectedMin {
				t.Errorf("FindFiles() found %d files, expected at least %d", len(results), tt.expectedMin)
			}
		})
	}
}

func TestSearchService_FindBySize(t *testing.T) {
	tempDir := t.TempDir()
	service := NewService()

	// 创建不同大小的测试文件
	smallFile := filepath.Join(tempDir, "small.txt")
	largeFile := filepath.Join(tempDir, "large.txt")

	// 小文件 (100 bytes)
	if err := os.WriteFile(smallFile, make([]byte, 100), 0644); err != nil {
		t.Fatalf("Failed to create small file: %v", err)
	}

	// 大文件 (1000 bytes)
	if err := os.WriteFile(largeFile, make([]byte, 1000), 0644); err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	// 测试按大小查找
	results, err := service.FindBySize(tempDir, 500, true) // 查找大于500字节的文件
	if err != nil {
		t.Errorf("FindBySize() error = %v", err)
		return
	}

	// 应该只找到大文件
	found := false
	for _, result := range results {
		if filepath.Base(result.Path) == "large.txt" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("FindBySize() should find large.txt")
	}
}

func TestSearchService_FindDuplicates(t *testing.T) {
	tempDir := t.TempDir()
	service := NewService()

	// 创建重复内容的文件
	content := []byte("duplicate content for testing")

	file1 := filepath.Join(tempDir, "dup1.txt")
	file2 := filepath.Join(tempDir, "dup2.txt")
	file3 := filepath.Join(tempDir, "unique.txt")

	if err := os.WriteFile(file1, content, 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}

	if err := os.WriteFile(file2, content, 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	if err := os.WriteFile(file3, []byte("unique content"), 0644); err != nil {
		t.Fatalf("Failed to create file3: %v", err)
	}

	// 查找重复文件
	duplicates, err := service.FindDuplicates(tempDir)
	if err != nil {
		t.Errorf("FindDuplicates() error = %v", err)
		return
	}

	// 验证找到重复文件
	if len(duplicates) == 0 {
		t.Errorf("FindDuplicates() should find duplicate files")
	}
}

func BenchmarkSearchService_FindFiles(b *testing.B) {
	tempDir := b.TempDir()
	service := NewService()

	// 创建一些测试文件
	for i := 0; i < 100; i++ {
		filename := filepath.Join(tempDir, "bench_test_"+string(rune(i))+".txt")
		os.WriteFile(filename, []byte("benchmark test"), 0644)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := service.FindFiles(tempDir, "*.txt", false)
		if err != nil {
			b.Errorf("FindFiles() error = %v", err)
		}
	}
}
