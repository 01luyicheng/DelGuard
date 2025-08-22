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
	KindNone      ErrKind = iota
	KindCancelled         // 2
	KindInvalidArgs
	KindPermission
	KindIO
	KindNotFound
	KindProtected
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
		return ""
	}
	return err.Error()
}

// E 构造 DGError
func E(kind ErrKind, op, path string, cause error, advice string) *DGError {
	return &DGError{Kind: kind, Op: op, Path: path, Cause: cause, Advice: advice}
}

// WrapE 用于将外部 error 包装为 DGError（按常见系统错误归类）
func WrapE(op, path string, err error) *DGError {
	if err == nil {
		return nil
	}
	switch {
	case os.IsPermission(err):
		return E(KindPermission, op, path, err, "")
	case os.IsNotExist(err):
		return E(KindNotFound, op, path, err, "")
	default:
		return E(KindIO, op, path, err, "")
	}
}

// ExitCodeFrom 根据 error 推断退出码（若不是 DGError，则按常见系统错误推断）
func ExitCodeFrom(err error) int {
	if err == nil {
		return 0
	}
	var de *DGError
	if errors.As(err, &de) && de != nil {
		return de.Kind.ExitCode()
	}
	switch {
	case os.IsPermission(err):
		return KindPermission.ExitCode()
	case os.IsNotExist(err):
		return KindNotFound.ExitCode()
	default:
		return KindIO.ExitCode()
	}
}

// ChooseExitCode 聚合多目标删除时的退出码优先级：
// 5(权限) > 11(不存在) > 12(受保护路径拦截且无其他错误与成功) > 10(其他I/O/预处理错误) > 0
func ChooseExitCode(permDenied, notFound, protected, success, ioErr, preErr int) int {
	if permDenied > 0 {
		return KindPermission.ExitCode()
	}
	if notFound > 0 {
		return KindNotFound.ExitCode()
	}
	if protected > 0 && success == 0 && ioErr == 0 && preErr == 0 {
		return KindProtected.ExitCode()
	}
	if ioErr > 0 || preErr > 0 {
		return KindIO.ExitCode()
	}
	return 0
}

var (
	// ErrUnsupportedPlatform 不支持的平台错误
	ErrUnsupportedPlatform = errors.New("不支持的操作系统平台")

	// ErrFileNotFound 文件不存在错误
	ErrFileNotFound = errors.New("文件不存在")

	// ErrPermissionDenied 权限不足错误
	ErrPermissionDenied = errors.New("权限不足")

	// ErrTrashOperationFailed 回收站操作失败错误
	ErrTrashOperationFailed = errors.New("回收站操作失败")

	// ErrCriticalPath 关键路径错误
	ErrCriticalPath = errors.New("关键受保护路径")

	// ErrContainsDelGuard DelGuard程序目录错误
	ErrContainsDelGuard = errors.New("包含DelGuard程序目录")

	// ErrTrashDirectory 回收站目录错误
	ErrTrashDirectory = errors.New("回收站/废纸篓目录")

	// ErrReadOnlyFile 只读文件错误
	ErrReadOnlyFile = errors.New("只读文件")

	// ErrUserCancelled 用户取消操作错误
	ErrUserCancelled = errors.New("用户取消操作")

	// ErrConfigLoadFailed 配置文件加载失败错误
	ErrConfigLoadFailed = errors.New("配置文件加载失败")

	// ErrInvalidPath 无效路径错误
	ErrInvalidPath = errors.New("无效的文件路径")
)

// IsCriticalError 检查是否为关键错误（需要特殊处理的错误类型）
func IsCriticalError(err error) bool {
	if err == nil {
		return false
	}

	var dgErr *DGError
	if errors.As(err, &dgErr) {
		switch dgErr.Kind {
		case KindProtected, KindPermission:
			return true
		}
	}

	// 检查特定的错误消息
	errMsg := err.Error()
	return strings.Contains(errMsg, "critical") ||
		strings.Contains(errMsg, "protected") ||
		strings.Contains(errMsg, "permission") ||
		strings.Contains(errMsg, "回收站") ||
		strings.Contains(errMsg, "关键路径") ||
		strings.Contains(errMsg, "DelGuard程序")
}

// GetErrorAdvice 根据错误类型提供建议信息
func GetErrorAdvice(err error) string {
	if err == nil {
		return ""
	}

	var dgErr *DGError
	if errors.As(err, &dgErr) && dgErr.Advice != "" {
		return dgErr.Advice
	}

	// 根据错误类型提供默认建议
	errMsg := err.Error()
	switch {
	case strings.Contains(errMsg, "permission") || strings.Contains(errMsg, "权限"):
		return "请检查文件权限或使用管理员权限运行"
	case strings.Contains(errMsg, "not exist") || strings.Contains(errMsg, "不存在"):
		return "请检查文件路径是否正确"
	case strings.Contains(errMsg, "critical") || strings.Contains(errMsg, "关键路径"):
		return "此路径受到保护，无法删除"
	case strings.Contains(errMsg, "trash") || strings.Contains(errMsg, "回收站"):
		return "无法删除回收站目录"
	case strings.Contains(errMsg, "read-only") || strings.Contains(errMsg, "只读"):
		return "请先取消文件的只读属性"
	default:
		return "请检查文件是否被其他程序占用或磁盘空间是否充足"
	}
}

// FormatErrorForDisplay 格式化错误信息用于显示
func FormatErrorForDisplay(err error) string {
	if err == nil {
		return ""
	}

	advice := GetErrorAdvice(err)
	if advice != "" {
		return fmt.Sprintf("错误: %s\n建议: %s", err.Error(), advice)
	}
	return fmt.Sprintf("错误: %s", err.Error())
}
