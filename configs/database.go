package configs

type PostgresConfig struct {
	Host           string `env:"POSTGRES_HOST"            env-default:"localhost"`
	Port           int    `env:"POSTGRES_PORT"            env-default:"5432"`
	Username       string `env:"POSTGRES_USER"            env-default:"postgres"`
	Password       string `env:"POSTGRES_PASSWORD"        env-default:"root"`
	Database       string `env:"POSTGRES_DB"              env-default:"postgres"`
	SSLMode        string `env:"POSTGRES_SSLMODE"         env-default:"disable"`
	MaxConnections int    `env:"POSTGRES_MAX_CONNECTIONS" env-default:"100"`
	MinConnections int    `env:"POSTGRES_MIN_CONNECTIONS" env-default:"3"`
}
