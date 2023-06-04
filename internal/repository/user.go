package repository

import (
	"context"
	"fmt"
	"github.com/chucky-1/finance/internal/model"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

//go:generate mockery --name=User

type User interface {
	Create(ctx context.Context, user *model.User) (bool, error)
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

func (u *UserPostgres) Create(ctx context.Context, user *model.User) (bool, error) {
	query := `INSERT INTO finance.users (username, password, country, timezone) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`
	commandTag, err := u.conn.Exec(ctx, query, user.Username, user.Password, user.Country, user.Timezone)
	if err != nil {
		return false, fmt.Errorf("repository.User, create user error: %v", err)
	}
	if commandTag.RowsAffected() != 1 {
		return false, nil
	}
	return true, nil
}

func (u *UserPostgres) Get(ctx context.Context, username string) (*model.User, error) {
	query := `SELECT username, password, country, timezone FROM finance.users WHERE username=$1`
	var user model.User
	err := u.conn.QueryRow(ctx, query, username).Scan(&user.Username, &user.Password, &user.Country, &user.Timezone)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("repository.User, get user error: %v", err)
	} else if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &user, nil
}
