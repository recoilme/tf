package main

import (
	"log"
	"testing"

	"strings"

	"net/http"

	"strconv"

	"golang.org/x/net/html"
)

func TestRssdomains(t *testing.T) {
	domains := rssdomains()
	log.Println("group1", domains)
}

func getVal(t html.Token, key string) (ok bool, val string) {
	// Iterate over all of the Token's attributes until we find an "href"
	for _, a := range t.Attr {
		if a.Key == key {
			val = a.Val
			ok = true
		}
	}

	// "bare" return will return the variables (ok, href) as defined in
	// the function definition
	return
}

func extractLinks(s string) (links []string) {
	z := html.NewTokenizer(strings.NewReader(s))
	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return
		case tt == html.StartTagToken:
			t := z.Token()
			switch t.Data {
			case "img":
				ok, href := getVal(t, "src")
				if !ok {
					continue
				}
				links = append(links, href)
				//log.Println(href)
			case "a":
				ok, href := getVal(t, "href")
				if !ok {
					continue
				}
				links = append(links, href)
			default:
				continue
			}
		}
	}
}

func getImgs(links []string) (imgs map[string]int) {
	imgs = make(map[string]int)
	for _, link := range links {
		resp, err := http.Head(link)
		if err != nil {
			continue
		}
		len, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
		// 10 - 500kb~
		if len < 10000 || len > 500000 {
			continue
		}
		isImg := strings.HasPrefix(resp.Header.Get("Content-Type"), "image")
		if isImg {
			imgs[link] = len
		}
	}
	return imgs
}

func TestRss2(t *testing.T) {
	s := `oft <a href="http://1.ru">12</a>Build 2017. <img src="https://png.cmtt.space/paper-preview-fox/m/ic/microsoft-build-announcements/8d1c780b2eba-original.jpg">`
	links := extractLinks(s)
	imgs := getImgs(links)
	var max = 0
	var maximg = ""
	for img, len := range imgs {
		if len > max {
			max = len
			maximg = img
		}
		log.Println(img, "len", len)
	}
	log.Println(maximg)
}
