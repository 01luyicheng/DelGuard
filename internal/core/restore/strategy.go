package restore

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RestoreStrategy 恢复策略接口
type RestoreStrategy interface {
	CanRestore(ctx context.Context, meta *DeleteMetadata, options *RestoreOptions) (bool, string)
	PrepareRestore(ctx context.Context, meta *DeleteMetadata, options *RestoreOptions) (*RestorePlan, error)
	ExecuteRestore(ctx context.Context, plan *RestorePlan) *RestoreResult
	PostRestore(ctx context.Context, result *RestoreResult) error
}

// RestorePlan 恢复计划
type RestorePlan struct {
	Metadata      *DeleteMetadata
	SourcePath    string
	TargetPath    string
	BackupPath    string
	Strategy      string
	Steps         []RestoreStep
	EstimatedTime time.Duration
	RequiredSpace int64
	Conflicts     []string
	Dependencies  []string
}

// RestoreStep 恢复步骤
type RestoreStep struct {
	Type        string
	Description string
	Source      string
	Target      string
	Required    bool
	Completed   bool
	Error       string
}

// DefaultRestoreStrategy 默认恢复策略
type DefaultRestoreStrategy struct {
	metadataManager *MetadataManager
}

// NewDefaultRestoreStrategy 创建默认恢复策略
func NewDefaultRestoreStrategy(mm *MetadataManager) *DefaultRestoreStrategy {
	return &DefaultRestoreStrategy{
		metadataManager: mm,
	}
}

// CanRestore 检查是否可以恢复
func (s *DefaultRestoreStrategy) CanRestore(ctx context.Context, meta *DeleteMetadata, options *RestoreOptions) (bool, string) {
	// 检查源文件是否存在
	if _, err := os.Stat(meta.TrashPath); os.IsNotExist(err) {
		return false, "源文件不存在于回收站"
	}

	// 检查目标路径
	targetPath := meta.OriginalPath
	if options.TargetDirectory != "" {
		targetPath = filepath.Join(options.TargetDirectory, meta.Name)
	}

	// 检查目标文件是否已存在
	if _, err := os.Stat(targetPath); err == nil {
		if !options.OverwriteExisting {
			return false, "目标文件已存在"
		}
	}

	// 检查目标目录是否可写
	targetDir := filepath.Dir(targetPath)
	if err := s.checkDirectoryWritable(targetDir); err != nil {
		return false, fmt.Sprintf("目标目录不可写: %v", err)
	}

	// 检查磁盘空间
	if err := s.checkDiskSpace(targetDir, meta.Size); err != nil {
		return false, fmt.Sprintf("磁盘空间不足: %v", err)
	}

	return true, ""
}

// PrepareRestore 准备恢复
func (s *DefaultRestoreStrategy) PrepareRestore(ctx context.Context, meta *DeleteMetadata, options *RestoreOptions) (*RestorePlan, error) {
	plan := &RestorePlan{
		Metadata:   meta,
		SourcePath: meta.TrashPath,
		Strategy:   "default",
		Steps:      []RestoreStep{},
	}

	// 确定目标路径
	if options.TargetDirectory != "" {
		plan.TargetPath = filepath.Join(options.TargetDirectory, meta.Name)
	} else {
		plan.TargetPath = meta.OriginalPath
	}

	// 检查冲突
	conflicts := s.checkConflicts(plan.TargetPath, meta)
	plan.Conflicts = conflicts

	// 添加恢复步骤
	s.addRestoreSteps(plan, options)

	// 估算时间和空间
	plan.EstimatedTime = s.estimateRestoreTime(meta)
	plan.RequiredSpace = meta.Size

	return plan, nil
}

// ExecuteRestore 执行恢复
func (s *DefaultRestoreStrategy) ExecuteRestore(ctx context.Context, plan *RestorePlan) *RestoreResult {
	result := &RestoreResult{
		OriginalPath: plan.Metadata.OriginalPath,
		RestoredPath: plan.TargetPath,
	}

	// 执行每个步骤
	for i := range plan.Steps {
		step := &plan.Steps[i]

		// 检查上下文是否被取消
		select {
		case <-ctx.Done():
			result.Error = "恢复被取消"
			return result
		default:
		}

		if err := s.executeStep(ctx, step, plan); err != nil {
			step.Error = err.Error()
			if step.Required {
				result.Error = fmt.Sprintf("步骤 '%s' 失败: %v", step.Description, err)
				return result
			}
		} else {
			step.Completed = true
		}
	}

	result.Success = true
	return result
}

// PostRestore 恢复后处理
func (s *DefaultRestoreStrategy) PostRestore(ctx context.Context, result *RestoreResult) error {
	if !result.Success {
		return nil
	}

	// 更新元数据
	if meta, exists := s.metadataManager.GetDeletedFile(result.OriginalPath); exists {
		s.metadataManager.UpdateRestoreAttempt(meta.ID)

		// 如果恢复成功，从元数据中移除
		s.metadataManager.RemoveDeletedFile(meta.ID)
	}

	return s.metadataManager.Save()
}

// addRestoreSteps 添加恢复步骤
func (s *DefaultRestoreStrategy) addRestoreSteps(plan *RestorePlan, options *RestoreOptions) {
	// 1. 创建目标目录
	targetDir := filepath.Dir(plan.TargetPath)
	plan.Steps = append(plan.Steps, RestoreStep{
		Type:        "create_directory",
		Description: "创建目标目录",
		Target:      targetDir,
		Required:    true,
	})

	// 2. 备份现有文件（如果需要）
	if options.CreateBackup {
		if _, err := os.Stat(plan.TargetPath); err == nil {
			backupPath := plan.TargetPath + ".backup." + time.Now().Format("20060102150405")
			plan.BackupPath = backupPath
			plan.Steps = append(plan.Steps, RestoreStep{
				Type:        "create_backup",
				Description: "备份现有文件",
				Source:      plan.TargetPath,
				Target:      backupPath,
				Required:    false,
			})
		}
	}

	// 3. 移动文件
	plan.Steps = append(plan.Steps, RestoreStep{
		Type:        "move_file",
		Description: "恢复文件",
		Source:      plan.SourcePath,
		Target:      plan.TargetPath,
		Required:    true,
	})

	// 4. 验证完整性（如果需要）
	if options.VerifyIntegrity {
		plan.Steps = append(plan.Steps, RestoreStep{
			Type:        "verify_integrity",
			Description: "验证文件完整性",
			Target:      plan.TargetPath,
			Required:    false,
		})
	}

	// 5. 恢复相关文件
	for _, relatedFile := range plan.Metadata.RelatedFiles {
		if s.shouldRestoreRelatedFile(relatedFile) {
			plan.Steps = append(plan.Steps, RestoreStep{
				Type:        "restore_related",
				Description: fmt.Sprintf("恢复相关文件: %s", filepath.Base(relatedFile)),
				Source:      relatedFile,
				Required:    false,
			})
		}
	}
}

// executeStep 执行单个步骤
func (s *DefaultRestoreStrategy) executeStep(ctx context.Context, step *RestoreStep, plan *RestorePlan) error {
	switch step.Type {
	case "create_directory":
		return os.MkdirAll(step.Target, 0755)

	case "create_backup":
		return os.Rename(step.Source, step.Target)

	case "move_file":
		return os.Rename(step.Source, step.Target)

	case "verify_integrity":
		return s.verifyFileIntegrity(step.Target, plan.Metadata)

	case "restore_related":
		// 恢复相关文件的逻辑
		return s.restoreRelatedFile(step.Source, plan)

	default:
		return fmt.Errorf("未知的步骤类型: %s", step.Type)
	}
}

// checkConflicts 检查冲突
func (s *DefaultRestoreStrategy) checkConflicts(targetPath string, meta *DeleteMetadata) []string {
	var conflicts []string

	// 检查目标文件是否存在
	if info, err := os.Stat(targetPath); err == nil {
		if info.Size() != meta.Size {
			conflicts = append(conflicts, "文件大小不同")
		}
		if info.ModTime().After(meta.DeletedTime) {
			conflicts = append(conflicts, "目标文件更新")
		}
	}

	// 检查目录权限
	targetDir := filepath.Dir(targetPath)
	if err := s.checkDirectoryWritable(targetDir); err != nil {
		conflicts = append(conflicts, "目录不可写")
	}

	return conflicts
}

// checkDirectoryWritable 检查目录是否可写
func (s *DefaultRestoreStrategy) checkDirectoryWritable(dir string) error {
	// 确保目录存在
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 尝试创建临时文件
	tempFile := filepath.Join(dir, ".delguard_write_test")
	file, err := os.Create(tempFile)
	if err != nil {
		return err
	}
	file.Close()
	os.Remove(tempFile)

	return nil
}

// checkDiskSpace 检查磁盘空间
func (s *DefaultRestoreStrategy) checkDiskSpace(dir string, requiredSize int64) error {
	// 简化实现：假设有足够空间
	// 在实际实现中，应该检查磁盘可用空间
	return nil
}

// estimateRestoreTime 估算恢复时间
func (s *DefaultRestoreStrategy) estimateRestoreTime(meta *DeleteMetadata) time.Duration {
	// 基于文件大小估算时间
	// 假设传输速度为 100MB/s
	const transferSpeed = 100 * 1024 * 1024 // 100MB/s

	seconds := float64(meta.Size) / float64(transferSpeed)
	if seconds < 1 {
		seconds = 1
	}

	return time.Duration(seconds) * time.Second
}

// verifyFileIntegrity 验证文件完整性
func (s *DefaultRestoreStrategy) verifyFileIntegrity(path string, meta *DeleteMetadata) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	// 检查文件大小
	if info.Size() != meta.Size {
		return fmt.Errorf("文件大小不匹配: 期望 %d，实际 %d", meta.Size, info.Size())
	}

	// 检查校验和（如果有）
	if meta.Checksum != "" {
		mm := &MetadataManager{}
		checksum, err := mm.calculateChecksum(path)
		if err != nil {
			return fmt.Errorf("计算校验和失败: %v", err)
		}
		if checksum != meta.Checksum {
			return fmt.Errorf("校验和不匹配")
		}
	}

	return nil
}

// shouldRestoreRelatedFile 检查是否应该恢复相关文件
func (s *DefaultRestoreStrategy) shouldRestoreRelatedFile(relatedFile string) bool {
	// 检查相关文件是否在回收站中
	// 这里简化处理，实际应该检查元数据
	return false
}

// restoreRelatedFile 恢复相关文件
func (s *DefaultRestoreStrategy) restoreRelatedFile(source string, plan *RestorePlan) error {
	// 恢复相关文件的逻辑
	// 这里简化处理
	return nil
}

// SmartRestoreStrategy 智能恢复策略
type SmartRestoreStrategy struct {
	*DefaultRestoreStrategy
}

// NewSmartRestoreStrategy 创建智能恢复策略
func NewSmartRestoreStrategy(mm *MetadataManager) *SmartRestoreStrategy {
	return &SmartRestoreStrategy{
		DefaultRestoreStrategy: NewDefaultRestoreStrategy(mm),
	}
}

// CanRestore 智能检查是否可以恢复
func (s *SmartRestoreStrategy) CanRestore(ctx context.Context, meta *DeleteMetadata, options *RestoreOptions) (bool, string) {
	// 首先执行基本检查
	canRestore, reason := s.DefaultRestoreStrategy.CanRestore(ctx, meta, options)
	if !canRestore {
		return false, reason
	}

	// 智能检查：分析恢复风险
	risk := s.analyzeRestoreRisk(meta, options)
	if risk > 0.8 {
		return false, "恢复风险过高"
	}

	// 检查依赖关系
	if len(meta.Dependencies) > 0 {
		missing := s.checkMissingDependencies(meta.Dependencies)
		if len(missing) > 0 {
			return false, fmt.Sprintf("缺少依赖文件: %s", strings.Join(missing, ", "))
		}
	}

	return true, ""
}

// analyzeRestoreRisk 分析恢复风险
func (s *SmartRestoreStrategy) analyzeRestoreRisk(meta *DeleteMetadata, options *RestoreOptions) float64 {
	risk := 0.0

	// 文件年龄风险
	age := time.Since(meta.DeletedTime)
	if age > 30*24*time.Hour { // 超过30天
		risk += 0.3
	}

	// 恢复尝试次数风险
	if meta.RestoreAttempts > 3 {
		risk += 0.2
	}

	// 文件大小风险
	if meta.Size > 1024*1024*1024 { // 超过1GB
		risk += 0.1
	}

	// 系统文件风险
	if s.isSystemFile(meta.OriginalPath) {
		risk += 0.4
	}

	return risk
}

// checkMissingDependencies 检查缺失的依赖
func (s *SmartRestoreStrategy) checkMissingDependencies(dependencies []string) []string {
	var missing []string

	for _, dep := range dependencies {
		if _, err := os.Stat(dep); os.IsNotExist(err) {
			missing = append(missing, dep)
		}
	}

	return missing
}

// isSystemFile 检查是否为系统文件
func (s *SmartRestoreStrategy) isSystemFile(path string) bool {
	systemPaths := []string{
		"C:\\Windows",
		"C:\\Program Files",
		"C:\\Program Files (x86)",
		"/System",
		"/usr/bin",
		"/usr/lib",
		"/etc",
	}

	for _, sysPath := range systemPaths {
		if strings.HasPrefix(strings.ToLower(path), strings.ToLower(sysPath)) {
			return true
		}
	}

	return false
}
