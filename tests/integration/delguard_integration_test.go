package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	delguardBinary = "../../build/delguard.exe"
)

func TestMain(m *testing.M) {
	// 构建测试二进制文件
	buildCmd := exec.Command("go", "build", "-o", delguardBinary, "../../cmd/delguard")
	if err := buildCmd.Run(); err != nil {
		panic("Failed to build delguard binary: " + err.Error())
	}

	// 运行测试
	code := m.Run()

	// 清理
	os.Remove(delguardBinary)
	os.Exit(code)
}

func TestDelGuard_Help(t *testing.T) {
	cmd := exec.Command(delguardBinary, "--help")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Failed to run delguard --help: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "DelGuard") {
		t.Errorf("Help output should contain 'DelGuard', got: %s", outputStr)
	}
}

func TestDelGuard_Version(t *testing.T) {
	cmd := exec.Command(delguardBinary, "--version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Failed to run delguard --version: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "2.0") {
		t.Errorf("Version output should contain version info, got: %s", outputStr)
	}
}

func TestDelGuard_SafeDelete(t *testing.T) {
	// 创建临时测试文件
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "integration_test.txt")

	content := []byte("integration test content")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 执行安全删除
	cmd := exec.Command(delguardBinary, "delete", "--safe", testFile)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Logf("Command output: %s", string(output))
		t.Fatalf("Failed to run safe delete: %v", err)
	}

	// 验证文件是否被删除
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Errorf("File should be deleted but still exists")
	}
}

func TestDelGuard_Search(t *testing.T) {
	// 创建临时测试目录和文件
	tempDir := t.TempDir()

	testFiles := []string{
		"search_test1.txt",
		"search_test2.log",
		"other_file.pdf",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		if err := os.WriteFile(fullPath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// 搜索 .txt 文件
	cmd := exec.Command(delguardBinary, "search", "--pattern", "*.txt", tempDir)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Logf("Command output: %s", string(output))
		t.Fatalf("Failed to run search: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "search_test1.txt") {
		t.Errorf("Search output should contain 'search_test1.txt', got: %s", outputStr)
	}
}

func TestDelGuard_BatchDelete(t *testing.T) {
	// 创建临时测试文件
	tempDir := t.TempDir()

	testFiles := []string{
		filepath.Join(tempDir, "batch1.txt"),
		filepath.Join(tempDir, "batch2.txt"),
		filepath.Join(tempDir, "batch3.txt"),
	}

	for _, file := range testFiles {
		if err := os.WriteFile(file, []byte("batch test"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// 执行批量删除
	args := append([]string{"delete", "--batch"}, testFiles...)
	cmd := exec.Command(delguardBinary, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Logf("Command output: %s", string(output))
		t.Fatalf("Failed to run batch delete: %v", err)
	}

	// 验证所有文件都被删除
	for _, file := range testFiles {
		if _, err := os.Stat(file); !os.IsNotExist(err) {
			t.Errorf("File %s should be deleted but still exists", file)
		}
	}
}

func TestDelGuard_ConfigLoad(t *testing.T) {
	// 创建临时配置文件
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.json")

	configContent := `{
		"language": "en-us",
		"max_file_size": 1048576,
		"enable_recycle_bin": true,
		"log_level": "debug"
	}`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// 使用配置文件运行
	cmd := exec.Command(delguardBinary, "--config", configFile, "--help")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Logf("Command output: %s", string(output))
		t.Fatalf("Failed to run with config: %v", err)
	}
}

}
