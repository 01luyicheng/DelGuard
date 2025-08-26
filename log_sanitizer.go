package main

import (
	"regexp"
	"strings"
)

// LogSanitizer 日志清理器，用于过滤敏感信息
type LogSanitizer struct {
	sensitivePatterns []*regexp.Regexp
	replacements      map[string]string
}

// NewLogSanitizer 创建日志清理器
func NewLogSanitizer() *LogSanitizer {
	sanitizer := &LogSanitizer{
		replacements: make(map[string]string),
	}

	// 初始化敏感信息模式
	sanitizer.initSensitivePatterns()

	return sanitizer
}

// initSensitivePatterns 初始化敏感信息匹配模式
func (ls *LogSanitizer) initSensitivePatterns() {
	patterns := []string{
		// 密码相关
		`(?i)password["\s]*[:=]["\s]*[^\s"]+`,
		`(?i)passwd["\s]*[:=]["\s]*[^\s"]+`,
		`(?i)pwd["\s]*[:=]["\s]*[^\s"]+`,

		// API密钥
		`(?i)api[_-]?key["\s]*[:=]["\s]*[^\s"]+`,
		`(?i)secret[_-]?key["\s]*[:=]["\s]*[^\s"]+`,
		`(?i)access[_-]?token["\s]*[:=]["\s]*[^\s"]+`,

		// 数据库连接字符串
		`(?i)connection[_-]?string["\s]*[:=]["\s]*[^\s"]+`,
		`(?i)database[_-]?url["\s]*[:=]["\s]*[^\s"]+`,

		// 邮箱地址（部分隐藏）
		`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`,

		// 电话号码
		`\b\d{3}-\d{3}-\d{4}\b`,
		`\b\d{11}\b`,

		// IP地址（内网地址保留，公网地址隐藏）
		`\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`,

		// 文件路径中的用户名
		`(?i)\\users\\[^\\]+\\`,
		`(?i)/home/[^/]+/`,
		`(?i)/Users/[^/]+/`,

		// Windows SID
		`S-1-[0-9-]+`,

		// 信用卡号（简单匹配）
		`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`,

		// 身份证号
		`\b\d{17}[\dXx]\b`,

		// JWT Token
		`eyJ[A-Za-z0-9_-]*\.eyJ[A-Za-z0-9_-]*\.[A-Za-z0-9_-]*`,

		// 哈希值（可能是密码哈希）
		`\b[a-fA-F0-9]{32}\b`, // MD5
		`\b[a-fA-F0-9]{40}\b`, // SHA1
		`\b[a-fA-F0-9]{64}\b`, // SHA256
	}

	for _, pattern := range patterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			ls.sensitivePatterns = append(ls.sensitivePatterns, regex)
		}
	}
}

// SanitizeMessage 清理日志消息中的敏感信息
func (ls *LogSanitizer) SanitizeMessage(message string) string {
	if message == "" {
		return message
	}

	sanitized := message

	// 应用所有敏感信息模式
	for _, pattern := range ls.sensitivePatterns {
		sanitized = pattern.ReplaceAllStringFunc(sanitized, ls.maskSensitiveData)
	}

	// 应用自定义替换规则
	for original, replacement := range ls.replacements {
		sanitized = strings.ReplaceAll(sanitized, original, replacement)
	}

	return sanitized
}

// maskSensitiveData 遮蔽敏感数据
func (ls *LogSanitizer) maskSensitiveData(match string) string {
	// 根据匹配内容的类型进行不同的处理
	lower := strings.ToLower(match)

	// 密码相关 - 完全隐藏
	if strings.Contains(lower, "password") || strings.Contains(lower, "passwd") || strings.Contains(lower, "pwd") {
		return ls.replaceWithMask(match, "password", "***")
	}

	// API密钥 - 显示前几位
	if strings.Contains(lower, "key") || strings.Contains(lower, "token") {
		return ls.replaceWithMask(match, "key", "***")
	}

	// 邮箱地址 - 部分隐藏
	if strings.Contains(match, "@") {
		return ls.maskEmail(match)
	}

	// IP地址 - 检查是否为内网地址
	if ls.isIPAddress(match) {
		if ls.isPrivateIP(match) {
			return match // 内网地址保留
		}
		return ls.maskIP(match)
	}

	// 文件路径 - 隐藏用户名
	if strings.Contains(lower, "users") || strings.Contains(lower, "home") {
		return ls.maskUserPath(match)
	}

	// 默认处理 - 部分隐藏
	return ls.partialMask(match)
}

// replaceWithMask 用掩码替换敏感部分
func (ls *LogSanitizer) replaceWithMask(original, keyword, mask string) string {
	parts := strings.Split(original, ":")
	if len(parts) >= 2 {
		return parts[0] + ": " + mask
	}
	parts = strings.Split(original, "=")
	if len(parts) >= 2 {
		return parts[0] + "=" + mask
	}
	return mask
}

// maskEmail 遮蔽邮箱地址
func (ls *LogSanitizer) maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***@***.***"
	}

	username := parts[0]
	domain := parts[1]

	// 保留用户名的前1-2个字符
	maskedUsername := ""
	if len(username) <= 2 {
		maskedUsername = "*"
	} else if len(username) <= 4 {
		maskedUsername = username[:1] + "***"
	} else {
		maskedUsername = username[:2] + "***"
	}

	// 域名部分保留
	return maskedUsername + "@" + domain
}

// maskIP 遮蔽IP地址
func (ls *LogSanitizer) maskIP(ip string) string {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return "***.***.***.***.***"
	}

	// 保留前两段，隐藏后两段
	return parts[0] + "." + parts[1] + ".***.***.***"
}

// maskUserPath 遮蔽用户路径
func (ls *LogSanitizer) maskUserPath(path string) string {
	// Windows路径
	if strings.Contains(strings.ToLower(path), "\\users\\") {
		re := regexp.MustCompile(`(?i)\\users\\[^\\]+\\`)
		return re.ReplaceAllString(path, "\\users\\***\\")
	}

	// Unix路径
	if strings.Contains(path, "/home/") {
		re := regexp.MustCompile(`/home/[^/]+/`)
		return re.ReplaceAllString(path, "/home/***/")
	}

	if strings.Contains(path, "/Users/") {
		re := regexp.MustCompile(`/Users/[^/]+/`)
		return re.ReplaceAllString(path, "/Users/***/")
	}

	return path
}

// partialMask 部分遮蔽
func (ls *LogSanitizer) partialMask(text string) string {
	if len(text) <= 4 {
		return "***"
	}

	if len(text) <= 8 {
		return text[:2] + "***"
	}

	return text[:3] + "***" + text[len(text)-2:]
}

// isIPAddress 检查是否为IP地址
func (ls *LogSanitizer) isIPAddress(text string) bool {
	parts := strings.Split(text, ".")
	if len(parts) != 4 {
		return false
	}

	for _, part := range parts {
		if len(part) == 0 || len(part) > 3 {
			return false
		}
		for _, char := range part {
			if char < '0' || char > '9' {
				return false
			}
		}
	}

	return true
}

// isPrivateIP 检查是否为内网IP
func (ls *LogSanitizer) isPrivateIP(ip string) bool {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}

	// 10.0.0.0/8
	if parts[0] == "10" {
		return true
	}

	// 172.16.0.0/12
	if parts[0] == "172" {
		if len(parts[1]) > 0 {
			second := parts[1]
			if second >= "16" && second <= "31" {
				return true
			}
		}
	}

	// 192.168.0.0/16
	if parts[0] == "192" && parts[1] == "168" {
		return true
	}

	// 127.0.0.0/8 (localhost)
	if parts[0] == "127" {
		return true
	}

	return false
}

// AddCustomReplacement 添加自定义替换规则
func (ls *LogSanitizer) AddCustomReplacement(original, replacement string) {
	ls.replacements[original] = replacement
}

// RemoveCustomReplacement 移除自定义替换规则
func (ls *LogSanitizer) RemoveCustomReplacement(original string) {
	delete(ls.replacements, original)
}

// SanitizeFilePath 清理文件路径中的敏感信息
func (ls *LogSanitizer) SanitizeFilePath(path string) string {
	if path == "" {
		return path
	}

	// 使用通用的消息清理方法
	return ls.SanitizeMessage(path)
}

// SanitizeErrorMessage 清理错误消息中的敏感信息
func (ls *LogSanitizer) SanitizeErrorMessage(errorMsg string) string {
	if errorMsg == "" {
		return errorMsg
	}

	// 使用通用的消息清理方法
	return ls.SanitizeMessage(errorMsg)
}
