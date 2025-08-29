package ui

import (
	"fmt"

	"github.com/01luyicheng/DelGuard/internal/config"
	"github.com/01luyicheng/DelGuard/internal/utils/version"
)

// Interface 用户界面
type Interface struct {
	config *config.Config
}

// NewInterface 创建新的用户界面
func NewInterface(cfg *config.Config) *Interface {
	return &Interface{
		config: cfg,
	}
}

// ShowHelp 显示帮助信息
func (ui *Interface) ShowHelp() {
	fmt.Println("DelGuard - 安全文件删除和恢复工具")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  delguard <命令> [选项] [参数...]")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  delete, del, rm    删除文件或目录")
	fmt.Println("  search, find       搜索文件")
	fmt.Println("  restore            恢复已删除的文件")
	fmt.Println("  config             配置管理")
	fmt.Println("  help               显示帮助信息")
	fmt.Println("  version            显示版本信息")
	fmt.Println()
	fmt.Println("删除选项:")
	fmt.Println("  -f, --force        强制删除，跳过安全检查")
	fmt.Println("  -r, --recursive    递归删除目录")
	fmt.Println()
	fmt.Println("搜索选项:")
	fmt.Println("  -n, --name         按文件名搜索")
	fmt.Println("  -s, --size         按文件大小搜索")
	fmt.Println("  -t, --type         按文件类型搜索")
	fmt.Println("  -r, --recursive    递归搜索")
	fmt.Println()
	fmt.Println("恢复选项:")
	fmt.Println("  -l, --list         仅列出可恢复的文件")
	fmt.Println("  -i, --interactive  交互式恢复")
	fmt.Println("  --max <数量>       限制显示的文件数量")
	fmt.Println()
	fmt.Println("配置选项:")
	fmt.Println("  show               显示当前配置")
	fmt.Println("  set <key> <value>  设置配置项")
	fmt.Println("  reset              重置为默认配置")
	fmt.Println()
	fmt.Println("全局选项:")
	fmt.Println("  --config <文件>    指定配置文件")
	fmt.Println("  -h, --help         显示帮助信息")
	fmt.Println("  -v, --version      显示版本信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  delguard delete file.txt")
	fmt.Println("  delguard delete -rf directory/")
	fmt.Println("  delguard search -n \"*.log\"")
	fmt.Println("  delguard search -s \">1MB\"")
	fmt.Println("  delguard restore -l")
	fmt.Println("  delguard restore -i \"*.txt\"")
	fmt.Println("  delguard config show")
}

// ShowVersion 显示版本信息
func (ui *Interface) ShowVersion() {
	version.PrintVersion()
}

// ShowError 显示错误信息
func (ui *Interface) ShowError(err error) {
	fmt.Printf("❌ 错误: %v\n", err)
}

// ShowSuccess 显示成功信息
func (ui *Interface) ShowSuccess(message string) {
	fmt.Printf("✅ %s\n", message)
}

// ShowWarning 显示警告信息
func (ui *Interface) ShowWarning(message string) {
	fmt.Printf("⚠️  %s\n", message)
}

// ShowInfo 显示信息
func (ui *Interface) ShowInfo(message string) {
	fmt.Printf("ℹ️  %s\n", message)
}

// Confirm 确认操作
func (ui *Interface) Confirm(message string) bool {
	fmt.Printf("%s (y/N): ", message)
	var response string
	fmt.Scanln(&response)
	return response == "y" || response == "Y" || response == "yes" || response == "Yes"
}
