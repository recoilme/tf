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
	"time"

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
	wrbot.Debug = true
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v+\n", time.Now())
	go forever()
	select {} // block forever
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

	last := len(feed.Items) - 1
	for i := range feed.Items {
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

func pubpost(domain string, p *gofeed.Item, users map[int]bool) {
	var vkcnt int64 = -1001067277325 //myakotka
	log.Println("pubpost", p.GUID)
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
