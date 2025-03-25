package configs

import "time"

type ShutdownConfig struct {
	Timeout time.Duration `env:"SHUTDOWN_TIMEOUT" env-default:"5s"`
}
