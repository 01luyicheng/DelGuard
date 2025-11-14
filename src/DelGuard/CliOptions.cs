using System;
using System.Collections.Generic;
using System.Linq;

namespace DelGuard
{
    internal sealed class CliOptions
    {
        public bool Force { get; private set; }
        public bool Preview { get; private set; }
        public bool Recursive { get; private set; }
        public bool Silent { get; private set; }
        public bool Verbose { get; private set; }
        public List<string> Paths { get; private set; } = new();

        public static CliOptions Parse(string[] args)
        {
            var opts = new CliOptions();
            bool afterDoubleDash = false;
            foreach (var a in args)
            {
                if (!afterDoubleDash && a == "--")
                {
                    afterDoubleDash = true;
                    continue;
                }

                if (!afterDoubleDash && a.StartsWith("-"))
                {
                    switch (a)
                    {
                        case "-f":
                        case "--force":
                            opts.Force = true; break;
                        case "-p":
                        case "--preview":
                            opts.Preview = true; break;
                        case "-r":
                        case "--recursive":
                            opts.Recursive = true; break;
                        case "-s":
                        case "--silent":
                            opts.Silent = true; break;
                        case "-v":
                        case "--verbose":
                            opts.Verbose = true; break;
                        default:
                            break;
                    }
                }
                else
                {
                    opts.Paths.Add(a);
                }
            }

            return opts;
        }
    }
}