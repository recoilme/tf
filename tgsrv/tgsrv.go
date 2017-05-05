package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/url"
	"strconv"
	"strings"

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
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, params.Hello))
			} else {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, params.SomeErr))
			}
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
	return httputils.HttpPut(url, nil, b)
}

func pubFind(msg *tgbotapi.Message, txt string) {
	log.Println("pubFind")
	var delete = false
	words := strings.Split(txt, " ")
	for i := range words {
		word := words[i]
		if word == "delete" {
			delete = true
		}
		urls, err := url.Parse(word)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Domain:'"+word+"'\n"+params.NotFound+params.Example))
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
			bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Domain:'"+word+"'"+params.NotFound))
		}
	}
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
			isNew := "New"
			if isDelete {
				isNew = "Removed"
			}
			bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Wow! "+isNew+" domain: https://vk.com/"+group.ScreenName+"\n"+
				params.Psst+"\n"+params.HowDelete))
		}
	}
}
