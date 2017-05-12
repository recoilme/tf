package main

import (
	"log"
	"strings"
	"testing"

	"github.com/recoilme/tf/httputils"

	"golang.org/x/net/html"
	"golang.org/x/net/publicsuffix"
)

func TestAverage(t *testing.T) {
	groupName := "myakotkapub"
	group := pubDbGet(groupName)
	//if group == nil {
	log.Println("group1", group)
}

func TestVkWallUpd(t *testing.T) {
	log.Println("vkWallUpd")
	//vkapi.vkWallUpd()
}

func TestLinkExtract(t *testing.T) {
	var defHeaders = make(map[string]string)
	defHeaders["User-Agent"] = "script::recoilme:v1"
	defHeaders["Authorization"] = "Client-ID 4191ffe3736cfcb"

	b := httputils.HttpGet("https://www.reddit.com/.rss?feed=32f7ac01a37b80c88037018e186bb2581de14d55&user=recoilme", defHeaders)
	if b == nil {
		return
	}
	s := string(b)
	log.Println(s)
	var rss string
	//s := `<link rel="alternate" type="application/rss+xml" href="https://vc.ru/feed">`
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		log.Fatal(err)
	}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "link" {
			var isRss bool
			for _, a := range n.Attr {
				if a.Key == "type" {
					if a.Val == "application/rss+xml" || a.Val == "application/atom+xml" {
						isRss = true
						break
					}
				}
			}
			if isRss {
				for _, a := range n.Attr {
					if a.Key == "href" {
						rss = a.Val
						break
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
			if rss != "" {
				break
			}
		}
	}
	f(doc)
	log.Println("rss", rss)
}

func TestMainDomain(t *testing.T) {
	var result string
	result, _ = publicsuffix.EffectiveTLDPlusOne("www.reddit.com")
	if result != "reddit.com" {
		t.Error("Expected reddit.com, got ", result)
	}
	result, _ = publicsuffix.EffectiveTLDPlusOne("m.vk.com")
	if result != "vk.com" {
		t.Error("Expected reddit.com, got ", result)
	}
	result, _ = publicsuffix.EffectiveTLDPlusOne("adsfffsf")
	log.Println("result:", result)
	if result != "" {
		t.Error("Expected reddit.com, got ", result)
	}
	result, _ = publicsuffix.EffectiveTLDPlusOne("en.reddit.com")
	if result != "reddit.com" {
		t.Error("Expected reddit.com, got ", result)
	}
}

func TestRss(t *testing.T) {
	url := "https://www.reddit.com/.rss?feed=32f7ac01a37b80c88037018e186bb2581de14d55&user=recoilme"
	//url := "https://vc.ru/feed"
	link := getFeedLink(url)
	if link != url {
		t.Error("Expected got", link)
	}
}

func TestRssExtract(t *testing.T) {
	url := "https://vc.ru/"
	urlexpect := "https://vc.ru/feed"
	link := rssExtract(url)
	if link != urlexpect {
		t.Error("Expected got", link)
	}
}

func TestSubs(t *testing.T) {
	subs := usersub("", 1263310, true)
	for k, v := range subs {
		log.Println("k", k, "v", v)
	}
}
