package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSafeCopySimple(t *testing.T) {
	// 创建临时目录用于测试
	tempDir := t.TempDir()
	
	// 创建测试源文件
	srcFile := filepath.Join(tempDir, "source.txt")
	content := "Hello, DelGuard!"
	err := os.WriteFile(srcFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("无法创建测试源文件: %v", err)
	}
	
	// 测试基本复制功能
	dstFile := filepath.Join(tempDir, "destination.txt")
	opts := SafeCopyOptions{}
	
	err = SafeCopy(srcFile, dstFile, opts)
	if err != nil {
		t.Errorf("安全复制失败: %v", err)
	}
	
	// 验证文件是否正确复制
	dstContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Errorf("无法读取目标文件: %v", err)
	}
	
	if string(dstContent) != content {
		t.Errorf("复制的文件内容不匹配，期望 %s，实际 %s", content, string(dstContent))
	}
}