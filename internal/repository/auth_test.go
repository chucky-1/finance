package repository

import (
	"context"
	"fmt"
	"github.com/chucky-1/finance/internal/model"
	"github.com/stretchr/testify/require"
	"os"
	"os/exec"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/sirupsen/logrus"
)

var (
	dbPool   *pgxpool.Pool
	authRepo *Auth
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		logrus.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.Run("postgres", "14.1-alpine", []string{"POSTGRES_PASSWORD=password123"})
	if err != nil {
		logrus.Fatalf("Could not start resource: %s", err)
	}

	var dbHostAndPort string

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = pool.Retry(func() error {
		dbHostAndPort = resource.GetHostPort("5432/tcp")

		dbPool, err = pgxpool.Connect(ctx, fmt.Sprintf("postgresql://postgres:password123@%v/postgres", dbHostAndPort))
		if err != nil {
			return err
		}

		return dbPool.Ping(ctx)
	})
	if err != nil {
		logrus.Fatalf("Could not connect to database: %s", err)
	}

	authRepo = NewAuth(dbPool)

	cmd := exec.Command("flyway",
		"-user=postgres",
		"-password=password123",
		"-locations=filesystem:../../migrations",
		fmt.Sprintf("-url=jdbc:postgresql://%v/postgres", dbHostAndPort),
		"migrate")

	err = cmd.Run()
	if err != nil {
		logrus.Fatalf("There are errors in migrations: %s", err)
	}

	code := m.Run()

	if err = pool.Purge(resource); err != nil {
		logrus.Fatalf("Could not purge resource: %s", err)
	}

	err = resource.Expire(1)
	if err != nil {
		logrus.Fatal(err)
	}

	os.Exit(code)
}

func TestAuth_CreateUser(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		_, err := dbPool.Exec(ctx, `TRUNCATE TABLE finance.users`)
		if err != nil {
			t.Fatal(err)
		}
	}()

	user := model.User{
		Username: "user",
		Password: "secret",
	}
	err := authRepo.CreateUser(ctx, &user)
	if err != nil {
		t.Fatal(err)
	}

	var u model.User
	err = dbPool.QueryRow(ctx, `SELECT username, password FROM finance.users WHERE username='user'`).Scan(&u.Username, &u.Password)
	if err != nil {
		t.Fatal(err)
	}

	logrus.Infof("recieved user: %v", u)
	require.Equal(t, user, u)
}