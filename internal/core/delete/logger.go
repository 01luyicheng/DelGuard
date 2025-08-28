package delete

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// String 返回日志级别的字符串表示
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger 日志记录器
type Logger struct {
	level      LogLevel
	output     io.Writer
	mu         sync.Mutex
	prefix     string
	timeFormat string
	logger     *log.Logger
}

// NewLogger 创建新的日志记录器
func NewLogger(level LogLevel, output io.Writer) *Logger {
	if output == nil {
		output = os.Stdout
	}
	
	return &Logger{
		level:      level,
		output:     output,
		prefix:     "[DelGuard] ",
		timeFormat: "2006-01-02 15:04:05",
		logger:     log.New(output, "", 0),
	}
}

// NewFileLogger 创建文件日志记录器
func NewFileLogger(level LogLevel, filePath string) (*Logger, error) {
	// 确保日志目录存在
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %v", err)
	}
	
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("打开日志文件失败: %v", err)
	}
	
	return NewLogger(level, file), nil
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetPrefix 设置日志前缀
func (l *Logger) SetPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
}

// log 记录日志
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if level < l.level {
		return
	}
	
	// 获取调用者信息
	_, file, line, ok := runtime.Caller(2)
	caller := ""
	if ok {
		caller = fmt.Sprintf(" %s:%d", filepath.Base(file), line)
	}
	
	// 格式化消息
	message := fmt.Sprintf(format, args...)
	
	// 构建日志条目
	timestamp := time.Now().Format(l.timeFormat)
	logEntry := fmt.Sprintf("%s%s [%s]%s %s\n", 
		l.prefix, timestamp, level.String(), caller, message)
	
	l.logger.Print(logEntry)
}

// Debug 记录调试日志
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LogLevelDebug, format, args...)
}

// Info 记录信息日志
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LogLevelInfo, format, args...)
}

// Warn 记录警告日志
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LogLevelWarn, format, args...)
}

// Error 记录错误日志
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LogLevelError, format, args...)
}

// Fatal 记录致命错误日志并退出程序
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(LogLevelFatal, format, args...)
	os.Exit(1)
}

// LogDeleteOperation 记录删除操作
func (l *Logger) LogDeleteOperation(path string, success bool, err error) {
	if success {
		l.Info("删除成功: %s", path)
	} else {
		l.Error("删除失败: %s, 错误: %v", path, err)
	}
}

// LogBatchDeleteOperation 记录批量删除操作
func (l *Logger) LogBatchDeleteOperation(results []DeleteResult) {
	successful := 0
	failed := 0
	
	for _, result := range results {
		if result.Success {
			successful++
			l.Debug("批量删除成功: %s", result.Path)
		} else {
			failed++
			l.Error("批量删除失败: %s, 错误: %v", result.Path, result.Error)
		}
	}
	
	l.Info("批量删除完成: 成功 %d, 失败 %d, 总计 %d", successful, failed, len(results))
}

// Close 关闭日志记录器
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if closer, ok := l.output.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// DefaultLogger 默认日志记录器
var DefaultLogger = NewLogger(LogLevelInfo, os.Stdout)

// 全局日志函数
func Debug(format string, args ...interface{}) {
	DefaultLogger.Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	DefaultLogger.Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	DefaultLogger.Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	DefaultLogger.Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	DefaultLogger.Fatal(format, args...)
}