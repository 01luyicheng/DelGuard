package delete

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/01luyicheng/DelGuard/internal/config"
)

// DeleteResult 删除结果
type DeleteResult struct {
	Success     bool      `json:"success"`
	Path        string    `json:"path"`
	TrashPath   string    `json:"trashPath,omitempty"`
	Error       string    `json:"error,omitempty"`
	Size        int64     `json:"size"`
	DeletedTime time.Time `json:"deletedTime"`
}

// Config 删除配置
type Config struct {
	TrashPath       string   `yaml:"trashPath" json:"trashPath"`
	SafeMode        bool     `yaml:"safeMode" json:"safeMode"`
	ConfirmDelete   bool     `yaml:"confirmDelete" json:"confirmDelete"`
	MaxFileSize     int64    `yaml:"maxFileSize" json:"maxFileSize"`
	ExcludePatterns []string `yaml:"excludePatterns" json:"excludePatterns"`
	MaxConcurrency  int      `yaml:"maxConcurrency" json:"maxConcurrency"`
	ProtectedPaths  []string `yaml:"protectedPaths" json:"protectedPaths"`
}

// DefaultConfig 默认配置
var DefaultConfig = &Config{
	TrashPath:     "",
	SafeMode:      true,
	ConfirmDelete: true,
	MaxFileSize:   1024 * 1024 * 1024, // 1GB
	ExcludePatterns: []string{
		"*.tmp",
		"*.log",
		".git/*",
	},
}

// Service 删除服务
type Service struct {
	config *config.Config
}

// NewService 创建删除服务
func NewService(cfg *config.Config) *Service {
	return &Service{
		config: cfg,
	}
}

// Execute 执行删除命令
func (s *Service) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("请指定要删除的文件或目录")
	}

	// 解析参数
	var targets []string
	force := false
	recursive := false

	for i, arg := range args {
		switch arg {
		case "-f", "--force":
			force = true
		case "-r", "-R", "--recursive":
			recursive = true
		case "-rf", "-fr":
			force = true
			recursive = true
		default:
			if !strings.HasPrefix(arg, "-") {
				targets = append(targets, arg)
			} else {
				return fmt.Errorf("未知参数: %s", arg)
			}
		}
		_ = i // 避免未使用变量警告
	}

	if len(targets) == 0 {
		return fmt.Errorf("请指定要删除的文件或目录")
	}

	// 执行删除操作
	for _, target := range targets {
		if err := s.deleteTarget(ctx, target, force, recursive); err != nil {
			return fmt.Errorf("删除 %s 失败: %v", target, err)
		}
	}

	return nil
}

// deleteTarget 删除目标文件或目录
func (s *Service) deleteTarget(ctx context.Context, target string, force, recursive bool) error {
	// 检查文件是否存在
	info, err := os.Stat(target)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("文件或目录不存在: %s", target)
		}
		return err
	}

	// 检查是否为目录
	if info.IsDir() && !recursive {
		return fmt.Errorf("是目录，请使用 -r 参数: %s", target)
	}

	// 安全检查
	if err := s.performSafetyChecks(target, force); err != nil {
		return err
	}

	// 执行删除
	if s.config.Delete.UseTrash {
		return s.moveToTrash(target)
	} else {
		if info.IsDir() {
			return os.RemoveAll(target)
		} else {
			return os.Remove(target)
		}
	}
}

// performSafetyChecks 执行安全检查
func (s *Service) performSafetyChecks(target string, force bool) error {
	// 获取绝对路径
	absPath, err := filepath.Abs(target)
	if err != nil {
		return err
	}

	// 检查受保护路径
	for _, protectedPath := range s.config.Security.ProtectedPaths {
		if strings.HasPrefix(absPath, protectedPath) {
			if !force {
				return fmt.Errorf("受保护的路径，请使用 -f 参数强制删除: %s", target)
			}
		}
	}

	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(target))
	for _, blockedExt := range s.config.Security.BlockedExtensions {
		if ext == blockedExt {
			if !force {
				return fmt.Errorf("危险文件类型，请使用 -f 参数强制删除: %s", target)
			}
		}
	}

	// 需要确认
	if s.config.Security.RequireConfirmation && !force {
		fmt.Printf("确定要删除 %s 吗？(y/N): ", target)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			return fmt.Errorf("用户取消删除操作")
		}
	}

	return nil
}

// moveToTrash 移动到回收站
func (s *Service) moveToTrash(target string) error {
	trashPath, err := s.getTrashPath()
	if err != nil {
		return err
	}

	// 确保回收站目录存在
	if err := os.MkdirAll(trashPath, 0755); err != nil {
		return err
	}

	// 生成唯一的目标文件名
	baseName := filepath.Base(target)
	destPath := filepath.Join(trashPath, baseName)

	// 如果文件已存在，添加时间戳
	if _, err := os.Stat(destPath); err == nil {
		ext := filepath.Ext(baseName)
		name := strings.TrimSuffix(baseName, ext)
		destPath = filepath.Join(trashPath, fmt.Sprintf("%s_%d%s", name,
			os.Getpid(), ext))
	}

	// 移动文件到回收站
	return os.Rename(target, destPath)
}

// getTrashPath 获取回收站路径
func (s *Service) getTrashPath() (string, error) {
	// 如果配置了自定义回收站路径
	if s.config.Delete.TrashPath != "" {
		return s.config.Delete.TrashPath, nil
	}

	// 根据操作系统获取默认回收站路径
	switch runtime.GOOS {
	case "windows":
		return s.getWindowsTrashPath()
	case "darwin":
		return s.getMacOSTrashPath()
	case "linux":
		return s.getLinuxTrashPath()
	default:
		return "", fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// getWindowsTrashPath 获取Windows回收站路径
func (s *Service) getWindowsTrashPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".delguard", "trash"), nil
}

// getMacOSTrashPath 获取macOS废纸篓路径
func (s *Service) getMacOSTrashPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".Trash"), nil
}

// getLinuxTrashPath 获取Linux回收站路径
func (s *Service) getLinuxTrashPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "Trash", "files"), nil
}
