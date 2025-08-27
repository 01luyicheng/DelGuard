//go:build windows

package delete

import (
	"fmt"
	"path/filepath"
	"syscall"
	"unsafe"
)

// moveToUnixTrash 在Windows系统上不可用
func (s *Service) moveToUnixTrash(filePath string) error {
	return fmt.Errorf("Unix回收站功能在此平台不可用")
}

// moveToWindowsRecycleBin Windows回收站实现
func (s *Service) moveToWindowsRecycleBin(filePath string) error {
	// 使用Windows API将文件移动到回收站
	shell32 := syscall.NewLazyDLL("shell32.dll")
	shFileOperation := shell32.NewProc("SHFileOperationW")

	// 转换为绝对路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("获取绝对路径失败: %v", err)
	}

	// 文件路径需要以双null结尾
	pathUTF16, err := syscall.UTF16FromString(absPath)
	if err != nil {
		return fmt.Errorf("转换文件路径失败: %v", err)
	}
	// 手动添加双null终止符
	pathUTF16 = append(pathUTF16, 0)

	// SHFILEOPSTRUCT 结构 - 正确的字段对齐
	type shFileOpStruct struct {
		hwnd                  uintptr
		wFunc                 uint32
		pFrom                 *uint16
		pTo                   *uint16
		fFlags                uint16
		fAnyOperationsAborted int32
		hNameMappings         uintptr
		lpszProgressTitle     *uint16
	}

	var fileOp shFileOpStruct
	fileOp.wFunc = 3                // FO_DELETE
	fileOp.pFrom = &pathUTF16[0]    // 指向路径字符串
	fileOp.fFlags = 0x0040 | 0x0010 // FOF_ALLOWUNDO | FOF_NOCONFIRMATION

	ret, _, _ := shFileOperation.Call(uintptr(unsafe.Pointer(&fileOp)))
	if ret != 0 {
		return fmt.Errorf("移动到回收站失败，错误代码: %d", ret)
	}

	return nil
}
