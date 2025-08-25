package main

import (
	"strings"
)

// SearchResult 搜索结果结构体
type SearchResult struct {
	Path       string  // 文件路径
	Name       string  // 文件名
	Similarity float64 // 相似度 (0-100)
	MatchType  string  // 匹配类型: exact, filename, content, regex, parent_filename, parent_content, subdir_filename, subdir_content
	Context    string  // 匹配上下文（用于内容匹配时显示匹配的行）
}

// SmartSearchConfig 智能搜索配置
type SmartSearchConfig struct {
	SimilarityThreshold float64 // 相似度阈值
	MaxResults          int     // 最大结果数量
	SearchContent       bool    // 是否搜索文件内容
	Recursive           bool    // 是否递归搜索
	SearchParent        bool    // 是否搜索父目录
}

// SmartFileSearch 智能文件搜索引擎
type SmartFileSearch struct {
	config SmartSearchConfig
}

// InteractiveUI 交互式用户界面
type InteractiveUI struct {
}

// 搜索相似度常量
const (
	ExactMatchSimilarity              = 100.0 // 精确匹配
	FilenameMatchSimilarity           = 85.0  // 文件名匹配相似度
	ParentMatchSimilarity             = 0.8   // 父目录匹配相似度系数
	EnhancedContentBaseSimilarity     = 70.0  // 增强内容搜索基础相似度
	EnhancedContentSimilarityPerMatch = 5     // 每个匹配增加的相似度
)

// 搜索限制常量
const (
	MaxEnhancedMatches    = 10 // 增强搜索最大匹配数
	TruncateContextLength = 80 // 上下文截断长度
)

// NewSmartFileSearch 创建智能文件搜索引擎
func NewSmartFileSearch(config SmartSearchConfig) *SmartFileSearch {
	return &SmartFileSearch{
		config: config,
	}
}

// NewInteractiveUI 创建交互式用户界面
func NewInteractiveUI() *InteractiveUI {
	return &InteractiveUI{}
}

// SearchFiles 搜索文件
func (s *SmartFileSearch) SearchFiles(target string, searchDir string) ([]SearchResult, error) {
	var allResults []SearchResult

	// 1. 按相似度搜索
	results, err := s.searchBySimilarity(target, searchDir)
	if err == nil {
		allResults = append(allResults, results...)
	}

	// 2. 如果启用内容搜索
	if s.config.SearchContent {
		contentResults, err := s.searchByContent(target, searchDir)
		if err == nil {
			allResults = append(allResults, contentResults...)
		}
	}

	// 3. 如果启用父目录搜索
	if s.config.SearchParent {
		parentResults, err := s.searchInParentDirectories(target, searchDir)
		if err == nil {
			allResults = append(allResults, parentResults...)
		}
	}

	// 4. 搜索子目录
	if s.config.Recursive {
		subdirResults, err := s.searchInSubdirectories(target, searchDir)
		if err == nil {
			allResults = append(allResults, subdirResults...)
		}
	}

	// 排序并限制结果
	return s.sortAndLimitResults(allResults), nil
}

// searchBySimilarity 按相似度搜索文件
func (s *SmartFileSearch) searchBySimilarity(target string, searchDir string) ([]SearchResult, error) {
	var results []SearchResult

	// 这里应该实现相似度搜索逻辑
	// 为了简化，暂时返回空结果
	return results, nil
}

// searchInParentDirectories 在父目录中搜索
func (s *SmartFileSearch) searchInParentDirectories(target string, searchDir string) ([]SearchResult, error) {
	var results []SearchResult

	// 这里应该实现父目录搜索逻辑
	// 为了简化，暂时返回空结果
	return results, nil
}

// isRegexPattern 检查是否为正则表达式模式
func isRegexPattern(pattern string) bool {
	// 检查是否包含正则表达式特殊字符
	regexChars := []string{"*", "?", "[", "]", "^", "$", "+", "{", "}", "|", "(", ")", "\\"}
	for _, char := range regexChars {
		if strings.Contains(pattern, char) {
			return true
		}
	}
	return false
}
