package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// isStdinInteractive 检查标准输入是否为交互式
func isStdinInteractive() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// readLineWithTimeout 带超时的读取输入
func readLineWithTimeout(timeout time.Duration) (string, bool) {
	done := make(chan string, 1)
	go func() {
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			done <- ""
		} else {
			done <- strings.TrimSpace(line)
		}
	}()

	select {
	case line := <-done:
		return line, true
	case <-time.After(timeout):
		fmt.Println("\n超时，默认取消操作")
		return "", false
	}
}
