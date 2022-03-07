//go:build unit

package data

import (
	"fmt"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	udb "github.com/upper/db/v4"
)

func TestNew(t *testing.T) {
	fakeDB, _, _ := sqlmock.New()
	defer func() {
		_ = fakeDB.Close()
	}()

	_ = os.Setenv("DATABASE_TYPE", "postgres")
	m := New(fakeDB)
	if fmt.Sprintf("%T", m) != "data.Models" {
		t.Errorf("Wrong type: %T", m)
	}

	_ = os.Setenv("DATABASE_TYPE", "mysql")
	m = New(fakeDB)
	if fmt.Sprintf("%T", m) != "data.Models" {
		t.Errorf("Wrong type: %T", m)
	}
}

func TestGetInsertedID(t *testing.T) {
	var id udb.ID

	id = int64(1)
	returnedID := getInsertedID(id)
	if fmt.Sprintf("%T", returnedID) != "int" {
		t.Errorf("Wrong type: %T", returnedID)
	}

	id = 1
	returnedID = getInsertedID(id)
	if fmt.Sprintf("%T", returnedID) != "int" {
		t.Errorf("Wrong type: %T", returnedID)
	}
}
