package repository

import (
	"context"
	"testing"
	"time"

	"github.com/chucky-1/finance/internal/model"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

var (
	authRepo *Postgres
)

func TestUserPostgres_CreateGet(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		_, err := postgresPool.Exec(ctx, `TRUNCATE TABLE finance.users`)
		if err != nil {
			t.Fatal(err)
		}
	}()

	user := model.User{
		Username: "user",
		Password: "secret",
		Country:  "Belarus",
		Timezone: 3 * time.Hour,
	}
	success, err := authRepo.Create(ctx, &user)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, true, success)

	u, err := authRepo.Get(ctx, user.Username)

	logrus.Infof("recieved user: %v", u)
	require.Equal(t, &user, u)
}

func TestUserPostgres_CreateGetTimezoneWith30Minute(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		_, err := postgresPool.Exec(ctx, `TRUNCATE TABLE finance.users`)
		if err != nil {
			t.Fatal(err)
		}
	}()

	user := model.User{
		Username: "user",
		Password: "secret",
		Country:  "Belarus",
		Timezone: 3*time.Hour + 30*time.Minute,
	}
	success, err := authRepo.Create(ctx, &user)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, true, success)

	u, err := authRepo.Get(ctx, user.Username)

	logrus.Infof("recieved user: %v", u)
	require.Equal(t, &user, u)
}

func TestUserPostgres_CreateSetDuplicate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		_, err := postgresPool.Exec(ctx, `TRUNCATE TABLE finance.users`)
		if err != nil {
			t.Fatal(err)
		}
	}()

	user := model.User{
		Username: "user",
		Password: "secret",
		Country:  "Belarus",
		Timezone: 3 * time.Hour,
	}
	success, err := authRepo.Create(ctx, &user)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, true, success)

	success, err = authRepo.Create(ctx, &user)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, false, success)
}
