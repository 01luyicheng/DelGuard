package search

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// SearchFilter 搜索过滤器
type SearchFilter struct {
	NamePattern   string
	PathPattern   string
	Extension     string
	MinSize       int64
	MaxSize       int64
	ModifiedAfter *time.Time
	ModifiedBefore *time.Time
	UseRegex      bool
	FuzzyMatch    bool
	CaseSensitive bool
}

// EnhancedResult 增强搜索结果
type EnhancedResult struct {
	Path        string
	Size        int64
	ModTime     time.Time
	Score       float64
	MatchType   string
}

// EnhancedSearchEngine 增强搜索引擎
type EnhancedSearchEngine struct {
	maxResults int
}

// NewEnhancedSearchEngine 创建增强搜索引擎
func NewEnhancedSearchEngine() *EnhancedSearchEngine {
	return &EnhancedSearchEngine{
		maxResults: 1000,
	}
}

// MultiSearch 多条件搜索
func (e *EnhancedSearchEngine) MultiSearch(ctx context.Context, rootPath string, filter SearchFilter) ([]EnhancedResult, error) {
	var results []EnhancedResult
	
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		if err != nil {
			return nil
		}
		
		if info.IsDir() {
			return nil
		}
		
		if match, score, matchType := e.evaluateFile(path, info, filter); match {
			result := EnhancedResult{
				Path:      path,
				Size:      info.Size(),
				ModTime:   info.ModTime(),
				Score:     score,
				MatchType: matchType,
			}
			results = append(results, result)
			
			if len(results) >= e.maxResults {
				return fmt.Errorf("max results reached")
			}
		}
		
		return nil
	})
	
	if err != nil && err.Error() != "max results reached" {
		return nil, err
	}
	
	// 按分数排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	
	return results, nil
}

// evaluateFile 评估文件是否匹配条件
func (e *EnhancedSearchEngine) evaluateFile(path string, info os.FileInfo, filter SearchFilter) (bool, float64, string) {
	var score float64
	var matchTypes []string
	
	fileName := filepath.Base(path)
	
	// 文件名匹配
	if filter.NamePattern != "" {
		if match, nameScore := e.matchPattern(fileName, filter.NamePattern, filter); match {
			score += nameScore * 0.5
			matchTypes = append(matchTypes, "name")
		} else {
			return false, 0, ""
		}
	}
	
	// 路径匹配
	if filter.PathPattern != "" {
		if match, pathScore := e.matchPattern(path, filter.PathPattern, filter); match {
			score += pathScore * 0.3
			matchTypes = append(matchTypes, "path")
		} else {
			return false, 0, ""
		}
	}
	
	// 扩展名匹配
	if filter.Extension != "" {
		ext := strings.ToLower(filepath.Ext(fileName))
		targetExt := strings.ToLower(filter.Extension)
		if !strings.HasPrefix(targetExt, ".") {
			targetExt = "." + targetExt
		}
		
		if ext == targetExt {
			score += 0.2
			matchTypes = append(matchTypes, "extension")
		} else {
			return false, 0, ""
		}
	}
	
	// 大小过滤
	if filter.MinSize > 0 && info.Size() < filter.MinSize {
		return false, 0, ""
	}
	if filter.MaxSize > 0 && info.Size() > filter.MaxSize {
		return false, 0, ""
	}
	
	// 时间过滤
	if filter.ModifiedAfter != nil && info.ModTime().Before(*filter.ModifiedAfter) {
		return false, 0, ""
	}
	if filter.ModifiedBefore != nil && info.ModTime().After(*filter.ModifiedBefore) {
		return false, 0, ""
	}
	
	if score == 0 {
		score = 0.1
		matchTypes = append(matchTypes, "basic")
	}
	
	return true, score, strings.Join(matchTypes, ",")
}

// matchPattern 模式匹配
func (e *EnhancedSearchEngine) matchPattern(text, pattern string, filter SearchFilter) (bool, float64) {
	if !filter.CaseSensitive {
		text = strings.ToLower(text)
		pattern = strings.ToLower(pattern)
	}
	
	if filter.UseRegex {
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return false, 0
		}
		return regex.MatchString(text), 1.0
	}
	
	if filter.FuzzyMatch {
		similarity := calculateSimilarity(text, pattern)
		return similarity >= SimilarityThreshold, similarity
	}
	
	// 通配符匹配
	matched, err := filepath.Match(pattern, text)
	if err == nil && matched {
		return true, 1.0
	}
	
	// 子字符串匹配
	if strings.Contains(text, pattern) {
		return true, 0.8
	}
	
	return false, 0
}

// QuickSearch 快速搜索
func (e *EnhancedSearchEngine) QuickSearch(rootPath, query string) ([]string, error) {
	var results []string
	
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		
		fileName := strings.ToLower(filepath.Base(path))
		queryLower := strings.ToLower(query)
		
		if strings.Contains(fileName, queryLower) {
			results = append(results, path)
			if len(results) >= 50 { // 快速搜索限制结果数
				return fmt.Errorf("quick search limit reached")
			}
		}
		
		return nil
	})
	
	if err != nil && err.Error() != "quick search limit reached" {
		return nil, err
	}
	
	return results, nil
}