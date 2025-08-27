package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

// SearchTool 搜索工具结构体
type SearchTool struct {
	config     SearchToolConfig
	search     *EnhancedSmartSearch
	ui         *InteractiveUI
	searchDir  string
	outputMode string
	verbose    bool
}

// NewSearchTool 创建新的搜索工具
func NewSearchTool() *SearchTool {
	return &SearchTool{
		ui: NewInteractiveUI(),
	}
}

// Run 运行搜索工具
func (st *SearchTool) Run(args []string) error {
	// 解析命令行参数
	if err := st.parseArgs(args); err != nil {
		return err
	}

	// 验证搜索目录
	if st.searchDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("无法获取当前目录: %v", err)
		}
		st.searchDir = cwd
	}

	// 检查目录是否存在
	if _, err := os.Stat(st.searchDir); err != nil {
		return fmt.Errorf("搜索目录不存在: %s", st.searchDir)
	}

	// 创建搜索引擎
	config := SmartSearchConfig{
		SimilarityThreshold: st.config.SimilarityThreshold,
		MaxResults:          st.config.MaxResults,
		SearchContent:       st.config.SearchContent,
		Recursive:           st.config.Recursive,
		SearchParent:        st.config.SearchParent,
		CaseSensitive:       st.config.CaseSensitive,
	}
	st.search = NewEnhancedSmartSearch(config)

	// 显示搜索信息
	if st.verbose {
		st.printSearchInfo()
	}

	// 执行搜索
	return st.performSearch()
}

// parseArgs 解析命令行参数
func (st *SearchTool) parseArgs(args []string) error {
	fs := flag.NewFlagSet("search", flag.ExitOnError)

	// 搜索配置
	fs.Float64Var(&st.config.SimilarityThreshold, "threshold", DefaultSimilarityThreshold, "相似度阈值 (0-100)")
	fs.IntVar(&st.config.MaxResults, "max-results", DefaultMaxResults, "最大结果数量")
	fs.BoolVar(&st.config.SearchContent, "content", false, "搜索文件内容")
	fs.BoolVar(&st.config.Recursive, "recursive", false, "递归搜索子目录")
	fs.BoolVar(&st.config.SearchParent, "parent", false, "搜索父目录")
	fs.BoolVar(&st.config.CaseSensitive, "case-sensitive", false, "区分大小写")

	// 工具配置
	fs.StringVar(&st.searchDir, "dir", "", "搜索目录 (默认为当前目录)")
	fs.StringVar(&st.outputMode, "output", "table", "输出模式: table, json, csv, list")
	fs.BoolVar(&st.verbose, "verbose", false, "显示详细信息")

	// 过滤选项
	var extensions string
	var minSize, maxSize int64
	var afterDate, beforeDate string

	fs.StringVar(&extensions, "ext", "", "文件扩展名过滤 (逗号分隔)")
	fs.Int64Var(&minSize, "min-size", 0, "最小文件大小 (字节)")
	fs.Int64Var(&maxSize, "max-size", 0, "最大文件大小 (字节)")
	fs.StringVar(&afterDate, "after", "", "修改日期之后 (格式: 2006-01-02)")
	fs.StringVar(&beforeDate, "before", "", "修改日期之前 (格式: 2006-01-02)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// 检查必需参数
	if fs.NArg() == 0 {
		return fmt.Errorf("请提供搜索目标")
	}

	// 解析扩展名
	if extensions != "" {
		extList := strings.Split(extensions, ",")
		for i, ext := range extList {
			extList[i] = strings.TrimSpace(ext)
			if !strings.HasPrefix(extList[i], ".") {
				extList[i] = "." + extList[i]
			}
		}
		st.config.extensions = extList
	}

	// 解析日期
	var afterTime, beforeTime time.Time
	if afterDate != "" {
		t, err := time.Parse("2006-01-02", afterDate)
		if err != nil {
			return fmt.Errorf("无效的日期格式: %s", afterDate)
		}
		afterTime = t
	}
	if beforeDate != "" {
		t, err := time.Parse("2006-01-02", beforeDate)
		if err != nil {
			return fmt.Errorf("无效的日期格式: %s", beforeDate)
		}
		beforeTime = t
	}

	// 存储过滤条件
	st.config.afterDate = afterTime
	st.config.beforeDate = beforeTime
	st.config.minSize = minSize
	st.config.maxSize = maxSize

	return nil
}

// printSearchInfo 打印搜索信息
func (st *SearchTool) printSearchInfo() {
	fmt.Printf("搜索工具配置:\n")
	fmt.Printf("  搜索目录: %s\n", st.searchDir)
	fmt.Printf("  相似度阈值: %.1f%%\n", st.config.SimilarityThreshold)
	fmt.Printf("  最大结果数: %d\n", st.config.MaxResults)
	fmt.Printf("  搜索内容: %t\n", st.config.SearchContent)
	fmt.Printf("  递归搜索: %t\n", st.config.Recursive)
	fmt.Printf("  搜索父目录: %t\n", st.config.SearchParent)
	fmt.Printf("  区分大小写: %t\n", st.config.CaseSensitive)

	if st.config.extensions != nil {
		fmt.Printf("  文件扩展名: %s\n", strings.Join(st.config.extensions, ", "))
	}
	if st.config.minSize > 0 {
		fmt.Printf("  最小文件大小: %d bytes\n", st.config.minSize)
	}
	if st.config.maxSize > 0 {
		fmt.Printf("  最大文件大小: %d bytes\n", st.config.maxSize)
	}
	if !st.config.afterDate.IsZero() {
		fmt.Printf("  修改日期之后: %s\n", st.config.afterDate.Format("2006-01-02"))
	}
	if !st.config.beforeDate.IsZero() {
		fmt.Printf("  修改日期之前: %s\n", st.config.beforeDate.Format("2006-01-02"))
	}
	fmt.Println()
}

// performSearch 执行搜索
func (st *SearchTool) performSearch() error {
	target := flag.Arg(0)

	// 检查是否为正则表达式
	var results []SearchResult
	var err error

	if isRegexPattern(target) {
		fmt.Printf("使用正则表达式搜索: %s\n", target)
		results, err = st.search.SearchRegexWithCache(target, st.searchDir)
	} else {
		fmt.Printf("搜索目标: %s\n", target)
		results, err = st.search.SearchWithCache(target, st.searchDir)

		// 如果基本搜索没找到且启用了内容搜索，尝试内容搜索
		if err == nil && len(results) == 0 && st.config.SearchContent {
			fmt.Printf("未找到文件名匹配，正在搜索文件内容...\n")
			results, err = st.search.SearchContentWithCache(target, st.searchDir)
		}
	}

	if err != nil {
		return fmt.Errorf("搜索失败: %v", err)
	}

	// 应用过滤条件
	if len(st.config.extensions) > 0 {
		results = st.search.FilterByExtension(results, st.config.extensions)
	}
	// 应用文件大小过滤
	if st.config.minSize > 0 || st.config.maxSize > 0 {
		results = st.search.FilterBySize(results, st.config.minSize, st.config.maxSize)
	}
	if !st.config.afterDate.IsZero() || !st.config.beforeDate.IsZero() {
		results = st.search.FilterByDate(results, st.config.afterDate, st.config.beforeDate)
	}

	// 显示结果
	return st.displayResults(results, target)
}

// displayResults 显示搜索结果
func (st *SearchTool) displayResults(results []SearchResult, target string) error {
	if len(results) == 0 {
		fmt.Printf("未找到与 '%s' 匹配的文件\n", target)
		return nil
	}

	// 显示统计信息
	stats := st.search.GetSearchStats(results)
	fmt.Printf("找到 %d 个结果:\n", stats["total_results"])
	if st.verbose {
		fmt.Printf("  文件: %d\n", stats["files"])
		fmt.Printf("  目录: %d\n", stats["directories"])
		fmt.Printf("  平均相似度: %.1f%%\n", stats["avg_similarity"])

		if fileTypes, ok := stats["file_types"].(map[string]int); ok && len(fileTypes) > 0 {
			fmt.Printf("  文件类型分布:\n")
			for ext, count := range fileTypes {
				fmt.Printf("    %s: %d\n", ext, count)
			}
		}
		fmt.Println()
	}

	// 根据输出模式显示结果
	switch st.outputMode {
	case "json":
		return st.displayJSON(results)
	case "csv":
		return st.displayCSV(results)
	case "list":
		return st.displayList(results)
	default:
		return st.displayTable(results)
	}
}

// displayTable 以表格形式显示结果
func (st *SearchTool) displayTable(results []SearchResult) error {
	fmt.Printf("%-60s %-20s %-10s %-15s %s\n", "路径", "名称", "相似度", "匹配类型", "上下文")
	fmt.Printf("%-60s %-20s %-10s %-15s %s\n", strings.Repeat("-", 60), strings.Repeat("-", 20), strings.Repeat("-", 10), strings.Repeat("-", 15), strings.Repeat("-", 30))

	for _, result := range results {
		path := result.Path
		if len(path) > 57 {
			path = "..." + path[len(path)-54:]
		}

		name := result.Name
		if len(name) > 17 {
			name = name[:17] + "..."
		}

		context := result.Context
		if len(context) > 30 {
			context = context[:27] + "..."
		}

		fmt.Printf("%-60s %-20s %-10.1f %-15s %s\n", path, name, result.Similarity, result.MatchType, context)
	}

	return nil
}

// displayList 以列表形式显示结果
func (st *SearchTool) displayList(results []SearchResult) error {
	for i, result := range results {
		fmt.Printf("%d. %s\n", i+1, result.Path)
		fmt.Printf("   名称: %s\n", result.Name)
		fmt.Printf("   相似度: %.1f%%\n", result.Similarity)
		fmt.Printf("   匹配类型: %s\n", result.MatchType)
		if result.Context != "" {
			fmt.Printf("   上下文: %s\n", result.Context)
		}
		fmt.Println()
	}
	return nil
}

// displayJSON 以JSON形式显示结果（简化实现）
func (st *SearchTool) displayJSON(results []SearchResult) error {
	fmt.Printf("{\n")
	fmt.Printf("  \"results\": [\n")
	for i, result := range results {
		fmt.Printf("    {\n")
		fmt.Printf("      \"path\": \"%s\",\n", escapeJSON(result.Path))
		fmt.Printf("      \"name\": \"%s\",\n", escapeJSON(result.Name))
		fmt.Printf("      \"similarity\": %.1f,\n", result.Similarity)
		fmt.Printf("      \"matchType\": \"%s\",\n", result.MatchType)
		if result.Context != "" {
			fmt.Printf("      \"context\": \"%s\"\n", escapeJSON(result.Context))
		} else {
			fmt.Printf("      \"context\": \"\"\n")
		}
		if i < len(results)-1 {
			fmt.Printf("    },\n")
		} else {
			fmt.Printf("    }\n")
		}
	}
	fmt.Printf("  ]\n")
	fmt.Printf("}\n")
	return nil
}

// displayCSV 以CSV形式显示结果
func (st *SearchTool) displayCSV(results []SearchResult) error {
	fmt.Printf("路径,名称,相似度,匹配类型,上下文\n")
	for _, result := range results {
		fmt.Printf("\"%s\",\"%s\",%.1f,\"%s\",\"%s\"\n",
			escapeCSV(result.Path),
			escapeCSV(result.Name),
			result.Similarity,
			result.MatchType,
			escapeCSV(result.Context))
	}
	return nil
}

// escapeJSON 转义JSON字符串
func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

// escapeCSV 转义CSV字符串
func escapeCSV(s string) string {
	if strings.Contains(s, "\"") || strings.Contains(s, ",") || strings.Contains(s, "\n") {
		return "\"" + strings.ReplaceAll(s, "\"", "\"\"") + "\""
	}
	return s
}

// printUsage 打印使用帮助
func (st *SearchTool) printUsage() {
	fmt.Printf(`智能文件搜索工具

使用方法:
  search [选项] <搜索目标>

参数:
  搜索目标    要搜索的文件名、路径或正则表达式

选项:
  -threshold float    相似度阈值 (0-100, 默认: %.1f)
  -max-results int    最大结果数量 (默认: %d)
  -content           搜索文件内容
  -recursive         递归搜索子目录
  -parent            搜索父目录
  -case-sensitive    区分大小写
  -dir string        搜索目录 (默认为当前目录)
  -output string     输出模式: table, json, csv, list (默认: table)
  -ext string        文件扩展名过滤 (逗号分隔, 如: .txt,.md,.go)
  -min-size int      最小文件大小 (字节)
  -max-size int      最大文件大小 (字节)
  -after string      修改日期之后 (格式: 2006-01-02)
  -before string     修改日期之前 (格式: 2006-01-02)
  -verbose           显示详细信息

示例:
  search myfile.txt
  search -content -recursive "hello world"
  search -ext .go -max-results 5 "main"
  search -output json "*.log"
  search -dir /path/to/dir -after 2024-01-01 "config"

正则表达式搜索:
  使用包含正则表达式字符的搜索目标会自动启用正则搜索
  例如: search "file.*\.txt" 或 search "[0-9]+_backup"
`, DefaultSimilarityThreshold, DefaultMaxResults)
}

// SearchToolConfig 扩展配置
type SearchToolConfig struct {
	SimilarityThreshold   float64
	MaxResults            int
	SearchContent         bool
	Recursive             bool
	SearchParent          bool
	CaseSensitive         bool
	extensions            []string
	minSize, maxSize      int64
	afterDate, beforeDate time.Time
}
