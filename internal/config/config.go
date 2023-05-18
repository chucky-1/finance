package config

type Config struct {
	Telegram         Telegram
	PostgresEndpoint string `env:"POSTGRES_ENDPOINT"`
	AuthSalt         string `env:"AUTHORIZATION_SALT" envDefault:"ugtyvmfldo"` // 10 characters is the maximum length
}

type Telegram struct {
	Timeout int `env:"TIMEOUT" envDefault:"60"`
}
