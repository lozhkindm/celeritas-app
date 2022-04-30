package data

import (
	"time"

	up "github.com/upper/db/v4"
)

type RememberToken struct {
	ID            int       `db:"id,omitempty"`
	UserID        int       `db:"user_id"`
	RememberToken string    `db:"remember_token"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

func (r *RememberToken) Table() string {
	return "remember_tokens"
}

func (r *RememberToken) Insert(userID int, token string) error {
	coll := upper.Collection(r.Table())
	tkn := RememberToken{
		UserID:        userID,
		RememberToken: token,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if _, err := coll.Insert(tkn); err != nil {
		return err
	}
	return nil
}

func (r *RememberToken) Delete(token string) error {
	coll := upper.Collection(r.Table())
	res := coll.Find(up.Cond{"remember_token": token})
	if err := res.Delete(); err != nil {
		return err
	}
	return nil
}
