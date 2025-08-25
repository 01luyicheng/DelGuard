package main

import (
	"os"
	"path/filepath"
	"testing"
)

// 每个测试前清理全局缓存，避免路径与结果被缓存影响
func resetDefaultConfig() { defaultConfig = nil }

func TestConfig_Load_JSONC(t *testing.T) {
	resetDefaultConfig()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.jsonc")
	content := `{
	  // comment
	  "log_level": "debug",
	  "use_recycle_bin": false
	}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write jsonc failed: %v", err)
	}
	cfg, err := LoadConfigWithOverride(path)
	if err != nil { t.Fatalf("LoadConfigWithOverride jsonc failed: %v", err) }
	if cfg.LogLevel != "debug" { t.Fatalf("log_level mismatch: got %s", cfg.LogLevel) }
	if cfg.UseRecycleBin != false { t.Fatalf("use_recycle_bin mismatch: got %v", cfg.UseRecycleBin) }
}

func TestConfig_Load_INI_WithPrefixedKeys(t *testing.T) {
	resetDefaultConfig()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.ini")
	content := "" +
		"[general]\n" +
		"DELGUARD_LOG_LEVEL = error\n" +
		"DELGUARD_USE_RECYCLE_BIN = false\n" +
		"DELGUARD_MAX_FILE_SIZE = 1024\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write ini failed: %v", err)
	}
	cfg, err := LoadConfigWithOverride(path)
	if err != nil { t.Fatalf("LoadConfigWithOverride ini failed: %v", err) }
	if cfg.LogLevel != "error" { t.Fatalf("log_level mismatch: got %s", cfg.LogLevel) }
	if cfg.UseRecycleBin != false { t.Fatalf("use_recycle_bin mismatch: got %v", cfg.UseRecycleBin) }
	if cfg.MaxFileSize != 1024 { t.Fatalf("max_file_size mismatch: got %d", cfg.MaxFileSize) }
}

func TestConfig_Load_Properties_WithPrefixedKeys(t *testing.T) {
	resetDefaultConfig()
	dir := t.TempDir()
	path := filepath.Join(dir, "delguard.properties")
	content := "" +
		"DELGUARD_ENABLE_SECURITY_CHECKS=true\n" +
		"DELGUARD_INTERACTIVE_MODE=confirm\n" +
		"DELGUARD_SAFE_MODE=normal\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write properties failed: %v", err)
	}
	cfg, err := LoadConfigWithOverride(path)
	if err != nil { t.Fatalf("LoadConfigWithOverride properties failed: %v", err) }
	if cfg.EnableSecurityChecks != true { t.Fatalf("enable_security_checks mismatch: got %v", cfg.EnableSecurityChecks) }
	if cfg.InteractiveMode != "confirm" { t.Fatalf("interactive_mode mismatch: got %s", cfg.InteractiveMode) }
	if cfg.SafeMode != "normal" { t.Fatalf("safe_mode mismatch: got %s", cfg.SafeMode) }
}
