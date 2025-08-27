package search

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// 相似度阈值，用于确定文件名是否足够相似
const SimilarityThreshold = 0.6

// SmartSearch 在给定路径不存在时，查找相似的文件
func SmartSearch(filePath string) ([]string, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); err == nil {
		// 文件存在，不需要查找相似文件
		return []string{filePath}, nil
	}

	// 获取目录和文件名
	dir := filepath.Dir(filePath)
	baseName := filepath.Base(filePath)

	// 检查目录是否存在
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// 目录不存在，尝试在当前目录查找相似文件
		dir = "."
	}

	// 读取目录内容
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("无法读取目录 %s: %v", dir, err)
	}

	// 存储相似文件及其相似度
	var similarFiles []SimilarFile

	// 计算每个文件与目标文件的相似度
	for _, entry := range entries {
		if entry.IsDir() {
			continue // 跳过目录
		}

		entryName := entry.Name()
		similarity := calculateSimilarity(baseName, entryName)

		if similarity >= SimilarityThreshold {
			fullPath := filepath.Join(dir, entryName)
			similarFiles = append(similarFiles, SimilarFile{
				Path:       fullPath,
				Similarity: similarity,
			})
		}
	}

	// 如果没有找到相似文件，尝试在子目录中查找
	if len(similarFiles) == 0 {
		subDirSimilarFiles, err := searchInSubdirectories(dir, baseName)
		if err == nil {
			similarFiles = append(similarFiles, subDirSimilarFiles...)
		}
	}

	// 按相似度排序
	sort.Slice(similarFiles, func(i, j int) bool {
		return similarFiles[i].Similarity > similarFiles[j].Similarity
	})

	// 提取路径
	result := make([]string, len(similarFiles))
	for i, sf := range similarFiles {
		result[i] = sf.Path
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("未找到与 %s 相似的文件", filePath)
	}

	return result, nil
}
