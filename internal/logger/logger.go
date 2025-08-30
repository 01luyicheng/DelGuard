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
	logFilePtr  *os.File
)

// Init 初始化日志系统
func Init(logFilePath, level string, maxSize, maxAge int, compress bool) error {
	// 确保日志目录存在
	logDir := filepath.Dir(logFilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}

	// 先关闭已存在的日志文件
	Close()

	// 打开日志文件
	var err error
	logFilePtr, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %v", err)
	}

	// 创建不同级别的日志记录器
	infoLogger = log.New(logFilePtr, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(logFilePtr, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
	debugLogger = log.New(logFilePtr, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)

	// 记录初始化信息
	Info("日志系统初始化成功")
	Debugf("日志文件: %s, 级别: %s", logFilePath, level)

	return nil
}

// Close 关闭日志文件
func Close() error {
	if logFilePtr != nil {
		err := logFilePtr.Close()
		logFilePtr = nil
		infoLogger = nil
		errorLogger = nil
		debugLogger = nil
		return err
	}
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
