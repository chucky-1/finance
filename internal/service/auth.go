package service

import (
	"context"
	"crypto/sha1"
	"fmt"
	"github.com/chucky-1/finance/internal/model"
	"github.com/chucky-1/finance/internal/repository"
)

type Authorization interface {
	CreateUser(ctx context.Context, user *model.User) error
}

type Auth struct {
	repo repository.Authorization
	salt string
}

func NewAuth(repo repository.Authorization, salt string) *Auth {
	return &Auth{
		repo: repo,
		salt: salt,
	}
}

func (a *Auth) CreateUser(ctx context.Context, user *model.User) error {
	user.Password = a.generatePassword(user.Password)
	return a.repo.CreateUser(ctx, user)
}

func (a *Auth) generatePassword(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))
	return fmt.Sprintf("%x", hash.Sum([]byte(a.salt)))
}
