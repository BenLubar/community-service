package main

import (
	"log"
	"net/http"
	"strconv"
)

func TopicByID(id uint64) (*Topic, error) {
	t := new(Topic)
	err := Bucket.Get("topic@"+strconv.FormatUint(id, 10), t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func init() {
	http.HandleFunc("/t/", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/t/"):]
		id_, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		start := 0 // TODO: make configurable
		if s := r.FormValue("start"); s != "" {

		}

		topic, err := TopicByID(id_)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		forum, err := ForumByID(topic.Forum)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		var v struct {
			Rows []struct {
				Doc struct {
					Json *Post
				}
			}
		}
		err = Bucket.ViewCustom(DDocName, ViewTopicPost, map[string]interface{}{
			"include_docs": true,
			"startkey":     []interface{}{id},
			"endkey":       []interface{}{id, map[string]interface{}{}},
			"limit":        50, // TODO: make configurable
			"skip":         start,
		}, &v)
		if err != nil {
			log.Println(r.URL, err)
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}

		posts := make([]*Post, len(v.Rows))
		for i, r := range v.Rows {
			posts[i] = r.Doc.Json
		}

		status := http.StatusOK
		if len(posts) == 0 {
			status = http.StatusNotFound
		}

		ShowTemplate(w, r, "topic.html", &TmplTopic{
			Meta:  GetTmplMeta(r),
			Forum: forum,
			Topic: topic,
			Posts: posts,
		}, status)
	})
}
