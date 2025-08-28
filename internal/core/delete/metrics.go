package delete

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics 删除操作统计信息
type Metrics struct {
	// 计数器
	TotalOperations    int64 `json:"total_operations"`
	SuccessfulDeletes  int64 `json:"successful_deletes"`
	FailedDeletes      int64 `json:"failed_deletes"`
	FilesProcessed     int64 `json:"files_processed"`
	BytesDeleted       int64 `json:"bytes_deleted"`
	
	// 时间统计
	TotalDuration      time.Duration `json:"total_duration"`
	AverageDuration    time.Duration `json:"average_duration"`
	MaxDuration        time.Duration `json:"max_duration"`
	MinDuration        time.Duration `json:"min_duration"`
	
	// 错误统计
	ErrorsByType       map[ErrorCode]int64 `json:"errors_by_type"`
	
	// 并发统计
	MaxConcurrency     int32 `json:"max_concurrency"`
	CurrentConcurrency int32 `json:"current_concurrency"`
	
	// 内部字段
	mu                 sync.RWMutex
	startTime          time.Time
	lastOperationTime  time.Time
}

// NewMetrics 创建新的统计信息
func NewMetrics() *Metrics {
	return &Metrics{
		ErrorsByType:      make(map[ErrorCode]int64),
		startTime:         time.Now(),
		lastOperationTime: time.Now(),
		MinDuration:       time.Duration(^uint64(0) >> 1), // 最大值
	}
}

// RecordOperation 记录操作
func (m *Metrics) RecordOperation(success bool, duration time.Duration, bytesDeleted int64, err error) {
	atomic.AddInt64(&m.TotalOperations, 1)
	atomic.AddInt64(&m.FilesProcessed, 1)
	
	if success {
		atomic.AddInt64(&m.SuccessfulDeletes, 1)
		atomic.AddInt64(&m.BytesDeleted, bytesDeleted)
	} else {
		atomic.AddInt64(&m.FailedDeletes, 1)
		
		// 记录错误类型
		if deleteErr, ok := err.(*DeleteError); ok {
			m.mu.Lock()
			m.ErrorsByType[deleteErr.Code]++
			m.mu.Unlock()
		}
	}
	
	// 更新时间统计
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.TotalDuration += duration
	m.lastOperationTime = time.Now()
	
	// 更新平均时间
	if m.TotalOperations > 0 {
		m.AverageDuration = time.Duration(int64(m.TotalDuration) / m.TotalOperations)
	}
	
	// 更新最大最小时间
	if duration > m.MaxDuration {
		m.MaxDuration = duration
	}
	if duration < m.MinDuration {
		m.MinDuration = duration
	}
}

// RecordConcurrency 记录并发数
func (m *Metrics) RecordConcurrency(current int32) {
	atomic.StoreInt32(&m.CurrentConcurrency, current)
	
	// 更新最大并发数
	for {
		max := atomic.LoadInt32(&m.MaxConcurrency)
		if current <= max || atomic.CompareAndSwapInt32(&m.MaxConcurrency, max, current) {
			break
		}
	}
}

// GetSuccessRate 获取成功率
func (m *Metrics) GetSuccessRate() float64 {
	total := atomic.LoadInt64(&m.TotalOperations)
	if total == 0 {
		return 0
	}
	
	successful := atomic.LoadInt64(&m.SuccessfulDeletes)
	return float64(successful) / float64(total) * 100
}

// GetThroughput 获取吞吐量 (操作/秒)
func (m *Metrics) GetThroughput() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	elapsed := time.Since(m.startTime)
	if elapsed == 0 {
		return 0
	}
	
	total := atomic.LoadInt64(&m.TotalOperations)
	return float64(total) / elapsed.Seconds()
}

// GetErrorRate 获取错误率
func (m *Metrics) GetErrorRate() float64 {
	total := atomic.LoadInt64(&m.TotalOperations)
	if total == 0 {
		return 0
	}
	
	failed := atomic.LoadInt64(&m.FailedDeletes)
	return float64(failed) / float64(total) * 100
}

// Reset 重置统计信息
func (m *Metrics) Reset() {
	atomic.StoreInt64(&m.TotalOperations, 0)
	atomic.StoreInt64(&m.SuccessfulDeletes, 0)
	atomic.StoreInt64(&m.FailedDeletes, 0)
	atomic.StoreInt64(&m.FilesProcessed, 0)
	atomic.StoreInt64(&m.BytesDeleted, 0)
	atomic.StoreInt32(&m.MaxConcurrency, 0)
	atomic.StoreInt32(&m.CurrentConcurrency, 0)
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.TotalDuration = 0
	m.AverageDuration = 0
	m.MaxDuration = 0
	m.MinDuration = time.Duration(^uint64(0) >> 1)
	m.ErrorsByType = make(map[ErrorCode]int64)
	m.startTime = time.Now()
	m.lastOperationTime = time.Now()
}

// GetSnapshot 获取统计信息快照
func (m *Metrics) GetSnapshot() *Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	snapshot := &Metrics{
		TotalOperations:    atomic.LoadInt64(&m.TotalOperations),
		SuccessfulDeletes:  atomic.LoadInt64(&m.SuccessfulDeletes),
		FailedDeletes:      atomic.LoadInt64(&m.FailedDeletes),
		FilesProcessed:     atomic.LoadInt64(&m.FilesProcessed),
		BytesDeleted:       atomic.LoadInt64(&m.BytesDeleted),
		TotalDuration:      m.TotalDuration,
		AverageDuration:    m.AverageDuration,
		MaxDuration:        m.MaxDuration,
		MinDuration:        m.MinDuration,
		MaxConcurrency:     atomic.LoadInt32(&m.MaxConcurrency),
		CurrentConcurrency: atomic.LoadInt32(&m.CurrentConcurrency),
		ErrorsByType:       make(map[ErrorCode]int64),
		startTime:          m.startTime,
		lastOperationTime:  m.lastOperationTime,
	}
	
	// 复制错误统计
	for code, count := range m.ErrorsByType {
		snapshot.ErrorsByType[code] = count
	}
	
	return snapshot
}

// String 返回统计信息的字符串表示
func (m *Metrics) String() string {
	snapshot := m.GetSnapshot()
	
	return fmt.Sprintf(`删除操作统计:
  总操作数: %d
  成功删除: %d
  失败删除: %d
  处理文件数: %d
  删除字节数: %d
  成功率: %.2f%%
  错误率: %.2f%%
  吞吐量: %.2f ops/s
  平均耗时: %v
  最大耗时: %v
  最小耗时: %v
  最大并发数: %d
  当前并发数: %d`,
		snapshot.TotalOperations,
		snapshot.SuccessfulDeletes,
		snapshot.FailedDeletes,
		snapshot.FilesProcessed,
		snapshot.BytesDeleted,
		snapshot.GetSuccessRate(),
		snapshot.GetErrorRate(),
		snapshot.GetThroughput(),
		snapshot.AverageDuration,
		snapshot.MaxDuration,
		snapshot.MinDuration,
		snapshot.MaxConcurrency,
		snapshot.CurrentConcurrency,
	)
}