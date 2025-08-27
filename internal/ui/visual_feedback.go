package ui

import (
	"fmt"
	"strings"
)

// 颜色代码常量
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

// VisualFeedback 提供增强的视觉反馈功能
type VisualFeedback struct {
	UseColors      bool   // 是否使用颜色
	UseUnicode     bool   // 是否使用Unicode符号
	IndentSize     int    // 缩进大小
	ProgressFormat string // 进度条格式
}

// NewVisualFeedback 创建一个新的视觉反馈实例
func NewVisualFeedback() *VisualFeedback {
	return &VisualFeedback{
		UseColors:      true,
		UseUnicode:     true,
		IndentSize:     2,
		ProgressFormat: "[{bar}] {percent}%",
	}
}

// ShowSuccess 显示成功消息
func (vf *VisualFeedback) ShowSuccess(message string) {
	vf.showColoredMessage(message, ColorGreen, "✓ ")
}

// ShowError 显示错误消息
func (vf *VisualFeedback) ShowError(message string) {
	vf.showColoredMessage(message, ColorRed, "✗ ")
}

// ShowWarning 显示警告消息
func (vf *VisualFeedback) ShowWarning(message string) {
	vf.showColoredMessage(message, ColorYellow, "⚠ ")
}

// ShowInfo 显示信息消息
func (vf *VisualFeedback) ShowInfo(message string) {
	vf.showColoredMessage(message, ColorCyan, "ℹ ")
}

// ShowPrompt 显示提示消息
func (vf *VisualFeedback) ShowPrompt(message string) {
	vf.showColoredMessage(message, ColorPurple, "? ")
}

// ShowListItem 显示列表项
func (vf *VisualFeedback) ShowListItem(message string) {
	indent := strings.Repeat(" ", vf.IndentSize)
	vf.showColoredMessage(indent+message, ColorWhite, "• ")
}

// ShowProgress 显示进度条
func (vf *VisualFeedback) ShowProgress(current, total int) {
	if total <= 0 {
		total = 1
	}

	percent := int(float64(current) / float64(total) * 100)
	width := 30
	completed := int(float64(width) * float64(current) / float64(total))

	bar := strings.Repeat("█", completed) + strings.Repeat("░", width-completed)

	progressText := vf.ProgressFormat
	progressText = strings.Replace(progressText, "{bar}", bar, -1)
	progressText = strings.Replace(progressText, "{percent}", fmt.Sprintf("%d", percent), -1)
	progressText = strings.Replace(progressText, "{current}", fmt.Sprintf("%d", current), -1)
	progressText = strings.Replace(progressText, "{total}", fmt.Sprintf("%d", total), -1)

	fmt.Printf("\r%s", progressText)

	if current >= total {
		fmt.Println()
	}
}

// ShowTable 显示表格数据
func (vf *VisualFeedback) ShowTable(headers []string, rows [][]string) {
	if len(headers) == 0 || len(rows) == 0 {
		return
	}

	// 计算每列的最大宽度
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// 打印表头
	fmt.Print("| ")
	for i, header := range headers {
		fmt.Printf("%-*s | ", colWidths[i], header)
	}
	fmt.Println()

	// 打印分隔线
	fmt.Print("| ")
	for i := range headers {
		fmt.Print(strings.Repeat("-", colWidths[i]), " | ")
	}
	fmt.Println()

	// 打印数据行
	for _, row := range rows {
		fmt.Print("| ")
		for i, cell := range row {
			if i < len(colWidths) {
				fmt.Printf("%-*s | ", colWidths[i], cell)
			}
		}
		fmt.Println()
	}
}

// showColoredMessage 显示带颜色的消息
func (vf *VisualFeedback) showColoredMessage(message, color, prefix string) {
	if vf.UseColors {
		if vf.UseUnicode {
			fmt.Printf("%s%s%s%s\n", color, prefix, message, ColorReset)
		} else {
			fmt.Printf("%s%s%s\n", color, message, ColorReset)
		}
	} else {
		if vf.UseUnicode {
			fmt.Printf("%s%s\n", prefix, message)
		} else {
			fmt.Println(message)
		}
	}
}
