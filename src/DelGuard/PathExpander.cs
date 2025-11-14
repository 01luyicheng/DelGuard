using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;

namespace DelGuard
{
    public static class PathExpander
    {
        public static List<string> Expand(IEnumerable<string> inputs, Logger logger, bool recursive = false)
        {
            var list = new List<string>();
            foreach (var input in inputs)
            {
                if (input.IndexOfAny(new[] { '*', '?' }) >= 0)
                {
                    var baseDir = Path.GetDirectoryName(input);
                    var pattern = Path.GetFileName(input);
                    if (string.IsNullOrEmpty(baseDir)) baseDir = Directory.GetCurrentDirectory();
                    if (!Directory.Exists(baseDir))
                    {
                        logger.Error("DG001", $"目录不存在：{baseDir}");
                        continue;
                    }
                    var searchOption = recursive ? SearchOption.AllDirectories : SearchOption.TopDirectoryOnly;
                    var matches = Directory.EnumerateFileSystemEntries(baseDir, pattern, searchOption);
                    list.AddRange(matches);
                }
                else
                {
                    // 对于非通配符路径，如果是目录且启用了递归，则添加目录及其所有内容
                    if (recursive && Directory.Exists(input))
                    {
                        // 先添加目录本身
                        list.Add(input);
                        // 然后添加目录下的所有文件和子目录
                        try
                        {
                            var allEntries = Directory.EnumerateFileSystemEntries(input, "*", SearchOption.AllDirectories);
                            list.AddRange(allEntries);
                        }
                        catch (UnauthorizedAccessException)
                        {
                            logger.Warn($"无权访问目录部分内容：{input}");
                            list.Add(input);
                        }
                    }
                    else
                    {
                        list.Add(input);
                    }
                }
            }

            return list.Distinct(StringComparer.OrdinalIgnoreCase).ToList();
        }
    }
}