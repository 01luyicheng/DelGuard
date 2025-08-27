package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// SmartPromptSystem 智能提示系统
type SmartPromptSystem struct {
	ui     *EnhancedInterface
	locale string
}

// NewSmartPromptSystem 创建智能提示系统
func NewSmartPromptSystem(ui *EnhancedInterface, locale string) *SmartPromptSystem {
	return &SmartPromptSystem{
		ui:     ui,
		locale: locale,
	}
}

// PromptContext 提示上下文
type PromptContext struct {
	Operation    string
	FilePath     string
	FileType     string
	FileSize     int64
	IsDirectory  bool
	IsSystemFile bool
	IsReadOnly   bool
	RiskLevel    string
}

// SmartDeletePrompt 智能删除提示
func (sps *SmartPromptSystem) SmartDeletePrompt(ctx PromptContext) (bool, error) {
	dialog := ConfirmationDialog{
		Title:     sps.getDeleteTitle(ctx),
		Message:   sps.getDeleteMessage(ctx),
		Details:   sps.getDeleteDetails(ctx),
		ShowRisks: true,
		RiskLevel: ctx.RiskLevel,
		Timeout:   30 * time.Second,
	}

	// 根据风险级别设置默认选项
	dialog.DefaultYes = ctx.RiskLevel == "low"

	return sps.ui.ShowConfirmationDialog(dialog)
}

// SmartBatchPrompt 智能批量操作提示
func (sps *SmartPromptSystem) SmartBatchPrompt(operation string, files []string, totalSize int64) (bool, error) {
	riskLevel := sps.assessBatchRisk(files, totalSize)

	dialog := ConfirmationDialog{
		Title:     fmt.Sprintf("批量%s确认", operation),
		Message:   fmt.Sprintf("即将%s %d 个文件/目录", operation, len(files)),
		Details:   sps.getBatchDetails(files, totalSize),
		ShowRisks: true,
		RiskLevel: riskLevel,
		Timeout:   60 * time.Second,
	}

	return sps.ui.ShowConfirmationDialog(dialog)
}

// SmartErrorPrompt 智能错误提示
func (sps *SmartPromptSystem) SmartErrorPrompt(err error, context string) {
	errorType := sps.classifyError(err)
	suggestions := sps.getErrorSuggestions(errorType, context)

	sps.ui.ShowNotification(NotificationError, "操作失败", err.Error())

	if len(suggestions) > 0 {
		fmt.Println("🔧 建议解决方案:")
		for i, suggestion := range suggestions {
			fmt.Printf("   %d. %s\n", i+1, suggestion)
		}
		fmt.Println()
	}
}

// SmartSuccessPrompt 智能成功提示
func (sps *SmartPromptSystem) SmartSuccessPrompt(operation string, details map[string]interface{}) {
	message := sps.getSuccessMessage(operation, details)
	nextSteps := sps.getNextSteps(operation)

	sps.ui.ShowNotification(NotificationSuccess, "操作成功", message)

	if len(nextSteps) > 0 {
		fmt.Println("📋 后续操作建议:")
		for i, step := range nextSteps {
			fmt.Printf("   %d. %s\n", i+1, step)
		}
		fmt.Println()
	}
}

// SmartWarningPrompt 智能警告提示
func (sps *SmartPromptSystem) SmartWarningPrompt(warning, context string) bool {
	dialog := ConfirmationDialog{
		Title:     "⚠️ 警告",
		Message:   warning,
		Details:   sps.getWarningDetails(context),
		ShowRisks: true,
		RiskLevel: "medium",
		Timeout:   30 * time.Second,
	}

	confirmed, _ := sps.ui.ShowConfirmationDialog(dialog)
	return confirmed
}

// 私有方法实现

func (sps *SmartPromptSystem) getDeleteTitle(ctx PromptContext) string {
	if ctx.IsDirectory {
		return "🗂️ 删除目录确认"
	}
	return "🗑️ 删除文件确认"
}

func (sps *SmartPromptSystem) getDeleteMessage(ctx PromptContext) string {
	fileName := filepath.Base(ctx.FilePath)

	if ctx.IsDirectory {
		return fmt.Sprintf("即将删除目录: %s", fileName)
	}

	sizeStr := sps.formatFileSize(ctx.FileSize)
	return fmt.Sprintf("即将删除文件: %s (%s)", fileName, sizeStr)
}

func (sps *SmartPromptSystem) getDeleteDetails(ctx PromptContext) []string {
	details := []string{
		fmt.Sprintf("完整路径: %s", ctx.FilePath),
	}

	if ctx.FileType != "" {
		details = append(details, fmt.Sprintf("文件类型: %s", ctx.FileType))
	}

	if ctx.IsReadOnly {
		details = append(details, "⚠️ 只读文件")
	}

	if ctx.IsSystemFile {
		details = append(details, "🔒 系统文件")
	}

	if ctx.IsDirectory {
		details = append(details, "📁 目录（包含子文件）")
	}

	return details
}

func (sps *SmartPromptSystem) getBatchDetails(files []string, totalSize int64) []string {
	details := []string{
		fmt.Sprintf("总大小: %s", sps.formatFileSize(totalSize)),
	}

	// 分析文件类型
	typeCount := make(map[string]int)
	dirCount := 0

	for _, file := range files {
		if sps.isDirectory(file) {
			dirCount++
		} else {
			ext := strings.ToLower(filepath.Ext(file))
			typeCount[ext]++
		}
	}

	if dirCount > 0 {
		details = append(details, fmt.Sprintf("目录: %d 个", dirCount))
	}

	// 显示主要文件类型
	for ext, count := range typeCount {
		if count > 0 {
			if ext == "" {
				details = append(details, fmt.Sprintf("无扩展名文件: %d 个", count))
			} else {
				details = append(details, fmt.Sprintf("%s 文件: %d 个", ext, count))
			}
		}
	}

	return details
}

func (sps *SmartPromptSystem) assessBatchRisk(files []string, totalSize int64) string {
	// 大文件或大量文件认为是高风险
	if totalSize > 10*1024*1024*1024 { // 10GB
		return "high"
	}

	if len(files) > 1000 {
		return "high"
	}

	// 检查是否包含系统文件
	for _, file := range files {
		if sps.isSystemPath(file) {
			return "critical"
		}
	}

	if len(files) > 100 || totalSize > 1024*1024*1024 { // 1GB
		return "medium"
	}

	return "low"
}

func (sps *SmartPromptSystem) classifyError(err error) string {
	errStr := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errStr, "permission") || strings.Contains(errStr, "access"):
		return "permission"
	case strings.Contains(errStr, "not found") || strings.Contains(errStr, "no such"):
		return "not_found"
	case strings.Contains(errStr, "in use") || strings.Contains(errStr, "busy"):
		return "file_in_use"
	case strings.Contains(errStr, "space") || strings.Contains(errStr, "disk"):
		return "disk_space"
	case strings.Contains(errStr, "network") || strings.Contains(errStr, "connection"):
		return "network"
	case strings.Contains(errStr, "invalid") || strings.Contains(errStr, "illegal"):
		return "invalid_path"
	default:
		return "unknown"
	}
}

func (sps *SmartPromptSystem) getErrorSuggestions(errorType, context string) []string {
	switch errorType {
	case "permission":
		return []string{
			"以管理员身份运行程序",
			"检查文件或目录的权限设置",
			"确认文件没有被其他程序占用",
			"尝试关闭可能占用文件的程序",
		}
	case "not_found":
		return []string{
			"检查文件路径是否正确",
			"确认文件是否已被移动或删除",
			"使用智能搜索功能查找相似文件",
			"检查文件名是否包含特殊字符",
		}
	case "file_in_use":
		return []string{
			"关闭正在使用该文件的程序",
			"等待文件操作完成后重试",
			"检查是否有后台进程占用文件",
			"重启计算机后重试",
		}
	case "disk_space":
		return []string{
			"清理磁盘空间",
			"删除临时文件和回收站内容",
			"移动文件到其他磁盘",
			"检查磁盘健康状态",
		}
	case "network":
		return []string{
			"检查网络连接",
			"确认网络路径可访问",
			"尝试重新连接网络驱动器",
			"检查网络权限设置",
		}
	case "invalid_path":
		return []string{
			"检查路径中是否包含非法字符",
			"确认路径长度不超过系统限制",
			"使用引号包围包含空格的路径",
			"避免使用特殊字符如 < > | \" : * ? \\ /",
		}
	default:
		return []string{
			"检查系统日志获取更多信息",
			"尝试重启程序",
			"联系技术支持",
		}
	}
}

func (sps *SmartPromptSystem) getSuccessMessage(operation string, details map[string]interface{}) string {
	switch operation {
	case "delete":
		if count, ok := details["count"].(int); ok && count > 1 {
			return fmt.Sprintf("成功删除 %d 个文件/目录", count)
		}
		return "文件删除成功"
	case "restore":
		if count, ok := details["count"].(int); ok && count > 1 {
			return fmt.Sprintf("成功恢复 %d 个文件", count)
		}
		return "文件恢复成功"
	case "search":
		if count, ok := details["count"].(int); ok {
			return fmt.Sprintf("搜索完成，找到 %d 个匹配项", count)
		}
		return "搜索完成"
	default:
		return "操作完成"
	}
}

func (sps *SmartPromptSystem) getNextSteps(operation string) []string {
	switch operation {
	case "delete":
		return []string{
			"可以使用 restore 命令恢复已删除的文件",
			"定期清理回收站以释放磁盘空间",
			"使用 --dry-run 参数预览删除操作",
		}
	case "restore":
		return []string{
			"检查恢复的文件是否完整",
			"更新相关程序的文件路径",
			"考虑备份重要文件",
		}
	case "search":
		return []string{
			"使用搜索结果进行批量操作",
			"保存搜索条件以便重复使用",
			"使用更精确的搜索条件缩小范围",
		}
	default:
		return []string{}
	}
}

func (sps *SmartPromptSystem) getWarningDetails(context string) []string {
	// 根据上下文返回相关的警告详情
	return []string{
		"此操作可能产生不可预期的后果",
		"建议在继续之前备份重要数据",
		"如不确定，请选择取消操作",
	}
}

func (sps *SmartPromptSystem) formatFileSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case size >= TB:
		return fmt.Sprintf("%.2f TB", float64(size)/TB)
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d 字节", size)
	}
}

func (sps *SmartPromptSystem) isDirectory(path string) bool {
	// 这里应该调用实际的文件系统检查
	// 为了示例，简单检查路径特征
	return strings.HasSuffix(path, "/") || strings.HasSuffix(path, "\\")
}

func (sps *SmartPromptSystem) isSystemPath(path string) bool {
	systemPaths := []string{
		"C:\\Windows",
		"C:\\Program Files",
		"C:\\System32",
		"/bin",
		"/sbin",
		"/usr/bin",
		"/usr/sbin",
		"/etc",
		"/sys",
		"/proc",
	}

	path = strings.ToLower(path)
	for _, sysPath := range systemPaths {
		if strings.HasPrefix(path, strings.ToLower(sysPath)) {
			return true
		}
	}

	return false
}
