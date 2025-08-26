package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// ArgumentType 参数类型枚举
type ArgumentType int

const (
	ArgTypeUnknown   ArgumentType = iota
	ArgTypeFile                   // 文件路径
	ArgTypeDirectory              // 目录路径
	ArgTypePattern                // 通配符模式
	ArgTypeFlag                   // 命令行标志
	ArgTypeOption                 // 选项参数
	ArgTypeValue                  // 纯值参数
)

// ParsedArgument 解析后的参数
type ParsedArgument struct {
	Raw            string                 `json:"raw"`                   // 原始输入
	Type           ArgumentType           `json:"type"`                  // 参数类型
	NormalizedPath string                 `json:"normalized_path"`       // 标准化路径
	Exists         bool                   `json:"exists"`                // 文件/目录是否存在
	IsAbsolute     bool                   `json:"is_absolute"`           // 是否为绝对路径
	HasWildcard    bool                   `json:"has_wildcard"`          // 是否包含通配符
	Confidence     float64                `json:"confidence"`            // 识别置信度 (0-1)
	Suggestions    []string               `json:"suggestions,omitempty"` // 修正建议
	Metadata       map[string]interface{} `json:"metadata,omitempty"`    // 额外元数据
}

// SmartParser 智能参数解析器
type SmartParser struct {
	workingDir     string
	caseSensitive  bool
	enableFuzzy    bool
	maxSuggestions int
}

// NewSmartParser 创建智能解析器
func NewSmartParser() *SmartParser {
	wd, _ := os.Getwd()
	return &SmartParser{
		workingDir:     wd,
		caseSensitive:  runtime.GOOS != "windows", // Windows默认不区分大小写
		enableFuzzy:    true,
		maxSuggestions: 5,
	}
}

// ParseArguments 解析命令行参数
func (sp *SmartParser) ParseArguments(args []string) ([]ParsedArgument, error) {
	fmt.Printf("🔍 智能解析 %d 个参数...\n", len(args))

	var results []ParsedArgument

	for i, arg := range args {
		parsed := sp.parseArgument(arg, i, args)
		results = append(results, parsed)

		// 显示解析结果
		sp.displayParseResult(parsed)
	}

	// 执行后处理优化
	results = sp.postProcessArguments(results)

	return results, nil
}

// parseArgument 解析单个参数
func (sp *SmartParser) parseArgument(arg string, index int, allArgs []string) ParsedArgument {
	result := ParsedArgument{
		Raw:        arg,
		Type:       ArgTypeUnknown,
		Confidence: 0.0,
		Metadata:   make(map[string]interface{}),
	}

	// 1. 检查是否为命令行标志
	if sp.isFlag(arg) {
		result.Type = ArgTypeFlag
		result.Confidence = 0.95
		return result
	}

	// 2. 检查是否为选项参数
	if sp.isOption(arg, index, allArgs) {
		result.Type = ArgTypeOption
		result.Confidence = 0.90
		return result
	}

	// 3. 尝试作为路径解析
	pathResult := sp.parseAsPath(arg)
	if pathResult.Confidence > 0.5 {
		return pathResult
	}

	// 4. 检查是否为纯值参数
	if sp.isPureValue(arg) {
		result.Type = ArgTypeValue
		result.Confidence = 0.3
		return result
	}

	// 5. 默认尝试作为文件路径处理
	return sp.parseAsPath(arg)
}

// parseAsPath 作为路径解析
func (sp *SmartParser) parseAsPath(arg string) ParsedArgument {
	result := ParsedArgument{
		Raw:        arg,
		Type:       ArgTypeFile,
		Confidence: 0.0,
		Metadata:   make(map[string]interface{}),
	}

	// 清理和标准化路径
	cleanPath := sp.cleanPath(arg)
	result.NormalizedPath = cleanPath

	// 检查是否为绝对路径
	result.IsAbsolute = filepath.IsAbs(cleanPath)

	// 检查是否包含通配符
	result.HasWildcard = sp.hasWildcard(cleanPath)

	// 如果包含通配符，设置为模式类型
	if result.HasWildcard {
		result.Type = ArgTypePattern
		result.Confidence = 0.8

		// 查找匹配的文件
		matches := sp.findMatches(cleanPath)
		result.Metadata["matches"] = matches
		result.Metadata["match_count"] = len(matches)

		if len(matches) > 0 {
			result.Confidence = 0.9
		}

		return result
	}

	// 检查文件/目录是否存在
	if info, err := os.Stat(cleanPath); err == nil {
		result.Exists = true
		if info.IsDir() {
			result.Type = ArgTypeDirectory
			result.Confidence = 0.95
		} else {
			result.Type = ArgTypeFile
			result.Confidence = 0.95
		}

		// 添加文件信息到元数据
		result.Metadata["size"] = info.Size()
		result.Metadata["mod_time"] = info.ModTime()
		result.Metadata["is_dir"] = info.IsDir()

		return result
	}

	// 文件不存在，尝试智能建议
	suggestions := sp.findSimilarPaths(cleanPath)
	result.Suggestions = suggestions

	// 根据路径特征判断类型
	if sp.looksLikeDirectory(cleanPath) {
		result.Type = ArgTypeDirectory
		result.Confidence = 0.6
	} else {
		result.Type = ArgTypeFile
		result.Confidence = 0.5
	}

	// 如果有建议，提高置信度
	if len(suggestions) > 0 {
		result.Confidence += 0.2
	}

	return result
}
