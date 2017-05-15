package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/publicsuffix"

	"github.com/mmcdole/gofeed"
	"github.com/recoilme/tf/httputils"
	"github.com/recoilme/tf/params"
	"github.com/recoilme/tf/vkapi"

	"gopkg.in/telegram-bot-api.v4"
)

var (
	bot *tgbotapi.BotAPI
)

func catch(e error) {
	if e != nil {
		log.Panic(e.Error)
	}
}

func main() {
	var err error
	tlgrmtoken, err := ioutil.ReadFile(params.Telefeedfile)
	catch(err)
	tgtoken := strings.Replace(string(tlgrmtoken), "\n", "", -1)
	bot, err = tgbotapi.NewBotAPI(tgtoken)
	catch(err)

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		if update.Message.Chat.ID == -1001067277325 {
			//skip updates from myakotkapub
			continue
		}
		switch update.Message.Text {
		case "/start":
			user := update.Message.From
			if userNew(user) {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, params.Hello))
			} else {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, params.SomeErr))
			}
		case "/help":
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "http://telegra.ph/telefeedbot-05-12"))
		case "/list":
			subs := usersub("", update.Message.From.ID, true)
			var s string
			for k, _ := range subs {
				if strings.Contains(k, params.Publics) {
					s = s + "\nhttps://vk.com/" + strings.Replace(k, params.Publics, "", -1)
				}
				if strings.Contains(k, params.Feeds) {
					b := httputils.HttpGet(k, nil)
					if b != nil {
						s = s + "\n" + string(b)
					}
				}
			}
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, s))
		default:
			msg := update.Message.Text
			pubFind(update.Message, msg)
		}
	}

}

func userNew(user *tgbotapi.User) bool {
	url := params.Users + strconv.Itoa(user.ID)
	log.Println("userNew", url)
	b, _ := json.Marshal(user)
	httputils.HttpPut(params.UserName+user.UserName, nil, b)
	return httputils.HttpPut(url, nil, b)
}

func pubFind(msg *tgbotapi.Message, txt string) {
	log.Println("pubFind")
	var delete = false
	words := strings.Split(txt, " ")
	for i := range words {
		var word = words[i]
		if word == "delete" {
			delete = true
			continue
		}
		if strings.HasPrefix(word, "http") == false {
			//default sheme is https
			word = "https://" + word
		}
		urls, err := url.Parse(word)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Domain:'"+word+"'\n"+params.NotFound+params.Example))
			return
		}
		mainDomain, _ := publicsuffix.EffectiveTLDPlusOne(urls.Host)

		switch mainDomain {
		case "twitter.com":
			parts := strings.Split(urls.Path, "/")
			for _, part := range parts {
				if part != "" {
					findFeed("https://twitrss.me/twitter_user_to_rss/?user="+part, msg, delete)
				}
			}
		case "instagram.com":
			parts := strings.Split(urls.Path, "/")
			for _, part := range parts {
				if part != "" {
					findFeed("https://web.stagram.com/rss/n/"+part, msg, delete)
				}
			}
		case "vk.com":
			parts := strings.Split(urls.Path, "/")
			for j := range parts {
				if parts[j] != "" {
					domain := parts[j]
					log.Println(domain)
					//bot.Send(tgbotapi.NewMessage(chatId, "Found vk domain:'"+parts[j]+"'"))
					groupDb := pubDbGet(domain)
					if groupDb.Gid == 0 {
						// public not found
						groups := vkapi.GroupsGetById(domain)
						if len(groups) > 0 {
							// we have group
							groupVk := groups[0]
							// save group to DB
							if pubDbSet(groupVk) {
								// new group set
								pubSubTgAdd(groupVk, msg, delete)
							} else {
								// group not set
								bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Error create domain:'"+domain+"'"))
							}
						} else {
							// group not found
							bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Error vk domain:'"+domain+"'"+" not found"))
						}

					} else {
						// public exists
						pubSubTgAdd(groupDb, msg, delete)
					}
				}
			}
		default:
			findFeed(word, msg, delete)
		}
	}
}

func findFeed(word string, msg *tgbotapi.Message, isDelete bool) {
	log.Println("word", word)
	var feedlink = getFeedLink(word)
	if feedlink == "" {
		log.Println("feedlink", feedlink)
		rss := rssExtract(word)
		if rss != "" {
			log.Println("rss", rss)
			feedlink = getFeedLink(rss)
			log.Println("feedlink", feedlink)
		}
	}
	if feedlink != "" {
		feedkey := GetMD5Hash(feedlink)
		//create feed or overwrite
		httputils.HttpPut(params.Feeds+feedkey, nil, []byte(feedlink))
		feedSubTgAdd(feedlink, msg, isDelete)
	} else {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Domain: "+word+"\n"+params.NotFound))
	}
}

func feedSubTgAdd(feedlink string, msg *tgbotapi.Message, isDelete bool) {
	url := params.FeedSubs + GetMD5Hash(feedlink)
	log.Println("feedSubTgAdd", url)
	body := httputils.HttpGet(url, nil)
	users := make(map[int]bool)
	json.Unmarshal(body, &users)
	delete(users, msg.From.ID)
	if !isDelete {
		users[msg.From.ID] = true
	}
	log.Println("feedSubTgAdd users ", users)

	//user subs
	usersub(params.Feeds+GetMD5Hash(feedlink), msg.From.ID, isDelete)

	data, err := json.Marshal(users)
	if err == nil {
		log.Println("feedSubTgAdd data ", string(data))
		result := httputils.HttpPut(url, nil, data)
		if result == true {
			if isDelete {
				bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Ups! Removed domain: "+feedlink+"\n"))
			} else {
				bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Wow! New domain: "+feedlink+"\n"+
					params.Psst))
			}
		}
	}
}

func usersub(url string, userid int, isDelete bool) map[string]bool {
	suburl := params.UserSubs + strconv.Itoa(userid)
	bodysub := httputils.HttpGet(suburl, nil)
	subs := make(map[string]bool)
	json.Unmarshal(bodysub, &subs)
	delete(subs, url)
	if !isDelete {
		subs[url] = true
	}
	if url == "" {
		return subs
	}
	bsubs, _ := json.Marshal(subs)
	httputils.HttpPut(suburl, nil, bsubs)
	return subs
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func getFeedLink(link string) (feedlink string) {
	var defHeaders = make(map[string]string)
	defHeaders["User-Agent"] = "script::recoilme:v1"
	defHeaders["Authorization"] = "Client-ID 4191ffe3736cfcb"
	b := httputils.HttpGet(link, defHeaders)
	if b == nil {
		return feedlink
	}
	fp := gofeed.NewParser()
	feed, err := fp.Parse(bytes.NewReader(b))
	if err != nil {
		return feedlink
	}
	if len(feed.Items) > 0 {
		feedlink = link
	}
	return feedlink
}

func pubDbGet(domain string) (group vkapi.Group) {
	log.Println("pubDbGet")
	url := params.Publics + domain
	body := httputils.HttpGet(url, nil)
	if body != nil {
		json.Unmarshal(body, &group)
	}
	return
}

func pubDbSet(group vkapi.Group) bool {
	log.Println("pubDbSet")
	domain := group.ScreenName
	b, err := json.Marshal(group)
	if err != nil {
		return false
	}
	return httputils.HttpPut(params.Publics+domain, nil, b)
}

func pubSubTgAdd(group vkapi.Group, msg *tgbotapi.Message, isDelete bool) {

	gid := strconv.Itoa(group.Gid)
	url := params.Subs + gid
	log.Println("pubSubTgAdd", url)
	body := httputils.HttpGet(url, nil)

	users := make(map[int]bool)
	json.Unmarshal(body, &users)
	delete(users, msg.From.ID)
	if !isDelete {
		users[msg.From.ID] = true
	}
	log.Println("pubSubTgAdd users ", users)
	data, err := json.Marshal(users)
	if err == nil {
		log.Println("pubSubTgAdd data ", string(data))
		result := httputils.HttpPut(url, nil, data)
		if result == true {
			if isDelete {
				bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Ups! Removed domain: https://vk.com/"+group.ScreenName+"\n"))
			} else {
				bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Wow! New domain: https://vk.com/"+group.ScreenName+"\n"+
					params.Psst))
			}
			//user subs
			usersub(params.Publics+group.ScreenName, msg.From.ID, isDelete)
		}
	}
}

func rssExtract(link string) string {
	var rss string
	var defHeaders = make(map[string]string)
	defHeaders["User-Agent"] = "script::recoilme:v1"
	defHeaders["Authorization"] = "Client-ID 4191ffe3736cfcb"
	b := httputils.HttpGet(link, defHeaders)
	if b == nil {
		return rss
	}
	//s := `<link rel="alternate" type="application/rss+xml" href="https://vc.ru/feed">`
	doc, err := html.Parse(bytes.NewReader(b)) //strings.NewReader(s))
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
	return rss
}
