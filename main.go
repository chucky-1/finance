package main

import (
	"context"
	"fmt"
	"github.com/chucky-1/finance/internal/repository"
	"github.com/chucky-1/finance/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v8"
	"github.com/chucky-1/finance/internal/consumer"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"

	"github.com/chucky-1/finance/internal/config"
)

func main() {
	//logrus.SetFormatter(new(logrus.JSONFormatter))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := godotenv.Load(); err != nil {
		logrus.Fatal("No .env file found")
	}

	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}

	conn, err := pgxpool.Connect(ctx, cfg.PostgresEndpoint)
	if err != nil {
		logrus.Fatalf("couldn't connect to database: %v", err)
	}
	if err = conn.Ping(ctx); err != nil {
		logrus.Fatalf("couldn't ping database: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TG_TOKEN"))
	if err != nil {
		logrus.Fatal(err)
	}
	//bot.Debug = true
	u := tgbotapi.NewUpdate(0)
	u.Timeout = cfg.Telegram.Timeout
	updatesChan := bot.GetUpdatesChan(u)

	myValidator := validator.New()

	authRepository := repository.NewAuth(conn)
	authService := service.NewAuth(authRepository, cfg.AuthSalt)

	tgBot := consumer.NewBot(bot, updatesChan, myValidator, authService)
	tgBot.Consume(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)
	<-quit
	cancel()
	<-time.After(2 * time.Second)
}
