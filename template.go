package main

import (
	"compress/gzip"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

var tmpl = template.Must(template.ParseGlob("tmpl/*.html"))

type TmplMeta struct {
	URL         string
	LoggedIn    *User
	IsLoginPage bool
}

type TmplIndex struct {
	Meta *TmplMeta
}

type TmplLogin struct {
	Meta  *TmplMeta
	User  string
	Error string
}

func GetTmplMeta(r *http.Request) *TmplMeta {
	m := &TmplMeta{}
	m.URL = r.URL.String()
	m.LoggedIn, _ = UserByCookie(r)
	return m
}

func init() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w, gzipClose := maybeGzip(w, r)
		defer gzipClose()

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		err := tmpl.ExecuteTemplate(w, "index.html", &TmplIndex{
			Meta: GetTmplMeta(r),
		})
		if err != nil {
			log.Println(r.URL, err)
		}
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/login" {
			http.NotFound(w, r)
			return
		}

		signInError := ""
		if r.Method == "POST" {
			if _, err := UserByCookie(r); err == nil {
				signInError = "You are already logged in."
			} else if r.PostFormValue("user") == "" || r.PostFormValue("pass") == "" {
				signInError = "All fields are required."
			} else {
				u, err := UserByLogin(r.PostFormValue("user"))
				if err == nil && u.CheckPassword([]byte(r.PostFormValue("pass"))) == nil {
					ref := r.FormValue("ref")
					if len(ref) < 2 || ref[0] != '/' || ref[1] == '/' {
						ref = "/"
					}
					u.SetAuthCookie(w, time.Hour*24*30)
					http.Redirect(w, r, ref, http.StatusFound)
					return
				}

				signInError = "Incorrect login information."
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

	fs := http.FileServer(http.Dir("tmpl/"))
	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		// forbid directory indexes
		if r.URL.Path[len(r.URL.Path)-1] == '/' {
			http.Error(w, "", http.StatusForbidden)
			return
		}

		// add expires a year in the future
		w.Header().Add("Expires", time.Now().AddDate(1, 0, 0).Format(http.TimeFormat))

		// gzip, perhaps?
		w, gzipClose := maybeGzip(w, r)
		defer gzipClose()

		fs.ServeHTTP(w, r)
	})
}

type gzipWriter struct {
	http.ResponseWriter
	w *gzip.Writer
}

func (g *gzipWriter) Write(b []byte) (int, error) {
	return g.w.Write(b)
}

func maybeGzip(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, func() error) {
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && w.Header().Get("Content-Encoding") == "" {
		g, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			panic(err)
		}
		w.Header().Set("Content-Encoding", "gzip")
		return &gzipWriter{w, g}, g.Close
	}
	return w, func() error { return nil }
}
