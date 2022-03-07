//go:build integration

package data

import (
	"testing"
	"time"
)

var tokenUser = User{
	FirstName: "Ignat",
	LastName:  "Senkin",
	Email:     "token@integration.test",
	Active:    1,
	Password:  "password",
}

func TestToken_Table(t *testing.T) {
	if tbl := Mds.Tokens.Table(); tbl != "tokens" {
		t.Errorf("wrong table name: %s", tbl)
	}
}

func TestToken_Generate(t *testing.T) {
	uid, err := Mds.Users.Insert(&tokenUser)
	if err != nil {
		t.Errorf("failed to insert a user: %s", err)
	}

	if _, err := Mds.Tokens.Generate(uid, time.Hour*24*365); err != nil {
		t.Errorf("failed to generate a token: %s", err)
	}
}

func TestToken_Insert(t *testing.T) {
	u, err := Mds.Users.GetByEmail(tokenUser.Email)
	if err != nil {
		t.Errorf("failed to get a user: %s", err)
	}

	tkn, err := Mds.Tokens.Generate(u.ID, time.Hour*24*365)
	if err != nil {
		t.Errorf("failed to generate a token: %s", err)
	}

	if err := Mds.Tokens.Insert(tkn, *u); err != nil {
		t.Errorf("failed to insert a token: %s", err)
	}
}

func TestToken_GetUser(t *testing.T) {
	if _, err := Mds.Tokens.GetUser("non_existing_token"); err == nil {
		t.Error("err is nil when it should not be")
	}

	u, err := Mds.Users.GetByEmail(tokenUser.Email)
	if err != nil {
		t.Errorf("failed to get a user: %s", err)
	}

	if _, err := Mds.Tokens.GetUser(u.Token.PlainText); err != nil {
		t.Errorf("failed to get a user: %s", err)
	}
}

func TestToken_GetUserTokens(t *testing.T) {
	tokens, err := Mds.Tokens.GetUserTokens(999)
	if err != nil {
		t.Errorf("failed to get user tokens: %s", err)
	}

	if len(tokens) > 0 {
		t.Error("tokens exist when it should not")
	}
}

func TestToken_Get(t *testing.T) {
	u, err := Mds.Users.GetByEmail(tokenUser.Email)
	if err != nil {
		t.Errorf("failed to get a user: %s", err)
	}

	if _, err := Mds.Tokens.Get(u.Token.ID); err != nil {
		t.Errorf("failed to get a token: %s", err)
	}
}

func TestToken_GetByToken(t *testing.T) {
	u, err := Mds.Users.GetByEmail(tokenUser.Email)
	if err != nil {
		t.Errorf("failed to get a user: %s", err)
	}

	if _, err := Mds.Tokens.GetByToken(u.Token.PlainText); err != nil {
		t.Errorf("failed to get a token: %s", err)
	}
	if _, err := Mds.Tokens.GetByToken("non_existent_token"); err == nil {
		t.Error("err is nil when it should not be")
	}
}
