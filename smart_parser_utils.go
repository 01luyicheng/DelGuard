package main

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"unicode"
)

// cleanPath 清理和标准化路径
func (sp *SmartParser) cleanPath(path string) string {
	// 移除引号
	path = strings.Trim(path, `"'`)

	// 处理转义字符
	path = sp.unescapePath(path)

	// 展开用户目录
	if strings.HasPrefix(path, "~") {
		if homeDir, err := os.UserHomeDir(); err == nil {
			path = strings.Replace(path, "~", homeDir, 1)
		}
	}

	// 处理相对路径
	if !filepath.IsAbs(path) {
		path = filepath.Join(sp.workingDir, path)
	}

	// 清理路径
	path = filepath.Clean(path)

	return path
}

// unescapePath 处理路径中的转义字符
func (sp *SmartParser) unescapePath(path string) string {
	// 处理常见的转义序列
	replacements := map[string]string{
		`\ `: ` `,  // 转义空格
		`\t`: "\t", // 转义制表符
		`\n`: "\n", // 转义换行符
		`\\`: `\`,  // 转义反斜杠
		`\"`: `"`,  // 转义双引号
		`\'`: `'`,  // 转义单引号
	}

	for escaped, unescaped := range replacements {
		path = strings.ReplaceAll(path, escaped, unescaped)
	}

	return path
}

// hasWildcard 检查是否包含通配符
func (sp *SmartParser) hasWildcard(path string) bool {
	wildcards := []string{"*", "?", "[", "]"}
	for _, wildcard := range wildcards {
		if strings.Contains(path, wildcard) {
			return true
		}
	}
	return false
}

// findMatches 查找通配符匹配的文件
func (sp *SmartParser) findMatches(pattern string) []string {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return []string{}
	}

	// 限制返回结果数量
	if len(matches) > sp.maxSuggestions*2 {
		matches = matches[:sp.maxSuggestions*2]
	}

	return matches
}

// findSimilarPaths 查找相似的路径
func (sp *SmartParser) findSimilarPaths(targetPath string) []string {
	if !sp.enableFuzzy {
		return []string{}
	}

	var suggestions []string

	// 获取目标路径的目录和文件名
	dir := filepath.Dir(targetPath)
	filename := filepath.Base(targetPath)

	// 如果目录不存在，尝试查找相似目录
	if _, err := os.Stat(dir); err != nil {
		parentDir := filepath.Dir(dir)
		if entries, err := os.ReadDir(parentDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					similarity := sp.calculateSimilarity(filepath.Base(dir), entry.Name())
					if similarity > 0.6 {
						suggestedPath := filepath.Join(parentDir, entry.Name(), filename)
						suggestions = append(suggestions, suggestedPath)
					}
				}
			}
		}
	} else {
		// 目录存在，查找相似文件名
		if entries, err := os.ReadDir(dir); err == nil {
			for _, entry := range entries {
				similarity := sp.calculateSimilarity(filename, entry.Name())
				if similarity > 0.6 {
					suggestedPath := filepath.Join(dir, entry.Name())
					suggestions = append(suggestions, suggestedPath)
				}
			}
		}
	}

	// 限制建议数量
	if len(suggestions) > sp.maxSuggestions {
		suggestions = suggestions[:sp.maxSuggestions]
	}

	return suggestions
}

// calculateSimilarity 计算字符串相似度（简化的编辑距离算法）
func (sp *SmartParser) calculateSimilarity(s1, s2 string) float64 {
	if !sp.caseSensitive {
		s1 = strings.ToLower(s1)
		s2 = strings.ToLower(s2)
	}

	if s1 == s2 {
		return 1.0
	}

	// 使用简化的编辑距离算法
	maxLen := len(s1)
	if len(s2) > maxLen {
		maxLen = len(s2)
	}

	if maxLen == 0 {
		return 1.0
	}

	distance := sp.editDistance(s1, s2)
	return 1.0 - float64(distance)/float64(maxLen)
}

// editDistance 计算编辑距离
func (sp *SmartParser) editDistance(s1, s2 string) int {
	m, n := len(s1), len(s2)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 0; i <= m; i++ {
		dp[i][0] = i
	}
	for j := 0; j <= n; j++ {
		dp[0][j] = j
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if s1[i-1] == s2[j-1] {
				dp[i][j] = dp[i-1][j-1]
			} else {
				dp[i][j] = 1 + minInt(dp[i-1][j], dp[i][j-1], dp[i-1][j-1])
			}
		}
	}

	return dp[m][n]
}

// minInt 返回三个数中的最小值
func minInt(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}

// isFlag 检查是否为命令行标志
func (sp *SmartParser) isFlag(arg string) bool {
	// 短标志 (-v, -h)
	if len(arg) == 2 && arg[0] == '-' && unicode.IsLetter(rune(arg[1])) {
		return true
	}

	// 长标志 (--verbose, --help)
	if len(arg) > 2 && strings.HasPrefix(arg, "--") {
		return true
	}

	// Windows风格标志 (/v, /h)
	if runtime.GOOS == "windows" && len(arg) == 2 && arg[0] == '/' && unicode.IsLetter(rune(arg[1])) {
		return true
	}

	return false
}

// isOption 检查是否为选项参数
func (sp *SmartParser) isOption(arg string, index int, allArgs []string) bool {
	// 检查前一个参数是否为需要值的标志
	if index > 0 {
		prevArg := allArgs[index-1]
		if sp.isFlag(prevArg) {
			// 检查是否为需要值的标志
			needsValue := []string{"-o", "--output", "-c", "--config", "-l", "--log-level"}
			for _, flag := range needsValue {
				if prevArg == flag {
					return true
				}
			}
		}
	}

	return false
}

// isPureValue 检查是否为纯值参数
func (sp *SmartParser) isPureValue(arg string) bool {
	// 检查是否为数字
	if _, err := strconv.Atoi(arg); err == nil {
		return true
	}

	// 检查是否为浮点数
	if _, err := strconv.ParseFloat(arg, 64); err == nil {
		return true
	}

	// 检查是否为布尔值
	if arg == "true" || arg == "false" {
		return true
	}

	// 检查是否为URL
	if sp.isURL(arg) {
		return true
	}

	return false
}

// isURL 检查是否为URL
func (sp *SmartParser) isURL(arg string) bool {
	urlPattern := regexp.MustCompile(`^https?://`)
	return urlPattern.MatchString(arg)
}

// looksLikeDirectory 检查路径是否看起来像目录
func (sp *SmartParser) looksLikeDirectory(path string) bool {
	// 以斜杠结尾
	if strings.HasSuffix(path, "/") || strings.HasSuffix(path, "\\") {
		return true
	}

	// 没有文件扩展名
	if filepath.Ext(path) == "" {
		return true
	}

	// 包含常见目录名
	dirNames := []string{"bin", "lib", "src", "docs", "config", "tmp", "temp"}
	baseName := strings.ToLower(filepath.Base(path))
	for _, dirName := range dirNames {
		if baseName == dirName {
			return true
		}
	}

	return false
}
