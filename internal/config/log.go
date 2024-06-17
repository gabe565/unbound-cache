package config

import "log/slog"

func initLog(levelText string) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(levelText)); err != nil {
		slog.Warn("Invalid log level. Defaulting to info.", "value", levelText)
		level = slog.LevelInfo
	}
	slog.SetLogLoggerLevel(level)
}
