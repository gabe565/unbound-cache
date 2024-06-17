package config

import "log/slog"

func initLog(levelText string) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(levelText)); err != nil {
		level = slog.LevelInfo
	}
	slog.SetLogLoggerLevel(level)
}
