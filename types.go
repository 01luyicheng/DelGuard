package main

import (
	"strings"
)

// CommandMode 命令模式类型
type CommandMode int

const (
	ModeDel     CommandMode = iota // del命令模式
	ModeRM                         // rm命令模式
	ModeCP                         // cp命令模式
	ModeDefault                    // 默认delguard模式
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
	CaseSensitive       bool    // 是否区分大小写
}

// SmartFileSearch 智能文件搜索引擎
type SmartFileSearch struct {
	config SmartSearchConfig
}

// InteractiveUI 交互式用户界面
type InteractiveUI struct {
}

// DiskUsage 磁盘使用情况（跨平台定义）
type DiskUsage struct {
	Total uint64 // 总空间（字节）
	Free  uint64 // 可用空间（字节）
	Used  uint64 // 已用空间（字节）
}

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
