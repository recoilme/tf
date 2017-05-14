package main

import (
	"log"
	"testing"
)

func TestRssdomains(t *testing.T) {
	domains := rssdomains()
	log.Println("group1", domains)
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
