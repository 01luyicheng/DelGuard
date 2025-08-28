package search

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FileInfo 文件信息
type FileInfo struct {
	Path    string
	Size    int64
	ModTime int64
	IsDir   bool
	Hash    string
}

// DuplicateGroup 重复文件组
type DuplicateGroup struct {
	Hash  string
	Files []FileInfo
}

// Service 搜索服务
type Service struct {
	config       interface{}
	enhancedEngine *EnhancedSearchEngine
	searchIndex    *SearchIndex
}

// NewService 创建搜索服务 - 支持无参数调用
func NewService(config ...interface{}) *Service {
	var cfg interface{}
	if len(config) > 0 {
		cfg = config[0]
	}
	return &Service{
		config:         cfg,
		enhancedEngine: NewEnhancedSearchEngine(),
		searchIndex:    NewSearchIndex(),
	}
}

// FindFiles 查找文件
func (s *Service) FindFiles(rootPath, pattern string, recursive bool) ([]FileInfo, error) {
	var results []FileInfo

	if recursive {
		err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // 忽略错误，继续搜索
			}

			if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
				fileInfo := FileInfo{
					Path:    path,
					Size:    info.Size(),
					ModTime: info.ModTime().Unix(),
					IsDir:   info.IsDir(),
				}
				results = append(results, fileInfo)
			}

			return nil
		})
		return results, err
	} else {
		// 非递归搜索，只搜索当前目录
		entries, err := os.ReadDir(rootPath)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			if matched, _ := filepath.Match(pattern, entry.Name()); matched {
				fullPath := filepath.Join(rootPath, entry.Name())
				info, err := entry.Info()
				if err != nil {
					continue
				}

				fileInfo := FileInfo{
					Path:    fullPath,
					Size:    info.Size(),
					ModTime: info.ModTime().Unix(),
					IsDir:   info.IsDir(),
				}
				results = append(results, fileInfo)
			}
		}

		return results, nil
	}
}

// FindBySize 按大小查找文件
func (s *Service) FindBySize(rootPath string, minSize int64, recursive bool) ([]FileInfo, error) {
	var results []FileInfo

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续搜索
		}

		if !info.IsDir() && info.Size() >= minSize {
			fileInfo := FileInfo{
				Path:    path,
				Size:    info.Size(),
				ModTime: info.ModTime().Unix(),
				IsDir:   false,
			}
			results = append(results, fileInfo)
		}

		return nil
	}

	if recursive {
		err := filepath.Walk(rootPath, walkFunc)
		return results, err
	} else {
		entries, err := os.ReadDir(rootPath)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			fullPath := filepath.Join(rootPath, entry.Name())
			info, err := entry.Info()
			if err != nil {
				continue
			}

			if !info.IsDir() && info.Size() >= minSize {
				fileInfo := FileInfo{
					Path:    fullPath,
					Size:    info.Size(),
					ModTime: info.ModTime().Unix(),
					IsDir:   false,
				}
				results = append(results, fileInfo)
			}
		}

		return results, nil
	}
}

// FindDuplicates 查找重复文件
func (s *Service) FindDuplicates(rootPath string) ([]DuplicateGroup, error) {
	fileHashes := make(map[string][]FileInfo)

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续搜索
		}

		if !info.IsDir() {
			hash, err := s.calculateFileHash(path)
			if err != nil {
				return nil // 忽略无法计算哈希的文件
			}

			fileInfo := FileInfo{
				Path:    path,
				Size:    info.Size(),
				ModTime: info.ModTime().Unix(),
				IsDir:   false,
				Hash:    hash,
			}

			fileHashes[hash] = append(fileHashes[hash], fileInfo)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	var duplicates []DuplicateGroup
	for hash, files := range fileHashes {
		if len(files) > 1 {
			duplicates = append(duplicates, DuplicateGroup{
				Hash:  hash,
				Files: files,
			})
		}
	}

	return duplicates, nil
}

// calculateFileHash 计算文件哈希值
func (s *Service) calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// Execute 执行搜索操作
func (s *Service) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("请指定搜索目标")
	}

	for _, target := range args {
		if err := s.searchTarget(ctx, target); err != nil {
			return fmt.Errorf("搜索 %s 失败: %v", target, err)
		}
	}

	return nil
}

// searchTarget 搜索目标
func (s *Service) searchTarget(ctx context.Context, target string) error {
	// 这里会实现具体的搜索逻辑
	fmt.Printf("搜索: %s\n", target)
	return nil
}

// EnhancedSearch 增强搜索
func (s *Service) EnhancedSearch(ctx context.Context, rootPath string, filter SearchFilter) ([]EnhancedResult, error) {
	return s.enhancedEngine.MultiSearch(ctx, rootPath, filter)
}

// QuickSearch 快速搜索
func (s *Service) QuickSearch(rootPath, query string) ([]string, error) {
	return s.enhancedEngine.QuickSearch(rootPath, query)
}

// BuildSearchIndex 构建搜索索引
func (s *Service) BuildSearchIndex(rootPath string) error {
	return s.searchIndex.BuildIndex(rootPath)
}

// SearchByName 按名称搜索（使用索引）
func (s *Service) SearchByName(pattern string) []*IndexEntry {
	return s.searchIndex.SearchByName(pattern)
}

// SearchByExtension 按扩展名搜索（使用索引）
func (s *Service) SearchByExtension(ext string) []*IndexEntry {
	return s.searchIndex.SearchByExtension(ext)
}

// GetIndexStats 获取索引统计
func (s *Service) GetIndexStats() map[string]int {
	return s.searchIndex.GetStats()
}
