// Package boltapi contains handler for working with boltdb.
// For server example see boltapi_test
/*

PUT:

# params
host/database/backet/key
and value in body

curl -X PUT -H "Content-Type: application/octet-stream" --data-binary "@durov.jpg" localhost:5000/bolt/images/durov
curl -X PUT -H "Content-Type: text/html" -d '{"username":"xyz","password":"xyz"}' localhost:5000/bolt/users/user1
curl -X PUT -H "Content-Type: text/html" -d 'some value' localhost:5000/bolt/users/user2

GET:

# params
host/database/backet/key

curl localhost:5000/bolt/images/durov
return: bytes
curl localhost:5000/bolt/users/user1
return: {"username":"xyz","password":"xyz"}
curl -v localhost:5000/bolt/images/durov2
return 404 Error

POST:

# params
host/database/backet/key?cnt=1000&order=desc&vals=false

key: first key, possible values "some_your_key" or "some_your_key*" for prefix scan, Last, First - default Last
cnt: return count records, default 1000
order: sorting order (keys ordered as strings!), default desc
vals: return values, default false

curl -X POST localhost:5000/bolt/users
return: {"user2","user1"}

curl -X POST localhost:5000/bolt/users/First?cnt=2&order=asc
return: {"user1"}

curl -X POST "http://localhost:5000/bolt/users/use*?order=asc&vals=true"
return: {"user1":"{"username":"xyz","password":"xyz"}","user2":"some value"}

curl -X POST "http://localhost:5000/bolt/users/user2?order=desc&vals=true"
return: {"user2":"some value","user1":"{"username":"xyz","password":"xyz"}"}

DELETE:

curl -X DELETE http://localhost:5000/bolt/users/user2
return 200 Ok (or 404 Error if bucket! not found)
*/
package boltapi

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/boltdb/bolt"
)

// BoltAPI contains handler for rest api to boltdb
func BoltAPI(db *bolt.DB, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	method := r.Method
	urlPart := strings.Split(r.URL.Path, "/")
	var bucketstr = ""
	var keystr = ""
	//log.Println("len", len(urlPart))
	if len(urlPart) == 4 {
		bucketstr = urlPart[2]
		keystr = urlPart[3]
	}
	if len(urlPart) == 3 {
		bucketstr = urlPart[2]
	}
	//log.Println("bucketstr", bucketstr)
	//log.Println("keystr", keystr)
	switch method {
	case "GET":
		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(bucketstr))
			if b == nil {
				w.WriteHeader(http.StatusNotFound)
				return nil
			}
			val := b.Get([]byte(keystr))
			if len(val) == 0 {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.Write(val)
			}
			return nil
		})

	case "PUT":
		err := db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte(bucketstr))
			if err != nil {
				return err
			}
			val, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return err
			}
			err = b.Put([]byte(keystr), val)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			log.Println(err)
		}

	case "POST":
		//log.Println("POST")
		cnt := r.URL.Query().Get("cnt")
		var order = r.URL.Query().Get("order")
		var vals = r.URL.Query().Get("vals")
		var max = 1000
		var prefix []byte
		m, e := strconv.Atoi(cnt)
		if e == nil {
			max = m
		}
		var buffer bytes.Buffer
		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(bucketstr))
			if b == nil {
				w.WriteHeader(http.StatusNotFound)
				return nil
			}
			c := b.Cursor()
			if keystr == "Last" || keystr == "" {
				k, _ := c.Last()
				keystr = string(k)
			}
			if keystr == "First" {
				k, _ := c.First()
				keystr = string(k)
			}
			if order == "" {
				order = "desc"
			}
			if vals == "" {
				vals = "false"
			}
			if strings.HasSuffix(keystr, "*") {
				prefix = []byte(keystr[:len(keystr)-1])
				keystr = keystr[:len(keystr)-1]
			}
			var comp = func(i int, m int, k []byte) bool {
				if prefix != nil {
					return i < m && bytes.HasPrefix(k, prefix)
				}
				return i < m
			}
			i := 0
			buffer.WriteString("[")
			switch order {
			case "asc":
				for k, v := c.Seek([]byte(keystr)); k != nil && comp(i, max, k); k, v = c.Next() {
					if i != 0 {
						buffer.WriteString(",")
					}
					if vals == "false" {
						buffer.WriteString(fmt.Sprintf("\"%s\"", k))
					} else {
						buffer.WriteString(fmt.Sprintf("{\"%s\":\"%s\"}", k, v))
					}
					i++
				}
			default:
				for k, v := c.Seek([]byte(keystr)); k != nil && comp(i, max, k); k, v = c.Prev() {
					if i != 0 {
						buffer.WriteString(",")
					}
					if vals == "false" {
						buffer.WriteString(fmt.Sprintf("\"%s\"", k))
					} else {
						buffer.WriteString(fmt.Sprintf("{\"%s\":\"%s\"}", k, v))
					}
					i++
				}

			}
			buffer.WriteString("]")
			w.Write(buffer.Bytes())
			return nil
		})
	case "DELETE":
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(bucketstr))
			if b == nil {
				w.WriteHeader(http.StatusNotFound)
				return nil
			}
			b.Delete([]byte(keystr))
			w.WriteHeader(http.StatusOK)
			return nil
		})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
