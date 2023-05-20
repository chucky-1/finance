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
	Login(ctx context.Context, user *model.User) (bool, error)
}

type Auth struct {
	repo repository.User
	salt string
}

func NewAuth(repo repository.User, salt string) *Auth {
	return &Auth{
		repo: repo,
		salt: salt,
	}
}

func (a *Auth) CreateUser(ctx context.Context, user *model.User) error {
	user.Password = a.generatePassword(user.Password)
	return a.repo.Create(ctx, user)
}

func (a *Auth) Login(ctx context.Context, user *model.User) (bool, error) {
	repUser, err := a.repo.Get(ctx, user.Username)
	if err != nil {
		return false, err
	}
	if repUser == nil {
		return false, nil
	}
	user.Password = a.generatePassword(user.Password)
	if user.Password != repUser.Password {
		return false, nil
	}
	return true, nil
}

func (a *Auth) generatePassword(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))
	return fmt.Sprintf("%x", hash.Sum([]byte(a.salt)))
}
