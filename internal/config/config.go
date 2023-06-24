package config

type Config struct {
	TGMainBotToken            string `env:"TG_MAIN_BOT_TOKEN"`
	TGMainTimeout             int    `env:"TG_MAIN_TIMEOUT"`
	TGDailyReporterBotToken   string `env:"TG_DAILY_REPORTER_BOT_TOKEN"`
	TGDailyTimeout            int    `env:"TG_DAILY_TIMEOUT"`
	TGMonthlyReporterBotToken string `env:"TG_MONTHLY_REPORTER_BOT_TOKEN"`
	TGMonthlyTimeout          int    `env:"TG_MONTHLY_TIMEOUT"`
	PostgresDB                string `env:"POSTGRES_DB"`
	PostgresUser              string `env:"POSTGRES_USER"`
	PostgresPassword          string `env:"POSTGRES_PASSWORD"`
	PostgresPort              string `env:"POSTGRES_PORT"`
	PostgresEndpoint          string `env:"POSTGRES_ENDPOINT"`
	MongoURI                  string `env:"MONGODB_URI"`
	AuthSalt                  string `env:"AUTHORIZATION_SALT"` // 10 characters is the maximum length
}
