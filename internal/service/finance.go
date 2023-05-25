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
	err := f.repo.Add(ctx, entry)
	if err != nil {
		return err
	}
	entry.Type = fmt.Sprintf("today_%s", entry.Type)
	return f.repo.Add(ctx, entry)
}
