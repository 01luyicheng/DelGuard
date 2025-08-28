package monitor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SimpleFileMonitor 简化的文件监控器
type SimpleFileMonitor struct {
	watchPaths    map[string]bool
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	eventChan     chan FileEvent
	isRunning     bool
	pollInterval  time.Duration
	fileStates    map[string]os.FileInfo
}

// FileEvent 文件事件
type FileEvent struct {
	Path      string
	Operation string
	Timestamp time.Time
	Size      int64
}

// NewSimpleFileMonitor 创建简化文件监控器
func NewSimpleFileMonitor() *SimpleFileMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	return &SimpleFileMonitor{
		watchPaths:   make(map[string]bool),
		ctx:          ctx,
		cancel:       cancel,
		eventChan:    make(chan FileEvent, 100),
		pollInterval: 5 * time.Second,
		fileStates:   make(map[string]os.FileInfo),
	}
}

// Start 启动监控
func (fm *SimpleFileMonitor) Start() error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if fm.isRunning {
		return fmt.Errorf("文件监控器已在运行")
	}

	fm.isRunning = true
	go fm.pollFiles()
	return nil
}

// AddPath 添加监控路径
func (fm *SimpleFileMonitor) AddPath(path string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	fm.watchPaths[absPath] = true
	return nil
}

// RemovePath 移除监控路径
func (fm *SimpleFileMonitor) RemovePath(path string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	delete(fm.watchPaths, absPath)
	return nil
}

// Events 获取事件通道
func (fm *SimpleFileMonitor) Events() <-chan FileEvent {
	return fm.eventChan
}

// Close 关闭监控器
func (fm *SimpleFileMonitor) Close() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if !fm.isRunning {
		return
	}

	fm.cancel()
	fm.isRunning = false
	close(fm.eventChan)
}

// pollFiles 轮询文件变化
func (fm *SimpleFileMonitor) pollFiles() {
	ticker := time.NewTicker(fm.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-fm.ctx.Done():
			return
		case <-ticker.C:
			fm.checkFiles()
		}
	}
}

// checkFiles 检查文件变化
func (fm *SimpleFileMonitor) checkFiles() {
	fm.mu.RLock()
	paths := make([]string, 0, len(fm.watchPaths))
	for path := range fm.watchPaths {
		paths = append(paths, path)
	}
	fm.mu.RUnlock()

	for _, path := range paths {
		fm.checkPath(path)
	}
}

// checkPath 检查单个路径
func (fm *SimpleFileMonitor) checkPath(path string) {
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			// 文件可能被删除
			if oldInfo, exists := fm.fileStates[filePath]; exists {
				delete(fm.fileStates, filePath)
				fm.sendEvent(FileEvent{
					Path:      filePath,
					Operation: "REMOVE",
					Timestamp: time.Now(),
					Size:      oldInfo.Size(),
				})
			}
			return nil
		}

		oldInfo, exists := fm.fileStates[filePath]
		if !exists {
			// 新文件
			fm.fileStates[filePath] = info
			fm.sendEvent(FileEvent{
				Path:      filePath,
				Operation: "CREATE",
				Timestamp: time.Now(),
				Size:      info.Size(),
			})
		} else if info.ModTime().After(oldInfo.ModTime()) {
			// 文件被修改
			fm.fileStates[filePath] = info
			fm.sendEvent(FileEvent{
				Path:      filePath,
				Operation: "WRITE",
				Timestamp: time.Now(),
				Size:      info.Size(),
			})
		}

		return nil
	})

	if err != nil {
		// 路径可能被删除
		for filePath := range fm.fileStates {
			if filepath.HasPrefix(filePath, path) {
				delete(fm.fileStates, filePath)
				fm.sendEvent(FileEvent{
					Path:      filePath,
					Operation: "REMOVE",
					Timestamp: time.Now(),
					Size:      0,
				})
			}
		}
	}
}

// sendEvent 发送事件
func (fm *SimpleFileMonitor) sendEvent(event FileEvent) {
	select {
	case fm.eventChan <- event:
	default:
		// 通道满了，丢弃事件
	}
}