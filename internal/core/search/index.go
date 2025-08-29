package search

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/01luyicheng/DelGuard/internal/config"
)

// Index 搜索索引
type Index struct {
	config    *config.Config
	entries   map[string]*IndexEntry
	mutex     sync.RWMutex
	indexPath string
}

// IndexEntry 索引条目
type IndexEntry struct {
	Path        string    `json:"path"`
	Name        string    `json:"name"`
	Size        int64     `json:"size"`
	ModTime     time.Time `json:"modTime"`
	IsDir       bool      `json:"isDir"`
	Type        string    `json:"type"`
	Checksum    string    `json:"checksum"`
	IndexedTime time.Time `json:"indexedTime"`
}

// NewIndex 创建新索引
func NewIndex(cfg *config.Config) *Index {
	indexPath := cfg.Search.IndexPath
	if indexPath == "" {
		home, _ := os.UserHomeDir()
		indexPath = filepath.Join(home, ".delguard", "index.json")
	}

	return &Index{
		config:    cfg,
		entries:   make(map[string]*IndexEntry),
		indexPath: indexPath,
	}
}

// Load 加载索引
func (idx *Index) Load() error {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	if _, err := os.Stat(idx.indexPath); os.IsNotExist(err) {
		return nil // 索引文件不存在，返回空索引
	}

	data, err := os.ReadFile(idx.indexPath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &idx.entries)
}

// Save 保存索引
func (idx *Index) Save() error {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	// 确保目录存在
	dir := filepath.Dir(idx.indexPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(idx.entries, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(idx.indexPath, data, 0644)
}

// Update 更新索引
func (idx *Index) Update(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		// 文件不存在，从索引中删除
		idx.mutex.Lock()
		delete(idx.entries, path)
		idx.mutex.Unlock()
		return nil
	}

	entry := &IndexEntry{
		Path:        path,
		Name:        info.Name(),
		Size:        info.Size(),
		ModTime:     info.ModTime(),
		IsDir:       info.IsDir(),
		Type:        getFileType(path, info),
		IndexedTime: time.Now(),
	}

	idx.mutex.Lock()
	idx.entries[path] = entry
	idx.mutex.Unlock()

	return nil
}

// Search 在索引中搜索
func (idx *Index) Search(pattern string, options *SearchOptions) []*IndexEntry {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	var results []*IndexEntry

	for _, entry := range idx.entries {
		if idx.matchesEntry(entry, pattern, options) {
			results = append(results, entry)
		}
	}

	return results
}

// matchesEntry 检查索引条目是否匹配
func (idx *Index) matchesEntry(entry *IndexEntry, pattern string, options *SearchOptions) bool {
	// 这里实现索引匹配逻辑
	// 可以复用 Service 中的匹配逻辑
	return true
}

// getFileType 获取文件类型（辅助函数）
func getFileType(path string, info os.FileInfo) string {
	if info.IsDir() {
		return "directory"
	}

	ext := filepath.Ext(path)
	switch ext {
	case ".txt", ".md":
		return "text"
	case ".jpg", ".png", ".gif":
		return "image"
	case ".mp4", ".avi":
		return "video"
	case ".mp3", ".wav":
		return "audio"
	default:
		return "file"
	}
}
