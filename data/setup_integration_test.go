//go:build integration

package data

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var (
	dbHost     = "localhost"
	dbUser     = "postgres"
	dbPassword = "secret"
	dbName     = "celeritas_test"
	dbPort     = "5435"
	dsn        = "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable timezone=UTC connect_timeout=5"
)

var (
	Mds      Models
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
		testDB, err = sql.Open("pgx", fmt.Sprintf(dsn, dbHost, dbPort, dbUser, dbPassword, dbName))
		if err != nil {
			return err
		}
		return testDB.Ping()
	}); err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("could not connect to docker: %s", err)
	}

	if err := createTables(testDB); err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("could not create tables: %s", err)
	}

	Mds = New(testDB)
	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		log.Fatalf("could not purge resource: %s", err)
	}

	os.Exit(code)
}

func createTables(db *sql.DB) error {
	files := []string{"auth.sql"}
	for _, file := range files {
		content, err := ioutil.ReadFile(fmt.Sprintf("../db-sql/%s", file))
		if err != nil {
			return err
		}
		if err := runQuery(db, string(content)); err != nil {
			return err
		}
	}
	return nil
}

func runQuery(db *sql.DB, query string) error {
	if _, err := db.Exec(query); err != nil {
		return err
	}
	return nil
}
