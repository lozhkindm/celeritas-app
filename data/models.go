package data

import (
	"database/sql"
	"fmt"
	udb "github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/mysql"
	"github.com/upper/db/v4/adapter/postgresql"
	"os"
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
	dbType := os.Getenv("DATABASE_TYPE")

	if dbType == "mysql" || dbType == "mariadb" {
		upper, _ = mysql.New(db)
	} else {
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
