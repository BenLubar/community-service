package main

import (
	"crypto/rand"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var authSalt []byte

func initAuth() {
	var err error
	authSalt, err = Bucket.GetRaw("meta/authSalt")
	if err != nil {
		authSalt = make([]byte, 20)
		_, err = io.ReadFull(rand.Reader, authSalt)
		if err != nil {
			panic(err)
		}
		err = Bucket.SetRaw("meta/authSalt", 0, authSalt)
		if err != nil {
			panic(err)
		}
	}
}

func (u *User) SetAuthCookie(w http.ResponseWriter, valid time.Duration) {
	validUntil := time.Now().Add(valid)

	value := fmt.Sprintf("%d:%d:%s:%s", u.ID, validUntil.Unix(), u.Password, authSalt)
	value = fmt.Sprintf("%d:%d:%x", u.ID, validUntil.Unix(), sha1.Sum([]byte(value)))

	http.SetCookie(w, &http.Cookie{
		Name:     "commserv_auth",
		Value:    value,
		Expires:  validUntil,
		HttpOnly: true,
	})
}

func (u *User) CheckAuthCookie(r *http.Request) bool {
	cookie, err := r.Cookie("commserv_auth")
	if err != nil {
		return false
	}
	parts := strings.Split(cookie.Value, ":")
	if len(parts) != 3 {
		return false
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return false
	}
	if id != u.ID {
		return false
	}

	validUntil, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return false
	}
	if validUntil < time.Now().Unix() {
		return false
	}

	value := fmt.Sprintf("%d:%d:%s:%s", u.ID, validUntil, u.Password, authSalt)
	return parts[2] == fmt.Sprintf("%x", sha1.Sum([]byte(value)))
}

var ErrInvalidCookie = errors.New("Invalid cookie.")

func UserByCookie(r *http.Request) (*User, error) {
	cookie, err := r.Cookie("commserv_auth")
	if err != nil {
		return nil, err
	}
	parts := strings.Split(cookie.Value, ":")
	if len(parts) != 3 {
		return nil, ErrInvalidCookie
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return nil, err
	}

	u, err := UserByID(id)
	if err != nil {
		return nil, err
	}

	if !u.CheckAuthCookie(r) {
		return nil, ErrInvalidCookie
	}
	return u, nil
}

func init() {
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/logout" {
			http.NotFound(w, r)
			return
		}

		if r.Method == "POST" {
			http.SetCookie(w, &http.Cookie{
				Name:   "commserv_auth",
				MaxAge: -1,
			})
		}
		http.Redirect(w, r, "/", http.StatusFound)
	})
}
