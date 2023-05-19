// Package consumer
package consumer

import (
	"context"
	"fmt"
	"github.com/chucky-1/finance/internal/model"
	"github.com/chucky-1/finance/internal/service"
	"time"

	"github.com/go-playground/validator/v10"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

const (
	start    = "start"
	register = "register"
	login    = "login"
)

const (
	usernameMinLength = 3
	usernameMaxLength = 15
	passwordMaxLength = 15
)

// Bot sends requests to the telegram server every n seconds and if there are new messages it receives them
type Bot struct {
	bot         *tgbotapi.BotAPI
	updatesChan tgbotapi.UpdatesChannel
	validator   *validator.Validate
	auth        service.Authorization
	chats       service.Chats
}

func NewBot(bot *tgbotapi.BotAPI, updatesChan tgbotapi.UpdatesChannel, validator *validator.Validate, auth service.Authorization,
	chats service.Chats) *Bot {
	return &Bot{
		bot:         bot,
		updatesChan: updatesChan,
		validator:   validator,
		auth:        auth,
		chats:       chats,
	}
}

// TODO не будет работать если регистрируется больше одного человека одновременно. Нужен кэш
var (
	waitRegisterMessageWithUsername int
	waitRegisterMessageWithPassword int
	waitLoginMessageWithUsername    int
	waitLoginMessageWithPassword    int
	username                        string
	password                        string
)

func (b *Bot) Consume(ctx context.Context) {
	logrus.Infof("telegram bot %s started consuming", b.bot.Self.UserName)

	for {
		select {
		case <-ctx.Done():
			logrus.Infof("bot consumer stopped: %v", ctx.Err())
			return

		case update := <-b.updatesChan:
			if update.Message.MessageID == waitRegisterMessageWithUsername {
				if err := b.handleUsername(register, update.Message); err != nil {
					logrus.Errorf("register error: %v", err)
					continue
				}
				if err := b.requestForPassword(register, update.Message, "Enter your password"); err != nil {
					logrus.Errorf("register error: %v", err)
					continue
				}
				continue
			}

			if update.Message.MessageID == waitRegisterMessageWithPassword {
				if err := b.handlePassword(register, update.Message); err != nil {
					logrus.Errorf("register error: %v", err)
					continue
				}

				newCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
				if err := b.auth.CreateUser(newCtx, &model.User{
					Username: username,
					Password: password,
				}); err != nil {
					logrus.Errorf("register error: %v", err)
					cancel()
					continue
				}
				cancel()

				newCtx, cancel = context.WithTimeout(ctx, 10*time.Second)
				if err := b.chats.Add(newCtx, update.Message.Chat.ID, username); err != nil {
					logrus.Errorf("register error: %v", err)
					cancel()
					continue
				}
				cancel()

				if err := b.sendMessage(update.Message, fmt.Sprintf("Thank you, %s! You have successfully registered", username)); err != nil {
					logrus.Errorf("register error: %v", err)
					continue
				}

				logrus.Infof("user %s successful registered", username)
				continue
			}

			if update.Message.MessageID == waitLoginMessageWithUsername {
				if err := b.handleUsername(login, update.Message); err != nil {
					logrus.Errorf("login error: %v", err)
					continue
				}
				if err := b.requestForPassword(login, update.Message, "Enter your password"); err != nil {
					logrus.Errorf("login error: %v", err)
					continue
				}
			}

			if update.Message.MessageID == waitLoginMessageWithPassword {
				if err := b.handlePassword(login, update.Message); err != nil {
					logrus.Errorf("login error: %v", err)
					continue
				}

				newCtx, cancel := context.WithCancel(ctx)
				ok, err := b.auth.Login(newCtx, &model.User{
					Username: username,
					Password: password,
				})
				if err != nil {
					logrus.Errorf("login error: %v", err)
					cancel()
					continue
				}
				cancel()
				if !ok {
					logrus.Infof("user %s inputted invalid password", username)
					if err = b.requestForPassword(login, update.Message, "%s, you inputted invalid password. Try again!"); err != nil {
						logrus.Errorf("login error: %v", err)
						continue
					}
				}

				newCtx, cancel = context.WithCancel(ctx)
				if err = b.chats.Add(newCtx, update.Message.Chat.ID, username); err != nil {
					logrus.Errorf("login error: %v", err)
					cancel()
					continue
				}
				cancel()

				if err = b.sendMessage(update.Message, "%s, you are authorized!"); err != nil {
					logrus.Errorf("login error: %v", err)
					continue
				}

				logrus.Infof("user %s is authorized", username)
			}

			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case start:
					logrus.Info("start command executed")
					continue
				case register:
					logrus.Info("register command started executing")
					registerText := fmt.Sprintf("Enter your username. Minimum %d, maximum %d characters", usernameMinLength, usernameMaxLength)
					if err := b.requestForUsername(register, update.Message, registerText); err != nil {
						logrus.Errorf("register error: %v", err)
						continue
					}
					continue
				case login:
					logrus.Info("login command started executing")
					loginText := fmt.Sprintf("Enter your username")
					if err := b.requestForUsername(login, update.Message, loginText); err != nil {
						logrus.Errorf("login error: %v", err)
					}
					continue
				default:
					logrus.Info("unknown command: %s", update.Message.Text)
					continue
				}
			}

			logrus.Infof("recieved message: %s", update.Message.Text)
		}
	}
}

func (b *Bot) handleUsername(action string, message *tgbotapi.Message) error {
	username = message.Text
	if !b.validate(username, fmt.Sprintf("min=%d,max=%d", usernameMinLength, usernameMaxLength)) {
		err := b.requestForUsername(action, message, "You entered the wrong username. Try again!")
		if err != nil {
			return err
		}
		return fmt.Errorf("user entered the wrong username: %s", username)
	}
	logrus.Infof("register, user entered username: %s", username)
	return nil
}

func (b *Bot) handlePassword(action string, message *tgbotapi.Message) error {
	password = message.Text
	if !b.validate(password, fmt.Sprintf("max=%d", passwordMaxLength)) {
		err := b.requestForPassword(action, message, fmt.Sprintf("%s, you entered the wrong password. Try again!", username))
		if err != nil {
			return err
		}
		return fmt.Errorf("user %s entered the wrong password: %s", username, password)
	}
	logrus.Infof("register, user %s entered password: %s", username, password)
	return nil
}

func (b *Bot) requestForUsername(action string, message *tgbotapi.Message, text string) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyToMessageID = message.MessageID

	switch action {
	case register:
		waitRegisterMessageWithUsername = msg.ReplyToMessageID + 2
	case login:
		waitLoginMessageWithUsername = msg.ReplyToMessageID + 2
	}

	_, err := b.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("requestForUsername, telegram bot couldn't send message: %v", err)
	}
	return nil
}

func (b *Bot) requestForPassword(action string, message *tgbotapi.Message, text string) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyToMessageID = message.MessageID

	switch action {
	case register:
		waitRegisterMessageWithPassword = msg.ReplyToMessageID + 2
	case login:
		waitLoginMessageWithPassword = msg.ReplyToMessageID + 2
	}

	_, err := b.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("requestForPassword, telegram bot couldn't send message: %v", err)
	}
	return nil
}

func (b *Bot) sendMessage(message *tgbotapi.Message, text string) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyToMessageID = message.MessageID

	_, err := b.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("sendMessage, telegram bot couldn't send message: %v", err)
	}
	return nil
}

func (b *Bot) validate(value string, tags string) bool {
	err := b.validator.Var(value, tags)
	if err != nil {
		return false
	}
	return true
}
