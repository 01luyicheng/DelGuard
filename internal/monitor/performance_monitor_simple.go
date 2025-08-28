package monitor

import (
	"context"
	"runtime"
	"sync"
	"time"
)

// SimplePerformanceMonitor 简化的性能监控器
type SimplePerformanceMonitor struct {
	ctx           context.Context
	cancel        context.CancelFunc
	mu            sync.RWMutex
	isRunning     bool
	stats         *PerformanceStats
	startTime     time.Time
	operations    map[string][]time.Duration
}

// PerformanceStats 性能统计
type PerformanceStats struct {
	CPUUsage       float64
	MemoryUsage    uint64
	MonitoredFiles int
	DeletedFiles   int
	Uptime         time.Duration
	LastUpdate     time.Time
}

// NewSimplePerformanceMonitor 创建简化性能监控器
func NewSimplePerformanceMonitor() *SimplePerformanceMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	return &SimplePerformanceMonitor{
		ctx:        ctx,
		cancel:     cancel,
		stats:      &PerformanceStats{},
		startTime:  time.Now(),
		operations: make(map[string][]time.Duration),
	}
}

// Start 启动性能监控
func (pm *SimplePerformanceMonitor) Start() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.isRunning {
		return
	}

	pm.isRunning = true
	pm.startTime = time.Now()
	go pm.collectStats()
}

// Close 关闭性能监控
func (pm *SimplePerformanceMonitor) Close() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.isRunning {
		return
	}

	pm.cancel()
	pm.isRunning = false
}

// GetStats 获取性能统计
func (pm *SimplePerformanceMonitor) GetStats() *PerformanceStats {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if pm.stats == nil {
		return &PerformanceStats{}
	}

	// 返回副本
	stats := *pm.stats
	return &stats
}

// RecordOperation 记录操作性能
func (pm *SimplePerformanceMonitor) RecordOperation(operation string, duration time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.operations[operation] == nil {
		pm.operations[operation] = make([]time.Duration, 0)
	}

	pm.operations[operation] = append(pm.operations[operation], duration)

	// 保持最近100个记录
	if len(pm.operations[operation]) > 100 {
		pm.operations[operation] = pm.operations[operation][len(pm.operations[operation])-100:]
	}
}

// collectStats 收集性能统计
func (pm *SimplePerformanceMonitor) collectStats() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-pm.ctx.Done():
			return
		case <-ticker.C:
			pm.updateStats()
		}
	}
}

// updateStats 更新统计信息
func (pm *SimplePerformanceMonitor) updateStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.stats.MemoryUsage = m.Alloc
	pm.stats.CPUUsage = pm.calculateCPUUsage()
	pm.stats.Uptime = time.Since(pm.startTime)
	pm.stats.LastUpdate = time.Now()
}

// calculateCPUUsage 计算CPU使用率（简化版本）
func (pm *SimplePerformanceMonitor) calculateCPUUsage() float64 {
	// 简化的CPU使用率计算
	// 在实际应用中，这里应该使用更精确的方法
	return float64(runtime.NumGoroutine()) * 0.1
}

// GetOperationStats 获取操作统计
func (pm *SimplePerformanceMonitor) GetOperationStats(operation string) []time.Duration {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if durations, exists := pm.operations[operation]; exists {
		result := make([]time.Duration, len(durations))
		copy(result, durations)
		return result
	}

	return nil
}

// GetAverageOperationTime 获取平均操作时间
func (pm *SimplePerformanceMonitor) GetAverageOperationTime(operation string) time.Duration {
	durations := pm.GetOperationStats(operation)
	if len(durations) == 0 {
		return 0
	}

	var total time.Duration
	for _, d := range durations {
		total += d
	}

	return total / time.Duration(len(durations))
}