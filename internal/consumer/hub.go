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
	financeChannels map[int64]chan struct{}
}

func NewHub(bot *tgbotapi.BotAPI, updatesChan tgbotapi.UpdatesChannel, validator *validator.Validate, authService service.Authorization) *Hub {
	return &Hub{
		bot:             bot,
		updatesChan:     updatesChan,
		validator:       validator,
		authService:     authService,
		authChannels:    make(map[int64]chan tgbotapi.Update),
		financeChannels: make(map[int64]chan struct{}),
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
					authorized, err := h.checkUserAlreadyAuthorized(update.Message.Chat.ID)
					if err != nil {
						logrus.Errorf("register/login error: %v", err)
						continue
					}
					if authorized {
						logrus.Errorf("register/login error: user with chat %d already is authorized", update.Message.Chat.ID)
						if err = h.sendMessage(update.Message, "you are already authorized"); err != nil {
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
				default:
					logrus.Info("unknown command: %s", update.Message.Text)
					continue
				}
			}
			financeCh, ok := h.financeChannels[update.Message.Chat.ID]
			if ok {
				financeCh <- struct{}{}
				continue
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

func (h *Hub) startAuthConsumer(ctx context.Context, chatID int64) (chan tgbotapi.Update, chan int64) {
	finishChan := make(chan int64)
	newUpdatesChan := make(chan tgbotapi.Update)
	h.authChannels[chatID] = newUpdatesChan
	authConsumer := NewAuth(h.bot, newUpdatesChan, h.validator, h.authService, finishChan)
	newCtx, _ := context.WithCancel(ctx)
	go authConsumer.Consume(newCtx)
	return newUpdatesChan, finishChan
}

func (h *Hub) listenFinish(ctx context.Context, finishChan chan int64) {
	select {
	case <-ctx.Done():
		return
	case chatID := <-finishChan:
		logrus.Infof("hub received message in finish chat with chat id %d", chatID)
		delete(h.authChannels, chatID)
		h.financeChannels[chatID] = make(chan struct{})
	}
}

func (h *Hub) checkUserAlreadyAuthorized(chatID int64) (bool, error) {
	_, ok := h.financeChannels[chatID]
	if !ok {
		return false, nil
	}
	return true, nil
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
