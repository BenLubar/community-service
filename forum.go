package main

import (
	"log"
	"net/http"
	"strconv"
)

func init() {
	http.HandleFunc("/f/", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/f/"):]
		_, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// TODO: get forum data

		var v struct {
			Rows []struct {
				Doc struct {
					Json *Topic
				}
			}
		}
		err = Bucket.ViewCustom(DDocName, ViewForumTopic, map[string]interface{}{
			"include_docs": true,
			"descending":   true,
			"startkey":     []interface{}{id, map[string]interface{}{}},
			"endkey":       []interface{}{id},
		}, &v)
		if err != nil {
			log.Println(r.URL, err)
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}

		topics := make([]*Topic, len(v.Rows))
		for i, r := range v.Rows {
			topics[i] = r.Doc.Json
		}

		ShowTemplate(w, r, "forum.html", &TmplForum{
			Meta:   GetTmplMeta(r),
			Topics: topics,
		})
	})
}
