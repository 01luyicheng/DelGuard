using System;
using System.IO;

namespace DelGuard
{
    public enum LogLevel
    {
        None = 0,
        Error = 1,
        Warn = 2,
        Info = 3,
        Debug = 4
    }

    public sealed class Logger
    {
        private readonly LogLevel level;
        private readonly string? filePath;

        public Logger(LogLevel level)
        {
            this.level = level;
            try
            {
                var dir = Path.Combine(AppContext.BaseDirectory, "logs");
                Directory.CreateDirectory(dir);
                filePath = Path.Combine(dir, "delguard.log");
            }
            catch
            {
                filePath = null;
            }
        }

        public void Debug(string message)
        {
            if (level < LogLevel.Debug) return;
            Write("DEBUG", message);
        }

        public void Info(string message)
        {
            if (level < LogLevel.Info) return;
            Write("INFO", message);
        }

        public void Warn(string message)
        {
            if (level < LogLevel.Warn) return;
            Write("WARN", message);
        }

        public void Error(string code, string message)
        {
            if (level < LogLevel.Error) return;
            Write("ERROR", $"{code} {message}");
        }

        private void Write(string lvl, string message)
        {
            var line = $"[{DateTime.Now:yyyy-MM-dd HH:mm:ss}] {lvl} {message}";
            Console.WriteLine(line);
            if (filePath != null)
            {
                try { File.AppendAllText(filePath, line + Environment.NewLine); } catch { }
            }
        }
    }
}