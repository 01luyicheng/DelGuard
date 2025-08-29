package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"delguard/internal/filesystem"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// listCmd 列表命令
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls", "show"},
	Short:   "查看回收站中的文件",
	Long: `查看回收站中的文件和目录列表。

显示文件名、大小、删除时间等信息。
支持按不同条件排序和过滤。

示例:
  delguard list
  delguard list --sort=size
  delguard list --filter="*.txt"
  delguard ls  # 别名`,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)

	// 添加标志
	listCmd.Flags().StringP("sort", "s", "time", "排序方式: name, size, time")
	listCmd.Flags().BoolP("reverse", "r", false, "反向排序")
	listCmd.Flags().StringP("filter", "f", "", "文件名过滤器（支持通配符）")
	listCmd.Flags().BoolP("long", "l", false, "详细列表格式")
	listCmd.Flags().Bool("human", true, "人类可读的文件大小格式")
	listCmd.Flags().IntP("limit", "n", 0, "限制显示的文件数量（0表示无限制）")
}

func runList(cmd *cobra.Command, args []string) error {
	// 获取标志值
	sortBy, _ := cmd.Flags().GetString("sort")
	reverse, _ := cmd.Flags().GetBool("reverse")
	filter, _ := cmd.Flags().GetString("filter")
	longFormat, _ := cmd.Flags().GetBool("long")
	humanReadable, _ := cmd.Flags().GetBool("human")
	limit, _ := cmd.Flags().GetInt("limit")
	quiet := viper.GetBool("quiet")

	// 获取回收站管理器
	manager, err := filesystem.GetTrashManager()
	if err != nil {
		return fmt.Errorf("初始化回收站管理器失败: %v", err)
	}

	// 获取回收站文件列表
	trashFiles, err := manager.ListTrashFiles()
	if err != nil {
		return fmt.Errorf("获取回收站文件列表失败: %v", err)
	}

	if len(trashFiles) == 0 {
		if !quiet {
			fmt.Println("🗑️  回收站是空的")
		}
		return nil
	}

	// 应用过滤器
	if filter != "" {
		var filteredFiles []filesystem.TrashFile
		for _, file := range trashFiles {
			matched, err := matchPattern(file.Name, filter)
			if err != nil {
				return fmt.Errorf("过滤器模式错误: %v", err)
			}
			if matched {
				filteredFiles = append(filteredFiles, file)
			}
		}
		trashFiles = filteredFiles
	}

	// 排序
	sortTrashFiles(trashFiles, sortBy, reverse)

	// 限制数量
	if limit > 0 && len(trashFiles) > limit {
		trashFiles = trashFiles[:limit]
	}

	// 显示文件列表
	if longFormat {
		displayLongFormat(trashFiles, humanReadable)
	} else {
		displayShortFormat(trashFiles, humanReadable)
	}

	// 显示统计信息
	if !quiet {
		totalSize := int64(0)
		for _, file := range trashFiles {
			totalSize += file.Size
		}
		fmt.Printf("\n📊 总计: %d 个项目", len(trashFiles))
		if humanReadable {
			fmt.Printf(", %s", filesystem.FormatFileSize(totalSize))
		} else {
			fmt.Printf(", %d 字节", totalSize)
		}
		fmt.Println()
	}

	return nil
}

// sortTrashFiles 排序回收站文件
func sortTrashFiles(files []filesystem.TrashFile, sortBy string, reverse bool) {
	sort.Slice(files, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "name":
			less = strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
		case "size":
			less = files[i].Size < files[j].Size
		case "time":
			less = files[i].DeletedTime.Before(files[j].DeletedTime)
		default:
			less = files[i].DeletedTime.Before(files[j].DeletedTime)
		}

		if reverse {
			return !less
		}
		return less
	})
}

// matchPattern 匹配文件名模式
func matchPattern(name, pattern string) (bool, error) {
	// 简单的通配符匹配
	if pattern == "" {
		return true, nil
	}

	// 转换为小写进行匹配
	name = strings.ToLower(name)
	pattern = strings.ToLower(pattern)

	// 简单的 * 通配符支持
	if strings.Contains(pattern, "*") {
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			prefix := parts[0]
			suffix := parts[1]
			return strings.HasPrefix(name, prefix) && strings.HasSuffix(name, suffix), nil
		}
	}

	// 精确匹配或包含匹配
	return strings.Contains(name, pattern), nil
}

// displayLongFormat 显示详细格式
func displayLongFormat(files []filesystem.TrashFile, humanReadable bool) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// 表头
	fmt.Fprintln(w, "类型\t名称\t大小\t删除时间\t原始路径")
	fmt.Fprintln(w, "----\t----\t----\t--------\t--------")

	for _, file := range files {
		// 文件类型图标
		typeIcon := "📄"
		if file.IsDirectory {
			typeIcon = "📁"
		}

		// 格式化大小
		var sizeStr string
		if humanReadable {
			sizeStr = filesystem.FormatFileSize(file.Size)
		} else {
			sizeStr = fmt.Sprintf("%d", file.Size)
		}

		// 格式化时间
		timeStr := file.DeletedTime.Format("2006-01-02 15:04:05")

		// 原始路径
		originalPath := file.OriginalPath
		if originalPath == "" {
			originalPath = "-"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			typeIcon, file.Name, sizeStr, timeStr, originalPath)
	}

	w.Flush()
}

// displayShortFormat 显示简短格式
func displayShortFormat(files []filesystem.TrashFile, humanReadable bool) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// 表头
	fmt.Fprintln(w, "名称\t大小\t删除时间")
	fmt.Fprintln(w, "----\t----\t--------")

	for _, file := range files {
		// 文件名（带图标）
		nameWithIcon := file.Name
		if file.IsDirectory {
			nameWithIcon = "📁 " + file.Name
		} else {
			nameWithIcon = "📄 " + file.Name
		}

		// 格式化大小
		var sizeStr string
		if humanReadable {
			sizeStr = filesystem.FormatFileSize(file.Size)
		} else {
			sizeStr = fmt.Sprintf("%d", file.Size)
		}

		// 格式化时间（相对时间）
		timeStr := formatRelativeTime(file.DeletedTime)

		fmt.Fprintf(w, "%s\t%s\t%s\n", nameWithIcon, sizeStr, timeStr)
	}

	w.Flush()
}

// formatRelativeTime 格式化相对时间
func formatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "刚刚"
	} else if diff < time.Hour {
		return fmt.Sprintf("%d分钟前", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%d小时前", int(diff.Hours()))
	} else if diff < 7*24*time.Hour {
		return fmt.Sprintf("%d天前", int(diff.Hours()/24))
	} else {
		return t.Format("2006-01-02")
	}
}
