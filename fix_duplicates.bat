@echo off
echo 正在清理重复函数声明...
echo.

:: 备份原始文件
if not exist "backup" mkdir backup

:: 备份关键文件
copy i18n.go backup\i18n.go.bak >nul 2>&1
copy errors.go backup\errors.go.bak >nul 2>&1
copy logger.go backup\logger.go.bak >nul 2>&1
copy restore.go backup\restore.go.bak >nul 2>&1
copy file_validator.go backup\file_validator.go.bak >nul 2>&1
copy fsutil_other.go backup\fsutil_other.go.bak >nul 2>&1
copy security_check.go backup\security_check.go.bak >nul 2>&1

echo 已创建备份文件到 backup 目录
echo.

:: 创建清理后的文件
echo 正在创建清理后的文件...

:: 创建清理后的i18n.go (移除重复的T函数)
(
echo package main

echo import ^(
echo	"fmt"
echo	"strings"
echo ^)

echo // 国际化支持
echo type I18n struct { }

echo // T 返回翻译后的文本
echo func T(key string, args ...interface{}) string {
echo	switch key {
echo	case "error_path_traversal":
echo		return "检测到路径遍历攻击"
echo	case "error_hidden_file":
echo		return "检测到隐藏文件"
echo	case "error_system_file":
echo		return "检测到系统文件"
echo	case "error_permission_denied":
echo		return "权限不足"
echo	case "error_file_not_found":
echo		return "文件不存在"
echo	case "error_invalid_path":
echo		return "无效的路径"
echo	case "security_check_passed":
echo		return "安全检查通过"
echo	case "security_check_failed":
echo		return "安全检查失败"
echo	default:
echo		return key
	}
}

echo // GetErrorAdvice 返回错误处理建议
echo func GetErrorAdvice(errorType string) string {
echo	switch errorType {
echo	case "path_traversal":
echo		return "请使用绝对路径，避免使用 .. 路径"
echo	case "hidden_file":
echo		return "请确认是否真的要删除隐藏文件"
echo	case "system_file":
echo		return "系统文件不建议删除，请确认操作"
echo	case "permission_denied":
echo		return "请检查文件权限或以管理员身份运行"
echo	default:
echo		return "请检查错误信息并重试"
	}
}
) > i18n_clean.go

:: 创建清理后的logger.go (移除重复的LogError)
(
echo package main

echo import ^(
echo	"fmt"
echo	"log"
echo	"os"
echo	"path/filepath"
echo	"time"
echo ^)

echo // Logger 日志记录器
echo type Logger struct { }

echo // NewLogger 创建新的日志记录器
echo func NewLogger() *Logger {
echo	return &Logger{}
}

echo // Info 记录信息日志
echo func (l *Logger) Info(format string, args ...interface{}) {
echo	log.Printf("[INFO] "+format, args...)
}

echo // Error 记录错误日志
echo func (l *Logger) Error(format string, args ...interface{}) {
echo	log.Printf("[ERROR] "+format, args...)
}

echo // Security 记录安全事件
echo func (l *Logger) Security(format string, args ...interface{}) {
echo	log.Printf("[SECURITY] "+format, args...)
}
) > logger_clean.go

echo 清理完成！
echo.
echo 下一步：
echo 1. 手动检查清理后的文件
echo 2. 替换原始文件
echo 3. 运行 go build 验证
echo.
pause