package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// SearchFiles 搜索文件（带超时机制）
func (s *SmartFileSearch) SearchFiles(target string, searchDir string) ([]SearchResult, error) {
	// 创建带超时的上下文（减少超时时间到5秒）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 在搜索前提示用户
	fmt.Printf(T("🔍 未找到文件 '%s'，正在智能搜索相似文件...\n"), target)

	var allResults []SearchResult
	resultChan := make(chan []SearchResult, 3)
	errorChan := make(chan error, 3)

	// 1. 基于文件名相似度搜索
	go func() {
		nameResults, err := s.searchBySimilarityWithContext(ctx, target, searchDir)
		if err != nil {
			errorChan <- err
		} else {
			resultChan <- nameResults
		}
	}()

	// 2. 如果启用内容搜索
	if s.config.SearchContent {
		go func() {
			contentResults, err := s.searchByContentWithContext(ctx, target, searchDir)
			if err != nil {
				errorChan <- err
			} else {
				resultChan <- contentResults
			}
		}()
	}

	// 3. 如果启用父目录搜索
	if s.config.SearchParent {
		go func() {
			parentDir := filepath.Dir(searchDir)
			if parentDir != searchDir && parentDir != "." {
				parentResults, err := s.searchInSubdirectoriesWithContext(ctx, target, parentDir)
				if err != nil {
					errorChan <- err
				} else {
					resultChan <- parentResults
				}
			} else {
				resultChan <- []SearchResult{}
			}
		}()
	}

	// 收集结果
	expectedResults := 1
	if s.config.SearchContent {
		expectedResults++
	}
	if s.config.SearchParent {
		expectedResults++
	}

	timeout := time.NewTimer(5 * time.Second)
	defer timeout.Stop()

	for i := 0; i < expectedResults; i++ {
		select {
		case results := <-resultChan:
			allResults = append(allResults, results...)
		case <-errorChan:
			// 忽略单个搜索的错误，继续其他搜索
		case <-timeout.C:
			fmt.Printf(T("⏰ 搜索超时（5秒），返回已找到的 %d 个结果\n"), len(allResults))
			goto processResults
		case <-ctx.Done():
			fmt.Printf(T("⏰ 搜索被取消，返回已找到的 %d 个结果\n"), len(allResults))
			goto processResults
		}
	}

processResults:
	// 去重并排序
	allResults = s.removeDuplicates(allResults)
	allResults = s.sortAndLimitResults(allResults)

	return allResults, nil
}

// searchBySimilarity 基于相似度搜索文件和目录
func (s *SmartFileSearch) searchBySimilarity(target string, searchDir string) ([]SearchResult, error) {
	var results []SearchResult

	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续搜索
		}

		// 跳过根目录本身
		if path == searchDir {
			return nil
		}

		filename := info.Name()
		similarity := s.calculateSimilarity(target, filename)

		if similarity >= s.config.SimilarityThreshold {
			matchType := s.getMatchType(target, filename)

			// 为目录添加特殊标识和上下文信息
			context := ""
			if info.IsDir() {
				matchType = "dir_" + matchType
				context = T("目录")
			} else {
				context = s.getFileTypeContext(info)
			}

			results = append(results, SearchResult{
				Path:       path,
				Name:       filename,
				Similarity: similarity,
				MatchType:  matchType,
				Context:    context,
			})
		}

		// 如果是目录且不启用递归，跳过该目录
		if info.IsDir() && !s.config.Recursive {
			return filepath.SkipDir
		}

		return nil
	})

	return results, err
}

// getFileTypeContext 获取文件类型上下文信息
func (s *SmartFileSearch) getFileTypeContext(info os.FileInfo) string {
	// 允许处理目录和文件

	// 根据文件大小和类型提供上下文
	size := info.Size()
	if size > 1024*1024*1024 { // > 1GB
		return fmt.Sprintf(T("大文件 (%.1fGB)"), float64(size)/(1024*1024*1024))
	} else if size > 1024*1024 { // > 1MB
		return fmt.Sprintf(T("文件 (%.1fMB)"), float64(size)/(1024*1024))
	} else if size > 1024 { // > 1KB
		return fmt.Sprintf(T("文件 (%.1fKB)"), float64(size)/1024)
	} else {
		return fmt.Sprintf(T("文件 (%d字节)"), size)
	}
}

// calculateSimilarity 计算文件名相似度
func (s *SmartFileSearch) calculateSimilarity(target, filename string) float64 {
	if !s.config.CaseSensitive {
		target = strings.ToLower(target)
		filename = strings.ToLower(filename)
	}

	// 精确匹配
	if target == filename {
		return ExactMatchSimilarity
	}

	// 前缀匹配
	if strings.HasPrefix(filename, target) {
		return PrefixMatchSimilarity
	}

	// 后缀匹配
	if strings.HasSuffix(filename, target) {
		return SuffixMatchSimilarity
	}

	// 包含匹配
	if strings.Contains(filename, target) {
		return ContainsMatchSimilarity
	}

	// 模糊匹配（编辑距离）
	distance := s.levenshteinDistance(target, filename)
	maxLen := len(target)
	if len(filename) > maxLen {
		maxLen = len(filename)
	}

	if maxLen == 0 {
		return 0
	}

	similarity := (1.0 - float64(distance)/float64(maxLen)) * FuzzyMatchSimilarity
	if similarity < 0 {
		similarity = 0
	}

	return similarity
}

// getMatchType 获取匹配类型
func (s *SmartFileSearch) getMatchType(target, filename string) string {
	if !s.config.CaseSensitive {
		target = strings.ToLower(target)
		filename = strings.ToLower(filename)
	}

	if target == filename {
		return "exact"
	} else if strings.HasPrefix(filename, target) {
		return "prefix"
	} else if strings.HasSuffix(filename, target) {
		return "suffix"
	} else if strings.Contains(filename, target) {
		return "contains"
	} else {
		return "fuzzy"
	}
}

// levenshteinDistance 计算编辑距离
func (s *SmartFileSearch) levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}

	for j := 0; j <= len(b); j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}

			// 手动实现min函数以兼容旧版本Go
			deletion := matrix[i-1][j] + 1
			insertion := matrix[i][j-1] + 1
			substitution := matrix[i-1][j-1] + cost

			minVal := deletion
			if insertion < minVal {
				minVal = insertion
			}
			if substitution < minVal {
				minVal = substitution
			}
			matrix[i][j] = minVal
		}
	}

	return matrix[len(a)][len(b)]
}

// searchBySimilarityWithContext 基于相似度搜索文件和目录（带上下文）
func (s *SmartFileSearch) searchBySimilarityWithContext(ctx context.Context, target string, searchDir string) ([]SearchResult, error) {
	var results []SearchResult

	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		// 检查上下文是否被取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			return nil // 忽略错误，继续搜索
		}

		// 跳过根目录本身
		if path == searchDir {
			return nil
		}

		filename := info.Name()
		similarity := s.calculateSimilarity(target, filename)

		if similarity >= s.config.SimilarityThreshold {
			matchType := s.getMatchType(target, filename)

			// 为目录添加特殊标识和上下文信息
			context := ""
			if info.IsDir() {
				matchType = "dir_" + matchType
				context = T("目录")
			} else {
				context = s.getFileTypeContext(info)
			}

			results = append(results, SearchResult{
				Path:       path,
				Name:       filename,
				Similarity: similarity,
				MatchType:  matchType,
				Context:    context,
			})

			// 如果结果太多，提前结束搜索
			if len(results) >= s.config.MaxResults*2 {
				return filepath.SkipAll
			}
		}

		// 如果是目录且不启用递归，跳过该目录
		if info.IsDir() && !s.config.Recursive {
			return filepath.SkipDir
		}

		return nil
	})

	return results, err
}

// searchByContentWithContext 内容搜索方法（带上下文）
func (s *SmartFileSearch) searchByContentWithContext(ctx context.Context, target string, searchDir string) ([]SearchResult, error) {
	var results []SearchResult

	// 检查目标是否为空
	if strings.TrimSpace(target) == "" {
		return results, nil
	}

	// 限制搜索的文件大小，避免大文件影响性能
	maxFileSize := int64(10 * 1024 * 1024) // 10MB

	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		// 检查上下文是否被取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			return nil // 忽略错误，继续搜索
		}

		// 跳过目录
		if info.IsDir() {
			// 如果是目录且不启用递归，跳过该目录
			if !s.config.Recursive && path != searchDir {
				return filepath.SkipDir
			}
			return nil
		}

		// 跳过过大的文件
		if info.Size() > maxFileSize {
			return nil
		}

		// 检查文件扩展名，只搜索文本文件
		ext := strings.ToLower(filepath.Ext(path))
		isTextFile := false
		for _, textExt := range TextFileExtensions {
			if ext == textExt {
				isTextFile = true
				break
			}
		}

		if !isTextFile {
			return nil
		}

		// 读取文件内容并搜索
		content, err := os.ReadFile(path)
		if err != nil {
			return nil // 忽略无法读取的文件
		}

		contentStr := string(content)
		if !s.config.CaseSensitive {
			contentStr = strings.ToLower(contentStr)
			target = strings.ToLower(target)
		}

		// 搜索内容
		if strings.Contains(contentStr, target) {
			// 计算相似度（基于匹配次数和位置）
			matchCount := strings.Count(contentStr, target)
			similarity := SearchContentSimilarity + float64(matchCount)*5.0
			if similarity > 95.0 {
				similarity = 95.0
			}

			// 获取匹配上下文
			context := s.extractContext(contentStr, target)

			results = append(results, SearchResult{
				Path:       path,
				Name:       info.Name(),
				Similarity: similarity,
				MatchType:  "content",
				Context:    context,
			})

			// 限制结果数量
			if len(results) >= s.config.MaxResults*2 {
				return filepath.SkipAll
			}
		}

		return nil
	})

	return results, err
}

// extractContext 提取匹配内容的上下文
func (s *SmartFileSearch) extractContext(content, target string) string {
	// 查找目标字符串的位置
	index := strings.Index(content, target)
	if index == -1 {
		return ""
	}

	// 计算上下文范围
	start := index - 30
	if start < 0 {
		start = 0
	}

	end := index + len(target) + 30
	if end > len(content) {
		end = len(content)
	}

	context := content[start:end]

	// 清理上下文，移除多余的空白字符
	context = strings.TrimSpace(context)
	context = strings.ReplaceAll(context, "\n", " ")
	context = strings.ReplaceAll(context, "\r", " ")
	context = strings.ReplaceAll(context, "\t", " ")

	// 如果上下文太长，进行截断
	if len(context) > TruncateContextLength {
		context = context[:TruncateContextLength] + "..."
	}

	return context
}

// searchInSubdirectoriesWithContext 子目录搜索方法（带上下文）
func (s *SmartFileSearch) searchInSubdirectoriesWithContext(ctx context.Context, target string, searchDir string) ([]SearchResult, error) {
	var results []SearchResult

	// 限制搜索深度
	maxDepth := MaxSearchDepth
	currentDepth := 0

	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		// 检查上下文是否被取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			return nil // 忽略错误，继续搜索
		}

		// 计算当前深度
		relPath, _ := filepath.Rel(searchDir, path)
		currentDepth = strings.Count(relPath, string(os.PathSeparator))

		// 如果超过最大深度，跳过该目录
		if info.IsDir() && currentDepth >= maxDepth {
			return filepath.SkipDir
		}

		// 跳过根目录本身
		if path == searchDir {
			return nil
		}

		// 只搜索当前目录下的直接子目录
		if currentDepth > 1 {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		filename := info.Name()
		similarity := s.calculateSimilarity(target, filename)

		// 降低子目录匹配的阈值要求
		subdirThreshold := s.config.SimilarityThreshold * SubdirMatchSimilarity
		if subdirThreshold < 30.0 {
			subdirThreshold = 30.0
		}

		if similarity >= subdirThreshold {
			matchType := s.getMatchType(target, filename)

			// 为子目录结果添加特殊标识
			context := ""
			if info.IsDir() {
				matchType = "subdir_" + matchType
				context = T("子目录")
			} else {
				context = s.getFileTypeContext(info)
			}

			results = append(results, SearchResult{
				Path:       path,
				Name:       filename,
				Similarity: similarity * SubdirMatchSimilarity,
				MatchType:  matchType,
				Context:    context,
			})

			// 限制结果数量
			if len(results) >= s.config.MaxResults {
				return filepath.SkipAll
			}
		}

		// 如果是目录且不启用递归，跳过该目录
		if info.IsDir() && !s.config.Recursive {
			return filepath.SkipDir
		}

		return nil
	})

	return results, err
}

// searchByContent 内容搜索方法
func (s *SmartFileSearch) searchByContent(target string, searchDir string) ([]SearchResult, error) {
	ctx := context.Background()
	return s.searchByContentWithContext(ctx, target, searchDir)
}

// searchInSubdirectories 子目录搜索方法
func (s *SmartFileSearch) searchInSubdirectories(target string, searchDir string) ([]SearchResult, error) {
	ctx := context.Background()
	return s.searchInSubdirectoriesWithContext(ctx, target, searchDir)
}

// removeDuplicates 去重方法
func (s *SmartFileSearch) removeDuplicates(results []SearchResult) []SearchResult {
	if len(results) <= 1 {
		return results
	}

	seen := make(map[string]bool)
	var unique []SearchResult

	for _, result := range results {
		if !seen[result.Path] {
			seen[result.Path] = true
			unique = append(unique, result)
		}
	}

	return unique
}

// sortAndLimitResults 排序和限制结果方法
func (s *SmartFileSearch) sortAndLimitResults(results []SearchResult) []SearchResult {
	if len(results) == 0 {
		return results
	}

	// 按相似度降序排序
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Similarity < results[j].Similarity {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// 限制结果数量到配置的最大值（默认10个）
	maxResults := s.config.MaxResults
	if maxResults <= 0 {
		maxResults = 10 // 确保默认值为10
	}

	if len(results) > maxResults {
		fmt.Printf(T("🔍 找到 %d 个相似文件，显示前 %d 个最相似的结果\n"), len(results), maxResults)
		results = results[:maxResults]
	}

	return results
}

// SearchByRegex 使用正则表达式搜索文件
func (s *SmartFileSearch) SearchByRegex(pattern string, searchDir string) ([]SearchResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.searchByRegexWithContext(ctx, pattern, searchDir)
}

// searchByRegexWithContext 使用正则表达式搜索文件（带上下文）
func (s *SmartFileSearch) searchByRegexWithContext(ctx context.Context, pattern string, searchDir string) ([]SearchResult, error) {
	var results []SearchResult

	// 编译正则表达式
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf(T("正则表达式编译失败: %v"), err)
	}

	// 限制搜索的文件大小
	maxFileSize := int64(10 * 1024 * 1024) // 10MB

	err = filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		// 检查上下文是否被取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			return nil // 忽略错误，继续搜索
		}

		// 跳过根目录本身
		if path == searchDir {
			return nil
		}

		filename := info.Name()

		// 在文件名中搜索
		if regex.MatchString(filename) {
			matchType := "regex_filename"
			context := ""

			if info.IsDir() {
				matchType = "dir_" + matchType
				context = T("目录")
			} else {
				context = s.getFileTypeContext(info)
			}

			results = append(results, SearchResult{
				Path:       path,
				Name:       filename,
				Similarity: 85.0, // 正则匹配给予较高相似度
				MatchType:  matchType,
				Context:    context,
			})
		}

		// 如果是文件且在文本文件列表中，搜索文件内容
		if !info.IsDir() && info.Size() <= maxFileSize {
			ext := strings.ToLower(filepath.Ext(path))
			isTextFile := false
			for _, textExt := range TextFileExtensions {
				if ext == textExt {
					isTextFile = true
					break
				}
			}

			if isTextFile {
				content, readErr := os.ReadFile(path)
				if readErr == nil {
					contentStr := string(content)
					matches := regex.FindAllStringIndex(contentStr, -1)

					if len(matches) > 0 {
						// 获取第一个匹配的上下文
						match := matches[0]
						start := match[0] - 30
						if start < 0 {
							start = 0
						}

						end := match[1] + 30
						if end > len(contentStr) {
							end = len(contentStr)
						}

						context := strings.TrimSpace(contentStr[start:end])
						context = strings.ReplaceAll(context, "\n", " ")
						context = strings.ReplaceAll(context, "\r", " ")
						context = strings.ReplaceAll(context, "\t", " ")

						if len(context) > TruncateContextLength {
							context = context[:TruncateContextLength] + "..."
						}

						results = append(results, SearchResult{
							Path:       path,
							Name:       info.Name(),
							Similarity: 80.0 + float64(len(matches))*2.0,
							MatchType:  "regex_content",
							Context:    context,
						})
					}
				}
			}
		}

		// 限制结果数量
		if len(results) >= s.config.MaxResults*3 {
			return filepath.SkipAll
		}

		// 如果是目录且不启用递归，跳过该目录
		if info.IsDir() && !s.config.Recursive {
			return filepath.SkipDir
		}

		return nil
	})

	return results, err
}

// FilterByExtension 按文件扩展名过滤结果
func (s *SmartFileSearch) FilterByExtension(results []SearchResult, extensions []string) []SearchResult {
	if len(extensions) == 0 {
		return results
	}

	var filtered []SearchResult
	extMap := make(map[string]bool)

	// 将扩展名转换为小写并存储在map中
	for _, ext := range extensions {
		extMap[strings.ToLower(ext)] = true
	}

	for _, result := range results {
		fileExt := strings.ToLower(filepath.Ext(result.Path))
		if extMap[fileExt] {
			filtered = append(filtered, result)
		}
	}

	return filtered
}

// FilterBySize 按文件大小过滤结果
func (s *SmartFileSearch) FilterBySize(results []SearchResult, minSize, maxSize int64) []SearchResult {
	if minSize <= 0 && maxSize <= 0 {
		return results
	}

	var filtered []SearchResult

	for _, result := range results {
		info, err := os.Stat(result.Path)
		if err != nil {
			continue
		}

		if info.IsDir() {
			continue
		}

		size := info.Size()
		if (minSize <= 0 || size >= minSize) && (maxSize <= 0 || size <= maxSize) {
			filtered = append(filtered, result)
		}
	}

	return filtered
}

// FilterByDate 按文件修改日期过滤结果
func (s *SmartFileSearch) FilterByDate(results []SearchResult, after, before time.Time) []SearchResult {
	if after.IsZero() && before.IsZero() {
		return results
	}

	var filtered []SearchResult

	for _, result := range results {
		info, err := os.Stat(result.Path)
		if err != nil {
			continue
		}

		modTime := info.ModTime()
		if (after.IsZero() || modTime.After(after)) && (before.IsZero() || modTime.Before(before)) {
			filtered = append(filtered, result)
		}
	}

	return filtered
}

// GroupResultsByDirectory 按目录分组结果
func (s *SmartFileSearch) GroupResultsByDirectory(results []SearchResult) map[string][]SearchResult {
	groups := make(map[string][]SearchResult)

	for _, result := range results {
		dir := filepath.Dir(result.Path)
		groups[dir] = append(groups[dir], result)
	}

	return groups
}

// GetSearchStats 获取搜索统计信息
func (s *SmartFileSearch) GetSearchStats(results []SearchResult) map[string]interface{} {
	stats := make(map[string]interface{})

	if len(results) == 0 {
		stats["total_files"] = 0
		stats["directories"] = 0
		stats["avg_similarity"] = 0.0
		return stats
	}

	// 统计信息
	fileCount := 0
	dirCount := 0
	totalSimilarity := 0.0
	fileTypes := make(map[string]int)

	for _, result := range results {
		info, err := os.Stat(result.Path)
		if err != nil {
			continue
		}

		if info.IsDir() {
			dirCount++
		} else {
			fileCount++

			// 统计文件类型
			ext := strings.ToLower(filepath.Ext(result.Path))
			if ext == "" {
				ext = "no_extension"
			}
			fileTypes[ext]++
		}

		totalSimilarity += result.Similarity
	}

	avgSimilarity := totalSimilarity / float64(len(results))

	stats["total_results"] = len(results)
	stats["files"] = fileCount
	stats["directories"] = dirCount
	stats["avg_similarity"] = avgSimilarity
	stats["file_types"] = fileTypes

	return stats
}
