package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
)

// Init 初始化日志系统
func Init(logFile, level string, maxSize, maxAge int, compress bool) error {
	// 确保日志目录存在
	logDir := filepath.Dir(logFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}

	// 打开日志文件
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %v", err)
	}

	// 创建不同级别的日志记录器
	infoLogger = log.New(file, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(file, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
	debugLogger = log.New(file, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)

	return nil
}

// Info 记录信息日志
func Info(msg string) {
	if infoLogger != nil {
		infoLogger.Println(msg)
	}
}

// Error 记录错误日志
func Error(msg string) {
	if errorLogger != nil {
		errorLogger.Println(msg)
	}
}

// Debug 记录调试日志
func Debug(msg string) {
	if debugLogger != nil {
		debugLogger.Println(msg)
	}
}

// Infof 格式化记录信息日志
func Infof(format string, args ...interface{}) {
	Info(fmt.Sprintf(format, args...))
}

// Errorf 格式化记录错误日志
func Errorf(format string, args ...interface{}) {
	Error(fmt.Sprintf(format, args...))
}

// Debugf 格式化记录调试日志
func Debugf(format string, args ...interface{}) {
	Debug(fmt.Sprintf(format, args...))
}
