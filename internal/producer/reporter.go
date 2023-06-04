package producer

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

type TGUser struct {
	TGUsername string
	Username   string
}

type Reporter struct {
	dailyReporterBot   *tgbotapi.BotAPI
	dailyUpdatesChan   tgbotapi.UpdatesChannel
	monthlyReporterBot *tgbotapi.BotAPI
	monthlyUpdatesChan tgbotapi.UpdatesChannel

	// key: tgUserName, value: username
	// receiving from hub consumer
	tgUsers     map[string]string
	tgUsersChan <-chan TGUser
}

func NewReporter(dailyReporterBot, monthlyReporterBot *tgbotapi.BotAPI, dailyUpdatesChan, monthlyUpdatesChan tgbotapi.UpdatesChannel,
	tgUsersChan chan TGUser) *Reporter {
	return &Reporter{
		dailyReporterBot:   dailyReporterBot,
		monthlyReporterBot: monthlyReporterBot,
		tgUsers:            make(map[string]string),
		tgUsersChan:        tgUsersChan,
	}
}

func (r *Reporter) Produce(ctx context.Context) {
	logrus.Info("reporter producer started")
	for {
		select {
		case <-ctx.Done():
			logrus.Infof("reporter producer stopped: %v", ctx.Err())
			return
		case tgUser := <-r.tgUsersChan:
			logrus.Infof("reporter producer received message: %v", tgUser)
			r.tgUsers[tgUser.TGUsername] = tgUser.Username
		}
	}
}
