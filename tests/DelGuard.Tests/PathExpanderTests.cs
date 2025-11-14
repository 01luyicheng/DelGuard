using System;
using System.IO;
using Xunit;

namespace DelGuard.Tests
{
    public class PathExpanderTests
    {
        [Fact]
        public void Expand_Wildcard_MatchesFiles()
        {
            var tmp = Path.Combine(Path.GetTempPath(), Guid.NewGuid().ToString("N"));
            Directory.CreateDirectory(tmp);
            try
            {
                File.WriteAllText(Path.Combine(tmp, "a.log"), "x");
                File.WriteAllText(Path.Combine(tmp, "b.log"), "y");
                var logger = new DelGuard.Logger(DelGuard.LogLevel.None);
                var list = DelGuard.PathExpander.Expand(new[] { Path.Combine(tmp, "*.log") }, logger);
                Assert.Contains(Path.Combine(tmp, "a.log"), list);
                Assert.Contains(Path.Combine(tmp, "b.log"), list);
            }
            finally
            {
                Directory.Delete(tmp, true);
            }
        }
    }
}