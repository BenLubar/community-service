package main

import (
	"log"
	"net/http"
	"strconv"
)

func PostByID(id uint64) (*Post, error) {
	p := new(Post)
	err := Bucket.Get("post@"+strconv.FormatUint(id, 10), p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func init() {
	http.HandleFunc("/p/", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/p/"):]
		id_, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		post, err := PostByID(id_)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		topicID := strconv.FormatUint(post.Topic, 10)

		var v struct {
			Rows []struct {
				ID string `json:"id"`
			}
		}
		err = Bucket.ViewCustom(DDocName, ViewTopicPost, map[string]interface{}{
			"startkey": []interface{}{topicID},
			"endkey":   []interface{}{topicID, map[string]interface{}{}},
		}, &v)
		if err != nil {
			log.Println(r.URL, err)
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		postID := "post@" + id

		for i, p := range v.Rows {
			if p.ID == postID {
				start := (i / 50) * 50 // TODO: make configurable
				if start == 0 {
					http.Redirect(w, r, "/t/"+topicID+"#"+id, http.StatusMovedPermanently)
				} else {
					http.Redirect(w, r, "/t/"+topicID+"?start="+strconv.Itoa(start)+"#"+id, http.StatusMovedPermanently)
				}
				return
			}
		}

		http.Error(w, "unknown error", http.StatusInternalServerError)
	})
}
