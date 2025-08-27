//go:build windows
// +build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

// moveToTrashWindows Windows平台回收站删除
func (cd *CoreDeleter) moveToTrashWindows(filePath string) error {
	// 1. 基本输入验证
	if filePath == "" {
		return fmt.Errorf("文件路径不能为空")
	}

	// 2. 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", filePath)
	}

	// 3. 获取绝对路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("无法获取绝对路径: %v", err)
	}

	// 4. 验证路径长度上限，防止缓冲区溢出
	if len(absPath) > 32760 { // Windows 最大路径长度
		return fmt.Errorf("路径太长，超过Windows限制")
	}

	// 5. 安全的UTF16转换 - 确保双空终止
	u16, err := syscall.UTF16FromString(absPath)
	if err != nil {
		return fmt.Errorf("UTF16转换失败: %v", err)
	}
	// 添加第二个空终止符（Windows SHFileOperation API 要求双空终止）
	u16 = append(u16, 0)

	// 检查是否为目录
	info, err := os.Stat(absPath)
	var flags uint16 = FOF_ALLOWUNDO | FOF_NOCONFIRMATION | FOF_SILENT | FOF_NOERRORUI
	if err == nil && info.IsDir() {
		// 对于目录，确保使用正确的标志
		flags = FOF_ALLOWUNDO | FOF_NOCONFIRMATION | FOF_SILENT | FOF_NOERRORUI
	}

	// 初始化SHFILEOPSTRUCT结构体
	fileOp := SHFILEOPSTRUCT{
		Hwnd:                  0,
		WFunc:                 FO_DELETE,
		PFrom:                 &u16[0],
		PTo:                   nil,
		FFlags:                flags,
		FAnyOperationsAborted: false,
		HNameMappings:         0,
		LpszProgressTitle:     nil,
	}

	// 调用SHFileOperationW API
	ret, _, err := procSHFileOperationW.Call(uintptr(unsafe.Pointer(&fileOp)))
	if ret != 0 {
		// 根据返回值提供更详细的错误信息
		switch ret {
		case 0x75: // ERROR_CANCELLED
			return fmt.Errorf("用户取消了操作")
		case 0x5: // ERROR_ACCESS_DENIED
			return fmt.Errorf("权限不足，无法访问文件")
		case 0x2: // ERROR_FILE_NOT_FOUND
			return fmt.Errorf("文件不存在")
		case 0x78: // ERROR_ALREADY_EXISTS
			return fmt.Errorf("目标文件已存在")
		case 0x6: // ERROR_INVALID_HANDLE
			return fmt.Errorf("无效的文件句柄")
		default:
			return fmt.Errorf("Windows API 错误码: 0x%x", ret)
		}
	}

	// 检查操作是否被中止
	if fileOp.FAnyOperationsAborted {
		return fmt.Errorf("操作被中止")
	}

	return nil
}

// moveToTrashMacOS macOS平台回收站删除（存根）
func (cd *CoreDeleter) moveToTrashMacOS(path string) error {
	// 暂时使用永久删除
	return os.Remove(path)
}

// moveToTrashLinux Linux平台回收站删除（存根）
func (cd *CoreDeleter) moveToTrashLinux(path string) error {
	// 暂时使用永久删除
	return os.Remove(path)
}
