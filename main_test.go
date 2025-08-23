package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// --- helpers ---

func withEnv(t *testing.T, key, val string, fn func()) {
	t.Helper()
	old, had := os.LookupEnv(key)
	if val == "" {
		_ = os.Unsetenv(key)
	} else {
		_ = os.Setenv(key, val)
	}
	defer func() {
		if had {
			_ = os.Setenv(key, old)
		} else {
			_ = os.Unsetenv(key)
		}
	}()
	fn()
}

func TestChooseExitCode_Priority(t *testing.T) {
	// 权限优先
	if got := ChooseExitCode(1, 1, 1, 0, 1, 1); got != 5 {
		t.Fatalf("perm>0 should be 5, got %d", got)
	}
	// 未找到其次（无权限）
	if got := ChooseExitCode(0, 1, 1, 0, 0, 0); got != 11 {
		t.Fatalf("notfound>0 should be 11, got %d", got)
	}
	// 受保护：仅当无成功且无其他错误时为 12
	if got := ChooseExitCode(0, 0, 1, 0, 0, 0); got != 12 {
		t.Fatalf("protected only should be 12, got %d", got)
	}
	if got := ChooseExitCode(0, 0, 1, 1, 0, 0); got != 0 {
		t.Fatalf("protected with success should be 0, got %d", got)
	}
	if got := ChooseExitCode(0, 0, 1, 0, 1, 0); got != 10 {
		t.Fatalf("protected with ioErr should be 10, got %d", got)
	}
	// 其他 I/O 或预处理错误
	if got := ChooseExitCode(0, 0, 0, 0, 1, 0); got != 10 {
		t.Fatalf("ioErr>0 should be 10, got %d", got)
	}
	if got := ChooseExitCode(0, 0, 0, 0, 0, 1); got != 10 {
		t.Fatalf("preErr>0 should be 10, got %d", got)
	}
	// 全部正常
	if got := ChooseExitCode(0, 0, 0, 1, 0, 0); got != 0 {
		t.Fatalf("all ok should be 0, got %d", got)
	}
}

func withTempConfig(t *testing.T, cfg any, fn func()) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	b, _ := json.Marshal(cfg)
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	SetConfigOverride(path)
	defer SetConfigOverride("")
	fn()
}

// --- tests ---

// --- 安全删除操作单元测试 ---

func TestIsCriticalPath_Advanced(t *testing.T) {
	// 测试关键路径检测
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"normal path", "/home/user/file.txt", false},
		{"relative path", "file.txt", false},
		{"current directory", ".", false},
		{"parent directory", "..", false},
	}

	// 平台特定测试
	if runtime.GOOS == "windows" {
		windowsTests := []struct {
			name string
			path string
			want bool
		}{
			{"windows system32", "C:\\Windows\\System32", true},
			{"windows program files", "C:\\Program Files", true},
			{"windows program files x86", "C:\\Program Files (x86)", true},
			{"windows programdata", "C:\\ProgramData", true},
			{"recycle bin path", "C:\\$RECYCLE.BIN", true},
			{"windows root", "C:\\", true},
			{"windows drive root", "D:\\", true},
		}
		tests = append(tests, windowsTests...)
	} else {
		unixTests := []struct {
			name string
			path string
			want bool
		}{
			{"bin directory", "/bin", true},
			{"usr directory", "/usr", true},
			{"etc directory", "/etc", true},
			{"dev directory", "/dev", true},
			{"root directory", "/", true},
			{"sbin directory", "/sbin", true},
			{"lib directory", "/lib", true},
			{"var directory", "/var", true},
			{"boot directory", "/boot", true},
		}
		tests = append(tests, unixTests...)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsCriticalPath(tt.path); got != tt.want {
				t.Errorf("IsCriticalPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"valid path", "/home/user/file.txt", true},
		{"empty path", "", false},
		{"path with null char", "/home/user/\x00file.txt", false},
	}

	if runtime.GOOS == "windows" {
		windowsTests := []struct {
			name string
			path string
			want bool
		}{
			{"path with <", "C:\\file<.txt", false},
			{"path with >", "C:\\file>.txt", false},
			{"path with :", "C:\\file:.txt", false},
			{"path with \"", "C:\\file\".txt", false},
			{"path with |", "C:\\file|.txt", false},
			{"path with ?", "C:\\file?.txt", false},
			{"path with *", "C:\\file*.txt", false},
			{"reserved name CON", "CON", false},
			{"reserved name PRN", "PRN", false},
			{"reserved name AUX", "AUX", false},
			{"reserved name NUL", "NUL", false},
			{"reserved name COM1", "COM1", false},
			{"reserved name LPT1", "LPT1", false},
		}
		tests = append(tests, windowsTests...)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validatePath(tt.path); got != tt.want {
				t.Errorf("validatePath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsTrashDirectory(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"normal path", "/home/user/file.txt", false},
	}

	if runtime.GOOS == "windows" {
		windowsTests := []struct {
			name string
			path string
			want bool
		}{
			{"recycle bin", "C:\\$Recycle.Bin", true},
			{"recycler", "C:\\Recycler", true},
			{"RECYCLER", "C:\\RECYCLER", true},
		}
		tests = append(tests, windowsTests...)
	} else if runtime.GOOS == "darwin" {
		home, _ := os.UserHomeDir()
		macosTests := []struct {
			name string
			path string
			want bool
		}{
			{"mac trash", filepath.Join(home, ".Trash"), true},
		}
		tests = append(tests, macosTests...)
	} else {
		home, _ := os.UserHomeDir()
		linuxTests := []struct {
			name string
			path string
			want bool
		}{
			{"linux trash", filepath.Join(home, ".local", "share", "Trash"), true},
			{"linux trash alt", filepath.Join(home, ".Trash"), true},
		}
		tests = append(tests, linuxTests...)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTrashDirectory(tt.path); got != tt.want {
				t.Errorf("IsTrashDirectory(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
