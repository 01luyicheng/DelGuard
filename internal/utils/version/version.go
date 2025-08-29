package version

import (
	"fmt"
	"runtime"
)

const (
	Version   = "1.3.0"
	BuildDate = "2025-08-29"
	GitCommit = "dev"
)

// PrintVersion 打印版本信息
func PrintVersion() {
	fmt.Printf("DelGuard v%s\n", Version)
	fmt.Printf("构建日期: %s\n", BuildDate)
	fmt.Printf("Git提交: %s\n", GitCommit)
	fmt.Printf("Go版本: %s\n", runtime.Version())
	fmt.Printf("操作系统: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

// GetVersion 获取版本号
func GetVersion() string {
	return Version
}

// GetBuildInfo 获取构建信息
func GetBuildInfo() map[string]string {
	return map[string]string{
		"version":   Version,
		"buildDate": BuildDate,
		"gitCommit": GitCommit,
		"goVersion": runtime.Version(),
		"os":        runtime.GOOS,
		"arch":      runtime.GOARCH,
	}
}
