package config

type Config struct {
	Telegram Telegram
}

type Telegram struct {
	Timeout int `env:"TIMEOUT" envDefault:"60"`
}
