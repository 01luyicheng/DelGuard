package delete

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// DeleteResult 删除结果
type DeleteResult struct {
	Path    string `json:"path"`
	Success bool   `json:"success"`
	Error   error  `json:"error,omitempty"`
}

// Config 删除服务配置
type Config struct {
	MaxConcurrency int      `json:"max_concurrency"`
	ProtectedPaths []string `json:"protected_paths"`
	EnableLogging  bool     `json:"enable_logging"`
}

// Service 删除服务
type Service struct {
	config  *Config
	logger  *Logger
	metrics *Metrics
	mu      sync.RWMutex
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		MaxConcurrency: 10,
		ProtectedPaths: []string{
			"C:\\Windows",
			"C:\\Program Files",
			"C:\\Program Files (x86)",
			"/bin",
			"/sbin",
			"/usr/bin",
			"/usr/sbin",
			"/System",
			"/Applications",
		},
		EnableLogging: true,
	}
}

// NewService 创建删除服务
func NewService(configs ...*Config) *Service {
	var cfg *Config
	if len(configs) > 0 && configs[0] != nil {
		cfg = configs[0]
	} else {
		cfg = DefaultConfig()
	}
	
	// 创建日志记录器
	var logger *Logger
	if cfg.EnableLogging {
		logger = DefaultLogger
	} else {
		logger = NewLogger(LogLevelError, io.Discard)
	}
	
	return &Service{
		config:  cfg,
		logger:  logger,
		metrics: NewMetrics(),
	}
}

// NewServiceWithLogger 创建带有自定义日志记录器的删除服务
func NewServiceWithLogger(config *Config, logger *Logger) *Service {
	if config == nil {
		config = DefaultConfig()
	}
	if logger == nil {
		logger = DefaultLogger
	}
	
	return &Service{
		config:  config,
		logger:  logger,
		metrics: NewMetrics(),
	}
}

// GetMetrics 获取统计信息
func (s *Service) GetMetrics() *Metrics {
	return s.metrics.GetSnapshot()
}

// ResetMetrics 重置统计信息
func (s *Service) ResetMetrics() {
	s.metrics.Reset()
}

// ValidateFile 验证文件路径
func (s *Service) ValidateFile(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("文件路径不能为空")
	}

	// 规范化路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("无法解析文件路径: %v", err)
	}

	s.mu.RLock()
	protectedPaths := s.config.ProtectedPaths
	s.mu.RUnlock()

	// 检查是否为受保护的系统路径
	for _, protected := range protectedPaths {
		absProtected, err := filepath.Abs(protected)
		if err != nil {
			continue
		}
		
		if strings.HasPrefix(strings.ToLower(absPath), strings.ToLower(absProtected)) {
			return fmt.Errorf("不能删除受保护的系统路径: %s", filePath)
		}
	}

	return nil
}

// SafeDelete 安全删除文件
func (s *Service) SafeDelete(filePath string) error {
	startTime := time.Now()
	s.logger.Debug("开始删除文件: %s", filePath)
	
	var fileSize int64
	var success bool
	var err error
	
	defer func() {
		duration := time.Since(startTime)
		s.metrics.RecordOperation(success, duration, fileSize, err)
	}()
	
	// 验证文件路径
	if err = s.ValidateFile(filePath); err != nil {
		deleteErr := NewDeleteError("validate", filePath, err)
		s.logger.Error("文件路径验证失败: %v", deleteErr)
		err = deleteErr
		return deleteErr
	}

	// 检查文件是否存在并获取文件大小
	if stat, statErr := os.Stat(filePath); os.IsNotExist(statErr) {
		deleteErr := NewDeleteError("stat", filePath, fmt.Errorf("文件不存在: %s", filePath))
		s.logger.Warn("文件不存在: %s", filePath)
		err = deleteErr
		return deleteErr
	} else if statErr != nil {
		deleteErr := NewDeleteError("stat", filePath, statErr)
		s.logger.Error("获取文件信息失败: %v", deleteErr)
		err = deleteErr
		return deleteErr
	} else {
		fileSize = stat.Size()
	}

	// 尝试移动到回收站
	if err = s.MoveToRecycleBin(filePath); err != nil {
		deleteErr := NewDeleteError("move_to_recycle_bin", filePath, err)
		s.logger.Error("移动到回收站失败: %v", deleteErr)
		err = deleteErr
		return deleteErr
	}

	success = true
	s.logger.Info("文件删除成功: %s (大小: %d 字节)", filePath, fileSize)
	return nil
}

// BatchDelete 批量删除文件 - 支持并发处理
func (s *Service) BatchDelete(filePaths []string) []DeleteResult {
	if len(filePaths) == 0 {
		return []DeleteResult{}
	}

	s.mu.RLock()
	maxConcurrency := s.config.MaxConcurrency
	s.mu.RUnlock()

	results := make([]DeleteResult, len(filePaths))
	
	// 使用信号量控制并发数
	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	for i, filePath := range filePaths {
		wg.Add(1)
		go func(index int, path string) {
			defer wg.Done()
			
			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			err := s.SafeDelete(path)
			results[index] = DeleteResult{
				Path:    path,
				Success: err == nil,
				Error:   err,
			}
		}(i, filePath)
	}

	wg.Wait()
	return results
}

// BatchDeleteWithContext 支持上下文的批量删除
func (s *Service) BatchDeleteWithContext(ctx context.Context, filePaths []string) []DeleteResult {
	if len(filePaths) == 0 {
		return []DeleteResult{}
	}

	s.mu.RLock()
	maxConcurrency := s.config.MaxConcurrency
	s.mu.RUnlock()

	results := make([]DeleteResult, len(filePaths))
	
	// 使用信号量控制并发数
	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	for i, filePath := range filePaths {
		wg.Add(1)
		go func(index int, path string) {
			defer wg.Done()
			
			select {
			case <-ctx.Done():
				results[index] = DeleteResult{
					Path:    path,
					Success: false,
					Error:   ctx.Err(),
				}
				return
			case semaphore <- struct{}{}:
			}
			
			defer func() { <-semaphore }()

			err := s.SafeDelete(path)
			results[index] = DeleteResult{
				Path:    path,
				Success: err == nil,
				Error:   err,
			}
		}(i, filePath)
	}

	wg.Wait()
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

// ExecuteOptions 执行选项
type ExecuteOptions struct {
	Verbose    bool
	DryRun     bool
	Force      bool
	Recursive  bool
	BatchMode  bool
}

// parseArgs 解析命令行参数
func (s *Service) parseArgs(args []string) ([]string, *ExecuteOptions, error) {
	var filePaths []string
	options := &ExecuteOptions{}

	for i, arg := range args {
		switch {
		case arg == "-v" || arg == "--verbose":
			options.Verbose = true
		case arg == "-n" || arg == "--dry-run":
			options.DryRun = true
		case arg == "-f" || arg == "--force":
			options.Force = true
		case arg == "-r" || arg == "--recursive":
			options.Recursive = true
		case arg == "-b" || arg == "--batch":
			options.BatchMode = true
		case strings.HasPrefix(arg, "-"):
			return nil, nil, fmt.Errorf("未知选项: %s", arg)
		default:
			// 展开通配符
			matches, err := filepath.Glob(arg)
			if err != nil {
				return nil, nil, fmt.Errorf("无效的文件模式 %s: %v", arg, err)
			}
			if len(matches) == 0 {
				filePaths = append(filePaths, arg) // 保留原始路径，即使不匹配
			} else {
				filePaths = append(filePaths, matches...)
			}
		}
	}

	return filePaths, options, nil
}

// Execute 执行删除操作
func (s *Service) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("请指定要删除的文件或目录")
	}

	filePaths, options, err := s.parseArgs(args)
	if err != nil {
		return err
	}

	if len(filePaths) == 0 {
		return fmt.Errorf("请指定要删除的文件或目录")
	}

	s.mu.RLock()
	enableLogging := s.config.EnableLogging
	s.mu.RUnlock()

	// 干运行模式
	if options.DryRun {
		if options.Verbose || enableLogging {
			fmt.Println("干运行模式 - 以下文件将被删除:")
		}
		for _, target := range filePaths {
			if options.Verbose || enableLogging {
				fmt.Printf("  %s\n", target)
			}
		}
		return nil
	}

	// 批量模式
	if options.BatchMode && len(filePaths) > 1 {
		results := s.BatchDeleteWithContext(ctx, filePaths)
		var errors []string
		
		for _, result := range results {
			if result.Error != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", result.Path, result.Error))
			} else if options.Verbose || enableLogging {
				fmt.Printf("成功删除: %s\n", result.Path)
			}
		}
		
		if len(errors) > 0 {
			return fmt.Errorf("批量删除失败:\n%s", strings.Join(errors, "\n"))
		}
		return nil
	}

	// 逐个删除
	for _, target := range filePaths {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if options.Verbose || enableLogging {
			fmt.Printf("正在删除: %s\n", target)
		}

		if err := s.SafeDelete(target); err != nil {
			if !options.Force {
				return fmt.Errorf("删除 %s 失败: %v", target, err)
			}
			if options.Verbose || enableLogging {
				fmt.Printf("警告: 删除 %s 失败: %v\n", target, err)
			}
			continue
		}

		if options.Verbose || enableLogging {
			fmt.Printf("成功删除: %s\n", target)
		}
	}

	return nil
}
