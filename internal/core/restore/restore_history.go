package restore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// RestoreRecord 恢复记录
type RestoreRecord struct {
	ID            string    `json:"id"`
	Timestamp     time.Time `json:"timestamp"`
	SourcePath    string    `json:"source_path"`
	DestPath      string    `json:"dest_path"`
	FileSize      int64     `json:"file_size"`
	Checksum      string    `json:"checksum"`
	Success       bool      `json:"success"`
	Error         string    `json:"error,omitempty"`
	BackupPath    string    `json:"backup_path,omitempty"`
	Duration      int64     `json:"duration_ms"`
	RestoreMethod string    `json:"restore_method"`
}

// RestoreSession 恢复会话
type RestoreSession struct {
	ID          string          `json:"id"`
	StartTime   time.Time       `json:"start_time"`
	EndTime     time.Time       `json:"end_time"`
	TotalFiles  int             `json:"total_files"`
	SuccessCount int            `json:"success_count"`
	FailedCount int             `json:"failed_count"`
	Records     []RestoreRecord `json:"records"`
	Options     RestoreOptions  `json:"options"`
}

// RestoreHistoryManager 恢复历史管理器
type RestoreHistoryManager struct {
	historyFile string
	sessions    map[string]*RestoreSession
	mu          sync.RWMutex
}

// NewRestoreHistoryManager 创建恢复历史管理器
func NewRestoreHistoryManager(historyDir string) (*RestoreHistoryManager, error) {
	if err := os.MkdirAll(historyDir, 0755); err != nil {
		return nil, fmt.Errorf("创建历史目录失败: %v", err)
	}

	historyFile := filepath.Join(historyDir, "restore_history.json")
	
	manager := &RestoreHistoryManager{
		historyFile: historyFile,
		sessions:    make(map[string]*RestoreSession),
	}

	// 加载现有历史记录
	if err := manager.loadHistory(); err != nil {
		return nil, fmt.Errorf("加载历史记录失败: %v", err)
	}

	return manager, nil
}

// StartSession 开始恢复会话
func (rhm *RestoreHistoryManager) StartSession(options RestoreOptions) string {
	rhm.mu.Lock()
	defer rhm.mu.Unlock()

	sessionID := fmt.Sprintf("session_%d", time.Now().UnixNano())
	session := &RestoreSession{
		ID:        sessionID,
		StartTime: time.Now(),
		Records:   make([]RestoreRecord, 0),
		Options:   options,
	}

	rhm.sessions[sessionID] = session
	return sessionID
}

// AddRecord 添加恢复记录
func (rhm *RestoreHistoryManager) AddRecord(sessionID string, record RestoreRecord) error {
	rhm.mu.Lock()
	defer rhm.mu.Unlock()

	session, exists := rhm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("会话不存在: %s", sessionID)
	}

	record.ID = fmt.Sprintf("record_%d", time.Now().UnixNano())
	record.Timestamp = time.Now()

	session.Records = append(session.Records, record)
	session.TotalFiles = len(session.Records)

	if record.Success {
		session.SuccessCount++
	} else {
		session.FailedCount++
	}

	return nil
}

// EndSession 结束恢复会话
func (rhm *RestoreHistoryManager) EndSession(sessionID string) error {
	rhm.mu.Lock()
	defer rhm.mu.Unlock()

	session, exists := rhm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("会话不存在: %s", sessionID)
	}

	session.EndTime = time.Now()

	// 保存历史记录
	return rhm.saveHistory()
}

// GetSession 获取会话信息
func (rhm *RestoreHistoryManager) GetSession(sessionID string) (*RestoreSession, error) {
	rhm.mu.RLock()
	defer rhm.mu.RUnlock()

	session, exists := rhm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("会话不存在: %s", sessionID)
	}

	// 返回副本
	sessionCopy := *session
	return &sessionCopy, nil
}

// ListSessions 列出所有会话
func (rhm *RestoreHistoryManager) ListSessions() []*RestoreSession {
	rhm.mu.RLock()
	defer rhm.mu.RUnlock()

	sessions := make([]*RestoreSession, 0, len(rhm.sessions))
	for _, session := range rhm.sessions {
		sessionCopy := *session
		sessions = append(sessions, &sessionCopy)
	}

	// 按时间排序
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartTime.After(sessions[j].StartTime)
	})

	return sessions
}

// GetRecentSessions 获取最近的会话
func (rhm *RestoreHistoryManager) GetRecentSessions(limit int) []*RestoreSession {
	sessions := rhm.ListSessions()
	if len(sessions) > limit {
		sessions = sessions[:limit]
	}
	return sessions
}

// RollbackSession 回滚会话
func (rhm *RestoreHistoryManager) RollbackSession(sessionID string) error {
	session, err := rhm.GetSession(sessionID)
	if err != nil {
		return err
	}

	var rollbackErrors []string

	for _, record := range session.Records {
		if !record.Success {
			continue // 跳过失败的记录
		}

		if err := rhm.rollbackRecord(record); err != nil {
			rollbackErrors = append(rollbackErrors, 
				fmt.Sprintf("回滚 %s 失败: %v", record.DestPath, err))
		}
	}

	if len(rollbackErrors) > 0 {
		return fmt.Errorf("部分回滚失败: %v", rollbackErrors)
	}

	return nil
}

// rollbackRecord 回滚单个记录
func (rhm *RestoreHistoryManager) rollbackRecord(record RestoreRecord) error {
	// 删除恢复的文件
	if _, err := os.Stat(record.DestPath); err == nil {
		if err := os.Remove(record.DestPath); err != nil {
			return fmt.Errorf("删除恢复文件失败: %v", err)
		}
	}

	// 恢复备份文件
	if record.BackupPath != "" {
		if _, err := os.Stat(record.BackupPath); err == nil {
			if err := os.Rename(record.BackupPath, record.DestPath); err != nil {
				return fmt.Errorf("恢复备份文件失败: %v", err)
			}
		}
	}

	return nil
}

// CleanupOldSessions 清理旧会话
func (rhm *RestoreHistoryManager) CleanupOldSessions(maxAge time.Duration) error {
	rhm.mu.Lock()
	defer rhm.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	var toDelete []string

	for sessionID, session := range rhm.sessions {
		if session.StartTime.Before(cutoff) {
			toDelete = append(toDelete, sessionID)
		}
	}

	for _, sessionID := range toDelete {
		delete(rhm.sessions, sessionID)
	}

	if len(toDelete) > 0 {
		return rhm.saveHistory()
	}

	return nil
}

// GetStatistics 获取统计信息
func (rhm *RestoreHistoryManager) GetStatistics() map[string]interface{} {
	rhm.mu.RLock()
	defer rhm.mu.RUnlock()

	var totalFiles, successFiles, failedFiles int
	var totalSize int64
	var totalDuration int64

	for _, session := range rhm.sessions {
		totalFiles += session.TotalFiles
		successFiles += session.SuccessCount
		failedFiles += session.FailedCount

		for _, record := range session.Records {
			totalSize += record.FileSize
			totalDuration += record.Duration
		}
	}

	return map[string]interface{}{
		"total_sessions":    len(rhm.sessions),
		"total_files":       totalFiles,
		"success_files":     successFiles,
		"failed_files":      failedFiles,
		"success_rate":      float64(successFiles) / float64(totalFiles) * 100,
		"total_size_bytes":  totalSize,
		"avg_duration_ms":   totalDuration / int64(totalFiles),
	}
}

// ExportHistory 导出历史记录
func (rhm *RestoreHistoryManager) ExportHistory(exportPath string) error {
	rhm.mu.RLock()
	defer rhm.mu.RUnlock()

	sessions := make([]*RestoreSession, 0, len(rhm.sessions))
	for _, session := range rhm.sessions {
		sessions = append(sessions, session)
	}

	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化历史记录失败: %v", err)
	}

	return os.WriteFile(exportPath, data, 0644)
}

// loadHistory 加载历史记录
func (rhm *RestoreHistoryManager) loadHistory() error {
	if _, err := os.Stat(rhm.historyFile); os.IsNotExist(err) {
		return nil // 文件不存在，跳过加载
	}

	data, err := os.ReadFile(rhm.historyFile)
	if err != nil {
		return err
	}

	var sessions []*RestoreSession
	if err := json.Unmarshal(data, &sessions); err != nil {
		return err
	}

	for _, session := range sessions {
		rhm.sessions[session.ID] = session
	}

	return nil
}

// saveHistory 保存历史记录
func (rhm *RestoreHistoryManager) saveHistory() error {
	sessions := make([]*RestoreSession, 0, len(rhm.sessions))
	for _, session := range rhm.sessions {
		sessions = append(sessions, session)
	}

	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(rhm.historyFile, data, 0644)
}