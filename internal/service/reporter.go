package service

import (
	"context"
	"github.com/chucky-1/finance/internal/repository"
	"sync"
	"time"
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

func (r *Reporter) AddTimezone(timezone time.Duration, username string) {
	r.timezones.add(timezone, username)
}

func (r *Reporter) DailyReport(ctx context.Context) (map[string]map[string]float64, error) {
	utc := time.Now().UTC()
	hourPlusMinute := time.Duration(utc.Hour() + utc.Minute())
	users := r.timezones.getIfTheDayChanges(hourPlusMinute)
	if len(users) == 0 {
		return make(map[string]map[string]float64), nil
	}
	reports, err := r.getter.GetByUsers(ctx, users, "expenses", dailyPeriod)
	if err != nil {
		return nil, err
	}
	err = r.cleaner.DeleteByUsers(ctx, users, "expenses", dailyPeriod)
	if err != nil {
		return nil, err
	}
	return reports, nil
}

func (r *Reporter) MonthlyReport(ctx context.Context) (map[string]map[string]float64, error) {
	utc := time.Now().UTC()
	hourPlusMinute := time.Duration(utc.Hour() + utc.Minute())
	users := r.timezones.getIfTheDayChanges(hourPlusMinute)
	if len(users) == 0 {
		return make(map[string]map[string]float64), nil
	}
	reports, err := r.getter.GetByUsers(ctx, users, "expenses", time.Now().UTC().Add(24*-time.Hour).Format(monthlyPeriod))
	if err != nil {
		return nil, err
	}
	return reports, nil
}

func (t *timezones) add(key time.Duration, value string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.timezones[key] = append(t.timezones[key], value)
}

func (t *timezones) getIfTheDayChanges(tm time.Duration) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var values []string
	if tm >= 12*time.Hour {
		v, ok := t.timezones[24*time.Hour-tm]
		if ok {
			values = append(values, v...)
		}
	}
	if tm <= 12*time.Hour {
		v, ok := t.timezones[-tm]
		if ok {
			values = append(values, v...)
		}
	}
	return values
}
