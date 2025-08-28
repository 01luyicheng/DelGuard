package monitor

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// SimpleDeleteDetector 简化的删除检测器
type SimpleDeleteDetector struct {
	fileMonitor   *SimpleFileMonitor
	ctx           context.Context
	cancel        context.CancelFunc
	mu            sync.RWMutex
	isRunning     bool
	deletedFiles  []DeletedFileInfo
	onDeleteFunc  func(DeletedFileInfo)
}

// DeletedFileInfo 删除文件信息
type DeletedFileInfo struct {
	Path         string
	DeleteTime   time.Time
	Size         int64
	BackupPath   string
	IsProtected  bool
}

// NewSimpleDeleteDetector 创建简化删除检测器
func NewSimpleDeleteDetector(fileMonitor *SimpleFileMonitor) *SimpleDeleteDetector {
	ctx, cancel := context.WithCancel(context.Background())
	return &SimpleDeleteDetector{
		fileMonitor:  fileMonitor,
		ctx:          ctx,
		cancel:       cancel,
		deletedFiles: make([]DeletedFileInfo, 0),
	}
}

// Start 启动删除检测
func (dd *SimpleDeleteDetector) Start() {
	dd.mu.Lock()
	defer dd.mu.Unlock()

	if dd.isRunning {
		return
	}

	dd.isRunning = true
	go dd.processEvents()
}

// Close 关闭检测器
func (dd *SimpleDeleteDetector) Close() {
	dd.mu.Lock()
	defer dd.mu.Unlock()

	if !dd.isRunning {
		return
	}

	dd.cancel()
	dd.isRunning = false
}

// SetOnDeleteCallback 设置删除回调
func (dd *SimpleDeleteDetector) SetOnDeleteCallback(callback func(DeletedFileInfo)) {
	dd.mu.Lock()
	defer dd.mu.Unlock()
	dd.onDeleteFunc = callback
}

// GetDeletedFiles 获取删除文件列表
func (dd *SimpleDeleteDetector) GetDeletedFiles() []DeletedFileInfo {
	dd.mu.RLock()
	defer dd.mu.RUnlock()
	
	result := make([]DeletedFileInfo, len(dd.deletedFiles))
	copy(result, dd.deletedFiles)
	return result
}

// processEvents 处理文件事件
func (dd *SimpleDeleteDetector) processEvents() {
	for {
		select {
		case <-dd.ctx.Done():
			return
		case event := <-dd.fileMonitor.Events():
			if event.Operation == "REMOVE" {
				dd.handleDeleteEvent(event)
			}
		}
	}
}

// handleDeleteEvent 处理删除事件
func (dd *SimpleDeleteDetector) handleDeleteEvent(event FileEvent) {
	// 检查是否为受保护的文件类型
	isProtected := dd.isProtectedFile(event.Path)
	
	deletedInfo := DeletedFileInfo{
		Path:        event.Path,
		DeleteTime:  event.Timestamp,
		Size:        event.Size,
		IsProtected: isProtected,
	}

	dd.mu.Lock()
	dd.deletedFiles = append(dd.deletedFiles, deletedInfo)
	
	// 保持最近1000个删除记录
	if len(dd.deletedFiles) > 1000 {
		dd.deletedFiles = dd.deletedFiles[len(dd.deletedFiles)-1000:]
	}
	
	callback := dd.onDeleteFunc
	dd.mu.Unlock()

	// 调用回调函数
	if callback != nil {
		callback(deletedInfo)
	}

	if isProtected {
		fmt.Printf("检测到受保护文件被删除: %s\n", event.Path)
	}
}

// isProtectedFile 检查是否为受保护文件
func (dd *SimpleDeleteDetector) isProtectedFile(path string) bool {
	// 简单的文件类型保护逻辑
	protectedExtensions := []string{
		".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
		".pdf", ".txt", ".jpg", ".jpeg", ".png", ".gif",
		".mp4", ".avi", ".mov", ".mp3", ".wav",
	}

	path = strings.ToLower(path)
	for _, ext := range protectedExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}

	return false
}