package search

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// IndexEntry 索引条目
type IndexEntry struct {
	Path     string
	Name     string
	Size     int64
	ModTime  time.Time
	Hash     string
	Keywords []string
}

// SearchIndex 搜索索引
type SearchIndex struct {
	entries map[string]*IndexEntry
	nameMap map[string][]*IndexEntry
	extMap  map[string][]*IndexEntry
	mu      sync.RWMutex
	lastUpdate time.Time
}

// NewSearchIndex 创建搜索索引
func NewSearchIndex() *SearchIndex {
	return &SearchIndex{
		entries: make(map[string]*IndexEntry),
		nameMap: make(map[string][]*IndexEntry),
		extMap:  make(map[string][]*IndexEntry),
	}
}

// BuildIndex 构建索引
func (si *SearchIndex) BuildIndex(rootPath string) error {
	si.mu.Lock()
	defer si.mu.Unlock()
	
	// 清空现有索引
	si.entries = make(map[string]*IndexEntry)
	si.nameMap = make(map[string][]*IndexEntry)
	si.extMap = make(map[string][]*IndexEntry)
	
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		
		entry := &IndexEntry{
			Path:    path,
			Name:    info.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			Keywords: si.extractKeywords(info.Name()),
		}
		
		// 添加到主索引
		si.entries[path] = entry
		
		// 添加到名称索引
		nameLower := strings.ToLower(info.Name())
		si.nameMap[nameLower] = append(si.nameMap[nameLower], entry)
		
		// 添加到扩展名索引
		ext := strings.ToLower(filepath.Ext(info.Name()))
		if ext != "" {
			si.extMap[ext] = append(si.extMap[ext], entry)
		}
		
		return nil
	})
	
	si.lastUpdate = time.Now()
	return err
}

// SearchByName 按名称搜索
func (si *SearchIndex) SearchByName(pattern string) []*IndexEntry {
	si.mu.RLock()
	defer si.mu.RUnlock()
	
	pattern = strings.ToLower(pattern)
	var results []*IndexEntry
	
	for name, entries := range si.nameMap {
		if strings.Contains(name, pattern) {
			results = append(results, entries...)
		}
	}
	
	return results
}

// SearchByExtension 按扩展名搜索
func (si *SearchIndex) SearchByExtension(ext string) []*IndexEntry {
	si.mu.RLock()
	defer si.mu.RUnlock()
	
	ext = strings.ToLower(ext)
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	
	return si.extMap[ext]
}

// extractKeywords 提取关键词
func (si *SearchIndex) extractKeywords(filename string) []string {
	// 移除扩展名
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	
	// 按常见分隔符分割
	separators := []string{"_", "-", " ", ".", "(", ")", "[", "]"}
	keywords := []string{name}
	
	for _, sep := range separators {
		var newKeywords []string
		for _, keyword := range keywords {
			parts := strings.Split(keyword, sep)
			for _, part := range parts {
				if len(part) > 0 {
					newKeywords = append(newKeywords, strings.ToLower(part))
				}
			}
		}
		keywords = newKeywords
	}
	
	// 去重
	keywordMap := make(map[string]bool)
	var uniqueKeywords []string
	for _, keyword := range keywords {
		if !keywordMap[keyword] && len(keyword) > 1 {
			keywordMap[keyword] = true
			uniqueKeywords = append(uniqueKeywords, keyword)
		}
	}
	
	return uniqueKeywords
}

// IsStale 检查索引是否过期
func (si *SearchIndex) IsStale(maxAge time.Duration) bool {
	si.mu.RLock()
	defer si.mu.RUnlock()
	
	return time.Since(si.lastUpdate) > maxAge
}

// GetStats 获取索引统计
func (si *SearchIndex) GetStats() map[string]int {
	si.mu.RLock()
	defer si.mu.RUnlock()
	
	return map[string]int{
		"total_files": len(si.entries),
		"name_entries": len(si.nameMap),
		"ext_entries": len(si.extMap),
	}
}