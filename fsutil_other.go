//go:build !linux
// +build !linux

package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// isEXDEV 判断是否跨设备移动错误（非 Linux 平台用于编译满足，返回尽量合理结果）
func isEXDEV(err error) bool {
	var errno syscall.Errno
	if errors.As(err, &errno) {
		// 在大多数平台 EXDEV 常量存在
		return errno == syscall.EXDEV
	}
	// 兜底：匹配常见文案
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "cross-device") || strings.Contains(msg, "different device")
}

func removeOriginal(p string) error {
	fi, err := os.Lstat(p)
	if err != nil {
		return err
	}
	if fi.IsDir() {
		return os.RemoveAll(p)
	}
	return os.Remove(p)
}

func copyTree(src, dst string) error {
	fi, err := os.Lstat(src)
	if err != nil {
		return err
	}

	switch {
	case fi.Mode()&os.ModeSymlink != 0:
		return copySymlink(src, dst)
	case fi.IsDir():
		if err := os.MkdirAll(dst, fi.Mode().Perm()); err != nil {
			return err
		}
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		for _, e := range entries {
			sChild := filepath.Join(src, e.Name())
			dChild := filepath.Join(dst, e.Name())
			if err := copyTree(sChild, dChild); err != nil {
				return err
			}
		}
		return nil
	default:
		return copyFile(src, dst, fi)
	}
}

func copyFile(src, dst string, fi os.FileInfo) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fi.Mode().Perm())
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

func copySymlink(src, dst string) error {
	target, err := os.Readlink(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	// 在 Windows 上可能需要管理员/开发者模式才可创建符号链接；此路径在非 Linux 平台仅为编译满足，正常不会被调用
	return os.Symlink(target, dst)
}
