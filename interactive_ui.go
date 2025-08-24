package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// InteractiveUI 交互式用户界面
type InteractiveUI struct {
	reader *bufio.Reader
}

// NewInteractiveUI 创建新的交互式UI
func NewInteractiveUI() *InteractiveUI {
	return &InteractiveUI{
		reader: bufio.NewReader(os.Stdin),
	}
}

// ShowLoadingMessage 显示加载消息
func (ui *InteractiveUI) ShowLoadingMessage(message string) {
	fmt.Printf("🔍 %s", message)
	// 简单的加载动画
	for i := 0; i < 3; i++ {
		time.Sleep(500 * time.Millisecond)
		fmt.Print(".")
	}
	fmt.Println()
}

// ShowSearchResults 显示搜索结果并获取用户选择
func (ui *InteractiveUI) ShowSearchResults(results []SearchResult, target string) (string, error) {
	if len(results) == 0 {
		fmt.Printf("❌ 未找到与 '%s' 匹配的文件\n", target)
		return "", fmt.Errorf("未找到匹配的文件")
	}

	fmt.Printf("🔍 未找到文件 '%s'，找到以下相似文件：\n\n", target)

	// 显示搜索结果
	for i, result := range results {
		icon := ui.getMatchTypeIcon(result.MatchType)
		fmt.Printf("[%d] %s %s", i+1, icon, result.Path)

		if result.Similarity < 100.0 {
			fmt.Printf(" (相似度: %.1f%%)", result.Similarity)
		}

		if result.MatchType == "content" && result.Context != "" {
			fmt.Printf("\n    💡 匹配内容: %s", result.Context)
		}
		fmt.Println()
	}

	fmt.Printf("\n请选择文件编号 (1-%d)，或输入 'n' 取消操作: ", len(results))

	input, err := ui.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if strings.ToLower(input) == "n" {
		return "", fmt.Errorf("用户取消操作")
	}

	// 解析用户输入
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(results) {
		return "", fmt.Errorf("无效的选择")
	}

	return results[choice-1].Path, nil
}

// getMatchTypeIcon 获取匹配类型图标
func (ui *InteractiveUI) getMatchTypeIcon(matchType string) string {
	switch matchType {
	case "exact":
		return "✅"
	case "filename":
		return "📄"
	case "content":
		return "🔍"
	case "regex":
		return "🎯"
	default:
		return "📁"
	}
}

// ShowBatchConfirmation 显示批量操作确认
func (ui *InteractiveUI) ShowBatchConfirmation(files []string, operation string, force bool) (bool, error) {
	if force {
		return true, nil
	}

	if len(files) == 0 {
		fmt.Println("没有找到匹配的文件")
		return false, nil
	}

	fmt.Printf("⚠️  准备%s %d 个文件：\n\n", operation, len(files))

	// 显示文件列表（分页显示）
	pageSize := 10
	totalPages := (len(files) + pageSize - 1) / pageSize
	currentPage := 1

	for {
		// 显示当前页的文件
		start := (currentPage - 1) * pageSize
		end := start + pageSize
		if end > len(files) {
			end = len(files)
		}

		fmt.Printf("📄 第 %d/%d 页：\n", currentPage, totalPages)
		for i := start; i < end; i++ {
			fmt.Printf("  %d. %s\n", i+1, files[i])
		}

		// 显示操作选项
		fmt.Printf("\n🎯 选项：\n")
		fmt.Printf("  y - 确认%s所有文件\n", operation)
		fmt.Printf("  n - 取消操作\n")
		if totalPages > 1 {
			if currentPage < totalPages {
				fmt.Printf("  > - 下一页\n")
			}
			if currentPage > 1 {
				fmt.Printf("  < - 上一页\n")
			}
		}
		fmt.Printf("  s - 跳过确认（强制执行）\n")
		fmt.Print("\n请选择: ")

		input, err := ui.reader.ReadString('\n')
		if err != nil {
			return false, err
		}

		input = strings.ToLower(strings.TrimSpace(input))

		switch input {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		case "s", "skip":
			return true, nil
		case ">", "next":
			if currentPage < totalPages {
				currentPage++
			}
		case "<", "prev":
			if currentPage > 1 {
				currentPage--
			}
		default:
			fmt.Println("❌ 无效的选择，请重新输入")
		}

		fmt.Println() // 空行分隔
	}
}

// ShowProgressBar 显示进度条
func (ui *InteractiveUI) ShowProgressBar(current, total int, message string) {
	if total == 0 {
		return
	}

	percentage := float64(current) / float64(total) * 100
	barLength := 50
	filledLength := int(percentage / 100 * float64(barLength))

	bar := strings.Repeat("█", filledLength) + strings.Repeat("░", barLength-filledLength)

	fmt.Printf("\r%s [%s] %.1f%% (%d/%d)", message, bar, percentage, current, total)

	if current == total {
		fmt.Println() // 完成时换行
	}
}

// ShowError 显示错误信息
func (ui *InteractiveUI) ShowError(err error, suggestion string) {
	fmt.Printf("❌ 错误: %v\n", err)
	if suggestion != "" {
		fmt.Printf("💡 建议: %s\n", suggestion)
	}
}

// ShowSuccess 显示成功信息
func (ui *InteractiveUI) ShowSuccess(message string) {
	fmt.Printf("✅ %s\n", message)
}

// ShowWarning 显示警告信息
func (ui *InteractiveUI) ShowWarning(message string) {
	fmt.Printf("⚠️  %s\n", message)
}

// ShowInfo 显示信息
func (ui *InteractiveUI) ShowInfo(message string) {
	fmt.Printf("ℹ️  %s\n", message)
}

// ConfirmAction 确认操作
func (ui *InteractiveUI) ConfirmAction(message string) bool {
	fmt.Printf("%s [y/N]: ", message)

	input, err := ui.reader.ReadString('\n')
	if err != nil {
		return false
	}

	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes"
}

// ShowSummary 显示操作总结
func (ui *InteractiveUI) ShowSummary(successCount, failCount int, operation string) {
	total := successCount + failCount
	if total == 0 {
		return
	}

	fmt.Printf("\n📊 %s完成: ", operation)
	if successCount > 0 {
		fmt.Printf("✅ 成功 %d 个", successCount)
	}
	if failCount > 0 {
		if successCount > 0 {
			fmt.Print("，")
		}
		fmt.Printf("❌ 失败 %d 个", failCount)
	}
	fmt.Printf("，总计 %d 个\n", total)
}
