package main

import (
	"time"
)

// 版本信息
const (
	Version = "1.0.0"
	AppName = "DelGuard"
)

// 搜索相关常量
const (
	// 相似度阈值
	ExactMatchSimilarity              = 100.0
	PrefixMatchSimilarity             = 90.0
	SuffixMatchSimilarity             = 80.0
	ContainsMatchSimilarity           = 70.0
	FuzzyMatchSimilarity              = 60.0
	SearchContentSimilarity           = 50.0
	SubdirMatchSimilarity             = 0.8  // 子目录匹配相似度权重
	EnhancedContentBaseSimilarity     = 40.0 // 增强内容搜索基础相似度
	EnhancedContentSimilarityPerMatch = 5.0  // 每个匹配增加的相似度

	// 默认配置
	DefaultSimilarityThreshold = 60.0
	DefaultMaxResults          = 20
	MaxSearchDepth             = 5   // 最大搜索深度
	MaxEnhancedMatches         = 10  // 最大增强匹配数
	TruncateContextLength      = 100 // 截断上下文长度

	// 分页相关
	PageSize = 10

	// 超时设置
	DefaultTimeout       = 30 * time.Second
	InteractiveTimeout   = 30 * time.Second
	FileOperationTimeout = 60 * time.Second

	// 文件大小限制
	MaxFileSize        = 10 * 1024 * 1024 * 1024 // 10GB
	DefaultMaxFileSize = 100 * 1024 * 1024       // 100MB 默认最大文件大小
	MaxBatchSize       = 1000                    // 最大批处理大小

	// 单位常量
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB

	// 路径相关
	MaxPathLength     = 4096
	MaxFilenameLength = 255

	// 并发相关
	DefaultMaxWorkers = 4
	MaxConcurrentOps  = 10

	// 日志相关
	MaxLogFileSize  = 100 * 1024 * 1024 // 100MB
	MaxLogBackups   = 5
	LogRotationDays = 30

	// 日志级别字符串
	LogLevelDebugStr = "DEBUG"
	LogLevelInfoStr  = "INFO"
	LogLevelWarnStr  = "WARN"
	LogLevelErrorStr = "ERROR"
	LogLevelFatalStr = "FATAL"

	// 时间格式
	TimeFormatWithMillis = "2006-01-02 15:04:05.000"
	TimeFormatStandard   = "2006-01-02 15:04:05"

	// 回收站确认字符串
	ConfirmYes              = "y"
	ConfirmYesUnderstand    = "yes"
	ConfirmDeleteRecycleBin = "delete_recycle_bin"
)

// 错误代码
const (
	ExitSuccess          = 0
	ExitGeneralError     = 1
	ExitFileNotFound     = 2
	ExitPermissionDenied = 3
	ExitInvalidArgument  = 4
	ExitUserCancelled    = 5
	ExitSystemError      = 6
	ExitConfigError      = 7
	ExitNetworkError     = 8
	ExitSecurityError    = 9
)

// 平台相关常量
const (
	WindowsPlatform = "windows"
	LinuxPlatform   = "linux"
	DarwinPlatform  = "darwin"
)

// 文件类型
const (
	FileTypeRegular   = "regular"
	FileTypeDirectory = "directory"
	FileTypeSymlink   = "symlink"
	FileTypeDevice    = "device"
	FileTypeSocket    = "socket"
	FileTypePipe      = "pipe"
	FileTypeUnknown   = "unknown"
)

// 操作类型
const (
	OpDelete  = "delete"
	OpCopy    = "copy"
	OpMove    = "move"
	OpRestore = "restore"
	OpList    = "list"
	OpSearch  = "search"
)

// 匹配类型
const (
	MatchExact    = "exact"
	MatchPrefix   = "prefix"
	MatchSuffix   = "suffix"
	MatchContains = "contains"
	MatchFuzzy    = "fuzzy"
	MatchContent  = "content"
	MatchRegex    = "regex"
)

// 安全级别
const (
	SecurityLevelLow    = "low"
	SecurityLevelMedium = "medium"
	SecurityLevelHigh   = "high"
	SecurityLevelStrict = "strict"
)

// 默认智能搜索配置
var DefaultSmartSearchConfig = SmartSearchConfig{
	SimilarityThreshold: DefaultSimilarityThreshold,
	MaxResults:          DefaultMaxResults,
	SearchContent:       false,
	SearchParent:        false,
}

// 配置文件名
var ConfigFileNames = []string{
	"config.json",
	"config.jsonc",
	"config.ini",
	"config.cfg",
	"config.conf",
	".env",
	"delguard.properties",
}

// 支持的语言代码
var SupportedLanguages = []string{
	"zh-CN", "zh-TW", "en-US", "ja", "ko-KR",
	"fr-FR", "de-DE", "es-ES", "it-IT", "pt-BR",
	"ru-RU", "ar-SA", "hi-IN", "th-TH", "vi-VN",
	"nl-NL", "sv-SE", "no-NO", "fi-FI",
}

// 危险路径模式 - 基础模式（Unix/Linux）
var DangerousPathPatterns = []string{
	"/",
	"/bin", "/sbin", "/usr", "/etc", "/var", "/sys", "/proc", "/dev",
}

// 获取平台特定的危险路径模式
func getPlatformSpecificDangerousPatterns() []string {
	paths := PathUtils.GetSystemPaths()
	var result []string
	for _, path := range paths {
		result = append(result, path)
	}
	return result
}

// 系统文件扩展名
var SystemFileExtensions = []string{
	".sys", ".dll", ".exe", ".com", ".bat", ".cmd",
	".msi", ".scr", ".pif", ".lnk", ".url",
}

// 可执行文件扩展名
var ExecutableExtensions = []string{
	".exe", ".com", ".bat", ".cmd", ".scr", ".pif",
	".sh", ".bash", ".zsh", ".fish", ".ps1", ".psm1",
	".py", ".pl", ".rb", ".js", ".vbs", ".jar",
}

// 文本文件扩展名（用于内容搜索）
var TextFileExtensions = []string{
	".txt", ".md", ".log", ".cfg", ".conf", ".ini",
	".json", ".xml", ".yaml", ".yml", ".csv", ".sql",
	".sh", ".bat", ".ps1", ".py", ".js", ".html", ".css",
	".go", ".java", ".c", ".cpp", ".h", ".hpp", ".cs",
	".php", ".rb", ".pl", ".swift", ".kt", ".rs", ".ts",
}
