package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ShowLoadingMessage 显示加载消息
// 显示带有动画效果的加载消息，给用户反馈正在进行的操作
// 参数:
//
//	message: 要显示的加载消息文本
func (ui *InteractiveUI) ShowLoadingMessage(message string) {
	fmt.Printf(T("🔍 %s"), message)
	// 简单的加载动画
	for i := 0; i < 3; i++ {
		time.Sleep(500 * time.Millisecond)
		fmt.Print(".")
	}
	fmt.Println()
}

// ShowSearchResults 显示搜索结果并获取用户选择
// 显示智能搜索的结果列表，并提示用户选择要操作的文件
// 参数:
//
//	results: 搜索结果列表
//	target: 用户搜索的目标字符串
//
// 返回值:
//
//	string: 用户选择的文件路径
//	error: 用户取消操作或选择无效时返回的错误
func (ui *InteractiveUI) ShowSearchResults(results []SearchResult, target string) (string, error) {
	if len(results) == 0 {
		fmt.Printf(T("❌ 未找到与 '%s' 匹配的文件\n"), target)
		return "", fmt.Errorf(T("未找到匹配的文件"))
	}

	fmt.Printf(T("🔍 未找到文件 '%s'，找到以下相似文件：\n\n"), target)

	// 显示搜索结果
	for i, result := range results {
		icon := ui.getMatchTypeIcon(result.MatchType)
		fmt.Printf(T("[%d] %s %s"), i+1, icon, result.Path)

		if result.Similarity < 100.0 {
			fmt.Printf(T(" (相似度: %.1f%%)"), result.Similarity)
		}

		if result.MatchType == "content" && result.Context != "" {
			fmt.Printf(T("\n    匹配内容: %s"), result.Context)
		}
		fmt.Println()
	}

	fmt.Printf(T("\n请选择文件编号 (1-%d)，或输入 'n' 取消操作: "), len(results))

	var input string
	if isStdinInteractive() {
		if s, ok := readLineWithTimeout(20 * time.Second); ok {
			input = strings.TrimSpace(s)
		} else {
			input = ""
		}
	} else {
		input = ""
	}
	if strings.ToLower(input) == "n" {
		return "", fmt.Errorf(T("用户取消操作"))
	}

	// 解析用户输入
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(results) {
		return "", fmt.Errorf(T("无效的选择"))
	}

	return results[choice-1].Path, nil
}

// getMatchTypeIcon 获取匹配类型图标
// 根据匹配类型返回相应的图标字符，用于在界面中直观显示匹配类型
// 参数:
//
//	matchType: 匹配类型字符串
//
// 返回值:
//
//	string: 对应的图标字符
func (ui *InteractiveUI) getMatchTypeIcon(matchType string) string {
	switch matchType {
	case "exact":
		return ""
	case "filename":
		return ""
	case "content":
		return ""
	case "regex":
		return ""
	case "parent_filename":
		return "" // 父目录+文件名
	case "parent_content":
		return "" // 父目录+内容
	case "subdir_filename":
		return "" // 子目录+文件名
	case "subdir_content":
		return "" // 子目录+内容
	default:
		return ""
	}
}

// ShowBatchConfirmation 显示批量操作确认
// 显示批量操作的文件列表，并提示用户确认是否执行操作
// 参数:
//
//	files: 要操作的文件列表
//	operation: 操作类型（如"删除"、"复制"等）
//	force: 是否强制执行（跳过确认）
//
// 返回值:
//
//	bool: 用户是否确认执行操作
//	error: 操作过程中可能发生的错误
func (ui *InteractiveUI) ShowBatchConfirmation(files []string, operation string, force bool) (bool, error) {
	if force {
		return true, nil
	}

	if len(files) == 0 {
		fmt.Println(T("没有找到匹配的文件"))
		return false, nil
	}

	fmt.Printf(T("  准备%s %d 个文件：\n\n"), operation, len(files))

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

		fmt.Printf(T(" 第 %d/%d 页：\n"), currentPage, totalPages)
		for i := start; i < end; i++ {
			fmt.Printf(T("  %d. %s\n"), i+1, files[i])
		}

		// 显示操作选项
		fmt.Printf(T("\n选项：\n"))
		fmt.Printf(T("  y - 确认%s所有文件\n"), operation)
		fmt.Printf(T("  n - 取消操作\n"))
		if totalPages > 1 {
			if currentPage < totalPages {
				fmt.Printf(T("  > - 下一页\n"))
			}
			if currentPage > 1 {
				fmt.Printf(T("  < - 上一页\n"))
			}
		}
		fmt.Printf(T("  s - 跳过确认（强制执行）\n"))
		fmt.Print(T("\n请选择: "))

		var input string
		if isStdinInteractive() {
			if s, ok := readLineWithTimeout(30 * time.Second); ok {
				input = strings.ToLower(strings.TrimSpace(s))
			} else {
				input = ""
			}
		} else {
			input = ""
		}

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
			fmt.Println(T(" 无效的选择，请重新输入"))
		}

		fmt.Println() // 空行分隔
	}
}

// ShowProgressBar 显示进度条
// 显示操作进度的可视化进度条
// 参数:
//
//	current: 当前进度
//	total: 总进度
//	message: 进度消息
func (ui *InteractiveUI) ShowProgressBar(current, total int, message string) {
	if total == 0 {
		return
	}

	percentage := float64(current) / float64(total) * 100
	barLength := 50
	filledLength := int(percentage / 100 * float64(barLength))

	bar := strings.Repeat("", filledLength) + strings.Repeat("", barLength-filledLength)

	fmt.Printf("\r%s [%s] %.1f%% (%d/%d)", message, bar, percentage, current, total)

	if current == total {
		fmt.Println() // 完成时换行
	}
}

// ShowError 显示错误信息
// 显示错误信息和建议解决方案
// 参数:
//
//	err: 错误对象
//	suggestion: 建议的解决方案
func (ui *InteractiveUI) ShowError(err error, suggestion string) {
	// 记录错误：operation=交互界面错误, filePath留空, err为真实错误, message为建议
	logger.Error("交互界面错误", "", err, suggestion)
	if suggestion != "" {
		fmt.Printf(T(" 建议: %s\n"), suggestion)
	}
}

// ShowSmartError 显示智能错误信息和建议
// 根据错误类型显示智能的错误信息和针对性的建议
// 参数:
//
//	err: 错误对象
//	context: 错误上下文描述
func (ui *InteractiveUI) ShowSmartError(err error, context string) {
	fmt.Printf(T(" %s: %v\n"), context, err)

	// 根据错误类型提供智能建议
	errorMsg := err.Error()
	switch {
	case strings.Contains(errorMsg, "no such file") || strings.Contains(errorMsg, "not exist"):
		fmt.Printf(T(" 建议：\n"))
		fmt.Printf(T("   1. 检查文件路径是否正确\n"))
		fmt.Printf(T("   2. 使用 --smart-search 启用智能搜索\n"))
		fmt.Printf(T("   3. 使用通配符如 *.txt 或 file*\n"))
		fmt.Printf(T("   4. 使用 --search-content 搜索文件内容\n"))
	case strings.Contains(errorMsg, "permission"):
		fmt.Printf(T(" 建议：\n"))
		fmt.Printf(T("   1. 以管理员身份运行\n"))
		fmt.Printf(T("   2. 检查文件权限设置\n"))
		fmt.Printf(T("   3. 确认文件没有被其他程序占用\n"))
	case strings.Contains(errorMsg, "invalid"):
		fmt.Printf(T(" 建议：\n"))
		fmt.Printf(T("   1. 检查文件名中是否包含非法字符\n"))
		fmt.Printf(T("   2. 避免使用特殊字符如 < > | \" : * ? \\ /\n"))
		fmt.Printf(T("   3. 使用引号包围包含空格的文件名\n"))
	case strings.Contains(errorMsg, "too long"):
		fmt.Printf(T(" 建议：\n"))
		fmt.Printf(T("   1. 缩短文件路径或文件名\n"))
		fmt.Printf(T("   2. 移动到更短的目录路径\n"))
	default:
		// 记录一般建议：operation=操作建议, filePath留空, message为提示
		logger.Info("操作建议", "", T("请检查错误信息并尝试重新操作"))
	}
}

// ShowSuccess 显示成功信息
// 显示操作成功的消息
// 参数:
//
//	message: 成功消息
func (ui *InteractiveUI) ShowSuccess(message string) {
	fmt.Printf(T(" %s\n"), message)
}

// ShowWarning 显示警告信息
// 显示警告消息
// 参数:
//
//	message: 警告消息
func (ui *InteractiveUI) ShowWarning(message string) {
	fmt.Printf(T("  %s\n"), message)
}

// ShowInfo 显示信息
// 显示普通信息消息
// 参数:
//
//	message: 信息消息
func (ui *InteractiveUI) ShowInfo(message string) {
	fmt.Printf(T("  %s\n"), message)
}

// ConfirmAction 确认操作
// 提示用户确认是否执行某个操作
// 参数:
//
//	message: 确认消息
//
// 返回值:
//
//	bool: 用户是否确认执行操作
func (ui *InteractiveUI) ConfirmAction(message string) bool {
	fmt.Printf("%s [y/N]: ", message)

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
	return input == "y" || input == "yes"
}

// ShowSummary 显示操作总结
// 显示操作完成后的总结信息
// 参数:
//
//	successCount: 成功操作的数量
//	failCount: 失败操作的数量
//	operation: 操作类型
func (ui *InteractiveUI) ShowSummary(successCount, failCount int, operation string) {
	total := successCount + failCount
	if total == 0 {
		return
	}

	fmt.Printf(T("\n📊 %s完成: "), operation)
	if successCount > 0 {
		fmt.Printf(T("✅ 成功 %d 个"), successCount)
	}
	if failCount > 0 {
		if successCount > 0 {
			fmt.Print(T("，"))
		}
		fmt.Printf(T("❌ 失败 %d 个"), failCount)
	}
	fmt.Printf(T("，总计 %d 个\n"), total)
}
