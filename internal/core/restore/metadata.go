package restore

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// DeleteMetadata 删除元数据
type DeleteMetadata struct {
	ID              string            `json:"id"`
	OriginalPath    string            `json:"originalPath"`
	TrashPath       string            `json:"trashPath"`
	Name            string            `json:"name"`
	Size            int64             `json:"size"`
	DeletedTime     time.Time         `json:"deletedTime"`
	DeletedBy       string            `json:"deletedBy"`
	DeleteReason    string            `json:"deleteReason"`
	Type            string            `json:"type"`
	Checksum        string            `json:"checksum"`
	Permissions     string            `json:"permissions"`
	Owner           string            `json:"owner"`
	Group           string            `json:"group"`
	Attributes      map[string]string `json:"attributes"`
	Dependencies    []string          `json:"dependencies"`    // 依赖的其他文件
	RelatedFiles    []string          `json:"relatedFiles"`    // 相关文件
	BackupLocation  string            `json:"backupLocation"`  // 备份位置
	RestoreAttempts int               `json:"restoreAttempts"` // 恢复尝试次数
	LastRestoreTime time.Time         `json:"lastRestoreTime"` // 最后恢复时间
}

// MetadataManager 元数据管理器
type MetadataManager struct {
	metadataPath string
	metadata     map[string]*DeleteMetadata
}

// NewMetadataManager 创建元数据管理器
func NewMetadataManager(trashPath string) *MetadataManager {
	metadataPath := filepath.Join(trashPath, ".metadata", "deleted_files.json")
	return &MetadataManager{
		metadataPath: metadataPath,
		metadata:     make(map[string]*DeleteMetadata),
	}
}

// Load 加载元数据
func (mm *MetadataManager) Load() error {
	if _, err := os.Stat(mm.metadataPath); os.IsNotExist(err) {
		return nil // 元数据文件不存在
	}

	data, err := os.ReadFile(mm.metadataPath)
	if err != nil {
		return err
	}

	var metadataList []*DeleteMetadata
	if err := json.Unmarshal(data, &metadataList); err != nil {
		return err
	}

	// 重建索引
	mm.metadata = make(map[string]*DeleteMetadata)
	for _, meta := range metadataList {
		mm.metadata[meta.ID] = meta
	}

	return nil
}

// Save 保存元数据
func (mm *MetadataManager) Save() error {
	// 确保目录存在
	dir := filepath.Dir(mm.metadataPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 转换为列表
	var metadataList []*DeleteMetadata
	for _, meta := range mm.metadata {
		metadataList = append(metadataList, meta)
	}

	data, err := json.MarshalIndent(metadataList, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(mm.metadataPath, data, 0644)
}

// AddDeletedFile 添加删除的文件元数据
func (mm *MetadataManager) AddDeletedFile(originalPath, trashPath string, info os.FileInfo) (*DeleteMetadata, error) {
	id := mm.generateID(originalPath, info.ModTime())
	
	metadata := &DeleteMetadata{
		ID:           id,
		OriginalPath: originalPath,
		TrashPath:    trashPath,
		Name:         info.Name(),
		Size:         info.Size(),
		DeletedTime:  time.Now(),
		DeletedBy:    mm.getCurrentUser(),
		Type:         mm.getFileType(originalPath, info),
		Attributes:   make(map[string]string),
	}

	// 计算校验和
	if !info.IsDir() && info.Size() < 100*1024*1024 { // 100MB以下的文件
		checksum, err := mm.calculateChecksum(originalPath)
		if err == nil {
			metadata.Checksum = checksum
		}
	}

	// 获取文件权限
	metadata.Permissions = info.Mode().String()

	// 获取文件属性
	mm.collectFileAttributes(originalPath, info, metadata)

	// 查找相关文件
	metadata.RelatedFiles = mm.findRelatedFiles(originalPath)

	mm.metadata[id] = metadata
	return metadata, nil
}

// GetDeletedFile 获取删除的文件元数据
func (mm *MetadataManager) GetDeletedFile(id string) (*DeleteMetadata, bool) {
	meta, exists := mm.metadata[id]
	return meta, exists
}

// ListDeletedFiles 列出所有删除的文件
func (mm *MetadataManager) ListDeletedFiles() []*DeleteMetadata {
	var list []*DeleteMetadata
	for _, meta := range mm.metadata {
		list = append(list, meta)
	}
	return list
}

// RemoveDeletedFile 移除删除的文件元数据
func (mm *MetadataManager) RemoveDeletedFile(id string) {
	delete(mm.metadata, id)
}

// UpdateRestoreAttempt 更新恢复尝试
func (mm *MetadataManager) UpdateRestoreAttempt(id string) {
	if meta, exists := mm.metadata[id]; exists {
		meta.RestoreAttempts++
		meta.LastRestoreTime = time.Now()
	}
}

// SearchDeletedFiles 搜索删除的文件
func (mm *MetadataManager) SearchDeletedFiles(pattern string) []*DeleteMetadata {
	var results []*DeleteMetadata
	
	for _, meta := range mm.metadata {
		if mm.matchesPattern(meta, pattern) {
			results = append(results, meta)
		}
	}
	
	return results
}

// matchesPattern 检查是否匹配模式
func (mm *MetadataManager) matchesPattern(meta *DeleteMetadata, pattern string) bool {
	if pattern == "" {
		return true
	}

	// 文件名匹配
	matched, _ := filepath.Match(pattern, meta.Name)
	if matched {
		return true
	}

	// 路径匹配
	if filepath.Match(pattern, meta.OriginalPath) == nil {
		return true
	}

	// 类型匹配
	if meta.Type == pattern {
		return true
	}

	return false
}

// generateID 生成唯一ID
func (mm *MetadataManager) generateID(path string, modTime time.Time) string {
	data := fmt.Sprintf("%s_%d", path, modTime.Unix())
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// calculateChecksum 计算文件校验和
func (mm *MetadataManager) calculateChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// getCurrentUser 获取当前用户
func (mm *MetadataManager) getCurrentUser() string {
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	if user := os.Getenv("USERNAME"); user != "" {
		return user
	}
	return "unknown"
}

// getFileType 获取文件类型
func (mm *MetadataManager) getFileType(path string, info os.FileInfo) string {
	if info.IsDir() {
		return "directory"
	}

	ext := filepath.Ext(path)
	switch ext {
	case ".txt", ".md", ".rst", ".log":
		return "text"
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg":
		return "image"
	case ".mp4", ".avi", ".mkv", ".mov", ".wmv":
		return "video"
	case ".mp3", ".wav", ".flac", ".aac", ".ogg":
		return "audio"
	case ".pdf":
		return "pdf"
	case ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx":
		return "office"
	case ".zip", ".rar", ".7z", ".tar", ".gz":
		return "archive"
	case ".exe", ".msi", ".deb", ".rpm":
		return "executable"
	case ".go", ".py", ".js", ".java", ".cpp", ".c", ".h":
		return "code"
	default:
		return "file"
	}
}

// collectFileAttributes 收集文件属性
func (mm *MetadataManager) collectFileAttributes(path string, info os.FileInfo, metadata *DeleteMetadata) {
	// 基本属性
	metadata.Attributes["size_human"] = mm.formatSize(info.Size())
	metadata.Attributes["mod_time"] = info.ModTime().Format(time.RFC3339)
	
	// 扩展名
	if ext := filepath.Ext(path); ext != "" {
		metadata.Attributes["extension"] = ext
	}

	// 目录深度
	depth := len(filepath.SplitList(path)) - 1
	metadata.Attributes["depth"] = fmt.Sprintf("%d", depth)

	// 父目录
	metadata.Attributes["parent_dir"] = filepath.Dir(path)
}

// findRelatedFiles 查找相关文件
func (mm *MetadataManager) findRelatedFiles(path string) []string {
	var related []string
	
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	nameWithoutExt := base[:len(base)-len(filepath.Ext(base))]

	// 查找同名不同扩展名的文件
	if entries, err := os.ReadDir(dir); err == nil {
		for _, entry := range entries {
			name := entry.Name()
			if name != base && filepath.Base(name[:len(name)-len(filepath.Ext(name))]) == nameWithoutExt {
				related = append(related, filepath.Join(dir, name))
			}
		}
	}

	return related
}

// formatSize 格式化文件大小
func (mm *MetadataManager) formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GetStats 获取统计信息
func (mm *MetadataManager) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	totalFiles := len(mm.metadata)
	totalSize := int64(0)
	typeCount := make(map[string]int)
	
	for _, meta := range mm.metadata {
		totalSize += meta.Size
		typeCount[meta.Type]++
	}
	
	stats["total_files"] = totalFiles
	stats["total_size"] = totalSize
	stats["total_size_human"] = mm.formatSize(totalSize)
	stats["type_distribution"] = typeCount
	
	return stats
}

// Cleanup 清理过期的元数据
func (mm *MetadataManager) Cleanup(maxAge time.Duration) int {
	cutoff := time.Now().Add(-maxAge)
	removed := 0
	
	for id, meta := range mm.metadata {
		if meta.DeletedTime.Before(cutoff) {
			// 检查文件是否还存在于回收站
			if _, err := os.Stat(meta.TrashPath); os.IsNotExist(err) {
				delete(mm.metadata, id)
				removed++
			}
		}
	}
	
	return removed
}