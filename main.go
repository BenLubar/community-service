package main

import (
	"log"
	"net"
	"net/http"

	"github.com/couchbaselabs/go-couchbase"
)

var Bucket *couchbase.Bucket

func main() {
	ln, err := net.Listen("tcp", ":3024") //":0")
	if err != nil {
		log.Fatalf("Could not listen: %v", err)
	}
	defer ln.Close()
	log.Println("Now listening on", ln.Addr())

	Bucket, err = couchbase.GetBucket("http://127.0.0.1:8091", "default", "commserv")
	if err != nil {
		log.Fatalf("Could not connect to db: %v", err)
	}

	initAuth()

	log.Fatalf("http.Serve failed: %v", http.Serve(ln, nil))
}
