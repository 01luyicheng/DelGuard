package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// TrashOperationMonitor 回收站操作监控器
type TrashOperationMonitor struct {
	config *Config
}

// isStdinInteractive 判断是否为交互式终端（避免在无TTY/管道环境中阻塞）
func isStdinInteractive() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	// 在非交互/管道场景下，Stdin 不是字符设备
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// readLineWithTimeout 从标准输入读取一行，带超时；返回(文本, 是否读取成功)
func readLineWithTimeout(timeout time.Duration) (string, bool) {
	if !isStdinInteractive() {
		return "", false
	}
	ch := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			ch <- scanner.Text()
			return
		}
		ch <- ""
	}()

	select {
	case s := <-ch:
		return s, true
	case <-time.After(timeout):
		return "", false
	}
}

// NewTrashOperationMonitor 创建回收站操作监控器
func NewTrashOperationMonitor(config *Config) *TrashOperationMonitor {
	return &TrashOperationMonitor{config: config}
}

// TrashOperation 回收站操作类型
type TrashOperation struct {
	Type        string    // 操作类型: delete_from_trash, empty_trash, delete_trash_dir
	Path        string    // 操作路径
	Timestamp   time.Time // 操作时间
	Description string    // 操作描述
	RiskLevel   string    // 风险级别: low, medium, high, critical
}

// DetectTrashOperation 检测回收站相关操作
func (m *TrashOperationMonitor) DetectTrashOperation(path string) (*TrashOperation, error) {
	cleanPath := filepath.Clean(path)

	// 获取回收站路径
	trashPaths, err := m.getSystemTrashPaths()
	if err != nil {
		return nil, err
	}

	// 检查是否为回收站相关操作
	for _, trashPath := range trashPaths {
		if m.isPathInTrash(cleanPath, trashPath) {
			return m.analyzeTrashOperation(cleanPath, trashPath)
		}
	}

	// 启发式匹配：在跨平台测试或目录不存在时，通过常见路径特征识别回收站
	if base, ok := deriveTrashRootFromPath(cleanPath); ok {
		return m.analyzeTrashOperation(cleanPath, base)
	}

	return nil, nil // 不是回收站操作
}

// getSystemTrashPaths 获取系统回收站路径
func (m *TrashOperationMonitor) getSystemTrashPaths() ([]string, error) {
	var trashPaths []string

	switch runtime.GOOS {
	case "windows":
		// Windows回收站路径
		drives := []string{"C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
		for _, drive := range drives {
			recycleBin := fmt.Sprintf("%s:\\$Recycle.Bin", drive)
			if _, err := os.Stat(recycleBin); err == nil {
				trashPaths = append(trashPaths, recycleBin)
			}

			// 老版本Windows回收站
			recycler := fmt.Sprintf("%s:\\RECYCLER", drive)
			if _, err := os.Stat(recycler); err == nil {
				trashPaths = append(trashPaths, recycler)
			}
		}

	case "darwin":
		// macOS回收站路径
		homeDir, err := os.UserHomeDir()
		if err == nil {
			trashPaths = append(trashPaths, filepath.Join(homeDir, ".Trash"))
		}

		// 系统级回收站
		trashPaths = append(trashPaths, "/.Trashes")

	case "linux":
		// Linux回收站路径
		homeDir, err := os.UserHomeDir()
		if err == nil {
			trashPaths = append(trashPaths, filepath.Join(homeDir, ".local/share/Trash"))
			trashPaths = append(trashPaths, filepath.Join(homeDir, ".Trash"))
		}

		// 系统级回收站
		trashPaths = append(trashPaths, "/tmp/.Trash-1000")
	}

	return trashPaths, nil
}

// isPathInTrash 检查路径是否在回收站内
func (m *TrashOperationMonitor) isPathInTrash(targetPath, trashPath string) bool {
	targetPath = filepath.Clean(strings.ToLower(targetPath))
	trashPath = filepath.Clean(strings.ToLower(trashPath))

	// 检查是否为回收站目录本身
	if targetPath == trashPath {
		return true
	}

	// 检查是否在回收站目录内
	return strings.HasPrefix(targetPath, trashPath+string(filepath.Separator))
}

// analyzeTrashOperation 分析回收站操作
func (m *TrashOperationMonitor) analyzeTrashOperation(targetPath, trashPath string) (*TrashOperation, error) {
	operation := &TrashOperation{
		Path:      targetPath,
		Timestamp: time.Now(),
	}

	targetPath = filepath.Clean(strings.ToLower(targetPath))
	trashPath = filepath.Clean(strings.ToLower(trashPath))

	if targetPath == trashPath {
		// 直接删除回收站目录
		operation.Type = "delete_trash_dir"
		operation.Description = "尝试删除回收站目录本身"
		operation.RiskLevel = "critical"
	} else {
		// 删除回收站内的文件
		operation.Type = "delete_from_trash"
		operation.Description = "从回收站中永久删除文件"
		operation.RiskLevel = "high"

		// 检查是否为批量删除
		if m.isBulkTrashOperation(targetPath, trashPath) {
			operation.Type = "empty_trash"
			operation.Description = "批量清空回收站"
			operation.RiskLevel = "critical"
		}
	}

	return operation, nil
}

// isBulkTrashOperation 检查是否为批量回收站操作
func (m *TrashOperationMonitor) isBulkTrashOperation(targetPath, trashPath string) bool {
	// 如果目标是回收站的主要子目录，可能是批量操作
	relativePath := strings.TrimPrefix(targetPath, trashPath+string(filepath.Separator))
	parts := strings.Split(relativePath, string(filepath.Separator))

	// 如果只有一级目录，可能是批量操作
	return len(parts) <= 2
}

// deriveTrashRootFromPath 根据常见路径特征推断回收站根目录
func deriveTrashRootFromPath(p string) (string, bool) {
	lp := filepath.Clean(strings.ToLower(p))
	// 常见特征列表（跨平台）
	candidates := []string{
		string(filepath.Separator) + "$recycle.bin",
		string(filepath.Separator) + "recycler",
		string(filepath.Separator) + ".trashes",
		string(filepath.Separator) + ".trash",
		string(filepath.Separator) + ".trash-", // linux /tmp/.Trash-1000
		string(filepath.Separator) + ".local" + string(filepath.Separator) + "share" + string(filepath.Separator) + "trash",
	}

	for _, mark := range candidates {
		idx := strings.Index(lp, mark)
		if idx >= 0 {
			// 回收站根目录 = 路径中 mark 的结束位置
			root := lp[:idx+len(mark)]
			// 规范化根目录分隔符
			return filepath.Clean(root), true
		}
	}
	return "", false
}

// WarnTrashOperation 警告回收站操作
func (m *TrashOperationMonitor) WarnTrashOperation(operation *TrashOperation, force bool) (bool, error) {
	if operation == nil {
		return true, nil // 不是回收站操作，允许继续
	}

	// 显示警告信息
	m.displayTrashWarning(operation)

	// 强制模式下仍然显示警告，但不阻止操作
	if force {
		fmt.Printf(T("⚠️ 强制模式：跳过确认，继续执行危险操作\n"))
		return true, nil
	}

	// 根据风险级别决定确认方式
	switch operation.RiskLevel {
	case "critical":
		return m.confirmCriticalTrashOperation(operation)
	case "high":
		return m.confirmHighRiskTrashOperation(operation)
	case "medium":
		return m.confirmMediumRiskTrashOperation(operation)
	default:
		return m.confirmLowRiskTrashOperation(operation)
	}
}

// displayTrashWarning 显示回收站警告
func (m *TrashOperationMonitor) displayTrashWarning(operation *TrashOperation) {
	fmt.Println(T(""))
	fmt.Println(T("🗑️ =================== 回收站操作警告 ==================="))

	switch operation.Type {
	case "delete_trash_dir":
		fmt.Println(T("❌ 危险操作：尝试删除回收站目录"))
		fmt.Println(T("📍 路径："), operation.Path)
		fmt.Println(T("⚠️  警告：这将删除整个回收站目录，导致："))
		fmt.Println(T("   • 回收站中的所有文件将被永久删除"))
		fmt.Println(T("   • 系统回收站功能可能受到影响"))
		fmt.Println(T("   • 可能需要重启系统才能恢复回收站功能"))

	case "empty_trash":
		fmt.Println(T("⚠️ 批量操作：清空回收站"))
		fmt.Println(T("📍 路径："), operation.Path)
		fmt.Println(T("⚠️  警告：这将永久删除回收站中的大量文件"))
		fmt.Println(T("   • 建议使用系统自带的清空回收站功能"))
		fmt.Println(T("   • 删除后无法恢复"))

	case "delete_from_trash":
		fmt.Println(T("ℹ️ 回收站操作：永久删除文件"))
		fmt.Println(T("📍 路径："), operation.Path)
		fmt.Println(T("⚠️  提醒：从回收站删除的文件无法恢复"))
		fmt.Println(T("   • 这是永久删除操作"))
		fmt.Println(T("   • 建议先确认文件不再需要"))
	}

	fmt.Println(T(""))
	fmt.Println(T("💡 建议："))
	fmt.Println(T("   • 如需清空回收站，推荐使用系统自带功能"))
	fmt.Println(T("   • 如需恢复文件，请使用系统的还原功能"))
	fmt.Println(T("   • 重要文件建议先备份"))
	fmt.Println(T("========================================================"))
	fmt.Println(T(""))
}

// confirmCriticalTrashOperation 确认关键回收站操作
func (m *TrashOperationMonitor) confirmCriticalTrashOperation(operation *TrashOperation) (bool, error) {
	fmt.Println(T("🚨 这是一个极其危险的操作！"))
	fmt.Println(T(""))

	// 第一次确认
	fmt.Printf(T("请输入 '%s' 以确认删除回收站: "), ConfirmDeleteRecycleBin)
	text, ok := readLineWithTimeout(30 * time.Second)
	if !ok {
		// 非交互或超时：默认不允许，防止危险操作在无确认下继续
		return false, nil
	}

	if strings.TrimSpace(text) != ConfirmDeleteRecycleBin {
		fmt.Println(T("❌ 确认失败，操作已取消"))
		return false, nil
	}

	// 第二次确认
	fmt.Println(T(""))
	fmt.Println(T("⚠️ 最后警告：此操作将永久删除回收站及其所有内容！"))
	fmt.Printf(T("请再次输入 '%s' 以最终确认: "), ConfirmYesUnderstand)

	text2, ok2 := readLineWithTimeout(30 * time.Second)
	if !ok2 {
		return false, nil
	}

	if strings.TrimSpace(text2) != ConfirmYesUnderstand {
		fmt.Println(T("❌ 最终确认失败，操作已取消"))
		return false, nil
	}

	fmt.Println(T("✅ 确认完成，将执行危险操作..."))
	return true, nil
}

// confirmHighRiskTrashOperation 确认高风险回收站操作
func (m *TrashOperationMonitor) confirmHighRiskTrashOperation(operation *TrashOperation) (bool, error) {
	fmt.Printf(T("这是一个高风险操作，请输入 '%s' 确认继续: "), ConfirmYes)
	text, ok := readLineWithTimeout(20 * time.Second)
	if !ok {
		// 非交互或超时默认不继续
		return false, nil
	}

	input := strings.TrimSpace(text)
	if input == ConfirmYes {
		fmt.Println(T("✅ 确认继续执行高风险操作"))
		return true, nil
	}

	fmt.Println(T("❌ 操作已取消"))
	return false, nil
}

// confirmMediumRiskTrashOperation 确认中风险回收站操作
func (m *TrashOperationMonitor) confirmMediumRiskTrashOperation(operation *TrashOperation) (bool, error) {
	fmt.Printf(T("确认从回收站永久删除文件? (y/N): "))
	text, ok := readLineWithTimeout(15 * time.Second)
	if !ok {
		// 非交互或超时默认不继续
		return false, nil
	}

	input := strings.ToLower(strings.TrimSpace(text))
	if input == "y" || input == "yes" {
		return true, nil
	}

	fmt.Println(T("❌ 操作已取消"))
	return false, nil
}

// confirmLowRiskTrashOperation 确认低风险回收站操作
func (m *TrashOperationMonitor) confirmLowRiskTrashOperation(operation *TrashOperation) (bool, error) {
	fmt.Printf(T("继续操作? (Y/n): "))
	text, ok := readLineWithTimeout(10 * time.Second)
	if !ok {
		// 保持低风险默认继续的策略
		return true, nil
	}

	input := strings.ToLower(strings.TrimSpace(text))
	if input == "n" || input == "no" {
		fmt.Println(T("❌ 操作已取消"))
		return false, nil
	}

	return true, nil
}

// LogTrashOperation 记录回收站操作
func (m *TrashOperationMonitor) LogTrashOperation(operation *TrashOperation, result string) {
	if operation == nil {
		return
	}

	logEntry := fmt.Sprintf("[%s] 回收站操作: %s | 路径: %s | 风险级别: %s | 结果: %s\n",
		operation.Timestamp.Format(TimeFormatStandard),
		operation.Description,
		operation.Path,
		operation.RiskLevel,
		result)

	// 写入日志文件
	logFile := filepath.Join(os.TempDir(), "delguard_trash_operations.log")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return // 静默失败
	}
	defer f.Close()

	f.WriteString(logEntry)
}

// GetTrashStatistics 获取回收站统计信息
func (m *TrashOperationMonitor) GetTrashStatistics(trashPath string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	if _, err := os.Stat(trashPath); os.IsNotExist(err) {
		stats["exists"] = false
		return stats, nil
	}

	stats["exists"] = true

	// 统计文件数量和总大小
	var fileCount int64
	var totalSize int64

	err := filepath.Walk(trashPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续统计
		}

		if !info.IsDir() {
			fileCount++
			totalSize += info.Size()
		}

		return nil
	})

	stats["file_count"] = fileCount
	stats["total_size"] = totalSize
	stats["size_mb"] = float64(totalSize) / (1024 * 1024)

	return stats, err
}
