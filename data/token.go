package data

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"net/http"
	"strings"
	"time"

	up "github.com/upper/db/v4"
)

type Token struct {
	ID        int       `db:"id,omitempty" json:"id"`
	UserID    int       `db:"user_id" json:"user_id"`
	FirstName string    `db:"first_name" json:"first_name"`
	Email     string    `db:"email" json:"email"`
	PlainText string    `db:"token" json:"token"`
	Hash      []byte    `db:"token_hash" json:"-"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	Expires   time.Time `db:"expiry" json:"expiry"`
}

func (t *Token) Table() string {
	return "tokens"
}

func (t *Token) GetUser(token string) (*User, error) {
	var usr *User
	var tkn Token

	coll := upper.Collection(t.Table())
	res := coll.Find(up.Cond{"token": token})
	if err := res.One(&tkn); err != nil {
		return nil, err
	}

	coll = upper.Collection(usr.Table())
	res = coll.Find(tkn.UserID)
	if err := res.One(&usr); err != nil {
		return nil, err
	}

	usr.Token = tkn
	return usr, nil
}

func (t *Token) GetUserTokens(userID int) ([]*Token, error) {
	var tokens []*Token
	coll := upper.Collection(t.Table())
	res := coll.Find(up.Cond{"user_id": userID})
	if err := res.All(&tokens); err != nil {
		return nil, err
	}
	return tokens, nil
}

func (t *Token) Get(id int) (*Token, error) {
	var token *Token
	coll := upper.Collection(t.Table())
	res := coll.Find(id)
	if err := res.One(&token); err != nil {
		return nil, err
	}
	return token, nil
}

func (t *Token) GetByToken(token string) (*Token, error) {
	var tkn *Token
	coll := upper.Collection(t.Table())
	res := coll.Find(up.Cond{"token": token})
	if err := res.One(&tkn); err != nil {
		return nil, err
	}
	return tkn, nil
}

func (t *Token) DeleteById(id int) error {
	coll := upper.Collection(t.Table())
	res := coll.Find(id)
	if err := res.Delete(); err != nil {
		return err
	}
	return nil
}

func (t *Token) Delete(token string) error {
	coll := upper.Collection(t.Table())
	res := coll.Find(up.Cond{"token": token})
	if err := res.Delete(); err != nil {
		return err
	}
	return nil
}

func (t *Token) Insert(token *Token, user User) error {
	coll := upper.Collection(t.Table())
	res := coll.Find(up.Cond{"user_id": user.ID})
	if err := res.Delete(); err != nil {
		return err
	}

	token.FirstName = user.FirstName
	token.Email = user.Email
	token.CreatedAt = time.Now()
	token.UpdatedAt = time.Now()

	if _, err := coll.Insert(token); err != nil {
		return err
	}
	return nil
}

func (t *Token) Generate(userID int, ttl time.Duration) (*Token, error) {
	token := &Token{
		UserID:  userID,
		Expires: time.Now().Add(ttl),
	}

	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}

	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes)
	hash := sha256.Sum256([]byte(token.PlainText))
	token.Hash = hash[:]
	return token, nil
}

func (t *Token) Authenticate(r *http.Request) (*User, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return nil, errors.New("no authorization header provided")
	}

	parts := strings.Split(header, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, errors.New("no authorization header provided")
	}

	token := parts[1]
	if len(token) != 26 {
		return nil, errors.New("wrong token size")
	}

	tkn, err := t.GetByToken(token)
	if err != nil {
		return nil, errors.New("no matching token found")
	}

	if tkn.Expires.Before(time.Now()) {
		return nil, errors.New("expired token")
	}

	user, err := t.GetUser(token)
	if err != nil {
		return nil, errors.New("no matching user found")
	}
	return user, nil
}

func (t *Token) Validate(token string) (bool, error) {
	user, err := t.GetUser(token)
	if err != nil {
		return false, errors.New("no matching user found")
	}
	if user.Token.PlainText == "" {
		return false, errors.New("no matching token found")
	}
	if user.Token.Expires.Before(time.Now()) {
		return false, errors.New("expired token")
	}
	return true, nil
}
