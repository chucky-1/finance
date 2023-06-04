package consumer

import (
	"context"
	"fmt"
	"github.com/chucky-1/finance/internal/model"
	"github.com/chucky-1/finance/internal/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

type Finance struct {
	bot         *tgbotapi.BotAPI
	username    string
	updatesChan chan tgbotapi.Update
	serv        service.Finance
}

func NewFinance(bot *tgbotapi.BotAPI, username string, updatesChan chan tgbotapi.Update, serv service.Finance) *Finance {
	return &Finance{
		bot:         bot,
		username:    username,
		updatesChan: updatesChan,
		serv:        serv,
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

			newCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			err = f.serv.Add(newCtx, &model.Entry{
				Type: "expenses",
				Item: args[0],
				User: f.username,
				Date: time.Now().UTC(),
				Sum:  sum,
			})
			if err != nil {
				logrus.Errorf("finance consumer couldn't Add: %v", err)
				cancel()
				continue
			}
			cancel()

			err = f.sendMessage(update.Message, fmt.Sprintf("Added expenses %s: %.2f", args[0], sum))
			if err != nil {
				logrus.Errorf("finance consumer send message error: %v", err)
				continue
			}

			logrus.Infof("%s added expenses: %s: %f", f.username, args[0], sum)
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