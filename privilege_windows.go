//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

// IsElevated 检查当前用户是否具有管理员权限
func IsElevated() bool {
	// 使用Windows API检查管理员权限
	// 首先尝试使用shell32.dll的IsUserAnAdmin函数
	if isUserAnAdmin, err := isUserAdminShell32(); err == nil {
		return isUserAnAdmin
	}

	// 备用方法：检查管理员组成员身份
	return checkAdminGroupMembership()
}

// isUserAdminShell32 使用shell32.dll检查管理员权限
func isUserAdminShell32() (bool, error) {
	shell32 := syscall.NewLazyDLL("shell32.dll")
	procIsUserAnAdmin := shell32.NewProc("IsUserAnAdmin")

	r1, _, err := procIsUserAnAdmin.Call()
	if err != nil && err.Error() != "The operation completed successfully" {
		return false, err
	}
	return r1 != 0, nil
}

// checkAdminGroupMembership 检查管理员组成员身份
func checkAdminGroupMembership() bool {
	// 检查管理员组成员身份
	var sid *syscall.SID

	// SECURITY_NT_AUTHORITY: {0,0,0,0,0,5}
	ntAuthority := struct{ Value [6]byte }{Value: [6]byte{0, 0, 0, 0, 0, 5}}

	// 分配并初始化管理员组SID
	// SECURITY_BUILTIN_DOMAIN_RID: 0x00000020
	// DOMAIN_ALIAS_RID_ADMINS: 0x00000220
	err := syscallAllocateAndInitializeSid(&ntAuthority, 2,
		0x00000020, // SECURITY_BUILTIN_DOMAIN_RID
		0x00000220, // DOMAIN_ALIAS_RID_ADMINS
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		return false
	}
	defer syscallFreeSid(sid)

	// 检查当前进程令牌是否属于管理员组
	var isAdmin int32
	token := syscall.Token(0) // 当前进程令牌

	// 使用CheckTokenMembership检查
	err = checkTokenMembership(token, sid, &isAdmin)
	if err != nil {
		// 如果CheckTokenMembership失败，尝试备用方法
		return checkTokenElevation()
	}
	return isAdmin != 0
}

// checkTokenElevation 检查令牌提升状态
func checkTokenElevation() bool {
	var token syscall.Token
	err := openProcessToken(syscall.Handle(^uintptr(0)), syscall.TOKEN_QUERY, &token)
	if err != nil {
		return false
	}
	defer closeHandle(syscall.Handle(token))

	var elevation uint32
	var returnLength uint32
	err = getTokenInformation(token, syscall.TokenElevation,
		(*byte)(unsafe.Pointer(&elevation)),
		uint32(unsafe.Sizeof(elevation)),
		&returnLength)
	if err != nil {
		return false
	}
	return elevation != 0
}

// Windows API函数包装
func syscallAllocateAndInitializeSid(identAuth *struct{ Value [6]byte }, subAuthCount uint8, subAuth0, subAuth1 uint32, subAuth2, subAuth3, subAuth4, subAuth5, subAuth6, subAuth7 uint32, sid **syscall.SID) error {
	advapi32 := syscall.NewLazyDLL("advapi32.dll")
	proc := advapi32.NewProc("AllocateAndInitializeSid")

	r1, _, err := proc.Call(
		uintptr(unsafe.Pointer(identAuth)),
		uintptr(subAuthCount),
		uintptr(subAuth0), uintptr(subAuth1), uintptr(subAuth2), uintptr(subAuth3),
		uintptr(subAuth4), uintptr(subAuth5), uintptr(subAuth6), uintptr(subAuth7),
		uintptr(unsafe.Pointer(sid)),
	)
	if r1 == 0 {
		return err
	}
	return nil
}

func syscallFreeSid(sid *syscall.SID) {
	advapi32 := syscall.NewLazyDLL("advapi32.dll")
	proc := advapi32.NewProc("FreeSid")
	proc.Call(uintptr(unsafe.Pointer(sid)))
}

func checkTokenMembership(token syscall.Token, sid *syscall.SID, isMember *int32) error {
	advapi32 := syscall.NewLazyDLL("advapi32.dll")
	proc := advapi32.NewProc("CheckTokenMembership")

	r1, _, err := proc.Call(
		uintptr(token),
		uintptr(unsafe.Pointer(sid)),
		uintptr(unsafe.Pointer(isMember)),
	)
	if r1 == 0 {
		return err
	}
	return nil
}

func openProcessToken(process syscall.Handle, desiredAccess uint32, token *syscall.Token) error {
	advapi32 := syscall.NewLazyDLL("advapi32.dll")
	proc := advapi32.NewProc("OpenProcessToken")

	r1, _, err := proc.Call(
		uintptr(process),
		uintptr(desiredAccess),
		uintptr(unsafe.Pointer(token)),
	)
	if r1 == 0 {
		return err
	}
	return nil
}

func closeHandle(handle syscall.Handle) error {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("CloseHandle")

	r1, _, err := proc.Call(uintptr(handle))
	if r1 == 0 {
		return err
	}
	return nil
}

func getTokenInformation(token syscall.Token, infoClass uint32, info *byte, infoLength uint32, returnLength *uint32) error {
	advapi32 := syscall.NewLazyDLL("advapi32.dll")
	proc := advapi32.NewProc("GetTokenInformation")

	r1, _, err := proc.Call(
		uintptr(token),
		uintptr(infoClass),
		uintptr(unsafe.Pointer(info)),
		uintptr(infoLength),
		uintptr(unsafe.Pointer(returnLength)),
	)
	if r1 == 0 {
		return err
	}
	return nil
}

// 增强版Windows权限管理 - 添加UAC支持和安全检查

// UAC相关常量
const (
	SEE_MASK_DEFAULT        = 0x00000000
	SEE_MASK_NOCLOSEPROCESS = 0x00000040
	SEE_MASK_FLAG_NO_UI     = 0x00000400

	SW_HIDE            = 0
	SW_SHOWNORMAL      = 1
	SW_SHOWMINIMIZED   = 2
	SW_SHOWMAXIMIZED   = 3
	SW_SHOWNOACTIVATE  = 4
	SW_SHOW            = 5
	SW_MINIMIZE        = 6
	SW_SHOWMINNOACTIVE = 7
	SW_SHOWNA          = 8
	SW_RESTORE         = 9
	SW_SHOWDEFAULT     = 10
	SW_FORCEMINIMIZE   = 11
)

// SHELLEXECUTEINFO Windows API结构体
type SHELLEXECUTEINFO struct {
	CbSize       uint32
	FMask        uint32
	Hwnd         uintptr
	LpVerb       *uint16
	LpFile       *uint16
	LpParameters *uint16
	LpDirectory  *uint16
	NShow        int32
	HInstApp     uintptr
	LpIDList     uintptr
	LpClass      *uint16
	HkeyClass    uintptr
	DwHotKey     uint32
	HMonitor     uintptr
	HProcess     uintptr
}

// CheckUACAndPrompt 检查UAC并提示用户
func CheckUACAndPrompt(operation string) (bool, error) {
	if IsElevated() {
		return true, nil
	}

	// 如果不是管理员，提示用户
	fmt.Printf("需要管理员权限执行操作: %s\n", operation)
	fmt.Println("系统将请求UAC权限提升...")

	// 使用ShellExecuteEx请求UAC提升
	return RequestUACElevation(operation)
}

// RequestUACElevation 请求UAC权限提升
func RequestUACElevation(operation string) (bool, error) {
	// 获取当前可执行文件路径
	exePath, err := os.Executable()
	if err != nil {
		return false, fmt.Errorf("无法获取可执行文件路径: %v", err)
	}

	// 构建参数
	args := fmt.Sprintf("--uac-operation=%s", operation)

	// 使用ShellExecuteEx以管理员权限运行
	return runAsAdmin(exePath, args)
}

// runAsAdmin 以管理员权限运行程序
func runAsAdmin(executable, parameters string) (bool, error) {
	shell32 := syscall.NewLazyDLL("shell32.dll")
	procShellExecuteEx := shell32.NewProc("ShellExecuteExW")

	verb, _ := syscall.UTF16PtrFromString("runas")
	file, _ := syscall.UTF16PtrFromString(executable)
	params, _ := syscall.UTF16PtrFromString(parameters)

	execInfo := SHELLEXECUTEINFO{
		CbSize:       uint32(unsafe.Sizeof(SHELLEXECUTEINFO{})),
		FMask:        SEE_MASK_NOCLOSEPROCESS,
		Hwnd:         0,
		LpVerb:       verb,
		LpFile:       file,
		LpParameters: params,
		NShow:        SW_SHOWNORMAL,
	}

	r1, _, err := procShellExecuteEx.Call(uintptr(unsafe.Pointer(&execInfo)))
	if r1 == 0 {
		if err != nil && strings.Contains(err.Error(), "ERROR_CANCELLED") {
			return false, fmt.Errorf("用户取消了UAC权限请求")
		}
		return false, fmt.Errorf("UAC权限请求失败: %v", err)
	}

	// 等待进程完成
	if execInfo.HProcess != 0 {
		kernel32 := syscall.NewLazyDLL("kernel32.dll")
		procWaitForSingleObject := kernel32.NewProc("WaitForSingleObject")
		procCloseHandle := kernel32.NewProc("CloseHandle")

		procWaitForSingleObject.Call(execInfo.HProcess, 0xFFFFFFFF) // INFINITE
		procCloseHandle.Call(execInfo.HProcess)
	}

	return true, nil
}

// CheckFilePermissions 检查文件权限
func CheckFilePermissions(filePath string) (bool, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false, fmt.Errorf("文件不存在: %s", filePath)
	}

	// 尝试以写权限打开文件
	file, err := os.OpenFile(filePath, os.O_WRONLY, 0666)
	if err != nil {
		if os.IsPermission(err) {
			return false, nil // 无权限
		}
		return false, err
	}
	file.Close()
	return true, nil
}

// CheckDirectoryPermissions 检查目录权限
func CheckDirectoryPermissions(dirPath string) (bool, error) {
	// 检查目录是否存在
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// 如果目录不存在，检查父目录权限
		parent := filepath.Dir(dirPath)
		return CheckDirectoryPermissions(parent)
	}

	// 检查目录是否可写
	testFile := filepath.Join(dirPath, ".delguard_permission_test")
	file, err := os.Create(testFile)
	if err != nil {
		if os.IsPermission(err) {
			return false, nil
		}
		return false, err
	}
	file.Close()
	os.Remove(testFile)
	return true, nil
}

// EnsureWriteAccess 确保有写权限
func EnsureWriteAccess(path string) error {
	// 检查路径是否存在
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// 检查父目录权限
			parent := filepath.Dir(path)
			return EnsureWriteAccess(parent)
		}
		return err
	}

	// 如果是目录，检查目录权限
	if info.IsDir() {
		if ok, err := CheckDirectoryPermissions(path); err != nil {
			return err
		} else if !ok {
			return fmt.Errorf("目录无写权限: %s", path)
		}
		return nil
	}

	// 如果是文件，检查文件权限
	if ok, err := CheckFilePermissions(path); err != nil {
		return err
	} else if !ok {
		// 尝试请求UAC权限
		if elevated, err := RequestUACElevation(fmt.Sprintf("获取文件写权限: %s", path)); err != nil {
			return err
		} else if !elevated {
			return fmt.Errorf("需要管理员权限修改文件: %s", path)
		}
	}

	return nil
}

// GetProcessIntegrityLevel 获取进程完整性级别
func GetProcessIntegrityLevel() (string, error) {
	advapi32 := syscall.NewLazyDLL("advapi32.dll")
	kernel32 := syscall.NewLazyDLL("kernel32.dll")

	procGetCurrentProcess := kernel32.NewProc("GetCurrentProcess")
	procOpenProcessToken := advapi32.NewProc("OpenProcessToken")
	procGetTokenInformation := advapi32.NewProc("GetTokenInformation")
	procGetSidSubAuthority := advapi32.NewProc("GetSidSubAuthority")
	procGetSidSubAuthorityCount := advapi32.NewProc("GetSidSubAuthorityCount")

	// 获取当前进程
	process, _, _ := procGetCurrentProcess.Call()

	// 打开进程令牌
	var token syscall.Token
	r1, _, err := procOpenProcessToken.Call(process, syscall.TOKEN_QUERY, uintptr(unsafe.Pointer(&token)))
	if r1 == 0 {
		return "unknown", err
	}
	defer token.Close()

	// 获取完整性级别SID
	var needed uint32
	procGetTokenInformation.Call(uintptr(token), syscall.TokenIntegrityLevel, 0, 0, uintptr(unsafe.Pointer(&needed)))
	if needed == 0 {
		return "unknown", nil
	}

	buffer := make([]byte, needed)
	r1, _, err = procGetTokenInformation.Call(uintptr(token), syscall.TokenIntegrityLevel,
		uintptr(unsafe.Pointer(&buffer[0])), uintptr(needed), uintptr(unsafe.Pointer(&needed)))
	if r1 == 0 {
		return "unknown", err
	}

	// 解析SID获取完整性级别
	sid := (*syscall.SID)(unsafe.Pointer(&buffer[0]))
	count, _, _ := procGetSidSubAuthorityCount.Call(uintptr(unsafe.Pointer(sid)))
	if count == 0 {
		return "unknown", nil
	}

	lastSubAuthority, _, _ := procGetSidSubAuthority.Call(uintptr(unsafe.Pointer(sid)), uintptr(count-1))
	if lastSubAuthority == 0 {
		return "unknown", nil
	}

	// 安全地转换结果
	level := uint32(lastSubAuthority)
	switch level {
	case 0x0000:
		return "untrusted", nil
	case 0x1000:
		return "low", nil
	case 0x2000:
		return "medium", nil
	case 0x3000:
		return "high", nil
	case 0x4000:
		return "system", nil
	default:
		return fmt.Sprintf("level-%d", level), nil
	}
}

// IsProtectedMode 检查是否在保护模式下运行
func IsProtectedMode() bool {
	level, err := GetProcessIntegrityLevel()
	if err != nil {
		return false
	}

	// 低完整性级别或以下被认为是保护模式
	return level == "low" || level == "untrusted"
}

// ValidateSystemPath 验证系统路径的安全性
func ValidateSystemPath(path string) error {
	// 清理路径
	cleanPath := filepath.Clean(path)

	// 检查是否为系统关键路径
	systemPaths := []string{
		`C:\Windows`,
		`C:\Program Files`,
		`C:\Program Files (x86)`,
		`C:\ProgramData`,
		`C:\Users\All Users`,
	}

	cleanPathLower := strings.ToLower(cleanPath)
	for _, sysPath := range systemPaths {
		if strings.HasPrefix(cleanPathLower, strings.ToLower(sysPath)) {
			if !IsElevated() {
				return fmt.Errorf("需要管理员权限访问系统路径: %s", path)
			}
		}
	}

	// 检查文件所有权和权限
	if err := checkFileOwnershipWindows(path); err != nil {
		return fmt.Errorf("文件所有权检查失败: %v", err)
	}

	return nil
}

// checkFileOwnershipWindows 检查Windows文件所有权
func checkFileOwnershipWindows(filePath string) error {
	// Windows平台的所有权检查相对复杂，这里实现基础检查
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

	// 检查文件是否为系统文件
	if isWindowsSystemFile(filePath) {
		return fmt.Errorf("系统文件不允许操作")
	}

	// 检查文件是否被其他进程锁定
	if err := checkFileLock(filePath); err != nil {
		return fmt.Errorf("文件被锁定: %v", err)
	}

	return nil
}

// isWindowsSystemFile 检查是否为Windows系统文件
func isWindowsSystemFile(filePath string) bool {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getFileAttributes := kernel32.NewProc("GetFileAttributesW")

	pathPtr, err := syscall.UTF16PtrFromString(filePath)
	if err != nil {
		return false
	}

	attrs, _, err := getFileAttributes.Call(uintptr(unsafe.Pointer(pathPtr)))
	if attrs == 0xffffffff {
		return false
	}

	// 检查系统文件属性 (FILE_ATTRIBUTE_SYSTEM = 0x4)
	const FILE_ATTRIBUTE_SYSTEM = 0x00000004
	return (attrs & FILE_ATTRIBUTE_SYSTEM) != 0
}

// checkFileLock 检查文件是否被锁定
func checkFileLock(filePath string) error {
	// 尝试以独占模式打开文件
	file, err := os.OpenFile(filePath, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	// 尝试获取文件锁（Windows需要特殊处理）
	// 这里简化处理，实际应该使用LockFileEx
	return nil
}
