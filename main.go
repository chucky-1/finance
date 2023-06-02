package main

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v8"
	"github.com/go-playground/validator/v10"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"github.com/chucky-1/finance/internal/config"
	"github.com/chucky-1/finance/internal/consumer"
	"github.com/chucky-1/finance/internal/repository"
	"github.com/chucky-1/finance/internal/service"
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
		logrus.Fatalf("%+v\n", err)
	}

	conn, err := pgxpool.Connect(ctx, cfg.PostgresEndpoint)
	if err != nil {
		logrus.Fatalf("couldn't connect to database: %v", err)
	}
	if err = conn.Ping(ctx); err != nil {
		logrus.Fatalf("couldn't ping database: %v", err)
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		logrus.Fatalf("couldn't connect to mongo: %v", err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		logrus.Fatalf("failed ping to mongo: %v", err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			logrus.Fatalf("couldn't disconnect to mongo: %v", err)
		}
	}()

	bot, err := tgbotapi.NewBotAPI(cfg.TgToken)
	if err != nil {
		logrus.Fatal(err)
	}
	//bot.Debug = true
	u := tgbotapi.NewUpdate(0)
	u.Timeout = cfg.TgTimeout
	updatesChan := bot.GetUpdatesChan(u)

	myValidator := validator.New()

	userRepository := repository.NewUserPostgres(conn)
	financeRepository := repository.NewFinance(client)
	authService := service.NewAuth(userRepository, cfg.AuthSalt)
	financeService := service.NewFinance(financeRepository)

	tgBot := consumer.NewHub(bot, updatesChan, myValidator, authService, financeService)
	go tgBot.Consume(ctx)

	cleaner := consumer.NewCleaner(financeRepository)
	go cleaner.Consume(ctx)

	logrus.Infof("app has started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)
	<-quit
	cancel()
	<-time.After(2 * time.Second)
}
