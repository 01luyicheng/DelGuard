package main

import "unicode/utf8"

// truncateString 将字符串按最大长度安全截断（按rune计数），超出则在末尾追加"..."
// 参数:
//   - s: 原始字符串
//   - maxLen: 最大保留的字符数（按rune）
//
// 返回:
//   - string: 截断后的字符串
func truncateString(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	// 逐rune截断，确保不破坏utf-8
	count := 0
	endIdx := 0
	for i := range s {
		if count == maxLen {
			endIdx = i
			break
		}
		count++
	}
	if endIdx == 0 {
		endIdx = len(s)
	}
	return s[:endIdx] + "..."
}
