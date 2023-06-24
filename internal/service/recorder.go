package service

import (
	"context"
	"github.com/chucky-1/finance/internal/model"
	"github.com/chucky-1/finance/internal/repository"
)

const (
	monthlyPeriod = "2006-01"
	dailyPeriod   = "today"
)

type Recorder struct {
	repo repository.Recorder
}

func NewRecorder(repo repository.Recorder) *Recorder {
	return &Recorder{
		repo: repo,
	}
}

func (f *Recorder) Add(ctx context.Context, entry *model.Entry) error {
	if err := f.repo.Add(ctx, entry, entry.Date.Format(monthlyPeriod)); err != nil {
		return err
	}
	return f.repo.Add(ctx, entry, dailyPeriod)
}
