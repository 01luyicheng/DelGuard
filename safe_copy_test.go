package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSafeCopy(t *testing.T) {
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
	
	// 测试相同文件复制（应跳过）
	err = SafeCopy(srcFile, dstFile, opts)
	if err != nil {
		t.Errorf("复制相同文件时出错: %v", err)
	}
	
	// 测试不同的文件覆盖
	newContent := "New content for testing"
	err = os.WriteFile(srcFile, []byte(newContent), 0644)
	if err != nil {
		t.Fatalf("无法更新源文件: %v", err)
	}
	
	// 在非交互模式下应不覆盖
	err = SafeCopy(srcFile, dstFile, opts)
	if err != nil {
		t.Errorf("安全复制不同文件时出错: %v", err)
	}
	
	// 验证文件未被覆盖（仍应是原始内容）
	dstContent, err = os.ReadFile(dstFile)
	if err != nil {
		t.Errorf("无法读取目标文件: %v", err)
	}
	
	if string(dstContent) != content {
		t.Errorf("在非交互模式下文件被错误覆盖，期望 %s，实际 %s", content, string(dstContent))
	}
	
	// 测试强制复制模式
	opts.Force = true
	err = SafeCopy(srcFile, dstFile, opts)
	if err != nil {
		t.Errorf("强制安全复制失败: %v", err)
	}
	
	// 验证文件是否被覆盖
	dstContent, err = os.ReadFile(dstFile)
	if err != nil {
		t.Errorf("无法读取目标文件: %v", err)
	}
	
	if string(dstContent) != newContent {
		t.Errorf("强制复制后文件内容不匹配，期望 %s，实际 %s", newContent, string(dstContent))
	}
}

func TestSafeCopyToDirectory(t *testing.T) {
	// 创建临时目录用于测试
	tempDir := t.TempDir()
	
	// 创建源文件
	srcFile := filepath.Join(tempDir, "source.txt")
	content := "Test content"
	err := os.WriteFile(srcFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("无法创建测试源文件: %v", err)
	}
	
	// 创建目标目录
	dstDir := filepath.Join(tempDir, "destination")
	err = os.Mkdir(dstDir, 0755)
	if err != nil {
		t.Fatalf("无法创建目标目录: %v", err)
	}
	
	// 测试复制到目录
	opts := SafeCopyOptions{}
	err = SafeCopy(srcFile, dstDir, opts)
	if err != nil {
		t.Errorf("复制到目录失败: %v", err)
	}
	
	// 验证文件是否正确复制到目录
	dstFile := filepath.Join(dstDir, "source.txt")
	dstContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Errorf("无法读取目标文件: %v", err)
	}
	
	if string(dstContent) != content {
		t.Errorf("复制到目录的文件内容不匹配，期望 %s，实际 %s", content, string(dstContent))
	}
}