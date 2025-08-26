package main

import (
	"fmt"
	"os"
)

// smartResolveFile 智能解析文件路径
func smartResolveFile(target string) ([]string, error) {
	// 先提示用户没有找到指定文件
	fmt.Printf(T("⚠️  未找到文件 '%s'，正在进行智能搜索...\n"), target)

	// 获取当前目录
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// 创建增强版智能搜索配置
	searchConfig := SmartSearchConfig{
		SimilarityThreshold: similarityThreshold,
		MaxResults:          10, // 限制为10个结果
		SearchContent:       searchContent,
		Recursive:           recursive,
		SearchParent:        searchParent,
		CaseSensitive:       false, // 默认不区分大小写
	}

	// 创建增强版搜索引擎
	search := NewEnhancedSmartSearch(searchConfig)

	// 搜索相似文件（使用缓存）
	results, err := search.SearchWithCache(target, currentDir)
	if err != nil {
		return nil, fmt.Errorf(T("智能搜索失败: %v"), err)
	}

	if len(results) == 0 {
		// 如果基本搜索没找到，尝试内容搜索
		if searchContent {
			fmt.Printf(T("🔍 未找到文件名匹配，正在搜索文件内容...\n"))
			contentResults, contentErr := search.SearchContentWithCache(target, currentDir)
			if contentErr == nil && len(contentResults) > 0 {
				results = contentResults
			} else {
				return nil, fmt.Errorf(T("未找到与 '%s' 匹配的文件"), target)
			}
		} else {
			return nil, fmt.Errorf(T("未找到与 '%s' 匹配的文件"), target)
		}
	}

	// 如果只有一个结果且相似度很高，直接返回
	if len(results) == 1 && results[0].Similarity >= 90.0 {
		fmt.Printf(T("🔍 自动选择高相似度文件: %s (%.1f%%)\n"), results[0].Path, results[0].Similarity)
		return []string{results[0].Path}, nil
	}

	// 创建交互式UI
	ui := NewInteractiveUI()

	// 显示搜索结果让用户选择
	selectedPath, err := ui.ShowSearchResults(results, target)
	if err != nil {
		return nil, err
	}

	return []string{selectedPath}, nil
}

// smartResolveWithRegex 使用正则表达式智能解析
func smartResolveWithRegex(pattern string) ([]string, error) {
	parser, err := NewRegexParser(pattern)
	if err != nil {
		return nil, err
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	matches, err := parser.FindMatches(currentDir, recursive)
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		// 如果正则匹配失败，尝试智能搜索
		if smartSearch {
			fmt.Printf("⚠️  正则表达式 '%s' 未匹配到文件，尝试智能搜索\n", pattern)
			return smartResolveFile(pattern)
		}
		return nil, fmt.Errorf("正则表达式 '%s' 未匹配到任何文件", pattern)
	}

	// 如果匹配到多个文件，需要确认
	if len(matches) > 1 && !force && !forceConfirm {
		ui := NewInteractiveUI()
		ui.ShowInfo(fmt.Sprintf("正则表达式 '%s' 匹配到 %d 个文件", pattern, len(matches)))

		// 显示匹配的文件
		for i, match := range matches {
			if i >= 5 { // 最多显示5个
				ui.ShowInfo(fmt.Sprintf("... 还有 %d 个文件", len(matches)-5))
				break
			}
			ui.ShowInfo(fmt.Sprintf("  %s", match))
		}

		if !ui.ConfirmAction("确认删除这些文件吗？") {
			return nil, fmt.Errorf("用户取消操作")
		}
	}

	return matches, nil
}

// enhancedFileResolver 增强的文件解析器
func enhancedFileResolver(target string) ([]string, error) {
	// 首先检查文件是否直接存在
	if _, err := os.Stat(target); err == nil {
		return []string{target}, nil
	}

	// 检查是否为正则表达式或通配符
	if isRegexPattern(target) {
		return smartResolveWithRegex(target)
	}

	// 如果启用智能搜索，尝试查找相似文件
	if smartSearch {
		return smartResolveFile(target)
	}

	return nil, fmt.Errorf("文件不存在: %s", target)
}
