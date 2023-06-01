package service

import (
	"context"
	"fmt"

	"github.com/chucky-1/finance/internal/model"
	"github.com/chucky-1/finance/internal/repository"
)

type Finance interface {
	Add(ctx context.Context, entry *model.Entry) error
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
	if err := f.repo.Add(ctx, entry); err != nil {
		return err
	}

	entry.Type = fmt.Sprintf("last_%s", tp)
	return f.repo.Replace(ctx, entry)
}
