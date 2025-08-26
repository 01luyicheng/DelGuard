package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// 统一错误种类与退出码映射
type ErrKind int

const (
	KindNone              ErrKind = iota
	KindCancelled                 // 2
	KindInvalidArgs               // 3
	KindPermission                // 5
	KindIO                        // 10
	KindNotFound                  // 11
	KindProtected                 // 12
	KindSecurity                  // 13 - 安全相关错误
	KindMalware                   // 14 - 恶意软件检测
	KindPathTraversal             // 15 - 路径遍历攻击
	KindHiddenFile                // 16 - 隐藏文件
	KindSystemFile                // 17 - 系统文件
	KindSpecialFile               // 18 - 特殊文件类型
	KindIntegrity                 // 19 - 完整性检查失败
	KindQuota                     // 20 - 配额限制
	KindConfig                    // 21 - 配置错误
	KindNetwork                   // 22 - 网络错误
	KindValidation                // 23 - 验证失败
	KindResourceExhausted         // 24 - 资源耗尽
	KindTimeout                   // 25 - 操作超时
	KindConflict                  // 26 - 文件冲突
	KindTrashOperation            // 27 - 回收站操作
	KindDelGuardProject           // 28 - DelGuard项目文件
	KindLongFileName              // 29 - 文件名过长
	KindSpecialCharacters         // 30 - 特殊字符
	KindUnicodeIssue              // 31 - Unicode问题
	KindSpaceIssue                // 32 - 空格问题
	KindReadOnlyFile              // 33 - 只读文件
	KindRecoverable               // 34 - 可恢复错误
	KindTransient                 // 35 - 临时错误
	KindRetryable                 // 36 - 可重试错误
	KindCorrupted                 // 37 - 文件损坏
	KindDiskFull                  // 38 - 磁盘空间不足
	KindConcurrency               // 39 - 并发冲突
	KindDeadlock                  // 40 - 死锁
	KindCircularRef               // 41 - 循环引用
)

// 预定义错误
var (
	ErrUnsupportedPlatform = errors.New("不支持的操作系统平台")
	ErrFileNotFound        = errors.New("文件不存在")
)

func (k ErrKind) ExitCode() int {
	switch k {
	case KindCancelled:
		return 2
	case KindInvalidArgs:
		return 3
	case KindPermission:
		return 5
	case KindIO:
		return 10
	case KindNotFound:
		return 11
	case KindProtected:
		return 12
	case KindSecurity:
		return 13
	case KindMalware:
		return 14
	case KindPathTraversal:
		return 15
	case KindHiddenFile:
		return 16
	case KindSystemFile:
		return 17
	case KindSpecialFile:
		return 18
	case KindIntegrity:
		return 19
	case KindQuota:
		return 20
	case KindConfig:
		return 21
	case KindNetwork:
		return 22
	case KindValidation:
		return 23
	case KindResourceExhausted:
		return 24
	case KindTimeout:
		return 25
	case KindConflict:
		return 26
	case KindTrashOperation:
		return 27
	case KindDelGuardProject:
		return 28
	case KindLongFileName:
		return 29
	case KindSpecialCharacters:
		return 30
	case KindUnicodeIssue:
		return 31
	case KindSpaceIssue:
		return 32
	case KindReadOnlyFile:
		return 33
	case KindRecoverable:
		return 34
	case KindTransient:
		return 35
	case KindRetryable:
		return 36
	case KindCorrupted:
		return 37
	case KindDiskFull:
		return 38
	case KindConcurrency:
		return 39
	case KindDeadlock:
		return 40
	case KindCircularRef:
		return 41
	default:
		return 0
	}
}

// GetUserFriendlyMessage 获取用户友好的错误消息
func (k ErrKind) GetUserFriendlyMessage() string {
	switch k {
	case KindCancelled:
		return "操作被取消"
	case KindInvalidArgs:
		return "命令行参数错误"
	case KindPermission:
		return "权限不足，无法执行操作"
	case KindIO:
		return "文件读写错误"
	case KindNotFound:
		return "文件或目录不存在"
	case KindProtected:
		return "文件受到保护，不允许删除"
	case KindSecurity:
		return "安全检查失败"
	case KindMalware:
		return "检测到恶意软件"
	case KindPathTraversal:
		return "检测到路径遍历攻击"
	case KindHiddenFile:
		return "检测到隐藏文件"
	case KindSystemFile:
		return "检测到系统关键文件"
	case KindSpecialFile:
		return "检测到特殊文件类型"
	case KindIntegrity:
		return "文件完整性检查失败"
	case KindQuota:
		return "超出配额限制"
	case KindConfig:
		return "配置错误"
	case KindNetwork:
		return "网络错误"
	case KindValidation:
		return "验证失败"
	case KindResourceExhausted:
		return "系统资源不足"
	case KindTimeout:
		return "操作超时"
	case KindConflict:
		return "文件冲突"
	case KindTrashOperation:
		return "回收站操作警告"
	case KindDelGuardProject:
		return "DelGuard项目文件保护"
	case KindLongFileName:
		return "文件名过长"
	case KindSpecialCharacters:
		return "文件名包含特殊字符"
	case KindUnicodeIssue:
		return "文件名Unicode问题"
	case KindSpaceIssue:
		return "文件名空格问题"
	case KindReadOnlyFile:
		return "只读文件"
	case KindRecoverable:
		return "可恢复错误"
	case KindTransient:
		return "临时错误"
	case KindRetryable:
		return "可重试错误"
	case KindCorrupted:
		return "文件损坏"
	case KindDiskFull:
		return "磁盘空间不足"
	case KindConcurrency:
		return "并发冲突"
	case KindDeadlock:
		return "死锁检测"
	case KindCircularRef:
		return "循环引用"
	default:
		return "未知错误"
	}
}

// GetSuggestion 获取错误建议
func (k ErrKind) GetSuggestion() string {
	switch k {
	case KindCancelled:
		return "用户主动取消操作"
	case KindInvalidArgs:
		return "请检查命令行参数，使用 --help 查看帮助"
	case KindPermission:
		return "请以管理员身份运行，或检查文件权限"
	case KindIO:
		return "请检查磁盘空间和文件系统状态"
	case KindNotFound:
		return "请检查文件路径是否正确，或使用智能搜索功能"
	case KindProtected:
		return "如果确实需要删除，请使用 --force 参数"
	case KindSecurity:
		return "请检查文件安全性，确认操作合法"
	case KindMalware:
		return "建议使用杀毒软件扫描，确认安全后再操作"
	case KindPathTraversal:
		return "检测到可能的攻击行为，请使用安全的文件路径"
	case KindHiddenFile:
		return "隐藏文件可能包含重要数据，请谨慎操作"
	case KindSystemFile:
		return "系统文件对系统稳定性至关重要，建议不要删除"
	case KindSpecialFile:
		return "特殊文件类型可能有特殊用途，请确认后操作"
	case KindIntegrity:
		return "文件可能已损坏，请检查文件完整性"
	case KindQuota:
		return "请清理磁盘空间或联系管理员提高配额"
	case KindConfig:
		return "请检查配置文件语法和参数设置"
	case KindNetwork:
		return "请检查网络连接和防火墙设置"
	case KindValidation:
		return "请检查输入参数的格式和合法性"
	case KindResourceExhausted:
		return "请等待系统资源释放或重启程序"
	case KindTimeout:
		return "操作超时，请检查网络或系统负载"
	case KindConflict:
		return "文件冲突，请检查目标文件状态"
	case KindTrashOperation:
		return "正在操作回收站，请谨慎确认"
	case KindDelGuardProject:
		return "DelGuard项目文件受保护，如需删除请使用 --force"
	case KindLongFileName:
		return "文件名过长可能导致兼容性问题"
	case KindSpecialCharacters:
		return "文件名包含特殊字符，可能影响跨平台兼容性"
	case KindUnicodeIssue:
		return "文件名包含Unicode问题，可能影响显示和处理"
	case KindSpaceIssue:
		return "文件名空格问题可能导致命令行操作困难"
	case KindReadOnlyFile:
		return "只读文件通常包含重要数据，请确认后操作"
	case KindRecoverable:
		return "错误可恢复，系统将自动尝试恢复"
	case KindTransient:
		return "临时错误，稍后重试可能成功"
	case KindRetryable:
		return "操作失败，但可以重试"
	case KindCorrupted:
		return "检查文件完整性，必要时修复后重试"
	case KindDiskFull:
		return "清理磁盘空间或选择其他位置"
	case KindConcurrency:
		return "检测到并发冲突，稍后重试或避免同时操作"
	case KindDeadlock:
		return "检测到死锁，系统将释放资源后重试"
	case KindCircularRef:
		return "检测到循环引用，请检查文件结构后操作"
	default:
		return "请检查系统日志或联系技术支持"
	}
}

// DGError 携带错误分类与上下文
type DGError struct {
	Kind      ErrKind
	Op        string
	Path      string
	Cause     error
	Advice    string
	Code      string       // 错误代码，用于国际化
	Timestamp string       // 错误发生时间
	Stack     []StackFrame // 堆栈跟踪信息
}

// StackFrame 堆栈帧信息
type StackFrame struct {
	Function string
	File     string
	Line     int
}

func (e *DGError) Error() string {
	if e == nil {
		return ""
	}
	if e.Op != "" && e.Path != "" {
		return e.Op + " " + e.Path + ": " + unwrapMsg(e.Cause)
	}
	if e.Op != "" {
		return e.Op + ": " + unwrapMsg(e.Cause)
	}
	return unwrapMsg(e.Cause)
}

func (e *DGError) Unwrap() error { return e.Cause }

func unwrapMsg(err error) string {
	if err == nil {
		return "<nil>"
	}
	// 展开嵌套错误
	msg := err.Error()
	for {
		unwrapable, ok := err.(interface{ Unwrap() error })
		if !ok {
			break
		}
		err = unwrapable.Unwrap()
		if err == nil {
			break
		}
		msg += ": " + err.Error()
	}
	return msg
}

// E 创建新的DGError
func E(kind ErrKind, op, path string, cause error, advice string) *DGError {
	return &DGError{
		Kind:      kind,
		Op:        op,
		Path:      path,
		Cause:     cause,
		Advice:    advice,
		Timestamp: getCurrentTime(),
		Stack:     captureStackTrace(2), // 跳过当前和E函数
	}
}

// WrapE 包装错误，提供上下文信息
func WrapE(operation string, path string, err error) *DGError {
	var advice string
	if path != "" {
		advice = fmt.Sprintf("操作 '%s' 在路径 '%s' 失败", operation, path)
	} else {
		advice = fmt.Sprintf("操作 '%s' 失败", operation)
	}

	// 根据错误类型提供更具体的错误信息和类型
	kind := KindIO // 默认类型
	if err != nil {
		switch {
		case os.IsNotExist(err):
			advice = fmt.Sprintf("文件或目录不存在: %s", path)
			kind = KindNotFound
		case os.IsPermission(err):
			advice = fmt.Sprintf("权限不足，无法执行操作: %s", path)
			kind = KindPermission
		case os.IsTimeout(err):
			advice = fmt.Sprintf("操作超时: %s", path)
			kind = KindTimeout
		case os.IsExist(err):
			advice = fmt.Sprintf("文件已存在: %s", path)
			kind = KindConflict
		default:
			// 检查是否为特定的系统错误
			if strings.Contains(err.Error(), "access denied") || strings.Contains(err.Error(), "permission denied") {
				kind = KindPermission
			} else if strings.Contains(err.Error(), "no space left") {
				kind = KindResourceExhausted
			} else if strings.Contains(err.Error(), "invalid argument") {
				kind = KindInvalidArgs
			}
			advice = fmt.Sprintf("%s: %v", advice, err)
		}
	}

	// 如果已经是DGError，保留原始的Kind和堆栈信息
	if dgerr, ok := err.(*DGError); ok {
		return &DGError{
			Kind:      dgerr.Kind,
			Op:        operation,
			Path:      path,
			Cause:     err,
			Advice:    advice,
			Timestamp: getCurrentTime(),
			Stack:     dgerr.Stack, // 保留原始堆栈
		}
	}

	return &DGError{
		Kind:      kind,
		Op:        operation,
		Path:      path,
		Cause:     err,
		Advice:    advice,
		Timestamp: getCurrentTime(),
		Stack:     captureStackTrace(2), // 跳过当前和WrapE函数
	}
}

// Errorf 创建格式化的错误
func Errorf(kind ErrKind, op, path, advice, format string, a ...interface{}) *DGError {
	return &DGError{
		Kind:   kind,
		Op:     op,
		Path:   path,
		Cause:  fmt.Errorf(format, a...),
		Advice: advice,
	}
}

// FormatErrorForDisplay 格式化错误用于显示
func FormatErrorForDisplay(err error) string {
	if dgerr, ok := err.(*DGError); ok {
		var sb strings.Builder
		// 优先加入用户友好消息（基于 Kind）
		if dgerr.Kind != KindNone {
			sb.WriteString(dgerr.Kind.GetUserFriendlyMessage())
			sb.WriteString("\n")
		}
		// 追加原始错误（包含操作与路径）
		sb.WriteString(dgerr.Error())
		if dgerr.Advice != "" {
			sb.WriteString("\n💡 建议: ")
			sb.WriteString(dgerr.Advice)
		}
		if dgerr.Kind != KindNone {
			sb.WriteString(fmt.Sprintf(" (错误代码: DG%02d)", dgerr.Kind.ExitCode()))
		}
		return sb.String()
	}
	if err == nil {
		return ""
	}
	return err.Error()
}

// As 将错误转换为目标类型
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// ExitHandler 可注入的退出处理器，默认调用 os.Exit，可测试时替换
var ExitHandler = func(code int) {
	os.Exit(code)
}

// ExitWithCode 根据错误种类退出程序（可测试）
func ExitWithCode(err error) {
	code := 1
	if dgerr, ok := err.(*DGError); ok && dgerr.Kind != KindNone {
		code = dgerr.Kind.ExitCode()
	}
	ExitHandler(code)
}

// captureStackTrace 捕获堆栈跟踪信息
func captureStackTrace(skip int) []StackFrame {
	var frames []StackFrame

	// 最多捕莱10层堆栈
	for i := skip; i < skip+10; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		var funcName string
		if fn != nil {
			funcName = fn.Name()
		} else {
			funcName = "unknown"
		}

		frames = append(frames, StackFrame{
			Function: funcName,
			File:     file,
			Line:     line,
		})
	}

	return frames
}

// getCurrentTime 获取当前时间的格式化字符串
func getCurrentTime() string {
	return time.Now().Format(TimeFormatWithMillis)
}

// StackString 返回堆栈跟踪的字符串表示
func (e *DGError) StackString() string {
	if len(e.Stack) == 0 {
		return "无堆栈信息"
	}

	var sb strings.Builder
	sb.WriteString("堆栈跟踪:\n")
	for i, frame := range e.Stack {
		sb.WriteString(fmt.Sprintf("%d. %s\n\t%s:%d\n", i+1, frame.Function, frame.File, frame.Line))
	}
	return sb.String()
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxAttempts       int           // 最大重试次数
	InitialDelay      time.Duration // 初始延迟
	BackoffMultiplier float64       // 退避倍数
	MaxDelay          time.Duration // 最大延迟
	RetryableErrors   []ErrKind     // 可重试的错误类型
}

// DefaultRetryConfig 默认重试配置
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:       3,
		InitialDelay:      100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		MaxDelay:          5 * time.Second,
		RetryableErrors: []ErrKind{
			KindRetryable,
			KindTransient,
			KindTimeout,
			KindNetwork,
			KindIO,
			KindResourceExhausted,
			KindConcurrency,
		},
	}
}

// IsRetryable 判断错误是否可重试
func (k ErrKind) IsRetryable() bool {
	switch k {
	case KindRetryable, KindTransient, KindTimeout, KindNetwork, KindIO, KindResourceExhausted, KindConcurrency:
		return true
	default:
		return false
	}
}

// GetRecoveryStrategy 获取错误恢复策略
func (k ErrKind) GetRecoveryStrategy() string {
	switch k {
	case KindRecoverable:
		return "尝试自动恢复"
	case KindTransient:
		return "等待后重试"
	case KindRetryable:
		return "立即重试"
	case KindCorrupted:
		return "检查文件完整性后修复"
	case KindDiskFull:
		return "清理磁盘空间后重试"
	case KindConcurrency:
		return "避免并发冲突后重试"
	case KindDeadlock:
		return "释放资源后重试"
	case KindCircularRef:
		return "修复循环引用后重试"
	default:
		return "无特定恢复策略"
	}
}

// WithRetry 使用重试机制执行操作
func WithRetry(config *RetryConfig, operation func() error) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	var lastErr error
	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		err := operation()
		if err == nil {
			return nil // 成功
		}

		lastErr = err

		// 检查是否为可重试错误
		if dgerr, ok := err.(*DGError); ok {
			if !dgerr.Kind.IsRetryable() {
				return err // 不可重试的错误，立即返回
			}
		} else {
			// 非DGError，根据内容判断是否可重试
			if !isRetryableError(err) {
				return err
			}
		}

		if attempt < config.MaxAttempts {
			// 计算延迟时间
			delay := time.Duration(float64(config.InitialDelay) * float64(attempt-1) * config.BackoffMultiplier)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
			time.Sleep(delay)
		}
	}

	return lastErr
}

// isRetryableError 判断普通错误是否可重试
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	retryablePatterns := []string{
		"timeout",
		"connection refused",
		"temporary",
		"try again",
		"resource temporarily unavailable",
		"device busy",
		"operation would block",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// RecoveryManager 错误恢复管理器
type RecoveryManager struct {
	config *RetryConfig
	logger func(msg string) // 日志记录函数
}

// NewRecoveryManager 创建错误恢复管理器
func NewRecoveryManager(config *RetryConfig, logger func(string)) *RecoveryManager {
	if config == nil {
		config = DefaultRetryConfig()
	}
	if logger == nil {
		logger = func(string) {} // 空日志函数
	}
	return &RecoveryManager{
		config: config,
		logger: logger,
	}
}

// TryRecover 尝试恢复错误
func (rm *RecoveryManager) TryRecover(err error, operation func() error) error {
	if err == nil {
		return nil
	}

	// 检查是否为可恢复错误
	if dgerr, ok := err.(*DGError); ok {
		switch dgerr.Kind {
		case KindRecoverable, KindTransient, KindRetryable:
			rm.logger(fmt.Sprintf("尝试恢复错误: %s, 策略: %s", dgerr.Error(), dgerr.Kind.GetRecoveryStrategy()))
			return WithRetry(rm.config, operation)
		case KindDiskFull:
			rm.logger("检测到磁盘空间不足，尝试清理临时文件")
			// 这里可以添加清理临时文件的逻辑
			return WithRetry(rm.config, operation)
		case KindConcurrency:
			rm.logger("检测到并发冲突，稍后重试")
			// 增加随机延迟避免并发冲突
			time.Sleep(time.Duration(50) * time.Millisecond)
			return WithRetry(rm.config, operation)
		default:
			return err // 不可恢复的错误
		}
	}

	return err
}

// ErrorCollector 错误收集器，用于批量操作
type ErrorCollector struct {
	errors []error
	mu     sync.Mutex
}

// NewErrorCollector 创建错误收集器
func NewErrorCollector() *ErrorCollector {
	return &ErrorCollector{
		errors: make([]error, 0),
	}
}

// Add 添加错误
func (ec *ErrorCollector) Add(err error) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.errors = append(ec.errors, err)
}

// HasErrors 检查是否有错误
func (ec *ErrorCollector) HasErrors() bool {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	return len(ec.errors) > 0
}

// GetErrors 获取所有错误
func (ec *ErrorCollector) GetErrors() []error {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	errors := make([]error, len(ec.errors))
	copy(errors, ec.errors)
	return errors
}

// Summary 获取错误摘要
func (ec *ErrorCollector) Summary() string {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	if len(ec.errors) == 0 {
		return "无错误"
	}

	// 统计错误类型
	kindCount := make(map[ErrKind]int)
	for _, err := range ec.errors {
		if dgerr, ok := err.(*DGError); ok {
			kindCount[dgerr.Kind]++
		} else {
			kindCount[KindIO]++ // 默认归类为IO错误
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("总计错误: %d\n", len(ec.errors)))
	for kind, count := range kindCount {
		sb.WriteString(fmt.Sprintf("- %s: %d\n", kind.GetUserFriendlyMessage(), count))
	}

	return sb.String()
}

// Clear 清空错误
func (ec *ErrorCollector) Clear() {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.errors = ec.errors[:0]
}

// ErrorHandler 错误处理器接口
type ErrorHandler interface {
	HandleError(err error, operation string, critical bool) bool
}

// DefaultErrorHandler 默认错误处理器
type DefaultErrorHandler struct {
	exitOnCritical bool
}

// NewDefaultErrorHandler 创建默认错误处理器
func NewDefaultErrorHandler(exitOnCritical bool) *DefaultErrorHandler {
	return &DefaultErrorHandler{
		exitOnCritical: exitOnCritical,
	}
}

// HandleError 处理错误
func (h *DefaultErrorHandler) HandleError(err error, operation string, critical bool) bool {
	if err == nil {
		return true
	}

	// 记录错误日志
	LogError(operation, "", err)

	// 如果是关键错误且配置为退出
	if critical && h.exitOnCritical {
		fmt.Fprintf(os.Stderr, T("关键错误: %v\n"), err)
		if ExitHandler != nil {
			ExitHandler(1)
		}
		return false
	}

	// 非关键错误，只输出警告
	fmt.Fprintf(os.Stderr, T("警告: %s 操作失败: %v\n"), operation, err)
	return false
}

// 全局错误处理器实例
var globalErrorHandler ErrorHandler = NewDefaultErrorHandler(true)

// SetErrorHandler 设置全局错误处理器
func SetErrorHandler(handler ErrorHandler) {
	globalErrorHandler = handler
}

// HandleError 全局错误处理函数
func HandleError(err error, operation string, critical bool) bool {
	return globalErrorHandler.HandleError(err, operation, critical)
}
