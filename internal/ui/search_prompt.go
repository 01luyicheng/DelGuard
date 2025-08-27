package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/01luyicheng/DelGuard/internal/core/search"
)

// SearchPrompt 表示智能搜索提示界面
type SearchPrompt struct {
	visualFeedback *VisualFeedback
}

// NewSearchPrompt 创建一个新的搜索提示界面
func NewSearchPrompt(vf *VisualFeedback) *SearchPrompt {
	return &SearchPrompt{
		visualFeedback: vf,
	}
}

// HandleFileNotFound 处理文件不存在的情况，提供智能搜索建议
func (sp *SearchPrompt) HandleFileNotFound(filePath string) (string, bool) {
	// 使用智能搜索查找相似文件
	similarFiles, err := search.SmartSearch(filePath)

	if err != nil {
		// 没有找到相似文件
		sp.visualFeedback.ShowError(fmt.Sprintf("文件不存在: %s", filePath))
		sp.visualFeedback.ShowInfo("未找到相似文件，请检查文件路径是否正确。")
		return "", false
	}

	// 如果只找到一个相似文件，直接询问是否使用
	if len(similarFiles) == 1 {
		similarFile := similarFiles[0]
		confirmed := sp.askForConfirmation(fmt.Sprintf(
			"文件 '%s' 不存在。\n是否使用找到的相似文件: '%s'?",
			filePath, similarFile))

		if confirmed {
			return similarFile, true
		}
		return "", false
	}

	// 如果找到多个相似文件，显示列表供用户选择
	sp.visualFeedback.ShowWarning(fmt.Sprintf("文件 '%s' 不存在", filePath))
	sp.visualFeedback.ShowInfo("找到以下相似文件:")

	for i, file := range similarFiles {
		sp.visualFeedback.ShowListItem(fmt.Sprintf("%d. %s", i+1, file))
	}

	sp.visualFeedback.ShowInfo("请输入文件编号选择，或输入0取消:")

	var choice int
	fmt.Scanln(&choice)

	if choice <= 0 || choice > len(similarFiles) {
		sp.visualFeedback.ShowInfo("已取消操作")
		return "", false
	}

	selectedFile := similarFiles[choice-1]
	sp.visualFeedback.ShowSuccess(fmt.Sprintf("已选择: %s", selectedFile))
	return selectedFile, true
}

// askForConfirmation 询问用户确认
func (sp *SearchPrompt) askForConfirmation(message string) bool {
	sp.visualFeedback.ShowPrompt(message + " (y/n)")

	var response string
	fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// SuggestCorrection 根据错误类型提供纠正建议
func (sp *SearchPrompt) SuggestCorrection(filePath string, err error) {
	if os.IsNotExist(err) {
		// 文件不存在错误
		sp.HandleFileNotFound(filePath)
	} else if os.IsPermission(err) {
		// 权限错误
		sp.visualFeedback.ShowError(fmt.Sprintf("无权访问文件: %s", filePath))
		sp.visualFeedback.ShowInfo("请检查文件权限或尝试以管理员/root身份运行。")
	} else {
		// 其他错误
		sp.visualFeedback.ShowError(fmt.Sprintf("访问文件时出错: %s", err.Error()))
		sp.visualFeedback.ShowInfo("请检查文件路径是否正确，或者文件系统是否有问题。")
	}
}
