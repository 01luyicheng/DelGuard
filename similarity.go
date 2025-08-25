package main

import (
	"math"
	"strings"
	"unicode/utf8"
)

// LevenshteinDistance 计算两个字符串之间的Levenshtein距离
func LevenshteinDistance(s1, s2 string) int {
	// 转换为rune切片以正确处理Unicode字符
	r1 := []rune(s1)
	r2 := []rune(s2)

	len1 := len(r1)
	len2 := len(r2)

	// 创建距离矩阵
	matrix := make([][]int, len1+1)
	for i := range matrix {
		matrix[i] = make([]int, len2+1)
	}

	// 初始化第一行和第一列
	for i := 0; i <= len1; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		matrix[0][j] = j
	}

	// 填充矩阵
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 0
			if r1[i-1] != r2[j-1] {
				cost = 1
			}

			matrix[i][j] = min3(
				matrix[i-1][j]+1,      // 删除
				matrix[i][j-1]+1,      // 插入
				matrix[i-1][j-1]+cost, // 替换
			)
		}
	}

	return matrix[len1][len2]
}

// min3 返回三个整数中的最小值
func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// CalculateSimilarity 计算两个字符串的相似度百分比
func CalculateSimilarity(s1, s2 string) float64 {
	// 预处理：转换为小写并去除空格
	s1 = strings.ToLower(strings.TrimSpace(s1))
	s2 = strings.ToLower(strings.TrimSpace(s2))

	if s1 == s2 {
		return 100.0
	}

	if s1 == "" || s2 == "" {
		return 0.0
	}

	// 计算Levenshtein距离
	distance := LevenshteinDistance(s1, s2)

	// 计算最大长度
	maxLen := math.Max(float64(utf8.RuneCountInString(s1)), float64(utf8.RuneCountInString(s2)))

	// 计算相似度百分比
	similarity := (1.0 - float64(distance)/maxLen) * 100.0

	// 确保结果在0-100范围内
	if similarity < 0 {
		similarity = 0
	}

	return similarity
}

// FindSimilarStrings 支持传入阈值（建议由 config.SimilarityThreshold 提供）
func FindSimilarStrings(target string, candidates []string, threshold float64) []SimilarMatch {
	var matches []SimilarMatch

	for _, candidate := range candidates {
		similarity := CalculateSimilarity(target, candidate)
		if similarity >= threshold {
			matches = append(matches, SimilarMatch{
				Text:       candidate,
				Similarity: similarity,
			})
		}
	}

	// 按相似度降序排序
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[i].Similarity < matches[j].Similarity {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	return matches
}

// SimilarMatch 表示一个相似匹配结果
type SimilarMatch struct {
	Text       string  // 匹配的文本
	Similarity float64 // 相似度百分比
}
