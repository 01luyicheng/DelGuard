package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	mu                sync.RWMutex
	StartTime         time.Time
	EndTime           time.Time
	Duration          time.Duration
	MemoryUsage       runtime.MemStats
	GoroutineCount    int
	FileOperations    int64
	BytesProcessed    int64
	ErrorCount        int64
	SuccessCount      int64
	CacheHits         int64
	CacheMisses       int64
	DiskIOOperations  int64
	NetworkOperations int64
}

// PerformanceAnalyzer 性能分析器
type PerformanceAnalyzer struct {
	mu              sync.RWMutex
	metrics         *PerformanceMetrics
	operationTimes  map[string][]time.Duration
	bottlenecks     []string
	recommendations []string
	outputManager   *OutputManager
	enabled         bool
}

// NewPerformanceAnalyzer 创建性能分析器
func NewPerformanceAnalyzer(outputManager *OutputManager) *PerformanceAnalyzer {
	return &PerformanceAnalyzer{
		metrics:         &PerformanceMetrics{},
		operationTimes:  make(map[string][]time.Duration),
		bottlenecks:     make([]string, 0),
		recommendations: make([]string, 0),
		outputManager:   outputManager,
		enabled:         true,
	}
}

// StartProfiling 开始性能分析
func (pa *PerformanceAnalyzer) StartProfiling() {
	if !pa.enabled {
		return
	}

	pa.mu.Lock()
	defer pa.mu.Unlock()

	pa.metrics.StartTime = time.Now()
	runtime.ReadMemStats(&pa.metrics.MemoryUsage)
	pa.metrics.GoroutineCount = runtime.NumGoroutine()

	pa.outputManager.Debug("性能分析已启动")
}

// StopProfiling 停止性能分析
func (pa *PerformanceAnalyzer) StopProfiling() *PerformanceMetrics {
	if !pa.enabled {
		return nil
	}

	pa.mu.Lock()
	defer pa.mu.Unlock()

	pa.metrics.EndTime = time.Now()
	pa.metrics.Duration = pa.metrics.EndTime.Sub(pa.metrics.StartTime)

	var finalMemStats runtime.MemStats
	runtime.ReadMemStats(&finalMemStats)
	pa.metrics.MemoryUsage = finalMemStats
	pa.metrics.GoroutineCount = runtime.NumGoroutine()

	pa.outputManager.Debug("性能分析已完成，耗时: %v", pa.metrics.Duration)
	return pa.metrics
}

// RecordOperation 记录操作性能
func (pa *PerformanceAnalyzer) RecordOperation(operation string, duration time.Duration) {
	if !pa.enabled {
		return
	}

	pa.mu.Lock()
	defer pa.mu.Unlock()

	if pa.operationTimes[operation] == nil {
		pa.operationTimes[operation] = make([]time.Duration, 0)
	}
	pa.operationTimes[operation] = append(pa.operationTimes[operation], duration)

	// 检查是否为性能瓶颈
	if duration > 5*time.Second {
		pa.bottlenecks = append(pa.bottlenecks, fmt.Sprintf("%s: %v", operation, duration))
	}
}

// IncrementFileOperations 增加文件操作计数
func (pa *PerformanceAnalyzer) IncrementFileOperations() {
	if !pa.enabled {
		return
	}
	pa.mu.Lock()
	defer pa.mu.Unlock()
	pa.metrics.FileOperations++
}

// AddBytesProcessed 添加处理的字节数
func (pa *PerformanceAnalyzer) AddBytesProcessed(bytes int64) {
	if !pa.enabled {
		return
	}
	pa.mu.Lock()
	defer pa.mu.Unlock()
	pa.metrics.BytesProcessed += bytes
}

// IncrementErrors 增加错误计数
func (pa *PerformanceAnalyzer) IncrementErrors() {
	if !pa.enabled {
		return
	}
	pa.mu.Lock()
	defer pa.mu.Unlock()
	pa.metrics.ErrorCount++
}

// IncrementSuccess 增加成功计数
func (pa *PerformanceAnalyzer) IncrementSuccess() {
	if !pa.enabled {
		return
	}
	pa.mu.Lock()
	defer pa.mu.Unlock()
	pa.metrics.SuccessCount++
}

// RecordCacheHit 记录缓存命中
func (pa *PerformanceAnalyzer) RecordCacheHit() {
	if !pa.enabled {
		return
	}
	pa.mu.Lock()
	defer pa.mu.Unlock()
	pa.metrics.CacheHits++
}

// RecordCacheMiss 记录缓存未命中
func (pa *PerformanceAnalyzer) RecordCacheMiss() {
	if !pa.enabled {
		return
	}
	pa.mu.Lock()
	defer pa.mu.Unlock()
	pa.metrics.CacheMisses++
}

// AnalyzePerformance 分析性能
func (pa *PerformanceAnalyzer) AnalyzePerformance() map[string]interface{} {
	pa.mu.RLock()
	defer pa.mu.RUnlock()

	analysis := make(map[string]interface{})

	// 基本统计
	analysis["duration"] = pa.metrics.Duration
	analysis["file_operations"] = pa.metrics.FileOperations
	analysis["bytes_processed"] = pa.metrics.BytesProcessed
	analysis["error_rate"] = float64(pa.metrics.ErrorCount) / float64(pa.metrics.SuccessCount+pa.metrics.ErrorCount)
	analysis["success_rate"] = float64(pa.metrics.SuccessCount) / float64(pa.metrics.SuccessCount+pa.metrics.ErrorCount)

	// 缓存效率
	totalCacheOps := pa.metrics.CacheHits + pa.metrics.CacheMisses
	if totalCacheOps > 0 {
		analysis["cache_hit_rate"] = float64(pa.metrics.CacheHits) / float64(totalCacheOps)
	}

	// 内存使用
	analysis["memory_alloc"] = pa.metrics.MemoryUsage.Alloc
	analysis["memory_sys"] = pa.metrics.MemoryUsage.Sys
	analysis["gc_cycles"] = pa.metrics.MemoryUsage.NumGC

	// 吞吐量
	if pa.metrics.Duration > 0 {
		analysis["throughput_ops_per_sec"] = float64(pa.metrics.FileOperations) / pa.metrics.Duration.Seconds()
		analysis["throughput_bytes_per_sec"] = float64(pa.metrics.BytesProcessed) / pa.metrics.Duration.Seconds()
	}

	// 操作统计
	operationStats := make(map[string]map[string]interface{})
	for op, times := range pa.operationTimes {
		stats := pa.calculateOperationStats(times)
		operationStats[op] = stats
	}
	analysis["operation_stats"] = operationStats

	// 瓶颈和建议
	analysis["bottlenecks"] = pa.bottlenecks
	analysis["recommendations"] = pa.generateRecommendations()

	return analysis
}

// calculateOperationStats 计算操作统计
func (pa *PerformanceAnalyzer) calculateOperationStats(times []time.Duration) map[string]interface{} {
	if len(times) == 0 {
		return nil
	}

	stats := make(map[string]interface{})

	// 计算平均值
	var total time.Duration
	for _, t := range times {
		total += t
	}
	avg := total / time.Duration(len(times))
	stats["average"] = avg
	stats["count"] = len(times)
	stats["total"] = total

	// 计算最小值和最大值
	min := times[0]
	max := times[0]
	for _, t := range times {
		if t < min {
			min = t
		}
		if t > max {
			max = t
		}
	}
	stats["min"] = min
	stats["max"] = max

	return stats
}

// generateRecommendations 生成优化建议
func (pa *PerformanceAnalyzer) generateRecommendations() []string {
	recommendations := make([]string, 0)

	// 内存使用建议
	if pa.metrics.MemoryUsage.Alloc > 512*1024*1024 { // 512MB
		recommendations = append(recommendations, "内存使用过高，建议优化内存分配")
	}

	// GC建议
	if pa.metrics.MemoryUsage.NumGC > 100 {
		recommendations = append(recommendations, "GC次数过多，建议减少内存分配")
	}

	// 错误率建议
	totalOps := pa.metrics.SuccessCount + pa.metrics.ErrorCount
	if totalOps > 0 {
		errorRate := float64(pa.metrics.ErrorCount) / float64(totalOps)
		if errorRate > 0.1 { // 10%
			recommendations = append(recommendations, "错误率过高，建议检查错误处理逻辑")
		}
	}

	// 缓存建议
	totalCacheOps := pa.metrics.CacheHits + pa.metrics.CacheMisses
	if totalCacheOps > 0 {
		hitRate := float64(pa.metrics.CacheHits) / float64(totalCacheOps)
		if hitRate < 0.8 { // 80%
			recommendations = append(recommendations, "缓存命中率较低，建议优化缓存策略")
		}
	}

	// 操作时间建议
	for op, times := range pa.operationTimes {
		if len(times) > 0 {
			var total time.Duration
			for _, t := range times {
				total += t
			}
			avg := total / time.Duration(len(times))
			if avg > 1*time.Second {
				recommendations = append(recommendations, fmt.Sprintf("操作 %s 平均耗时过长: %v", op, avg))
			}
		}
	}

	return recommendations
}

// GenerateReport 生成性能报告
func (pa *PerformanceAnalyzer) GenerateReport() string {
	analysis := pa.AnalyzePerformance()

	report := fmt.Sprintf("=== DelGuard 性能分析报告 ===\n")
	report += fmt.Sprintf("分析时间: %v\n", time.Now().Format("2006-01-02 15:04:05"))
	report += fmt.Sprintf("运行时长: %v\n", analysis["duration"])
	report += fmt.Sprintf("文件操作数: %d\n", analysis["file_operations"])
	report += fmt.Sprintf("处理字节数: %d\n", analysis["bytes_processed"])

	if errorRate, ok := analysis["error_rate"].(float64); ok {
		report += fmt.Sprintf("错误率: %.2f%%\n", errorRate*100)
	}

	if successRate, ok := analysis["success_rate"].(float64); ok {
		report += fmt.Sprintf("成功率: %.2f%%\n", successRate*100)
	}

	if cacheHitRate, ok := analysis["cache_hit_rate"].(float64); ok {
		report += fmt.Sprintf("缓存命中率: %.2f%%\n", cacheHitRate*100)
	}

	if throughputOps, ok := analysis["throughput_ops_per_sec"].(float64); ok {
		report += fmt.Sprintf("操作吞吐量: %.2f ops/sec\n", throughputOps)
	}

	if throughputBytes, ok := analysis["throughput_bytes_per_sec"].(float64); ok {
		report += fmt.Sprintf("数据吞吐量: %.2f bytes/sec\n", throughputBytes)
	}

	report += fmt.Sprintf("内存分配: %d bytes\n", analysis["memory_alloc"])
	report += fmt.Sprintf("系统内存: %d bytes\n", analysis["memory_sys"])
	report += fmt.Sprintf("GC次数: %d\n", analysis["gc_cycles"])

	// 瓶颈
	if bottlenecks, ok := analysis["bottlenecks"].([]string); ok && len(bottlenecks) > 0 {
		report += "\n=== 性能瓶颈 ===\n"
		for _, bottleneck := range bottlenecks {
			report += fmt.Sprintf("- %s\n", bottleneck)
		}
	}

	// 建议
	if recommendations, ok := analysis["recommendations"].([]string); ok && len(recommendations) > 0 {
		report += "\n=== 优化建议 ===\n"
		for _, rec := range recommendations {
			report += fmt.Sprintf("- %s\n", rec)
		}
	}

	return report
}

// MeasureOperation 测量操作性能
func (pa *PerformanceAnalyzer) MeasureOperation(ctx context.Context, operation string, fn func() error) error {
	if !pa.enabled {
		return fn()
	}

	start := time.Now()
	defer func() {
		duration := time.Since(start)
		pa.RecordOperation(operation, duration)
	}()

	err := fn()
	if err != nil {
		pa.IncrementErrors()
	} else {
		pa.IncrementSuccess()
	}

	return err
}

// Enable 启用性能分析
func (pa *PerformanceAnalyzer) Enable() {
	pa.mu.Lock()
	defer pa.mu.Unlock()
	pa.enabled = true
}

// Disable 禁用性能分析
func (pa *PerformanceAnalyzer) Disable() {
	pa.mu.Lock()
	defer pa.mu.Unlock()
	pa.enabled = false
}

// Reset 重置性能数据
func (pa *PerformanceAnalyzer) Reset() {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	pa.metrics = &PerformanceMetrics{}
	pa.operationTimes = make(map[string][]time.Duration)
	pa.bottlenecks = make([]string, 0)
	pa.recommendations = make([]string, 0)
}

// 全局性能分析器
var globalPerformanceAnalyzer = NewPerformanceAnalyzer(globalOutputManager)

// 全局函数
func StartProfiling() {
	globalPerformanceAnalyzer.StartProfiling()
}

func StopProfiling() *PerformanceMetrics {
	return globalPerformanceAnalyzer.StopProfiling()
}

func MeasureOperation(ctx context.Context, operation string, fn func() error) error {
	return globalPerformanceAnalyzer.MeasureOperation(ctx, operation, fn)
}

func GeneratePerformanceReport() string {
	return globalPerformanceAnalyzer.GenerateReport()
}
