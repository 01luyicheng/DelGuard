package filesystem

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGetTrashManager(t *testing.T) {
	manager, err := GetTrashManager()
	if err != nil {
		t.Fatalf("获取回收站管理器失败: %v", err)
	}

	if manager == nil {
		t.Fatal("回收站管理器为空")
	}

	// 测试获取回收站路径
	trashPath, err := manager.GetTrashPath()
	if err != nil {
		t.Fatalf("获取回收站路径失败: %v", err)
	}

	if trashPath == "" {
		t.Fatal("回收站路径为空")
	}

	t.Logf("操作系统: %s, 回收站路径: %s", runtime.GOOS, trashPath)
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		size     int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, test := range tests {
		result := FormatFileSize(test.size)
		if result != test.expected {
			t.Errorf("FormatFileSize(%d) = %s, 期望 %s", test.size, result, test.expected)
		}
	}
}

func TestIsValidPath(t *testing.T) {
	// 测试空路径
	if IsValidPath("") {
		t.Error("空路径应该返回false")
	}

	// 测试不存在的路径
	if IsValidPath("/nonexistent/path/file.txt") {
		t.Error("不存在的路径应该返回false")
	}

	// 测试当前目录
	if !IsValidPath(".") {
		t.Error("当前目录应该返回true")
	}
}

func TestGetAbsolutePath(t *testing.T) {
	// 测试相对路径
	absPath, err := GetAbsolutePath(".")
	if err != nil {
		t.Fatalf("获取绝对路径失败: %v", err)
	}

	if !filepath.IsAbs(absPath) {
		t.Error("返回的路径不是绝对路径")
	}
}

func TestCreateDirIfNotExists(t *testing.T) {
	// 创建临时目录进行测试
	tempDir := filepath.Join(os.TempDir(), "delguard_test")
	defer os.RemoveAll(tempDir)

	// 测试创建不存在的目录
	err := CreateDirIfNotExists(tempDir)
	if err != nil {
		t.Fatalf("创建目录失败: %v", err)
	}

	// 检查目录是否存在
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Error("目录创建失败")
	}

	// 测试已存在的目录
	err = CreateDirIfNotExists(tempDir)
	if err != nil {
		t.Fatalf("处理已存在目录失败: %v", err)
	}
}

// 集成测试 - 测试完整的删除和恢复流程
func TestTrashIntegration(t *testing.T) {
	manager, err := GetTrashManager()
	if err != nil {
		t.Fatalf("获取回收站管理器失败: %v", err)
	}

	// 创建测试文件
	tempDir := filepath.Join(os.TempDir(), "delguard_integration_test")
	err = CreateDirIfNotExists(tempDir)
	if err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test_file.txt")
	testContent := "这是一个测试文件"
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 测试移动到回收站
	err = manager.MoveToTrash(testFile)
	if err != nil {
		t.Fatalf("移动到回收站失败: %v", err)
	}

	// 检查文件是否已从原位置删除
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("文件应该已从原位置删除")
	}

	// 测试列出回收站文件
	trashFiles, err := manager.ListTrashFiles()
	if err != nil {
		t.Fatalf("列出回收站文件失败: %v", err)
	}

	t.Logf("回收站中有 %d 个文件", len(trashFiles))

	// 查找我们的测试文件
	var testTrashFile *TrashFile
	for _, file := range trashFiles {
		if file.Name == "test_file.txt" {
			testTrashFile = &file
			break
		}
	}

	if testTrashFile == nil {
		t.Fatal("在回收站中未找到测试文件")
	}

	t.Logf("找到测试文件: %s, 大小: %s", testTrashFile.Name, FormatFileSize(testTrashFile.Size))

	// 测试恢复文件
	restoreFile := filepath.Join(tempDir, "restored_test_file.txt")
	err = manager.RestoreFile(*testTrashFile, restoreFile)
	if err != nil {
		t.Fatalf("恢复文件失败: %v", err)
	}

	// 检查文件是否已恢复
	if !IsValidPath(restoreFile) {
		t.Error("文件恢复失败")
	}

	// 检查文件内容
	restoredContent, err := os.ReadFile(restoreFile)
	if err != nil {
		t.Fatalf("读取恢复文件失败: %v", err)
	}

	if string(restoredContent) != testContent {
		t.Error("恢复文件内容不匹配")
	}

	t.Log("集成测试通过")
}
