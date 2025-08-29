package security

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"delguard/internal/errors"
)

// PathValidator 路径验证器
type PathValidator struct {
	// 受保护的路径列表
	protectedPaths []string
	// 系统关键目录
	systemPaths []string
}

// NewPathValidator 创建路径验证器
func NewPathValidator() *PathValidator {
	return &PathValidator{
		protectedPaths: getProtectedPaths(),
		systemPaths:    getSystemPaths(),
	}
}

// ValidateDeletePath 验证删除路径是否安全
func (pv *PathValidator) ValidateDeletePath(path string) error {
	// 转换为绝对路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		return errors.NewInvalidPathError(path)
	}

	// 检查路径是否存在
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return errors.NewFileNotFoundError(absPath)
	}

	// 检查是否为受保护路径
	if pv.isProtectedPath(absPath) {
		return errors.NewError(errors.ErrTypePermissionDenied, 
			fmt.Sprintf("不能删除受保护的路径: %s", absPath), nil)
	}

	// 检查是否为系统关键路径
	if pv.isSystemPath(absPath) {
		return errors.NewError(errors.ErrTypePermissionDenied, 
			fmt.Sprintf("不能删除系统关键路径: %s", absPath), nil)
	}

	return nil
}

// isProtectedPath 检查是否为受保护路径
func (pv *PathValidator) isProtectedPath(path string) bool {
	for _, protected := range pv.protectedPaths {
		if strings.HasPrefix(path, protected) {
			return true
		}
	}
	return false
}

// isSystemPath 检查是否为系统路径
func (pv *PathValidator) isSystemPath(path string) bool {
	for _, system := range pv.systemPaths {
		if strings.HasPrefix(path, system) {
			return true
		}
	}
	return false
}

// getProtectedPaths 获取受保护路径列表
func getProtectedPaths() []string {
	var paths []string
	
	// 添加用户主目录的重要子目录
	if homeDir, err := os.UserHomeDir(); err == nil {
		paths = append(paths, 
			filepath.Join(homeDir, "Desktop"),
			filepath.Join(homeDir, "Documents"),
			filepath.Join(homeDir, "Downloads"),
		)
	}
	
	return paths
}

// getSystemPaths 获取系统关键路径列表
func getSystemPaths() []string {
	var paths []string
	
	switch filepath.Separator {
	case '\\': // Windows
		paths = []string{
			"C:\\Windows",
			"C:\\Program Files",
			"C:\\Program Files (x86)",
			"C:\\System Volume Information",
		}
	case '/': // Unix-like
		paths = []string{
			"/bin",
			"/sbin",
			"/usr/bin",
			"/usr/sbin",
			"/etc",
			"/boot",
			"/sys",
			"/proc",
		}
	}
	
	return paths
}

// ValidateRestorePath 验证恢复路径是否安全
func (pv *PathValidator) ValidateRestorePath(targetPath string) error {
	// 转换为绝对路径
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return errors.NewInvalidPathError(targetPath)
	}

	// 检查目标目录是否存在
	targetDir := filepath.Dir(absPath)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return errors.NewError(errors.ErrTypeInvalidPath, 
			fmt.Sprintf("目标目录不存在: %s", targetDir), nil)
	}

	// 检查是否有写权限
	if !pv.hasWritePermission(targetDir) {
		return errors.NewPermissionDeniedError(targetDir)
	}

	return nil
}

// hasWritePermission 检查是否有写权限
func (pv *PathValidator) hasWritePermission(path string) bool {
	// 尝试在目录中创建临时文件来测试写权限
	tempFile := filepath.Join(path, ".delguard_test")
	file, err := os.Create(tempFile)
	if err != nil {
		return false
	}
	file.Close()
	os.Remove(tempFile)
	return true
}