package main

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/go.crypto/bcrypt"
)

const PasswordCryptCost = bcrypt.DefaultCost

var (
	ErrPassTooShort = errors.New("Passwords must have a length of at least 3 characters.")
	ErrMissingUser  = errors.New("The username field is required.")
	ErrMissingEmail = errors.New("The email field is required.")
	ErrUserExists   = errors.New("The specified username already exists.")
	ErrUserNotExist = errors.New("The specified user does not exist.")
)

func NewUser(username, email string, password []byte) (uint64, error, bool) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)

	if len(username) == 0 {
		return 0, ErrMissingUser, true
	}
	if !strings.Contains(email, "@") {
		return 0, ErrMissingEmail, true
	}
	if len(password) < 3 {
		return 0, ErrPassTooShort, true
	}

	pass, err := bcrypt.GenerateFromPassword(password, PasswordCryptCost)
	if err != nil {
		return 0, err, false
	}

	usernameKey := "user/" + strings.ToLower(username)

	added, err := Bucket.Add(usernameKey, 0, uint64(0))
	if err != nil {
		return 0, err, false
	}
	if !added {
		return 0, ErrUserExists, true
	}

	id, err := Bucket.Incr("incr/userID", 1, 1, 0)
	if err != nil {
		// Attempt to delete the username key we added as it is invalid.
		// If this fails, something horrible has happened - possibly
		// a crashed database server, so there's nothing we can do.
		Bucket.Delete(usernameKey)

		return 0, err, false
	}

	userIDKey := "user@" + strconv.FormatUint(id, 10)

	user := &User{
		ID:          id,
		Email:       email,
		DisplayName: username,
		LoginName:   username,
		Password:    pass,
		Registered:  time.Now().UTC(),
	}

	err = Bucket.Set(userIDKey, 0, user)
	if err != nil {
		return 0, err, false
	}
	err = Bucket.Set(usernameKey, 0, id)
	if err != nil {
		return 0, err, false
	}

	return id, nil, false
}

func (u *User) update() {
	dirty := false
	if u.Registered.IsZero() {
		u.Registered = time.Now()
		dirty = true
	}
	if u.Registered.UTC() != u.Registered {
		u.Registered = u.Registered.UTC()
		dirty = true
	}
	if dirty {
		Bucket.Set("user@"+strconv.FormatUint(u.ID, 10), 0, u)
	}
}

func UserByID(id uint64) (*User, error) {
	u := new(User)
	err := Bucket.Get("user@"+strconv.FormatUint(id, 10), u)
	if err != nil {
		return nil, err
	}
	u.update()
	return u, nil
}

func UserByLogin(login string) (*User, error) {
	var id uint64
	err := Bucket.Get("user/"+strings.ToLower(strings.TrimSpace(login)), &id)
	if err != nil {
		return nil, err
	}
	if id == 0 {
		return nil, ErrUserNotExist
	}
	return UserByID(id)
}

func (u *User) CheckPassword(password []byte) error {
	return bcrypt.CompareHashAndPassword(u.Password, password)
}
