package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// BoundaryConditionHandler 边界条件处理器
type BoundaryConditionHandler struct {
	outputManager *OutputManager
}

// NewBoundaryConditionHandler 创建边界条件处理器
func NewBoundaryConditionHandler(outputManager *OutputManager) *BoundaryConditionHandler {
	return &BoundaryConditionHandler{
		outputManager: outputManager,
	}
}

// ValidateInput 验证输入参数
func (bch *BoundaryConditionHandler) ValidateInput(input interface{}, inputType string) error {
	switch inputType {
	case "path":
		return bch.validatePath(input.(string))
	case "file_size":
		return bch.validateFileSize(input.(int64))
	case "timeout":
		return bch.validateTimeout(input.(time.Duration))
	case "count":
		return bch.validateCount(input.(int))
	default:
		return fmt.Errorf("未知的输入类型: %s", inputType)
	}
}

// validatePath 验证路径
func (bch *BoundaryConditionHandler) validatePath(path string) error {
	if path == "" {
		return fmt.Errorf("路径不能为空")
	}

	// 检查路径长度
	maxPathLength := 260 // Windows默认限制
	if runtime.GOOS != "windows" {
		maxPathLength = 4096 // Unix系统通常更大
	}

	if len(path) > maxPathLength {
		return fmt.Errorf("路径长度超过限制: %d > %d", len(path), maxPathLength)
	}

	// 检查路径中的非法字符
	if err := bch.checkIllegalCharacters(path); err != nil {
		return err
	}

	// 检查路径是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("路径不存在: %s", path)
	}

	return nil
}

// checkIllegalCharacters 检查非法字符
func (bch *BoundaryConditionHandler) checkIllegalCharacters(path string) error {
	// Windows非法字符
	if runtime.GOOS == "windows" {
		illegalChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
		for _, char := range illegalChars {
			if strings.Contains(path, char) {
				return fmt.Errorf("路径包含非法字符: %s", char)
			}
		}
	}

	// 检查控制字符
	for _, r := range path {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return fmt.Errorf("路径包含控制字符: %d", r)
		}
	}

	return nil
}

// validateFileSize 验证文件大小
func (bch *BoundaryConditionHandler) validateFileSize(size int64) error {
	if size < 0 {
		return fmt.Errorf("文件大小不能为负数: %d", size)
	}

	// 检查是否超过合理限制 (100GB)
	maxSize := int64(100 * 1024 * 1024 * 1024)
	if size > maxSize {
		return fmt.Errorf("文件大小超过限制: %d > %d", size, maxSize)
	}

	return nil
}

// validateTimeout 验证超时时间
func (bch *BoundaryConditionHandler) validateTimeout(timeout time.Duration) error {
	if timeout < 0 {
		return fmt.Errorf("超时时间不能为负数: %v", timeout)
	}

	// 检查是否超过合理限制 (24小时)
	maxTimeout := 24 * time.Hour
	if timeout > maxTimeout {
		return fmt.Errorf("超时时间过长: %v > %v", timeout, maxTimeout)
	}

	return nil
}

// validateCount 验证计数
func (bch *BoundaryConditionHandler) validateCount(count int) error {
	if count < 0 {
		return fmt.Errorf("计数不能为负数: %d", count)
	}

	// 检查是否超过合理限制
	maxCount := 1000000 // 100万
	if count > maxCount {
		return fmt.Errorf("计数超过限制: %d > %d", count, maxCount)
	}

	return nil
}

// CheckDiskSpace 检查磁盘空间
func (bch *BoundaryConditionHandler) CheckDiskSpace(path string, requiredBytes int64) error {
	if path == "" {
		return fmt.Errorf("路径不能为空")
	}

	if requiredBytes < 0 {
		return fmt.Errorf("所需空间不能为负数: %d", requiredBytes)
	}

	// 获取磁盘使用情况
	usage, err := getDiskUsage(path)
	if err != nil {
		return fmt.Errorf("获取磁盘使用情况失败: %v", err)
	}

	if int64(usage.Free) < requiredBytes {
		return fmt.Errorf("磁盘空间不足: 需要 %d 字节，可用 %d 字节",
			requiredBytes, int64(usage.Free))
	}

	// 检查是否接近磁盘满
	usagePercent := float64(usage.Used) / float64(usage.Total) * 100
	if usagePercent > 95 {
		bch.outputManager.Warn("磁盘使用率过高: %.1f%%", usagePercent)
	}

	return nil
}

// CheckMemoryUsage 检查内存使用情况
func (bch *BoundaryConditionHandler) CheckMemoryUsage() error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 检查内存使用是否过高
	if m.Sys > 1024*1024*1024 { // 1GB
		bch.outputManager.Warn("系统内存使用过高: %d MB", m.Sys/1024/1024)
	}

	if m.Alloc > 512*1024*1024 { // 512MB
		bch.outputManager.Warn("分配内存过高: %d MB", m.Alloc/1024/1024)
	}

	return nil
}

// CheckBoundaryFilePermissions 检查文件权限（重命名避免冲突）
func (bch *BoundaryConditionHandler) CheckBoundaryFilePermissions(path string, operation string) error {
	if path == "" {
		return fmt.Errorf("路径不能为空")
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %v", err)
	}

	mode := info.Mode()

	switch operation {
	case "read":
		// 检查读权限
		if mode&0400 == 0 && runtime.GOOS != "windows" {
			return fmt.Errorf("没有读权限: %s", path)
		}
	case "write":
		// 检查写权限
		if mode&0200 == 0 && runtime.GOOS != "windows" {
			return fmt.Errorf("没有写权限: %s", path)
		}
	case "execute":
		// 检查执行权限
		if mode&0100 == 0 && runtime.GOOS != "windows" {
			return fmt.Errorf("没有执行权限: %s", path)
		}
	}

	return nil
}

// CheckSystemLimits 检查系统限制
func (bch *BoundaryConditionHandler) CheckSystemLimits() error {
	// 检查文件描述符限制
	if runtime.GOOS != "windows" {
		// Unix系统检查
		// 这里可以添加更多系统限制检查
	}

	// 检查goroutine数量
	numGoroutines := runtime.NumGoroutine()
	if numGoroutines > 1000 {
		bch.outputManager.Warn("Goroutine数量过多: %d", numGoroutines)
	}

	return nil
}

// ValidateConfiguration 验证配置
func (bch *BoundaryConditionHandler) ValidateConfiguration(config *Config) error {
	if config == nil {
		return fmt.Errorf("配置不能为空")
	}

	// 验证最大文件大小
	if config.MaxFileSize < 0 {
		return fmt.Errorf("最大文件大小不能为负数: %d", config.MaxFileSize)
	}

	// 验证语言设置
	if config.Language == "" {
		return fmt.Errorf("语言设置不能为空")
	}

	validLanguages := []string{"en", "zh-cn", "zh-tw", "ja", "ko"}
	isValidLang := false
	for _, lang := range validLanguages {
		if config.Language == lang {
			isValidLang = true
			break
		}
	}
	if !isValidLang {
		return fmt.Errorf("不支持的语言: %s", config.Language)
	}

	return nil
}

// HandleEdgeCases 处理边界情况
func (bch *BoundaryConditionHandler) HandleEdgeCases(operation string, params map[string]interface{}) error {
	switch operation {
	case "delete_empty_directory":
		return bch.handleEmptyDirectory(params["path"].(string))
	case "delete_large_file":
		return bch.handleLargeFile(params["path"].(string), params["size"].(int64))
	case "delete_many_files":
		return bch.handleManyFiles(params["count"].(int))
	case "delete_special_characters":
		return bch.handleSpecialCharacters(params["path"].(string))
	default:
		return fmt.Errorf("未知的边界情况: %s", operation)
	}
}

// handleEmptyDirectory 处理空目录
func (bch *BoundaryConditionHandler) handleEmptyDirectory(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("读取目录失败: %v", err)
	}

	if len(entries) == 0 {
		bch.outputManager.Info("检测到空目录: %s", path)
		return nil
	}

	return fmt.Errorf("目录不为空: %s", path)
}

// handleLargeFile 处理大文件
func (bch *BoundaryConditionHandler) handleLargeFile(path string, size int64) error {
	// 大于1GB的文件需要特殊处理
	if size > 1024*1024*1024 {
		bch.outputManager.Warn("检测到大文件: %s (%.2f GB)", path, float64(size)/(1024*1024*1024))

		// 检查磁盘空间
		if err := bch.CheckDiskSpace(filepath.Dir(path), size); err != nil {
			return err
		}
	}

	return nil
}

// handleManyFiles 处理大量文件
func (bch *BoundaryConditionHandler) handleManyFiles(count int) error {
	if count > 10000 {
		bch.outputManager.Warn("检测到大量文件操作: %d 个文件", count)

		// 检查系统资源
		if err := bch.CheckMemoryUsage(); err != nil {
			return err
		}

		if err := bch.CheckSystemLimits(); err != nil {
			return err
		}
	}

	return nil
}

// handleSpecialCharacters 处理特殊字符
func (bch *BoundaryConditionHandler) handleSpecialCharacters(path string) error {
	// 检查Unicode字符
	for _, r := range path {
		if r > 127 {
			bch.outputManager.Info("检测到Unicode字符: %s", path)
			break
		}
	}

	// 检查空格
	if strings.HasPrefix(path, " ") || strings.HasSuffix(path, " ") {
		bch.outputManager.Warn("路径包含前导或尾随空格: %s", path)
	}

	return nil
}

// 全局边界条件处理器
var globalBoundaryHandler = NewBoundaryConditionHandler(globalOutputManager)

// 全局函数
func ValidateInput(input interface{}, inputType string) error {
	return globalBoundaryHandler.ValidateInput(input, inputType)
}

func CheckDiskSpace(path string, requiredBytes int64) error {
	return globalBoundaryHandler.CheckDiskSpace(path, requiredBytes)
}

func CheckBoundaryFilePermissions(path string, operation string) error {
	return globalBoundaryHandler.CheckBoundaryFilePermissions(path, operation)
}

func ValidateConfiguration(config *Config) error {
	return globalBoundaryHandler.ValidateConfiguration(config)
}
