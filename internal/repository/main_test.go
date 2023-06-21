package repository

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	postgresPool *pgxpool.Pool
	mongoCli     *mongo.Client
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := dockertest.NewPool("")
	if err != nil {
		logrus.Fatalf("Could not construct pool: %s", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		logrus.Fatalf("Could not connect to Docker: %s", err)
	}

	mongoResource := initialMongo(ctx, pool)
	postgresResource := initialPostgres(ctx, pool)

	// run tests
	code := m.Run()
	purgeResources(pool, mongoResource, postgresResource)
	os.Exit(code)
}

func purgeResources(dockerPool *dockertest.Pool, resources ...*dockertest.Resource) {
	for i := range resources {
		if err := dockerPool.Purge(resources[i]); err != nil {
			logrus.Errorf("Could not purge resource: %s", err.Error())
		}

		err := resources[i].Expire(1)
		if err != nil {
			logrus.Error(err.Error())
		}
	}
}

func initialPostgres(ctx context.Context, pool *dockertest.Pool) *dockertest.Resource {
	resource, err := pool.Run("postgres", "14.1-alpine", []string{"POSTGRES_PASSWORD=password123"})
	if err != nil {
		logrus.Fatalf("Could not start resource: %s", err)
	}

	var dbHostAndPort string

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = pool.Retry(func() error {
		dbHostAndPort = resource.GetHostPort("5432/tcp")

		postgresPool, err = pgxpool.Connect(ctx, fmt.Sprintf("postgresql://postgres:password123@%v/postgres", dbHostAndPort))
		if err != nil {
			return err
		}

		return postgresPool.Ping(ctx)
	})
	if err != nil {
		logrus.Fatalf("Could not connect to database: %s", err)
	}

	authRepo = NewPostgres(postgresPool)

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

	return resource
}

func initialMongo(ctx context.Context, pool *dockertest.Pool) *dockertest.Resource {
	// pull mongodb docker image for version 5.0
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mongo",
		Tag:        "latest",
		Env: []string{
			// username and password for mongodb superuser
			"MONGO_INITDB_ROOT_USERNAME=root",
			"MONGO_INITDB_ROOT_PASSWORD=password",
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		logrus.Fatalf("Could not start resource: %s", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	err = pool.Retry(func() error {
		uri := fmt.Sprintf("mongodb://root:password@localhost:%s", resource.GetPort("27017/tcp"))
		logrus.Infof("mongo URI: %s", uri)
		mongoCli, err = mongo.Connect(
			ctx,
			options.Client().ApplyURI(uri),
		)
		if err != nil {
			return err
		}
		financeRepo = NewMongo(mongoCli)
		return mongoCli.Ping(ctx, nil)
	})
	if err != nil {
		logrus.Fatalf("Could not connect to docker: %s", err)
	}

	//// disconnect mongodb client
	//if err = cli.Disconnect(ctx); err != nil {
	//	panic(err)
	//}

	return resource
}
