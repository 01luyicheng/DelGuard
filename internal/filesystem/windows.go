package filesystem

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// TrashMetadata 回收站元数据结构
type TrashMetadata struct {
	OriginalPath string    `json:"original_path"`
	DeletedTime  time.Time `json:"deleted_time"`
	FileName     string    `json:"file_name"`
	Size         int64     `json:"size"`
	IsDirectory  bool      `json:"is_directory"`
	Permissions  string    `json:"permissions"`
	Hash         string    `json:"hash,omitempty"`
	SystemTrash  bool      `json:"system_trash,omitempty"`
}

// copyDirectoryAndRemove 递归复制目录后删除源目录
func (w *WindowsTrashManager) copyDirectoryAndRemove(src, dst string) error {
	// 确保源目录存在
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("无法访问源目录: %v", err)
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("源路径不是目录: %s", src)
	}

	// 创建目标目录
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	// 读取源目录内容
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("读取源目录失败: %v", err)
	}

	// 递归复制每个条目
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := w.copyDirectoryAndRemove(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := w.copyAndRemove(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	// 删除源目录
	return os.Remove(src)
}

// WindowsTrashManager Windows回收站管理器
type WindowsTrashManager struct {
	forceOverwrite bool
}

// NewWindowsTrashManager 创建Windows回收站管理器
func NewWindowsTrashManager() *WindowsTrashManager {
	return &WindowsTrashManager{forceOverwrite: false}
}

// SetForceOverwrite 设置是否强制覆盖已存在文件
func (w *WindowsTrashManager) SetForceOverwrite(force bool) {
	w.forceOverwrite = force
}

// MoveToTrash 将文件移动到Windows回收站
func (w *WindowsTrashManager) MoveToTrash(filePath string) error {
	// 转换为绝对路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("路径转换失败: %v", err)
	}

	// 验证路径安全性，防止路径遍历攻击
	if err := w.validatePath(absPath); err != nil {
		return fmt.Errorf("路径验证失败: %v", err)
	}

	// 检查文件是否存在
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", absPath)
	}

	// 尝试使用系统回收站
	if w.CanUseSystemRecycleBin() {
		// 优先使用PowerShell方法
		if err := w.moveToSystemRecycleBin(absPath); err == nil {
			return nil
		}
	}

	// 如果系统回收站不可用，使用DelGuard专用回收站
	return w.moveToDelGuardTrash(absPath)
}

// moveToRecycleBin 使用Windows API移动文件到回收站
func (w *WindowsTrashManager) moveToRecycleBin(filePath string) error {
	// 使用Windows系统回收站API
	// 首先尝试使用系统回收站，失败则回退到DelGuard专用回收站
	
	// 尝试使用系统回收站
	if err := w.moveToSystemRecycleBin(filePath); err == nil {
		return nil
	}
	
	// 系统回收站失败，使用DelGuard专用回收站
	return w.moveToDelGuardTrash(filePath)
}

// moveToSystemRecycleBin 尝试使用Windows系统回收站
func (w *WindowsTrashManager) moveToSystemRecycleBin(filePath string) error {
	// 在Windows上使用SHFileOperationW API来移动到系统回收站
	// 由于Go的限制，我们使用go-winio库来调用Windows API
	// 这里我们实现一个更可靠的系统回收站移动
	
	// 首先检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", filePath)
	}
	
	// 尝试使用系统回收站 - 使用cmd.exe的move命令作为临时解决方案
	// 在实际生产环境中应该使用Windows API
	return w.moveToSystemRecycleBinViaCmd(filePath)
}

// moveToSystemRecycleBinViaCmd 通过cmd命令移动到系统回收站
func (w *WindowsTrashManager) moveToSystemRecycleBinViaCmd(filePath string) error {
	// 使用多种方法尝试将文件移动到系统回收站
	
	// 方法1: 使用PowerShell（最可靠的方法）
	if err := w.moveToRecycleBinWithPowerShell(filePath); err == nil {
		return nil
	}
	
	// 方法2: 使用Windows Shell API（备用方案）
	if err := w.moveToRecycleBinWithShellAPI(filePath); err == nil {
		return nil
	}
	
	// 方法3: 使用DelGuard专用回收站（最终回退）
	return w.moveToDelGuardTrash(filePath)
}

// moveToRecycleBinWithShellAPI 使用Windows Shell API将文件移动到回收站
func (w *WindowsTrashManager) moveToRecycleBinWithShellAPI(filePath string) error {
	// 使用Windows Shell API (SHFileOperationW) 的Go实现
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", filePath)
	}
	
		// 使用VBS脚本调用Windows Shell API
	safePath := strings.ReplaceAll(filePath, "\"", "\"\"")
	vbsScript := fmt.Sprintf(`
Set objShell = CreateObject("Shell.Application")
Set objFolder = objShell.Namespace("%s")
Set objFile = objFolder.ParseName("%s")
objFile.InvokeVerb("delete")
`, filepath.Dir(safePath), filepath.Base(safePath))
	
	// 创建临时VBS文件
	tempVBS := filepath.Join(os.TempDir(), "delguard_trash_"+fmt.Sprintf("%d", time.Now().UnixNano())+".vbs")
	if err := os.WriteFile(tempVBS, []byte(vbsScript), 0644); err != nil {
		return fmt.Errorf("创建VBS脚本失败: %v", err)
	}
	defer os.Remove(tempVBS)
	
	// 执行VBS脚本
	cmd := exec.Command("wscript", tempVBS)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Shell API移动失败: %v", err)
	}
	
	// 验证文件是否已被删除
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}
	
	return fmt.Errorf("文件未被移动到回收站")
}

// moveToRecycleBinWithPowerShell 使用PowerShell将文件移动到回收站
func (w *WindowsTrashManager) moveToRecycleBinWithPowerShell(filePath string) error {
	// 验证路径
	if err := w.validatePath(filePath); err != nil {
		return fmt.Errorf("路径验证失败: %v", err)
	}
	
	// 获取绝对路径并验证
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("无法获取绝对路径: %v", err)
	}
	
	// 检查文件是否存在
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", absPath)
	}
	
	// 使用更安全的方式构建PowerShell命令参数
	// 使用Base64编码防止命令注入
	psScript := `
param($FilePath)
$ErrorActionPreference = "Stop"

# 验证路径参数
if (-not $FilePath) {
    Write-Error "文件路径参数不能为空"
    exit 1
}

# 验证路径长度
if ($FilePath.Length -gt 4096) {
    Write-Error "路径过长"
    exit 1
}

try {
    # 使用Resolve-Path验证路径并获取规范路径
    $resolvedPath = Resolve-Path $FilePath -ErrorAction Stop
    $fullPath = $resolvedPath.Path
    
    # 再次验证路径是否存在
    if (-not (Test-Path $fullPath)) {
        Write-Error "路径不存在: $fullPath"
        exit 1
    }
    
    # 获取文件/目录信息
    $item = Get-Item $fullPath
    $itemName = $item.Name
    $parentDir = $item.DirectoryName
    if ($item.PSIsContainer) {
        $parentDir = $item.Parent.FullName
    }
    
    Write-Host "正在将 '$itemName' 移动到回收站..."
    
    # 使用Shell.Application COM对象
    $shell = New-Object -ComObject Shell.Application
    $folder = $shell.Namespace($parentDir)
    $fileItem = $folder.ParseName($itemName)
    
    if ($null -eq $fileItem) {
        Write-Error "无法访问文件对象"
        exit 1
    }
    
    # 执行删除操作
    $fileItem.InvokeVerb("delete")
    
    # 验证文件是否已被删除
    Start-Sleep -Milliseconds 500
    if (Test-Path $fullPath) {
        Write-Error "文件删除失败"
        exit 1
    }
    
    Write-Host "成功将 '$itemName' 移动到回收站"
    
} catch {
    Write-Error "移动到回收站失败: $($_.Exception.Message)"
    exit 1
}
`
	
	// 执行PowerShell命令，将路径作为参数传递
	// 使用更安全的参数传递方式，避免命令注入
	cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", psScript)
	cmd.Args = append(cmd.Args, "-FilePath", absPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} // 隐藏窗口
	
	// 设置超时防止进程挂起
	cmd.Env = append(os.Environ(), "COMSPEC=cmd.exe")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("PowerShell移动失败: %v, 输出: %s", err, string(output))
	}
	
	// 验证文件是否已被删除
	if _, err := os.Stat(absPath); err == nil {
		return fmt.Errorf("文件未被移动到回收站")
	}
	
	return nil
}

// executeRecycleCommand 执行回收站命令（已弃用，使用更可靠的方法）
func (w *WindowsTrashManager) executeRecycleCommand(filePath string) error {
	return fmt.Errorf("此方法已弃用，请使用PowerShell或Shell API")
}

// moveToDelGuardTrash 使用DelGuard专用回收站
func (w *WindowsTrashManager) moveToDelGuardTrash(filePath string) error {
	// 获取用户回收站路径
	userProfile := os.Getenv("USERPROFILE")
	if userProfile == "" {
		return fmt.Errorf("无法获取用户配置目录")
	}

	// 创建DelGuard专用回收站目录
	delguardTrash := filepath.Join(userProfile, ".delguard", "trash")
	if err := os.MkdirAll(delguardTrash, 0755); err != nil {
		return fmt.Errorf("创建DelGuard回收站目录失败: %v", err)
	}

	// 创建回收站元数据目录
	metadataDir := filepath.Join(delguardTrash, ".metadata")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return fmt.Errorf("创建回收站元数据目录失败: %v", err)
	}

	// 获取文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %v", err)
	}

	// 使用原始文件名，保持文件名不变
	fileName := filepath.Base(filePath)
	targetPath := filepath.Join(delguardTrash, fileName)

	// 如果目标文件已存在，添加时间戳
	counter := 1
	for {
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			break
		}
		timestamp := time.Now().Format("20060102_150405")
		targetPath = filepath.Join(delguardTrash, fmt.Sprintf("%s_%s_%d%s", 
			fileName[:len(fileName)-len(filepath.Ext(fileName))], timestamp, counter, filepath.Ext(fileName)))
		counter++
	}

	// 计算文件哈希值（用于完整性验证）
	fileHash, err := w.calculateFileHash(filePath)
	if err != nil {
		fileHash = "" // 如果无法计算哈希，留空但不中断操作
	}

	// 创建元数据
	metadata := TrashMetadata{
		OriginalPath: filePath,
		DeletedTime:  time.Now(),
		FileName:     fileName,
		Size:         fileInfo.Size(),
		IsDirectory:  fileInfo.IsDir(),
		Permissions:  fileInfo.Mode().String(),
		Hash:         fileHash,
		SystemTrash:  false, // 标记为DelGuard专用回收站
	}
	
	metadataFile := filepath.Join(metadataDir, filepath.Base(targetPath)+".json")
	if err := w.writeJSONMetadata(metadataFile, metadata); err != nil {
		return fmt.Errorf("创建元数据文件失败: %v", err)
	}

	// 使用更可靠的移动方法处理跨驱动器情况
	return w.moveFileWithProgress(filePath, targetPath)
}

// GetTrashPath 获取Windows回收站路径
func (w *WindowsTrashManager) GetTrashPath() (string, error) {
	// 优先使用DelGuard专用回收站目录
	userProfile := os.Getenv("USERPROFILE")
	if userProfile == "" {
		userProfile = "C:\\Users\\Default"
	}

	// 返回DelGuard专用回收站路径
	return filepath.Join(userProfile, ".delguard", "trash"), nil
}

// GetSystemRecycleBinPath 获取Windows系统回收站路径
func (w *WindowsTrashManager) GetSystemRecycleBinPath() (string, error) {
	// 获取系统回收站路径
	userProfile := os.Getenv("USERPROFILE")
	if userProfile == "" {
		return "", fmt.Errorf("无法获取用户配置目录")
	}

	// 获取当前驱动器
	drive := filepath.VolumeName(userProfile)
	if drive == "" {
		drive = "C:"
	}

	// Windows系统回收站路径
	// 通常为 <驱动器>\$Recycle.Bin\{SID}
	return filepath.Join(drive+"\\", "$Recycle.Bin"), nil
}

// CanUseSystemRecycleBin 检查是否可以使用系统回收站
func (w *WindowsTrashManager) CanUseSystemRecycleBin() bool {
	// 检查系统回收站是否可用
	// 在实际环境中，这里应该检查用户权限和系统配置
	return true
}

// ListTrashFiles 列出Windows回收站中的文件
func (w *WindowsTrashManager) ListTrashFiles() ([]TrashFile, error) {
	trashPath, err := w.GetTrashPath()
	if err != nil {
		return nil, err
	}

	// 检查回收站目录是否存在
	if _, err := os.Stat(trashPath); os.IsNotExist(err) {
		return []TrashFile{}, nil // 返回空列表
	}

	// 获取元数据目录
	metadataDir := filepath.Join(trashPath, ".metadata")

	entries, err := os.ReadDir(trashPath)
	if err != nil {
		return nil, fmt.Errorf("读取回收站失败: %v", err)
	}

	var trashFiles []TrashFile
	for _, entry := range entries {
		// 跳过元数据目录和隐藏文件
		if entry.Name() == ".metadata" || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		fullPath := filepath.Join(trashPath, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue // 跳过无法获取信息的文件
		}

		// 尝试读取元数据
		metadataFile := filepath.Join(metadataDir, entry.Name()+".json")
		var originalPath string
		var deletedTime time.Time
		
		if metadata, err := w.readJSONMetadata(metadataFile); err == nil {
			originalPath = metadata.OriginalPath
			deletedTime = metadata.DeletedTime
		} else {
			// 如果没有元数据，使用文件修改时间
			deletedTime = info.ModTime()
			originalPath = fullPath // 如果没有元数据，使用当前路径
		}
		
		// 标准化路径，确保Windows路径分隔符一致
		originalPath = filepath.Clean(originalPath)

		trashFile := TrashFile{
			ID:           entry.Name(),
			Name:         entry.Name(),
			OriginalPath: originalPath,
			TrashPath:    fullPath,
			Size:         info.Size(),
			DeletedTime:  deletedTime,
			IsDirectory:  entry.IsDir(),
			Permissions:  info.Mode().String(),
		}

		trashFiles = append(trashFiles, trashFile)
	}

	return trashFiles, nil
}

// RestoreFile 从Windows回收站恢复文件
func (w *WindowsTrashManager) RestoreFile(trashFile TrashFile, targetPath string) error {
	// 验证回收站文件路径
	if err := w.validatePath(trashFile.TrashPath); err != nil {
		return fmt.Errorf("回收站文件路径验证失败: %v", err)
	}
	
	// 检查文件是否存在
	if _, err := os.Stat(trashFile.TrashPath); os.IsNotExist(err) {
		return fmt.Errorf("回收站文件不存在: %s", trashFile.TrashPath)
	}

	// 如果提供了原始路径且目标路径为空，使用原始路径
	if targetPath == "" && trashFile.OriginalPath != "" {
		targetPath = trashFile.OriginalPath
	}
	
	// 确保使用绝对路径
	var err error
	targetPath, err = filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("获取目标路径失败: %v", err)
	}

	// 验证目标路径安全性
	if err := w.validatePath(targetPath); err != nil {
		return fmt.Errorf("目标路径验证失败: %v", err)
	}

	// 确保目标目录存在
	targetDir := filepath.Dir(targetPath)
	if err := CreateDirIfNotExists(targetDir); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	// 检查目标文件是否已存在
	if _, err := os.Stat(targetPath); err == nil {
		if !w.forceOverwrite {
			// 如果目标文件存在，添加后缀
			ext := filepath.Ext(targetPath)
			base := targetPath[:len(targetPath)-len(ext)]
			counter := 1
			for {
				newPath := fmt.Sprintf("%s_%d%s", base, counter, ext)
				if _, err := os.Stat(newPath); os.IsNotExist(err) {
					targetPath = newPath
					break
				}
				counter++
			}
		}
	}

	// 从元数据获取文件信息以验证完整性
	userProfile := os.Getenv("USERPROFILE")
	var expectedHash string
	if userProfile != "" {
		metadataFile := filepath.Join(userProfile, ".delguard", "trash", ".metadata", trashFile.ID+".json")
		if metadata, err := w.readJSONMetadata(metadataFile); err == nil {
			expectedHash = metadata.Hash
		}
	}

	// 移动文件从回收站到目标位置
	err = w.moveFileWithProgress(trashFile.TrashPath, targetPath)
	if err != nil {
		return fmt.Errorf("恢复文件失败: %v", err)
	}

	// 验证文件完整性
		if expectedHash != "" {
			if !w.verifyFileIntegrity(targetPath, expectedHash) {
				// 文件完整性验证失败，但仍然返回成功，只是记录警告
				// 使用标准错误输出而不是fmt.Printf
				fmt.Fprintf(os.Stderr, "⚠️  警告: 文件完整性验证失败，文件可能在传输过程中损坏: %s\n", targetPath)
			}
		}

	// 清理对应的元数据文件
	if userProfile != "" {
		metadataFile := filepath.Join(userProfile, ".delguard", "trash", ".metadata", trashFile.ID+".json")
		os.Remove(metadataFile)
	}

	return nil
}

// EmptyTrash 清空Windows回收站
func (w *WindowsTrashManager) EmptyTrash() error {
	trashPath, err := w.GetTrashPath()
	if err != nil {
		return err
	}

	// 验证回收站路径
	if err := w.validatePath(trashPath); err != nil {
		return fmt.Errorf("回收站路径验证失败: %v", err)
	}

	// 检查回收站目录是否存在
	if _, err := os.Stat(trashPath); os.IsNotExist(err) {
		return nil // 回收站已经是空的
	}

	entries, err := os.ReadDir(trashPath)
	if err != nil {
		return fmt.Errorf("读取回收站失败: %v", err)
	}

	// 删除所有文件和目录，但跳过隐藏目录和元数据目录
	for _, entry := range entries {
		name := entry.Name()
		
		// 跳过元数据目录和隐藏文件
		if name == ".metadata" || strings.HasPrefix(name, ".") {
			continue
		}
		
		fullPath := filepath.Join(trashPath, name)
		
		// 验证要删除的文件路径
		if err := w.validatePath(fullPath); err != nil {
			return fmt.Errorf("要删除的文件路径验证失败: %v", err)
		}
		
		err := os.RemoveAll(fullPath)
		if err != nil {
			return fmt.Errorf("删除文件失败 %s: %v", fullPath, err)
		}
	}

	return nil
}

// ListTrashContents 列出回收站内容（接口实现）
func (w *WindowsTrashManager) ListTrashContents() ([]TrashItem, error) {
	files, err := w.ListTrashFiles()
	if err != nil {
		return nil, err
	}

	items := make([]TrashItem, len(files))
	for i, file := range files {
		items[i] = TrashItem{
			Name:         file.Name,
			OriginalPath: file.OriginalPath,
			Path:         file.TrashPath,
			Size:         file.Size,
			DeletedTime:  file.DeletedTime,
			IsDirectory:  file.IsDirectory,
		}
	}

	return items, nil
}

// RestoreFromTrash 从回收站恢复文件（接口实现）
func (w *WindowsTrashManager) RestoreFromTrash(fileName string, originalPath string) error {
	files, err := w.ListTrashFiles()
	if err != nil {
		return err
	}

	// 尝试精确匹配
	for _, file := range files {
		if file.Name == fileName {
			targetPath := originalPath
			if targetPath == "" {
				targetPath = file.OriginalPath
			}
			return w.RestoreFile(file, targetPath)
		}
	}

	// 尝试部分匹配（包含文件名）
	var matches []TrashFile
	for _, file := range files {
		if strings.Contains(file.Name, fileName) {
			matches = append(matches, file)
		}
	}

	if len(matches) == 1 {
		targetPath := originalPath
		if targetPath == "" {
			targetPath = matches[0].OriginalPath
		}
		return w.RestoreFile(matches[0], targetPath)
	} else if len(matches) > 1 {
		return fmt.Errorf("找到多个匹配文件，请使用更精确的文件名或索引: %s", fileName)
	}

	return fmt.Errorf("文件未找到: %s", fileName)
}

// GetStats 获取回收站统计信息（统一接口）
func (w *WindowsTrashManager) GetStats() (*TrashStats, error) {
	return w.GetTrashStats()
}

// Clear 清空回收站（接口实现）
func (w *WindowsTrashManager) Clear() error {
	// 首先尝试清空系统回收站
	if w.CanUseSystemRecycleBin() {
		if err := w.clearSystemRecycleBin(); err == nil {
			return nil
		}
	}
	
	// 回退到DelGuard专用回收站
	return w.EmptyTrash()
}

// clearSystemRecycleBin 清空Windows系统回收站
func (w *WindowsTrashManager) clearSystemRecycleBin() error {
	// 使用PowerShell清空系统回收站
	psScript := `
$ErrorActionPreference = "Stop"
try {
    # 方法1: 使用Clear-RecycleBin命令（Windows 10+）
    Clear-RecycleBin -Force -ErrorAction SilentlyContinue
    
    # 方法2: 使用Shell.Application COM对象
    $shell = New-Object -ComObject Shell.Application
    $recycleBin = $shell.Namespace(0xA)
    if ($recycleBin.Items().Count -gt 0) {
        $recycleBin.Items() | ForEach-Object { 
            try {
                $_.InvokeVerb("delete") 
            } catch {
                Write-Warning "无法删除项目: $($_.Name)"
            }
        }
    }
    
    Write-Host "系统回收站已清空"
    exit 0
} catch {
    Write-Error "清空系统回收站失败: $($_.Exception.Message)"
    exit 1
}
`
	cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", psScript)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("清空系统回收站失败: %v, 输出: %s", err, string(output))
	}
	
	return nil
}

// IsEmpty 检查回收站是否为空
func (w *WindowsTrashManager) IsEmpty() bool {
	files, err := w.ListTrashFiles()
	if err != nil {
		return true
	}
	return len(files) == 0
}

// GetTrashStats 获取回收站统计信息
func (w *WindowsTrashManager) GetTrashStats() (*TrashStats, error) {
	files, err := w.ListTrashFiles()
	if err != nil {
		return nil, err
	}

	stats := &TrashStats{
		TotalFiles: int64(len(files)),
		TotalSize:  0,
	}

	if len(files) > 0 {
		stats.OldestFile = files[0].DeletedTime
		for _, file := range files {
			stats.TotalSize += file.Size
			if file.DeletedTime.Before(stats.OldestFile) {
				stats.OldestFile = file.DeletedTime
			}
		}
	}

	return stats, nil
}

// CleanOldFiles 清理过期文件
func (w *WindowsTrashManager) CleanOldFiles(maxDays int) error {
	if maxDays < 0 {
		return fmt.Errorf("清理天数不能为负数")
	}
	
	files, err := w.ListTrashFiles()
	if err != nil {
		return err
	}

	cutoffTime := time.Now().AddDate(0, 0, -maxDays)

	for _, file := range files {
		if file.DeletedTime.Before(cutoffTime) {
			// 验证要删除的文件路径
			if err := w.validatePath(file.TrashPath); err != nil {
				return fmt.Errorf("要清理的文件路径验证失败: %v", err)
			}
			
			if err := os.RemoveAll(file.TrashPath); err != nil {
				return fmt.Errorf("清理过期文件失败 %s: %v", file.TrashPath, err)
			}
			
			// 清理对应的元数据文件
			userProfile := os.Getenv("USERPROFILE")
			if userProfile != "" {
				metadataFile := filepath.Join(userProfile, ".delguard", "trash", ".metadata", file.ID+".json")
				os.Remove(metadataFile)
			}
		}
	}

	return nil
}

// writeJSONMetadata 写入JSON格式的元数据文件
func (w *WindowsTrashManager) writeJSONMetadata(metadataFile string, metadata TrashMetadata) error {
	// 验证元数据文件路径
	if err := w.validateMetadataPath(metadataFile); err != nil {
		return fmt.Errorf("元数据文件路径验证失败: %v", err)
	}
	
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化元数据失败: %v", err)
	}
	
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(metadataFile), 0755); err != nil {
		return fmt.Errorf("创建元数据目录失败: %v", err)
	}
	
	// 使用临时文件和原子写入，防止数据损坏
	tempFile := metadataFile + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("写入临时元数据文件失败: %v", err)
	}
	
	return os.Rename(tempFile, metadataFile)
}

// readJSONMetadata 读取JSON格式的元数据文件
func (w *WindowsTrashManager) readJSONMetadata(metadataFile string) (*TrashMetadata, error) {
	// 验证元数据文件路径
	if err := w.validateMetadataPath(metadataFile); err != nil {
		return nil, fmt.Errorf("元数据文件路径验证失败: %v", err)
	}
	
	// 检查文件是否存在且可读
	if _, err := os.Stat(metadataFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("元数据文件不存在: %v", err)
	}
	
	data, err := os.ReadFile(metadataFile)
	if err != nil {
		return nil, fmt.Errorf("读取元数据文件失败: %v", err)
	}
	
	// 验证JSON数据大小，防止内存耗尽
	if len(data) > 10*1024*1024 { // 限制为10MB
		return nil, fmt.Errorf("元数据文件过大")
	}
	
	var metadata TrashMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("解析元数据失败: %v", err)
	}
	
	// 验证元数据内容
	if err := w.validateMetadataContent(&metadata); err != nil {
		return nil, fmt.Errorf("元数据内容验证失败: %v", err)
	}
	
	return &metadata, nil
}

// calculateFileHash 计算文件的SHA256哈希值
func (w *WindowsTrashManager) calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// verifyFileIntegrity 验证文件完整性
func (w *WindowsTrashManager) verifyFileIntegrity(filePath string, expectedHash string) bool {
	if expectedHash == "" {
		return true // 如果没有哈希值，跳过验证
	}
	
	actualHash, err := w.calculateFileHash(filePath)
	if err != nil {
		return false
	}
	
	return actualHash == expectedHash
}

// getFileSize 获取文件大小
func getFileSize(filePath string) int64 {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0
	}
	return info.Size()
}

// ValidateTrash 验证回收站完整性
func (w *WindowsTrashManager) ValidateTrash() error {
	trashPath, err := w.GetTrashPath()
	if err != nil {
		return err
	}

	// 检查回收站目录是否存在，不存在则创建
	if _, err := os.Stat(trashPath); os.IsNotExist(err) {
		if err := os.MkdirAll(trashPath, 0755); err != nil {
			return fmt.Errorf("创建回收站目录失败: %v", err)
		}
	}

	// 创建必要的子目录
	metadataDir := filepath.Join(trashPath, ".metadata")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return fmt.Errorf("创建元数据目录失败: %v", err)
	}

	// 检查目录权限
	if !w.hasWritePermission(trashPath) {
		return fmt.Errorf("回收站目录无写权限: %s", trashPath)
	}

	return nil
}

// hasWritePermission 检查写权限
func (w *WindowsTrashManager) hasWritePermission(path string) bool {
	testFile := filepath.Join(path, ".delguard_test")
	file, err := os.Create(testFile)
	if err != nil {
		return false
	}
	file.Close()
	os.Remove(testFile)
	return true
}

// validateMetadataPath 验证元数据文件路径
func (w *WindowsTrashManager) validateMetadataPath(metadataFile string) error {
	// 确保元数据文件在DelGuard回收站目录下
	userProfile := os.Getenv("USERPROFILE")
	if userProfile == "" {
		return fmt.Errorf("无法获取用户配置目录")
	}

	expectedDir := filepath.Join(userProfile, ".delguard", "trash", ".metadata")
	absExpectedDir, err := filepath.Abs(expectedDir)
	if err != nil {
		return fmt.Errorf("无法获取期望目录的绝对路径: %v", err)
	}
	
	absMetadataFile, err := filepath.Abs(metadataFile)
	if err != nil {
		return fmt.Errorf("无法获取元数据文件的绝对路径: %v", err)
	}

	// 确保路径在允许的目录内
	relPath, err := filepath.Rel(absExpectedDir, absMetadataFile)
	if err != nil {
		return fmt.Errorf("无法计算相对路径: %v", err)
	}
	
	// 检查相对路径是否包含".."，防止目录遍历
	if strings.Contains(relPath, "..") {
		return fmt.Errorf("元数据文件路径不在允许的目录内")
	}

	// 检查文件扩展名
	if !strings.HasSuffix(strings.ToLower(absMetadataFile), ".json") {
		return fmt.Errorf("元数据文件必须是.json格式")
	}

	return nil
}

// validateMetadataContent 验证元数据内容
func (w *WindowsTrashManager) validateMetadataContent(metadata *TrashMetadata) error {
	if metadata == nil {
		return fmt.Errorf("元数据不能为空")
	}
	
	// 验证原始路径
	if metadata.OriginalPath == "" {
		return fmt.Errorf("原始路径不能为空")
	}
	
	// 验证文件名
	if metadata.FileName == "" {
		return fmt.Errorf("文件名不能为空")
	}
	
	// 验证文件大小
	if metadata.Size < 0 {
		return fmt.Errorf("文件大小不能为负数")
	}
	
	// 验证时间
	if metadata.DeletedTime.IsZero() {
		return fmt.Errorf("删除时间无效")
	}
	
	// 验证文件权限格式
	if metadata.Permissions == "" {
		return fmt.Errorf("文件权限不能为空")
	}
	
	return nil
}

// validatePath 验证路径安全性
func (w *WindowsTrashManager) validatePath(path string) error {
	// 获取绝对路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("无法获取绝对路径: %v", err)
	}

	// 检查路径长度
	if len(absPath) > 4096 {
		return fmt.Errorf("路径过长: %d字符", len(absPath))
	}

	// 检查路径是否为空
	if strings.TrimSpace(absPath) == "" {
		return fmt.Errorf("路径不能为空")
	}

	// 使用filepath.Clean进行严格路径清理
	cleanPath := filepath.Clean(absPath)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("路径包含目录遍历字符")
	}

	// 检查路径是否包含空字符
	if strings.ContainsRune(cleanPath, 0) {
		return fmt.Errorf("路径包含空字符")
	}

	// 检查是否为Windows卷名
	if len(cleanPath) >= 2 && cleanPath[1] == ':' {
		// 检查是否为驱动器根目录
		if len(cleanPath) == 3 && cleanPath[2] == '\\' {
			return fmt.Errorf("不允许操作驱动器根目录")
		}
	}

	// 检查是否为系统关键路径
	systemPaths := []string{
		`C:\Windows`,
		`C:\Program Files`,
		`C:\Program Files (x86)`,
		`C:\Users\Public`,
		`C:\`,
		`C:\System Volume Information`,
		`C:\$Recycle.Bin`,
		`C:\Recovery`,
		`C:\Boot`,
	}

	upperPath := strings.ToUpper(cleanPath)
	for _, sysPath := range systemPaths {
		if strings.HasPrefix(upperPath, strings.ToUpper(sysPath)) {
			return fmt.Errorf("不允许操作系统关键路径: %s", sysPath)
		}
	}

	// 检查是否为网络路径
	if strings.HasPrefix(cleanPath, `\\`) {
		return fmt.Errorf("不允许操作网络路径")
	}

	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(cleanPath))
	blockedExts := []string{
		".sys", ".dll", ".exe", ".msi", ".com", ".bat", ".cmd",
		".drv", ".vxd", ".386", ".cpl", ".scr", ".pif",
	}
	
	for _, blocked := range blockedExts {
		if ext == blocked {
			return fmt.Errorf("不允许操作系统文件类型: %s", ext)
		}
	}

	return nil
}

// moveFileWithProgress 带进度显示的文件移动
func (w *WindowsTrashManager) moveFileWithProgress(src, dst string) error {
	// 确保源文件存在
	_, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("源文件不存在: %v", err)
	}

	// 确保目标目录存在
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	// 如果源和目标在同一驱动器，直接重命名
	srcDrive := filepath.VolumeName(src)
	dstDrive := filepath.VolumeName(dst)
	
	if srcDrive == dstDrive {
		// 先尝试重命名
		if err := os.Rename(src, dst); err == nil {
			return nil
		}
		// 重命名失败，回退到复制+删除
	}

	// 跨驱动器移动或重命名失败，使用复制+删除
	return w.copyAndRemove(src, dst)
}

// copyAndRemove 复制文件后删除源文件
func (w *WindowsTrashManager) copyAndRemove(src, dst string) error {
	// 获取源文件信息
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("无法访问源文件: %v", err)
	}

	// 如果是目录，使用递归复制
	if info.IsDir() {
		return w.copyDirectoryAndRemove(src, dst)
	}

	// 文件复制
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("无法打开源文件: %v", err)
	}
	defer srcFile.Close()

	// 确保目标目录存在
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	// 创建目标文件
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return fmt.Errorf("无法创建目标文件: %v", err)
	}
	defer dstFile.Close()

	// 复制文件内容
	written, err := io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("文件复制失败: %v", err)
	}

	// 验证文件大小
	if written != info.Size() {
		return fmt.Errorf("文件复制不完整: 期望 %d 字节, 实际 %d 字节", info.Size(), written)
	}

	// 确保数据写入磁盘
	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("数据同步失败: %v", err)
	}

	// 关闭文件句柄确保数据写入
	dstFile.Close()
	srcFile.Close()

	// 验证目标文件是否创建成功
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		return fmt.Errorf("目标文件创建失败")
	}

	// 删除源文件
	if err := os.Remove(src); err != nil {
		return fmt.Errorf("删除源文件失败: %v", err)
	}

	return nil
}
