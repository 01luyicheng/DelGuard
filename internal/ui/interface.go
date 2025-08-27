package ui

import (
	"fmt"
)

// Interface 用户界面
type Interface struct {
	config interface{}
}

// NewInterface 创建用户界面
func NewInterface(config interface{}) *Interface {
	return &Interface{
		config: config,
	}
}

// ShowHelp 显示帮助信息
func (ui *Interface) ShowHelp() {
	fmt.Println(`DelGuard - 安全文件删除工具

用法:
  delguard <命令> [选项] [文件...]

命令:
  delete, del, rm    删除文件或目录
  search, find       搜索文件
  restore           恢复已删除的文件
  config            配置管理
  help              显示帮助信息
  version           显示版本信息

选项:
  -f, --force       强制删除，跳过确认
  -r, --recursive   递归删除目录
  -v, --verbose     详细输出
  --dry-run         预览模式，不实际删除

示例:
  delguard delete file.txt
  delguard search "*.log"
  delguard config show`)
}

// ShowVersion 显示版本信息
func (ui *Interface) ShowVersion() {
	fmt.Println("DelGuard v2.0.0")
	fmt.Println("安全文件删除工具")
}
