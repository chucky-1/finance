package repository

import (
	"context"
	"fmt"
)

type Chats interface {
	Add(ctx context.Context, chatID int64, username string) error
	Get(ctx context.Context, chatID int64) (string, error)
}

type ChatsLocalStorage struct {
	m map[int64]string
}

func NewChatsLocalStorage() *ChatsLocalStorage {
	return &ChatsLocalStorage{
		m: make(map[int64]string),
	}
}

func (l *ChatsLocalStorage) Add(_ context.Context, chatID int64, username string) error {
	l.m[chatID] = username
	return nil
}

func (l *ChatsLocalStorage) Get(_ context.Context, chatID int64) (string, error) {
	v, ok := l.m[chatID]
	if !ok {
		return "", fmt.Errorf("repository.ChatsLocalStorage.Get value with key: %d doesn't exist", chatID)
	}
	return v, nil
}
