package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

// WindowsTrashManager Windows回收站管理器
type WindowsTrashManager struct {
	trashPath string
}

// NewWindowsTrashManager 创建Windows回收站管理器
func NewWindowsTrashManager() *WindowsTrashManager {
	return &WindowsTrashManager{}
}

// MoveToTrash 将文件移动到Windows回收站
func (w *WindowsTrashManager) MoveToTrash(filePath string) error {
	// 转换为绝对路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("路径转换失败: %v", err)
	}

	// 检查文件是否存在
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", absPath)
	}

	// 使用Windows Shell API移动到回收站
	return w.moveToRecycleBin(absPath)
}

// moveToRecycleBin 使用Windows API移动文件到回收站
func (w *WindowsTrashManager) moveToRecycleBin(filePath string) error {
	// 加载shell32.dll
	shell32 := syscall.NewLazyDLL("shell32.dll")
	shFileOperation := shell32.NewProc("SHFileOperationW")

	// 准备SHFILEOPSTRUCT结构
	type SHFILEOPSTRUCT struct {
		hwnd                  uintptr
		wFunc                 uint32
		pFrom                 *uint16
		pTo                   *uint16
		fFlags                uint16
		fAnyOperationsAborted int32
		hNameMappings         uintptr
		lpszProgressTitle     *uint16
	}

	// 转换路径为UTF-16
	utf16Path := syscall.StringToUTF16(filePath)
	fromPath := &utf16Path[0]

	// 设置操作参数
	fileOp := SHFILEOPSTRUCT{
		hwnd:   0,
		wFunc:  0x0003, // FO_DELETE
		pFrom:  fromPath,
		pTo:    nil,
		fFlags: 0x0040, // FOF_ALLOWUNDO (移动到回收站)
	}

	// 调用API
	ret, _, _ := shFileOperation.Call(uintptr(unsafe.Pointer(&fileOp)))
	if ret != 0 {
		return fmt.Errorf("移动到回收站失败，错误代码: %d", ret)
	}

	return nil
}

// GetTrashPath 获取Windows回收站路径
func (w *WindowsTrashManager) GetTrashPath() (string, error) {
	// Windows回收站路径通常需要管理员权限访问
	// 我们返回一个虚拟路径，因为实际的删除操作通过Shell API处理
	userProfile := os.Getenv("USERPROFILE")
	if userProfile == "" {
		userProfile = "C:\\Users\\Default"
	}

	drive := filepath.VolumeName(userProfile)
	if drive == "" {
		drive = "C:"
	}

	// 返回回收站的概念路径
	return filepath.Join(drive, "$Recycle.Bin"), nil
}

// ListTrashFiles 列出Windows回收站中的文件
func (w *WindowsTrashManager) ListTrashFiles() ([]TrashFile, error) {
	// Windows回收站访问需要特殊权限，建议用户使用系统回收站查看
	return nil, fmt.Errorf("Windows回收站列表功能需要管理员权限，请使用系统回收站查看文件")
}

// RestoreFile 从Windows回收站恢复文件
func (w *WindowsTrashManager) RestoreFile(trashFile TrashFile, targetPath string) error {
	// 确保目标目录存在
	targetDir := filepath.Dir(targetPath)
	if err := CreateDirIfNotExists(targetDir); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	// 移动文件从回收站到目标位置
	err := os.Rename(trashFile.TrashPath, targetPath)
	if err != nil {
		return fmt.Errorf("恢复文件失败: %v", err)
	}

	return nil
}

// EmptyTrash 清空Windows回收站
func (w *WindowsTrashManager) EmptyTrash() error {
	// 加载shell32.dll
	shell32 := syscall.NewLazyDLL("shell32.dll")
	shEmptyRecycleBin := shell32.NewProc("SHEmptyRecycleBinW")

	// 调用API清空回收站
	ret, _, _ := shEmptyRecycleBin.Call(0, 0, 0x0001) // SHERB_NOCONFIRMATION
	if ret != 0 {
		return fmt.Errorf("清空回收站失败，错误代码: %d", ret)
	}

	return nil
}
