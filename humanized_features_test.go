package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSmartArgumentParser 测试智能参数解析器
func TestSmartArgumentParser(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected *ArgumentResult
		wantErr  bool
	}{
		{
			name: "正常参数",
			args: []string{"file.txt", "-v"},
			expected: &ArgumentResult{
				Targets: []string{"file.txt"},
				Flags:   []string{"-v"},
			},
			wantErr: false,
		},
		{
			name: "分隔符测试",
			args: []string{"-v", "--", "-file.txt"},
			expected: &ArgumentResult{
				Targets: []string{"-file.txt"},
				Flags:   []string{"-v"},
			},
			wantErr: false,
		},
		{
			name:     "未知标志",
			args:     []string{"-unknown"},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewSmartArgumentParser(tt.args)
			result, err := parser.ParseArguments()

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseArguments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result != nil && tt.expected != nil {
				if len(result.Targets) != len(tt.expected.Targets) {
					t.Errorf("ParseArguments() targets = %v, want %v", result.Targets, tt.expected.Targets)
				}
				if len(result.Flags) != len(tt.expected.Flags) {
					t.Errorf("ParseArguments() flags = %v, want %v", result.Flags, tt.expected.Flags)
				}
			}
		})
	}
}

// TestSpecialFileHandler 测试特殊文件处理器
func TestSpecialFileHandler(t *testing.T) {
	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "delguard_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试文件
	testFiles := map[string][]byte{
		"normal.txt": []byte("normal file content"),
		".hidden":    []byte("hidden file content"),
		"very_long_name_" + strings.Repeat("x", 300) + ".txt": []byte("long name file"),
		"special\u0000char.txt":                               []byte("file with null byte"),
		"space  multiple.txt":                                 []byte("file with multiple spaces"),
		" leading_space.txt":                                  []byte("file with leading space"),
		"trailing_space.txt ":                                 []byte("file with trailing space"),
	}

	config := &Config{}
	handler := NewSpecialFileHandler(config)

	for filename, content := range testFiles {
		filePath := filepath.Join(tempDir, filename)

		// 某些非法文件名可能无法创建，跳过这些测试
		if err := ioutil.WriteFile(filePath, content, 0644); err != nil {
			t.Logf("跳过无法创建的文件: %s", filename)
			continue
		}

		t.Run(filename, func(t *testing.T) {
			issues, err := handler.AnalyzeFile(filePath)
			if err != nil {
				t.Logf("分析文件 %s 时出错: %v", filename, err)
				return
			}

			// 验证结果
			if filename == ".hidden" {
				found := false
				for _, issue := range issues {
					if issue.Type == "hidden_file" {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("隐藏文件 %s 未被检测到", filename)
				}
			}

			if strings.Contains(filename, strings.Repeat("x", 300)) {
				found := false
				for _, issue := range issues {
					if issue.Type == "long_filename" {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("长文件名 %s 未被检测到", filename)
				}
			}

			if strings.Contains(filename, "\u0000") {
				found := false
				for _, issue := range issues {
					if issue.Type == "unicode_issue" {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Unicode问题文件 %s 未被检测到", filename)
				}
			}

			if strings.Contains(filename, "  ") || strings.HasPrefix(filename, " ") || strings.HasSuffix(filename, " ") {
				found := false
				for _, issue := range issues {
					if issue.Type == "space_issue" {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("空格问题文件 %s 未被检测到", filename)
				}
			}
		})
	}
}

// TestTrashOperationMonitor 测试回收站操作监控器
func TestTrashOperationMonitor(t *testing.T) {
	config := &Config{}
	monitor := NewTrashOperationMonitor(config)

	tests := []struct {
		name     string
		path     string
		expected bool // 是否应该检测到回收站操作
	}{
		{
			name:     "Windows回收站",
			path:     "C:\\$Recycle.Bin\\S-1-5-21-123456789\\test.txt",
			expected: true,
		},
		{
			name:     "Linux回收站",
			path:     "/home/user/.local/share/Trash/files/test.txt",
			expected: true,
		},
		{
			name:     "macOS回收站",
			path:     "/Users/user/.Trash/test.txt",
			expected: true,
		},
		{
			name:     "普通文件",
			path:     "/home/user/Documents/test.txt",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operation, err := monitor.DetectTrashOperation(tt.path)
			if err != nil {
				t.Errorf("DetectTrashOperation() error = %v", err)
				return
			}

			if tt.expected && operation == nil {
				t.Errorf("期望检测到回收站操作，但没有检测到")
			}

			if !tt.expected && operation != nil {
				t.Errorf("不期望检测到回收站操作，但检测到了: %+v", operation)
			}
		})
	}
}

// TestProtectionMechanisms 测试保护机制
func TestProtectionMechanisms(t *testing.T) {
	// 创建临时目录模拟DelGuard项目
	tempDir, err := ioutil.TempDir("", "delguard_project")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// 创建DelGuard特征文件
	delguardFiles := []string{"main.go", "config.go", "protect.go"}
	for _, file := range delguardFiles {
		filePath := filepath.Join(tempDir, file)
		if err := ioutil.WriteFile(filePath, []byte("// DelGuard file"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// 测试DelGuard项目检测
	if !isDelGuardProject(tempDir) {
		t.Errorf("isDelGuardProject() 应该检测到DelGuard项目")
	}

	// 测试保护机制
	testFile := filepath.Join(tempDir, "main.go")
	err = checkCriticalProtection(testFile, false)
	if err == nil {
		t.Errorf("checkCriticalProtection() 应该阻止删除DelGuard项目文件")
	}

	// 测试强制模式（由于需要用户交互，这里只测试是否不会直接失败）
	// 在真实环境中，这会提示用户确认
}

// TestErrorFormatting 测试错误格式化
func TestErrorFormatting(t *testing.T) {
	// 测试DGError格式化
	dgErr := &DGError{
		Kind:   KindPermission,
		Op:     "delete",
		Path:   "/test/file.txt",
		Advice: "请以管理员身份运行",
	}

	formatted := FormatErrorForDisplay(dgErr)
	if !strings.Contains(formatted, "权限不足") {
		t.Errorf("格式化错误应该包含用户友好的消息")
	}

	if !strings.Contains(formatted, "/test/file.txt") {
		t.Errorf("格式化错误应该包含文件路径")
	}

	if !strings.Contains(formatted, "建议") {
		t.Errorf("格式化错误应该包含建议")
	}
}

// BenchmarkSmartArgumentParser 性能基准测试
func BenchmarkSmartArgumentParser(b *testing.B) {
	args := []string{"-v", "--recursive", "file1.txt", "file2.txt", "--force"}

	for i := 0; i < b.N; i++ {
		parser := NewSmartArgumentParser(args)
		_, _ = parser.ParseArguments()
	}
}

// BenchmarkSpecialFileAnalysis 特殊文件分析性能基准测试
func BenchmarkSpecialFileAnalysis(b *testing.B) {
	// 创建临时文件
	tempDir, err := ioutil.TempDir("", "delguard_bench")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	if err := ioutil.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		b.Fatal(err)
	}

	config := &Config{}
	handler := NewSpecialFileHandler(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = handler.AnalyzeFile(testFile)
	}
}
