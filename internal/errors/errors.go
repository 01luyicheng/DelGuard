package errors

import (
	"fmt"
	"runtime"
)

// ErrorType 错误类型
type ErrorType int

const (
	// ErrTypeUnknown 未知错误
	ErrTypeUnknown ErrorType = iota
	// ErrTypeFileNotFound 文件未找到
	ErrTypeFileNotFound
	// ErrTypePermissionDenied 权限拒绝
	ErrTypePermissionDenied
	// ErrTypeInvalidPath 无效路径
	ErrTypeInvalidPath
	// ErrTypeTrashFull 回收站已满
	ErrTypeTrashFull
	// ErrTypeConfigError 配置错误
	ErrTypeConfigError
	// ErrTypeNetworkError 网络错误
	ErrTypeNetworkError
)

// DelGuardError DelGuard自定义错误
type DelGuardError struct {
	Type    ErrorType
	Message string
	Cause   error
	File    string
	Line    int
}

// Error 实现error接口
func (e *DelGuardError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap 支持errors.Unwrap
func (e *DelGuardError) Unwrap() error {
	return e.Cause
}

// NewError 创建新的DelGuard错误
func NewError(errType ErrorType, message string, cause error) *DelGuardError {
	_, file, line, _ := runtime.Caller(1)
	return &DelGuardError{
		Type:    errType,
		Message: message,
		Cause:   cause,
		File:    file,
		Line:    line,
	}
}

// NewFileNotFoundError 创建文件未找到错误
func NewFileNotFoundError(path string) *DelGuardError {
	return NewError(ErrTypeFileNotFound, fmt.Sprintf("文件未找到: %s", path), nil)
}

// NewPermissionDeniedError 创建权限拒绝错误
func NewPermissionDeniedError(path string) *DelGuardError {
	return NewError(ErrTypePermissionDenied, fmt.Sprintf("权限不足: %s", path), nil)
}

// NewInvalidPathError 创建无效路径错误
func NewInvalidPathError(path string) *DelGuardError {
	return NewError(ErrTypeInvalidPath, fmt.Sprintf("无效路径: %s", path), nil)
}

// NewTrashFullError 创建回收站已满错误
func NewTrashFullError() *DelGuardError {
	return NewError(ErrTypeTrashFull, "回收站空间不足", nil)
}

// NewConfigError 创建配置错误
func NewConfigError(message string, cause error) *DelGuardError {
	return NewError(ErrTypeConfigError, fmt.Sprintf("配置错误: %s", message), cause)
}

// NewNetworkError 创建网络错误
func NewNetworkError(message string, cause error) *DelGuardError {
	return NewError(ErrTypeNetworkError, fmt.Sprintf("网络错误: %s", message), cause)
}

// IsType 检查错误类型
func IsType(err error, errType ErrorType) bool {
	if delErr, ok := err.(*DelGuardError); ok {
		return delErr.Type == errType
	}
	return false
}

// GetErrorMessage 获取用户友好的错误消息
func GetErrorMessage(err error) string {
	if delErr, ok := err.(*DelGuardError); ok {
		switch delErr.Type {
		case ErrTypeFileNotFound:
			return "指定的文件或目录不存在"
		case ErrTypePermissionDenied:
			return "权限不足，请检查文件权限或使用管理员权限运行"
		case ErrTypeInvalidPath:
			return "路径格式无效"
		case ErrTypeTrashFull:
			return "回收站空间不足，请清理回收站"
		case ErrTypeConfigError:
			return "配置文件错误，请检查配置"
		case ErrTypeNetworkError:
			return "网络连接失败，请检查网络设置"
		default:
			return delErr.Message
		}
	}
	return err.Error()
}