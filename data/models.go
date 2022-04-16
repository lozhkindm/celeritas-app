package data

import (
	"database/sql"
	"fmt"
	"os"

	udb "github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/mysql"
	"github.com/upper/db/v4/adapter/postgresql"
)

var (
	db    *sql.DB
	upper udb.Session
)

type Models struct {
	Users  User
	Tokens Token
}

func New(dbPool *sql.DB) Models {
	db = dbPool

	switch os.Getenv("DATABASE_TYPE") {
	case "mysql", "mariadb":
		upper, _ = mysql.New(db)
	case "postgres", "postgresql":
		upper, _ = postgresql.New(db)
	}

	return Models{
		Users:  User{},
		Tokens: Token{},
	}
}

func getInsertedID(id udb.ID) int {
	t := fmt.Sprintf("%T", id)
	if t == "int64" {
		return int(id.(int64))
	}
	return id.(int)
}
