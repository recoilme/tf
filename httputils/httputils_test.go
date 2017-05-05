package httputils

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"
)

type Resp struct {
	UserAgent string `json:"user-agent"`
}

var urls = []string{
	"http://www.google.com",        //good url, 200
	"http://www.googlegoogle.com/", //bad url
	"http://www.zoogle.com",        //500 example
}

func init() {
	//log.SetOutput(ioutil.Discard)
}

func main() {
	log.Println("main")
}

func TestGetUa(t *testing.T) {
	b := HttpGet("http://httpbin.org/user-agent", nil)
	var res = ""
	if b != nil {
		res = string(b)
		var result Resp
		json.Unmarshal(b, &result)
		res = result.UserAgent
	}
	if res != defHeaders["User-Agent"] {
		t.Error("User agent not match '%s'", res)
	}
}

func TestGetUaCustom(t *testing.T) {
	defHeader := make(map[string]string)
	defHeader["User-Agent"] = "bot"
	b := HttpGet("http://httpbin.org/user-agent", defHeader)
	var res = ""
	if b != nil {
		res = string(b)
		var result Resp
		json.Unmarshal(b, &result)
		res = result.UserAgent
	}
	if res != defHeader["User-Agent"] {
		t.Error("User agent not match:", res)
	}
}

func TestGetMissing(t *testing.T) {
	//TODO FAIL on mac, darwin problem?

	body := HttpGet("http://missinghostexample.com", nil)
	if body == nil {
		log.Println("not nil")
	}
	var b []byte
	log.Println("empty", string(b))
}

func MakeRequest(url string, ch chan<- string) {
	b := HttpGet(url, nil)
	if b == nil {
		ch <- fmt.Sprintf("url: %s, BAD", url)
	} else {
		ch <- fmt.Sprintf("url: %s, OK", url) // put response into a channel
	}
}

func TestConcurrentReq(t *testing.T) {
	output := make([][]string, 0) //define an array to hold responses

	//MAKE URL REQUESTS----------------------------------------------
	for _, url := range urls {
		ch := make(chan string)                 //create a channel for each request
		go MakeRequest(url, ch)                 //make concurrent http request
		output = append(output, []string{<-ch}) //append output to an array
	}

	//PRINT OUTPUT ----------------------
	for _, value := range output {
		fmt.Println(value)
	}
}
