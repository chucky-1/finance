package repository

import (
	"context"
	"fmt"
	"github.com/chucky-1/finance/internal/model"
	"github.com/jackc/pgx/v4/pgxpool"
)

//go:generate mockery --name=Authorization

type User interface {
	Create(ctx context.Context, user *model.User) error
	Get(ctx context.Context, username string) (*model.User, error)
}

type UserPostgres struct {
	conn *pgxpool.Pool
}

func NewUserPostgres(conn *pgxpool.Pool) *UserPostgres {
	return &UserPostgres{
		conn: conn,
	}
}

func (u *UserPostgres) Create(ctx context.Context, user *model.User) error {
	query := `INSERT INTO finance.users (username, password) VALUES ($1, $2)`
	_, err := u.conn.Exec(ctx, query, user.Username, user.Password)
	if err != nil {
		return fmt.Errorf("repository.User, create user error: %v", err)
	}
	return nil
}

func (u *UserPostgres) Get(ctx context.Context, username string) (*model.User, error) {
	query := `SELECT username, password FROM finance.users WHERE username=$1`
	var user model.User
	err := u.conn.QueryRow(ctx, query, username).Scan(&user.Username, &user.Password)
	if err != nil {
		return nil, fmt.Errorf("repository.User, get user error: %v", err)
	}
	return &user, nil
}
