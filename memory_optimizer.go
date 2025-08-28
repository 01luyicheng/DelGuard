package main

import (
	"context"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

// MemoryOptimizer 内存优化器
type MemoryOptimizer struct {
	mu               sync.RWMutex
	gcPercent        int
	memoryLimit      int64
	cleanupInterval  time.Duration
	forceGCThreshold int64
	outputManager    *OutputManager
	enabled          bool
	stopChan         chan struct{}
	cleanupTicker    *time.Ticker
}

// NewMemoryOptimizer 创建内存优化器
func NewMemoryOptimizer(outputManager *OutputManager) *MemoryOptimizer {
	return &MemoryOptimizer{
		gcPercent:        100,                // 默认GC百分比
		memoryLimit:      1024 * 1024 * 1024, // 1GB内存限制
		cleanupInterval:  30 * time.Second,   // 30秒清理间隔
		forceGCThreshold: 512 * 1024 * 1024,  // 512MB强制GC阈值
		outputManager:    outputManager,
		enabled:          true,
		stopChan:         make(chan struct{}),
	}
}

// Start 启动内存优化器
func (mo *MemoryOptimizer) Start(ctx context.Context) {
	if !mo.enabled {
		return
	}

	mo.mu.Lock()
	defer mo.mu.Unlock()

	// 设置GC百分比
	debug.SetGCPercent(mo.gcPercent)

	// 启动清理协程
	mo.cleanupTicker = time.NewTicker(mo.cleanupInterval)
	go mo.cleanupRoutine(ctx)

	mo.outputManager.Debug("内存优化器已启动")
}

// Stop 停止内存优化器
func (mo *MemoryOptimizer) Stop() {
	mo.mu.Lock()
	defer mo.mu.Unlock()

	if mo.cleanupTicker != nil {
		mo.cleanupTicker.Stop()
	}

	close(mo.stopChan)
	mo.outputManager.Debug("内存优化器已停止")
}

// cleanupRoutine 清理协程
func (mo *MemoryOptimizer) cleanupRoutine(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-mo.stopChan:
			return
		case <-mo.cleanupTicker.C:
			mo.performCleanup()
		}
	}
}

// performCleanup 执行清理
func (mo *MemoryOptimizer) performCleanup() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 检查内存使用情况
	if int64(m.Alloc) > mo.forceGCThreshold {
		// 只在必要时输出简化的调试信息
		if mo.outputManager != nil && mo.outputManager.IsDebugEnabled() {
			mo.outputManager.Debug("执行内存清理: %d MB", m.Alloc/1024/1024)
		}
		runtime.GC()

		// 再次检查
		runtime.ReadMemStats(&m)
		// 简化调试信息输出
		if mo.outputManager != nil && mo.outputManager.IsDebugEnabled() {
			mo.outputManager.Debug("内存清理完成: %d MB", m.Alloc/1024/1024)
		}
	}

	// 检查是否超过内存限制
	if int64(m.Sys) > mo.memoryLimit {
		// 只在警告级别输出关键信息
		if mo.outputManager != nil {
			mo.outputManager.Warn("内存使用超限: %d MB", m.Sys/1024/1024)
		}

		// 执行更激进的清理
		mo.aggressiveCleanup()
	}
}

// aggressiveCleanup 激进清理
func (mo *MemoryOptimizer) aggressiveCleanup() {
	// 只在调试模式下输出详细信息
	if mo.outputManager != nil && mo.outputManager.IsDebugEnabled() {
		mo.outputManager.Debug("执行内存清理")
	}

	// 强制GC
	runtime.GC()
	runtime.GC() // 执行两次确保彻底清理

	// 返回内存给操作系统
	debug.FreeOSMemory()

	// 清理全局缓存（如果存在）
	// 这里可以添加具体的缓存清理逻辑

	// 简化完成信息
	if mo.outputManager != nil && mo.outputManager.IsDebugEnabled() {
		mo.outputManager.Debug("内存清理完成")
	}
}

// OptimizeForOperation 为特定操作优化内存
func (mo *MemoryOptimizer) OptimizeForOperation(operation string) func() {
	if !mo.enabled {
		return func() {}
	}

	var originalGCPercent int

	switch operation {
	case "large_file_operation":
		// 大文件操作：降低GC频率
		originalGCPercent = debug.SetGCPercent(200)
		// 简化调试信息
		if mo.outputManager != nil && mo.outputManager.IsDebugEnabled() {
			mo.outputManager.Debug("优化大文件操作内存")
		}

	case "many_small_files":
		// 多小文件操作：提高GC频率
		originalGCPercent = debug.SetGCPercent(50)
		if mo.outputManager != nil && mo.outputManager.IsDebugEnabled() {
			mo.outputManager.Debug("优化小文件操作内存")
		}

	case "search_operation":
		// 搜索操作：平衡设置
		originalGCPercent = debug.SetGCPercent(100)
		if mo.outputManager != nil && mo.outputManager.IsDebugEnabled() {
			mo.outputManager.Debug("优化搜索操作内存")
		}

	default:
		return func() {}
	}

	// 返回恢复函数
	return func() {
		debug.SetGCPercent(originalGCPercent)
		// 只在调试模式下输出恢复信息
		if mo.outputManager != nil && mo.outputManager.IsDebugEnabled() {
			mo.outputManager.Debug("恢复内存设置")
		}
	}
}

// GetMemoryStats 获取内存统计
func (mo *MemoryOptimizer) GetMemoryStats() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	stats := make(map[string]interface{})

	// 基本内存信息
	stats["alloc"] = m.Alloc
	stats["total_alloc"] = m.TotalAlloc
	stats["sys"] = m.Sys
	stats["lookups"] = m.Lookups
	stats["mallocs"] = m.Mallocs
	stats["frees"] = m.Frees

	// 堆内存信息
	stats["heap_alloc"] = m.HeapAlloc
	stats["heap_sys"] = m.HeapSys
	stats["heap_idle"] = m.HeapIdle
	stats["heap_inuse"] = m.HeapInuse
	stats["heap_released"] = m.HeapReleased
	stats["heap_objects"] = m.HeapObjects

	// GC信息
	stats["gc_cycles"] = m.NumGC
	stats["gc_cpu_fraction"] = m.GCCPUFraction
	stats["last_gc"] = time.Unix(0, int64(m.LastGC))

	// 计算使用率
	if m.Sys > 0 {
		stats["memory_usage_percent"] = float64(m.Alloc) / float64(m.Sys) * 100
	}

	return stats
}

// SetGCPercent 设置GC百分比
func (mo *MemoryOptimizer) SetGCPercent(percent int) {
	mo.mu.Lock()
	defer mo.mu.Unlock()

	mo.gcPercent = percent
	debug.SetGCPercent(percent)
	// 简化调试信息
	if mo.outputManager != nil && mo.outputManager.IsDebugEnabled() {
		mo.outputManager.Debug("设置GC: %d%%", percent)
	}
}

// SetMemoryLimit 设置内存限制
func (mo *MemoryOptimizer) SetMemoryLimit(limit int64) {
	mo.mu.Lock()
	defer mo.mu.Unlock()

	mo.memoryLimit = limit
	// 简化调试信息
	if mo.outputManager != nil && mo.outputManager.IsDebugEnabled() {
		mo.outputManager.Debug("内存限制: %d MB", limit/1024/1024)
	}
}

// SetForceGCThreshold 设置强制GC阈值
func (mo *MemoryOptimizer) SetForceGCThreshold(threshold int64) {
	mo.mu.Lock()
	defer mo.mu.Unlock()

	mo.forceGCThreshold = threshold
	// 简化调试信息
	if mo.outputManager != nil && mo.outputManager.IsDebugEnabled() {
		mo.outputManager.Debug("GC阈值: %d MB", threshold/1024/1024)
	}
}

// ForceGC 强制执行GC
func (mo *MemoryOptimizer) ForceGC() {
	if !mo.enabled {
		return
	}

	start := time.Now()
	var beforeGC runtime.MemStats
	runtime.ReadMemStats(&beforeGC)

	runtime.GC()

	var afterGC runtime.MemStats
	runtime.ReadMemStats(&afterGC)

	duration := time.Since(start)
	freed := beforeGC.Alloc - afterGC.Alloc

	// 只在调试模式下输出详细信息，并简化输出
	if mo.outputManager != nil && mo.outputManager.IsDebugEnabled() {
		mo.outputManager.Debug("GC完成: %v, 释放 %d KB", duration, freed/1024)
	}
}

// CheckMemoryPressure 检查内存压力
func (mo *MemoryOptimizer) CheckMemoryPressure() (bool, string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 检查各种内存压力指标
	if int64(m.Sys) > mo.memoryLimit {
		return true, "系统内存使用超过限制"
	}

	if int64(m.Alloc) > mo.forceGCThreshold {
		return true, "分配内存超过强制GC阈值"
	}

	// 检查GC频率
	if m.GCCPUFraction > 0.1 { // GC占用CPU超过10%
		return true, "GC占用CPU过高"
	}

	// 检查堆使用率
	if m.HeapSys > 0 {
		heapUsage := float64(m.HeapInuse) / float64(m.HeapSys)
		if heapUsage > 0.9 { // 堆使用率超过90%
			return true, "堆内存使用率过高"
		}
	}

	return false, ""
}

// OptimizeMemoryLayout 优化内存布局
func (mo *MemoryOptimizer) OptimizeMemoryLayout() {
	if !mo.enabled {
		return
	}

	// 只在调试模式下输出信息
	if mo.outputManager != nil && mo.outputManager.IsDebugEnabled() {
		mo.outputManager.Debug("优化内存布局")
	}

	// 执行GC
	runtime.GC()

	// 返回内存给操作系统
	debug.FreeOSMemory()

	// 调整GC目标
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 根据当前内存使用情况调整GC百分比
	if m.Alloc < 100*1024*1024 { // 小于100MB
		debug.SetGCPercent(200) // 降低GC频率
	} else if m.Alloc > 500*1024*1024 { // 大于500MB
		debug.SetGCPercent(50) // 提高GC频率
	} else {
		debug.SetGCPercent(100) // 默认设置
	}

	// 简化完成信息
	if mo.outputManager != nil && mo.outputManager.IsDebugEnabled() {
		mo.outputManager.Debug("内存优化完成")
	}
}

// Enable 启用内存优化器
func (mo *MemoryOptimizer) Enable() {
	mo.mu.Lock()
	defer mo.mu.Unlock()
	mo.enabled = true
}

// Disable 禁用内存优化器
func (mo *MemoryOptimizer) Disable() {
	mo.mu.Lock()
	defer mo.mu.Unlock()
	mo.enabled = false
}

// 全局内存优化器
var globalMemoryOptimizer = NewMemoryOptimizer(globalOutputManager)

// 全局函数
func StartMemoryOptimizer(ctx context.Context) {
	globalMemoryOptimizer.Start(ctx)
}

func StopMemoryOptimizer() {
	globalMemoryOptimizer.Stop()
}

func OptimizeForOperation(operation string) func() {
	return globalMemoryOptimizer.OptimizeForOperation(operation)
}

func GetMemoryStats() map[string]interface{} {
	return globalMemoryOptimizer.GetMemoryStats()
}

func ForceGC() {
	globalMemoryOptimizer.ForceGC()
}

func CheckMemoryPressure() (bool, string) {
	return globalMemoryOptimizer.CheckMemoryPressure()
}
