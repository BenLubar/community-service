package main

import (
	"compress/gzip"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

var tmpl = template.Must(template.New("").Funcs(template.FuncMap{
	"User": UserByID,
	"Comma": func(n interface{}) string {
		if x, ok := n.(uint64); ok {
			return humanize.Comma(int64(x))
		}
		return humanize.Comma(n.(int64))
	},
	"RelTime": humanize.Time,
}).ParseGlob("tmpl/*.html"))

type TmplMeta struct {
	SiteTitle   string
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

type TmplRegister struct {
	Meta  *TmplMeta
	User  string
	Email string
	Error string
}

type TmplForum struct {
	Meta *TmplMeta
	// TODO: Forum *Forum
	Topics []*Topic
}

func GetTmplMeta(r *http.Request) *TmplMeta {
	m := &TmplMeta{}
	if err := Bucket.Get("meta/siteTitle", &m.SiteTitle); err != nil {
		m.SiteTitle = "Forum"
	}
	m.URL = r.URL.String()
	m.LoggedIn, _ = UserByCookie(r)
	return m
}

func ShowTemplate(w http.ResponseWriter, r *http.Request, file string, data interface{}) {
	w, gzipClose := maybeGzip(w, r)
	defer gzipClose()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err := tmpl.ExecuteTemplate(w, file, data)
	if err != nil {
		log.Println(r.URL, err)
	}
}

func init() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		ShowTemplate(w, r, "index.html", &TmplIndex{
			Meta: GetTmplMeta(r),
		})
	})

	http.Handle("/favicon.ico", http.RedirectHandler("/static/favicon.ico", http.StatusMovedPermanently))

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
