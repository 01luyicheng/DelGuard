//go:build windows
// +build windows

package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

var (
	shell32              = syscall.NewLazyDLL("shell32.dll")
	procSHFileOperationW = shell32.NewProc("SHFileOperationW")
)

// SHFILEOPSTRUCT Windows API结构体
type SHFILEOPSTRUCT struct {
	Hwnd                  uintptr
	WFunc                 uint32
	PFrom                 *uint16
	PTo                   *uint16
	FFlags                uint16
	FAnyOperationsAborted bool
	HNameMappings         uintptr
	LpszProgressTitle     *uint16
}

const (
	FO_DELETE          = 0x0003
	FOF_ALLOWUNDO      = 0x0040
	FOF_NOCONFIRMATION = 0x0010
	FOF_SILENT         = 0x0004
)

// moveToTrashWindows 将文件移动到Windows回收站
func moveToTrashWindows(filePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return ErrFileNotFound
	}

	// 获取绝对路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}

	// Windows SHFileOperation 要求 PFrom 为 双空终止 的列表（MULTI_SZ）
	// syscall.UTF16FromString返回单空终止，这里手动再追加一个 0
	u16, err := syscall.UTF16FromString(absPath)
	if err != nil {
		return err
	}
	u16 = append(u16, 0) // 追加第二个 0 形成双空终止

	// 确保切片不为空再获取指针
	if len(u16) == 0 {
		return E(KindIO, "moveToTrash", absPath, nil, "路径转换失败")
	}

	fileOp := SHFILEOPSTRUCT{
		WFunc:  FO_DELETE,
		PFrom:  &u16[0],
		FFlags: FOF_ALLOWUNDO | FOF_NOCONFIRMATION | FOF_SILENT,
	}

	// 调用SHFileOperationW API
	ret, _, err := procSHFileOperationW.Call(uintptr(unsafe.Pointer(&fileOp)))
	if ret != 0 {
		// 根据返回值提供更详细的错误信息
		switch ret {
		case 0x75: // ERROR_CANCELLED
			return E(KindCancelled, "moveToTrash", absPath, nil, "用户取消了操作")
		case 0x5: // ERROR_ACCESS_DENIED
			return E(KindPermission, "moveToTrash", absPath, nil, "权限不足，无法访问文件")
		case 0x2: // ERROR_FILE_NOT_FOUND
			return E(KindNotFound, "moveToTrash", absPath, nil, "文件不存在")
		default:
			// 检查是否有底层系统错误
			if err != nil {
				if syscallErr, ok := err.(syscall.Errno); ok && syscallErr != 0 {
					return E(KindIO, "moveToTrash", absPath, err, fmt.Sprintf("Windows API 错误码: 0x%x", uint32(ret)))
				}
			}
			return E(KindIO, "moveToTrash", absPath, nil, fmt.Sprintf("Windows API 错误码: 0x%x", ret))
		}
	}

	return nil
}

// 为Windows平台提供其他平台函数的存根
func moveToTrashMacOS(filePath string) error {
	return ErrUnsupportedPlatform
}

func moveToTrashLinux(filePath string) error {
	return ErrUnsupportedPlatform
}

// DecodeTrashInfoPath 解码.trashinfo中的Path字段，供其他平台使用
func DecodeTrashInfoPath(p string) string {
	return decodeTrashInfoPath(p)
}

// decodeTrashInfoPath 解码.trashinfo中的Path字段
func decodeTrashInfoPath(p string) string {
	if p == "" {
		return ""
	}
	parts := strings.Split(p, "/")
	for i := range parts {
		if parts[i] != "" {
			if decoded, err := url.PathUnescape(parts[i]); err == nil {
				parts[i] = decoded
			}
		}
	}
	return strings.Join(parts, "/")
}

// isWindowsHiddenFile 检查Windows文件是否为隐藏文件
func isWindowsHiddenFile(filePath string) bool {
	// 使用Windows API检查文件属性
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getFileAttributes := kernel32.NewProc("GetFileAttributesW")
	
	// 转换路径为UTF16
	pathPtr, err := syscall.UTF16PtrFromString(filePath)
	if err != nil {
		return false
	}
	
	// 调用Windows API获取文件属性
	attrs, _, err := getFileAttributes.Call(uintptr(unsafe.Pointer(pathPtr)))
	if attrs == 0xffffffff {
		// 获取属性失败，保守起见返回true
		return true
	}
	
	// 检查隐藏属性位 (FILE_ATTRIBUTE_HIDDEN = 0x2)
	const FILE_ATTRIBUTE_HIDDEN = 0x00000002
	return (attrs & FILE_ATTRIBUTE_HIDDEN) != 0
}

// checkFileOwnershipWindows 检查Windows文件所有权
func checkFileOwnershipWindows(filePath string) error {
	// Windows平台的所有权检查相对复杂，这里实现基础检查
	// 检查文件是否可访问
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("无法访问文件: %v", err)
	}
	defer file.Close()
	
	// 检查文件权限
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("无法获取文件信息: %v", err)
	}
	
	// 检查是否为只读
	if info.Mode().Perm()&0222 == 0 {
		return fmt.Errorf("文件为只读")
	}
	
	return nil
}
