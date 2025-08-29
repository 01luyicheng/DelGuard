package search

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// EnhancedIndex 增强的搜索索引
type EnhancedIndex struct {
	entries    map[string]*EnhancedIndexEntry
	nameIndex  map[string][]*EnhancedIndexEntry // 按文件名索引
	typeIndex  map[string][]*EnhancedIndexEntry // 按文件类型索引
	sizeIndex  []*EnhancedIndexEntry            // 按大小排序的索引
	timeIndex  []*EnhancedIndexEntry            // 按时间排序的索引
	mutex      sync.RWMutex
	indexPath  string
	lastUpdate time.Time
}

// EnhancedIndexEntry 增强的索引条目
type EnhancedIndexEntry struct {
	Path        string            `json:"path"`
	Name        string            `json:"name"`
	NameLower   string            `json:"nameLower"`
	Size        int64             `json:"size"`
	ModTime     time.Time         `json:"modTime"`
	IsDir       bool              `json:"isDir"`
	Type        string            `json:"type"`
	Extension   string            `json:"extension"`
	Checksum    string            `json:"checksum"`
	IndexedTime time.Time         `json:"indexedTime"`
	Keywords    []string          `json:"keywords"`
	Metadata    map[string]string `json:"metadata"`
	SearchScore float64           `json:"-"` // 运行时计算，不持久化
}

// NewEnhancedIndex 创建增强索引
func NewEnhancedIndex(indexPath string) *EnhancedIndex {
	if indexPath == "" {
		home, _ := os.UserHomeDir()
		indexPath = filepath.Join(home, ".delguard", "enhanced_index.json")
	}

	return &EnhancedIndex{
		entries:   make(map[string]*EnhancedIndexEntry),
		nameIndex: make(map[string][]*EnhancedIndexEntry),
		typeIndex: make(map[string][]*EnhancedIndexEntry),
		indexPath: indexPath,
	}
}

// BuildIndex 构建索引
func (idx *EnhancedIndex) BuildIndex(rootPath string) error {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	// 清空现有索引
	idx.entries = make(map[string]*EnhancedIndexEntry)
	idx.nameIndex = make(map[string][]*EnhancedIndexEntry)
	idx.typeIndex = make(map[string][]*EnhancedIndexEntry)
	idx.sizeIndex = nil
	idx.timeIndex = nil

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续遍历
		}

		entry, err := idx.createIndexEntry(path, info)
		if err != nil {
			return nil // 忽略错误，继续遍历
		}

		idx.addEntryToIndexes(entry)
		return nil
	})

	if err != nil {
		return err
	}

	// 构建排序索引
	idx.buildSortedIndexes()
	idx.lastUpdate = time.Now()

	return idx.Save()
}

// createIndexEntry 创建索引条目
func (idx *EnhancedIndex) createIndexEntry(path string, info os.FileInfo) (*EnhancedIndexEntry, error) {
	entry := &EnhancedIndexEntry{
		Path:        path,
		Name:        info.Name(),
		NameLower:   strings.ToLower(info.Name()),
		Size:        info.Size(),
		ModTime:     info.ModTime(),
		IsDir:       info.IsDir(),
		Extension:   strings.ToLower(filepath.Ext(path)),
		IndexedTime: time.Now(),
		Metadata:    make(map[string]string),
	}

	// 设置文件类型
	entry.Type = idx.getFileType(path, info)

	// 生成关键词
	entry.Keywords = idx.generateKeywords(path, info)

	// 计算校验和（仅对小文件）
	if !info.IsDir() && info.Size() < 1024*1024 { // 1MB以下的文件
		checksum, _ := idx.calculateChecksum(path)
		entry.Checksum = checksum
	}

	return entry, nil
}

// generateKeywords 生成搜索关键词
func (idx *EnhancedIndex) generateKeywords(path string, info os.FileInfo) []string {
	var keywords []string

	// 文件名关键词
	name := strings.ToLower(info.Name())
	nameWithoutExt := strings.ToLower(strings.TrimSuffix(info.Name(), filepath.Ext(info.Name())))

	// 分割文件名
	keywords = append(keywords, name, nameWithoutExt)

	// 按常见分隔符分割
	for _, sep := range []string{"_", "-", ".", " "} {
		parts := strings.Split(nameWithoutExt, sep)
		for _, part := range parts {
			if len(part) > 1 {
				keywords = append(keywords, part)
			}
		}
	}

	// 路径关键词
	pathParts := strings.Split(strings.ToLower(path), string(filepath.Separator))
	for _, part := range pathParts {
		if len(part) > 1 && part != "." && part != ".." {
			keywords = append(keywords, part)
		}
	}

	// 去重
	keywordMap := make(map[string]bool)
	var uniqueKeywords []string
	for _, keyword := range keywords {
		if !keywordMap[keyword] && len(keyword) > 0 {
			keywordMap[keyword] = true
			uniqueKeywords = append(uniqueKeywords, keyword)
		}
	}

	return uniqueKeywords
}

// calculateChecksum 计算文件校验和
func (idx *EnhancedIndex) calculateChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// getFileType 获取文件类型
func (idx *EnhancedIndex) getFileType(path string, info os.FileInfo) string {
	if info.IsDir() {
		return "directory"
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".txt", ".md", ".rst", ".log":
		return "text"
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg", ".webp":
		return "image"
	case ".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm":
		return "video"
	case ".mp3", ".wav", ".flac", ".aac", ".ogg", ".m4a":
		return "audio"
	case ".pdf":
		return "pdf"
	case ".doc", ".docx", ".odt", ".rtf":
		return "document"
	case ".xls", ".xlsx", ".ods", ".csv":
		return "spreadsheet"
	case ".ppt", ".pptx", ".odp":
		return "presentation"
	case ".zip", ".rar", ".7z", ".tar", ".gz", ".bz2":
		return "archive"
	case ".exe", ".msi", ".deb", ".rpm", ".dmg":
		return "executable"
	case ".go", ".py", ".js", ".java", ".cpp", ".c", ".h", ".cs", ".php":
		return "code"
	case ".json", ".xml", ".yaml", ".yml", ".toml", ".ini":
		return "config"
	default:
		return "file"
	}
}

// addEntryToIndexes 将条目添加到各种索引中
func (idx *EnhancedIndex) addEntryToIndexes(entry *EnhancedIndexEntry) {
	// 主索引
	idx.entries[entry.Path] = entry

	// 名称索引
	nameLower := entry.NameLower
	idx.nameIndex[nameLower] = append(idx.nameIndex[nameLower], entry)

	// 类型索引
	idx.typeIndex[entry.Type] = append(idx.typeIndex[entry.Type], entry)

	// 关键词索引
	for _, keyword := range entry.Keywords {
		idx.nameIndex[keyword] = append(idx.nameIndex[keyword], entry)
	}
}

// buildSortedIndexes 构建排序索引
func (idx *EnhancedIndex) buildSortedIndexes() {
	// 构建大小排序索引
	idx.sizeIndex = make([]*EnhancedIndexEntry, 0, len(idx.entries))
	for _, entry := range idx.entries {
		idx.sizeIndex = append(idx.sizeIndex, entry)
	}
	sort.Slice(idx.sizeIndex, func(i, j int) bool {
		return idx.sizeIndex[i].Size < idx.sizeIndex[j].Size
	})

	// 构建时间排序索引
	idx.timeIndex = make([]*EnhancedIndexEntry, 0, len(idx.entries))
	for _, entry := range idx.entries {
		idx.timeIndex = append(idx.timeIndex, entry)
	}
	sort.Slice(idx.timeIndex, func(i, j int) bool {
		return idx.timeIndex[i].ModTime.Before(idx.timeIndex[j].ModTime)
	})
}

// SmartSearch 智能搜索
func (idx *EnhancedIndex) SmartSearch(query string, options *SearchOptions) []*EnhancedIndexEntry {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	var results []*EnhancedIndexEntry
	queryLower := strings.ToLower(query)

	// 1. 精确匹配
	if entries, exists := idx.nameIndex[queryLower]; exists {
		results = append(results, entries...)
	}

	// 2. 前缀匹配
	for keyword, entries := range idx.nameIndex {
		if strings.HasPrefix(keyword, queryLower) && keyword != queryLower {
			results = append(results, entries...)
		}
	}

	// 3. 子字符串匹配
	for keyword, entries := range idx.nameIndex {
		if strings.Contains(keyword, queryLower) && !strings.HasPrefix(keyword, queryLower) {
			results = append(results, entries...)
		}
	}

	// 4. 模糊匹配
	for _, entry := range idx.entries {
		if idx.fuzzyMatch(entry, queryLower) {
			// 检查是否已经在结果中
			found := false
			for _, existing := range results {
				if existing.Path == entry.Path {
					found = true
					break
				}
			}
			if !found {
				results = append(results, entry)
			}
		}
	}

	// 计算搜索分数并排序
	idx.calculateSearchScores(results, queryLower)
	sort.Slice(results, func(i, j int) bool {
		return results[i].SearchScore > results[j].SearchScore
	})

	// 应用其他过滤条件
	results = idx.applyFilters(results, options)

	// 限制结果数量
	if options.MaxResults > 0 && len(results) > options.MaxResults {
		results = results[:options.MaxResults]
	}

	return results
}

// fuzzyMatch 模糊匹配
func (idx *EnhancedIndex) fuzzyMatch(entry *EnhancedIndexEntry, query string) bool {
	// 简单的模糊匹配算法
	text := entry.NameLower

	i, j := 0, 0
	for i < len(text) && j < len(query) {
		if text[i] == query[j] {
			j++
		}
		i++
	}

	return j == len(query)
}

// calculateSearchScores 计算搜索分数
func (idx *EnhancedIndex) calculateSearchScores(results []*EnhancedIndexEntry, query string) {
	for _, entry := range results {
		score := 0.0

		// 完全匹配得分最高
		if entry.NameLower == query {
			score += 100.0
		} else if strings.HasPrefix(entry.NameLower, query) {
			score += 80.0
		} else if strings.Contains(entry.NameLower, query) {
			score += 60.0
		}

		// 关键词匹配加分
		for _, keyword := range entry.Keywords {
			if keyword == query {
				score += 40.0
			} else if strings.Contains(keyword, query) {
				score += 20.0
			}
		}

		// 文件类型加分
		if strings.Contains(entry.Type, query) {
			score += 30.0
		}

		// 路径深度影响分数（越浅分数越高）
		depth := strings.Count(entry.Path, string(filepath.Separator))
		score -= float64(depth) * 2.0

		// 文件大小影响分数（适中大小得分更高）
		if entry.Size > 0 && entry.Size < 1024*1024*10 { // 10MB以下
			score += 10.0
		}

		// 最近修改的文件得分更高
		daysSinceModified := time.Since(entry.ModTime).Hours() / 24
		if daysSinceModified < 7 {
			score += 15.0 - daysSinceModified*2
		}

		entry.SearchScore = score
	}
}

// applyFilters 应用过滤条件
func (idx *EnhancedIndex) applyFilters(results []*EnhancedIndexEntry, options *SearchOptions) []*EnhancedIndexEntry {
	var filtered []*EnhancedIndexEntry

	for _, entry := range results {
		// 文件类型过滤
		if options.FileType != "" && !strings.EqualFold(entry.Type, options.FileType) {
			continue
		}

		// 大小过滤
		if options.MinSize > 0 && entry.Size < options.MinSize {
			continue
		}
		if options.MaxSize > 0 && entry.Size > options.MaxSize {
			continue
		}

		// 时间过滤
		if !options.ModifiedAfter.IsZero() && entry.ModTime.Before(options.ModifiedAfter) {
			continue
		}
		if !options.ModifiedBefore.IsZero() && entry.ModTime.After(options.ModifiedBefore) {
			continue
		}

		filtered = append(filtered, entry)
	}

	return filtered
}

// Load 加载索引
func (idx *EnhancedIndex) Load() error {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	if _, err := os.Stat(idx.indexPath); os.IsNotExist(err) {
		return nil // 索引文件不存在
	}

	data, err := os.ReadFile(idx.indexPath)
	if err != nil {
		return err
	}

	var indexData struct {
		Entries    map[string]*EnhancedIndexEntry `json:"entries"`
		LastUpdate time.Time                      `json:"lastUpdate"`
	}

	if err := json.Unmarshal(data, &indexData); err != nil {
		return err
	}

	idx.entries = indexData.Entries
	idx.lastUpdate = indexData.LastUpdate

	// 重建内存索引
	idx.nameIndex = make(map[string][]*EnhancedIndexEntry)
	idx.typeIndex = make(map[string][]*EnhancedIndexEntry)

	for _, entry := range idx.entries {
		idx.addEntryToIndexes(entry)
	}

	idx.buildSortedIndexes()
	return nil
}

// Save 保存索引
func (idx *EnhancedIndex) Save() error {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	// 确保目录存在
	dir := filepath.Dir(idx.indexPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	indexData := struct {
		Entries    map[string]*EnhancedIndexEntry `json:"entries"`
		LastUpdate time.Time                      `json:"lastUpdate"`
	}{
		Entries:    idx.entries,
		LastUpdate: idx.lastUpdate,
	}

	data, err := json.MarshalIndent(indexData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(idx.indexPath, data, 0644)
}

// IsBuilt 检查索引是否已构建
func (idx *EnhancedIndex) IsBuilt() bool {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()
	return len(idx.entries) > 0
}

// Clear 清空索引
func (idx *EnhancedIndex) Clear() {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	idx.entries = make(map[string]*EnhancedIndexEntry)
	idx.nameIndex = make(map[string][]*EnhancedIndexEntry)
	idx.typeIndex = make(map[string][]*EnhancedIndexEntry)
	idx.sizeIndex = nil
	idx.timeIndex = nil
	idx.lastUpdate = time.Time{}
}

// GetStats 获取索引统计信息
func (idx *EnhancedIndex) GetStats() map[string]interface{} {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	stats := make(map[string]interface{})
	stats["totalEntries"] = len(idx.entries)
	stats["lastUpdate"] = idx.lastUpdate

	// 按类型统计
	typeStats := make(map[string]int)
	for _, entry := range idx.entries {
		typeStats[entry.Type]++
	}
	stats["typeStats"] = typeStats

	return stats
}
