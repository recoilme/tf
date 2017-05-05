package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/recoilme/tf/boltapi"
)

var boltdb *bolt.DB
var baseuri = "bolt/"

// main handler
// default path localhost:5000/bolt/
func handler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	switch {
	case strings.HasPrefix(path, baseuri):
		boltapi.BoltAPI(boltdb, w, r)
	default:
		w.WriteHeader(http.StatusOK)
	}
}

// Serve run server
// example addr: ":5000"
func Serve(addr string) {
	var err error
	boltdb, err = bolt.Open("bolt.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer boltdb.Close()
	http.HandleFunc("/", handler)
	http.ListenAndServe(addr, nil)
}

func main() {
	Serve(":5000")
}
