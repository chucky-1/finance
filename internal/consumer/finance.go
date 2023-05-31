package consumer

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	expensesAdd = "a"
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
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case expensesAdd:
					logrus.Infof("received message in finance consumer to add from username: %s", f.username)
					args := strings.Split(update.Message.CommandArguments(), " ")
					if len(args) != 2 {
						err := f.sendMessage(update.Message, fmt.Sprintf("%s, I can't process it. You only have to enter two words separated by a space. Expense item and amount", f.username))
						if err != nil {
							logrus.Errorf("finance consumer adding expenses error: received invalid message: %s from user %s",
								update.Message.Text, f.username)
							continue
						}
						continue
					}
					logrus.Infof("adding expense: %s, %s", args[0], args[1])
					continue
				}
			}
			logrus.Infof("received message in finance consumer: %s", update.Message.Text)
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
