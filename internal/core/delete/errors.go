package delete

import (
	"fmt"
	"strings"
)

// DeleteError 删除操作错误类型
type DeleteError struct {
	Op       string // 操作类型
	Path     string // 文件路径
	Err      error  // 原始错误
	Code     ErrorCode
	Retryable bool   // 是否可重试
}

// ErrorCode 错误代码
type ErrorCode int

const (
	// ErrUnknown 未知错误
	ErrUnknown ErrorCode = iota
	// ErrFileNotFound 文件不存在
	ErrFileNotFound
	// ErrPermissionDenied 权限被拒绝
	ErrPermissionDenied
	// ErrProtectedPath 受保护的路径
	ErrProtectedPath
	// ErrInvalidPath 无效路径
	ErrInvalidPath
	// ErrFileInUse 文件正在使用
	ErrFileInUse
	// ErrDiskFull 磁盘空间不足
	ErrDiskFull
	// ErrNetworkError 网络错误
	ErrNetworkError
	// ErrTimeout 操作超时
	ErrTimeout
	// ErrCancelled 操作被取消
	ErrCancelled
)

// Error 实现error接口
func (e *DeleteError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("%s %s: %v", e.Op, e.Path, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

// Unwrap 返回原始错误
func (e *DeleteError) Unwrap() error {
	return e.Err
}

// Is 检查错误类型
func (e *DeleteError) Is(target error) bool {
	if t, ok := target.(*DeleteError); ok {
		return e.Code == t.Code
	}
	return false
}

// NewDeleteError 创建删除错误
func NewDeleteError(op, path string, err error) *DeleteError {
	code := classifyError(err)
	retryable := isRetryable(code)
	
	return &DeleteError{
		Op:        op,
		Path:      path,
		Err:       err,
		Code:      code,
		Retryable: retryable,
	}
}

// classifyError 分类错误
func classifyError(err error) ErrorCode {
	if err == nil {
		return ErrUnknown
	}
	
	errStr := strings.ToLower(err.Error())
	
	switch {
	case strings.Contains(errStr, "no such file") || strings.Contains(errStr, "cannot find"):
		return ErrFileNotFound
	case strings.Contains(errStr, "permission denied") || strings.Contains(errStr, "access denied"):
		return ErrPermissionDenied
	case strings.Contains(errStr, "protected") || strings.Contains(errStr, "system"):
		return ErrProtectedPath
	case strings.Contains(errStr, "invalid") || strings.Contains(errStr, "illegal"):
		return ErrInvalidPath
	case strings.Contains(errStr, "in use") || strings.Contains(errStr, "being used"):
		return ErrFileInUse
	case strings.Contains(errStr, "no space") || strings.Contains(errStr, "disk full"):
		return ErrDiskFull
	case strings.Contains(errStr, "network") || strings.Contains(errStr, "connection"):
		return ErrNetworkError
	case strings.Contains(errStr, "timeout") || strings.Contains(errStr, "timed out"):
		return ErrTimeout
	case strings.Contains(errStr, "cancel") || strings.Contains(errStr, "context"):
		return ErrCancelled
	default:
		return ErrUnknown
	}
}

// isRetryable 判断错误是否可重试
func isRetryable(code ErrorCode) bool {
	switch code {
	case ErrFileInUse, ErrDiskFull, ErrNetworkError, ErrTimeout:
		return true
	case ErrFileNotFound, ErrPermissionDenied, ErrProtectedPath, ErrInvalidPath, ErrCancelled:
		return false
	default:
		return false
	}
}

// ErrorCodeString 返回错误代码的字符串表示
func (e ErrorCode) String() string {
	switch e {
	case ErrUnknown:
		return "UNKNOWN"
	case ErrFileNotFound:
		return "FILE_NOT_FOUND"
	case ErrPermissionDenied:
		return "PERMISSION_DENIED"
	case ErrProtectedPath:
		return "PROTECTED_PATH"
	case ErrInvalidPath:
		return "INVALID_PATH"
	case ErrFileInUse:
		return "FILE_IN_USE"
	case ErrDiskFull:
		return "DISK_FULL"
	case ErrNetworkError:
		return "NETWORK_ERROR"
	case ErrTimeout:
		return "TIMEOUT"
	case ErrCancelled:
		return "CANCELLED"
	default:
		return "UNKNOWN"
	}
}

// BatchDeleteError 批量删除错误
type BatchDeleteError struct {
	Errors []DeleteResult
	Total  int
	Failed int
}

// Error 实现error接口
func (e *BatchDeleteError) Error() string {
	return fmt.Sprintf("批量删除失败: %d/%d 个文件删除失败", e.Failed, e.Total)
}

// GetFailedResults 获取失败的结果
func (e *BatchDeleteError) GetFailedResults() []DeleteResult {
	var failed []DeleteResult
	for _, result := range e.Errors {
		if !result.Success {
			failed = append(failed, result)
		}
	}
	return failed
}

// NewBatchDeleteError 创建批量删除错误
func NewBatchDeleteError(results []DeleteResult) *BatchDeleteError {
	failed := 0
	for _, result := range results {
		if !result.Success {
			failed++
		}
	}
	
	return &BatchDeleteError{
		Errors: results,
		Total:  len(results),
		Failed: failed,
	}
}