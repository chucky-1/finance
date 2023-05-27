package consumer

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

const (
	add = "add"
)

type Finance struct {
	updatesChan chan tgbotapi.Update
}

func NewFinance(updatesChan chan tgbotapi.Update) *Finance {
	return &Finance{
		updatesChan: updatesChan,
	}
}

func (f *Finance) Consume(ctx context.Context) {
	logrus.Info("finance consumer started")
	for {
		select {
		case <-ctx.Done():
			logrus.Infof("finance consumer stopped: %v", ctx.Err())
			return
		case update := <-f.updatesChan:
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case add:
					logrus.Infof("received message in finance consumer to add: %s", update.Message.Text)
					continue
				}
			}
			logrus.Infof("received message in finance consumer: %s", update.Message.Text)
		}
	}
}
