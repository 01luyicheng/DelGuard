package search

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	config interface{}
}

// NewService 创建搜索服务 - 支持无参数调用
func NewService(config ...interface{}) *Service {
	var cfg interface{}
	if len(config) > 0 {
		cfg = config[0]
	}
	return &Service{
		config: cfg,
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

	var pattern string
	var searchPath string = "."

	// 解析参数
	for i := 0; i < len(args); i++ {
		if args[i] == "--pattern" && i+1 < len(args) {
			pattern = args[i+1]
			i++ // 跳过pattern值
		} else if !strings.HasPrefix(args[i], "-") {
			// 这是搜索路径
			searchPath = args[i]
		}
	}

	if pattern == "" {
		// 如果没有指定pattern，使用第一个非选项参数作为pattern
		for _, arg := range args {
			if !strings.HasPrefix(arg, "-") {
				pattern = arg
				break
			}
		}
	}

	if pattern == "" {
		return fmt.Errorf("请指定搜索模式")
	}

	return s.searchTarget(ctx, pattern, searchPath)
}

// searchTarget 搜索目标
func (s *Service) searchTarget(ctx context.Context, pattern, searchPath string) error {
	// 执行实际的文件搜索
	results, err := s.FindFiles(searchPath, pattern, true)
	if err != nil {
		return fmt.Errorf("搜索失败: %v", err)
	}

	if len(results) == 0 {
		fmt.Printf("未找到匹配 '%s' 的文件\n", pattern)
		return nil
	}

	fmt.Printf("找到 %d 个匹配文件:\n", len(results))
	for _, file := range results {
		fmt.Printf("  %s\n", file.Path)
	}

	return nil
}
