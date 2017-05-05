package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/recoilme/httputils"
	"github.com/recoilme/vkapi"

	"gopkg.in/telegram-bot-api.v4"
)

const (
	api      = "http://badtobefat.ru/bolt"
	users    = "/usertg/"
	pubNames = "/pubNames/"
	pubSubTg = "/pubSubTg/"
	someErr  = "Something going wrong. Try later.. Ошибка, мать её!"
	hello    = "🇬🇧 Send me links to public pages from vk.com, and I will send you new articles.\n🇷🇺 Отправь мне ссылки на общедоступные страницы c vk.com, и я буду присылать тебе новые статьи.\nExample: https://vk.com/myakotkapub\nContacts: @recoilme"
	psst     = "🇬🇧 Psst. As soon as there are new articles here - I will immediately send them\n🇷🇺 Псст. Как только появятся новые статьи здесь -  я их сразу пришлю"
)

var (
	bot *tgbotapi.BotAPI
)

func catch(e error) {
	if e != nil {
		log.Panic(e.Error)
	}
}

func init() {
	log.SetOutput(ioutil.Discard)
	http.DefaultClient.Transport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 1 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 1 * time.Second,
	}
	http.DefaultClient = &http.Client{
		Timeout: time.Second * 10,
	}
}

func main() {
	var err error
	tlgrmtoken, err := ioutil.ReadFile("tokentg")
	catch(err)
	tgtoken := strings.Replace(string(tlgrmtoken), "\n", "", -1)
	bot, err = tgbotapi.NewBotAPI(tgtoken)
	catch(err)

	bot.Debug = true

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
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, hello))
			} else {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, someErr))
			}
		default:
			msg := update.Message.Text
			pubFind(update.Message, msg)
		}
	}

}

func userNew(user *tgbotapi.User) bool {
	url := api + users + strconv.Itoa(user.ID)
	log.Println("userNew", url)
	b, _ := json.Marshal(user)
	return httputils.HttpPut(url, nil, b)
}

func pubFind(msg *tgbotapi.Message, txt string) {
	log.Println("pubFind")
	words := strings.Split(txt, " ")
	for i := range words {
		word := words[i]
		urls, err := url.Parse(word)
		if err != nil {
			return
		}
		switch urls.Host {
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
								pubSubTgAdd(groupVk, msg)
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
						pubSubTgAdd(groupDb, msg)
					}
				}
			}
		}
	}
}

func pubDbGet(domain string) (group vkapi.Group) {
	log.Println("pubDbGet")
	url := api + pubNames + domain
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
	return httputils.HttpPut(api+pubNames+domain, nil, b)
}

func pubSubTgAdd(group vkapi.Group, msg *tgbotapi.Message) {

	gid := strconv.Itoa(group.Gid)
	url := api + pubSubTg + gid
	log.Println("pubSubTgAdd", url)
	body := httputils.HttpGet(url, nil)

	if body != nil {
		users := make(map[int]bool)
		json.Unmarshal(body, &users)
		delete(users, msg.From.ID)
		users[msg.From.ID] = true
		log.Println("pubSubTgAdd users ", users)
		data, err := json.Marshal(users)
		if err == nil {
			log.Println("pubSubTgAdd data ", string(data))
			result := httputils.HttpPut(url, nil, data)
			if result == true {
				bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Domain:'"+group.ScreenName+"'\n"+psst))
			}
		}
	}
}
