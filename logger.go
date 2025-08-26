package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// 全局日志记录器实例
var logger *Logger

// LogLevel 日志级别
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

const (
	// LogDirPerm 日志目录权限
	LogDirPerm = 0700

	// LogFilePerm 日志文件权限
	LogFilePerm = 0600
)

// Logger 日志记录器
type Logger struct {
	mu         sync.Mutex
	level      LogLevel
	file       *os.File
	logDir     string
	maxSize    int64 // 最大文件大小(MB)
	maxBackups int   // 最大备份文件数
}

// NewLogger 创建新的日志记录器
func NewLogger(level string) *Logger {
	// 获取配置实例
	config, err := LoadConfig()
	if err != nil {
		// 如果配置加载失败，使用默认值
		config = &Config{}
		config.setDefaults()
	}

	l := &Logger{
		level:      parseLogLevel(level),
		logDir:     getLogDir(),
		maxSize:    config.LogMaxSize,    // 使用配置中的值
		maxBackups: config.LogMaxBackups, // 使用配置中的值
	}

	// 确保日志目录存在，使用更安全的权限 LogDirPerm（仅所有者可访问）
	if err := os.MkdirAll(l.logDir, LogDirPerm); err != nil {
		fmt.Fprintf(os.Stderr, "无法创建日志目录: %v\n", err)
		return l
	}

	// 设置日志文件
	if err := l.setupLogFile(); err != nil {
		fmt.Fprintf(os.Stderr, "无法设置日志文件: %v\n", err)
	}

	return l
}

// InitGlobalLogger 初始化全局日志记录器
func InitGlobalLogger(level string) {
	logger = NewLogger(level)
}

// setupLogFile 设置日志文件
func (l *Logger) setupLogFile() error {
	logPath := filepath.Join(l.logDir, "delguard.log")

	// 检查是否需要轮转
	if l.needsRotation(logPath) {
		if err := l.rotateLog(logPath); err != nil {
			return err
		}
	}

	// 使用更安全的文件权限 LogFilePerm（仅所有者可读写）
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, LogFilePerm)
	if err != nil {
		return err
	}

	// 关闭旧文件
	if l.file != nil {
		l.file.Close()
	}

	l.file = file
	log.SetOutput(file)
	return nil
}

// needsRotation 检查是否需要日志轮转
func (l *Logger) needsRotation(logPath string) bool {
	info, err := os.Stat(logPath)
	if err != nil {
		return false // 文件不存在，不需要轮转
	}
	return info.Size() > l.maxSize*1024*1024
}

// rotateLog 执行日志轮转
func (l *Logger) rotateLog(logPath string) error {
	// 重命名当前日志文件
	timestamp := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("%s.%s", logPath, timestamp)

	if err := os.Rename(logPath, backupPath); err != nil {
		return err
	}

	// 清理旧备份
	return l.cleanupOldBackups(logPath)
}

// cleanupOldBackups 清理旧备份文件
func (l *Logger) cleanupOldBackups(logPath string) error {
	pattern := fmt.Sprintf("%s.*", logPath)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	// 如果备份文件过多，删除最旧的
	if len(matches) > l.maxBackups {
		// 按修改时间排序
		for i := 0; i < len(matches)-l.maxBackups; i++ {
			if err := os.Remove(matches[i]); err != nil {
				// 记录错误但不停止处理
				fmt.Fprintf(os.Stderr, "无法删除旧日志文件: %v\n", err)
			}
		}
	}

	return nil
}

// getLogDir 获取日志目录
func getLogDir() string {
	// 优先使用环境变量
	if logDir := os.Getenv("DELGUARD_LOG_DIR"); logDir != "" {
		return logDir
	}

	// 使用用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "delguard")
	}

	return filepath.Join(homeDir, ".delguard", "logs")
}

// parseLogLevel 解析日志级别字符串
func parseLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case LogLevelDebugStr:
		return LogLevelDebug
	case LogLevelInfoStr:
		return LogLevelInfo
	case LogLevelWarnStr, "warning":
		return LogLevelWarn
	case LogLevelErrorStr:
		return LogLevelError
	case LogLevelFatalStr:
		return LogLevelFatal
	default:
		return LogLevelInfo
	}
}

// Log 记录日志
func (l *Logger) Log(level LogLevel, operation, filePath string, err error, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	// 确保日志文件打开
	if l.file == nil {
		if setupErr := l.setupLogFile(); setupErr != nil {
			fmt.Fprintf(os.Stderr, "无法设置日志文件: %v\n", setupErr)
			return
		}
	}

	// 构建日志消息
	logEntry := l.buildLogEntry(level, operation, filePath, err, message)

	// 写入日志
	if _, writeErr := l.file.WriteString(logEntry); writeErr != nil {
		// 如果写入失败，尝试重新设置日志文件
		if setupErr := l.setupLogFile(); setupErr == nil {
			l.file.WriteString(logEntry)
		}
	}

	// 同时输出到控制台（错误级别）
	if level >= LogLevelError {
		fmt.Fprintf(os.Stderr, "[%s] %s\n", level.String(), message)
	}
}

// buildLogEntry 构建日志条目
func (l *Logger) buildLogEntry(level LogLevel, operation, filePath string, err error, message string) string {
	// 时间戳
	timestamp := time.Now().Format(TimeFormatWithMillis)

	// 调用信息（文件名和行号）
	_, file, line, ok := runtime.Caller(3)
	caller := "unknown"
	if ok {
		caller = fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}

	// 错误信息（脱敏处理）
	errorMsg := ""
	if err != nil {
		// 脱敏错误信息
		sanitizedErr := sanitizeErrorMessage(err.Error())
		errorMsg = fmt.Sprintf(" | Error: %s", sanitizedErr)
	}

	// 文件路径（脱敏处理）
	displayPath := sanitizeFilePath(filePath)

	// 消息脱敏处理
	sanitizedMessage := sanitizeMessage(message)

	return fmt.Sprintf("[%s] [%s] [%s] %s - %s - %s%s\n",
		timestamp, level.String(), operation, caller, displayPath, sanitizedMessage, errorMsg)
}

// String 返回日志级别的字符串表示
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "调试"
	case LogLevelInfo:
		return "信息"
	case LogLevelWarn:
		return "警告"
	case LogLevelError:
		return "错误"
	case LogLevelFatal:
		return "致命"
	default:
		return "未知"
	}
}

// 快捷日志方法
func (l *Logger) Debug(operation, filePath string, message string) {
	l.Log(LogLevelDebug, operation, filePath, nil, message)
}

func (l *Logger) Info(operation, filePath string, message string) {
	l.Log(LogLevelInfo, operation, filePath, nil, message)
}

func (l *Logger) Warn(operation, filePath string, message string) {
	l.Log(LogLevelWarn, operation, filePath, nil, message)
}

func (l *Logger) Error(operation, filePath string, err error, message string) {
	l.Log(LogLevelError, operation, filePath, err, message)
}

// Fatal 支持可测试的退出处理器
func (l *Logger) Fatal(operation, filePath string, err error, message string) {
	l.Log(LogLevelFatal, operation, filePath, err, message)
	if ExitHandler != nil {
		ExitHandler(1)
	} else {
		os.Exit(1)
	}
}

// Close 关闭日志文件
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// LogError 全局错误日志记录函数
func LogError(operation, filePath string, err error) {
	if logger != nil {
		logger.Error(operation, filePath, err, "操作失败")
	}
}

// LogInfo 全局信息日志记录函数
func LogInfo(operation, filePath string, message string) {
	if logger != nil {
		logger.Info(operation, filePath, message)
	}
}

// LogDebug 全局调试日志记录函数
func LogDebug(operation, filePath string, message string) {
	if logger != nil {
		logger.Debug(operation, filePath, message)
	}
}

// LogWarn 全局警告日志记录函数
func LogWarn(operation, filePath string, message string) {
	if logger != nil {
		logger.Warn(operation, filePath, message)
	}
}

// sanitizeFilePath 脱敏文件路径，隐藏敏感信息
func sanitizeFilePath(filePath string) string {
	if filePath == "" {
		return ""
	}

	// 获取用户主目录，用于脱敏
	homeDir, _ := os.UserHomeDir()

	// 替换用户主目录为~
	if homeDir != "" && strings.HasPrefix(filePath, homeDir) {
		filePath = "~" + strings.TrimPrefix(filePath, homeDir)
	}

	// 缩短过长的路径
	if len(filePath) > 50 {
		filePath = "..." + filePath[len(filePath)-47:]
	}

	return filePath
}

// sanitizeErrorMessage 脱敏错误信息
func sanitizeErrorMessage(errorMsg string) string {
	// 移除绝对路径信息
	if homeDir, err := os.UserHomeDir(); err == nil && homeDir != "" {
		errorMsg = strings.ReplaceAll(errorMsg, homeDir, "~")
	}

	// 移除Windows用户目录模式
	if runtime.GOOS == "windows" {
		systemDrive := os.Getenv("SYSTEMDRIVE")
		if systemDrive == "" {
			systemDrive = "C:"
		}
		usersPath := filepath.Join(systemDrive, "Users") + string(filepath.Separator)
		errorMsg = strings.ReplaceAll(errorMsg, usersPath, "<USER_DIR>"+string(filepath.Separator))
	}

	// 限制错误信息长度
	if len(errorMsg) > 200 {
		errorMsg = errorMsg[:197] + "..."
	}

	return errorMsg
}

// sanitizeMessage 脱敏日志消息
func sanitizeMessage(message string) string {
	// 移除用户目录信息
	if homeDir, err := os.UserHomeDir(); err == nil && homeDir != "" {
		message = strings.ReplaceAll(message, homeDir, "~")
	}

	// 移除系统路径信息
	if runtime.GOOS == "windows" {
		systemDrive := os.Getenv("SYSTEMDRIVE")
		if systemDrive == "" {
			systemDrive = "C:"
		}
		windowsPath := filepath.Join(systemDrive, "Windows") + string(filepath.Separator)
		programFilesPath := filepath.Join(systemDrive, "Program Files") + string(filepath.Separator)
		message = strings.ReplaceAll(message, windowsPath, "<WINDOWS_DIR>"+string(filepath.Separator))
		message = strings.ReplaceAll(message, programFilesPath, "<PROGRAM_FILES>"+string(filepath.Separator))
	}

	// 限制消息长度
	if len(message) > 500 {
		message = message[:497] + "..."
	}

	return message
}
