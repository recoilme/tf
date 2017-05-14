package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/net/html"

	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/mmcdole/gofeed"
	"github.com/recoilme/tf/httputils"
	"github.com/recoilme/tf/params"
)

const (
	MaxInt = int(^uint(0) >> 1)
	MinInt = -MaxInt - 1
)

var (
	bot, wrbot *tgbotapi.BotAPI
)

func main() {
	//botfile:= "telefeed.bot"
	log.Println("rsssrv")
	var err error
	tlgrmtoken, err := ioutil.ReadFile(params.Telefeedfile)
	if err != nil {
		log.Fatal(err)
	}
	writetoken, err := ioutil.ReadFile(params.Vkwriterfile)
	if err != nil {
		log.Fatal(err)
	}
	tgtoken := strings.Replace(string(tlgrmtoken), "\n", "", -1)
	wrtoken := strings.Replace(string(writetoken), "\n", "", -1)
	bot, err = tgbotapi.NewBotAPI(tgtoken)
	if err != nil {
		log.Fatal(err)
	}
	wrbot, err = tgbotapi.NewBotAPI(wrtoken)
	wrbot.Debug = false
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v+\n", time.Now())
	//go forever()
	parseRss()
	//select {} // block forever
}

func forever() {
	for {
		//fmt.Printf("%v+\n", time.Now())
		parseRss()
		time.Sleep(600 * time.Second)
	}
}

func parseRss() {
	domains := rssdomains()

	for hash, url := range domains {

		log.Println(hash)
		users := domUsers(hash)
		saveposts(url, users)
		time.Sleep(1 * time.Second)

	}
}

func domUsers(hash string) (users map[int]bool) {
	mask := params.FeedSubs + "%s"
	url := fmt.Sprintf(mask, hash)
	log.Println(url)
	b := httputils.HttpGet(url, nil)
	if b != nil {
		json.Unmarshal(b, &users)
	}
	return users
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func saveposts(link string, users map[int]bool) {
	log.Println(link)

	var defHeaders = make(map[string]string)
	defHeaders["User-Agent"] = "script::recoilme:v1"
	defHeaders["Authorization"] = "Client-ID 4191ffe3736cfcb"
	b := httputils.HttpGet(link, defHeaders)
	if b == nil {
		return
	}
	fp := gofeed.NewParser()
	feed, err := fp.Parse(bytes.NewReader(b))
	if err != nil {
		return
	}

	var last = len(feed.Items) - 1
	if last > 10 {
		last = 10
	}
	for i := range feed.Items {
		if i == last {
			break
		}
		item := feed.Items[last-i]

		log.Println(item.Link)
		key := GetMD5Hash(item.Link)
		body := httputils.HttpGet(params.Links+key, nil)
		if body != nil {
			continue
		}
		// pub feed
		b, err := json.Marshal(item)
		if err != nil {
			continue
		}
		httputils.HttpPut(params.Links+key, nil, b)
		pubpost(link, item, users)
		break
	}
}

func pubpost(domain string, p *gofeed.Item, users map[int]bool) {
	fmt.Printf("%+v\n", p)
	var vkcnt int64 = -1001067277325 //myakotka
	log.Println("pubpost", p.GUID)

	var content = p.Title + p.Description + p.Content
	var links = extractLinks(content)
	log.Println("lin1", links)
	if p.Enclosures != nil {
		for _, encl := range p.Enclosures {
			links = append(links, encl.URL)
		}
	}
	if p.Image != nil {
		links = append(links, p.Image.URL)
	}
	links = append(links, p.Link)
	imgs := getImgs(links)
	var max = 0
	var photo = ""
	for img, len := range imgs {
		if len > max {
			max = len
			photo = img
			log.Println("photo", photo, "len", len)
		}
	}
	log.Println("phot:", photo)
	if photo != "" && len(p.Title) > 3 && len(p.Title) < 200 {
		b := httputils.HttpGet(photo, nil)
		if b != nil {
			var caption = p.Title
			appendix := fmt.Sprintf("\n🔗 %s", p.Link)
			var maxlen = 248 - len(caption) - len(appendix)
			words := strings.Split(p.Description, " ")
			for i, word := range words {
				if i == 0 {
					caption = caption + "\n"
				}
				if len(word) < maxlen {
					maxlen = maxlen - len(word) - 1
					caption = caption + word + " "
				} else {
					caption = caption + ".."
					break
				}
			}
			caption = caption + appendix

			bb := tgbotapi.FileBytes{Name: photo, Bytes: b}
			msg := tgbotapi.NewPhotoUpload(vkcnt, bb)
			msg.Caption = caption
			msg.DisableNotification = true
			res, err := wrbot.Send(msg)
			if err == nil {
				for user := range users {

					bot.Send(tgbotapi.NewForward(int64(user), vkcnt, res.MessageID))
				}
			}
		}
	} else {
		msg := tgbotapi.NewMessage(vkcnt, p.Link)
		msg.DisableWebPagePreview = false
		msg.DisableNotification = true
		res, err := wrbot.Send(msg)
		if err == nil {
			for user := range users {
				log.Println(user)
				bot.Send(tgbotapi.NewForward(int64(user), vkcnt, res.MessageID))
			}
		}
	}
}

func rssdomains() map[string]string {
	domains := make(map[string]string)
	url := params.BaseUri + "feeds/Last?cnt=1000000&order=desc&vals=false"
	log.Println("rssdomains", url)
	resp, err := http.Post(url, "application/json", nil)
	if err == nil {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			var keys = make([]string, 500)
			json.Unmarshal(body, &keys)

			log.Println("keys", keys)
			for _, key := range keys {
				log.Println("key", key)
				b := httputils.HttpGet(params.Feeds+key, nil)
				if b != nil {
					domains[key] = string(b)
				}
			}
		}
	}
	return domains
}

func getVal(t html.Token, key string) (ok bool, val string) {
	// Iterate over all of the Token's attributes until we find an "href"
	for _, a := range t.Attr {
		if a.Key == key {
			val = a.Val
			ok = true
		}
	}
	return
}

func extractLinks(s string) (links []string) {
	z := html.NewTokenizer(strings.NewReader(s))
	for {
		tt := z.Next()
		//log.Println("tt", tt)
		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			log.Println("err tok")
			return
		case tt == html.StartTagToken || tt == html.SelfClosingTagToken:
			t := z.Token()
			//log.Println("t", t)
			switch t.Data {
			case "img":
				//log.Println("img")
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
				//		log.Println("t.data", t.Data)
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
		if len < 10000 || len > 1000000 {
			continue
		}
		isImg := strings.HasPrefix(resp.Header.Get("Content-Type"), "image")
		if isImg {
			imgs[link] = len
		}
	}
	return imgs
}
