package main

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

// ErrorSeverity 错误严重程度
type ErrorSeverity int

const (
	SeverityLow ErrorSeverity = iota
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

// ErrorCategory 错误分类
type ErrorCategory int

const (
	CategoryFileSystem ErrorCategory = iota
	CategoryPermission
	CategoryNetwork
	CategoryConfiguration
	CategoryValidation
	CategorySecurity
	CategorySystem
)

// EnhancedError 增强的错误类型
type EnhancedError struct {
	Code        string
	Message     string
	Cause       error
	Severity    ErrorSeverity
	Category    ErrorCategory
	Context     map[string]interface{}
	Timestamp   time.Time
	StackTrace  []string
	Suggestions []string
	Recoverable bool
}

// Error 实现error接口
func (e *EnhancedError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 支持错误链
func (e *EnhancedError) Unwrap() error {
	return e.Cause
}

// EnhancedErrorHandler 增强的错误处理器
type EnhancedErrorHandler struct {
	mu              sync.RWMutex
	errorCounts     map[string]int
	recentErrors    []*EnhancedError
	maxRecentErrors int
	outputManager   *OutputManager
	panicRecovery   bool
}

// NewEnhancedErrorHandler 创建新的增强错误处理器
func NewEnhancedErrorHandler(outputManager *OutputManager) *EnhancedErrorHandler {
	return &EnhancedErrorHandler{
		errorCounts:     make(map[string]int),
		recentErrors:    make([]*EnhancedError, 0),
		maxRecentErrors: 100,
		outputManager:   outputManager,
		panicRecovery:   true,
	}
}

// HandleEnhancedError 处理错误
func (eh *EnhancedErrorHandler) HandleEnhancedError(err error, operation string, critical bool) *EnhancedError {
	if err == nil {
		return nil
	}

	// 转换为增强错误
	enhancedErr := eh.enhanceError(err, operation)

	// 记录错误
	eh.recordError(enhancedErr)

	// 输出错误信息
	eh.displayError(enhancedErr, critical)

	// 如果是关键错误且启用了panic恢复，尝试恢复
	if critical && eh.panicRecovery {
		eh.attemptRecovery(enhancedErr)
	}

	return enhancedErr
}

// enhanceError 将普通错误转换为增强错误
func (eh *EnhancedErrorHandler) enhanceError(err error, operation string) *EnhancedError {
	// 如果已经是增强错误，直接返回
	if enhanced, ok := err.(*EnhancedError); ok {
		return enhanced
	}

	// 创建新的增强错误
	enhanced := &EnhancedError{
		Code:        eh.generateErrorCode(err, operation),
		Message:     err.Error(),
		Cause:       err,
		Severity:    eh.determineSeverity(err, operation),
		Category:    eh.determineCategory(err, operation),
		Context:     eh.gatherContext(operation),
		Timestamp:   time.Now(),
		StackTrace:  eh.captureStackTrace(),
		Suggestions: eh.generateSuggestions(err, operation),
		Recoverable: eh.isRecoverable(err, operation),
	}

	return enhanced
}

// generateErrorCode 生成错误代码
func (eh *EnhancedErrorHandler) generateErrorCode(err error, operation string) string {
	errStr := strings.ToLower(err.Error())

	// 根据错误内容生成代码
	switch {
	case strings.Contains(errStr, "permission denied"):
		return "PERM_001"
	case strings.Contains(errStr, "file not found"):
		return "FILE_001"
	case strings.Contains(errStr, "directory not found"):
		return "DIR_001"
	case strings.Contains(errStr, "access denied"):
		return "ACCESS_001"
	case strings.Contains(errStr, "disk full"):
		return "DISK_001"
	case strings.Contains(errStr, "network"):
		return "NET_001"
	case strings.Contains(errStr, "timeout"):
		return "TIME_001"
	case strings.Contains(errStr, "invalid"):
		return "VALID_001"
	case strings.Contains(errStr, "security"):
		return "SEC_001"
	default:
		return "GEN_001"
	}
}

// determineSeverity 确定错误严重程度
func (eh *EnhancedErrorHandler) determineSeverity(err error, operation string) ErrorSeverity {
	errStr := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errStr, "critical") || strings.Contains(errStr, "fatal"):
		return SeverityCritical
	case strings.Contains(errStr, "permission denied") || strings.Contains(errStr, "access denied"):
		return SeverityHigh
	case strings.Contains(errStr, "not found") || strings.Contains(errStr, "invalid"):
		return SeverityMedium
	default:
		return SeverityLow
	}
}

// determineCategory 确定错误分类
func (eh *EnhancedErrorHandler) determineCategory(err error, operation string) ErrorCategory {
	errStr := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errStr, "permission") || strings.Contains(errStr, "access"):
		return CategoryPermission
	case strings.Contains(errStr, "file") || strings.Contains(errStr, "directory"):
		return CategoryFileSystem
	case strings.Contains(errStr, "network") || strings.Contains(errStr, "connection"):
		return CategoryNetwork
	case strings.Contains(errStr, "config"):
		return CategoryConfiguration
	case strings.Contains(errStr, "invalid") || strings.Contains(errStr, "validation"):
		return CategoryValidation
	case strings.Contains(errStr, "security") || strings.Contains(errStr, "attack"):
		return CategorySecurity
	default:
		return CategorySystem
	}
}

// gatherContext 收集上下文信息
func (eh *EnhancedErrorHandler) gatherContext(operation string) map[string]interface{} {
	context := make(map[string]interface{})

	context["operation"] = operation
	context["timestamp"] = time.Now().Format(time.RFC3339)
	context["goroutines"] = runtime.NumGoroutine()

	// 获取内存信息
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	context["memory_alloc"] = m.Alloc
	context["memory_sys"] = m.Sys

	return context
}

// captureStackTrace 捕获堆栈跟踪
func (eh *EnhancedErrorHandler) captureStackTrace() []string {
	var stack []string

	// 获取调用栈
	for i := 2; i < 10; i++ { // 跳过当前函数和调用者
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		if fn != nil {
			stack = append(stack, fmt.Sprintf("%s:%d %s", file, line, fn.Name()))
		}
	}

	return stack
}

// generateSuggestions 生成解决建议
func (eh *EnhancedErrorHandler) generateSuggestions(err error, operation string) []string {
	errStr := strings.ToLower(err.Error())
	var suggestions []string

	switch {
	case strings.Contains(errStr, "permission denied"):
		suggestions = append(suggestions, "检查文件权限")
		suggestions = append(suggestions, "尝试以管理员身份运行")
		suggestions = append(suggestions, "确认用户有足够的权限")

	case strings.Contains(errStr, "file not found"):
		suggestions = append(suggestions, "检查文件路径是否正确")
		suggestions = append(suggestions, "确认文件是否存在")
		suggestions = append(suggestions, "检查文件名拼写")

	case strings.Contains(errStr, "disk full"):
		suggestions = append(suggestions, "清理磁盘空间")
		suggestions = append(suggestions, "删除临时文件")
		suggestions = append(suggestions, "移动文件到其他磁盘")

	case strings.Contains(errStr, "network"):
		suggestions = append(suggestions, "检查网络连接")
		suggestions = append(suggestions, "确认防火墙设置")
		suggestions = append(suggestions, "重试操作")

	default:
		suggestions = append(suggestions, "重试操作")
		suggestions = append(suggestions, "检查系统日志")
		suggestions = append(suggestions, "联系技术支持")
	}

	return suggestions
}

// isRecoverable 判断错误是否可恢复
func (eh *EnhancedErrorHandler) isRecoverable(err error, operation string) bool {
	errStr := strings.ToLower(err.Error())

	// 不可恢复的错误
	unrecoverablePatterns := []string{
		"fatal",
		"critical",
		"corrupted",
		"invalid signature",
	}

	for _, pattern := range unrecoverablePatterns {
		if strings.Contains(errStr, pattern) {
			return false
		}
	}

	return true
}

// recordError 记录错误
func (eh *EnhancedErrorHandler) recordError(err *EnhancedError) {
	eh.mu.Lock()
	defer eh.mu.Unlock()

	// 增加错误计数
	eh.errorCounts[err.Code]++

	// 添加到最近错误列表
	eh.recentErrors = append(eh.recentErrors, err)

	// 保持最近错误列表大小
	if len(eh.recentErrors) > eh.maxRecentErrors {
		eh.recentErrors = eh.recentErrors[1:]
	}
}

// displayError 显示错误信息
func (eh *EnhancedErrorHandler) displayError(err *EnhancedError, critical bool) {
	if eh.outputManager == nil {
		return
	}

	// 根据严重程度选择输出方式
	switch err.Severity {
	case SeverityCritical:
		eh.outputManager.Error("严重错误 [%s]: %s", err.Code, err.Message)
	case SeverityHigh:
		eh.outputManager.Error("高级错误 [%s]: %s", err.Code, err.Message)
	case SeverityMedium:
		eh.outputManager.Warn("中级错误 [%s]: %s", err.Code, err.Message)
	case SeverityLow:
		eh.outputManager.Info("低级错误 [%s]: %s", err.Code, err.Message)
	}

	// 显示建议
	if len(err.Suggestions) > 0 {
		eh.outputManager.Info("建议解决方案:")
		for i, suggestion := range err.Suggestions {
			eh.outputManager.Info("  %d. %s", i+1, suggestion)
		}
	}

	// 如果是关键错误，显示更多信息
	if critical {
		eh.outputManager.Debug("错误上下文: %+v", err.Context)
		if len(err.StackTrace) > 0 {
			eh.outputManager.Debug("堆栈跟踪:")
			for _, frame := range err.StackTrace {
				eh.outputManager.Debug("  %s", frame)
			}
		}
	}
}

// attemptRecovery 尝试错误恢复
func (eh *EnhancedErrorHandler) attemptRecovery(err *EnhancedError) {
	if !err.Recoverable {
		return
	}

	eh.outputManager.Info("尝试从错误中恢复...")

	// 根据错误类型尝试不同的恢复策略
	switch err.Category {
	case CategoryFileSystem:
		eh.recoverFileSystemError(err)
	case CategoryPermission:
		eh.recoverPermissionError(err)
	case CategoryNetwork:
		eh.recoverNetworkError(err)
	default:
		eh.outputManager.Warn("无法自动恢复此类型的错误")
	}
}

// recoverFileSystemError 恢复文件系统错误
func (eh *EnhancedErrorHandler) recoverFileSystemError(err *EnhancedError) {
	eh.outputManager.Info("尝试文件系统错误恢复...")
	// 实现文件系统错误恢复逻辑
}

// recoverPermissionError 恢复权限错误
func (eh *EnhancedErrorHandler) recoverPermissionError(err *EnhancedError) {
	eh.outputManager.Info("尝试权限错误恢复...")
	// 实现权限错误恢复逻辑
}

// recoverNetworkError 恢复网络错误
func (eh *EnhancedErrorHandler) recoverNetworkError(err *EnhancedError) {
	eh.outputManager.Info("尝试网络错误恢复...")
	// 实现网络错误恢复逻辑
}

// GetErrorStatistics 获取错误统计信息
func (eh *EnhancedErrorHandler) GetErrorStatistics() map[string]interface{} {
	eh.mu.RLock()
	defer eh.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_errors"] = len(eh.recentErrors)
	stats["error_counts"] = eh.errorCounts

	// 按严重程度统计
	severityStats := make(map[string]int)
	categoryStats := make(map[string]int)

	for _, err := range eh.recentErrors {
		switch err.Severity {
		case SeverityLow:
			severityStats["low"]++
		case SeverityMedium:
			severityStats["medium"]++
		case SeverityHigh:
			severityStats["high"]++
		case SeverityCritical:
			severityStats["critical"]++
		}

		switch err.Category {
		case CategoryFileSystem:
			categoryStats["filesystem"]++
		case CategoryPermission:
			categoryStats["permission"]++
		case CategoryNetwork:
			categoryStats["network"]++
		case CategoryConfiguration:
			categoryStats["configuration"]++
		case CategoryValidation:
			categoryStats["validation"]++
		case CategorySecurity:
			categoryStats["security"]++
		case CategorySystem:
			categoryStats["system"]++
		}
	}

	stats["severity_stats"] = severityStats
	stats["category_stats"] = categoryStats

	return stats
}

// SafeExecuteWithRecovery 安全执行函数，自动处理panic
func (eh *EnhancedErrorHandler) SafeExecuteWithRecovery(ctx context.Context, operation string, fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// 将panic转换为错误
			panicErr := fmt.Errorf("panic recovered: %v", r)
			err = eh.HandleEnhancedError(panicErr, operation, true)
		}
	}()

	// 执行函数
	err = fn()
	if err != nil {
		err = eh.HandleEnhancedError(err, operation, false)
	}

	return err
}

// 全局增强错误处理器实例
var globalEnhancedErrorHandler = NewEnhancedErrorHandler(globalOutputManager)

// 全局函数，方便使用
func HandleEnhancedError(err error, operation string, critical bool) *EnhancedError {
	return globalEnhancedErrorHandler.HandleEnhancedError(err, operation, critical)
}

func SafeExecuteWithRecovery(ctx context.Context, operation string, fn func() error) error {
	return globalEnhancedErrorHandler.SafeExecuteWithRecovery(ctx, operation, fn)
}

func GetEnhancedErrorStatistics() map[string]interface{} {
	return globalEnhancedErrorHandler.GetErrorStatistics()
}
