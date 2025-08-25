package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// SmartDeleteOptions 智能删除选项
type SmartDeleteOptions struct {
	Force               bool    // 强制删除，跳过确认
	Interactive         bool    // 交互模式
	Recursive           bool    // 递归删除
	DryRun              bool    // 试运行
	SmartSearch         bool    // 启用智能搜索
	SearchContent       bool    // 搜索文件内容
	SearchParent        bool    // 搜索父目录
	SimilarityThreshold float64 // 相似度阈值
	MaxResults          int     // 最大搜索结果数
}

// DefaultSmartDeleteOptions 返回默认选项
func DefaultSmartDeleteOptions() SmartDeleteOptions {
	return SmartDeleteOptions{
		Force:               false,
		Interactive:         true,
		Recursive:           false,
		DryRun:              false,
		SmartSearch:         true,
		SearchContent:       false,
		SearchParent:        false,
		SimilarityThreshold: DefaultSimilarityThreshold,
		MaxResults:          DefaultMaxResults,
	}
}

// SmartDelete 智能删除引擎
type SmartDelete struct {
	options SmartDeleteOptions
	ui      *InteractiveUI
	search  *SmartFileSearch
}

// NewSmartDelete 创建智能删除引擎
func NewSmartDelete(options SmartDeleteOptions) *SmartDelete {
	searchConfig := SmartSearchConfig{
		SimilarityThreshold: options.SimilarityThreshold,
		MaxResults:          options.MaxResults,
		SearchContent:       options.SearchContent,
		Recursive:           options.Recursive,
		SearchParent:        options.SearchParent,
	}

	return &SmartDelete{
		options: options,
		ui:      NewInteractiveUI(),
		search:  NewSmartFileSearch(searchConfig),
	}
}

// Delete 智能删除文件
func (sd *SmartDelete) Delete(targets []string) error {
	var processedFiles []string
	var errors []error

	for _, target := range targets {
		files, err := sd.resolveTarget(target)
		if err != nil {
			sd.ui.ShowError(err, "请检查文件路径是否正确")
			errors = append(errors, err)
			continue
		}

		processedFiles = append(processedFiles, files...)
	}

	if len(processedFiles) == 0 {
		return fmt.Errorf("没有找到要删除的文件")
	}

	// 批量确认
	if len(processedFiles) > 1 && !sd.options.Force {
		confirmed, err := sd.ui.ShowBatchConfirmation(processedFiles, "删除", sd.options.Force)
		if err != nil {
			return err
		}
		if !confirmed {
			return fmt.Errorf("用户取消操作")
		}
	}

	// 执行删除
	successCount, failCount := sd.executeDelete(processedFiles)

	// 显示总结
	sd.ui.ShowSummary(successCount, failCount, "删除")

	if failCount > 0 {
		return fmt.Errorf("部分文件删除失败")
	}

	return nil
}

// resolveTarget 解析目标文件
func (sd *SmartDelete) resolveTarget(target string) ([]string, error) {
	// 首先检查文件是否直接存在
	if _, err := os.Stat(target); err == nil {
		return []string{target}, nil
	}

	// 如果不存在且启用了智能搜索
	if sd.options.SmartSearch {
		return sd.smartResolve(target)
	}

	// 尝试通配符匹配
	if isRegexPattern(target) {
		return sd.regexResolve(target)
	}

	return nil, fmt.Errorf("文件不存在: %s", target)
}

// smartResolve 智能解析文件
func (sd *SmartDelete) smartResolve(target string) ([]string, error) {
	sd.ui.ShowLoadingMessage(fmt.Sprintf("正在搜索与 '%s' 相似的文件", target))

	// 获取当前目录
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// 搜索相似文件
	results, err := sd.search.SearchFiles(target, currentDir)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("未找到与 '%s' 匹配的文件", target)
	}

	// 如果只有一个结果且相似度很高，直接使用
	if len(results) == 1 && results[0].Similarity >= 90.0 {
		sd.ui.ShowInfo(fmt.Sprintf("自动选择高相似度文件: %s (%.1f%%)", results[0].Path, results[0].Similarity))
		return []string{results[0].Path}, nil
	}

	// 显示搜索结果让用户选择
	selectedPath, err := sd.ui.ShowSearchResults(results, target)
	if err != nil {
		return nil, err
	}

	return []string{selectedPath}, nil
}

// regexResolve 正则表达式解析
func (sd *SmartDelete) regexResolve(pattern string) ([]string, error) {
	parser, err := NewRegexParser(pattern)
	if err != nil {
		return nil, err
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	matches, err := parser.FindMatches(currentDir, sd.options.Recursive)
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		// 如果正则匹配失败，尝试智能搜索
		if sd.options.SmartSearch {
			sd.ui.ShowWarning(fmt.Sprintf("正则表达式 '%s' 未匹配到文件，尝试智能搜索", pattern))
			return sd.smartResolve(pattern)
		}
		return nil, fmt.Errorf("正则表达式 '%s' 未匹配到任何文件", pattern)
	}

	// 如果匹配到多个文件，需要确认
	if len(matches) > 1 && !sd.options.Force {
		sd.ui.ShowInfo(fmt.Sprintf("正则表达式 '%s' 匹配到 %d 个文件", pattern, len(matches)))

		// 显示匹配的文件
		for i, match := range matches {
			if i >= 5 { // 最多显示5个
				sd.ui.ShowInfo(fmt.Sprintf("... 还有 %d 个文件", len(matches)-5))
				break
			}
			sd.ui.ShowInfo(fmt.Sprintf("  %s", match))
		}

		if !sd.ui.ConfirmAction("确认删除这些文件吗？") {
			return nil, fmt.Errorf("用户取消操作")
		}
	}

	return matches, nil
}

// executeDelete 执行删除操作
func (sd *SmartDelete) executeDelete(files []string) (int, int) {
	successCount := 0
	failCount := 0

	for i, file := range files {
		// 显示进度
		sd.ui.ShowProgressBar(i+1, len(files), "删除进度")

		if sd.options.DryRun {
			sd.ui.ShowInfo(fmt.Sprintf("试运行: 将删除 %s", file))
			successCount++
			continue
		}

		// 交互式确认单个文件
		if sd.options.Interactive && len(files) == 1 {
			if !sd.ui.ConfirmAction(fmt.Sprintf("确认删除文件 '%s'？", file)) {
				sd.ui.ShowInfo("跳过文件: " + file)
				continue
			}
		}

		// 执行删除
		if err := sd.deleteFile(file); err != nil {
			sd.ui.ShowError(err, fmt.Sprintf("删除文件失败: %s", file))
			failCount++
		} else {
			sd.ui.ShowSuccess(fmt.Sprintf("已删除: %s", file))
			successCount++
		}
	}

	return successCount, failCount
}

// deleteFile 删除单个文件
func (sd *SmartDelete) deleteFile(filePath string) error {
	// 安全检查
	if err := sd.performSafetyChecks(filePath); err != nil {
		return err
	}

	// 使用现有的删除函数
	return moveToTrashPlatform(filePath)
}

// performSafetyChecks 执行安全检查
func (sd *SmartDelete) performSafetyChecks(filePath string) error {
	// 检查是否为关键系统路径
	if IsCriticalPath(filePath) {
		if !sd.options.Force {
			return fmt.Errorf("拒绝删除系统关键路径: %s", filePath)
		}
		sd.ui.ShowWarning(fmt.Sprintf("警告: 正在删除系统关键路径: %s", filePath))
	}

	// 检查文件权限
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	if err := checkFilePermissions(filePath, info); err != nil {
		return err
	}

	// 检查隐藏文件
	if isHidden, _ := isHiddenFile(info, filePath); isHidden {
		if !sd.options.Force && !confirmHiddenFileDeletion(filePath) {
			return fmt.Errorf("用户拒绝删除隐藏文件: %s", filePath)
		}
	}

	return nil
}

// ProcessSmartDelete 处理智能删除命令
func ProcessSmartDelete(args []string, options SmartDeleteOptions) error {
	if len(args) == 0 {
		return fmt.Errorf("请指定要删除的文件")
	}

	smartDelete := NewSmartDelete(options)
	return smartDelete.Delete(args)
}

// EnhanceMainWithSmartDelete 增强主程序以支持智能删除
func EnhanceMainWithSmartDelete(files []string, globalOptions SmartDeleteOptions) error {
	// 预处理文件列表，移除重复项
	uniqueFiles := make(map[string]bool)
	var processedFiles []string

	for _, file := range files {
		// 清理路径
		cleanPath := filepath.Clean(file)
		if !uniqueFiles[cleanPath] {
			uniqueFiles[cleanPath] = true
			processedFiles = append(processedFiles, cleanPath)
		}
	}

	// 使用智能删除处理
	return ProcessSmartDelete(processedFiles, globalOptions)
}
