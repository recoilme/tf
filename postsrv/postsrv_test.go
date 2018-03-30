package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"testing"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/recoilme/tf/httputils"
	"github.com/recoilme/tf/params"
	"github.com/recoilme/tf/vkapi"
)

func TestParse(t *testing.T) {
	/*
		initBot()
		userId := int64(-1001099791579)
		txt := "42"
		//m := tgbotapi.NewMessageToChannel("@memefeed", txt)
		m := tgbotapi.NewMessage(userId, trimTo(txt, 3500))
		m.DisableNotification = true
		m.DisableWebPagePreview = true
		res, err := bot.Send(m)
		fmt.Printf("res %+v err %s", res, err)*/
	//parse()
	s := `pic.twitter.com/Q9RRu2UUSy`

	//r := trimTo(s, 200)

	fmt.Println(len([]rune(s)))

	t0 := time.Now().Unix()
	time.Sleep(2 * time.Second)
	t1 := time.Now().Unix()
	sec := int64(t1 - t0)

	fmt.Printf("The call took %v to run.\n", sec)
	vir := virality(1496643621, 190, 30, 3) //54
	fmt.Printf("vir:%d", vir)
}

func testHead(t *testing.T) {
	link := "http://feeds.feedburner.com/~ff/ettoday/realtime?a=5V9lxn-vAwo:_Qjs5D4yEIg:yIl2AUoC8zA"
	resp, err := http.Head(link)
	fmt.Printf("1.3 %s Feed: %s %+v\n", time.Now().Format("15:04:05"), err, resp)

}

func TestParse2(t *testing.T) {
	link := "http://feeds.feedburner.com/~ff/ettoday/realtime?a=5V9lxn-vAwo:_Qjs5D4yEIg:yIl2AUoC8zA"
	resp, err := http.Head(link)
	fmt.Printf("1.3 %s Feed: %s %+v\n", time.Now().Format("15:04:05"), err, resp)

	len := httputils.HttpHead(link, nil)
	fmt.Printf("%n\n", len)
	initBot()
	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)
	//u := tgbotapi.NewUpdate(0)
	//u.Timeout = 60

	//updates, err := bot.GetUpdates(u)

	//fmt.Println("%s err:%s\n", updates, err)
	msgdub := tgMessage{
		msgtype:  "",
		userId:   366035536,
		txt:      "1 ☝",
		bytes:    nil,
		fileName: "",
	}
	_ = msgdub

	msg := tgbotapi.NewPhotoShare(366035536, "AgADAgADLqgxG6HxwUnyoQed8cyvC8zCDw4ABISHNR9Njvu61YQAAgI")
	msg.Caption = "Test"
	bot.Send(msg)

}

func TestShortUrl(t *testing.T) {
	u := shortenUrl("http://ya.ru")
	fmt.Println("%s:\n", u)
}

func TestUrl(t *testing.T) {
	//var wg sync.WaitGroup
	//var feedUsers map[int64]bool
	//feedUsers[1263310] = true
	//wg.Add(1)
	url := "https://www.reddit.com/r/gifs/top/.rss"
	//go getFeedPosts(url, feedUsers, &wg)

	//wg.Wait()
	var defHeaders = make(map[string]string)
	defHeaders["User-Agent"] = "script::recoilme:v1"
	defHeaders["Authorization"] = "Client-ID 4191ffe3736cfcb"
	b := httputils.HttpGet(url, defHeaders)
	if b == nil {
		fmt.Printf("b is nil\n")
	} else {
		s := string(b)
		fmt.Printf("b is %s\n", s)
	}
}

func TestPubTop(t *testing.T) {

	type kv struct {
		Key   vkapi.Group
		Value int
	}

	var ss []kv

	publics := getPubNames("desc")
	for _, pubName := range publics {
		b := httputils.HttpGet(params.Publics+pubName, nil)
		if b != nil {
			var public vkapi.Group
			err := json.Unmarshal(b, &public)
			if err == nil {
				pubusers := pubUsers(public)
				if len(pubusers) == 0 {
					continue
				}
				ss = append(ss, kv{public, len(pubusers)})
				//fmt.Printf("%s %d\n", pubName, len(pubusers))
				//go MakeRequestDeferred(i, "22", nil, "", nil)

			}
		}
	}

	type kvf struct {
		Key   string
		Value int
	}

	var ff []kvf
	feeds := getFeedNames()

	for _, hash := range feeds {
		//log.Println("getfeed", url, hash)
		b := httputils.HttpGet(params.Feeds+hash, nil)
		if b != nil {
			url := string(b)
			feedUsers := feedUsers(hash)
			ff = append(ff, kvf{url, len(feedUsers)})
		}
	}

	sort.Slice(ff, func(i, j int) bool {
		return ff[i].Value > ff[j].Value
	})
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	for i, kv := range ss {
		if i == 50 {
			break
		}
		fmt.Printf("%d https://vk.com/%s\n%s\n\n", i, kv.Key.ScreenName, trimTo(kv.Key.Description, 100))
	}

	for i, kvf := range ff {
		if i == 50 {
			break
		}
		fmt.Printf("%d %s %d\n\n", i, kvf.Key, kvf.Value)
	}
}
