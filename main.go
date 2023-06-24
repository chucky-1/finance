package main

import (
	"context"
	"github.com/chucky-1/finance/internal/producer"
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

	//if err := godotenv.Load(); err != nil {
	//	logrus.Fatal("No .env file found")
	//}

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

	mainBot, err := tgbotapi.NewBotAPI(cfg.TGMainBotToken)
	if err != nil {
		logrus.Fatal(err)
	}
	//bot.Debug = true
	u := tgbotapi.NewUpdate(0)
	u.Timeout = cfg.TGMainTimeout
	updatesChan := mainBot.GetUpdatesChan(u)

	myValidator := validator.New()

	postgresRepository := repository.NewPostgres(conn)
	mongoRepository := repository.NewMongo(client)

	authService := service.NewAuth(postgresRepository, cfg.AuthSalt)
	recorderService := service.NewRecorder(mongoRepository)
	reporterService := service.NewReporter(mongoRepository, mongoRepository)

	tgUsersChan := make(chan producer.TGUser)

	hub := consumer.NewHub(mainBot, updatesChan, myValidator, authService, recorderService, reporterService, tgUsersChan)
	go hub.Consume(ctx)

	dailyReporterBot, err := tgbotapi.NewBotAPI(cfg.TGDailyReporterBotToken)
	if err != nil {
		logrus.Fatal(err)
	}
	dailyUpdate := tgbotapi.NewUpdate(0)
	dailyUpdate.Timeout = cfg.TGDailyTimeout
	dailyUpdatesChan := dailyReporterBot.GetUpdatesChan(dailyUpdate)

	monthlyReporterBot, err := tgbotapi.NewBotAPI(cfg.TGMonthlyReporterBotToken)
	if err != nil {
		logrus.Fatal(err)
	}
	monthlyUpdate := tgbotapi.NewUpdate(0)
	monthlyUpdate.Timeout = cfg.TGMonthlyTimeout
	monthlyUpdatesChan := monthlyReporterBot.GetUpdatesChan(monthlyUpdate)

	reporterProducer := producer.NewReporter(dailyReporterBot, monthlyReporterBot, dailyUpdatesChan, monthlyUpdatesChan, reporterService, tgUsersChan)
	go reporterProducer.Produce(ctx)

	logrus.Infof("app has started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)
	<-quit
	cancel()
	<-time.After(2 * time.Second)
}
