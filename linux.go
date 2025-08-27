//go:build linux
// +build linux

package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// moveToTrashLinux 将文件移动到Linux回收站
func moveToTrashLinux(filePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return ErrFileNotFound
	}

	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// 确定回收站目录
	trashDir := filepath.Join(homeDir, ".local", "share", "Trash")
	trashFilesDir := filepath.Join(trashDir, "files")
	trashInfoDir := filepath.Join(trashDir, "info")

	// 确保回收站目录存在
	if err := os.MkdirAll(trashFilesDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(trashInfoDir, 0755); err != nil {
		return err
	}

	// 获取文件的绝对路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}

	// 生成唯一的文件名以避免冲突
	baseName := filepath.Base(absPath)
	trashFileName := baseName
	trashFilePath := filepath.Join(trashFilesDir, trashFileName)

	counter := 1
	for {
		if _, err := os.Stat(trashFilePath); os.IsNotExist(err) {
			break
		}
		// 文件名已存在，添加数字后缀
		ext := filepath.Ext(baseName)
		nameWithoutExt := strings.TrimSuffix(baseName, ext)
		trashFileName = fmt.Sprintf("%s.%d%s", nameWithoutExt, counter, ext)
		trashFilePath = filepath.Join(trashFilesDir, trashFileName)
		counter++
	}

	// 移动文件到回收站
	if err := os.Rename(absPath, trashFilePath); err != nil {
		// 如果跨设备，需要复制然后删除
		if isEXDEV(err) {
			if err := copyTree(absPath, trashFilePath); err != nil {
				return err
			}
			if err := removeOriginal(absPath); err != nil {
				// 如果删除失败，也要删除复制的文件
				os.Remove(trashFilePath)
				return err
			}
		} else {
			return err
		}
	}

	// 创建 .trashinfo 文件
	trashInfoPath := filepath.Join(trashInfoDir, trashFileName+".trashinfo")
	trashInfoContent := fmt.Sprintf(
		"[Trash Info]\nPath=%s\nDeletionDate=%s\n",
		absPath,
		time.Now().Format("2006-01-02T15:04:05"),
	)

	if err := os.WriteFile(trashInfoPath, []byte(trashInfoContent), 0644); err != nil {
		// 如果创建.trashinfo文件失败，应该将文件移回原位
		os.Rename(trashFilePath, absPath)
		os.Remove(trashInfoPath)
		return err
	}

	return nil
}

// getCurrentUserSID 获取当前用户的SID（Linux实现）
func getCurrentUserSID() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	return currentUser.Uid, nil
}

// CheckFilePermissions 检查文件权限（Linux实现）
func CheckFilePermissions(filePath string) (bool, error) {
	// Linux平台的文件权限检查
	info, err := os.Stat(filePath)
	if err != nil {
		return false, err
	}

	// 检查基本文件权限
	if info.Mode().Perm()&0002 != 0 {
		// 世界可写文件，可能存在风险
		return false, fmt.Errorf("文件权限过于宽松: %s", info.Mode().Perm())
	}

	// 检查文件所有者
	currentUser, err := user.Current()
	if err != nil {
		return false, fmt.Errorf("无法获取当前用户信息: %v", err)
	}

	// 检查文件所有者匹配
	fileInfo, err := os.Lstat(filePath)
	if err != nil {
		return false, fmt.Errorf("无法获取文件详细信息: %v", err)
	}

	// 获取文件系统统计信息
	var stat syscall.Stat_t
	if err := syscall.Stat(filePath, &stat); err != nil {
		return false, fmt.Errorf("无法获取文件系统统计信息: %v", err)
	}

	// 检查文件所有者
	fileUID := strconv.Itoa(int(stat.Uid))
	if fileUID != currentUser.Uid {
		// 文件不属于当前用户，需要额外权限检查
		return false, fmt.Errorf("文件不属于当前用户，权限不足")
	}

	// 检查SELinux上下文（如果可用）
	if hasSELinux() {
		context, err := getSELinuxContext(filePath)
		if err != nil {
			return false, fmt.Errorf("无法获取SELinux上下文: %v", err)
		}

		// 检查SELinux上下文是否允许访问
		if !isSELinuxAllowed(context) {
			return false, fmt.Errorf("SELinux策略禁止访问")
		}
	}

	// 检查ACL权限（如果可用）
	if hasACL() {
		aclPerms, err := getACLPermissions(filePath)
		if err != nil {
			return false, fmt.Errorf("无法获取ACL权限: %v", err)
		}

		// 检查ACL是否允许访问
		if !isACLAllowed(aclPerms) {
			return false, fmt.Errorf("ACL权限不足")
		}
	}

	return true, nil
}

// hasSELinux 检查系统是否启用了SELinux
func hasSELinux() bool {
	_, err := os.Stat("/sys/fs/selinux")
	return err == nil
}

// getSELinuxContext 获取文件的SELinux上下文
func getSELinuxContext(filePath string) (string, error) {
	// 使用getenforce命令检查SELinux状态
	cmd := exec.Command("getenforce")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("无法检查SELinux状态: %v", err)
	}

	// 如果SELinux是Enforcing状态，获取文件上下文
	if strings.TrimSpace(string(output)) == "Enforcing" {
		cmd := exec.Command("ls", "-Z", filePath)
		output, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("无法获取SELinux上下文: %v", err)
		}

		// 解析SELinux上下文
		parts := strings.Fields(string(output))
		if len(parts) >= 4 {
			return parts[3], nil
		}
	}

	return "", nil
}

// isSELinuxAllowed 检查SELinux上下文是否允许访问
func isSELinuxAllowed(context string) bool {
	// 简单的SELinux上下文检查
	// 在实际应用中，这里应该使用更复杂的SELinux策略检查
	if context == "" {
		return true // 没有SELinux上下文，允许访问
	}

	// 检查常见的受限上下文
	restrictedContexts := []string{
		"unconfined_u:object_r:admin_home_t:s0",
		"system_u:object_r:shadow_t:s0",
		"system_u:object_r:etc_t:s0",
	}

	for _, restricted := range restrictedContexts {
		if context == restricted {
			return false
		}
	}

	return true
}

// hasACL 检查系统是否支持ACL
func hasACL() bool {
	_, err := exec.LookPath("getfacl")
	return err == nil
}

// getACLPermissions 获取文件的ACL权限
func getACLPermissions(filePath string) (string, error) {
	cmd := exec.Command("getfacl", filePath)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("无法获取ACL权限: %v", err)
	}

	return string(output), nil
}

// isACLAllowed 检查ACL权限是否允许访问
func isACLAllowed(aclPerms string) bool {
	// 简单的ACL权限检查
	if aclPerms == "" {
		return true // 没有ACL权限，使用标准权限
	}

	// 检查ACL中是否有拒绝访问的条目
	lines := strings.Split(aclPerms, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "user::") && strings.Contains(line, "---") {
			return false // 用户没有权限
		}
		if strings.Contains(line, "group::") && strings.Contains(line, "---") {
			return false // 组没有权限
		}
		if strings.Contains(line, "other::") && strings.Contains(line, "---") {
			return false // 其他人没有权限
		}
	}

	return true
}
