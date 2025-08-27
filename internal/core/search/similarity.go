package search

import (
	"math"
	"os"
	"path/filepath"
	"strings"
)

// SimilarFile 表示一个相似文件及其相似度
type SimilarFile struct {
	Path       string
	Similarity float64
}

// calculateSimilarity 计算两个字符串的相似度（0-1之间）
func calculateSimilarity(s1, s2 string) float64 {
	// 转换为小写，减少大小写差异的影响
	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)

	// 如果字符串完全相同，直接返回1.0
	if s1 == s2 {
		return 1.0
	}

	// 如果其中一个是另一个的子串，给予较高的相似度
	if strings.Contains(s1, s2) || strings.Contains(s2, s1) {
		// 计算长度比例
		ratio := float64(len(s1)) / float64(len(s2))
		if ratio > 1.0 {
			ratio = 1.0 / ratio
		}
		// 子串相似度基础分为0.8，再乘以长度比例
		return 0.8 * ratio
	}

	// 计算编辑距离
	distance := levenshteinDistance(s1, s2)
	maxLen := math.Max(float64(len(s1)), float64(len(s2)))

	// 将编辑距离转换为相似度
	if maxLen == 0 {
		return 1.0 // 两个空字符串视为完全相似
	}

	similarity := 1.0 - float64(distance)/maxLen

	// 考虑文件扩展名
	ext1 := filepath.Ext(s1)
	ext2 := filepath.Ext(s2)

	// 如果扩展名相同，增加相似度
	if ext1 != "" && ext2 != "" && ext1 == ext2 {
		similarity += 0.1
		if similarity > 1.0 {
			similarity = 1.0
		}
	}

	return similarity
}

// levenshteinDistance 计算两个字符串之间的编辑距离
func levenshteinDistance(s1, s2 string) int {
	// 创建距离矩阵
	d := make([][]int, len(s1)+1)
	for i := range d {
		d[i] = make([]int, len(s2)+1)
	}

	// 初始化第一行和第一列
	for i := range d {
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}

	// 填充距离矩阵
	for j := 1; j <= len(s2); j++ {
		for i := 1; i <= len(s1); i++ {
			if s1[i-1] == s2[j-1] {
				d[i][j] = d[i-1][j-1] // 字符相同，不需要操作
			} else {
				// 取三种操作的最小值：替换、插入、删除
				d[i][j] = min(
					d[i-1][j-1]+1, // 替换
					min(
						d[i][j-1]+1,  // 插入
						d[i-1][j]+1)) // 删除
			}
		}
	}

	return d[len(s1)][len(s2)]
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// searchInSubdirectories 在子目录中查找相似文件
func searchInSubdirectories(rootDir, targetName string) ([]SimilarFile, error) {
	var similarFiles []SimilarFile
	maxDepth := 2 // 限制搜索深度，避免过度搜索

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续搜索
		}

		// 计算当前路径的深度
		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return nil
		}
		depth := len(strings.Split(relPath, string(filepath.Separator)))

		// 如果超过最大深度，跳过
		if depth > maxDepth {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 只处理文件
		if !info.IsDir() {
			fileName := filepath.Base(path)
			similarity := calculateSimilarity(targetName, fileName)

			if similarity >= SimilarityThreshold {
				similarFiles = append(similarFiles, SimilarFile{
					Path:       path,
					Similarity: similarity,
				})
			}
		}

		return nil
	})

	return similarFiles, err
}
