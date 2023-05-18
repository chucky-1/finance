package repository

import (
	"context"
	"fmt"
	"github.com/chucky-1/finance/internal/model"
	"github.com/jackc/pgx/v4/pgxpool"
)

//go:generate mockery --name=Authorization

type Authorization interface {
	CreateUser(ctx context.Context, user *model.User) error
}

type Auth struct {
	conn *pgxpool.Pool
}

func NewAuth(conn *pgxpool.Pool) *Auth {
	return &Auth{
		conn: conn,
	}
}

func (a *Auth) CreateUser(ctx context.Context, user *model.User) error {
	query := `INSERT INTO finance.users (username, password) VALUES ($1, $2)`
	_, err := a.conn.Exec(ctx, query, user.Username, user.Password)
	if err != nil {
		return fmt.Errorf("authorization repository, create user error: %v", err)
	}
	return nil
}
