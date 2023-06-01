package consumer

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

type Finance struct {
	bot         *tgbotapi.BotAPI
	username    string
	updatesChan chan tgbotapi.Update
}

func NewFinance(bot *tgbotapi.BotAPI, username string, updatesChan chan tgbotapi.Update) *Finance {
	return &Finance{
		bot:         bot,
		username:    username,
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
			logrus.Infof("received message in finance consumer from username: %s", f.username)
			args := strings.Split(update.Message.Text, " ")
			if len(args) != 2 {
				logrus.Errorf("finance consumer received invalid message: %s", update.Message.Text)
				err := f.sendMessage(update.Message, fmt.Sprintf("%s, I can't process it. You only have to enter two words separated by a space. Expense item and amount", f.username))
				if err != nil {
					logrus.Errorf("finance consumer send message error: %v", err)
					continue
				}
				continue
			}

			sum, err := strconv.ParseFloat(args[1], 64)
			if err != nil {
				logrus.Errorf("finance consumer couldn't parseFloat: %v", err)
				err = f.sendMessage(update.Message, fmt.Sprintf("%s, the second argument must be a number", f.username))
				if err != nil {
					logrus.Errorf("finance consumer send message error: %v", err)
					continue
				}
				continue
			}
			logrus.Infof("adding expense: %s, sum: %f", args[0], sum)
		}
	}
}

func (f *Finance) sendMessage(message *tgbotapi.Message, text string) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyToMessageID = message.MessageID

	_, err := f.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("sendMessage, telegram bot couldn't send message: %v", err)
	}
	return nil
}
