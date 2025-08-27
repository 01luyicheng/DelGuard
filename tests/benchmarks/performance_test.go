package benchmarks

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"delguard/internal/core/delete"
	"delguard/internal/core/search"
)

func BenchmarkLargeFileDelete(b *testing.B) {
	service := delete.NewService()
	tempDir := b.TempDir()

	// 创建大文件 (10MB)
	largeContent := make([]byte, 10*1024*1024)
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		testFile := filepath.Join(tempDir, "large_file.bin")
		os.WriteFile(testFile, largeContent, 0644)

		start := time.Now()
		err := service.SafeDelete(testFile)
		duration := time.Since(start)

		if err != nil {
			b.Errorf("SafeDelete failed: %v", err)
		}

		b.ReportMetric(float64(duration.Nanoseconds()), "ns/delete")
		b.ReportMetric(float64(len(largeContent))/duration.Seconds(), "bytes/sec")
	}
}

func BenchmarkManySmallFilesDelete(b *testing.B) {
	service := delete.NewService()
	tempDir := b.TempDir()

	// 创建1000个小文件
	fileCount := 1000
	files := make([]string, fileCount)

	for i := 0; i < fileCount; i++ {
		filename := filepath.Join(tempDir, "small_file_"+string(rune(i))+".txt")
		files[i] = filename
		os.WriteFile(filename, []byte("small file content"), 0644)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 重新创建文件
		for _, file := range files {
			os.WriteFile(file, []byte("small file content"), 0644)
		}

		start := time.Now()
		results := service.BatchDelete(files)
		duration := time.Since(start)

		// 检查结果
		for j, result := range results {
			if result.Error != nil {
				b.Errorf("BatchDelete failed for file %d: %v", j, result.Error)
			}
		}

		b.ReportMetric(float64(duration.Nanoseconds()), "ns/batch")
		b.ReportMetric(float64(fileCount)/duration.Seconds(), "files/sec")
	}
}

func BenchmarkSearchLargeDirectory(b *testing.B) {
	service := search.NewService()
	tempDir := b.TempDir()

	// 创建大量文件和目录结构
	fileCount := 5000
	dirCount := 100

	// 创建目录结构
	for i := 0; i < dirCount; i++ {
		dirPath := filepath.Join(tempDir, "dir_"+string(rune(i)))
		os.MkdirAll(dirPath, 0755)

		// 在每个目录中创建文件
		for j := 0; j < fileCount/dirCount; j++ {
			filename := filepath.Join(dirPath, "file_"+string(rune(j))+".txt")
			os.WriteFile(filename, []byte("search test content"), 0644)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		start := time.Now()
		results, err := service.FindFiles(tempDir, "*.txt", true)
		duration := time.Since(start)

		if err != nil {
			b.Errorf("FindFiles failed: %v", err)
		}

		if len(results) < fileCount/2 {
			b.Errorf("Expected to find at least %d files, got %d", fileCount/2, len(results))
		}

		b.ReportMetric(float64(duration.Nanoseconds()), "ns/search")
		b.ReportMetric(float64(len(results))/duration.Seconds(), "files/sec")
	}
}

func BenchmarkDuplicateDetection(b *testing.B) {
	service := search.NewService()
	tempDir := b.TempDir()

	// 创建重复文件
	content1 := []byte("duplicate content type 1")
	content2 := []byte("duplicate content type 2")
	uniqueContent := []byte("unique content")

	// 创建文件结构
	for i := 0; i < 100; i++ {
		// 重复文件组1
		file1 := filepath.Join(tempDir, "dup1_"+string(rune(i))+".txt")
		os.WriteFile(file1, content1, 0644)

		// 重复文件组2
		file2 := filepath.Join(tempDir, "dup2_"+string(rune(i))+".txt")
		os.WriteFile(file2, content2, 0644)

		// 唯一文件
		uniqueFile := filepath.Join(tempDir, "unique_"+string(rune(i))+".txt")
		os.WriteFile(uniqueFile, append(uniqueContent, byte(i)), 0644)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		start := time.Now()
		duplicates, err := service.FindDuplicates(tempDir)
		duration := time.Since(start)

		if err != nil {
			b.Errorf("FindDuplicates failed: %v", err)
		}

		if len(duplicates) < 2 {
			b.Errorf("Expected to find at least 2 duplicate groups, got %d", len(duplicates))
		}

		b.ReportMetric(float64(duration.Nanoseconds()), "ns/duplicate_scan")
		b.ReportMetric(float64(300)/duration.Seconds(), "files/sec")
	}
}

func BenchmarkMemoryUsage(b *testing.B) {
	service := delete.NewService()
	tempDir := b.TempDir()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 创建大量文件
		files := make([]string, 1000)
		for j := 0; j < 1000; j++ {
			filename := filepath.Join(tempDir, "mem_test_"+string(rune(j))+".txt")
			files[j] = filename
			os.WriteFile(filename, []byte("memory test content"), 0644)
		}

		// 测试内存使用
		var m1, m2 runtime.MemStats
		runtime.ReadMemStats(&m1)

		results := service.BatchDelete(files)

		runtime.ReadMemStats(&m2)

		// 检查结果
		for _, result := range results {
			if result.Error != nil {
				b.Errorf("BatchDelete failed: %v", result.Error)
			}
		}

		memUsed := m2.Alloc - m1.Alloc
		b.ReportMetric(float64(memUsed), "bytes/operation")
		b.ReportMetric(float64(memUsed)/float64(len(files)), "bytes/file")
	}
}
