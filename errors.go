package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
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
	default:
		return 0
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
	return err.Error()
}

// Is 判断错误类型
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As 将错误转换为目标类型
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// ExitWithCode 根据错误类型退出程序
func ExitWithCode(err error) {
	if dgerr, ok := err.(*DGError); ok && dgerr.Kind != KindNone {
		os.Exit(dgerr.Kind.ExitCode())
	}
	os.Exit(1)
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
	return time.Now().Format("2006-01-02 15:04:05.000")
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
