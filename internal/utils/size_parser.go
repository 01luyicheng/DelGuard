package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ParseSize 解析文件大小字符串，返回字节数
// 支持单位: B, KB, MB, GB, TB (不区分大小写)
// 示例: "1GB", "512MB", "2.5TB"
func ParseSize(sizeStr string) (int64, error) {
	// 去除空格
	sizeStr = strings.TrimSpace(sizeStr)
	if sizeStr == "" {
		return 0, fmt.Errorf("空的大小字符串")
	}

	// 正则匹配数字和单位
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*([KMGT]?B?)$`)
	matches := re.FindStringSubmatch(strings.ToUpper(sizeStr))
	
	if len(matches) != 3 {
		return 0, fmt.Errorf("无效的大小格式: %s", sizeStr)
	}

	// 解析数值
	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, fmt.Errorf("解析数值失败: %v", err)
	}

	// 解析单位
	unit := matches[2]
	var multiplier float64
	
	switch unit {
	case "B", "":
		multiplier = 1
	case "KB":
		multiplier = 1024
	case "MB":
		multiplier = 1024 * 1024
	case "GB":
		multiplier = 1024 * 1024 * 1024
	case "TB":
		multiplier = 1024 * 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("不支持的单位: %s", unit)
	}

	return int64(value * multiplier), nil
}

// FormatSize 将字节数格式化为易读的字符串
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// MustParseSize 解析文件大小字符串，如果解析失败则panic
// 用于配置初始化时
func MustParseSize(sizeStr string) int64 {
	size, err := ParseSize(sizeStr)
	if err != nil {
		panic(fmt.Sprintf("解析文件大小失败: %v", err))
	}
	return size
}