package main

import "time"

type Topic struct {
	ID    uint64 // unique identity
	Forum uint64 // -> Forum.ID

	// The fields that would normally be here in other forums are instead
	// in a Post with p.Topic = t.ID and p.ReplyTo = 0.
}

type Post struct {
	ID      uint64 // unique identity
	Topic   uint64 // -> Topic.ID
	Author  uint64 // -> User.ID
	ReplyTo uint64 // -> Post.ID
	Created time.Time
	LastMod time.Time

	Title   string
	Content string
	Tags    []string
}

type User struct {
	ID        uint64 // unique identity
	LoginName string // unique
	Email     string // unique
	Password  []byte

	DisplayName string
	Tagline     string
}
