package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

// Constants for default limits
const (
	// DefaultMemoryLimit 默认内存限制 (1GB)
	DefaultMemoryLimit = 1024 * 1024 * 1024

	// DefaultFileLimit 默认文件句柄限制
	DefaultFileLimit = 1000

	// DefaultTempFileLimit 默认临时文件限制
	DefaultTempFileLimit = 100

	// GCTimeoutInterval GC协程运行间隔时间
	GCTimeoutInterval = 30 * time.Second

	// CleanupTimeoutInterval 清理协程运行间隔时间
	CleanupTimeoutInterval = 60 * time.Second

	// MonitorTimeoutInterval 监控协程运行间隔时间
	MonitorTimeoutInterval = 10 * time.Second

	// UnusedFileThreshold 未使用文件清理阈值
	UnusedFileThreshold = 5 * time.Minute

	// UnusedPoolThreshold 未使用内存池清理阈值
	UnusedPoolThreshold = 10 * time.Minute
)

// ResourceManager 资源管理器
type ResourceManager struct {
	mu             sync.RWMutex
	openFiles      map[string]*ManagedFile  // 打开的文件映射
	openDirs       map[string]*ManagedDir   // 打开的目录映射
	memoryPools    map[string]*MemoryPool   // 内存池映射
	tempFiles      map[string]*TempFileInfo // 临时文件映射
	cleanupTasks   []CleanupTask            // 清理任务列表
	metrics        *ResourceMetrics         // 资源指标
	ctx            context.Context
	cancel         context.CancelFunc
	gcTicker       *time.Ticker // GC定时器
	cleanupTicker  *time.Ticker // 清理定时器
	memoryLimit    int64        // 内存限制
	fileLimit      int32        // 文件句柄限制
	tempFileLimit  int32        // 临时文件限制
	isShuttingDown int32        // 关闭状态标志
}

// ManagedFile 管理的文件
type ManagedFile struct {
	File      *os.File
	Path      string
	Mode      string
	CreatedAt time.Time
	AccessAt  time.Time
	RefCount  int32
	Size      int64
	mu        sync.RWMutex
}

// ManagedDir 管理的目录
type ManagedDir struct {
	File      *os.File
	Path      string
	CreatedAt time.Time
	AccessAt  time.Time
	RefCount  int32
	mu        sync.RWMutex
}

// MemoryPool 内存池
type MemoryPool struct {
	name       string
	buffers    [][]byte
	maxSize    int
	bufferSize int
	inUse      int32
	total      int32
	mu         sync.Mutex
	createdAt  time.Time
}

// TempFileInfo 临时文件信息
type TempFileInfo struct {
	Path      string
	File      *os.File
	Size      int64
	CreatedAt time.Time
	TTL       time.Duration
	AutoClean bool
	mu        sync.RWMutex
}

// ResourceMetrics 资源指标
type ResourceMetrics struct {
	OpenFiles      int32
	OpenDirs       int32
	TempFiles      int32
	MemoryUsage    int64
	MaxMemoryUsage int64
	GCRuns         int64
	CleanupRuns    int64
	FileLeaks      int64
	MemoryLeaks    int64
	mu             sync.RWMutex
}

// CleanupTask 清理任务
type CleanupTask struct {
	Name     string
	Function func() error
	Priority int
	RunAt    time.Time
}

// FileResource 文件资源接口
type FileResource interface {
	Close() error
	Name() string
	Size() int64
}

// MemoryResource 内存资源接口
type MemoryResource interface {
	Release()
	Size() int
}

// NewResourceManager 创建资源管理器
//
// 返回值:
//   - *ResourceManager: 资源管理器实例指针
func NewResourceManager() *ResourceManager {
	ctx, cancel := context.WithCancel(context.Background())

	rm := &ResourceManager{
		openFiles:     make(map[string]*ManagedFile),
		openDirs:      make(map[string]*ManagedDir),
		memoryPools:   make(map[string]*MemoryPool),
		tempFiles:     make(map[string]*TempFileInfo),
		cleanupTasks:  make([]CleanupTask, 0),
		metrics:       &ResourceMetrics{},
		ctx:           ctx,
		cancel:        cancel,
		memoryLimit:   DefaultMemoryLimit,   // 1GB默认限制
		fileLimit:     DefaultFileLimit,     // 1000个文件句柄限制
		tempFileLimit: DefaultTempFileLimit, // 100个临时文件限制
	}

	// 启动清理和监控协程
	rm.gcTicker = time.NewTicker(GCTimeoutInterval)
	rm.cleanupTicker = time.NewTicker(CleanupTimeoutInterval)

	go rm.gcRoutine()
	go rm.cleanupRoutine()
	go rm.monitorRoutine()

	return rm
}

// OpenFile 打开文件并管理
func (rm *ResourceManager) OpenFile(path string, flag int, perm os.FileMode) (*ManagedFile, error) {
	if atomic.LoadInt32(&rm.isShuttingDown) != 0 {
		return nil, fmt.Errorf("资源管理器正在关闭")
	}

	// 检查文件句柄限制
	if atomic.LoadInt32(&rm.metrics.OpenFiles) >= rm.fileLimit {
		// 尝试清理未使用的文件
		rm.cleanupUnusedFiles()
		if atomic.LoadInt32(&rm.metrics.OpenFiles) >= rm.fileLimit {
			return nil, fmt.Errorf("文件句柄数量已达上限: %d", rm.fileLimit)
		}
	}

	file, err := os.OpenFile(path, flag, perm)
	if err != nil {
		return nil, err
	}

	// 获取文件信息
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	managedFile := &ManagedFile{
		File:      file,
		Path:      path,
		Mode:      fmt.Sprintf("%v", flag),
		CreatedAt: time.Now(),
		AccessAt:  time.Now(),
		RefCount:  1,
		Size:      info.Size(),
	}

	rm.mu.Lock()
	rm.openFiles[path] = managedFile
	rm.mu.Unlock()

	atomic.AddInt32(&rm.metrics.OpenFiles, 1)
	return managedFile, nil
}

// CloseFile 关闭文件
func (rm *ResourceManager) CloseFile(path string) error {
	rm.mu.Lock()
	managedFile, exists := rm.openFiles[path]
	if !exists {
		rm.mu.Unlock()
		return fmt.Errorf("文件 %s 未打开", path)
	}
	delete(rm.openFiles, path)
	rm.mu.Unlock()

	err := managedFile.File.Close()
	atomic.AddInt32(&rm.metrics.OpenFiles, -1)
	return err
}

// GetFile 获取已打开的文件
func (rm *ResourceManager) GetFile(path string) (*ManagedFile, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	managedFile, exists := rm.openFiles[path]
	if exists {
		managedFile.mu.Lock()
		managedFile.AccessAt = time.Now()
		atomic.AddInt32(&managedFile.RefCount, 1)
		managedFile.mu.Unlock()
	}
	return managedFile, exists
}

// CreateTempFile 创建临时文件
func (rm *ResourceManager) CreateTempFile(dir, pattern string, ttl time.Duration) (*TempFileInfo, error) {
	if atomic.LoadInt32(&rm.metrics.TempFiles) >= rm.tempFileLimit {
		rm.cleanupExpiredTempFiles()
		if atomic.LoadInt32(&rm.metrics.TempFiles) >= rm.tempFileLimit {
			return nil, fmt.Errorf("临时文件数量已达上限: %d", rm.tempFileLimit)
		}
	}

	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return nil, err
	}

	tempInfo := &TempFileInfo{
		Path:      file.Name(),
		File:      file,
		CreatedAt: time.Now(),
		TTL:       ttl,
		AutoClean: true,
	}

	rm.mu.Lock()
	rm.tempFiles[tempInfo.Path] = tempInfo
	rm.mu.Unlock()

	atomic.AddInt32(&rm.metrics.TempFiles, 1)

	// 添加清理任务
	if ttl > 0 {
		rm.addCleanupTask(CleanupTask{
			Name:     fmt.Sprintf("cleanup_temp_%s", tempInfo.Path),
			Function: func() error { return rm.removeTempFile(tempInfo.Path) },
			Priority: 1,
			RunAt:    time.Now().Add(ttl),
		})
	}

	return tempInfo, nil
}

// GetMemoryPool 获取或创建内存池
func (rm *ResourceManager) GetMemoryPool(name string, bufferSize, maxBuffers int) *MemoryPool {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if pool, exists := rm.memoryPools[name]; exists {
		return pool
	}

	pool := &MemoryPool{
		name:       name,
		buffers:    make([][]byte, 0, maxBuffers),
		maxSize:    maxBuffers,
		bufferSize: bufferSize,
		createdAt:  time.Now(),
	}

	rm.memoryPools[name] = pool
	return pool
}

// Get 从内存池获取缓冲区
func (mp *MemoryPool) Get() []byte {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if len(mp.buffers) > 0 {
		buffer := mp.buffers[len(mp.buffers)-1]
		mp.buffers = mp.buffers[:len(mp.buffers)-1]
		atomic.AddInt32(&mp.inUse, 1)
		return buffer
	}

	// 创建新缓冲区
	buffer := make([]byte, mp.bufferSize)
	atomic.AddInt32(&mp.total, 1)
	atomic.AddInt32(&mp.inUse, 1)
	return buffer
}

// Put 归还缓冲区到内存池
func (mp *MemoryPool) Put(buffer []byte) {
	if len(buffer) != mp.bufferSize {
		// 大小不匹配，直接丢弃
		atomic.AddInt32(&mp.inUse, -1)
		return
	}

	mp.mu.Lock()
	defer mp.mu.Unlock()

	if len(mp.buffers) < mp.maxSize {
		// 清理缓冲区
		for i := range buffer {
			buffer[i] = 0
		}
		mp.buffers = append(mp.buffers, buffer)
	}

	atomic.AddInt32(&mp.inUse, -1)
}

// addCleanupTask 添加清理任务
func (rm *ResourceManager) addCleanupTask(task CleanupTask) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.cleanupTasks = append(rm.cleanupTasks, task)
}

// gcRoutine GC协程
func (rm *ResourceManager) gcRoutine() {
	defer rm.gcTicker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-rm.gcTicker.C:
			rm.runGC()
		}
	}
}

// cleanupRoutine 清理协程
func (rm *ResourceManager) cleanupRoutine() {
	defer rm.cleanupTicker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-rm.cleanupTicker.C:
			rm.runCleanup()
		}
	}
}

// monitorRoutine 监控协程
func (rm *ResourceManager) monitorRoutine() {
	ticker := time.NewTicker(MonitorTimeoutInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			rm.updateMetrics()
			rm.checkMemoryUsage()
			rm.detectLeaks()
		}
	}
}

// runGC 执行垃圾回收
func (rm *ResourceManager) runGC() {
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	runtime.GC()
	debug.FreeOSMemory()

	runtime.ReadMemStats(&m2)
	atomic.AddInt64(&rm.metrics.GCRuns, 1)

	// 更新内存使用统计
	memUsage := int64(m2.Alloc)
	atomic.StoreInt64(&rm.metrics.MemoryUsage, memUsage)

	// 更新最大内存使用
	for {
		max := atomic.LoadInt64(&rm.metrics.MaxMemoryUsage)
		if memUsage <= max || atomic.CompareAndSwapInt64(&rm.metrics.MaxMemoryUsage, max, memUsage) {
			break
		}
	}
}

// runCleanup 执行清理任务
func (rm *ResourceManager) runCleanup() {
	atomic.AddInt64(&rm.metrics.CleanupRuns, 1)

	rm.cleanupUnusedFiles()
	rm.cleanupExpiredTempFiles()
	rm.cleanupMemoryPools()
	rm.runScheduledCleanupTasks()
}

// cleanupUnusedFiles 清理未使用的文件
func (rm *ResourceManager) cleanupUnusedFiles() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	now := time.Now()
	unusedThreshold := 5 * time.Minute

	for path, managedFile := range rm.openFiles {
		managedFile.mu.RLock()
		refCount := atomic.LoadInt32(&managedFile.RefCount)
		lastAccess := managedFile.AccessAt
		managedFile.mu.RUnlock()

		if refCount <= 0 && now.Sub(lastAccess) > unusedThreshold {
			// 文件未被使用且超过阈值时间
			managedFile.File.Close()
			delete(rm.openFiles, path)
			atomic.AddInt32(&rm.metrics.OpenFiles, -1)
		}
	}
}

// cleanupExpiredTempFiles 清理过期临时文件
func (rm *ResourceManager) cleanupExpiredTempFiles() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	now := time.Now()

	for path, tempInfo := range rm.tempFiles {
		tempInfo.mu.RLock()
		expired := tempInfo.TTL > 0 && now.Sub(tempInfo.CreatedAt) > tempInfo.TTL
		autoClean := tempInfo.AutoClean
		tempInfo.mu.RUnlock()

		if expired && autoClean {
			rm.removeTempFileUnsafe(path)
		}
	}
}

// cleanupMemoryPools 清理内存池
func (rm *ResourceManager) cleanupMemoryPools() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	now := time.Now()
	unusedThreshold := 10 * time.Minute

	for name, pool := range rm.memoryPools {
		inUse := atomic.LoadInt32(&pool.inUse)
		if inUse == 0 && now.Sub(pool.createdAt) > unusedThreshold {
			// 内存池未被使用，清理它
			pool.mu.Lock()
			pool.buffers = pool.buffers[:0]
			pool.mu.Unlock()
			delete(rm.memoryPools, name)
		}
	}
}

// runScheduledCleanupTasks 执行计划的清理任务
func (rm *ResourceManager) runScheduledCleanupTasks() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	now := time.Now()
	pendingTasks := make([]CleanupTask, 0)

	for _, task := range rm.cleanupTasks {
		if now.After(task.RunAt) {
			// 执行清理任务
			if err := task.Function(); err != nil {
				fmt.Printf("清理任务 %s 执行失败: %v\n", task.Name, err)
			}
		} else {
			pendingTasks = append(pendingTasks, task)
		}
	}

	rm.cleanupTasks = pendingTasks
}

// removeTempFile 移除临时文件
func (rm *ResourceManager) removeTempFile(path string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	return rm.removeTempFileUnsafe(path)
}

// removeTempFileUnsafe 移除临时文件（不加锁版本）
func (rm *ResourceManager) removeTempFileUnsafe(path string) error {
	tempInfo, exists := rm.tempFiles[path]
	if !exists {
		return fmt.Errorf("临时文件 %s 不存在", path)
	}

	var err error
	if tempInfo.File != nil {
		err = tempInfo.File.Close()
	}

	if removeErr := os.Remove(path); removeErr != nil && err == nil {
		err = removeErr
	}

	delete(rm.tempFiles, path)
	atomic.AddInt32(&rm.metrics.TempFiles, -1)
	return err
}

// updateMetrics 更新指标
func (rm *ResourceManager) updateMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memUsage := int64(m.Alloc)
	atomic.StoreInt64(&rm.metrics.MemoryUsage, memUsage)
}

// checkMemoryUsage 检查内存使用
func (rm *ResourceManager) checkMemoryUsage() {
	currentUsage := atomic.LoadInt64(&rm.metrics.MemoryUsage)

	if currentUsage > rm.memoryLimit {
		// 内存使用超过限制，强制GC
		rm.runGC()

		// 如果仍然超过限制，清理缓存
		if atomic.LoadInt64(&rm.metrics.MemoryUsage) > rm.memoryLimit {
			rm.forceCleanup()
		}
	}
}

// detectLeaks 检测泄漏
func (rm *ResourceManager) detectLeaks() {
	// 检测文件句柄泄漏
	rm.mu.RLock()
	openFileCount := len(rm.openFiles)
	tempFileCount := len(rm.tempFiles)
	rm.mu.RUnlock()

	actualOpenFiles := atomic.LoadInt32(&rm.metrics.OpenFiles)
	actualTempFiles := atomic.LoadInt32(&rm.metrics.TempFiles)

	if int32(openFileCount) != actualOpenFiles {
		atomic.AddInt64(&rm.metrics.FileLeaks, 1)
	}

	if int32(tempFileCount) != actualTempFiles {
		atomic.AddInt64(&rm.metrics.FileLeaks, 1)
	}
}

// forceCleanup 强制清理
func (rm *ResourceManager) forceCleanup() {
	// 清理所有内存池
	rm.mu.Lock()
	for _, pool := range rm.memoryPools {
		pool.mu.Lock()
		pool.buffers = pool.buffers[:0]
		pool.mu.Unlock()
	}
	rm.mu.Unlock()

	// 强制GC
	runtime.GC()
	debug.FreeOSMemory()
}

// GetMetrics 获取资源指标
func (rm *ResourceManager) GetMetrics() *ResourceMetrics {
	return &ResourceMetrics{
		OpenFiles:      atomic.LoadInt32(&rm.metrics.OpenFiles),
		OpenDirs:       atomic.LoadInt32(&rm.metrics.OpenDirs),
		TempFiles:      atomic.LoadInt32(&rm.metrics.TempFiles),
		MemoryUsage:    atomic.LoadInt64(&rm.metrics.MemoryUsage),
		MaxMemoryUsage: atomic.LoadInt64(&rm.metrics.MaxMemoryUsage),
		GCRuns:         atomic.LoadInt64(&rm.metrics.GCRuns),
		CleanupRuns:    atomic.LoadInt64(&rm.metrics.CleanupRuns),
		FileLeaks:      atomic.LoadInt64(&rm.metrics.FileLeaks),
		MemoryLeaks:    atomic.LoadInt64(&rm.metrics.MemoryLeaks),
	}
}

// Close 关闭资源管理器
func (rm *ResourceManager) Close() error {
	atomic.StoreInt32(&rm.isShuttingDown, 1)
	rm.cancel()

	// 等待清理协程结束
	time.Sleep(100 * time.Millisecond)

	// 关闭所有打开的文件
	rm.mu.Lock()
	for path, managedFile := range rm.openFiles {
		managedFile.File.Close()
		delete(rm.openFiles, path)
	}

	for path, managedDir := range rm.openDirs {
		managedDir.File.Close()
		delete(rm.openDirs, path)
	}

	// 清理所有临时文件
	for path := range rm.tempFiles {
		rm.removeTempFileUnsafe(path)
	}

	// 清理内存池
	for _, pool := range rm.memoryPools {
		pool.mu.Lock()
		pool.buffers = pool.buffers[:0]
		pool.mu.Unlock()
	}
	rm.mu.Unlock()

	// 最终GC
	runtime.GC()
	debug.FreeOSMemory()

	return nil
}

// AddRef 增加文件引用计数
func (mf *ManagedFile) AddRef() {
	atomic.AddInt32(&mf.RefCount, 1)
	mf.mu.Lock()
	mf.AccessAt = time.Now()
	mf.mu.Unlock()
}

// Release 释放文件引用
func (mf *ManagedFile) Release() {
	atomic.AddInt32(&mf.RefCount, -1)
}

// Read 读取文件
func (mf *ManagedFile) Read(p []byte) (n int, err error) {
	mf.mu.Lock()
	mf.AccessAt = time.Now()
	mf.mu.Unlock()
	return mf.File.Read(p)
}

// Write 写入文件
func (mf *ManagedFile) Write(p []byte) (n int, err error) {
	mf.mu.Lock()
	mf.AccessAt = time.Now()
	mf.mu.Unlock()
	return mf.File.Write(p)
}

// GetStats 获取统计信息
func (mf *ManagedFile) GetStats() (path string, size int64, refCount int32, created, accessed time.Time) {
	mf.mu.RLock()
	defer mf.mu.RUnlock()

	return mf.Path, mf.Size, atomic.LoadInt32(&mf.RefCount), mf.CreatedAt, mf.AccessAt
}
