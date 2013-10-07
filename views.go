package main

import (
	"github.com/couchbaselabs/go-couchbase"
)

const ViewRevision = 3

const (
	DDocName       = "commserv"
	ViewForums     = "forums"
	ViewForumTopic = "forum_topic"
	ViewTopicPost  = "topic_post"
)

var views = couchbase.DDocJSON{
	"javascript",
	map[string]couchbase.ViewDefinition{
		ViewForums: {
			Map: `function(doc, meta) {
	if (meta.id.match(/^forum@\d+/)) {
		emit(doc.ID, null);
	}
}`,
		},
		ViewForumTopic: {
			Map: `function(doc, meta) {
	if (meta.id.match(/^topic@\d+/)) {
		emit([doc.Forum].concat(dateToArray(doc.LastMod)), null);
	}
}`,
		},
		ViewTopicPost: {
			Map: `function(doc, meta) {
	if (meta.id.match(/^post@\d+/)) {
		emit([doc.Topic].concat(dateToArray(doc.Created)), null);
	}
}`,
		},
	},
}

func initViews() {
	var currentViewRevision uint64
	err := Bucket.Get("meta/viewRevision", &currentViewRevision)
	if err == nil && currentViewRevision == ViewRevision {
		return
	}
	err = Bucket.PutDDoc(DDocName, views)
	if err != nil {
		panic(err)
	}
	err = Bucket.Set("meta/viewRevision", 0, uint64(ViewRevision))
	if err != nil {
		panic(err)
	}
}
