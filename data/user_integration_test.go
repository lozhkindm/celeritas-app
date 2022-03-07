//go:build integration

package data

import (
	"testing"
)

var user = User{
	FirstName: "Ignat",
	LastName:  "Senkin",
	Email:     "user@integration.test",
	Active:    1,
	Password:  "password",
}

func TestUser_Table(t *testing.T) {
	if tbl := Mds.Users.Table(); tbl != "users" {
		t.Errorf("wrong table name: %s", tbl)
	}
}

func TestUser_Insert(t *testing.T) {
	id, err := Mds.Users.Insert(&user)
	if err != nil {
		t.Fatalf("failed to insert a user: %s", err)
	}
	if id == 0 {
		t.Error("0 returned as user id")
	}
}

func TestUser_GetById(t *testing.T) {
	u, err := Mds.Users.GetById(1)
	if err != nil {
		t.Errorf("failed to get a user: %s", err)
	}
	if u.ID == 0 {
		t.Error("0 returned as user id")
	}
}

func TestUser_GetByEmail(t *testing.T) {
	u, err := Mds.Users.GetByEmail(user.Email)
	if err != nil {
		t.Errorf("failed to get a user: %s", err)
	}
	if u.ID == 0 {
		t.Error("0 returned as user id")
	}
}

func TestUser_GetAll(t *testing.T) {
	_, err := Mds.Users.GetAll()
	if err != nil {
		t.Errorf("failed to get all users: %s", err)
	}
}

func TestUser_Update(t *testing.T) {
	u, err := Mds.Users.GetById(1)
	if err != nil {
		t.Errorf("failed to get a user: %s", err)
	}

	u.FirstName = "Senka"
	err = Mds.Users.Update(u)
	if err != nil {
		t.Errorf("failed to update a user: %s", err)
	}

	u, err = Mds.Users.GetById(1)
	if err != nil {
		t.Errorf("failed to get a user: %s", err)
	}
	if u.FirstName != "Senka" {
		t.Error("user is not updated")
	}
}

func TestUser_CheckPassword(t *testing.T) {
	u, err := Mds.Users.GetById(1)
	if err != nil {
		t.Errorf("failed to get a user: %s", err)
	}

	match, err := u.CheckPassword("password")
	if err != nil {
		t.Errorf("failed to check a password: %s", err)
	}
	if !match {
		t.Error("password does not match when it should")
	}

	match, err = u.CheckPassword("should_not_match")
	if err != nil {
		t.Errorf("failed to check a password: %s", err)
	}
	if match {
		t.Error("password matches when it should not")
	}
}

func TestUser_ResetPassword(t *testing.T) {
	if err := Mds.Users.ResetPassword(1, "new_password"); err != nil {
		t.Errorf("failed to reset a password: %s", err)
	}
	if err := Mds.Users.ResetPassword(999, "new_password"); err == nil {
		t.Error("err is nil when it should not be")
	}
}

func TestUser_Delete(t *testing.T) {
	if err := Mds.Users.Delete(1); err != nil {
		t.Errorf("failed to delete a user: %s", err)
	}
	if _, err := Mds.Users.GetById(1); err == nil {
		t.Error("err is nil when it should not be")
	}
}
