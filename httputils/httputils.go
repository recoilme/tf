package httputils

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

var defHeaders = make(map[string]string)

type Config struct {
	ConnectTimeout   time.Duration
	ReadWriteTimeout time.Duration
}

type OtvetRes struct {
	Count     int `json:"count"`
	CountShow int `json:"count_show"`
	Number    int `json:"number"`
	Results   []struct {
		Count   int `json:"count"`
		TimeAgo int `json:"time_ago"`
		Author  struct {
			URL   string `json:"url"`
			Nick  string `json:"nick"`
			ID    string `json:"id"`
			Filin string `json:"filin"`
		} `json:"author"`
		URL      string `json:"url"`
		Time     int    `json:"time"`
		Question string `json:"question"`
		State    int    `json:"state"`
		Number   int    `json:"number"`
		Catname  string `json:"catname"`
		Banswer  string `json:"banswer"`
		Answer   string `json:"answer"`
		IsPoll   bool   `json:"is_poll"`
		ID       string `json:"id"`
	} `json:"results"`
	Start int `json:"start"`
}

func TimeoutDialer(config *Config) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, config.ConnectTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(config.ReadWriteTimeout))
		return conn, nil
	}
}

func NewTimeoutClient(args ...interface{}) *http.Client {
	// Default configuration
	//http.DefaultTransport.(*http.Transport).TLSNextProto = make(map[string]func(string, *tls.Conn) http.RoundTripper)
	config := &Config{
		ConnectTimeout:   7 * time.Second,
		ReadWriteTimeout: 7 * time.Second,
	}

	// merge the default with user input if there is one
	if len(args) == 1 {
		timeout := args[0].(time.Duration)
		config.ConnectTimeout = timeout
		config.ReadWriteTimeout = timeout
	}

	if len(args) == 2 {
		config.ConnectTimeout = args[0].(time.Duration)
		config.ReadWriteTimeout = args[1].(time.Duration)
	}
	http.DefaultTransport.(*http.Transport).TLSNextProto = make(map[string]func(string, *tls.Conn) http.RoundTripper)

	return &http.Client{
		Transport: &http.Transport{
			Dial: TimeoutDialer(config),
		},
	}
}

func init() {
	defHeaders["User-Agent"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.12; rv:52.0) Gecko/20100101 Firefox/52.0"
	defHeaders["Accept-Language"] = "ru-RU,ru;q=0.8,en-US;q=0.5,en;q=0.3"
	defHeaders["Referer"] = "https://ya.ru/"
	defHeaders["Cookie"] = ""
}

// HttpGet create request with default headers + custom headers
func HttpGet(url string, headers map[string]string) []byte {
	//log.Println("httpGet", url)

	client := NewTimeoutClient()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
		return nil
	}
	for k, v := range defHeaders {
		req.Header.Set(k, v)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	} else {
		return body
	}

	return nil
}

// HttpPut create request with default headers + custom headers
func HttpPut(url string, headers map[string]string, b []byte) (result bool) {
	log.Println("httpPut", url)
	client := NewTimeoutClient()
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(b))
	if err != nil {
		log.Println(err)
		return
	}
	for k, v := range defHeaders {
		req.Header.Set(k, v)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	} else {
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			result = true
		}
	}
	return
}

// HttpGet create request with default headers + custom headers
func HttpHead(url string, headers map[string]string) int64 {
	//log.Println("httpGet", url)

	client := NewTimeoutClient()
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		log.Println(err)
		return 0
	}
	for k, v := range defHeaders {
		req.Header.Set(k, v)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return 0
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return 0
	}
	if strings.HasPrefix(resp.Header.Get("Content-Type"), "image") {
		return resp.ContentLength
	} else {
		return 0
	}
}
