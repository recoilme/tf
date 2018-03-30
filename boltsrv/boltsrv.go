package main

import (
	"fmt"
	"net/http"
	"os"
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
		fmt.Printf("%s\n", err)
	}
	defer boltdb.Close()
	http.HandleFunc("/", handler)
	http.ListenAndServe(addr, nil)
}

func main() {
	argsWithProg := os.Args
	if len(argsWithProg) > 1 {
		Serve(os.Args[1])
	} else {
		Serve(":5000")
	}
}
