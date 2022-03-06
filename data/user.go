package data

import (
	"errors"
	"time"

	up "github.com/upper/db/v4"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int       `db:"id,omitempty"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Email     string    `db:"email"`
	Active    int       `db:"user_active"`
	Password  string    `db:"password"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Token     Token     `db:"-"`
}

func (u *User) Table() string {
	return "users"
}

func (u *User) GetAll() ([]*User, error) {
	var users []*User

	coll := upper.Collection(u.Table())
	res := coll.Find().OrderBy("last_name")
	if err := res.All(&users); err != nil {
		return nil, err
	}

	return users, nil
}

func (u *User) GetByEmail(email string) (*User, error) {
	var usr *User

	coll := upper.Collection(u.Table())
	res := coll.Find(up.Cond{"email": email})
	if err := res.One(&usr); err != nil {
		return nil, err
	}

	t, err := getUserToken(*usr)
	if err != nil {
		return nil, err
	}
	usr.Token = t

	return usr, nil
}

func (u *User) GetById(id int) (*User, error) {
	var usr *User

	coll := upper.Collection(u.Table())
	res := coll.Find(id)
	if err := res.One(&usr); err != nil {
		return nil, err
	}

	t, err := getUserToken(*usr)
	if err != nil {
		return nil, err
	}
	usr.Token = t

	return usr, nil
}

func (u *User) Update(user *User) error {
	user.UpdatedAt = time.Now()
	coll := upper.Collection(u.Table())
	res := coll.Find(user.ID)
	if err := res.Update(user); err != nil {
		return err
	}
	return nil
}

func (u *User) Delete(id int) error {
	coll := upper.Collection(u.Table())
	res := coll.Find(id)
	if err := res.Delete(); err != nil {
		return err
	}
	return nil
}

func (u *User) Insert(user *User) (int, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	if err != nil {
		return 0, err
	}

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Password = string(hash)

	coll := upper.Collection(u.Table())
	res, err := coll.Insert(user)
	if err != nil {
		return 0, err
	}

	id := getInsertedID(res.ID())
	return id, nil
}

func (u *User) ResetPassword(id int, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	usr, err := u.GetById(id)
	if err != nil {
		return err
	}

	usr.Password = string(hash)
	if err := u.Update(usr); err != nil {
		return err
	}

	return nil
}

func (u *User) CheckPassword(password string) (bool, error) {
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func getUserToken(user User) (Token, error) {
	var token Token

	coll := upper.Collection(token.Table())
	res := coll.Find(up.Cond{
		"user_id":  user.ID,
		"expiry >": time.Now(),
	}).OrderBy("created_at desc")

	if err := res.One(&token); err != nil && err != up.ErrNilRecord && err != up.ErrNoMoreRows {
		return token, err
	}

	return token, nil
}
