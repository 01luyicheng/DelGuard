package restore

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// RestoreOptions 恢复选项
type RestoreOptions struct {
	VerifyIntegrity   bool   // 验证文件完整性
	CreateBackup      bool   // 创建备份
	OverwriteExisting bool   // 覆盖已存在文件
	PreserveMetadata  bool   // 保留元数据
	MaxConcurrency    int    // 最大并发数
	ChunkSize         int64  // 分块大小（用于大文件）
	EnableResume      bool   // 启用断点续传
}

// RestoreResult 恢复结果
type RestoreResult struct {
	FilePath      string
	Success       bool
	Error         error
	BytesRestored int64
	Duration      time.Duration
	Checksum      string
}

// RestoreProgress 恢复进度
type RestoreProgress struct {
	TotalFiles     int
	CompletedFiles int
	TotalBytes     int64
	RestoredBytes  int64
	CurrentFile    string
	StartTime      time.Time
}

// EnhancedRestoreService 增强恢复服务
type EnhancedRestoreService struct {
	options    RestoreOptions
	progress   *RestoreProgress
	mu         sync.RWMutex
	cancelFunc context.CancelFunc
}

// NewEnhancedRestoreService 创建增强恢复服务
func NewEnhancedRestoreService(options RestoreOptions) *EnhancedRestoreService {
	if options.MaxConcurrency <= 0 {
		options.MaxConcurrency = 4
	}
	if options.ChunkSize <= 0 {
		options.ChunkSize = 1024 * 1024 // 1MB
	}

	return &EnhancedRestoreService{
		options: options,
		progress: &RestoreProgress{
			StartTime: time.Now(),
		},
	}
}

// RestoreFiles 恢复多个文件
func (ers *EnhancedRestoreService) RestoreFiles(ctx context.Context, files []*RecoverableFile) ([]RestoreResult, error) {
	ctx, cancel := context.WithCancel(ctx)
	ers.cancelFunc = cancel
	defer cancel()

	// 初始化进度
	ers.initProgress(files)

	// 创建结果通道
	resultChan := make(chan RestoreResult, len(files))
	
	// 创建工作池
	semaphore := make(chan struct{}, ers.options.MaxConcurrency)
	var wg sync.WaitGroup

	// 启动恢复任务
	for _, file := range files {
		wg.Add(1)
		go func(f *RecoverableFile) {
			defer wg.Done()
			
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
				result := ers.restoreFileWithIntegrity(ctx, f)
				resultChan <- result
			case <-ctx.Done():
				resultChan <- RestoreResult{
					FilePath: f.OriginalPath,
					Success:  false,
					Error:    ctx.Err(),
				}
			}
		}(file)
	}

	// 等待所有任务完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	var results []RestoreResult
	for result := range resultChan {
		results = append(results, result)
		ers.updateProgress(result)
	}

	return results, nil
}

// restoreFileWithIntegrity 带完整性验证的文件恢复
func (ers *EnhancedRestoreService) restoreFileWithIntegrity(ctx context.Context, file *RecoverableFile) RestoreResult {
	startTime := time.Now()
	
	result := RestoreResult{
		FilePath: file.OriginalPath,
	}

	// 更新当前处理文件
	ers.setCurrentFile(file.OriginalPath)

	// 验证源文件完整性
	if ers.options.VerifyIntegrity {
		checksum, err := ers.calculateChecksum(file.CurrentPath)
		if err != nil {
			result.Error = fmt.Errorf("计算源文件校验和失败: %v", err)
			return result
		}
		result.Checksum = checksum
	}

	// 创建目标目录
	destDir := filepath.Dir(file.OriginalPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		result.Error = fmt.Errorf("创建目标目录失败: %v", err)
		return result
	}

	// 处理已存在的文件
	if err := ers.handleExistingFile(file.OriginalPath); err != nil {
		result.Error = err
		return result
	}

	// 执行文件恢复
	if file.Size > ers.options.ChunkSize && ers.options.EnableResume {
		// 大文件使用分块恢复
		err := ers.restoreFileInChunks(ctx, file, &result)
		if err != nil {
			result.Error = err
			return result
		}
	} else {
		// 小文件直接复制
		err := ers.copyFile(file.CurrentPath, file.OriginalPath)
		if err != nil {
			result.Error = err
			return result
		}
		result.BytesRestored = file.Size
	}

	// 保留元数据
	if ers.options.PreserveMetadata {
		if err := ers.preserveMetadata(file.CurrentPath, file.OriginalPath); err != nil {
			// 元数据恢复失败不影响文件恢复成功
			fmt.Printf("警告: 保留元数据失败: %v\n", err)
		}
	}

	// 验证恢复后文件完整性
	if ers.options.VerifyIntegrity {
		restoredChecksum, err := ers.calculateChecksum(file.OriginalPath)
		if err != nil {
			result.Error = fmt.Errorf("验证恢复文件校验和失败: %v", err)
			return result
		}
		
		if restoredChecksum != result.Checksum {
			result.Error = fmt.Errorf("文件完整性验证失败: 校验和不匹配")
			return result
		}
	}

	result.Success = true
	result.Duration = time.Since(startTime)
	return result
}

// handleExistingFile 处理已存在的文件
func (ers *EnhancedRestoreService) handleExistingFile(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // 文件不存在，无需处理
	}

	if !ers.options.OverwriteExisting {
		return fmt.Errorf("目标文件已存在且未启用覆盖选项: %s", filePath)
	}

	if ers.options.CreateBackup {
		backupPath := filePath + ".backup." + time.Now().Format("20060102150405")
		if err := os.Rename(filePath, backupPath); err != nil {
			return fmt.Errorf("创建备份文件失败: %v", err)
		}
		fmt.Printf("已创建备份: %s\n", backupPath)
	} else {
		if err := os.Remove(filePath); err != nil {
			return fmt.Errorf("删除现有文件失败: %v", err)
		}
	}

	return nil
}

// restoreFileInChunks 分块恢复大文件
func (ers *EnhancedRestoreService) restoreFileInChunks(ctx context.Context, file *RecoverableFile, result *RestoreResult) error {
	srcFile, err := os.Open(file.CurrentPath)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %v", err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(file.OriginalPath)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %v", err)
	}
	defer destFile.Close()

	buffer := make([]byte, ers.options.ChunkSize)
	var totalCopied int64

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, err := srcFile.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("读取源文件失败: %v", err)
		}

		if n == 0 {
			break
		}

		if _, err := destFile.Write(buffer[:n]); err != nil {
			return fmt.Errorf("写入目标文件失败: %v", err)
		}

		totalCopied += int64(n)
		result.BytesRestored = totalCopied

		// 更新进度
		ers.addRestoredBytes(int64(n))
	}

	return nil
}

// copyFile 复制文件
func (ers *EnhancedRestoreService) copyFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}

// calculateChecksum 计算文件校验和
func (ers *EnhancedRestoreService) calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// preserveMetadata 保留文件元数据
func (ers *EnhancedRestoreService) preserveMetadata(src, dest string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// 设置文件权限
	if err := os.Chmod(dest, srcInfo.Mode()); err != nil {
		return err
	}

	// 设置修改时间
	return os.Chtimes(dest, srcInfo.ModTime(), srcInfo.ModTime())
}

// initProgress 初始化进度
func (ers *EnhancedRestoreService) initProgress(files []*RecoverableFile) {
	ers.mu.Lock()
	defer ers.mu.Unlock()

	ers.progress.TotalFiles = len(files)
	ers.progress.CompletedFiles = 0
	ers.progress.RestoredBytes = 0
	ers.progress.StartTime = time.Now()

	var totalBytes int64
	for _, file := range files {
		totalBytes += file.Size
	}
	ers.progress.TotalBytes = totalBytes
}

// updateProgress 更新进度
func (ers *EnhancedRestoreService) updateProgress(result RestoreResult) {
	ers.mu.Lock()
	defer ers.mu.Unlock()

	ers.progress.CompletedFiles++
}

// setCurrentFile 设置当前处理文件
func (ers *EnhancedRestoreService) setCurrentFile(filePath string) {
	ers.mu.Lock()
	defer ers.mu.Unlock()
	ers.progress.CurrentFile = filePath
}

// addRestoredBytes 添加已恢复字节数
func (ers *EnhancedRestoreService) addRestoredBytes(bytes int64) {
	ers.mu.Lock()
	defer ers.mu.Unlock()
	ers.progress.RestoredBytes += bytes
}

// GetProgress 获取恢复进度
func (ers *EnhancedRestoreService) GetProgress() RestoreProgress {
	ers.mu.RLock()
	defer ers.mu.RUnlock()
	return *ers.progress
}

// Cancel 取消恢复操作
func (ers *EnhancedRestoreService) Cancel() {
	if ers.cancelFunc != nil {
		ers.cancelFunc()
	}
}