package service

import (
	"context"
	"sync"
	"time"

	"github.com/chucky-1/finance/internal/repository"
)

type Reporter struct {
	getter    repository.Getter
	cleaner   repository.Cleaner
	timezones *timezones
}

type timezones struct {
	// key: timezone, value: usernames
	mu        sync.RWMutex
	timezones map[time.Duration][]string
}

func NewReporter(getter repository.Getter, cleaner repository.Cleaner) *Reporter {
	return &Reporter{
		getter:  getter,
		cleaner: cleaner,
		timezones: &timezones{
			timezones: make(map[time.Duration][]string),
		},
	}
}

func (r *Reporter) DailyReportsIfDayChanges(ctx context.Context, timeUTC time.Time) (map[string]map[string]float64, error) {
	usernames := r.timezones.getUsersWhoseDayChanges(timeUTC)
	if len(usernames) == 0 {
		return nil, nil
	}
	reports, err := r.getter.GetByUsernames(ctx, usernames, "expenses", dailyPeriod)
	if err != nil {
		return nil, err
	}
	err = r.cleaner.DeleteByUsernames(ctx, usernames, "expenses", dailyPeriod)
	if err != nil {
		return nil, err
	}
	return reports, nil
}

func (r *Reporter) MonthlyReportsIfMonthChanges(ctx context.Context, timeUTC time.Time) (map[string]map[string]float64, error) {
	usernames := r.timezones.getUsersWhoseMonthChanges(timeUTC)
	if len(usernames) == 0 {
		return nil, nil
	}
	reports, err := r.getter.GetByUsernames(ctx, usernames, "expenses", timeUTC.Add(24*-time.Hour).Format(monthlyPeriod))
	if err != nil {
		return nil, err
	}
	return reports, nil
}

func (r *Reporter) AddTimezone(timezone time.Duration, username string) {
	r.timezones.add(timezone, username)
}

// getUsersWhoseDayChanges returns users who have 00:00 local time
// When by UTC 12:00 the day changes in the time zones +12 and -12
func (t *timezones) getUsersWhoseDayChanges(timeUTC time.Time) []string {
	timezone := getTimezoneWhereDayChanges(timeUTC.Hour(), timeUTC.Minute())
	if timezone.Hours() == 12 {
		return append(t.get(timezone), t.get(-timezone)...)
	}
	return t.get(timezone)
}

// getUsersWhoseMonthChanges returns users who have the 1st date
// When by UTC 12:00 the day changes in time zones +12 and -12, but the dates will be different
func (t *timezones) getUsersWhoseMonthChanges(timeUTC time.Time) []string {
	timezone := getTimezoneWhereDayChanges(timeUTC.Hour(), timeUTC.Minute())
	if timezone.Hours() != 12 && timeUTC.Add(timezone).Day() == 1 {
		return t.get(timezone)
	}
	if timezone.Hours() == 12 && timeUTC.Add(timezone).Day() == 1 {
		return t.get(timezone)
	}
	if timezone.Hours() == 12 && timeUTC.Add(-timezone).Day() == 1 {
		return t.get(-timezone)
	}
	return nil
}

func (t *timezones) add(key time.Duration, value string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.timezones[key] = append(t.timezones[key], value)
}

func (t *timezones) get(key time.Duration) []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	users, ok := t.timezones[key]
	if !ok {
		return nil
	}
	return users
}

// getTimezoneWhereDayChanges returns the timezone in which the day changes.
// When by UTC 12:00 the day changes in the timezone +12 and -12, but the function returns only one value +12
func getTimezoneWhereDayChanges(hourUTC, minuteUTC int) time.Duration {
	hourPlusMinute := time.Duration(hourUTC)*time.Hour + time.Duration(minuteUTC)*time.Minute
	if hourPlusMinute >= 12*time.Hour {
		return 24*time.Hour - hourPlusMinute
	}
	if hourPlusMinute < 12*time.Hour {
		return -hourPlusMinute
	}
	return 0
}
