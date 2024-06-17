package config

import "github.com/spf13/cobra"

const (
	FlagLogLevel = "log-level"

	FlagDumpEvery = "dump-every"
	FlagCachePath = "path"
)

func (c *Config) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&c.LogLevel, FlagLogLevel, "l", c.LogLevel, "Log level (trace, debug, info, warn, error, fatal, panic)")
	cmd.Flags().DurationVar(&c.DumpEvery, FlagDumpEvery, c.DumpEvery, "Regular dump interval")
	cmd.Flags().StringVar(&c.CachePath, FlagCachePath, c.CachePath, "Cache file path")
}
