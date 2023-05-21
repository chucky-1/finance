package config

type Config struct {
	TgToken          string `env:"TG_TOKEN"`
	TgTimeout        int    `env:"TG_TIMEOUT"`
	PostgresDB       string `env:"POSTGRES_DB"`
	PostgresUser     string `env:"POSTGRES_USER"`
	PostgresPassword string `env:"POSTGRES_PASSWORD"`
	PostgresPort     string `env:"POSTGRES_PORT"`
	PostgresEndpoint string `env:"POSTGRES_ENDPOINT"`
	MongoURI         string `env:"MONGODB_URI"`
	AuthSalt         string `env:"AUTHORIZATION_SALT"` // 10 characters is the maximum length
}
