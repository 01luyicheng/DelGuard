package main

import (
	"os"
	"path/filepath"
)

// EnableOverwriteProtection 启用文件覆盖保护
func EnableOverwriteProtection() error {
	// 在用户的主目录下创建一个标记文件，表示启用覆盖保护
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	
	protectFile := filepath.Join(homeDir, ".delguard", "overwrite_protection_enabled")
	
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(protectFile), 0755); err != nil {
		return err
	}
	
	// 创建或更新标记文件
	if err := os.WriteFile(protectFile, []byte("enabled"), 0644); err != nil {
		return err
	}
	
	return nil
}

// DisableOverwriteProtection 禁用文件覆盖保护
func DisableOverwriteProtection() error {
	// 删除标记文件以禁用覆盖保护
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	
	protectFile := filepath.Join(homeDir, ".delguard", "overwrite_protection_enabled")
	
	// 删除标记文件
	if err := os.Remove(protectFile); err != nil && !os.IsNotExist(err) {
		return err
	}
	
	return nil
}

// IsOverwriteProtectionEnabled 检查是否启用了文件覆盖保护
func IsOverwriteProtectionEnabled() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	
	protectFile := filepath.Join(homeDir, ".delguard", "overwrite_protection_enabled")
	
	// 检查标记文件是否存在
	_, err = os.Stat(protectFile)
	return err == nil
}