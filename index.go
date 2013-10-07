package main

import (
	"log"
	"net/http"
)

func init() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		var v struct {
			Rows []struct {
				Doc struct {
					Json *Forum
				}
			}
		}
		err := Bucket.ViewCustom(DDocName, ViewForums, map[string]interface{}{
			"include_docs": true,
		}, &v)
		if err != nil {
			log.Println(r.URL, err)
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}

		forums := make([]*Forum, len(v.Rows))
		for i, r := range v.Rows {
			forums[i] = r.Doc.Json
		}

		ShowTemplate(w, r, "index.html", &TmplIndex{
			Meta:   GetTmplMeta(r),
			Forums: forums,
		}, http.StatusOK)
	})
}
