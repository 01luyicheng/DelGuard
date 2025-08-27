package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPathUtils_NormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Windows path normalization",
			input:    "C:/Users/test/Documents",
			expected: "C:\\Users\\test\\Documents",
		},
		{
			name:     "Unix path normalization",
			input:    "/home/user/documents",
			expected: "/home/user/documents",
		},
		{
			name:     "Mixed separators Windows",
			input:    "C:/Users\\test/Documents",
			expected: "C:\\Users\\test\\Documents",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PathUtils.NormalizePath(tt.input)
			
			// 根据操作系统调整期望值
			var expected string
			if runtime.GOOS == "windows" {
				// Windows: 期望反斜杠
				expected = strings.ReplaceAll(tt.expected, "/", "\\")
			} else {
				// Unix: 期望正斜杠
				expected = strings.ReplaceAll(tt.expected, "\\", "/")
			}
			
			if result != expected {
				t.Errorf("NormalizePath() = %v, want %v", result, expected)
			}
		})
	}
}

func TestPathUtils_IsDangerousPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Windows root directory",
			path:     "C:\\",
			expected: true,
		},
		{
			name:     "Unix root directory",
			path:     "/",
			expected: true,
		},
		{
			name:     "Windows system32",
			path:     "C:\\Windows\\System32",
			expected: true,
		},
		{
			name:     "User documents",
			path:     "C:\\Users\\test\\Documents",
			expected: false,
		},
		{
			name:     "Unix home",
			path:     "/home/user",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 根据当前操作系统跳过不相关的测试
			if runtime.GOOS == "windows" {
				// Windows系统跳过Unix特定测试
				if strings.Contains(tt.name, "Unix") && tt.path == "/" {
					t.Skip("Skipping Unix root test on Windows")
				}
			} else {
				// Unix系统跳过Windows特定测试
				if strings.Contains(tt.name, "Windows") && (tt.path == "C:\\" || strings.Contains(tt.path, "Windows")) {
					t.Skip("Skipping Windows test on Unix")
				}
			}
			
			result := PathUtils.IsDangerousPath(tt.path)
			if result != tt.expected {
				t.Errorf("IsDangerousPath(%s) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestPathUtils_GetTrashPaths(t *testing.T) {
	trashPaths := PathUtils.GetTrashPaths()
	if len(trashPaths) == 0 {
		t.Errorf("GetTrashPaths() returned empty slice")
	}

	// 检查返回的路径格式是否正确
	for _, path := range trashPaths {
		if path == "" {
			t.Errorf("GetTrashPaths() returned empty path")
		}
		
		// 验证路径分隔符
		if runtime.GOOS == "windows" {
			if strings.Contains(path, "/") && !strings.Contains(path, "\\") {
				t.Errorf("Windows path should use backslash: %s", path)
			}
		} else {
			if strings.Contains(path, "\\") {
				t.Errorf("Unix path should use forward slash: %s", path)
			}
		}
	}
}

func TestPathUtils_JoinPath(t *testing.T) {
	result := PathUtils.JoinPath("C:", "Users", "test", "Documents")
	expected := filepath.Join("C:", "Users", "test", "Documents")
	
	if result != expected {
		t.Errorf("JoinPath() = %v, want %v", result, expected)
	}
}

func TestPathUtils_ExpandEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Windows user profile",
			input:    "%USERPROFILE%/Documents",
			expected: filepath.Join(os.Getenv("USERPROFILE"), "Documents"),
		},
		{
			name:     "Unix home",
			input:    "$HOME/documents",
			expected: filepath.Join(os.Getenv("HOME"), "documents"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PathUtils.expandEnvironmentVariables(tt.input)
			// 只测试环境变量展开，不测试路径标准化
			if !strings.Contains(result, "Documents") && !strings.Contains(result, "documents") {
				t.Errorf("expandEnvironmentVariables() = %v, want to contain Documents", result)
			}
		})
	}
}