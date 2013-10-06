package main

import (
	"crypto/rand"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/xsrftoken"
)

const authCookieName = "commserv_auth"

var authSalt []byte
var xsrfSalt []byte

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

	xsrfSalt, err = Bucket.GetRaw("meta/xsrfSalt")
	if err != nil {
		xsrfSalt = make([]byte, 20)
		_, err = io.ReadFull(rand.Reader, xsrfSalt)
		if err != nil {
			panic(err)
		}
		err = Bucket.SetRaw("meta/xsrfSalt", 0, xsrfSalt)
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
		Name:     authCookieName,
		Value:    value,
		Expires:  validUntil,
		HttpOnly: true,
	})
}

func (u *User) CheckAuthCookie(r *http.Request) bool {
	cookie, err := r.Cookie(authCookieName)
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

func (u *User) GenerateNonce(actionID string) string {
	return xsrftoken.Generate(string(xsrfSalt), u.LoginName, actionID)
}

func (u *User) CheckNonce(nonce, actionID string) bool {
	return xsrftoken.Valid(nonce, string(xsrfSalt), u.LoginName, actionID)
}

var ErrInvalidCookie = errors.New("Invalid cookie.")

func UserByCookie(r *http.Request) (*User, error) {
	cookie, err := r.Cookie(authCookieName)
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
		defer http.Redirect(w, r, "/", http.StatusFound)

		if r.Method != "POST" {
			return
		}

		u, err := UserByCookie(r)
		if err != nil {
			return
		}
		if !u.CheckNonce(r.PostFormValue("logout_nonce"), "logout") {
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:   authCookieName,
			MaxAge: -1,
		})
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/login" {
			http.NotFound(w, r)
			return
		}

		signInError := ""
		if r.Method == "POST" {
			signInError = processLogin(w, r)
			if signInError == "" {
				return
			}
		}

		w, gzipClose := maybeGzip(w, r)
		defer gzipClose()

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		meta := GetTmplMeta(r)
		meta.IsLoginPage = true
		err := tmpl.ExecuteTemplate(w, "login.html", &TmplLogin{
			Meta:  meta,
			User:  r.PostFormValue("user"),
			Error: signInError,
		})
		if err != nil {
			log.Println(r.URL, err)
		}
	})
}

func processLogin(w http.ResponseWriter, r *http.Request) string {
	if _, err := UserByCookie(r); err == nil {
		return "You are already logged in."
	}
	if r.PostFormValue("user") == "" || r.PostFormValue("pass") == "" {
		return "All fields are required."
	}

	u, err := UserByLogin(r.PostFormValue("user"))
	if err != nil || u.CheckPassword([]byte(r.PostFormValue("pass"))) != nil {
		return "Incorrect login information."
	}

	ref := r.FormValue("ref")
	if len(ref) < 2 || ref[0] != '/' || ref[1] == '/' {
		ref = "/"
	}
	u.SetAuthCookie(w, time.Hour*24*30)
	http.Redirect(w, r, ref, http.StatusFound)
	return ""
}
