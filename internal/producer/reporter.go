package producer

import (
	"context"
	"fmt"
	"github.com/chucky-1/finance/internal/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	monthPeriod = "month"
	dayPeriod   = "day"
)

type TGUser struct {
	TGUsername string
	Username   string
}

type Reporter struct {
	dailyReporterBot    *tgbotapi.BotAPI
	dailySubscription   tgbotapi.UpdatesChannel
	monthlyReporterBot  *tgbotapi.BotAPI
	monthlySubscription tgbotapi.UpdatesChannel

	reporter *service.Reporter

	// receiving from hub consumer
	// key: tgUserName, value: username
	// TODO clear map after subscription
	tgUsersChan              <-chan TGUser
	expectedUsersToSubscribe map[string]string

	// key: username, value: chatID
	dailyChatsByUser   map[string]int64
	monthlyChatsByUser map[string]int64
}

func NewReporter(dailyReporterBot, monthlyReporterBot *tgbotapi.BotAPI, dailySubscription, monthlySubscription tgbotapi.UpdatesChannel,
	reporter *service.Reporter, tgUsersChan chan TGUser) *Reporter {
	return &Reporter{
		dailyReporterBot:         dailyReporterBot,
		dailySubscription:        dailySubscription,
		monthlyReporterBot:       monthlyReporterBot,
		monthlySubscription:      monthlySubscription,
		reporter:                 reporter,
		tgUsersChan:              tgUsersChan,
		expectedUsersToSubscribe: make(map[string]string),
		dailyChatsByUser:         make(map[string]int64),
		monthlyChatsByUser:       make(map[string]int64),
	}
}

func (r *Reporter) Produce(ctx context.Context) {
	logrus.Info("reporter producer started produce")
	go r.waitSubscribers(ctx)
	go r.waitTimeToSendReports(ctx)
}

func (r *Reporter) waitSubscribers(ctx context.Context) {
	logrus.Info("reporter producer started wait subscribers")
	for {
		select {
		case <-ctx.Done():
			logrus.Infof("reporter producer stopped wait subscribers: %v", ctx.Err())
			return
		case tgUser := <-r.tgUsersChan:
			logrus.Infof("reporter producer received message to wait for the user's subscriptions: %v", tgUser)
			r.expectedUsersToSubscribe[tgUser.TGUsername] = tgUser.Username
		case update := <-r.dailySubscription:
			logrus.Infof("reporter producer received message in dailySubscription from %s", update.SentFrom().UserName)
			username, ok := r.expectedUsersToSubscribe[update.SentFrom().UserName]
			if !ok {
				logrus.Infof("reporter producer received message in dailySubscription from unknown user: %s", update.SentFrom().UserName)
				continue
			}
			r.dailyChatsByUser[username] = update.Message.Chat.ID
			logrus.Infof("%s subscribed to daily reports", username)
		case update := <-r.monthlySubscription:
			logrus.Infof("reporter producer received message in monthlyUpdatesChan from %s", update.SentFrom().UserName)
			username, ok := r.expectedUsersToSubscribe[update.SentFrom().UserName]
			if !ok {
				logrus.Infof("reporter producer received message in monthlyUpdatesChan from unknown user: %s", update.SentFrom().UserName)
				continue
			}
			r.monthlyChatsByUser[username] = update.Message.Chat.ID
			logrus.Infof("%s subscribed to monthly reports", username)
		}
	}
}

func (r *Reporter) waitTimeToSendReports(ctx context.Context) {
	logrus.Info("reporter producer started wait time to send reports")
	t := tickerFromBeginningOrMiddleOfHour(ctx)
	defer t.Stop()
	timeUTC := time.Now().UTC()
	logrus.Infof("reporter produce created ticker from beginning of hour: %v", timeUTC)

	go r.sendAllReports(ctx, timeUTC)

	for {
		select {
		case <-ctx.Done():
			logrus.Infof("reporter producer stopped wait time to send reports: %v", ctx.Err())
			return
		case <-t.C:
			timeUTC = time.Now().UTC()
			logrus.Infof("reporter producer: ticker triggered in: %v", timeUTC)
			r.sendAllReports(ctx, timeUTC)
		}
	}
}

func (r *Reporter) sendAllReports(ctx context.Context, timeUTC time.Time) {
	if err := r.sendReports(ctx, timeUTC, dayPeriod); err != nil {
		logrus.Error(err)
	}
	if err := r.sendReports(ctx, timeUTC, monthPeriod); err != nil {
		logrus.Error(err)
	}
}

func (r *Reporter) sendReports(ctx context.Context, timeUTC time.Time, period string) error {
	var (
		reports map[string]map[string]float64
		err     error
	)
	switch period {
	case dayPeriod:
		if reports, err = r.reporter.DailyReportsIfDayChanges(ctx, timeUTC); err != nil {
			return fmt.Errorf("reporter producer couldn't get daily report: %v", err)
		}
	case monthPeriod:
		if reports, err = r.reporter.MonthlyReportsIfMonthChanges(ctx, timeUTC); err != nil {
			return fmt.Errorf("reporter producer couldn't get monthly report: %v", err)
		}
	}
	tgReports := convertToTGReports(reports, timeUTC, period)
	for user, report := range tgReports {
		if err = r.sendReport(user, report); err != nil {
			logrus.Error(err)
		}
	}
	return nil
}

func (r *Reporter) sendReport(user, report string) error {
	chatID, ok := r.dailyChatsByUser[user]
	if !ok {
		logrus.Infof("couldn't send a report because don't have a chat with user, user didn't subscribe: %s", user)
		return nil
	}
	message := tgbotapi.NewMessage(chatID, report)
	_, err := r.dailyReporterBot.Send(message)
	if err != nil {
		return fmt.Errorf("reporter producer couldn't send report: %v", err)
	}
	return nil
}

func tickerFromBeginningOrMiddleOfHour(ctx context.Context) *time.Ticker {
	timer := time.NewTimer(durationBeforeCreateTicker(time.Now().UTC()))
	select {
	case <-ctx.Done():
		return &time.Ticker{}
	case <-timer.C:
		return time.NewTicker(30 * time.Minute)
	}
}

func durationBeforeCreateTicker(timeUTC time.Time) time.Duration {
	return timeUTC.Truncate(30 * time.Minute).Add(30 * time.Minute).Sub(timeUTC)
}

func convertToTGReports(reports map[string]map[string]float64, timeUTC time.Time, period string) map[string]string {
	year, month, day := timeUTC.Add(-time.Hour).Date()
	var reportTitle string
	switch period {
	case dayPeriod:
		reportTitle = fmt.Sprintf("%d %s\n", day, month.String())
	case monthPeriod:
		reportTitle = fmt.Sprintf("%s %d\n", month.String(), year)
	}
	tgReports := make(map[string]string)
	for user, entry := range reports {
		report := reportTitle
		for item, sum := range entry {
			report += fmt.Sprintf("%s - %.2f\n", item, sum)
		}
		tgReports[user] = report
	}
	return tgReports
}
