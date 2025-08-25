package main

// 文件大小常量
const (
	KB = 1024
	MB = KB * 1024
	GB = MB * 1024
	TB = GB * 1024
)

// 文件大小限制常量
const (
	DefaultMaxFileSize   = 10 * GB  // 默认最大文件大小 10GB
	DefaultLargeFileSize = 100 * MB // 大文件阈值 100MB
	DefaultHugeFileSize  = 1 * GB   // 超大文件阈值 1GB
	DefaultBatchSize     = 100      // 默认批处理大小
	DefaultMaxResults    = 10       // 默认最大搜索结果数
	DefaultBufferSize    = 1024     // 默认缓冲区大小
)

// Windows系统限制常量
const (
	WindowsMaxPathLength     = 260   // Windows最大路径长度
	WindowsMaxFileNameLength = 255   // Windows最大文件名长度
	WindowsLongPathPrefix    = 32760 // Windows长路径前缀限制
)

// Unix系统限制常量
const (
	UnixMaxPathLength = 4096 // Unix最大路径长度
)

// 时间格式常量
const (
	TimeFormatStandard   = "2006-01-02 15:04:05"       // 标准时间格式
	TimeFormatWithMillis = "2006-01-02 15:04:05.000"   // 带毫秒的时间格式
	TimeFormatRFC3339    = "2006-01-02T15:04:05Z07:00" // RFC3339格式
	TimeFormatDate       = "2006-01-02"                // 日期格式
	TimeFormatTime       = "15:04:05"                  // 时间格式
)

// 确认字符串常量
const (
	ConfirmDeleteRecycleBin = "DELETE_RECYCLE_BIN" // 删除回收站确认
	ConfirmYesUnderstand    = "YES I UNDERSTAND"   // 理解确认
	ConfirmYes              = "YES"                // 是确认
	ConfirmDelete           = "DELETE"             // 删除确认
)

// 日志级别常量
const (
	LogLevelDebugStr = "debug" // 调试级别
	LogLevelInfoStr  = "info"  // 信息级别
	LogLevelWarnStr  = "warn"  // 警告级别
	LogLevelErrorStr = "error" // 错误级别
	LogLevelFatalStr = "fatal" // 致命级别
)

// 默认相似度阈值
const (
	DefaultSimilarityThreshold = 60.0 // 默认相似度阈值
)

// Windows驱动器字母常量
var WindowsDriveLetters = []string{
	"C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N",
	"O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
}

// 二进制文件扩展名
var BinaryFileExtensions = []string{
	".exe", ".dll", ".so", ".dylib", ".bin", ".obj", ".o", ".a", ".lib",
	".zip", ".tar", ".gz", ".jpg", ".png", ".gif", ".mp4", ".avi", ".mp3", ".wav",
}

// 脚本文件扩展名
var ScriptFileExtensions = []string{
	".sh", ".bat", ".cmd", ".ps1", ".py", ".pl", ".rb", ".js", ".vbs",
}

// 搜索相关常量
const (
	MaxSearchResults        = 10
	MaxSearchDepth          = 3
	HighSimilarityThreshold = 0.8
	ContentMatchSimilarity  = 95.0
	SubdirMatchSimilarity   = 0.7
	TruncateLineLength      = 100
)

// 默认智能搜索配置
var DefaultSmartSearchConfig = SmartSearchConfig{
	SimilarityThreshold: DefaultSimilarityThreshold,
	MaxResults:          MaxSearchResults,
	SearchContent:       true,
	Recursive:           true,
	SearchParent:        false,
}
