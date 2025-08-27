package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"delguard/utils"
)

// FeedbackLevel 反馈级别
type FeedbackLevel int

const (
	LevelMinimal FeedbackLevel = iota // 最简反馈
	LevelNormal                       // 普通反馈
	LevelVerbose                      // 详细反馈
	LevelDebug                        // 调试反馈
)

// FeedbackManager 增强的用户反馈管理器
type FeedbackManager struct {
	level          FeedbackLevel
	colorEnabled   bool
	progressStyle  string
	confirmStyle   string
	statisticsData *StatisticsData
	mu             sync.RWMutex
}

// StatisticsData 统计数据
type StatisticsData struct {
	OperationsTotal   int64            `json:"operations_total"`
	OperationsSuccess int64            `json:"operations_success"`
	OperationsFailed  int64            `json:"operations_failed"`
	FilesProcessed    int64            `json:"files_processed"`
	BytesProcessed    int64            `json:"bytes_processed"`
	TimeElapsed       time.Duration    `json:"time_elapsed"`
	AverageSpeed      float64          `json:"average_speed_mb_s"`
	ErrorTypes        map[string]int64 `json:"error_types"`
	LastOperation     time.Time        `json:"last_operation"`
	SessionStartTime  time.Time        `json:"session_start_time"`
}

// FeedbackMessage 反馈消息结构
type FeedbackMessage struct {
	Type        string                 `json:"type"`
	Level       FeedbackLevel          `json:"level"`
	Message     string                 `json:"message"`
	Context     map[string]interface{} `json:"context"`
	Timestamp   time.Time              `json:"timestamp"`
	Suggestions []string               `json:"suggestions"`
	Actions     []string               `json:"actions"`
}

// Colors 颜色常量
var Colors = struct {
	Reset  string
	Red    string
	Green  string
	Yellow string
	Blue   string
	Purple string
	Cyan   string
	Gray   string
	White  string
	Bold   string
}{
	Reset:  "\033[0m",
	Red:    "\033[31m",
	Green:  "\033[32m",
	Yellow: "\033[33m",
	Blue:   "\033[34m",
	Purple: "\033[35m",
	Cyan:   "\033[36m",
	Gray:   "\033[37m",
	White:  "\033[97m",
	Bold:   "\033[1m",
}

// Icons 图标常量
var Icons = struct {
	Success    string
	Error      string
	Warning    string
	Info       string
	Question   string
	Processing string
	Completed  string
	Cancelled  string
	File       string
	Folder     string
	Link       string
	Hidden     string
	System     string
	ReadOnly   string
	Executable string
	Archive    string
	Image      string
	Video      string
	Audio      string
	Document   string
	Code       string
}{
	Success:    "✅",
	Error:      "❌",
	Warning:    "⚠️",
	Info:       "ℹ️",
	Question:   "❓",
	Processing: "⏳",
	Completed:  "🎉",
	Cancelled:  "🚫",
	File:       "📄",
	Folder:     "📁",
	Link:       "🔗",
	Hidden:     "👻",
	System:     "⚙️",
	ReadOnly:   "🔒",
	Executable: "⚡",
	Archive:    "📦",
	Image:      "🖼️",
	Video:      "🎬",
	Audio:      "🎵",
	Document:   "📝",
	Code:       "💻",
}

// NewFeedbackManager 创建增强的反馈管理器
func NewFeedbackManager() *FeedbackManager {
	return &FeedbackManager{
		level:          LevelNormal,
		colorEnabled:   true,
		progressStyle:  "bar",
		confirmStyle:   "interactive",
		statisticsData: NewStatisticsData(),
	}
}

// NewStatisticsData 创建新的统计数据
func NewStatisticsData() *StatisticsData {
	return &StatisticsData{
		ErrorTypes:       make(map[string]int64),
		SessionStartTime: time.Now(),
	}
}

// SetLevel 设置反馈级别
func (fm *FeedbackManager) SetLevel(level FeedbackLevel) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.level = level
}

// SetColorEnabled 设置是否启用颜色
func (fm *FeedbackManager) SetColorEnabled(enabled bool) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.colorEnabled = enabled
}

// SetProgressStyle 设置进度条样式
func (fm *FeedbackManager) SetProgressStyle(style string) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.progressStyle = style
}

// colorize 应用颜色
func (fm *FeedbackManager) colorize(text, color string) string {
	if !fm.colorEnabled {
		return text
	}
	return color + text + Colors.Reset
}

// ShowSuccess 显示成功消息
func (fm *FeedbackManager) ShowSuccess(message string, context ...map[string]interface{}) {
	fm.show(FeedbackMessage{
		Type:      "success",
		Level:     LevelNormal,
		Message:   message,
		Context:   fm.mergeContext(context...),
		Timestamp: time.Now(),
	})
	fm.updateStatistics("success", 1)
}

// ShowError 显示错误消息
func (fm *FeedbackManager) ShowError(err error, suggestions ...string) {
	message := err.Error()
	errorType := "generic"

	// 智能错误分类
	if strings.Contains(message, "permission") || strings.Contains(message, "权限") {
		errorType = "permission"
		suggestions = append(suggestions, T("尝试以管理员身份运行"))
	} else if strings.Contains(message, "not found") || strings.Contains(message, "不存在") {
		errorType = "not_found"
		suggestions = append(suggestions, T("检查文件路径是否正确"))
		suggestions = append(suggestions, T("使用智能搜索功能"))
	} else if strings.Contains(message, "in use") || strings.Contains(message, "被使用") {
		errorType = "file_in_use"
		suggestions = append(suggestions, T("关闭使用该文件的程序"))
		suggestions = append(suggestions, T("等待一段时间后重试"))
	}

	fm.show(FeedbackMessage{
		Type:        "error",
		Level:       LevelNormal,
		Message:     message,
		Timestamp:   time.Now(),
		Suggestions: suggestions,
	})
	fm.updateStatistics("error", 1)
	fm.updateErrorType(errorType)
}

// ShowWarning 显示警告消息
func (fm *FeedbackManager) ShowWarning(message string, suggestions ...string) {
	fm.show(FeedbackMessage{
		Type:        "warning",
		Level:       LevelNormal,
		Message:     message,
		Timestamp:   time.Now(),
		Suggestions: suggestions,
	})
}

// ShowInfo 显示信息消息
func (fm *FeedbackManager) ShowInfo(message string, level FeedbackLevel) {
	if fm.level >= level {
		fm.show(FeedbackMessage{
			Type:      "info",
			Level:     level,
			Message:   message,
			Timestamp: time.Now(),
		})
	}
}

// ShowProgress 显示增强的进度条
func (fm *FeedbackManager) ShowProgress(current, total int, message string, details map[string]interface{}) {
	if fm.level == LevelMinimal {
		return
	}

	percentage := float64(current) / float64(total) * 100

	switch fm.progressStyle {
	case "bar":
		fm.showProgressBar(current, total, percentage, message)
	case "spinner":
		fm.showProgressSpinner(current, total, percentage, message)
	case "dots":
		fm.showProgressDots(current, total, percentage, message)
	case "detailed":
		fm.showProgressDetailed(current, total, percentage, message, details)
	default:
		fm.showProgressBar(current, total, percentage, message)
	}
}

// showProgressBar 显示进度条
func (fm *FeedbackManager) showProgressBar(current, total int, percentage float64, message string) {
	barLength := 50
	filledLength := int(percentage / 100 * float64(barLength))

	bar := strings.Repeat("█", filledLength) + strings.Repeat("░", barLength-filledLength)
	coloredBar := fm.colorize(bar, Colors.Blue)

	fmt.Printf("\r%s [%s] %.1f%% (%d/%d) %s",
		Icons.Processing, coloredBar, percentage, current, total, message)

	if current == total {
		fmt.Printf(" %s\n", Icons.Completed)
	}
}

// showProgressSpinner 显示旋转器进度
func (fm *FeedbackManager) showProgressSpinner(current, total int, percentage float64, message string) {
	spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinner := spinners[current%len(spinners)]

	fmt.Printf("\r%s %.1f%% (%d/%d) %s", spinner, percentage, current, total, message)

	if current == total {
		fmt.Printf(" %s\n", Icons.Completed)
	}
}

// showProgressDots 显示点状进度
func (fm *FeedbackManager) showProgressDots(current, total int, percentage float64, message string) {
	dots := int(percentage / 10)
	progress := strings.Repeat("●", dots) + strings.Repeat("○", 10-dots)

	fmt.Printf("\r[%s] %.1f%% %s", progress, percentage, message)

	if current == total {
		fmt.Printf(" %s\n", Icons.Completed)
	}
}

func (fm *FeedbackManager) showProgressDetailed(current, total int, percentage float64, message string, details map[string]interface{}) {
	speed := ""
	eta := ""

	if details != nil {
		if s, ok := details["speed"]; ok {
			speed = fmt.Sprintf(T(" | 速度: %v"), s)
		}
		if e, ok := details["eta"]; ok {
			eta = fmt.Sprintf(T(" | ETA: %v"), e)
		}
	}

	fmt.Printf("\r%s %.1f%% (%d/%d)%s%s | %s",
		Icons.Processing, percentage, current, total, speed, eta, message)

	if current == total {
		fmt.Printf(" %s\n", Icons.Completed)
	}
}

// ConfirmAction 增强的确认对话框
func (fm *FeedbackManager) ConfirmAction(message string, options ConfirmOptions) (ConfirmResult, error) {
	switch fm.confirmStyle {
	case "simple":
		return fm.confirmSimple(message)
	case "detailed":
		return fm.confirmDetailed(message, options)
	case "interactive":
		return fm.confirmInteractive(message, options)
	default:
		return fm.confirmInteractive(message, options)
	}
}

// ConfirmOptions 确认选项
type ConfirmOptions struct {
	Default     string                 `json:"default"`
	Timeout     time.Duration          `json:"timeout"`
	ShowHelp    bool                   `json:"show_help"`
	Context     map[string]interface{} `json:"context"`
	Suggestions []string               `json:"suggestions"`
	Risks       []string               `json:"risks"`
	Benefits    []string               `json:"benefits"`
}

// ConfirmResult 确认结果
type ConfirmResult struct {
	Action   string                 `json:"action"`
	Metadata map[string]interface{} `json:"metadata"`
}

// confirmSimple 简单确认
func (fm *FeedbackManager) confirmSimple(message string) (ConfirmResult, error) {
	fmt.Printf("%s %s [y/N]: ", Icons.Question, message)
	// 非交互或超时则默认 no，避免阻塞
	var input string
	if isStdinInteractive() {
		if s, ok := readLineWithTimeout(20 * time.Second); ok {
			input = strings.ToLower(strings.TrimSpace(s))
		} else {
			input = ""
		}
	} else {
		input = ""
	}
	if input == "y" || input == "yes" {
		return ConfirmResult{Action: "yes"}, nil
	}

	return ConfirmResult{Action: "no"}, nil
}

// confirmDetailed 详细确认
func (fm *FeedbackManager) confirmDetailed(message string, options ConfirmOptions) (ConfirmResult, error) {
	fmt.Printf("\n%s %s\n", Icons.Question, fm.colorize(message, Colors.Bold))

	if len(options.Risks) > 0 {
		fmt.Printf("\n%s %s\n", Icons.Warning, T("风险:"))
		for _, risk := range options.Risks {
			fmt.Printf("  • %s\n", fm.colorize(risk, Colors.Red))
		}
	}

	if len(options.Benefits) > 0 {
		fmt.Printf("\n%s %s\n", Icons.Info, T("好处:"))
		for _, benefit := range options.Benefits {
			fmt.Printf("  • %s\n", fm.colorize(benefit, Colors.Green))
		}
	}

	if len(options.Suggestions) > 0 {
		fmt.Printf("\n%s %s\n", Icons.Info, T("建议:"))
		for _, suggestion := range options.Suggestions {
			fmt.Printf("  • %s\n", fm.colorize(suggestion, Colors.Cyan))
		}
	}

	fmt.Printf("\n%s", T("选择:"))
	fmt.Printf(" %s", T("[y]是"))
	fmt.Printf(" %s", T("[n]否"))
	if options.Default != "" {
		fmt.Printf(T(" (默认: %s)"), options.Default)
	}
	fmt.Print(": ")

	var input string
	var timeout time.Duration
	if options.Timeout > 0 {
		timeout = options.Timeout
	} else {
		timeout = 30 * time.Second
	}
	if isStdinInteractive() {
		if s, ok := readLineWithTimeout(timeout); ok {
			input = strings.ToLower(strings.TrimSpace(s))
		} else {
			input = ""
		}
	} else {
		input = ""
	}
	if input == "" && options.Default != "" {
		input = options.Default
	}

	if input == "y" || input == "yes" {
		return ConfirmResult{Action: "yes"}, nil
	}

	return ConfirmResult{Action: "no"}, nil
}

// confirmInteractive 交互式确认
func (fm *FeedbackManager) confirmInteractive(message string, options ConfirmOptions) (ConfirmResult, error) {
	for {
		fmt.Printf("\n%s %s\n", Icons.Question, fm.colorize(message, Colors.Bold))
		fmt.Printf(T("\n选择:\n"))
		fmt.Printf(T("  [y] 是\n"))
		fmt.Printf(T("  [n] 否\n"))
		fmt.Printf(T("  [h] 帮助\n"))
		fmt.Printf(T("  [d] 详细信息\n"))
		fmt.Printf(T("  [q] 退出\n"))

		if options.Default != "" {
			fmt.Printf(T("\n默认选择: %s\n"), options.Default)
		}

		fmt.Print(T("\n请选择: "))
		var input string
		var timeout time.Duration
		if options.Timeout > 0 {
			timeout = options.Timeout
		} else {
			timeout = 30 * time.Second
		}
		if isStdinInteractive() {
			if s, ok := readLineWithTimeout(timeout); ok {
				input = strings.ToLower(strings.TrimSpace(s))
			} else {
				input = ""
			}
		} else {
			input = ""
		}
		if input == "" && options.Default != "" {
			input = options.Default
		}

		switch input {
		case "y", "yes":
			return ConfirmResult{Action: "yes"}, nil
		case "n", "no":
			return ConfirmResult{Action: "no"}, nil
		case "h", "help":
			fm.showHelp()
		case "d", "detail":
			fm.showDetailedInfo(options)
		case "q", "quit":
			return ConfirmResult{Action: "quit"}, nil
		default:
			fmt.Printf("%s %s\n", Icons.Error, T("无效选择，请重新输入"))
		}
	}
}

// showHelp 显示帮助信息
func (fm *FeedbackManager) showHelp() {
	fmt.Printf("\n%s %s\n", Icons.Info, T("帮助信息:"))
	fmt.Printf(T("  y/yes - 确认执行操作\n"))
	fmt.Printf(T("  n/no  - 取消操作\n"))
	fmt.Printf(T("  h     - 显示此帮助\n"))
	fmt.Printf(T("  d     - 显示详细信息\n"))
	fmt.Printf(T("  q     - 退出程序\n"))
}

// showDetailedInfo 显示详细信息
func (fm *FeedbackManager) showDetailedInfo(options ConfirmOptions) {
	fmt.Printf("\n%s %s\n", Icons.Info, T("详细信息:"))

	if options.Context != nil {
		for key, value := range options.Context {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}
}

// ShowStatistics 显示统计信息
func (fm *FeedbackManager) ShowStatistics() {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	stats := fm.statisticsData

	fmt.Printf("\n%s %s\n", Icons.Info, T("会话统计信息:"))
	fmt.Printf("──────────────────────────────\n")
	fmt.Printf("%s %d\n", T("总操作数:"), stats.OperationsTotal)
	fmt.Printf("%s %s\n", T("成功操作:"), fm.colorize(fmt.Sprintf("%d", stats.OperationsSuccess), Colors.Green))
	fmt.Printf("%s %s\n", T("失败操作:"), fm.colorize(fmt.Sprintf("%d", stats.OperationsFailed), Colors.Red))
	fmt.Printf("%s %d\n", T("处理文件:"), stats.FilesProcessed)
	fmt.Printf("%s %s\n", T("处理字节:"), fm.formatBytes(stats.BytesProcessed))
	fmt.Printf("%s %v\n", T("会话时长:"), time.Since(stats.SessionStartTime).Round(time.Second))
	fmt.Printf("%s %.2f MB/s\n", T("平均速度:"), stats.AverageSpeed)

	if len(stats.ErrorTypes) > 0 {
		logger.Info("错误类型统计", "message", "错误类型分布")
		for errorType, count := range stats.ErrorTypes {
			logger.Info("错误统计详情", fmt.Sprintf("类型: %s, 数量: %d", errorType, count), "stats")
		}
	}

	fmt.Printf("──────────────────────────────\n")
}

// formatBytes 格式化字节数
func (fm *FeedbackManager) formatBytes(bytes int64) string {
	return utils.FormatBytes(bytes)
}

// show 显示反馈消息
func (fm *FeedbackManager) show(msg FeedbackMessage) {
	if fm.level < msg.Level {
		return
	}

	var icon, color string
	switch msg.Type {
	case "success":
		icon, color = Icons.Success, Colors.Green
	case "error":
		icon, color = Icons.Error, Colors.Red
	case "warning":
		icon, color = Icons.Warning, Colors.Yellow
	case "info":
		icon, color = Icons.Info, Colors.Blue
	default:
		icon, color = Icons.Info, Colors.White
	}

	output := fmt.Sprintf("%s %s", icon, fm.colorize(msg.Message, color))

	if fm.level >= LevelVerbose && len(msg.Suggestions) > 0 {
		output += "\n" + T("💡 建议:")
		for _, suggestion := range msg.Suggestions {
			output += fmt.Sprintf("\n   • %s", suggestion)
		}
	}

	fmt.Println(output)
}

// updateStatistics 更新统计信息
func (fm *FeedbackManager) updateStatistics(operation string, count int64) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.statisticsData.OperationsTotal += count
	if operation == "success" {
		fm.statisticsData.OperationsSuccess += count
	} else if operation == "error" {
		fm.statisticsData.OperationsFailed += count
	}
	fm.statisticsData.LastOperation = time.Now()
}

// updateErrorType 更新错误类型统计
func (fm *FeedbackManager) updateErrorType(errorType string) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if fm.statisticsData.ErrorTypes == nil {
		fm.statisticsData.ErrorTypes = make(map[string]int64)
	}
	fm.statisticsData.ErrorTypes[errorType]++
}

// mergeContext 合并上下文
func (fm *FeedbackManager) mergeContext(contexts ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, ctx := range contexts {
		for k, v := range ctx {
			result[k] = v
		}
	}
	return result
}

// GetFileIcon 获取文件图标
func (fm *FeedbackManager) GetFileIcon(filePath string, isDir bool) string {
	if isDir {
		return Icons.Folder
	}

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filePath), "."))

	switch ext {
	case "jpg", "jpeg", "png", "gif", "bmp", "svg", "webp":
		return Icons.Image
	case "mp4", "avi", "mov", "wmv", "flv", "mkv":
		return Icons.Video
	case "mp3", "wav", "flac", "aac", "ogg":
		return Icons.Audio
	case "doc", "docx", "pdf", "txt", "rtf":
		return Icons.Document
	case "zip", "rar", "7z", "tar", "gz":
		return Icons.Archive
	case "exe", "msi", "app", "deb", "rpm":
		return Icons.Executable
	case "go", "py", "js", "html", "css", "java", "c", "cpp":
		return Icons.Code
	default:
		return Icons.File
	}
}

// ShowFileInfo 显示文件信息
func (fm *FeedbackManager) ShowFileInfo(filePath string, fileInfo os.FileInfo) {
	icon := fm.GetFileIcon(filePath, fileInfo.IsDir())
	name := fileInfo.Name()
	size := fm.formatBytes(fileInfo.Size())
	modTime := fileInfo.ModTime().Format(TimeFormatStandard)

	var attrs []string
	if fileInfo.IsDir() {
		attrs = append(attrs, T("目录"))
	}
	if fileInfo.Mode()&os.ModeSymlink != 0 {
		attrs = append(attrs, T("符号链接"))
	}
	if strings.HasPrefix(name, ".") {
		attrs = append(attrs, T("隐藏"))
	}
	if fileInfo.Mode()&0200 == 0 {
		attrs = append(attrs, T("只读"))
	}

	attrStr := ""
	if len(attrs) > 0 {
		attrStr = fmt.Sprintf(" [%s]", strings.Join(attrs, ", "))
	}

	fmt.Printf("%s %s%s - %s - %s\n", icon, name, attrStr, size, modTime)
}
