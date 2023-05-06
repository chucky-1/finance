// Package consumer
package consumer

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

// Bot sends requests to the telegram server every n seconds and if there are new messages it receives them
type Bot struct {
	bot     *tgbotapi.BotAPI
	timeout int
}

func NewBot(bot *tgbotapi.BotAPI, timeout int) *Bot {
	return &Bot{
		bot:     bot,
		timeout: timeout,
	}
}

func (b *Bot) Consume(ctx context.Context) {
	logrus.Infof("telegram bot %s started consuming", b.bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = b.timeout
	updates := b.bot.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			return
		case update := <-updates:
			if update.Message != nil {
				logrus.Infof("[%s] %s", update.Message.From.UserName, update.Message.Text)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
				msg.ReplyToMessageID = update.Message.MessageID

				_, err := b.bot.Send(msg)
				if err != nil {
					logrus.Errorf("telegram bot couldn't send message: %v", err)
					continue
				}
			}
		}
	}
}
