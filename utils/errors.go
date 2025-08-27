// Package utils 提供通用的工具函数
package utils

import (
	"fmt"
)

// StandardError 标准错误结构
type StandardError struct {
	Code    string       // 错误代码
	Message string       // 错误消息
	Details string       // 详细信息
	Cause   error        // 原始错误
	Context ErrorContext // 错误上下文
}

// ErrorContext 错误上下文信息
type ErrorContext struct {
	Operation string            // 操作名称
	Resource  string            // 涉及的资源
	Metadata  map[string]string // 其他元数据
}

// Error 实现error接口
func (e *StandardError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 实现错误解包接口
func (e *StandardError) Unwrap() error {
	return e.Cause
}

// NewStandardError 创建标准错误
func NewStandardError(code, message string, cause error, context ErrorContext) *StandardError {
	return &StandardError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Context: context,
	}
}

// Errorf 创建带格式化消息的标准错误
func Errorf(code, messageFormat string, cause error, context ErrorContext, a ...interface{}) *StandardError {
	return &StandardError{
		Code:    code,
		Message: fmt.Sprintf(messageFormat, a...),
		Cause:   cause,
		Context: context,
	}
}

// Wrap 包装现有错误，添加上下文信息
func Wrap(err error, code, message string, context ErrorContext) *StandardError {
	return &StandardError{
		Code:    code,
		Message: message,
		Cause:   err,
		Context: context,
	}
}
