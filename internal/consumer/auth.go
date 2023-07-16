package consumer

import (
	"context"
	"fmt"
	"github.com/chucky-1/finance/internal/model"
	"github.com/chucky-1/finance/internal/repository"
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

var explainingSubscriptionMessage = "Если вы хотите получать отчёты, отправьте команду \"start\" следующим ботам\n\n" +
	"Для получения ежедневных отчётов\n" +
	"%s\n" +
	"Для получения ежемесячных отчётов\n" +
	"%s\n\n" +
	"Эти боты не смогут с вами коммуницировать, они предназначены ТОЛЬКО для отчётов. Вся коммуникация с приложением осуществляется через этот чат\n"

var explainingCommunicationMessage = "Для того что бы записать расходы, вы должны отправить сообщение в формате\n\n" +
	"Кофе 3.5\n\n" +
	"Вы должны отправить мне только 2 слова, точнее одно слово и одну цифру, через пробел, иначе я не смогу обработать сообщение и буду ругаться :)\n" +
	"Приятного пользования :)"

var chooseCountryMessage = "Выберете свою страну и часовой пояс. " +
	"Это нужно для того, что бы мы понимали когда у вас наступают следующие сутки и могли разделять расходы по дням. " +
	"Вы сможете изменить эту настройку в будущем.\n\n" +
	"Пока мы работаем в бета версии, страну можно выбрать только из списка предложенных."

type finishData struct {
	username   string
	chatID     int64
	tgUsername string
}

type Auth struct {
	bot                      *tgbotapi.BotAPI
	updatesChan              chan tgbotapi.Update
	validator                *validator.Validate
	auth                     service.Authorization
	reporter                 *service.Reporter
	finish                   chan<- *finishData
	tgNameDailyReporterBot   string
	tgNameMonthlyReporterBot string

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
	reporter *service.Reporter, finish chan<- *finishData, TGNameDailyReporterBot, TGNameMonthlyReporterBot string) *Auth {
	return &Auth{
		bot:                      bot,
		updatesChan:              updatesChan,
		validator:                validator,
		auth:                     auth,
		reporter:                 reporter,
		finish:                   finish,
		tgNameDailyReporterBot:   TGNameDailyReporterBot,
		tgNameMonthlyReporterBot: TGNameMonthlyReporterBot,
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
				success, err := a.handleUsername(register, update.Message)
				if err != nil {
					logrus.Errorf("register error: %v", err)
					continue
				}
				if !success {
					continue
				}

				if err = a.requestForCountry(update.Message); err != nil {
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
					fmt.Sprintf("Введите пароль. Максимум %d символов", passwordMaxLength)); err != nil {
					logrus.Errorf("register error: %v", err)
					continue
				}
				continue
			}

			if !update.Message.IsCommand() && update.Message.MessageID == a.waitRegisterMessageWithPassword {
				success, err := a.handlePassword(register, update.Message)
				if err != nil {
					logrus.Errorf("register error: %v", err)
					continue
				}
				if !success {
					continue
				}

				newCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
				err = a.auth.Register(newCtx, &model.User{
					Username: a.username,
					Password: a.password,
					Country:  a.country,
					Timezone: a.timezone,
				})
				if err != nil && err != repository.DuplicateUserErr {
					logrus.Errorf("register error: %v", err)
					cancel()
					continue
				} else if err == repository.DuplicateUserErr {
					logrus.Debugf("user %s already exist", a.username)
					if err = a.requestForUsername(register, update.Message,
						fmt.Sprintf("Имя пользователя: %s уже существует. Попробуйте ещё раз! Введите ваше имя пользователя", a.username)); err != nil {
						logrus.Errorf("register error: %v", err)
						cancel()
						continue
					}
					cancel()
					continue
				}
				cancel()

				a.reporter.AddTimezone(a.timezone, a.username)

				if err = a.sendMessage(update.Message, fmt.Sprintf("Спасибо, %s! Вы успешно зарегистрировались", a.username)); err != nil {
					logrus.Errorf("register error: %v", err)
					continue
				}

				if err = a.sendMessage(update.Message, fmt.Sprintf(explainingSubscriptionMessage, a.tgNameDailyReporterBot, a.tgNameMonthlyReporterBot)); err != nil {
					logrus.Errorf("register error: coldn't send explanation subscribe message: %v", err)
				}

				if err = a.sendMessage(update.Message, explainingCommunicationMessage); err != nil {
					logrus.Errorf("register error: coldn't send explanation comminicate message: %v", err)
				}

				logrus.Debugf("user %s successful registered", a.username)
				logrus.Debugf("auth consumer for user %s stopped", a.username)
				a.finish <- &finishData{
					username:   a.username,
					chatID:     update.Message.Chat.ID,
					tgUsername: update.SentFrom().UserName,
				}
				return
			}

			if !update.Message.IsCommand() && update.Message.MessageID == a.waitLoginMessageWithUsername {
				success, err := a.handleUsername(login, update.Message)
				if err != nil {
					logrus.Errorf("login error: %v", err)
					continue
				}
				if !success {
					continue
				}

				if err := a.requestForPassword(login, update.Message, "Введите пароль"); err != nil {
					logrus.Errorf("login error: %v", err)
					continue
				}
				continue
			}

			if !update.Message.IsCommand() && update.Message.MessageID == a.waitLoginMessageWithPassword {
				success, err := a.handlePassword(login, update.Message)
				if err != nil {
					logrus.Errorf("login error: %v", err)
					continue
				}
				if !success {
					continue
				}

				newCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
				user, err := a.auth.Login(newCtx, a.username, a.password)
				if err != nil && err != service.UserNotFoundErr && err != service.WrongPasswordErr {
					logrus.Errorf("login error: %v", err)
					cancel()
					continue
				} else if err == service.UserNotFoundErr {
					logrus.Debugf("user %s already exists", a.username)
					if err = a.requestForUsername(login, update.Message, "Пользователь с таким именем не найден. Попробуйте ещё раз! Введите имя пользователя"); err != nil {
						logrus.Errorf("login error: %v", err)
						cancel()
						continue
					}
					cancel()
					continue
				} else if err == service.WrongPasswordErr {
					logrus.Debugf("user %s entered the wrong password %s", a.username, a.password)
					if err = a.requestForUsername(login, update.Message, "Вы ввели неверное имя пользователя или пароль. Попробуйте ещё раз! Введите ваше имя пользователя"); err != nil {
						logrus.Errorf("login error: %v", err)
						cancel()
						continue
					}
					cancel()
					continue
				}
				cancel()

				a.reporter.AddTimezone(user.Timezone, user.Username)

				if err = a.sendMessage(update.Message, fmt.Sprintf("%s, вы авторизованы!", a.username)); err != nil {
					logrus.Errorf("login error: %v", err)
					continue
				}

				if err = a.sendMessage(update.Message, fmt.Sprintf(explainingSubscriptionMessage, a.tgNameDailyReporterBot, a.tgNameMonthlyReporterBot)); err != nil {
					logrus.Errorf("login error: coldn't send explanation subscribe message: %v", err)
				}

				if err = a.sendMessage(update.Message, explainingCommunicationMessage); err != nil {
					logrus.Errorf("login error: coldn't send explanation communicate message: %v", err)
				}

				logrus.Debug("user %s is authorized", a.username)
				logrus.Debug("auth consumer for user %s stopped", a.username)
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
					logrus.Debug("register command started executing")
					if err := a.requestForUsername(register, update.Message,
						fmt.Sprintf("Введите имя пользователя. Минимум %d, максимум %d символов", usernameMinLength, usernameMaxLength)); err != nil {
						logrus.Errorf("register error: %v", err)
						continue
					}
					continue
				case login:
					logrus.Debug("login command started executing")
					if err := a.requestForUsername(login, update.Message, "Введите имя пользователя"); err != nil {
						logrus.Errorf("login error: %v", err)
						continue
					}
					continue
				}
			}
		}
	}
}

func (a *Auth) handleUsername(action string, message *tgbotapi.Message) (bool, error) {
	a.username = message.Text
	if !a.validate(a.username, fmt.Sprintf("min=%d,max=%d", usernameMinLength, usernameMaxLength)) {
		err := a.requestForUsername(action, message, "Вы ввели некорректное имя пользователя. Попробуйте ещё раз!")
		if err != nil {
			return false, err
		}
		logrus.Debugf("%s, user entered the wrong username: %s", action, a.username)
		return false, nil
	}
	logrus.Debugf("%s, user entered username: %s", action, a.username)
	return true, nil
}

func (a *Auth) handlePassword(action string, message *tgbotapi.Message) (bool, error) {
	a.password = message.Text
	if !a.validate(a.password, fmt.Sprintf("max=%d", passwordMaxLength)) {
		err := a.requestForPassword(action, message, fmt.Sprintf("%s, вы ввели некорректный пароль. Попробуйте ещё раз!", a.username))
		if err != nil {
			return false, err
		}
		logrus.Debugf("%s, user %s entered the wrong password: %s", action, a.username, a.password)
		return false, nil
	}
	logrus.Debugf("%s, user %s entered password: %s", action, a.username, a.password)
	return true, nil
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
	logrus.Debugf("%s chose country: %s and timezone: %v", a.username, a.country, a.timezone)
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
	msg := tgbotapi.NewMessage(message.Chat.ID, chooseCountryMessage)
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
