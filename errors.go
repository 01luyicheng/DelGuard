package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
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
	Kind   ErrKind
	Op     string
	Path   string
	Cause  error
	Advice string
	Code   string // 错误代码，用于国际化
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
		Kind:   kind,
		Op:     op,
		Path:   path,
		Cause:  cause,
		Advice: advice,
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

	// 根据错误类型提供更具体的错误信息
	if err != nil {
		switch {
		case os.IsNotExist(err):
			advice = fmt.Sprintf("文件或目录不存在: %s", path)
		case os.IsPermission(err):
			advice = fmt.Sprintf("权限不足，无法执行操作: %s", path)
		case os.IsTimeout(err):
			advice = fmt.Sprintf("操作超时: %s", path)
		default:
			advice = fmt.Sprintf("%s: %v", advice, err)
		}
	}

	// 如果已经是DGError，保留Kind
	if dgerr, ok := err.(*DGError); ok {
		return &DGError{
			Kind:   dgerr.Kind,
			Op:     operation,
			Path:   path,
			Cause:  err,
			Advice: advice,
		}
	}
	return &DGError{
		Kind:   KindIO,
		Op:     operation,
		Path:   path,
		Cause:  err,
		Advice: advice,
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
