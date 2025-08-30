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
	// 检查空路径
	if strings.TrimSpace(path) == "" {
		return errors.NewError(errors.ErrTypeInvalidPath, "路径不能为空", nil)
	}

	// 检查路径长度
	if len(path) > 4096 {
		return errors.NewError(errors.ErrTypeInvalidPath, 
			"路径过长", nil)
	}

	// 检查路径是否包含空字符
	if strings.ContainsRune(path, 0) {
		return errors.NewError(errors.ErrTypeInvalidPath, 
			"路径包含非法字符", nil)
	}

	// 转换为绝对路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		return errors.NewInvalidPathError(path)
	}

	// 清理路径，防止路径遍历攻击
	cleanPath := filepath.Clean(absPath)
	
	// 检查清理后的路径是否有效
	if strings.Contains(cleanPath, "..") {
		return errors.NewError(errors.ErrTypeInvalidPath, 
			"路径包含目录遍历字符", nil)
	}
	
	// 检查是否为符号链接
	if info, err := os.Lstat(cleanPath); err == nil && info.Mode()&os.ModeSymlink != 0 {
		// 获取符号链接的目标路径
		if target, err := os.Readlink(cleanPath); err == nil {
			// 重新验证目标路径
			if !filepath.IsAbs(target) {
				target = filepath.Join(filepath.Dir(cleanPath), target)
			}
			return pv.ValidateDeletePath(target)
		}
	}

	// 检查路径是否存在
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		return errors.NewFileNotFoundError(cleanPath)
	}

	// 检查是否为受保护路径
	if pv.isProtectedPath(cleanPath) {
		return errors.NewError(errors.ErrTypePermissionDenied, 
			fmt.Sprintf("不能删除受保护的路径: %s", cleanPath), nil)
	}

	// 检查是否为系统关键路径
	if pv.isSystemPath(cleanPath) {
		return errors.NewError(errors.ErrTypePermissionDenied, 
			fmt.Sprintf("不能删除系统关键路径: %s", cleanPath), nil)
	}

	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(cleanPath))
	blockedExts := []string{
		".sys", ".dll", ".exe", ".msi", ".com", ".bat", ".cmd",
		".drv", ".vxd", ".386", ".cpl", ".scr", ".pif",
	}
	
	for _, blocked := range blockedExts {
		if ext == blocked {
			return errors.NewError(errors.ErrTypePermissionDenied, 
				fmt.Sprintf("不能删除系统文件类型: %s", ext), nil)
		}
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
			filepath.Join(homeDir, "Pictures"),
			filepath.Join(homeDir, "Videos"),
			filepath.Join(homeDir, "Music"),
			filepath.Join(homeDir, "OneDrive"),
			filepath.Join(homeDir, "AppData"),
			filepath.Join(homeDir, "Application Data"),
			filepath.Join(homeDir, "Local Settings"),
			filepath.Join(homeDir, "Favorites"),
			filepath.Join(homeDir, "Contacts"),
			filepath.Join(homeDir, "Searches"),
		)
	}
	
	// 添加常见的敏感目录
	if homeDir, err := os.UserHomeDir(); err == nil {
		paths = append(paths,
			filepath.Join(homeDir, ".ssh"),
			filepath.Join(homeDir, ".gnupg"),
			filepath.Join(homeDir, ".gitconfig"),
			filepath.Join(homeDir, ".bashrc"),
			filepath.Join(homeDir, ".zshrc"),
			filepath.Join(homeDir, ".vimrc"),
			filepath.Join(homeDir, ".profile"),
			filepath.Join(homeDir, ".bash_profile"),
			filepath.Join(homeDir, ".git-credentials"),
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
			"C:\\ProgramData",
			"C:\\System Volume Information",
			"C:\\Recovery",
			"C:\\Boot",
			"C:\\MSOCache",
			"C:\\PerfLogs",
			"C:\\Users\\Public",
			"C:\\Users\\Default",
			"C:\\Users\\All Users",
		}
	case '/': // Unix-like
		paths = []string{
			"/bin",
			"/sbin",
			"/usr/bin",
			"/usr/sbin",
			"/usr/local/bin",
			"/usr/local/sbin",
			"/etc",
			"/boot",
			"/sys",
			"/proc",
			"/dev",
			"/lib",
			"/lib64",
			"/usr/lib",
			"/usr/lib64",
			"/opt",
			"/var",
			"/tmp",
			"/root",
		}
	}
	
	return paths
}

// ValidateRestorePath 验证恢复路径是否安全
func (pv *PathValidator) ValidateRestorePath(targetPath string) error {
	// 检查空路径
	if targetPath == "" {
		return errors.NewError(errors.ErrTypeInvalidPath, "恢复路径不能为空", nil)
	}

	// 转换为绝对路径
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return errors.NewInvalidPathError(targetPath)
	}

	// 清理路径，防止路径遍历攻击
	absPath = filepath.Clean(absPath)
	
	// 检查路径遍历攻击
	if strings.Contains(absPath, "..") {
		return errors.NewError(errors.ErrTypeInvalidPath, 
			"路径包含非法字符", nil)
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
	defer file.Close()
	defer os.Remove(tempFile)
	return true
}