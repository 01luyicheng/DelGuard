package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// SmartSearchConfig 智能搜索配置
type SmartSearchConfig struct {
	SimilarityThreshold float64 // 相似度阈值
	MaxResults          int     // 最大结果数量
	SearchContent       bool    // 是否搜索文件内容
	Recursive           bool    // 是否递归搜索
	SearchParent        bool    // 是否搜索父目录
}

// DefaultSmartSearchConfig 返回默认搜索配置
func DefaultSmartSearchConfig() SmartSearchConfig {
	return SmartSearchConfig{
		SimilarityThreshold: 60.0,
		MaxResults:          10,
		SearchContent:       false,
		Recursive:           true,
		SearchParent:        false,
	}
}

// SearchResult 搜索结果
type SearchResult struct {
	Path       string  // 文件路径
	Name       string  // 文件名
	Similarity float64 // 相似度
	MatchType  string  // 匹配类型：filename, content, regex
	Context    string  // 上下文信息（用于内容匹配）
}

// SmartFileSearch 智能文件搜索引擎
type SmartFileSearch struct {
	config SmartSearchConfig
}

// NewSmartFileSearch 创建新的智能搜索引擎
func NewSmartFileSearch(config SmartSearchConfig) *SmartFileSearch {
	return &SmartFileSearch{config: config}
}

// SearchFiles 搜索文件
func (s *SmartFileSearch) SearchFiles(target string, searchDir string) ([]SearchResult, error) {
	var results []SearchResult

	// 首先尝试直接匹配
	if _, err := os.Stat(filepath.Join(searchDir, target)); err == nil {
		results = append(results, SearchResult{
			Path:       filepath.Join(searchDir, target),
			Name:       target,
			Similarity: 100.0,
			MatchType:  "exact",
		})
		return results, nil
	}

	// 检查是否为正则表达式或通配符
	if isRegexPattern(target) {
		regexResults, err := s.searchByRegex(target, searchDir)
		if err == nil {
			results = append(results, regexResults...)
		}
	}

	// 文件名相似度搜索
	filenameResults, err := s.searchBySimilarity(target, searchDir)
	if err == nil {
		results = append(results, filenameResults...)
	}

	// 文件内容搜索
	if s.config.SearchContent {
		contentResults, err := s.searchByContent(target, searchDir)
		if err == nil {
			results = append(results, contentResults...)
		}
	}

	// 父目录搜索
	if s.config.SearchParent && len(results) == 0 {
		parentDir := filepath.Dir(searchDir)
		if parentDir != searchDir {
			parentResults, err := s.searchBySimilarity(target, parentDir)
			if err == nil {
				results = append(results, parentResults...)
			}
		}
	}

	// 排序并限制结果数量
	results = s.sortAndLimitResults(results)

	return results, nil
}

// isRegexPattern 检查是否为正则表达式或通配符模式
func isRegexPattern(pattern string) bool {
	// 检查通配符
	if strings.Contains(pattern, "*") || strings.Contains(pattern, "?") {
		return true
	}
	// 检查正则表达式特殊字符
	regexChars := []string{"[", "]", "(", ")", "{", "}", "^", "$", "+", "|", "\\"}
	for _, char := range regexChars {
		if strings.Contains(pattern, char) {
			return true
		}
	}
	return false
}

// searchByRegex 通过正则表达式搜索
func (s *SmartFileSearch) searchByRegex(pattern string, searchDir string) ([]SearchResult, error) {
	var results []SearchResult

	// 将通配符转换为正则表达式
	regexPattern := wildcardToRegexSmart(pattern)
	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, err
	}

	err = filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续搜索
		}

		// 跳过目录（除非递归搜索）
		if info.IsDir() && !s.config.Recursive {
			if path != searchDir {
				return filepath.SkipDir
			}
			return nil
		}

		filename := filepath.Base(path)
		if regex.MatchString(filename) {
			results = append(results, SearchResult{
				Path:       path,
				Name:       filename,
				Similarity: 90.0, // 正则匹配给予高相似度
				MatchType:  "regex",
			})
		}

		return nil
	})

	return results, err
}

// searchBySimilarity 通过文件名相似度搜索
func (s *SmartFileSearch) searchBySimilarity(target string, searchDir string) ([]SearchResult, error) {
	var results []SearchResult
	var candidates []string

	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续搜索
		}

		// 跳过目录（除非递归搜索）
		if info.IsDir() && !s.config.Recursive {
			if path != searchDir {
				return filepath.SkipDir
			}
			return nil
		}

		if !info.IsDir() {
			candidates = append(candidates, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 计算相似度
	for _, candidate := range candidates {
		filename := filepath.Base(candidate)
		similarity := CalculateSimilarity(target, filename)
		if similarity >= s.config.SimilarityThreshold {
			results = append(results, SearchResult{
				Path:       candidate,
				Name:       filename,
				Similarity: similarity,
				MatchType:  "filename",
			})
		}
	}

	return results, nil
}

// searchByContent 通过文件内容搜索
func (s *SmartFileSearch) searchByContent(target string, searchDir string) ([]SearchResult, error) {
	var results []SearchResult

	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续搜索
		}

		// 跳过目录和二进制文件
		if info.IsDir() || isBinaryFile(path) {
			if info.IsDir() && !s.config.Recursive && path != searchDir {
				return filepath.SkipDir
			}
			return nil
		}

		// 搜索文件内容
		if matches, context := searchInFile(path, target); len(matches) > 0 {
			results = append(results, SearchResult{
				Path:       path,
				Name:       filepath.Base(path),
				Similarity: 80.0, // 内容匹配给予中等相似度
				MatchType:  "content",
				Context:    context,
			})
		}

		return nil
	})

	return results, err
}

// isBinaryFile 检查是否为二进制文件
func isBinaryFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	binaryExts := []string{".exe", ".dll", ".so", ".dylib", ".bin", ".obj", ".o", ".a", ".lib", ".zip", ".tar", ".gz", ".jpg", ".png", ".gif", ".mp4", ".avi", ".mp3", ".wav"}
	for _, binaryExt := range binaryExts {
		if ext == binaryExt {
			return true
		}
	}
	return false
}

// searchInFile 在文件中搜索目标字符串
func searchInFile(filePath string, target string) ([]string, string) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, ""
	}
	defer file.Close()

	var matches []string
	var context string
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if strings.Contains(strings.ToLower(line), strings.ToLower(target)) {
			matches = append(matches, fmt.Sprintf("第%d行", lineNum))
			if context == "" {
				// 只保存第一个匹配的上下文
				context = fmt.Sprintf("第%d行: %s", lineNum, truncateString(line, 100))
			}
		}
	}

	return matches, context
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// wildcardToRegexSmart 将通配符转换为正则表达式（智能搜索版本）
func wildcardToRegexSmart(pattern string) string {
	// 转义正则表达式特殊字符
	pattern = regexp.QuoteMeta(pattern)
	// 将转义的通配符替换为正则表达式
	pattern = strings.ReplaceAll(pattern, "\\*", ".*")
	pattern = strings.ReplaceAll(pattern, "\\?", ".")
	return "^" + pattern + "$"
}

// sortAndLimitResults 排序并限制结果数量
func (s *SmartFileSearch) sortAndLimitResults(results []SearchResult) []SearchResult {
	// 按相似度降序排序
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Similarity < results[j].Similarity {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// 限制结果数量
	if len(results) > s.config.MaxResults {
		results = results[:s.config.MaxResults]
	}

	return results
}

// PromptUserSelection 提示用户选择文件
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
			fmt.Printf("\n    匹配内容: %s", result.Context)
		}
		fmt.Println()
	}

	fmt.Printf("\n请选择文件编号 (1-%d)，或输入 'n' 取消操作: ", len(results))

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
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
