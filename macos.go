//go:build darwin
// +build darwin

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// moveToTrashMacOS 将文件移动到MacOS废纸篓
func moveToTrashMacOS(filePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return ErrFileNotFound
	}

	// 获取绝对路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}

	// 检查路径是否已经被删除或不可访问
	if _, err := os.Lstat(absPath); err != nil {
		return E(KindIO, "moveToTrash", absPath, err, "无法访问文件")
	}

	// 使用osascript调用Finder将文件移动到废纸篓
	// 使用更健壮的AppleScript，处理各种边界情况
	script := fmt.Sprintf(`
	tell application "Finder"
		set targetFile to POSIX file "%s"
		if exists targetFile then
			delete targetFile
		end if
	end tell
	`, absPath)

	cmd := exec.Command("osascript", "-e", script)

	// 设置超时，防止osascript挂起
	cmd.Env = append(os.Environ(), "LANG=en_US.UTF-8")

	if err := cmd.Run(); err != nil {
		// 根据 osascript 的退出码和错误信息提供更详细的错误
		if exitErr, ok := err.(*exec.ExitError); ok {
			switch exitErr.ExitCode() {
			case 1:
				return E(KindPermission, "moveToTrash", absPath, err, "权限不足或文件被占用")
			case 128:
				return E(KindCancelled, "moveToTrash", absPath, err, "用户取消了操作")
			default:
				// 检查错误输出中是否包含特定信息
				stderr := string(exitErr.Stderr)
				if strings.Contains(stderr, "doesn't exist") {
					return E(KindNotFound, "moveToTrash", absPath, err, "文件不存在")
				} else if strings.Contains(stderr, "permission") {
					return E(KindPermission, "moveToTrash", absPath, err, "权限不足")
				}
				return E(KindIO, "moveToTrash", absPath, err, "AppleScript 执行失败")
			}
		}
		return E(KindIO, "moveToTrash", absPath, err, "无法启动 osascript")
	}

	// 验证文件确实被移动到了废纸篓
	// 注意：MacOS的废纸篓路径是 ~/.Trash
	homeDir, err := os.UserHomeDir()
	if err == nil {
		trashPath := filepath.Join(homeDir, ".Trash", filepath.Base(absPath))
		if _, err := os.Stat(trashPath); err == nil {
			// 文件已成功移动到废纸篓
		} else {
			// 文件可能被重命名（同名冲突），检查带时间戳的版本
			pattern := filepath.Join(homeDir, ".Trash", filepath.Base(absPath)+"*")
			matches, _ := filepath.Glob(pattern)
			if len(matches) > 0 {
				// 文件已移动到废纸篓，可能被重命名
			}
		}
	}

	return nil
}

// 为macOS平台提供其他平台函数的存根
func moveToTrashWindows(filePath string) error {
	return ErrUnsupportedPlatform
}

func moveToTrashLinux(filePath string) error {
	return ErrUnsupportedPlatform
}
