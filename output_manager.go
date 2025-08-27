package main

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// OutputLevel 定义输出级别
type OutputLevel int

const (
	OutputLevelSilent OutputLevel = iota
	OutputLevelError
	OutputLevelWarn
	OutputLevelInfo
	OutputLevelDebug
	OutputLevelVerbose
)

// OutputManager 统一管理所有输出
type OutputManager struct {
	level        OutputLevel
	writer       io.Writer
	errorWriter  io.Writer
	mu           sync.RWMutex
	colorEnabled bool
}

// NewOutputManager 创建新的输出管理器
func NewOutputManager(level OutputLevel) *OutputManager {
	return &OutputManager{
		level:        level,
		writer:       os.Stdout,
		errorWriter:  os.Stderr,
		colorEnabled: true,
	}
}

// SetLevel 设置输出级别
func (om *OutputManager) SetLevel(level OutputLevel) {
	om.mu.Lock()
	defer om.mu.Unlock()
	om.level = level
}

// SetColorEnabled 设置是否启用颜色
func (om *OutputManager) SetColorEnabled(enabled bool) {
	om.mu.Lock()
	defer om.mu.Unlock()
	om.colorEnabled = enabled
}

// SetWriter 设置输出流
func (om *OutputManager) SetWriter(w io.Writer) {
	om.mu.Lock()
	defer om.mu.Unlock()
	om.writer = w
}

// SetErrorWriter 设置错误输出流
func (om *OutputManager) SetErrorWriter(w io.Writer) {
	om.mu.Lock()
	defer om.mu.Unlock()
	om.errorWriter = w
}

// colorize 为文本添加颜色
func (om *OutputManager) colorize(text, color string) string {
	if !om.colorEnabled {
		return text
	}

	colors := map[string]string{
		"red":     "\033[31m",
		"green":   "\033[32m",
		"yellow":  "\033[33m",
		"blue":    "\033[34m",
		"magenta": "\033[35m",
		"cyan":    "\033[36m",
		"white":   "\033[37m",
		"reset":   "\033[0m",
	}

	if colorCode, exists := colors[color]; exists {
		return colorCode + text + colors["reset"]
	}
	return text
}

// shouldOutput 检查是否应该输出
func (om *OutputManager) shouldOutput(level OutputLevel) bool {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return om.level >= level
}

// Error 输出错误信息
func (om *OutputManager) Error(format string, args ...interface{}) {
	if om.shouldOutput(OutputLevelError) {
		message := fmt.Sprintf(format, args...)
		coloredMessage := om.colorize("❌ "+message, "red")
		fmt.Fprintln(om.errorWriter, coloredMessage)
	}
}

// Warn 输出警告信息
func (om *OutputManager) Warn(format string, args ...interface{}) {
	if om.shouldOutput(OutputLevelWarn) {
		message := fmt.Sprintf(format, args...)
		coloredMessage := om.colorize("⚠️  "+message, "yellow")
		fmt.Fprintln(om.writer, coloredMessage)
	}
}

// Info 输出信息
func (om *OutputManager) Info(format string, args ...interface{}) {
	if om.shouldOutput(OutputLevelInfo) {
		message := fmt.Sprintf(format, args...)
		coloredMessage := om.colorize("ℹ️  "+message, "blue")
		fmt.Fprintln(om.writer, coloredMessage)
	}
}

// Success 输出成功信息
func (om *OutputManager) Success(format string, args ...interface{}) {
	if om.shouldOutput(OutputLevelInfo) {
		message := fmt.Sprintf(format, args...)
		coloredMessage := om.colorize("✅ "+message, "green")
		fmt.Fprintln(om.writer, coloredMessage)
	}
}

// Debug 输出调试信息
func (om *OutputManager) Debug(format string, args ...interface{}) {
	if om.shouldOutput(OutputLevelDebug) {
		message := fmt.Sprintf(format, args...)
		coloredMessage := om.colorize("🐛 "+message, "magenta")
		fmt.Fprintln(om.writer, coloredMessage)
	}
}

// Verbose 输出详细信息
func (om *OutputManager) Verbose(format string, args ...interface{}) {
	if om.shouldOutput(OutputLevelVerbose) {
		message := fmt.Sprintf(format, args...)
		fmt.Fprintln(om.writer, message)
	}
}

// Print 直接输出（不受级别限制）
func (om *OutputManager) Print(format string, args ...interface{}) {
	fmt.Fprintf(om.writer, format, args...)
}

// Println 直接输出一行（不受级别限制）
func (om *OutputManager) Println(args ...interface{}) {
	fmt.Fprintln(om.writer, args...)
}

// Printf 直接格式化输出（不受级别限制）
func (om *OutputManager) Printf(format string, args ...interface{}) {
	fmt.Fprintf(om.writer, format, args...)
}

// Progress 输出进度信息
func (om *OutputManager) Progress(current, total int, message string) {
	if om.shouldOutput(OutputLevelInfo) {
		percentage := float64(current) / float64(total) * 100
		progressBar := om.generateProgressBar(current, total, 20)
		coloredMessage := om.colorize(fmt.Sprintf("🔄 %s [%s] %.1f%% (%d/%d)",
			message, progressBar, percentage, current, total), "cyan")
		fmt.Fprintf(om.writer, "\r%s", coloredMessage)
		if current == total {
			fmt.Fprintln(om.writer)
		}
	}
}

// generateProgressBar 生成进度条
func (om *OutputManager) generateProgressBar(current, total, width int) string {
	if total == 0 {
		return ""
	}

	filled := int(float64(current) / float64(total) * float64(width))
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return bar
}

// 全局输出管理器实例
var globalOutputManager = NewOutputManager(OutputLevelInfo)

// 全局函数，方便使用
func SetOutputLevel(level OutputLevel) {
	globalOutputManager.SetLevel(level)
}

func SetColorEnabled(enabled bool) {
	globalOutputManager.SetColorEnabled(enabled)
}

func OutputError(format string, args ...interface{}) {
	globalOutputManager.Error(format, args...)
}

func OutputWarn(format string, args ...interface{}) {
	globalOutputManager.Warn(format, args...)
}

func OutputInfo(format string, args ...interface{}) {
	globalOutputManager.Info(format, args...)
}

func OutputSuccess(format string, args ...interface{}) {
	globalOutputManager.Success(format, args...)
}

func OutputDebug(format string, args ...interface{}) {
	globalOutputManager.Debug(format, args...)
}

func OutputVerbose(format string, args ...interface{}) {
	globalOutputManager.Verbose(format, args...)
}

func OutputProgress(current, total int, message string) {
	globalOutputManager.Progress(current, total, message)
}
