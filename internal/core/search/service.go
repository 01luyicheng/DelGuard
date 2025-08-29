package search

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/01luyicheng/DelGuard/internal/config"
)

// Service 搜索服务
type Service struct {
	config        *config.Config
	index         *Index
	enhancedIndex *EnhancedIndex
}

// SearchResult 搜索结果
type SearchResult struct {
	Path        string    `json:"path"`
	Name        string    `json:"name"`
	Size        int64     `json:"size"`
	ModTime     time.Time `json:"modTime"`
	IsDir       bool      `json:"isDir"`
	Type        string    `json:"type"`
	Permissions string    `json:"permissions"`
	Owner       string    `json:"owner"`
	MatchReason string    `json:"matchReason"`
	Score       float64   `json:"score"`
}

// SearchOptions 搜索选项
type SearchOptions struct {
	Pattern        string
	Directory      string
	Recursive      bool
	CaseSensitive  bool
	UseRegex       bool
	FileType       string
	MinSize        int64
	MaxSize        int64
	ModifiedAfter  time.Time
	ModifiedBefore time.Time
	MaxResults     int
	IncludeHidden  bool
}

// NewService 创建搜索服务
func NewService(cfg *config.Config) *Service {
	service := &Service{
		config: cfg,
		index:  NewIndex(cfg),
	}

	// 初始化增强索引
	indexPath := ""
	if cfg.Search.IndexPath != "" {
		indexPath = cfg.Search.IndexPath
	}
	service.enhancedIndex = NewEnhancedIndex(indexPath)

	// 尝试加载现有索引
	service.enhancedIndex.Load()

	return service
}

// Execute 执行搜索命令
func (s *Service) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("请指定搜索模式")
	}

	// 解析搜索选项
	options, err := s.parseSearchArgs(args)
	if err != nil {
		return err
	}

	// 执行搜索
	results, err := s.Search(ctx, options)
	if err != nil {
		return err
	}

	// 显示结果
	s.displayResults(results, options)
	return nil
}

// parseSearchArgs 解析搜索参数
func (s *Service) parseSearchArgs(args []string) (*SearchOptions, error) {
	options := &SearchOptions{
		Directory:     ".",
		Recursive:     true,
		CaseSensitive: false, // 默认不区分大小写
		UseRegex:      false, // 默认不使用正则
		MaxResults:    100,   // 默认最大结果数
		IncludeHidden: false,
	}

	// 如果只有一个参数，将其作为搜索模式
	if len(args) == 1 && !strings.HasPrefix(args[0], "-") {
		options.Pattern = args[0]
		return options, nil
	}

	// 如果有两个参数且都不以-开头，第一个是模式，第二个是目录
	if len(args) == 2 && !strings.HasPrefix(args[0], "-") && !strings.HasPrefix(args[1], "-") {
		options.Pattern = args[0]
		options.Directory = args[1]
		return options, nil
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-n", "--name":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--name 需要指定模式")
			}
			options.Pattern = args[i+1]
			i++
		case "-d", "--directory":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--directory 需要指定目录")
			}
			options.Directory = args[i+1]
			i++
		case "-t", "--type":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--type 需要指定文件类型")
			}
			options.FileType = args[i+1]
			i++
		case "-s", "--size":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--size 需要指定大小条件")
			}
			if err := s.parseSizeCondition(args[i+1], options); err != nil {
				return nil, err
			}
			i++
		case "--max":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--max 需要指定数量")
			}
			max, err := strconv.Atoi(args[i+1])
			if err != nil {
				return nil, fmt.Errorf("无效的最大结果数: %s", args[i+1])
			}
			options.MaxResults = max
			i++
		case "-r", "--recursive":
			options.Recursive = true
		case "--no-recursive":
			options.Recursive = false
		case "-i", "--ignore-case":
			options.CaseSensitive = false
		case "-c", "--case-sensitive":
			options.CaseSensitive = true
		case "--regex":
			options.UseRegex = true
		case "--no-regex":
			options.UseRegex = false
		case "-a", "--all":
			options.IncludeHidden = true
		default:
			if !strings.HasPrefix(arg, "-") {
				if options.Pattern == "" {
					options.Pattern = arg
				} else {
					options.Directory = arg
				}
			} else {
				return nil, fmt.Errorf("未知参数: %s", arg)
			}
		}
	}

	if options.Pattern == "" {
		return nil, fmt.Errorf("请指定搜索模式")
	}

	return options, nil
}

// parseSizeCondition 解析大小条件
func (s *Service) parseSizeCondition(condition string, options *SearchOptions) error {
	// 支持格式: >1MB, <500KB, =1GB, 100-200MB
	if strings.Contains(condition, "-") {
		// 范围格式: 100-200MB
		parts := strings.Split(condition, "-")
		if len(parts) != 2 {
			return fmt.Errorf("无效的大小范围格式: %s", condition)
		}

		minSize, err := s.parseSize(parts[0])
		if err != nil {
			return err
		}
		maxSize, err := s.parseSize(parts[1])
		if err != nil {
			return err
		}

		options.MinSize = minSize
		options.MaxSize = maxSize
	} else if strings.HasPrefix(condition, ">") {
		// 大于格式: >1MB
		size, err := s.parseSize(condition[1:])
		if err != nil {
			return err
		}
		options.MinSize = size
	} else if strings.HasPrefix(condition, "<") {
		// 小于格式: <500KB
		size, err := s.parseSize(condition[1:])
		if err != nil {
			return err
		}
		options.MaxSize = size
	} else if strings.HasPrefix(condition, "=") {
		// 等于格式: =1GB
		size, err := s.parseSize(condition[1:])
		if err != nil {
			return err
		}
		options.MinSize = size
		options.MaxSize = size
	} else {
		// 直接大小: 1MB
		size, err := s.parseSize(condition)
		if err != nil {
			return err
		}
		options.MinSize = size
		options.MaxSize = size
	}

	return nil
}

// parseSize 解析大小字符串
func (s *Service) parseSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(strings.ToUpper(sizeStr))

	var multiplier int64 = 1
	var numStr string

	if strings.HasSuffix(sizeStr, "KB") {
		multiplier = 1024
		numStr = strings.TrimSuffix(sizeStr, "KB")
	} else if strings.HasSuffix(sizeStr, "MB") {
		multiplier = 1024 * 1024
		numStr = strings.TrimSuffix(sizeStr, "MB")
	} else if strings.HasSuffix(sizeStr, "GB") {
		multiplier = 1024 * 1024 * 1024
		numStr = strings.TrimSuffix(sizeStr, "GB")
	} else if strings.HasSuffix(sizeStr, "TB") {
		multiplier = 1024 * 1024 * 1024 * 1024
		numStr = strings.TrimSuffix(sizeStr, "TB")
	} else if strings.HasSuffix(sizeStr, "B") {
		multiplier = 1
		numStr = strings.TrimSuffix(sizeStr, "B")
	} else {
		numStr = sizeStr
	}

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("无效的大小格式: %s", sizeStr)
	}

	return int64(num * float64(multiplier)), nil
}

// Search 执行搜索
func (s *Service) Search(ctx context.Context, options *SearchOptions) ([]*SearchResult, error) {
	var results []*SearchResult

	// 编译正则表达式（如果需要）
	var pattern *regexp.Regexp
	var err error

	if options.UseRegex {
		regexPattern := options.Pattern
		if !options.CaseSensitive {
			regexPattern = "(?i)" + regexPattern
		}
		pattern, err = regexp.Compile(regexPattern)
		if err != nil {
			return nil, fmt.Errorf("无效的正则表达式: %v", err)
		}
	}

	// 遍历目录
	err = s.walkDirectory(ctx, options.Directory, options, pattern, &results)
	if err != nil {
		return nil, err
	}

	// 限制结果数量
	if len(results) > options.MaxResults {
		results = results[:options.MaxResults]
	}

	return results, nil
}

// walkDirectory 遍历目录
func (s *Service) walkDirectory(ctx context.Context, dir string, options *SearchOptions, pattern *regexp.Regexp, results *[]*SearchResult) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// 检查上下文是否被取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			return nil // 忽略错误，继续遍历
		}

		// 检查是否包含隐藏文件
		if !options.IncludeHidden && s.isHidden(path, info) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 检查递归选项
		if !options.Recursive && info.IsDir() && path != options.Directory {
			return filepath.SkipDir
		}

		// 检查文件是否匹配
		if s.matchesPattern(path, info, options, pattern) {
			result := &SearchResult{
				Path:        path,
				Name:        info.Name(),
				Size:        info.Size(),
				ModTime:     info.ModTime(),
				IsDir:       info.IsDir(),
				Type:        s.getFileType(path, info),
				Permissions: info.Mode().String(),
				Owner:       s.getFileOwner(path, info),
				MatchReason: s.getMatchReason(path, info, options, pattern),
				Score:       s.calculateScore(path, info, options),
			}
			*results = append(*results, result)
		}

		return nil
	})
}

// matchesPattern 检查文件是否匹配模式
func (s *Service) matchesPattern(path string, info os.FileInfo, options *SearchOptions, pattern *regexp.Regexp) bool {
	name := info.Name()

	// 检查文件类型
	if options.FileType != "" {
		fileType := s.getFileType(path, info)
		if !strings.EqualFold(fileType, options.FileType) {
			return false
		}
	}

	// 检查文件大小
	if options.MinSize > 0 && info.Size() < options.MinSize {
		return false
	}
	if options.MaxSize > 0 && info.Size() > options.MaxSize {
		return false
	}

	// 检查修改时间
	if !options.ModifiedAfter.IsZero() && info.ModTime().Before(options.ModifiedAfter) {
		return false
	}
	if !options.ModifiedBefore.IsZero() && info.ModTime().After(options.ModifiedBefore) {
		return false
	}

	// 如果没有指定模式，匹配所有文件
	if options.Pattern == "" {
		return true
	}

	// 检查名称模式
	if options.UseRegex && pattern != nil {
		return pattern.MatchString(name) || pattern.MatchString(path)
	} else {
		// 1. 通配符匹配（优先级最高）
		matched, err := filepath.Match(options.Pattern, name)
		if err == nil && matched {
			return true
		}

		// 2. 完整路径通配符匹配
		matched, err = filepath.Match(options.Pattern, path)
		if err == nil && matched {
			return true
		}

		// 3. 子字符串匹配
		searchName := name
		searchPath := path
		searchPattern := options.Pattern

		if !options.CaseSensitive {
			searchName = strings.ToLower(name)
			searchPath = strings.ToLower(path)
			searchPattern = strings.ToLower(options.Pattern)
		}

		// 在文件名中搜索
		if strings.Contains(searchName, searchPattern) {
			return true
		}

		// 在完整路径中搜索
		if strings.Contains(searchPath, searchPattern) {
			return true
		}

		// 4. 模糊匹配（去掉扩展名后匹配）
		nameWithoutExt := strings.TrimSuffix(searchName, filepath.Ext(searchName))
		if strings.Contains(nameWithoutExt, searchPattern) {
			return true
		}
	}

	return false
}

// isHidden 检查文件是否为隐藏文件
func (s *Service) isHidden(path string, info os.FileInfo) bool {
	name := info.Name()
	return strings.HasPrefix(name, ".")
}

// getFileType 获取文件类型
func (s *Service) getFileType(path string, info os.FileInfo) string {
	if info.IsDir() {
		return "directory"
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".txt", ".md", ".rst":
		return "text"
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg":
		return "image"
	case ".mp4", ".avi", ".mkv", ".mov", ".wmv":
		return "video"
	case ".mp3", ".wav", ".flac", ".aac", ".ogg":
		return "audio"
	case ".pdf":
		return "pdf"
	case ".doc", ".docx", ".odt":
		return "document"
	case ".xls", ".xlsx", ".ods":
		return "spreadsheet"
	case ".ppt", ".pptx", ".odp":
		return "presentation"
	case ".zip", ".rar", ".7z", ".tar", ".gz":
		return "archive"
	case ".exe", ".msi", ".deb", ".rpm":
		return "executable"
	default:
		return "file"
	}
}

// getFileOwner 获取文件所有者
func (s *Service) getFileOwner(path string, info os.FileInfo) string {
	// 这里可以根据不同平台实现具体的所有者获取逻辑
	return "unknown"
}

// getMatchReason 获取匹配原因
func (s *Service) getMatchReason(path string, info os.FileInfo, options *SearchOptions, pattern *regexp.Regexp) string {
	name := info.Name()

	if options.UseRegex && pattern != nil {
		if pattern.MatchString(name) {
			return "regex match"
		}
	} else {
		matched, _ := filepath.Match(options.Pattern, name)
		if matched {
			return "wildcard match"
		}

		if options.CaseSensitive {
			if strings.Contains(name, options.Pattern) {
				return "substring match"
			}
		} else {
			if strings.Contains(strings.ToLower(name), strings.ToLower(options.Pattern)) {
				return "case-insensitive match"
			}
		}
	}

	return "other criteria"
}

// calculateScore 计算匹配分数
func (s *Service) calculateScore(path string, info os.FileInfo, options *SearchOptions) float64 {
	score := 0.0
	name := info.Name()

	// 完全匹配得分最高
	if strings.EqualFold(name, options.Pattern) {
		score += 100.0
	} else if strings.HasPrefix(strings.ToLower(name), strings.ToLower(options.Pattern)) {
		score += 80.0
	} else if strings.Contains(strings.ToLower(name), strings.ToLower(options.Pattern)) {
		score += 60.0
	}

	// 文件类型匹配加分
	if options.FileType != "" {
		fileType := s.getFileType(path, info)
		if strings.EqualFold(fileType, options.FileType) {
			score += 20.0
		}
	}

	// 路径深度影响分数（越浅分数越高）
	depth := strings.Count(path, string(filepath.Separator))
	score -= float64(depth) * 2.0

	return score
}

// displayResults 显示搜索结果
func (s *Service) displayResults(results []*SearchResult, options *SearchOptions) {
	if len(results) == 0 {
		fmt.Println("未找到匹配的文件")
		return
	}

	fmt.Printf("找到 %d 个匹配的文件:\n\n", len(results))

	for i, result := range results {
		fmt.Printf("%d. %s\n", i+1, result.Path)
		fmt.Printf("   类型: %s | 大小: %s | 修改时间: %s\n",
			result.Type,
			s.formatSize(result.Size),
			result.ModTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("   匹配原因: %s | 分数: %.1f\n", result.MatchReason, result.Score)
		fmt.Println()
	}
}

// formatSize 格式化文件大小
func (s *Service) formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
