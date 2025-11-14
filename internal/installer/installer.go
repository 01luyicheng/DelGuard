package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// SystemInstaller 系统安装器接口
type SystemInstaller interface {
	// Install 安装DelGuard，替换系统命令
	Install() error

	// Uninstall 卸载DelGuard，恢复系统命令
	Uninstall() error

	// IsInstalled 检查是否已安装
	IsInstalled() bool

	// GetInstallPath 获取安装路径
	GetInstallPath() string

	// BackupOriginalCommands 备份原始命令
	BackupOriginalCommands() error

	// RestoreOriginalCommands 恢复原始命令
	RestoreOriginalCommands() error
}

// GetSystemInstaller 根据操作系统获取对应的安装器
func GetSystemInstaller() (SystemInstaller, error) {
	switch runtime.GOOS {
	case "windows":
		return NewWindowsInstaller(), nil
	case "darwin":
		return NewMacOSInstaller(), nil
	case "linux":
		return NewLinuxInstaller(), nil
	default:
		return nil, fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// InstallConfig 安装配置
type InstallConfig struct {
	InstallPath  string // 安装路径
	BackupPath   string // 备份路径
	CreateAlias  bool   // 是否创建别名
	SystemWide   bool   // 是否系统级安装
	ForceInstall bool   // 是否强制安装
}

// GetDefaultInstallConfig 获取默认安装配置
func GetDefaultInstallConfig() *InstallConfig {
	homeDir, _ := os.UserHomeDir()

	var installPath, backupPath string

	switch runtime.GOOS {
	case "windows":
		installPath = filepath.Join(homeDir, "AppData", "Local", "DelGuard")
		backupPath = filepath.Join(installPath, "backup")
	case "darwin":
		installPath = filepath.Join(homeDir, ".local", "bin")
		backupPath = filepath.Join(homeDir, ".local", "share", "delguard", "backup")
	case "linux":
		installPath = filepath.Join(homeDir, ".local", "bin")
		backupPath = filepath.Join(homeDir, ".local", "share", "delguard", "backup")
	default:
		installPath = filepath.Join(homeDir, ".delguard")
		backupPath = filepath.Join(installPath, "backup")
	}

	return &InstallConfig{
		InstallPath:  installPath,
		BackupPath:   backupPath,
		CreateAlias:  true,
		SystemWide:   false,
		ForceInstall: false,
	}
}

// CommandInfo 命令信息
type CommandInfo struct {
	Name         string // 命令名称
	OriginalPath string // 原始路径
	BackupPath   string // 备份路径
	AliasPath    string // 别名路径
}

// GetTargetCommands 获取需要替换的目标命令
func GetTargetCommands() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{"del", "rmdir", "erase"}
	case "darwin", "linux":
		return []string{"rm", "rmdir"}
	default:
		return []string{"rm"}
	}
}

// IsRunningAsAdmin 检查是否以管理员权限运行
func IsRunningAsAdmin() bool {
	switch runtime.GOOS {
	case "windows":
		return isWindowsAdmin()
	case "darwin", "linux":
		return os.Geteuid() == 0
	default:
		return false
	}
}

// RequiresAdmin 检查操作是否需要管理员权限
func RequiresAdmin(systemWide bool) bool {
	if systemWide {
		return true
	}

	// Windows的用户级安装通常不需要管理员权限
	// macOS和Linux的用户级安装也不需要
	return false
}
