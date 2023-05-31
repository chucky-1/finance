package consumer

import (
	"context"
	"fmt"
	"github.com/chucky-1/finance/internal/service"
	"github.com/go-playground/validator/v10"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

type Hub struct {
	bot             *tgbotapi.BotAPI
	updatesChan     tgbotapi.UpdatesChannel
	validator       *validator.Validate
	authService     service.Authorization
	authChannels    map[int64]chan tgbotapi.Update
	financeChannels map[int64]chan tgbotapi.Update
}

func NewHub(bot *tgbotapi.BotAPI, updatesChan tgbotapi.UpdatesChannel, validator *validator.Validate, authService service.Authorization) *Hub {
	return &Hub{
		bot:             bot,
		updatesChan:     updatesChan,
		validator:       validator,
		authService:     authService,
		authChannels:    make(map[int64]chan tgbotapi.Update),
		financeChannels: make(map[int64]chan tgbotapi.Update),
	}
}

func (h *Hub) Consume(ctx context.Context) {
	logrus.Info("hub consumer started")
	for {
		select {
		case <-ctx.Done():
			logrus.Infof("hub consumer stopped: %v", ctx.Err())
			return
		case update := <-h.updatesChan:
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case register, login:
					logrus.Infof("received message in hub consumer to register or login from chat %d", update.Message.Chat.ID)
					if h.authorized(update.Message.Chat.ID) {
						logrus.Errorf("register/login error: user with chat %d already is authorized", update.Message.Chat.ID)
						if err := h.sendMessage(update.Message, "you are already authorized"); err != nil {
							logrus.Errorf("register/login error: %v", err)
							continue
						}
						continue
					}

					ch, ok := h.authChannels[update.Message.Chat.ID]
					if !ok {
						// first touch with the user
						logrus.Infof("first touch with the user with chat id %d", update.Message.Chat.ID)
						newUpdatesChan, finishChan := h.startAuthConsumer(ctx, update.Message.Chat.ID)
						newUpdatesChan <- update
						go h.listenFinish(ctx, finishChan)
						continue
					}
					ch <- update
					continue
				case expensesAdd:
					logrus.Infof("received message in hub consumer to add from chat %d", update.Message.Chat.ID)
					if !h.authorized(update.Message.Chat.ID) {
						logrus.Errorf("add error: user with chat %d isn't authorized", update.Message.Chat.ID)
						if err := h.sendMessage(update.Message, "you aren't authorized"); err != nil {
							logrus.Errorf("add error: %v", err)
							continue
						}
						continue
					}
					financeCh, ok := h.financeChannels[update.Message.Chat.ID]
					if ok {
						financeCh <- update
						continue
					} else {
						logrus.Errorf("user with id %d is authorized but there's no channel in map finance")
						continue
					}
				default:
					logrus.Infof("unknown command: %s", update.Message.Text)
					continue
				}
			}
			authCh, ok := h.authChannels[update.Message.Chat.ID]
			if ok {
				authCh <- update
				continue
			}
			logrus.Infof("recieved message: %s", update.Message.Text)
		}
	}
}

func (h *Hub) startAuthConsumer(ctx context.Context, chatID int64) (chan tgbotapi.Update, chan *finishData) {
	finishChan := make(chan *finishData)
	newUpdatesChan := make(chan tgbotapi.Update)
	h.authChannels[chatID] = newUpdatesChan
	authConsumer := NewAuth(h.bot, newUpdatesChan, h.validator, h.authService, finishChan)
	newCtx, _ := context.WithCancel(ctx)
	go authConsumer.Consume(newCtx)
	return newUpdatesChan, finishChan
}

func (h *Hub) listenFinish(ctx context.Context, finishChan chan *finishData) {
	select {
	case <-ctx.Done():
		return
	case data := <-finishChan:
		logrus.Infof("hub received message in finish chat with chat id %d", data.chatID)
		delete(h.authChannels, data.chatID)
		financeChan := make(chan tgbotapi.Update)
		h.financeChannels[data.chatID] = financeChan
		go NewFinance(h.bot, data.username, financeChan).Consume(ctx)
	}
}

func (h *Hub) authorized(chatID int64) bool {
	_, ok := h.financeChannels[chatID]
	if !ok {
		return false
	}
	return true
}

func (h *Hub) sendMessage(message *tgbotapi.Message, text string) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyToMessageID = message.MessageID

	_, err := h.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("sendMessage, telegram bot couldn't send message: %v", err)
	}
	return nil
}
