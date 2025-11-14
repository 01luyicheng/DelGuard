using System;
using System.IO;
using Microsoft.VisualBasic.FileIO;

namespace DelGuard
{
    internal sealed class Deleter
    {
        private readonly Logger logger;
        private readonly AppConfig config;

        public Deleter(Logger logger, AppConfig config)
        {
            this.logger = logger;
            this.config = config;
        }

        public void SendToRecycleBin(string path)
        {
            if (File.Exists(path))
            {
                FileSystem.DeleteFile(path, config.UiOption, RecycleOption.SendToRecycleBin, UICancelOption.ThrowException);
                return;
            }

            if (Directory.Exists(path))
            {
                FileSystem.DeleteDirectory(path, config.UiOption, RecycleOption.SendToRecycleBin, UICancelOption.ThrowException);
                return;
            }

            throw new FileNotFoundException(path);
        }
    }
}