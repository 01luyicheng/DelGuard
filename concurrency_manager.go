package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// ConcurrencyManager 并发管理器
type ConcurrencyManager struct {
	maxConcurrentOps int32                    // 最大并发操作数
	currentOps       int32                    // 当前操作数
	semaphore        chan struct{}            // 信号量控制并发
	mu               sync.RWMutex             // 保护管理器状态
	opMap            map[string]*Operation    // 操作映射
	opMapMu          sync.RWMutex             // 保护操作映射
	resourceLocks    map[string]*ResourceLock // 资源锁映射
	resourceMu       sync.RWMutex             // 保护资源锁映射
	ctx              context.Context
	cancel           context.CancelFunc
	metrics          *ConcurrencyMetrics
}

// Operation 操作信息
type Operation struct {
	ID        string
	Type      string
	Path      string
	StartTime time.Time
	Context   context.Context
	Cancel    context.CancelFunc
	Progress  *OperationProgress
	Error     error
	Done      chan struct{}
	mu        sync.RWMutex
}

// OperationProgress 操作进度
type OperationProgress struct {
	Total     int64
	Current   int64
	Message   string
	StartTime time.Time
	mu        sync.RWMutex
}

// ResourceLock 资源锁
type ResourceLock struct {
	Path      string
	Type      string // read, write, exclusive
	Owner     string // 操作ID
	Count     int32  // 引用计数
	CreatedAt time.Time
	mu        sync.RWMutex
}

// ConcurrencyMetrics 并发指标
type ConcurrencyMetrics struct {
	TotalOperations    int64
	ActiveOperations   int64
	CompletedOps       int64
	FailedOps          int64
	AverageOpDuration  time.Duration
	MaxConcurrencyUsed int32
	DeadlockCount      int64
	RetryCount         int64
	mu                 sync.RWMutex
}

// LockType 锁类型
type LockType string

const (
	ReadLock      LockType = "read"
	WriteLock     LockType = "write"
	ExclusiveLock LockType = "exclusive"
)

// NewConcurrencyManager 创建并发管理器
func NewConcurrencyManager(maxConcurrent int) *ConcurrencyManager {
	if maxConcurrent <= 0 {
		maxConcurrent = runtime.NumCPU() * 2 // 默认为CPU核心数的2倍
	}

	ctx, cancel := context.WithCancel(context.Background())

	cm := &ConcurrencyManager{
		maxConcurrentOps: int32(maxConcurrent),
		semaphore:        make(chan struct{}, maxConcurrent),
		opMap:            make(map[string]*Operation),
		resourceLocks:    make(map[string]*ResourceLock),
		ctx:              ctx,
		cancel:           cancel,
		metrics:          &ConcurrencyMetrics{},
	}

	// 启动清理和监控协程
	go cm.cleanupRoutine()
	go cm.metricsRoutine()

	return cm
}

// AcquireOperation 获取操作许可
func (cm *ConcurrencyManager) AcquireOperation(opType, path string) (*Operation, error) {
	// 检查是否超过最大并发数
	select {
	case cm.semaphore <- struct{}{}:
		// 成功获取许可
	case <-cm.ctx.Done():
		return nil, fmt.Errorf("并发管理器已关闭")
	default:
		return nil, fmt.Errorf("已达到最大并发操作数限制: %d", cm.maxConcurrentOps)
	}

	// 创建操作实例
	opID := generateOperationID()
	opCtx, opCancel := context.WithCancel(cm.ctx)

	op := &Operation{
		ID:        opID,
		Type:      opType,
		Path:      path,
		StartTime: time.Now(),
		Context:   opCtx,
		Cancel:    opCancel,
		Progress:  &OperationProgress{StartTime: time.Now()},
		Done:      make(chan struct{}),
	}

	// 注册操作
	cm.opMapMu.Lock()
	cm.opMap[opID] = op
	cm.opMapMu.Unlock()

	// 更新指标
	atomic.AddInt32(&cm.currentOps, 1)
	atomic.AddInt64(&cm.metrics.TotalOperations, 1)
	atomic.AddInt64(&cm.metrics.ActiveOperations, 1)

	// 更新最大并发数使用记录
	current := atomic.LoadInt32(&cm.currentOps)
	for {
		max := atomic.LoadInt32(&cm.metrics.MaxConcurrencyUsed)
		if current <= max || atomic.CompareAndSwapInt32(&cm.metrics.MaxConcurrencyUsed, max, current) {
			break
		}
	}

	return op, nil
}

// ReleaseOperation 释放操作
func (cm *ConcurrencyManager) ReleaseOperation(op *Operation) {
	if op == nil {
		return
	}

	// 标记操作完成
	select {
	case <-op.Done:
		// 操作已经完成
		return
	default:
		close(op.Done)
	}

	// 取消操作上下文
	op.Cancel()

	// 从操作映射中移除
	cm.opMapMu.Lock()
	delete(cm.opMap, op.ID)
	cm.opMapMu.Unlock()

	// 释放信号量
	select {
	case <-cm.semaphore:
		// 成功释放许可
	default:
		// 理论上不应该到这里
	}

	// 更新指标
	atomic.AddInt32(&cm.currentOps, -1)
	atomic.AddInt64(&cm.metrics.ActiveOperations, -1)

	if op.Error != nil {
		atomic.AddInt64(&cm.metrics.FailedOps, 1)
	} else {
		atomic.AddInt64(&cm.metrics.CompletedOps, 1)
	}

	// 计算操作持续时间
	duration := time.Since(op.StartTime)
	cm.updateAverageOpDuration(duration)
}

// AcquireResourceLock 获取资源锁
func (cm *ConcurrencyManager) AcquireResourceLock(path string, lockType LockType, opID string) error {
	cm.resourceMu.Lock()
	defer cm.resourceMu.Unlock()

	// 检查是否存在冲突的锁
	if existingLock, exists := cm.resourceLocks[path]; exists {
		// 检查锁冲突
		if cm.hasLockConflict(existingLock, lockType) {
			return fmt.Errorf("资源 %s 已被操作 %s 锁定为 %s", path, existingLock.Owner, existingLock.Type)
		}

		// 如果是相同类型的读锁，可以共享
		if lockType == ReadLock && existingLock.Type == string(ReadLock) {
			atomic.AddInt32(&existingLock.Count, 1)
			return nil
		}
	}

	// 创建新锁
	lock := &ResourceLock{
		Path:      path,
		Type:      string(lockType),
		Owner:     opID,
		Count:     1,
		CreatedAt: time.Now(),
	}

	cm.resourceLocks[path] = lock
	return nil
}

// ReleaseResourceLock 释放资源锁
func (cm *ConcurrencyManager) ReleaseResourceLock(path string, opID string) {
	cm.resourceMu.Lock()
	defer cm.resourceMu.Unlock()

	if lock, exists := cm.resourceLocks[path]; exists {
		if lock.Owner == opID {
			if atomic.AddInt32(&lock.Count, -1) <= 0 {
				delete(cm.resourceLocks, path)
			}
		}
	}
}

// hasLockConflict 检查锁冲突
func (cm *ConcurrencyManager) hasLockConflict(existingLock *ResourceLock, requestedType LockType) bool {
	existing := LockType(existingLock.Type)

	switch requestedType {
	case ReadLock:
		// 读锁与写锁和独占锁冲突
		return existing == WriteLock || existing == ExclusiveLock
	case WriteLock:
		// 写锁与所有类型的锁冲突
		return true
	case ExclusiveLock:
		// 独占锁与所有类型的锁冲突
		return true
	default:
		return true
	}
}

// WaitForCompletion 等待所有操作完成
func (cm *ConcurrencyManager) WaitForCompletion(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(cm.ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("等待操作完成超时")
		case <-ticker.C:
			if atomic.LoadInt32(&cm.currentOps) == 0 {
				return nil
			}
		}
	}
}

// CancelAllOperations 取消所有操作
func (cm *ConcurrencyManager) CancelAllOperations() {
	cm.opMapMu.RLock()
	operations := make([]*Operation, 0, len(cm.opMap))
	for _, op := range cm.opMap {
		operations = append(operations, op)
	}
	cm.opMapMu.RUnlock()

	// 取消所有操作
	for _, op := range operations {
		op.Cancel()
	}

	// 等待操作清理
	time.Sleep(100 * time.Millisecond)
}

// GetActiveOperations 获取活跃操作列表
func (cm *ConcurrencyManager) GetActiveOperations() []*Operation {
	cm.opMapMu.RLock()
	defer cm.opMapMu.RUnlock()

	operations := make([]*Operation, 0, len(cm.opMap))
	for _, op := range cm.opMap {
		operations = append(operations, op)
	}
	return operations
}

// GetMetrics 获取并发指标
func (cm *ConcurrencyManager) GetMetrics() *ConcurrencyMetrics {
	cm.metrics.mu.RLock()
	defer cm.metrics.mu.RUnlock()

	// 返回指标的副本
	return &ConcurrencyMetrics{
		TotalOperations:    cm.metrics.TotalOperations,
		ActiveOperations:   atomic.LoadInt64(&cm.metrics.ActiveOperations),
		CompletedOps:       atomic.LoadInt64(&cm.metrics.CompletedOps),
		FailedOps:          atomic.LoadInt64(&cm.metrics.FailedOps),
		AverageOpDuration:  cm.metrics.AverageOpDuration,
		MaxConcurrencyUsed: atomic.LoadInt32(&cm.metrics.MaxConcurrencyUsed),
		DeadlockCount:      atomic.LoadInt64(&cm.metrics.DeadlockCount),
		RetryCount:         atomic.LoadInt64(&cm.metrics.RetryCount),
	}
}

// Close 关闭并发管理器
func (cm *ConcurrencyManager) Close() error {
	cm.cancel()

	// 等待所有操作完成或超时
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// 强制取消所有操作
			cm.CancelAllOperations()
			return fmt.Errorf("关闭超时，已强制取消所有操作")
		case <-ticker.C:
			if atomic.LoadInt32(&cm.currentOps) == 0 {
				return nil
			}
		}
	}
}

// cleanupRoutine 清理协程
func (cm *ConcurrencyManager) cleanupRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-cm.ctx.Done():
			return
		case <-ticker.C:
			cm.cleanupStaleOperations()
			cm.cleanupStaleResourceLocks()
		}
	}
}

// metricsRoutine 指标更新协程
func (cm *ConcurrencyManager) metricsRoutine() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-cm.ctx.Done():
			return
		case <-ticker.C:
			cm.updateMetrics()
		}
	}
}

// cleanupStaleOperations 清理过期操作
func (cm *ConcurrencyManager) cleanupStaleOperations() {
	cm.opMapMu.Lock()
	defer cm.opMapMu.Unlock()

	now := time.Now()
	staleThreshold := 5 * time.Minute

	for id, op := range cm.opMap {
		if now.Sub(op.StartTime) > staleThreshold {
			// 操作运行时间过长，可能有问题
			select {
			case <-op.Done:
				// 操作已完成但未清理
				delete(cm.opMap, id)
			default:
				// 操作仍在运行，记录警告
				fmt.Printf("警告: 操作 %s (%s) 运行时间过长: %v\n",
					id, op.Type, now.Sub(op.StartTime))
			}
		}
	}
}

// cleanupStaleResourceLocks 清理过期资源锁
func (cm *ConcurrencyManager) cleanupStaleResourceLocks() {
	cm.resourceMu.Lock()
	defer cm.resourceMu.Unlock()

	now := time.Now()
	staleThreshold := 10 * time.Minute

	for path, lock := range cm.resourceLocks {
		if now.Sub(lock.CreatedAt) > staleThreshold {
			// 检查锁的拥有者是否还在活动
			cm.opMapMu.RLock()
			_, exists := cm.opMap[lock.Owner]
			cm.opMapMu.RUnlock()

			if !exists {
				// 拥有者操作不存在，清理锁
				delete(cm.resourceLocks, path)
				fmt.Printf("清理过期资源锁: %s (拥有者: %s)\n", path, lock.Owner)
			}
		}
	}
}

// updateMetrics 更新指标
func (cm *ConcurrencyManager) updateMetrics() {
	// 检测死锁
	cm.detectDeadlocks()
}

// detectDeadlocks 检测死锁（简化实现）
func (cm *ConcurrencyManager) detectDeadlocks() {
	cm.resourceMu.RLock()
	defer cm.resourceMu.RUnlock()

	// 简化的死锁检测：检查长时间等待的操作
	now := time.Now()
	longWaitThreshold := 2 * time.Minute

	for _, lock := range cm.resourceLocks {
		if now.Sub(lock.CreatedAt) > longWaitThreshold {
			// 可能存在死锁
			atomic.AddInt64(&cm.metrics.DeadlockCount, 1)
		}
	}
}

// updateAverageOpDuration 更新平均操作持续时间
func (cm *ConcurrencyManager) updateAverageOpDuration(duration time.Duration) {
	cm.metrics.mu.Lock()
	defer cm.metrics.mu.Unlock()

	// 简化的移动平均计算
	if cm.metrics.AverageOpDuration == 0 {
		cm.metrics.AverageOpDuration = duration
	} else {
		// 使用指数加权移动平均
		alpha := 0.1
		cm.metrics.AverageOpDuration = time.Duration(
			float64(cm.metrics.AverageOpDuration)*(1-alpha) + float64(duration)*alpha,
		)
	}
}

// UpdateProgress 更新操作进度
func (op *Operation) UpdateProgress(current, total int64, message string) {
	op.Progress.mu.Lock()
	defer op.Progress.mu.Unlock()

	op.Progress.Current = current
	op.Progress.Total = total
	op.Progress.Message = message
}

// GetProgress 获取操作进度
func (op *Operation) GetProgress() (current, total int64, message string) {
	op.Progress.mu.RLock()
	defer op.Progress.mu.RUnlock()

	return op.Progress.Current, op.Progress.Total, op.Progress.Message
}

// SetError 设置操作错误
func (op *Operation) SetError(err error) {
	op.mu.Lock()
	defer op.mu.Unlock()
	op.Error = err
}

// GetError 获取操作错误
func (op *Operation) GetError() error {
	op.mu.RLock()
	defer op.mu.RUnlock()
	return op.Error
}

// generateOperationID 生成操作ID
func generateOperationID() string {
	return fmt.Sprintf("op_%d_%d", time.Now().UnixNano(), runtime.NumGoroutine())
}

// SafeExecute 安全执行函数，带有并发控制
func (cm *ConcurrencyManager) SafeExecute(opType, path string, fn func(context.Context) error) error {
	// 获取操作许可
	op, err := cm.AcquireOperation(opType, path)
	if err != nil {
		return fmt.Errorf("获取操作许可失败: %v", err)
	}
	defer cm.ReleaseOperation(op)

	// 获取资源锁
	var lockType LockType
	switch opType {
	case "delete", "move":
		lockType = WriteLock
	case "read", "stat":
		lockType = ReadLock
	default:
		lockType = ExclusiveLock
	}

	if err := cm.AcquireResourceLock(path, lockType, op.ID); err != nil {
		return fmt.Errorf("获取资源锁失败: %v", err)
	}
	defer cm.ReleaseResourceLock(path, op.ID)

	// 执行操作
	err = fn(op.Context)
	if err != nil {
		op.SetError(err)
	}

	return err
}
