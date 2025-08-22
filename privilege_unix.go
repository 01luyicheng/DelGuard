//go:build !windows
// +build !windows

package main

import "os"

// IsElevated 返回是否以 root 运行（Unix/macOS/Linux）
func IsElevated() bool {
	return os.Geteuid() == 0
}
