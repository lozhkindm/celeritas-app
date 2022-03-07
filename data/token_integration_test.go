//go:build integration

package data

import (
	"fmt"
	"net/http"
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

var authData = []struct {
	name          string
	token         string
	email         string
	errorExpected bool
	message       string
}{
	{"invalid", "abcdefghijklmnopqrstuvwxyz", "n@e.tkn", true, "invalid token accepted as valid"},
	{"invalid_length", "length", "n@e.tkn", true, "token of wrong length accepted as valid"},
	{"no_user", "abcdefghijklmnopqrstuvwxyz", "n@e.tkn", true, "token of wrong length accepted as valid"},
	{"valid", "", tokenUser.Email, false, "valid token reported as invalid"},
}

func TestToken_Authenticate(t *testing.T) {
	for _, tt := range authData {
		token := tt.token
		if tt.email == tokenUser.Email {
			user, err := Mds.Users.GetByEmail(tt.email)
			if err != nil {
				t.Errorf("failed to get a user: %s", err)
			}
			token = user.Token.PlainText
		}
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		if err != nil {
			t.Errorf("failed to create a request: %s", err)
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

		_, err = Mds.Tokens.Authenticate(req)
		if tt.errorExpected && err == nil {
			t.Errorf("%s: %s", tt.name, tt.message)
		} else if !tt.errorExpected && err != nil {
			t.Errorf("%s: %s: %s", tt.name, tt.message, err)
		}
	}
}

func TestToken_Delete(t *testing.T) {
	u, err := Mds.Users.GetByEmail(tokenUser.Email)
	if err != nil {
		t.Errorf("failed to get a user: %s", err)
	}

	if err := Mds.Tokens.Delete(u.Token.PlainText); err != nil {
		t.Errorf("failed to delete a token: %s", err)
	}
}

func TestToken_Expired(t *testing.T) {
	u, err := Mds.Users.GetByEmail(tokenUser.Email)
	if err != nil {
		t.Errorf("failed to get a user: %s", err)
	}

	tkn, err := Mds.Tokens.Generate(u.ID, -time.Hour)
	if err != nil {
		t.Errorf("failed to generate a token: %s", err)
	}

	if err := Mds.Tokens.Insert(tkn, *u); err != nil {
		t.Errorf("failed to insert a token: %s", err)
	}

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Errorf("failed to create a request: %s", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tkn.PlainText))

	if _, err := Mds.Tokens.Authenticate(req); err == nil {
		t.Error("err is nil when it should not be")
	}
}

func TestToken_BadHeader(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Errorf("failed to create a request: %s", err)
	}

	if _, err := Mds.Tokens.Authenticate(req); err == nil {
		t.Error("err is nil when it should not be")
	}

	req.Header.Add("Authorization", "bad_header")
	if _, err := Mds.Tokens.Authenticate(req); err == nil {
		t.Error("err is nil when it should not be")
	}

	newUser := User{
		FirstName: "New",
		LastName:  "User",
		Email:     "new@user.com",
		Active:    1,
		Password:  "password",
	}

	id, err := Mds.Users.Insert(&newUser)
	if err != nil {
		t.Errorf("failed to insert a user: %s", err)
	}

	tkn, err := Mds.Tokens.Generate(id, time.Hour)
	if err != nil {
		t.Errorf("failed to generate a token: %s", err)
	}

	if err := Mds.Tokens.Insert(tkn, newUser); err != nil {
		t.Errorf("failed to insert a token: %s", err)
	}

	if err := Mds.Users.Delete(id); err != nil {
		t.Errorf("failed to delete a user: %s", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tkn.PlainText))
	if _, err := Mds.Tokens.Authenticate(req); err == nil {
		t.Error("err is nil when it should not be")
	}
}

func TestToken_DeleteNonExistingToken(t *testing.T) {
	if err := Mds.Tokens.Delete("non_existing_token"); err != nil {
		t.Errorf("failed to delete a token: %s", err)
	}
}

func TestToken_ValidToken(t *testing.T) {
	u, err := Mds.Users.GetByEmail(tokenUser.Email)
	if err != nil {
		t.Errorf("failed to get a user: %s", err)
	}

	tkn, err := Mds.Tokens.Generate(u.ID, time.Hour*24)
	if err != nil {
		t.Errorf("failed to generate a token: %s", err)
	}

	if err := Mds.Tokens.Insert(tkn, *u); err != nil {
		t.Errorf("failed to insert a token: %s", err)
	}

	ok, err := Mds.Tokens.Validate(tkn.PlainText)
	if err != nil {
		t.Errorf("failed to validate a token: %s", err)
	}
	if !ok {
		t.Error("valid token reported as invalid")
	}

	ok, err = Mds.Tokens.Validate("invalid_token")
	if err == nil {
		t.Error("err is nil when it should not be")
	}
	if ok {
		t.Error("invalid token reported as valid")
	}

	u, err = Mds.Users.GetByEmail(tokenUser.Email)
	if err != nil {
		t.Errorf("failed to get a user: %s", err)
	}

	if err := Mds.Tokens.DeleteById(u.Token.ID); err != nil {
		t.Errorf("failed to delete a token: %s", err)
	}

	ok, err = Mds.Tokens.Validate(u.Token.PlainText)
	if err == nil {
		t.Error("err is nil when it should not be")
	}
	if ok {
		t.Error("non-existent token reported as valid")
	}
}
