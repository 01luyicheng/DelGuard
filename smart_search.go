package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ... (rest of the code remains the same)

// PromptUserSelection 提示用户选择文件
// 显示搜索结果并提示用户选择要操作的文件
// 参数:
//
//	results: 搜索结果列表
//	target: 用户搜索的目标字符串
//
// 返回值:
//
//	string: 用户选择的文件路径
//	error: 用户取消操作或选择无效时返回的错误
func PromptUserSelection(results []SearchResult, target string) (string, error) {
	if len(results) == 0 {
		return "", fmt.Errorf("未找到匹配的文件")
	}

	fmt.Printf("🔍 未找到文件 '%s'，找到以下相似文件：\n\n", target)

	// 显示搜索结果
	for i, result := range results {
		fmt.Printf("[%d] %s", i+1, result.Path)
		if result.Similarity < 100.0 {
			fmt.Printf(" (相似度: %.1f%%)", result.Similarity)
		}
		if result.MatchType == "content" && result.Context != "" {
			fmt.Printf("\n    💡 匹配内容: %s", result.Context)
		}
		fmt.Println()
	}

	fmt.Printf("\n请选择文件编号 (1-%d)，或输入 'n' 取消操作: ", len(results))

	var input string
	if isStdinInteractive() {
		if s, ok := readLineWithTimeout(30 * time.Second); ok {
			input = strings.TrimSpace(s)
		} else {
			input = ""
		}
	} else {
		input = ""
	}
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

// searchInSubdirectories 在子目录中搜索
// 在指定目录的子目录中搜索匹配的文件
// 参数:
//
//	target: 要搜索的目标字符串
//	searchDir: 搜索的根目录路径
//
// 返回值:
//
//	[]SearchResult: 在子目录中找到的匹配结果
//	error: 搜索过程中可能发生的错误
func (s *SmartFileSearch) searchInSubdirectories(target string, searchDir string) ([]SearchResult, error) {
	var results []SearchResult

	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续搜索
		}

		// 跳过根目录
		if path == searchDir {
			return nil
		}

		// 只处理子目录
		if info.IsDir() {
			// 在子目录中搜索
			subResults, subErr := s.searchBySimilarity(target, path)
			if subErr == nil {
				// 为子目录结果添加标识
				for i := range subResults {
					subResults[i].MatchType = "subdir_" + subResults[i].MatchType
					subResults[i].Similarity *= SubdirMatchSimilarity // 子目录结果相似度较低
				}
				results = append(results, subResults...)
			}

			// 控制递归深度，避免过深的目录遭历
			relPath, _ := filepath.Rel(searchDir, path)
			depth := len(strings.Split(relPath, string(filepath.Separator)))
			if depth > MaxSearchDepth { // 限制最大深度
				return filepath.SkipDir
			}
		}

		return nil
	})

	return results, err
}

// enhancedContentSearch 增强的内容搜索
// 在文本文件中进行更智能的内容搜索，支持更多文件类型
// 参数:
//
//	target: 要搜索的目标字符串
//	searchDir: 搜索的目录路径
//
// 返回值:
//
//	[]SearchResult: 包含匹配文件的搜索结果列表
//	error: 搜索过程中可能发生的错误
func (s *SmartFileSearch) enhancedContentSearch(target string, searchDir string) ([]SearchResult, error) {
	var results []SearchResult

	// 扩展可搜索的文件类型
	textExts := []string{".txt", ".md", ".log", ".cfg", ".conf", ".ini", ".json", ".xml", ".yaml", ".yml", ".csv", ".sql", ".sh", ".bat", ".ps1", ".py", ".js", ".html", ".css", ".go", ".java", ".c", ".cpp", ".h", ".hpp"}

	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if !s.config.Recursive && path != searchDir {
				return filepath.SkipDir
			}
			return nil
		}

		// 检查文件类型
		ext := strings.ToLower(filepath.Ext(path))
		isTextFile := false
		for _, textExt := range textExts {
			if ext == textExt {
				isTextFile = true
				break
			}
		}

		if !isTextFile {
			return nil
		}

		// 搜索文件内容
		matches, context := s.searchInFileEnhanced(path, target)
		if len(matches) > 0 {
			results = append(results, SearchResult{
				Path:       path,
				Name:       filepath.Base(path),
				Similarity: EnhancedContentBaseSimilarity + float64(len(matches)*EnhancedContentSimilarityPerMatch), // 根据匹配数量调整相似度
				MatchType:  "content",
				Context:    context,
			})
		}

		return nil
	})

	return results, err
}

// searchInFileEnhanced 增强的文件内容搜索
// 提供更智能的文件内容搜索，支持多种匹配模式
// 参数:
//
//	filePath: 要搜索的文件路径
//	target: 要搜索的目标字符串
//
// 返回值:
//
//	[]string: 包含匹配行号的字符串列表
//	string: 匹配行的上下文内容
func (s *SmartFileSearch) searchInFileEnhanced(filePath string, target string) ([]string, string) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, ""
	}
	defer file.Close()

	var matches []string
	var contexts []string
	scanner := bufio.NewScanner(file)
	lineNum := 0
	maxMatches := MaxEnhancedMatches // 限制最多匹配数量

	// 支持多种匹配模式
	targetLower := strings.ToLower(target)

	for scanner.Scan() && len(matches) < maxMatches {
		lineNum++
		line := scanner.Text()
		lineLower := strings.ToLower(line)

		// 精确匹配
		if strings.Contains(lineLower, targetLower) {
			matches = append(matches, fmt.Sprintf("第%d行", lineNum))
			contexts = append(contexts, fmt.Sprintf("第%d行: %s", lineNum, truncateString(line, TruncateContextLength)))
			continue
		}

		// 模糊匹配（去除空格和特殊字符）
		cleanLine := strings.ReplaceAll(strings.ReplaceAll(lineLower, " ", ""), "_", "")
		cleanTarget := strings.ReplaceAll(strings.ReplaceAll(targetLower, " ", ""), "_", "")
		if strings.Contains(cleanLine, cleanTarget) {
			matches = append(matches, fmt.Sprintf("第%d行(模糊)", lineNum))
			contexts = append(contexts, fmt.Sprintf("第%d行: %s", lineNum, truncateString(line, TruncateContextLength)))
		}
	}

	var context string
	if len(contexts) > 0 {
		if len(contexts) == 1 {
			context = contexts[0]
		} else {
			context = fmt.Sprintf("%s 等%d处匹配", contexts[0], len(contexts))
		}
	}

	return matches, context
}

// searchByContent 包装内容搜索，供对外统一调用
func (s *SmartFileSearch) searchByContent(target string, searchDir string) ([]SearchResult, error) {
    return s.enhancedContentSearch(target, searchDir)
}

// sortAndLimitResults 对结果按相似度降序排序并按配置限制数量
func (s *SmartFileSearch) sortAndLimitResults(results []SearchResult) []SearchResult {
    // 相似度高在前
    sort.Slice(results, func(i, j int) bool {
        return results[i].Similarity > results[j].Similarity
    })

    if s.config.MaxResults > 0 && len(results) > s.config.MaxResults {
        return results[:s.config.MaxResults]
    }
    return results
}
