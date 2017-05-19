package httputils

import (
	"bytes"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

var defHeaders = make(map[string]string)

type Config struct {
	ConnectTimeout   time.Duration
	ReadWriteTimeout time.Duration
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
