package configs

import (
	"context"
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
	"go.uber.org/zap"

	"github.com/flexer2006/case-back-restaurant-go/common"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger"
)

type Config struct {
	Database PostgresConfig `yaml:"postgres"`
	Shutdown ShutdownConfig `yaml:"shutdown"`
	Server   ServerConfig   `yaml:"server"`
	SMTP     *SMTPConfig    `yaml:"smtp"`
	LogLevel string         `env:"LOG_LEVEL" env-default:"info" yaml:"log_level"`
}

func Load(ctx context.Context) (*Config, error) {
	log, err := logger.NewLogger()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", common.ErrInitLogger, err)
	}

	ctx = logger.NewContext(ctx, log)

	log.Info(ctx, common.MsgConfigLoading)

	var cfg Config
	err = cleanenv.ReadConfig(".env", &cfg)
	if err != nil {
		log.Error(ctx, common.ErrConfigLoading, zap.Error(err))

		return nil, fmt.Errorf("%s: %w", common.ErrConfigLoad, err)
	}

	smtpConfig, err := NewSMTPConfig()
	if err != nil {
		log.Error(ctx, common.ErrInitSMTP, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", common.ErrInitSMTP, err)
	}
	cfg.SMTP = smtpConfig

	log.Info(ctx, common.MsgConfigLoaded)

	return &cfg, nil
}
