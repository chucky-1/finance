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
	recorder    *service.Recorder
}

func NewFinance(bot *tgbotapi.BotAPI, username string, updatesChan chan tgbotapi.Update, recorder *service.Recorder) *Finance {
	return &Finance{
		bot:         bot,
		username:    username,
		updatesChan: updatesChan,
		recorder:    recorder,
	}
}

func (f *Finance) Consume(ctx context.Context) {
	logrus.Debugf("finance consumer started")
	for {
		select {
		case <-ctx.Done():
			logrus.Debugf("finance consumer stopped: %v", ctx.Err())
			return
		case update := <-f.updatesChan:
			logrus.Debugf("received message in finance consumer from username: %s", f.username)
			args := strings.Split(update.Message.Text, " ")
			if len(args) != 2 {
				logrus.Debugf("finance consumer received invalid message: %s", update.Message.Text)
				err := f.sendMessage(update.Message, fmt.Sprintf("%s, мы не можем обработать ваш запрос. Вы должны ввести только 2 параметра разделённых пробелом: статью расходов и сумму", f.username))
				if err != nil {
					logrus.Errorf("finance consumer send message error: %v", err)
					continue
				}
				continue
			}

			sum, err := strconv.ParseFloat(args[1], 64)
			if err != nil {
				logrus.Debugf("finance consumer couldn't parseFloat: %v", err)
				err = f.sendMessage(update.Message, fmt.Sprintf("%s, второй параметр должен быть числом", f.username))
				if err != nil {
					logrus.Errorf("finance consumer send message error: %v", err)
					continue
				}
				continue
			}

			newCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			err = f.recorder.Add(newCtx, &model.Entry{
				Kind: "expenses",
				User: f.username,
				Date: time.Now().UTC(),
				Category: &model.Category{
					Name:   args[0],
					Amount: sum,
				},
			})
			if err != nil {
				logrus.Errorf("finance consumer couldn't Add: %v", err)
				cancel()
				continue
			}
			cancel()

			err = f.sendMessage(update.Message, fmt.Sprintf("Добавлены расходы\n%s: %.2f", args[0], sum))
			if err != nil {
				logrus.Errorf("finance consumer send message error: %v", err)
				continue
			}

			logrus.Debugf("%s added expenses: %s: %.2f", f.username, args[0], sum)
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
