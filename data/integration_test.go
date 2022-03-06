package data

import (
	"database/sql"
	"fmt"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"log"
	"os"
	"testing"
)

var (
	dbHost     = "localhost"
	dbUser     = "postgres"
	dbPassword = "secret"
	dbName     = "celeritas_test"
	dbPort     = "5435"
	dsn        = "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable timezone=UTC connect_timeout=5"
)

var dummyUser = User{
	FirstName: "Ignat",
	LastName:  "Senkin",
	Email:     "ignat@senkin.dog",
	Active:    1,
	Password:  "password",
}

var (
	models   Models
	testDB   *sql.DB
	resource *dockertest.Resource
	pool     *dockertest.Pool
)

func TestMain(m *testing.M) {
	_ = os.Setenv("DATABASE_TYPE", "postgres")

	p, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not connect to docker: %s", err)
	}

	pool = p
	opts := dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "13.4",
		Env: []string{
			fmt.Sprintf("POSTGRES_USER=%s", dbUser),
			fmt.Sprintf("POSTGRES_PASSWORD=%s", dbPassword),
			fmt.Sprintf("POSTGRES_DB=%s", dbName),
		},
		ExposedPorts: []string{"5432"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"5432": {
				{
					HostIP:   "0.0.0.0",
					HostPort: dbPort,
				},
			},
		},
	}

	resource, err = pool.RunWithOptions(&opts)
	if err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("could not start resource: %s", err)
	}

	if err := pool.Retry(func() error {
		var err error
		testDB, err = sql.Open("pgx", fmt.Sprintf(dsn, dbHost, dbPort, dbUser, dbPassword, dbName))
		if err != nil {
			return err
		}
		return testDB.Ping()
	}); err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("could not connect to docker: %s", err)
	}

	os.Exit(m.Run())
}
