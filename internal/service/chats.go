package service

import (
	"context"

	"github.com/chucky-1/finance/internal/repository"
)

type Chats interface {
	Add(ctx context.Context, chatID int64, username string) error
	Get(ctx context.Context, chatID int64) (string, error)
}

type chats struct {
	repo repository.Chats
}

func NewChats(repo repository.Chats) *chats {
	return &chats{
		repo: repo,
	}
}

func (c *chats) Add(ctx context.Context, chatID int64, username string) error {
	return c.repo.Add(ctx, chatID, username)
}

func (c *chats) Get(ctx context.Context, chatID int64) (string, error) {
	return c.repo.Get(ctx, chatID)
}
