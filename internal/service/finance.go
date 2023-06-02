package service

import (
	"context"
	"fmt"
	"time"

	"github.com/chucky-1/finance/internal/model"
	"github.com/chucky-1/finance/internal/repository"
)

type Finance interface {
	Add(ctx context.Context, entry *model.Entry) error
}

type Cleaner interface {
	ClearYesterday(ctx context.Context) error
}

type finance struct {
	repo *repository.Finance
}

func NewFinance(repo *repository.Finance) *finance {
	return &finance{
		repo: repo,
	}
}

func (f *finance) Add(ctx context.Context, entry *model.Entry) error {
	tp := entry.Type

	if err := f.repo.Add(ctx, entry); err != nil {
		return err
	}

	entry.Type = fmt.Sprintf("today_%s", tp)
	return f.repo.Add(ctx, entry)
}

func (f *finance) ClearYesterday(ctx context.Context) error {
	return f.repo.ClearYesterday(ctx, &model.Entry{
		Type: "today_expenses",
		Date: time.Now().UTC().Add(-time.Hour),
	})
}
