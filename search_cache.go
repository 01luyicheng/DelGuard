package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// SearchCache 搜索缓存结构体
type SearchCache struct {
	cache  map[string]*CacheEntry
	mutex  sync.RWMutex
	maxAge time.Duration
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Results    []SearchResult
	Timestamp  time.Time
	SearchDir  string
	Target     string
	SearchType string
}

// NewSearchCache 创建新的搜索缓存
func NewSearchCache(maxAge time.Duration) *SearchCache {
	return &SearchCache{
		cache:  make(map[string]*CacheEntry),
		maxAge: maxAge,
	}
}

// GenerateCacheKey 生成缓存键
func (sc *SearchCache) GenerateCacheKey(target, searchDir, searchType string) string {
	// 创建唯一的缓存键
	data := fmt.Sprintf("%s|%s|%s", target, searchDir, searchType)
	hash := sha1.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// Get 从缓存获取结果
func (sc *SearchCache) Get(target, searchDir, searchType string) ([]SearchResult, bool) {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()

	cacheKey := sc.GenerateCacheKey(target, searchDir, searchType)
	entry, exists := sc.cache[cacheKey]

	if !exists {
		return nil, false
	}

	// 检查缓存是否过期
	if time.Since(entry.Timestamp) > sc.maxAge {
		delete(sc.cache, cacheKey)
		return nil, false
	}

	// 检查搜索目录是否仍然存在
	if _, err := os.Stat(searchDir); err != nil {
		delete(sc.cache, cacheKey)
		return nil, false
	}

	return entry.Results, true
}

// Set 将结果存入缓存
func (sc *SearchCache) Set(target, searchDir, searchType string, results []SearchResult) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	cacheKey := sc.GenerateCacheKey(target, searchDir, searchType)

	// 创建缓存条目的副本以避免外部修改
	resultsCopy := make([]SearchResult, len(results))
	copy(resultsCopy, results)

	sc.cache[cacheKey] = &CacheEntry{
		Results:    resultsCopy,
		Timestamp:  time.Now(),
		SearchDir:  searchDir,
		Target:     target,
		SearchType: searchType,
	}
}

// Clear 清空缓存
func (sc *SearchCache) Clear() {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	sc.cache = make(map[string]*CacheEntry)
}

// Cleanup 清理过期缓存
func (sc *SearchCache) Cleanup() {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	now := time.Now()
	for key, entry := range sc.cache {
		if now.Sub(entry.Timestamp) > sc.maxAge {
			delete(sc.cache, key)
		}
	}
}

// GetStats 获取缓存统计信息
func (sc *SearchCache) GetStats() map[string]interface{} {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()

	stats := make(map[string]interface{})
	stats["total_entries"] = len(sc.cache)

	// 计算缓存总大小（估算）
	totalSize := 0
	for _, entry := range sc.cache {
		totalSize += len(entry.Results) * 200 // 估算每个结果约200字节
	}
	stats["estimated_size_bytes"] = totalSize

	// 计算平均缓存年龄
	if len(sc.cache) > 0 {
		totalAge := time.Duration(0)
		for _, entry := range sc.cache {
			totalAge += time.Since(entry.Timestamp)
		}
		stats["avg_age_minutes"] = totalAge.Minutes() / float64(len(sc.cache))
	} else {
		stats["avg_age_minutes"] = 0.0
	}

	return stats
}

// SearchHistory 搜索历史记录
type SearchHistory struct {
	entries []HistoryEntry
	maxSize int
	mutex   sync.RWMutex
}

// HistoryEntry 历史记录条目
type HistoryEntry struct {
	Target      string
	SearchDir   string
	Timestamp   time.Time
	ResultCount int
	SearchType  string
}

// NewSearchHistory 创建新的搜索历史
func NewSearchHistory(maxSize int) *SearchHistory {
	if maxSize <= 0 {
		maxSize = 100 // 默认最大100条记录
	}
	return &SearchHistory{
		entries: make([]HistoryEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

// Add 添加搜索历史记录
func (sh *SearchHistory) Add(target, searchDir, searchType string, resultCount int) {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	entry := HistoryEntry{
		Target:      target,
		SearchDir:   searchDir,
		Timestamp:   time.Now(),
		ResultCount: resultCount,
		SearchType:  searchType,
	}

	// 添加到开头
	sh.entries = append([]HistoryEntry{entry}, sh.entries...)

	// 限制历史记录大小
	if len(sh.entries) > sh.maxSize {
		sh.entries = sh.entries[:sh.maxSize]
	}
}

// GetRecent 获取最近的搜索历史
func (sh *SearchHistory) GetRecent(limit int) []HistoryEntry {
	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	if limit <= 0 || limit > len(sh.entries) {
		limit = len(sh.entries)
	}

	// 返回副本以避免外部修改
	result := make([]HistoryEntry, limit)
	copy(result, sh.entries[:limit])
	return result
}

// GetByTarget 获取特定目标的搜索历史
func (sh *SearchHistory) GetByTarget(target string) []HistoryEntry {
	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	var results []HistoryEntry
	for _, entry := range sh.entries {
		if entry.Target == target {
			results = append(results, entry)
		}
	}
	return results
}

// Clear 清空搜索历史
func (sh *SearchHistory) Clear() {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	sh.entries = sh.entries[:0]
}

// GetStats 获取历史记录统计信息
func (sh *SearchHistory) GetStats() map[string]interface{} {
	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	stats := make(map[string]interface{})
	stats["total_entries"] = len(sh.entries)
	stats["max_size"] = sh.maxSize

	// 统计最常用的搜索目标
	targetCount := make(map[string]int)
	for _, entry := range sh.entries {
		targetCount[entry.Target]++
	}

	// 找出前5个最常用的目标
	type targetFreq struct {
		Target string
		Count  int
	}

	var frequencies []targetFreq
	for target, count := range targetCount {
		frequencies = append(frequencies, targetFreq{target, count})
	}

	// 按使用频率排序
	for i := 0; i < len(frequencies)-1; i++ {
		for j := i + 1; j < len(frequencies); j++ {
			if frequencies[i].Count < frequencies[j].Count {
				frequencies[i], frequencies[j] = frequencies[j], frequencies[i]
			}
		}
	}

	// 只返回前5个
	if len(frequencies) > 5 {
		frequencies = frequencies[:5]
	}

	stats["top_targets"] = frequencies

	return stats
}

// EnhancedSmartSearch 增强版智能搜索
type EnhancedSmartSearch struct {
	*SmartFileSearch
	cache   *SearchCache
	history *SearchHistory
}

// NewEnhancedSmartSearch 创建增强版智能搜索
func NewEnhancedSmartSearch(config SmartSearchConfig) *EnhancedSmartSearch {
	return &EnhancedSmartSearch{
		SmartFileSearch: NewSmartFileSearch(config),
		cache:           NewSearchCache(30 * time.Minute), // 30分钟缓存
		history:         NewSearchHistory(100),            // 100条历史记录
	}
}

// SearchWithCache 使用缓存的搜索
func (ess *EnhancedSmartSearch) SearchWithCache(target, searchDir string) ([]SearchResult, error) {
	// 检查缓存
	if results, found := ess.cache.Get(target, searchDir, "similarity"); found {
		return results, nil
	}

	// 执行搜索
	results, err := ess.SmartFileSearch.SearchFiles(target, searchDir)
	if err != nil {
		return nil, err
	}

	// 存入缓存
	ess.cache.Set(target, searchDir, "similarity", results)

	// 添加到历史记录
	ess.history.Add(target, searchDir, "similarity", len(results))

	return results, nil
}

// SearchContentWithCache 使用缓存的内容搜索
func (ess *EnhancedSmartSearch) SearchContentWithCache(target, searchDir string) ([]SearchResult, error) {
	// 检查缓存
	if results, found := ess.cache.Get(target, searchDir, "content"); found {
		return results, nil
	}

	// 临时启用内容搜索
	oldConfig := ess.SmartFileSearch.config
	ess.SmartFileSearch.config.SearchContent = true

	// 执行搜索
	results, err := ess.SmartFileSearch.SearchFiles(target, searchDir)

	// 恢复配置
	ess.SmartFileSearch.config = oldConfig

	if err != nil {
		return nil, err
	}

	// 过滤出内容匹配的结果
	var contentResults []SearchResult
	for _, result := range results {
		if strings.Contains(result.MatchType, "content") {
			contentResults = append(contentResults, result)
		}
	}

	// 存入缓存
	ess.cache.Set(target, searchDir, "content", contentResults)

	// 添加到历史记录
	ess.history.Add(target, searchDir, "content", len(contentResults))

	return contentResults, nil
}

// SearchRegexWithCache 使用缓存的正则搜索
func (ess *EnhancedSmartSearch) SearchRegexWithCache(pattern, searchDir string) ([]SearchResult, error) {
	// 检查缓存
	if results, found := ess.cache.Get(pattern, searchDir, "regex"); found {
		return results, nil
	}

	// 执行正则搜索
	results, err := ess.SmartFileSearch.SearchByRegex(pattern, searchDir)
	if err != nil {
		return nil, err
	}

	// 存入缓存
	ess.cache.Set(pattern, searchDir, "regex", results)

	// 添加到历史记录
	ess.history.Add(pattern, searchDir, "regex", len(results))

	return results, nil
}

// GetCacheStats 获取缓存统计信息
func (ess *EnhancedSmartSearch) GetCacheStats() map[string]interface{} {
	return ess.cache.GetStats()
}

// GetHistoryStats 获取历史记录统计信息
func (ess *EnhancedSmartSearch) GetHistoryStats() map[string]interface{} {
	return ess.history.GetStats()
}

// ClearCache 清空缓存
func (ess *EnhancedSmartSearch) ClearCache() {
	ess.cache.Clear()
}

// ClearHistory 清空历史记录
func (ess *EnhancedSmartSearch) ClearHistory() {
	ess.history.Clear()
}

// GetRecentSearches 获取最近的搜索
func (ess *EnhancedSmartSearch) GetRecentSearches(limit int) []HistoryEntry {
	return ess.history.GetRecent(limit)
}

// GetSearchSuggestions 获取搜索建议
func (ess *EnhancedSmartSearch) GetSearchSuggestions(prefix string) []string {
	recent := ess.history.GetRecent(20)
	var suggestions []string
	seen := make(map[string]bool)

	for _, entry := range recent {
		if strings.HasPrefix(strings.ToLower(entry.Target), strings.ToLower(prefix)) && !seen[entry.Target] {
			suggestions = append(suggestions, entry.Target)
			seen[entry.Target] = true

			if len(suggestions) >= 5 { // 最多返回5个建议
				break
			}
		}
	}

	return suggestions
}
