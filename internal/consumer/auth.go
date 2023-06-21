package consumer

import (
	"context"
	"fmt"
	"github.com/chucky-1/finance/internal/model"
	"github.com/chucky-1/finance/internal/service"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

const (
	register = "register"
	login    = "login"
)

const (
	usernameMinLength = 3
	usernameMaxLength = 15
	passwordMaxLength = 15
)

type finishData struct {
	username   string
	chatID     int64
	tgUsername string
}

type Auth struct {
	bot         *tgbotapi.BotAPI
	updatesChan chan tgbotapi.Update
	validator   *validator.Validate
	auth        service.Authorization
	reporter    *service.Reporter
	finish      chan<- *finishData

	waitRegisterMessageWithUsername int
	waitRegisterMessageWithCountry  int
	waitRegisterMessageWithPassword int
	waitLoginMessageWithUsername    int
	waitLoginMessageWithPassword    int
	username                        string
	country                         string
	timezone                        time.Duration
	password                        string
}

func NewAuth(bot *tgbotapi.BotAPI, updatesChan chan tgbotapi.Update, validator *validator.Validate, auth service.Authorization,
	reporter *service.Reporter, finish chan<- *finishData) *Auth {
	return &Auth{
		bot:         bot,
		updatesChan: updatesChan,
		validator:   validator,
		auth:        auth,
		reporter:    reporter,
		finish:      finish,
	}
}

func (a *Auth) Consume(ctx context.Context) {
	logrus.Infof("auth consumer started consuming")

	for {
		select {
		case <-ctx.Done():
			logrus.Infof("auth consumer for user %s stopped: %v", a.username, ctx.Err())
			return

		case update := <-a.updatesChan:
			if !update.Message.IsCommand() && update.Message.MessageID == a.waitRegisterMessageWithUsername {
				if err := a.handleUsername(register, update.Message); err != nil {
					logrus.Errorf("register error: %v", err)
					continue
				}

				if err := a.requestForCountry(update.Message); err != nil {
					logrus.Errorf("register error: %v", err)
					continue
				}
			}

			if !update.Message.IsCommand() && update.Message.MessageID == a.waitRegisterMessageWithCountry {
				if err := a.handleCountry(update.Message); err != nil {
					logrus.Errorf("register error: %v", err)
					continue
				}

				if err := a.requestForPassword(register, update.Message,
					fmt.Sprintf("Enter your password. Maximum %d characters", passwordMaxLength)); err != nil {
					logrus.Errorf("register error: %v", err)
					continue
				}
				continue
			}

			if !update.Message.IsCommand() && update.Message.MessageID == a.waitRegisterMessageWithPassword {
				if err := a.handlePassword(register, update.Message); err != nil {
					logrus.Errorf("register error: %v", err)
					continue
				}

				newCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
				success, err := a.auth.CreateUser(newCtx, &model.User{
					Username: a.username,
					Password: a.password,
					Country:  a.country,
					Timezone: a.timezone,
				})
				if err != nil {
					logrus.Errorf("register error: %v", err)
					cancel()
					continue
				}
				cancel()
				if !success {
					logrus.Errorf("register error: user with username: %s already exist", a.username)
					if err = a.requestForUsername(register, update.Message,
						fmt.Sprintf("User with username: %s already exist. Try again! Enter your username", a.username)); err != nil {
						logrus.Errorf("register error: %v", err)
						continue
					}
					continue
				}

				a.reporter.AddTimezone(a.timezone, a.username)

				if err = a.sendMessage(update.Message, fmt.Sprintf("Thank you, %s! You have successfully registered", a.username)); err != nil {
					logrus.Errorf("register error: %v", err)
					continue
				}

				logrus.Infof("user %s successful registered", a.username)
				logrus.Infof("auth consumer for user %s stopped", a.username)
				a.finish <- &finishData{
					username:   a.username,
					chatID:     update.Message.Chat.ID,
					tgUsername: update.SentFrom().UserName,
				}
				return
			}

			if !update.Message.IsCommand() && update.Message.MessageID == a.waitLoginMessageWithUsername {
				if err := a.handleUsername(login, update.Message); err != nil {
					logrus.Errorf("login error: %v", err)
					continue
				}
				if err := a.requestForPassword(login, update.Message, "Enter your password"); err != nil {
					logrus.Errorf("login error: %v", err)
					continue
				}
				continue
			}

			if !update.Message.IsCommand() && update.Message.MessageID == a.waitLoginMessageWithPassword {
				if err := a.handlePassword(login, update.Message); err != nil {
					logrus.Errorf("login error: %v", err)
					continue
				}

				newCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
				ok, err := a.auth.Login(newCtx, &model.User{
					Username: a.username,
					Password: a.password,
				})
				if err != nil {
					logrus.Errorf("login error: %v", err)
					cancel()
					continue
				}
				cancel()
				if !ok {
					logrus.Errorf("login error: invalid username: %s or password: %s", a.username, a.password)
					if err = a.requestForUsername(login, update.Message, "You inputted invalid username or password. Try again! Enter your username"); err != nil {
						logrus.Errorf("login error: %v", err)
						continue
					}
					continue
				}

				if err = a.sendMessage(update.Message, fmt.Sprintf("%s, you are authorized!", a.username)); err != nil {
					logrus.Errorf("login error: %v", err)
					continue
				}

				logrus.Infof("user %s is authorized", a.username)
				logrus.Infof("auth consumer for user %s stopped", a.username)
				a.finish <- &finishData{
					username:   a.username,
					chatID:     update.Message.Chat.ID,
					tgUsername: update.SentFrom().UserName,
				}
				return
			}

			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case register:
					logrus.Info("register command started executing")
					registerText := fmt.Sprintf("Enter your username. Minimum %d, maximum %d characters", usernameMinLength, usernameMaxLength)
					if err := a.requestForUsername(register, update.Message, registerText); err != nil {
						logrus.Errorf("register error: %v", err)
						continue
					}
					continue
				case login:
					logrus.Info("login command started executing")
					loginText := fmt.Sprintf("Enter your username")
					if err := a.requestForUsername(login, update.Message, loginText); err != nil {
						logrus.Errorf("login error: %v", err)
						continue
					}
					continue
				}
			}
		}
	}
}

func (a *Auth) handleUsername(action string, message *tgbotapi.Message) error {
	a.username = message.Text
	if !a.validate(a.username, fmt.Sprintf("min=%d,max=%d", usernameMinLength, usernameMaxLength)) {
		err := a.requestForUsername(action, message, "You entered the wrong username. Try again!")
		if err != nil {
			return err
		}
		return fmt.Errorf("user entered the wrong username: %s", a.username)
	}
	logrus.Infof("%s, user entered username: %s", action, a.username)
	return nil
}

func (a *Auth) handlePassword(action string, message *tgbotapi.Message) error {
	a.password = message.Text
	if !a.validate(a.password, fmt.Sprintf("max=%d", passwordMaxLength)) {
		err := a.requestForPassword(action, message, fmt.Sprintf("%s, you entered the wrong password. Try again!", a.username))
		if err != nil {
			return err
		}
		return fmt.Errorf("user %s entered the wrong password: %s", a.username, a.password)
	}
	logrus.Infof("%s, user %s entered password: %s", action, a.username, a.password)
	return nil
}

func (a *Auth) handleCountry(message *tgbotapi.Message) error {
	a.country = strings.Split(strings.Trim(strings.Split(message.Text, "(")[0], " "), ",")[0]
	_, after, _ := strings.Cut(message.Text, "GMT")
	timezoneString := strings.Trim(after, ")")
	timezone, err := strconv.ParseFloat(timezoneString, 32)
	if err != nil {
		return fmt.Errorf("handle country couldn't parse int: %v", err)
	}
	hour := int(timezone)
	minute := int((timezone - float64(int(timezone))) * 100)
	a.timezone = time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute
	logrus.Infof("%s chose country: %s and timezone: %v", a.username, a.country, a.timezone)
	return nil
}

func (a *Auth) requestForUsername(action string, message *tgbotapi.Message, text string) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyToMessageID = message.MessageID

	switch action {
	case register:
		a.waitRegisterMessageWithUsername = msg.ReplyToMessageID + 2
	case login:
		a.waitLoginMessageWithUsername = msg.ReplyToMessageID + 2
	}

	_, err := a.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("requestForUsername, telegram bot couldn't send message: %v", err)
	}
	return nil
}

func (a *Auth) requestForPassword(action string, message *tgbotapi.Message, text string) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyToMessageID = message.MessageID
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)

	switch action {
	case register:
		a.waitRegisterMessageWithPassword = msg.ReplyToMessageID + 2
	case login:
		a.waitLoginMessageWithPassword = msg.ReplyToMessageID + 2
	}

	_, err := a.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("requestForPassword, telegram bot couldn't send message: %v", err)
	}
	return nil
}

func (a *Auth) requestForCountry(message *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Выберете свою страну и часовой пояс. Это нужно для того, что бы мы понимали когда у вас наступают следующие сутки и могли разделять расходы по дням. Вы сможете изменить эту настройку в будущем")
	msg.ReplyToMessageID = message.MessageID

	a.waitRegisterMessageWithCountry = msg.ReplyToMessageID + 2

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Belarus (GMT+3)"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Russia, Moscow (GMT+3)"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Poland (GMT+2)"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Ukraine (GMT+3)"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Georgia (GMT+4)"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Sri Lanka (GMT+5.30)"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("USA, California (GMT-7)"),
		),
	)
	msg.ReplyMarkup = keyboard

	_, err := a.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("sendMessage, telegram bot couldn't send message: %v", err)
	}
	return nil
}

func (a *Auth) sendMessage(message *tgbotapi.Message, text string) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyToMessageID = message.MessageID

	_, err := a.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("sendMessage, telegram bot couldn't send message: %v", err)
	}
	return nil
}

func (a *Auth) validate(value string, tags string) bool {
	err := a.validator.Var(value, tags)
	if err != nil {
		return false
	}
	return true
}
