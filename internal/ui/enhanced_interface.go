package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// EnhancedInterface 增强的用户界面
type EnhancedInterface struct {
	config interface{}
	locale string
}

// NewEnhancedInterface 创建增强的用户界面
func NewEnhancedInterface(config interface{}, locale string) *EnhancedInterface {
	return &EnhancedInterface{
		config: config,
		locale: locale,
	}
}

// ConfirmationDialog 确认对话框配置
type ConfirmationDialog struct {
	Title      string
	Message    string
	Details    []string
	DefaultYes bool
	Timeout    time.Duration
	ShowRisks  bool
	RiskLevel  string // "low", "medium", "high", "critical"
}

// ProgressIndicator 进度指示器
type ProgressIndicator struct {
	Title     string
	Current   int
	Total     int
	ShowETA   bool
	StartTime time.Time
}

// NotificationLevel 通知级别
type NotificationLevel int

const (
	NotificationInfo NotificationLevel = iota
	NotificationWarning
	NotificationError
	NotificationSuccess
)

// ShowConfirmationDialog 显示确认对话框
func (ui *EnhancedInterface) ShowConfirmationDialog(dialog ConfirmationDialog) (bool, error) {
	// 显示标题
	ui.printSeparator()
	fmt.Printf("🔔 %s\n", dialog.Title)
	ui.printSeparator()

	// 显示主要消息
	fmt.Printf("📝 %s\n\n", dialog.Message)

	// 显示详细信息
	if len(dialog.Details) > 0 {
		fmt.Println("📋 详细信息:")
		for i, detail := range dialog.Details {
			fmt.Printf("   %d. %s\n", i+1, detail)
		}
		fmt.Println()
	}

	// 显示风险提示
	if dialog.ShowRisks {
		ui.showRiskWarning(dialog.RiskLevel)
	}

	// 显示确认提示
	prompt := "确认继续吗?"
	if dialog.DefaultYes {
		prompt += " [Y/n]: "
	} else {
		prompt += " [y/N]: "
	}

	fmt.Print(prompt)

	// 读取用户输入
	input, ok := ui.readInputWithTimeout(dialog.Timeout)
	if !ok {
		fmt.Println("\n⏰ 操作超时，已自动取消")
		return false, nil
	}

	input = strings.ToLower(strings.TrimSpace(input))

	if dialog.DefaultYes {
		return input != "n" && input != "no", nil
	} else {
		return input == "y" || input == "yes", nil
	}
}

// ShowProgressIndicator 显示进度指示器
func (ui *EnhancedInterface) ShowProgressIndicator(progress ProgressIndicator) {
	if progress.Total == 0 {
		return
	}

	percentage := float64(progress.Current) / float64(progress.Total) * 100
	barLength := 40
	filledLength := int(percentage / 100 * float64(barLength))

	// 创建进度条
	bar := strings.Repeat("█", filledLength) + strings.Repeat("░", barLength-filledLength)

	// 计算预计剩余时间
	etaStr := ""
	if progress.ShowETA && progress.Current > 0 {
		elapsed := time.Since(progress.StartTime)
		avgTimePerItem := elapsed / time.Duration(progress.Current)
		remaining := avgTimePerItem * time.Duration(progress.Total-progress.Current)
		etaStr = fmt.Sprintf(" ETA: %s", ui.formatDuration(remaining))
	}

	// 显示进度
	fmt.Printf("\r🔄 %s [%s] %.1f%% (%d/%d)%s",
		progress.Title, bar, percentage, progress.Current, progress.Total, etaStr)

	if progress.Current == progress.Total {
		fmt.Println(" ✅ 完成!")
	}
}

// ShowNotification 显示通知消息
func (ui *EnhancedInterface) ShowNotification(level NotificationLevel, title, message string) {
	var icon, color string

	switch level {
	case NotificationInfo:
		icon = "ℹ️"
		color = "\033[36m" // 青色
	case NotificationWarning:
		icon = "⚠️"
		color = "\033[33m" // 黄色
	case NotificationError:
		icon = "❌"
		color = "\033[31m" // 红色
	case NotificationSuccess:
		icon = "✅"
		color = "\033[32m" // 绿色
	}

	reset := "\033[0m"

	if title != "" {
		fmt.Printf("%s%s %s%s\n", color, icon, title, reset)
	}

	if message != "" {
		// 多行消息处理
		lines := strings.Split(message, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				fmt.Printf("   %s\n", line)
			}
		}
	}
	fmt.Println()
}

// ShowOperationSummary 显示操作总结
func (ui *EnhancedInterface) ShowOperationSummary(operation string, successCount, failCount, skipCount int, duration time.Duration) {
	total := successCount + failCount + skipCount

	ui.printSeparator()
	fmt.Printf("📊 %s 操作总结\n", operation)
	ui.printSeparator()

	if successCount > 0 {
		fmt.Printf("✅ 成功: %d 个文件\n", successCount)
	}

	if failCount > 0 {
		fmt.Printf("❌ 失败: %d 个文件\n", failCount)
	}

	if skipCount > 0 {
		fmt.Printf("⏭️  跳过: %d 个文件\n", skipCount)
	}

	fmt.Printf("📈 总计: %d 个文件\n", total)
	fmt.Printf("⏱️  耗时: %s\n", ui.formatDuration(duration))

	if total > 0 {
		successRate := float64(successCount) / float64(total) * 100
		fmt.Printf("📊 成功率: %.1f%%\n", successRate)
	}

	ui.printSeparator()
}

// ShowFileList 显示文件列表（支持分页）
func (ui *EnhancedInterface) ShowFileList(title string, files []string, pageSize int) ([]int, error) {
	if len(files) == 0 {
		ui.ShowNotification(NotificationInfo, "提示", "没有找到任何文件")
		return nil, nil
	}

	totalPages := (len(files) + pageSize - 1) / pageSize
	currentPage := 1
	selectedIndices := make([]int, 0)

	for {
		// 显示当前页
		ui.printSeparator()
		fmt.Printf("📁 %s (第 %d/%d 页)\n", title, currentPage, totalPages)
		ui.printSeparator()

		start := (currentPage - 1) * pageSize
		end := start + pageSize
		if end > len(files) {
			end = len(files)
		}

		for i := start; i < end; i++ {
			marker := "  "
			if ui.contains(selectedIndices, i) {
				marker = "✓ "
			}
			fmt.Printf("%s[%d] %s\n", marker, i+1, files[i])
		}

		// 显示操作选项
		fmt.Println("\n📋 操作选项:")
		fmt.Println("  a - 全选当前页")
		fmt.Println("  c - 清除选择")
		fmt.Println("  数字 - 切换选择指定文件")
		if totalPages > 1 {
			if currentPage > 1 {
				fmt.Println("  p - 上一页")
			}
			if currentPage < totalPages {
				fmt.Println("  n - 下一页")
			}
		}
		fmt.Println("  q - 完成选择")
		fmt.Print("\n请选择操作: ")

		input, ok := ui.readInputWithTimeout(30 * time.Second)
		if !ok {
			return nil, fmt.Errorf("操作超时")
		}

		input = strings.ToLower(strings.TrimSpace(input))

		switch input {
		case "a":
			// 全选当前页
			for i := start; i < end; i++ {
				if !ui.contains(selectedIndices, i) {
					selectedIndices = append(selectedIndices, i)
				}
			}
		case "c":
			// 清除选择
			selectedIndices = selectedIndices[:0]
		case "p":
			if currentPage > 1 {
				currentPage--
			}
		case "n":
			if currentPage < totalPages {
				currentPage++
			}
		case "q":
			return selectedIndices, nil
		default:
			// 尝试解析为数字
			if num, err := strconv.Atoi(input); err == nil && num >= 1 && num <= len(files) {
				index := num - 1
				if ui.contains(selectedIndices, index) {
					// 取消选择
					selectedIndices = ui.removeIndex(selectedIndices, index)
				} else {
					// 添加选择
					selectedIndices = append(selectedIndices, index)
				}
			} else {
				ui.ShowNotification(NotificationWarning, "提示", "无效的选择，请重新输入")
			}
		}
	}
}

// ShowHelpTips 显示操作提示
func (ui *EnhancedInterface) ShowHelpTips(operation string) {
	tips := ui.getOperationTips(operation)
	if len(tips) == 0 {
		return
	}

	fmt.Println("💡 操作提示:")
	for _, tip := range tips {
		fmt.Printf("   • %s\n", tip)
	}
	fmt.Println()
}

// 私有辅助方法

func (ui *EnhancedInterface) printSeparator() {
	fmt.Println(strings.Repeat("─", 60))
}

func (ui *EnhancedInterface) showRiskWarning(riskLevel string) {
	switch riskLevel {
	case "critical":
		fmt.Println("🚨 危险级别: 极高")
		fmt.Println("⚠️  此操作可能导致系统不稳定或数据丢失")
		fmt.Println("🔒 建议: 请确保已备份重要数据")
	case "high":
		fmt.Println("⚠️  危险级别: 高")
		fmt.Println("📋 此操作可能影响系统功能")
		fmt.Println("💾 建议: 请谨慎操作并考虑备份")
	case "medium":
		fmt.Println("⚠️  危险级别: 中等")
		fmt.Println("📝 请确认操作的必要性")
	case "low":
		fmt.Println("ℹ️  危险级别: 低")
		fmt.Println("✅ 此操作相对安全")
	}
	fmt.Println()
}

func (ui *EnhancedInterface) readInputWithTimeout(timeout time.Duration) (string, bool) {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	done := make(chan string, 1)
	go func() {
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			done <- ""
		} else {
			done <- strings.TrimSpace(line)
		}
	}()

	select {
	case input := <-done:
		return input, true
	case <-time.After(timeout):
		return "", false
	}
}

func (ui *EnhancedInterface) formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1f秒", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.1f分钟", d.Minutes())
	} else {
		return fmt.Sprintf("%.1f小时", d.Hours())
	}
}

func (ui *EnhancedInterface) contains(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func (ui *EnhancedInterface) removeIndex(slice []int, index int) []int {
	result := make([]int, 0, len(slice))
	for _, v := range slice {
		if v != index {
			result = append(result, v)
		}
	}
	return result
}

func (ui *EnhancedInterface) getOperationTips(operation string) []string {
	switch operation {
	case "delete":
		return []string{
			"删除的文件将移动到回收站，可以恢复",
			"使用 --force 参数可以跳过确认",
			"使用 --dry-run 可以预览将要删除的文件",
			"系统重要文件会被自动保护",
		}
	case "search":
		return []string{
			"支持通配符 * 和 ? 进行模糊匹配",
			"使用 --content 可以搜索文件内容",
			"使用 --size 可以按文件大小过滤",
			"支持正则表达式搜索",
		}
	case "restore":
		return []string{
			"只能恢复通过 DelGuard 删除的文件",
			"恢复时会检查目标路径是否安全",
			"如果目标文件已存在，会提示是否覆盖",
		}
	default:
		return []string{}
	}
}
