package main

import "time"

type Topic struct {
	ID    uint64 `json:",string"` // unique identity
	Forum uint64 `json:",string"` // -> Forum.ID

	// The fields that would normally be here in other forums are instead
	// in a Post with p.Topic = t.ID and p.ReplyTo = 0.
}

type Post struct {
	ID      uint64 `json:",string"` // unique identity
	Topic   uint64 `json:",string"` // -> Topic.ID
	Author  uint64 `json:",string"` // -> User.ID
	ReplyTo uint64 `json:",string"` // -> Post.ID
	Created time.Time
	LastMod time.Time

	Title   string
	Content string
	Tags    []string
}

type User struct {
	ID        uint64 `json:",string"` // unique identity
	LoginName string // unique (case insensitive)
	Email     string
	Password  []byte

	DisplayName string
	Tagline     string
}
