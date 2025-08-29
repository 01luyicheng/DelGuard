package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MonitorManager 监控管理器
type MonitorManager struct {
	fileMonitor       *SimpleFileMonitor
	deleteDetector    *SimpleDeleteDetector
	performanceMonitor *SimplePerformanceMonitor
	ctx               context.Context
	cancel            context.CancelFunc
	mu                sync.RWMutex
	isRunning         bool
}

// NewMonitorManager 创建监控管理器
func NewMonitorManager() *MonitorManager {
	ctx, cancel := context.WithCancel(context.Background())

	// 创建文件监控器
	fileMonitor := NewSimpleFileMonitor()

	// 创建删除检测器
	deleteDetector := NewSimpleDeleteDetector(fileMonitor)

	// 创建性能监控器
	performanceMonitor := NewSimplePerformanceMonitor()

	return &MonitorManager{
		fileMonitor:        fileMonitor,
		deleteDetector:     deleteDetector,
		performanceMonitor: performanceMonitor,
		ctx:                ctx,
		cancel:             cancel,
		isRunning:          false,
	}
}

// Start 启动监控
func (mm *MonitorManager) Start() error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if mm.isRunning {
		return fmt.Errorf("监控管理器已在运行")
	}

	// 启动文件监控
	if err := mm.fileMonitor.Start(); err != nil {
		return fmt.Errorf("启动文件监控失败: %v", err)
	}

	// 启动删除检测器
	mm.deleteDetector.Start()

	// 启动性能监控器
	mm.performanceMonitor.Start()

	mm.isRunning = true
	fmt.Println("监控管理器已启动")
	return nil
}

// Stop 停止监控
func (mm *MonitorManager) Stop() error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if !mm.isRunning {
		return nil
	}

	// 停止所有监控器
	mm.fileMonitor.Close()
	mm.deleteDetector.Close()
	mm.performanceMonitor.Close()
	mm.cancel()

	mm.isRunning = false
	fmt.Println("监控管理器已停止")
	return nil
}

// AddWatchPath 添加监控路径
func (mm *MonitorManager) AddWatchPath(path string) error {
	return mm.fileMonitor.AddPath(path)
}

// RemoveWatchPath 移除监控路径
func (mm *MonitorManager) RemoveWatchPath(path string) error {
	return mm.fileMonitor.RemovePath(path)
}

// GetPerformanceStats 获取性能统计
func (mm *MonitorManager) GetPerformanceStats() *PerformanceStats {
	return mm.performanceMonitor.GetStats()
}

// IsRunning 检查是否运行中
func (mm *MonitorManager) IsRunning() bool {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	return mm.isRunning
}

// 兼容性函数 - 保持向后兼容
var globalManager *MonitorManager

// StartMemoryOptimizer 启动内存优化器
func StartMemoryOptimizer(ctx context.Context) {
	if globalManager == nil {
		globalManager = NewMonitorManager()
	}
	// 内存优化已集成在性能监控器中
}

// StopMemoryOptimizer 停止内存优化器
func StopMemoryOptimizer() {
	// 内存优化已集成在性能监控器中
}

// MeasureOperation 测量操作性能
func MeasureOperation(ctx context.Context, operation string, fn func() error) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if globalManager != nil && globalManager.performanceMonitor != nil {
			globalManager.performanceMonitor.RecordOperation(operation, duration)
		}
	}()

	return fn()
}