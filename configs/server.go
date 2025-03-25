package configs

type ServerConfig struct {
	Host string `env:"SERVER_HOST" env-default:"localhost"`
	Port int    `env:"SERVER_PORT" env-default:"8080"`
}
