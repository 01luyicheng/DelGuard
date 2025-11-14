using System;
using System.IO;
using System.Text.Json;
using Microsoft.VisualBasic.FileIO;

namespace DelGuard
{
    internal sealed class AppConfig
    {
        public UIOption UiOption { get; private set; } = UIOption.OnlyErrorDialogs;
        public LogLevel LogLevel { get; private set; } = LogLevel.Info;

        public static AppConfig Load()
        {
            var cfgPath = Path.Combine(AppContext.BaseDirectory, "config", "windows.json");
            if (!File.Exists(cfgPath))
            {
                return new AppConfig();
            }

            try
            {
                var json = File.ReadAllText(cfgPath);
                var dto = JsonSerializer.Deserialize<AppConfigDto>(json);
                if (dto == null) return new AppConfig();

                var cfg = new AppConfig();
                cfg.UiOption = dto.uiOption switch
                {
                    "AllDialogs" => UIOption.AllDialogs,
                    "NoDialogs" => UIOption.OnlyErrorDialogs, // 映射到仅显示错误对话框
                    _ => UIOption.OnlyErrorDialogs
                };
                cfg.LogLevel = dto.logLevel switch
                {
                    "debug" => LogLevel.Debug,
                    "warn" => LogLevel.Warn,
                    "error" => LogLevel.Error,
                    "none" => LogLevel.None,
                    _ => LogLevel.Info
                };
                return cfg;
            }
            catch
            {
                return new AppConfig();
            }
        }

        private sealed class AppConfigDto
        {
            public string uiOption { get; set; } = "OnlyErrorDialogs";
            public string logLevel { get; set; } = "info";
        }
    }
}