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
	timezones map[int][]string
}

func NewReporter(getter repository.Getter, cleaner repository.Cleaner) *Reporter {
	return &Reporter{
		getter:  getter,
		cleaner: cleaner,
		timezones: &timezones{
			timezones: make(map[int][]string),
		},
	}
}

func (r *Reporter) AddTimezone(timezone int, username string) {
	r.timezones.add(timezone, username)
}

func (r *Reporter) DailyReport(ctx context.Context) (map[string]map[string]float64, error) {
	users := r.timezones.getIfTheDayChanges(time.Now().UTC().Hour())
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
	users := r.timezones.getIfTheDayChanges(time.Now().UTC().Hour())
	if len(users) == 0 {
		return make(map[string]map[string]float64), nil
	}
	reports, err := r.getter.GetByUsers(ctx, users, "expenses", time.Now().UTC().Add(24*-time.Hour).Format(monthlyPeriod))
	if err != nil {
		return nil, err
	}
	return reports, nil
}

func (t *timezones) add(key int, value string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.timezones[key] = append(t.timezones[key], value)
}

func (t *timezones) getIfTheDayChanges(hourByUTC int) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var values []string
	if hourByUTC >= 12 {
		v, ok := t.timezones[24-hourByUTC]
		if ok {
			values = append(values, v...)
		}
	}
	if hourByUTC <= 12 {
		v, ok := t.timezones[-hourByUTC]
		if ok {
			values = append(values, v...)
		}
	}
	return values
}
