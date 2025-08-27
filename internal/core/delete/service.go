package delete

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DeleteResult 删除结果
type DeleteResult struct {
	Path    string
	Success bool
	Error   error
}

// Service 删除服务
type Service struct {
	config interface{} // 暂时使用interface{}，后续会替换为具体的配置类型
}

// NewService 创建删除服务 - 支持无参数调用
func NewService(config ...interface{}) *Service {
	var cfg interface{}
	if len(config) > 0 {
		cfg = config[0]
	}
	return &Service{
		config: cfg,
	}
}

// ValidateFile 验证文件路径
func (s *Service) ValidateFile(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("文件路径不能为空")
	}

	// 检查是否为受保护的系统路径
	protectedPaths := []string{
		"C:\\Windows",
		"C:\\Program Files",
		"C:\\Program Files (x86)",
		"/bin",
		"/sbin",
		"/usr/bin",
		"/usr/sbin",
		"/System",
		"/Applications",
	}

	for _, protected := range protectedPaths {
		if strings.HasPrefix(strings.ToLower(filePath), strings.ToLower(protected)) {
			return fmt.Errorf("不能删除受保护的系统路径: %s", filePath)
		}
	}

	return nil
}

// SafeDelete 安全删除文件
func (s *Service) SafeDelete(filePath string) error {
	// 验证文件路径
	if err := s.ValidateFile(filePath); err != nil {
		return err
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", filePath)
	}

	// 尝试移动到回收站
	if err := s.MoveToRecycleBin(filePath); err != nil {
		return fmt.Errorf("移动到回收站失败: %v", err)
	}

	return nil
}

// BatchDelete 批量删除文件
func (s *Service) BatchDelete(filePaths []string) []DeleteResult {
	results := make([]DeleteResult, len(filePaths))

	for i, filePath := range filePaths {
		err := s.SafeDelete(filePath)
		results[i] = DeleteResult{
			Path:    filePath,
			Success: err == nil,
			Error:   err,
		}
	}

	return results
}

// RecoverFromRecycleBin 从回收站恢复文件
func (s *Service) RecoverFromRecycleBin(filePath string) error {
	// 这是一个简化的实现，实际的回收站恢复需要更复杂的逻辑
	return fmt.Errorf("回收站恢复功能暂未实现")
}

// MoveToRecycleBin 移动文件到回收站
func (s *Service) MoveToRecycleBin(filePath string) error {
	// Windows 回收站实现
	if filepath.Separator == '\\' {
		return s.moveToWindowsRecycleBin(filePath)
	}

	// Unix/Linux 系统的回收站实现
	return s.moveToUnixTrash(filePath)
}

// Execute 执行删除操作
func (s *Service) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("请指定要删除的文件或目录")
	}

	// 过滤掉选项参数，只保留文件路径
	var filePaths []string
	verbose := false

	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			if arg == "-v" || arg == "--verbose" {
				verbose = true
			}
			// 忽略其他选项参数
			continue
		}
		filePaths = append(filePaths, arg)
	}

	if len(filePaths) == 0 {
		return fmt.Errorf("请指定要删除的文件或目录")
	}

	for _, target := range filePaths {
		if verbose {
			fmt.Printf("正在删除: %s\n", target)
		}

		if err := s.SafeDelete(target); err != nil {
			return fmt.Errorf("删除 %s 失败: %v", target, err)
		}

		if verbose {
			fmt.Printf("成功删除: %s\n", target)
		}
	}

	return nil
}
