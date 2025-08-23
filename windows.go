//go:build windows
// +build windows

package main

import (
	"fmt"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

var (
	shell32              = syscall.NewLazyDLL("shell32.dll")
	procSHFileOperationW = shell32.NewProc("SHFileOperationW")

	// advapi32 DLL for user SID functions
	advapi32           = syscall.NewLazyDLL("advapi32.dll")
	procGetUserNameExW = advapi32.NewProc("GetUserNameExW")
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

	// Name format constants for GetUserNameEx
	NameSamCompatible = 2
	NameUniqueId      = 6
	NameCanonical     = 7
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
		case 0x78: // ERROR_ALREADY_EXISTS
			return E(KindIO, "moveToTrash", absPath, nil, "目标文件已存在")
		case 0x6: // ERROR_INVALID_HANDLE
			return E(KindIO, "moveToTrash", absPath, nil, "无效的文件句柄")
		default:
			// 检查是否有底层系统错误
			if err != nil && err != syscall.Errno(0) {
				if syscallErr, ok := err.(syscall.Errno); ok && syscallErr != 0 {
					return E(KindIO, "moveToTrash", absPath, err, fmt.Sprintf("Windows API 错误码: 0x%x", uint32(ret)))
				}
			}
			return E(KindIO, "moveToTrash", absPath, nil, fmt.Sprintf("Windows API 错误码: 0x%x", ret))
		}
	}

	// 检查操作是否被中止
	if fileOp.FAnyOperationsAborted {
		return E(KindCancelled, "moveToTrash", absPath, nil, "操作被中止")
	}

	return nil
}

// getCurrentUserSID 获取当前用户的SID
func getCurrentUserSID() (string, error) {
	// 首先尝试使用标准库获取用户信息
	currentUser, err := user.Current()
	if err == nil && currentUser.Uid != "" {
		return currentUser.Uid, nil
	}

	// 如果标准方法失败，尝试使用Windows API
	username, err := getWindowsUsername()
	if err != nil {
		// 最后的备选方案
		return "S-1-5-21", nil
	}

	return username, nil
}

// getWindowsUsername 获取Windows用户名
func getWindowsUsername() (string, error) {
	// 使用GetUserNameExW获取用户名
	var size uint32
	// 首次调用获取所需缓冲区大小
	ret, _, _ := procGetUserNameExW.Call(
		uintptr(NameSamCompatible),
		uintptr(0),
		uintptr(unsafe.Pointer(&size)),
	)

	if ret != 0 || size == 0 {
		// 备选方法：使用环境变量
		username := os.Getenv("USERNAME")
		if username != "" {
			return username, nil
		}
		computername := os.Getenv("COMPUTERNAME")
		if computername != "" && username != "" {
			return computername + "\\" + username, nil
		}
		return "", fmt.Errorf("无法获取用户名")
	}

	// 分配缓冲区并再次调用
	buf := make([]uint16, size)
	ret, _, _ = procGetUserNameExW.Call(
		uintptr(NameSamCompatible),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)

	if ret == 0 {
		return "", fmt.Errorf("无法获取用户名")
	}

	// 转换UTF-16到字符串
	username := syscall.UTF16ToString(buf)
	return username, nil
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
	// 首先检查以点开头的文件名（Unix风格隐藏文件）
	filename := filepath.Base(filePath)
	if strings.HasPrefix(filename, ".") && filename != "." && filename != ".." {
		return true
	}

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

// checkDiskSpace Windows平台检查磁盘空间
func checkDiskSpace(path string, requiredBytes int64) error {
	if path == "" {
		return fmt.Errorf("路径不能为空")
	}

	// 获取路径所在的驱动器
	volumePath, err := getVolumePath(path)
	if err != nil {
		return fmt.Errorf("获取卷路径失败: %v", err)
	}

	// 使用Windows API获取磁盘空间信息
	var freeBytesAvailable, totalBytes, totalFreeBytes uint64
	volumePathPtr, err := syscall.UTF16PtrFromString(volumePath)
	if err != nil {
		return fmt.Errorf("转换卷路径失败: %v", err)
	}

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpaceEx := kernel32.NewProc("GetDiskFreeSpaceExW")
	r1, _, err := getDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(volumePathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)

	if r1 == 0 {
		return fmt.Errorf("获取磁盘空间失败: %v", err)
	}

	// 检查可用空间
	if int64(freeBytesAvailable) < requiredBytes {
		return fmt.Errorf("磁盘空间不足，需要 %s，可用 %s",
			formatBytes(requiredBytes), formatBytes(int64(freeBytesAvailable)))
	}

	return nil
}

// getVolumePath 获取路径所在的卷路径
func getVolumePath(path string) (string, error) {
	// 确保路径是绝对路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	// 处理UNC路径 (\\server\share\path)
	if strings.HasPrefix(absPath, `\\`) {
		// 查找第三个反斜杠的位置（在\\server\share之后）
		parts := strings.SplitN(absPath[2:], `\`, 3)
		if len(parts) >= 2 {
			// 返回 \\server\share 部分
			return `\\` + parts[0] + `\` + parts[1], nil
		}
		// 如果格式不正确，返回原始路径
		return absPath, nil
	}

	// Windows卷路径格式 (C:\path)
	if len(absPath) >= 2 && absPath[1] == ':' {
		return absPath[:3], nil
	}

	// 默认返回C盘
	return "C:\\", nil
}
