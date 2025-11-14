using System.IO;

namespace DelGuard
{
    internal static class AttributeHelper
    {
        public static void MakeWritable(string path)
        {
            if (File.Exists(path))
            {
                var attr = File.GetAttributes(path);
                if (attr.HasFlag(FileAttributes.ReadOnly))
                {
                    File.SetAttributes(path, attr & ~FileAttributes.ReadOnly);
                }
                return;
            }

            if (Directory.Exists(path))
            {
                ClearDirectoryAttributes(path);
            }
        }

        private static void ClearDirectoryAttributes(string dir)
        {
            var attr = File.GetAttributes(dir);
            if (attr.HasFlag(FileAttributes.ReadOnly))
            {
                File.SetAttributes(dir, attr & ~FileAttributes.ReadOnly);
            }
            foreach (var f in Directory.EnumerateFileSystemEntries(dir, "*", SearchOption.AllDirectories))
            {
                try
                {
                    var a = File.GetAttributes(f);
                    if (a.HasFlag(FileAttributes.ReadOnly))
                    {
                        File.SetAttributes(f, a & ~FileAttributes.ReadOnly);
                    }
                }
                catch { }
            }
        }
    }
}