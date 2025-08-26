package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// postProcessArguments 后处理参数
func (sp *SmartParser) postProcessArguments(args []ParsedArgument) []ParsedArgument {
	// 1. 处理相对路径关系
	sp.resolveRelativePaths(args)

	// 2. 验证路径安全性
	sp.validatePathSecurity(args)

	// 3. 优化建议
	sp.optimizeSuggestions(args)

	return args
}

// resolveRelativePaths 解析相对路径关系
func (sp *SmartParser) resolveRelativePaths(args []ParsedArgument) {
	for i := range args {
		if args[i].Type == ArgTypeFile || args[i].Type == ArgTypeDirectory {
			// 检查是否为相对于其他参数的路径
			sp.checkRelativeToOtherArgs(&args[i], args)
		}
	}
}

// checkRelativeToOtherArgs 检查是否相对于其他参数
func (sp *SmartParser) checkRelativeToOtherArgs(arg *ParsedArgument, allArgs []ParsedArgument) {
	if arg.Exists {
		return // 已存在的路径不需要检查
	}

	for _, otherArg := range allArgs {
		if otherArg.Type == ArgTypeDirectory && otherArg.Exists {
			// 尝试相对于这个目录解析
			relativePath := filepath.Join(otherArg.NormalizedPath, filepath.Base(arg.Raw))
			if _, err := os.Stat(relativePath); err == nil {
				arg.Suggestions = append(arg.Suggestions, relativePath)
				arg.Metadata["relative_to"] = otherArg.NormalizedPath
			}
		}
	}
}

// validatePathSecurity 验证路径安全性
func (sp *SmartParser) validatePathSecurity(args []ParsedArgument) {
	for i := range args {
		if args[i].Type == ArgTypeFile || args[i].Type == ArgTypeDirectory || args[i].Type == ArgTypePattern {
			// 检查路径遍历攻击
			if sp.hasPathTraversal(args[i].NormalizedPath) {
				args[i].Metadata["security_warning"] = "检测到潜在的路径遍历攻击"
				args[i].Confidence *= 0.5 // 降低置信度
			}

			// 检查是否为系统关键路径
			if IsCriticalPath(args[i].NormalizedPath) {
				args[i].Metadata["critical_path"] = true
				args[i].Metadata["warning"] = "这是系统关键路径，操作需要谨慎"
			}
		}
	}
}

// hasPathTraversal 检查是否包含路径遍历
func (sp *SmartParser) hasPathTraversal(path string) bool {
	// 检查常见的路径遍历模式
	traversalPatterns := []string{
		"..",
		"..\\",
		"../",
		"%2e%2e",
		"%252e%252e",
	}

	lowerPath := strings.ToLower(path)
	for _, pattern := range traversalPatterns {
		if strings.Contains(lowerPath, pattern) {
			return true
		}
	}

	return false
}

// optimizeSuggestions 优化建议
func (sp *SmartParser) optimizeSuggestions(args []ParsedArgument) {
	for i := range args {
		if len(args[i].Suggestions) > 1 {
			// 按相似度排序建议
			sp.sortSuggestionsBySimilarity(&args[i])

			// 限制建议数量
			if len(args[i].Suggestions) > sp.maxSuggestions {
				args[i].Suggestions = args[i].Suggestions[:sp.maxSuggestions]
			}
		}
	}
}

// sortSuggestionsBySimilarity 按相似度排序建议
func (sp *SmartParser) sortSuggestionsBySimilarity(arg *ParsedArgument) {
	if len(arg.Suggestions) <= 1 {
		return
	}

	// 简单的冒泡排序，按相似度降序
	for i := 0; i < len(arg.Suggestions)-1; i++ {
		for j := 0; j < len(arg.Suggestions)-1-i; j++ {
			sim1 := sp.calculateSimilarity(arg.Raw, filepath.Base(arg.Suggestions[j]))
			sim2 := sp.calculateSimilarity(arg.Raw, filepath.Base(arg.Suggestions[j+1]))
			if sim1 < sim2 {
				arg.Suggestions[j], arg.Suggestions[j+1] = arg.Suggestions[j+1], arg.Suggestions[j]
			}
		}
	}
}

// displayParseResult 显示解析结果
func (sp *SmartParser) displayParseResult(arg ParsedArgument) {
	typeStr := sp.getTypeString(arg.Type)
	confidenceStr := fmt.Sprintf("%.1f%%", arg.Confidence*100)

	fmt.Printf("  📄 '%s' → %s (置信度: %s)\n", arg.Raw, typeStr, confidenceStr)

	if arg.NormalizedPath != "" && arg.NormalizedPath != arg.Raw {
		fmt.Printf("     📍 标准化路径: %s\n", arg.NormalizedPath)
	}

	if arg.Exists {
		fmt.Printf("     ✅ 文件存在\n")
	} else if arg.Type == ArgTypeFile || arg.Type == ArgTypeDirectory {
		fmt.Printf("     ❌ 文件不存在\n")
	}

	if len(arg.Suggestions) > 0 {
		fmt.Printf("     💡 建议:\n")
		for _, suggestion := range arg.Suggestions {
			fmt.Printf("        • %s\n", suggestion)
		}
	}

	if warning, exists := arg.Metadata["security_warning"]; exists {
		fmt.Printf("     ⚠️  安全警告: %s\n", warning)
	}

	if arg.Metadata["critical_path"] == true {
		fmt.Printf("     🚨 关键路径警告\n")
	}
}

// getTypeString 获取类型字符串
func (sp *SmartParser) getTypeString(argType ArgumentType) string {
	switch argType {
	case ArgTypeFile:
		return "文件"
	case ArgTypeDirectory:
		return "目录"
	case ArgTypePattern:
		return "通配符模式"
	case ArgTypeFlag:
		return "命令标志"
	case ArgTypeOption:
		return "选项参数"
	case ArgTypeValue:
		return "值参数"
	default:
		return "未知"
	}
}

// GetValidPaths 获取所有有效的路径参数
func (sp *SmartParser) GetValidPaths(args []ParsedArgument) []string {
	var paths []string

	for _, arg := range args {
		if (arg.Type == ArgTypeFile || arg.Type == ArgTypeDirectory) && arg.Exists {
			paths = append(paths, arg.NormalizedPath)
		} else if arg.Type == ArgTypePattern {
			if matches, ok := arg.Metadata["matches"].([]string); ok {
				paths = append(paths, matches...)
			}
		}
	}

	return paths
}

// ShowParseReport 显示解析报告
func (sp *SmartParser) ShowParseReport(args []ParsedArgument) {
	fmt.Println("\n📊 参数解析报告")
	fmt.Println("=" + strings.Repeat("=", 50))

	var fileCount, dirCount, patternCount, flagCount, unknownCount int
	var existingCount, missingCount int

	for _, arg := range args {
		switch arg.Type {
		case ArgTypeFile:
			fileCount++
		case ArgTypeDirectory:
			dirCount++
		case ArgTypePattern:
			patternCount++
		case ArgTypeFlag, ArgTypeOption:
			flagCount++
		default:
			unknownCount++
		}

		if arg.Exists {
			existingCount++
		} else if arg.Type == ArgTypeFile || arg.Type == ArgTypeDirectory {
			missingCount++
		}
	}

	fmt.Printf("📁 目录: %d\n", dirCount)
	fmt.Printf("📄 文件: %d\n", fileCount)
	fmt.Printf("🔍 模式: %d\n", patternCount)
	fmt.Printf("🏷️  标志: %d\n", flagCount)
	fmt.Printf("❓ 未知: %d\n", unknownCount)
	fmt.Printf("✅ 存在: %d\n", existingCount)
	fmt.Printf("❌ 缺失: %d\n", missingCount)

	if missingCount > 0 {
		fmt.Printf("\n⚠️  发现 %d 个不存在的文件/目录\n", missingCount)
		fmt.Println("💡 建议检查路径拼写或使用智能建议")
	}

	fmt.Println()
}
