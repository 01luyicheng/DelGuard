package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfig_LoadDefault(t *testing.T) {
	config := NewConfig()

	// 测试默认配置加载
	if config.Language != "zh-cn" {
		t.Errorf("Expected default language 'zh-cn', got '%s'", config.Language)
	}

	if config.MaxFileSize != 1073741824 {
		t.Errorf("Expected default max file size 1073741824, got %d", config.MaxFileSize)
	}

	if !config.EnableRecycleBin {
		t.Errorf("Expected EnableRecycleBin to be true by default")
	}
}

func TestConfig_LoadFromFile(t *testing.T) {
	// 创建临时配置文件
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.json")

	configContent := `{
		"language": "en-us",
		"max_file_size": 2147483648,
		"enable_recycle_bin": false,
		"log_level": "debug"
	}`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config := NewConfig()
	err := config.LoadFromFile(configFile)
	if err != nil {
		t.Errorf("LoadFromFile() error = %v", err)
		return
	}

	// 验证配置是否正确加载
	if config.Language != "en-us" {
		t.Errorf("Expected language 'en-us', got '%s'", config.Language)
	}

	if config.MaxFileSize != 2147483648 {
		t.Errorf("Expected max file size 2147483648, got %d", config.MaxFileSize)
	}

	if config.EnableRecycleBin {
		t.Errorf("Expected EnableRecycleBin to be false")
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Language:         "zh-cn",
				MaxFileSize:      1024 * 1024 * 1024,
				EnableRecycleBin: true,
				LogLevel:         "info",
			},
			wantErr: false,
		},
		{
			name: "invalid language",
			config: &Config{
				Language:         "",
				MaxFileSize:      1024 * 1024 * 1024,
				EnableRecycleBin: true,
				LogLevel:         "info",
			},
			wantErr: true,
		},
		{
			name: "invalid max file size",
			config: &Config{
				Language:         "zh-cn",
				MaxFileSize:      -1,
				EnableRecycleBin: true,
				LogLevel:         "info",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_SaveToFile(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "save_test.json")

	config := &Config{
		Language:         "en-us",
		MaxFileSize:      2048,
		EnableRecycleBin: false,
		LogLevel:         "debug",
	}

	err := config.SaveToFile(configFile)
	if err != nil {
		t.Errorf("SaveToFile() error = %v", err)
		return
	}

	// 验证文件是否存在
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Errorf("Config file was not created")
		return
	}

	// 重新加载并验证
	newConfig := NewConfig()
	err = newConfig.LoadFromFile(configFile)
	if err != nil {
		t.Errorf("Failed to reload saved config: %v", err)
		return
	}

	if newConfig.Language != config.Language {
		t.Errorf("Language mismatch after save/load")
	}

	if newConfig.MaxFileSize != config.MaxFileSize {
		t.Errorf("MaxFileSize mismatch after save/load")
	}
}
