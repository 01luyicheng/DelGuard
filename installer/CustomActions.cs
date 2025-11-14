using System;
using System.IO;
using Microsoft.Win32;

namespace DelGuard.Installer
{
    public class CustomActions
    {
        [System.Runtime.InteropServices.Guid("A2A6E373-1D9A-4A8E-8F6F-3ED6B5D6C6E8")]
        public static class ConfigureDelGuard
        {
            public static void ConfigureDelGuardSession()
            {
                try
                {
                    // 创建配置目录
                    string configPath = Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.ProgramFiles), "DelGuard", "config");
                    Directory.CreateDirectory(configPath);

                    // 创建配置文件
                    string configFilePath = Path.Combine(configPath, "windows.json");
                    if (!File.Exists(configFilePath))
                    {
                        string configContent = @"{
  ""uiOption"": ""OnlyErrorDialogs"",
  ""logLevel"": ""info""
}";
                        File.WriteAllText(configFilePath, configContent);
                    }

                    // 创建日志目录
                    string logPath = Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.ProgramFiles), "DelGuard", "logs");
                    Directory.CreateDirectory(logPath);
                }
                catch (Exception)
                {
                    // 自定义操作中静默处理错误
                    throw;
                }
            }
        }

        [System.Runtime.InteropServices.Guid("B3B7F484-2E0B-5B9F-9G7G-4FE7C6E7D7F9")]
        public static class UninstallDelGuard
        {
            public static void UninstallDelGuardSession()
            {
                try
                {
                    // 清理注册表项
                    Registry.LocalMachine.DeleteSubKeyTree(@"SOFTWARE\TREA\DelGuard", false);
                    
                    // 清理用户配置（可选）
                    string appDataPath = Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData);
                    string appConfigPath = Path.Combine(appDataPath, "DelGuard");
                    if (Directory.Exists(appConfigPath))
                    {
                        Directory.Delete(appConfigPath, true);
                    }
                }
                catch (Exception)
                {
                    // 自定义操作中静默处理错误
                    throw;
                }
            }
        }
    }
}