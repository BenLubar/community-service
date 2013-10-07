package main

import "time"

type Forum struct {
	ID   uint64 `json:",string"` // unique identity
	Name string
}

type Topic struct {
	ID    uint64 `json:",string"` // unique identity
	Forum uint64 `json:",string"` // -> Forum.ID

	// Cache (these values are computable from other values, but it's easier to store them)
	FirstAuthor uint64    `json:",string"` // Post.Author : Post.Topic = Topic.ID ∧ Post.ReplyTo = 0
	LastAuthor  uint64    `json:",string"` // Post.Author : Post.Topic = Topic.ID ∧ (∀ Post2 : Post2.Topic = Topic.ID → Post2.ID ≤ Post.ID)
	FirstTitle  string    // Post.Title : Post.Topic = Topic.ID ∧ Post.ReplyTo = 0
	LastTitle   string    // Post.Title : Post.Topic = Topic.ID ∧ (∀ Post2 : Post2.Topic = Topic.ID → Post2.ID ≤ Post.ID)
	Created     time.Time // Post.Created : Post.Topic = Topic.ID ∧ Post.ReplyTo = 0
	LastMod     time.Time // max(Post.LastMod) : Post.Topic = Topic.ID
	LastPost    time.Time // max(Post.Created) : Post.Topic = Topic.ID
	Replies     uint64    // count(Post) - 1 : Post.Topic = Topic.ID
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

	Registered time.Time
	LastVisit  time.Time
	LastSeenIP string

	DisplayName string
	Tagline     string
}
