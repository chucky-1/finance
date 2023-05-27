package service

import (
	"context"
	"fmt"

	"github.com/chucky-1/finance/internal/model"
	"github.com/chucky-1/finance/internal/repository"
)

type Finance struct {
	repo *repository.Finance
}

func NewFinance(repo *repository.Finance) *Finance {
	return &Finance{
		repo: repo,
	}
}

func (f *Finance) Add(ctx context.Context, entry *model.Entry) error {
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
