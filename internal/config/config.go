package config

import "time"

type Config struct {
	LogLevel string

	DumpEvery time.Duration
	CachePath string
}

func New() *Config {
	return &Config{
		LogLevel: "info",

		DumpEvery: time.Hour,
		CachePath: "/opt/unbound/data/cache.txt.gz",
	}
}
