package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

func PostByID(id uint64) (*Post, error) {
	p := new(Post)
	err := Bucket.Get("post@"+strconv.FormatUint(id, 10), p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

var postLock sync.Mutex

func NewPost(author *User, replyTo *Post, title, content string, tags []string) (uint64, error) {
	if title == "" {
		title = replyTo.Title
		if !strings.HasPrefix(title, "Re: ") {
			title = "Re: " + title
		}
	}

	// easier and cheaper than trying to do compare-and-set on topics.
	postLock.Lock()
	defer postLock.Unlock()

	topic, err := TopicByID(replyTo.Topic)
	if err != nil {
		return 0, err
	}

	id, err := Bucket.Incr("incr/postID", 1, 1, 0)
	if err != nil {
		return 0, err
	}

	now := time.Now().UTC()

	topic.LastAuthor = author.ID
	topic.LastMod = now
	topic.LastPost = now
	topic.LastTitle = title
	topic.Replies++

	post := &Post{
		ID:      id,
		Topic:   topic.ID,
		ReplyTo: replyTo.ID,
		Author:  author.ID,
		LastMod: now,
		Created: now,
		Title:   title,
		Content: content,
		Tags:    tags,
	}

	err = Bucket.Set("post@"+strconv.FormatUint(id, 10), 0, post)
	if err != nil {
		return 0, err
	}
	err = Bucket.Set("topic@"+strconv.FormatUint(topic.ID, 10), 0, topic)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func NewTopic(author *User, forum *Forum, title, content string, tags []string) (uint64, error) {
	tid, err := Bucket.Incr("incr/topicID", 1, 1, 0)
	if err != nil {
		return 0, err
	}

	pid, err := Bucket.Incr("incr/postID", 1, 1, 0)
	if err != nil {
		return 0, err
	}

	now := time.Now().UTC()

	topic := &Topic{
		ID:          tid,
		Forum:       forum.ID,
		Created:     now,
		LastPost:    now,
		LastMod:     now,
		FirstTitle:  title,
		LastTitle:   title,
		FirstAuthor: author.ID,
		LastAuthor:  author.ID,
	}

	post := &Post{
		ID:      pid,
		Topic:   tid,
		ReplyTo: 0,
		Author:  author.ID,
		LastMod: now,
		Created: now,
		Title:   title,
		Content: content,
		Tags:    tags,
	}

	err = Bucket.Set("post@"+strconv.FormatUint(pid, 10), 0, post)
	if err != nil {
		return 0, err
	}
	err = Bucket.Set("topic@"+strconv.FormatUint(tid, 10), 0, topic)
	if err != nil {
		return 0, err
	}

	return tid, nil
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
