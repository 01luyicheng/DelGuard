using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using Microsoft.VisualBasic.FileIO;

namespace DelGuard
{
    internal static class Program
    {
        private static int Main(string[] args)
        {
            if (args.Length == 0 || args.Contains("--help") || args.Contains("-h"))
            {
                PrintHelp();
                return 0;
            }

            var options = CliOptions.Parse(args);

            var config = AppConfig.Load();
            var logger = new Logger(options.Silent ? LogLevel.None : (options.Verbose ? LogLevel.Debug : config.LogLevel));

            var targets = PathExpander.Expand(options.Paths, logger, options.Recursive);
            if (targets.Count == 0)
            {
                logger.Error("DG001", "未匹配到任何文件或目录");
                return 1;
            }

            // 对于递归删除，我们需要对目标进行排序，确保子项在父项之前被列出
            // 这样可以避免在删除父目录后尝试删除子目录的问题
            if (options.Recursive)
            {
                targets = targets.OrderByDescending(t => t.Length).ToList();
            }

            var deleter = new Deleter(logger, config);

            int failures = 0;
            foreach (var t in targets)
            {
                try
                {
                    if (options.Preview)
                    {
                        logger.Info($"预览：{t}");
                        continue;
                    }

                    if (options.Force)
                    {
                        AttributeHelper.MakeWritable(t);
                    }

                    deleter.SendToRecycleBin(t);
                    logger.Info($"已移动到回收站：{t}");
                }
                catch (UnauthorizedAccessException ua)
                {
                    logger.Error("DG002", $"权限不足：{t}：{ua.Message}");
                    failures++;
                }
                catch (FileNotFoundException)
                {
                    logger.Error("DG001", $"路径不存在：{t}");
                    failures++;
                }
                catch (DirectoryNotFoundException)
                {
                    logger.Error("DG001", $"路径不存在：{t}");
                    failures++;
                }
                catch (Exception ex)
                {
                    logger.Error("DG003", $"未知错误：{t}：{ex.Message}");
                    failures++;
                }
            }

            if (failures > 0)
            {
                logger.Warn($"完成，失败 {failures} 项");
                return 2;
            }

            return 0;
        }

        private static void PrintHelp()
        {
            Console.WriteLine("DelGuard - Windows文件删除工具\n");
            Console.WriteLine("用法：");
            Console.WriteLine("  delguard [选项] <路径...>\n");
            Console.WriteLine("选项：");
            Console.WriteLine("  -f, --force       强制删除（绕过系统限制删除被占用的文件）");
            Console.WriteLine("  -p, --preview     预览模式，先显示将要删除的文件");
            Console.WriteLine("  -r, --recursive   递归删除文件夹及其所有内容");
            Console.WriteLine("  -s, --silent      静默模式，最小化输出");
            Console.WriteLine("  -v, --verbose     详细模式，显示详细日志");
            Console.WriteLine("  -h, --help        显示帮助信息\n");
            Console.WriteLine("示例：");
            Console.WriteLine("  delguard -r C:\\Windows\\Temp\\*");
            Console.WriteLine("  delguard -f C:\\Windows\\Logs\\error.log");
            Console.WriteLine("  delguard -p -r C:\\Users\\%USERNAME%\\Downloads\\old_files");
            Console.WriteLine("  delguard -s -f C:\\temp\\*.tmp\n");
            Console.WriteLine("说明：所有匹配到的文件或目录将移动到回收站，而非永久删除。");
        }
    }
}