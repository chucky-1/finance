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
	users, ok := r.timezones.get(time.Now().UTC().Hour())
	if !ok {
		return nil, nil
	}
	reports, err := r.getter.GetByUsers(ctx, users, "expenses", dailyCollection)
	if err != nil {
		return nil, err
	}
	err = r.cleaner.DeleteByUsers(ctx, users, "expenses", dailyCollection)
	if err != nil {
		return nil, err
	}
	return reports, nil
}

func (r *Reporter) MonthlyReport(ctx context.Context) (map[string]map[string]float64, error) {
	users, ok := r.timezones.get(time.Now().UTC().Hour())
	if !ok {
		return nil, nil
	}
	reports, err := r.getter.GetByUsers(ctx, users, "expenses", time.Now().UTC().Add(-time.Hour).Format(monthlyCollection))
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

func (t *timezones) get(key int) ([]string, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	value, ok := t.timezones[key]
	if !ok {
		return nil, false
	}
	return value, true
}
