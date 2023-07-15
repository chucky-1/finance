package service

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/chucky-1/finance/internal/model"
	"github.com/chucky-1/finance/internal/repository"
)

var (
	UserNotFoundErr  = errors.New("user not found")
	WrongPasswordErr = errors.New("wrong password")
)

type Authorization interface {
	Register(ctx context.Context, user *model.User) error
	Login(ctx context.Context, username, password string) (*model.User, error)
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

func (a *Auth) Register(ctx context.Context, user *model.User) error {
	user.Password = a.generatePassword(user.Password)
	return a.repo.Create(ctx, user)
}

func (a *Auth) Login(ctx context.Context, username, password string) (*model.User, error) {
	user, err := a.repo.Get(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, UserNotFoundErr
	}
	password = a.generatePassword(password)
	if password != user.Password {
		return nil, WrongPasswordErr
	}
	return user, nil
}

func (a *Auth) generatePassword(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))
	return fmt.Sprintf("%x", hash.Sum([]byte(a.salt)))
}
