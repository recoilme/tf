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

	"sort"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	bot *tgbotapi.BotAPI
	//botYa botan.Botan
)

func catch(e error) {
	if e != nil {
		log.Println(e.Error)
	}
}

func main() {
	//botYa = botan.New(params.YaToken)
	var err error
	//log.Println(params.Telefeedfile)
	tlgrmtoken, err := ioutil.ReadFile(params.Telefeedfile)
	catch(err)
	tgtoken := strings.Replace(strings.Replace(string(tlgrmtoken), "\n", "", -1), "\r", "", -1)
	//fmt.Printf("token:'%s'", tgtoken)
	//tgtoken := strings.Replace(tgtoken, "\r", "", -1)
	bot, err = tgbotapi.NewBotAPI(tgtoken)
	catch(err)

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.CallbackQuery != nil && update.CallbackQuery.Message != nil {
			/*
				msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID,
						update.CallbackQuery.Message.MessageID,
						"send me bot token")
					bot.Send(msg)
			*/
			//callback
			//delete/feeds/dc28dee259a7d5f0ab73e9eaad050c23
			//delete/pubNames/dnodnaru
			//fmt.Printf("Callback:%+v\nCalb:%+v\n", update.CallbackQuery.From, update.CallbackQuery.Message)
			data := update.CallbackQuery.Data
			msgCancel := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID,
				update.CallbackQuery.Message.MessageID,
				"ZzZzZzzz ...")
			if strings.HasPrefix(data, "delete"+params.Feed) {
				feed := strings.Replace(data, "delete"+params.Feed, "", -1)
				b := httputils.HttpGet(params.Feeds+feed, nil)
				if b != nil {
					url := string(b)
					log.Println("delete " + url)
					pubFind(update.CallbackQuery.Message, "delete "+url, int64(update.CallbackQuery.From.ID))
					bot.Send(msgCancel)
					//usersub(params.Feed+feed, update.CallbackQuery.From.ID, true)
				}
			} else {
				if strings.HasPrefix(data, "delete"+params.PubNames) {
					screenname := strings.Replace(data, "delete"+params.PubNames, "", -1)
					pubFind(update.CallbackQuery.Message, "delete https://vk.com/"+screenname, int64(update.CallbackQuery.From.ID))
					log.Println("update.CallbackQuery.From.ID", update.CallbackQuery.From.ID)
					//usersub(params.PubNames+screenname, update.CallbackQuery.From.ID, true)
					bot.Send(msgCancel)
				} else {
					if strings.Contains(data, "_!_") {
						parts := strings.Split(data, "_!_")
						cmd := parts[0]
						cmdval := parts[1]
						switch cmd {
						case "channel":
							switch cmdval {
							case "new":
								msgNewCh := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID,
									update.CallbackQuery.Message.MessageID,
									params.NewChannel)
								bot.Send(msgNewCh)
							case "delete":
								msgDel := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID,
									update.CallbackQuery.Message.MessageID,
									"delete")

								if len(parts) > 2 {
									chanName := parts[2]

									url := params.UserName + chanName
									b := httputils.HttpGet(url, nil)
									var chat *tgbotapi.Chat
									json.Unmarshal(b, &chat)
									if chat != nil {
										subs := usersub("", chat.ID, true)
										if len(subs) > 0 {
											msgDel.Text = "Channel @" + chanName + " have subscriptions\nDelete urls before delete channel!"
										} else {
											addChannel(update.CallbackQuery.Message.Chat.ID, chat, true)
											msgDel.Text = "deleted @" + chanName
										}
									} else {
										msgDel.Text = "@" + chanName + " not found("
									}
								}
								bot.Send(msgDel)

							case "list":
								msgList := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID,
									update.CallbackQuery.Message.MessageID,
									"list\n")
								msgList.DisableWebPagePreview = true
								if len(parts) > 2 {
									chanName := parts[2]

									url := params.UserName + chanName
									b := httputils.HttpGet(url, nil)
									var chat *tgbotapi.Chat
									json.Unmarshal(b, &chat)
									if chat != nil {
										subs := usersub("", chat.ID, true)
										cmds := subs2cmds(subs)
										var txt = strings.Replace(params.SubsHelp, "channelname", chanName, -1) + "\n\nList of urls of @" + chanName + ":\n\n"
										for _, v := range cmds {
											txt = txt + strings.Replace(v, "delete ", "", -1) + "\n"
										}
										msgList.Text = txt + "\n"
									}
									bot.Send(msgList)
								}

							}
						default:
							bot.Send(msgCancel)
						}

					} else {
						//unknown cmd
						bot.Send(msgCancel)
					}
				}
			}

		} else {
			if update.Message == nil {
				//fmt.Println("ignore")
				continue
			}
			switch update.Message.Text {
			case "/start":
				//botYa.Track(update.Message.From.ID, nil, "start")
				user := update.Message.From
				if userNew(user) {
					m := tgbotapi.NewMessage(update.Message.Chat.ID, params.Hello)
					m.DisableWebPagePreview = true
					bot.Send(m)
				} else {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, params.SomeErr))
				}
			case "/top":
				//botYa.Track(update.Message.From.ID, nil, "top")
				m := tgbotapi.NewMessage(update.Message.Chat.ID, params.TopLinks)
				m.DisableWebPagePreview = true
				bot.Send(m)
			case "/rateme":
				//botYa.Track(update.Message.From.ID, nil, "rateme")
				m := tgbotapi.NewMessage(update.Message.Chat.ID, params.Rate)
				m.DisableWebPagePreview = true
				bot.Send(m)

			case "/help":
				//botYa.Track(update.Message.From.ID, nil, "help")
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "http://telegra.ph/telefeedbot-05-12"))
			case "/channels":
				//botYa.Track(update.Message.From.ID, nil, "channels")
				var cmds = make(map[string]string)
				cmds["channel_!_new"] = "new channel"
				url := params.Channels + strconv.FormatInt(update.Message.Chat.ID, 10)

				body := httputils.HttpGet(url, nil)
				channels := make(map[int64]*tgbotapi.Chat)
				json.Unmarshal(body, &channels)
				for _, channel := range channels {
					cmds["channel_!_delete_!_"+channel.UserName] = "delete @" + channel.UserName
					cmds["channel_!_list_!_"+channel.UserName] = "list of urls of @" + channel.UserName
				}

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Instruction: http://telegra.ph/telefeedbot-05-12\n\nYour channels:\n")
				msg.DisableWebPagePreview = true
				msg.ReplyMarkup = createButtons(cmds)
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)
			case "/list":
				//botYa.Track(update.Message.From.ID, nil, "list")
				//var cmds = make(map[string]string)
				//fmt.Printf("fromid:%d: %d\n", update.Message.From.ID, update.Message.Chat.ID)
				subs := usersub("", int64(update.Message.From.ID), true)
				//var s = "Subscriptions (send 'delete http://..' - for unsubscribe):\n"
				cmds := subs2cmds(subs)
				if len(cmds) == 0 {
					m := tgbotapi.NewMessage(update.Message.Chat.ID, "No feeds..\n\n"+params.Hello)
					m.DisableWebPagePreview = true
					bot.Send(m)
				} else {
					//msg := update.Message
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Subscriptions (press button bellow for unsubscribe):\n")
					msg.ReplyMarkup = createButtons(cmds)
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
				}
			case "/subs":
				//answ, _ := botYa.Track(update.Message.From.ID, nil, "subs")
				//fmt.Printf("arcw:%+v\n", answ)
				subs := usersub("", int64(update.Message.From.ID), true)
				cmds := subs2cmds(subs)
				msgList := tgbotapi.NewMessage(update.Message.Chat.ID, "")
				var txt = "List of urls:\nSend delete url(s) for unsubscribe\n\n"
				for _, v := range cmds {
					txt = txt + strings.Replace(v, "delete ", "", -1) + "\n"
				}
				msgList.Text = txt + "\n"
				bot.Send(msgList)
			default:
				//botYa.Track(update.Message.From.ID, nil, "subscribe")
				msg := update.Message.Text
				pubFind(update.Message, msg, int64(update.Message.From.ID))
			}
		}
	}

}

func subs2cmds(subs map[string]bool) map[string]string {
	var cmds = make(map[string]string)
	for k, _ := range subs {
		log.Println(k)
		if strings.Contains(k, params.PubNames) {
			cmd := "delete https://vk.com/" + strings.Replace(k, params.PubNames, "", -1)
			key := "delete" + k
			cmds[key] = cmd
		}
		if strings.Contains(k, params.Feed) {
			b := httputils.HttpGet(params.Api+k, nil)
			if b != nil {
				cmd := "delete " + string(b)
				key := "delete" + k
				cmds[key] = cmd
			}
		}
	}
	log.Println("cmds:", cmds)
	return cmds
}

func createButtons(buttonsCmds map[string]string) tgbotapi.InlineKeyboardMarkup {
	var buttons [][]tgbotapi.InlineKeyboardButton
	var keys []string
	for k := range buttonsCmds {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		choice := buttonsCmds[k]
		cleanedChoice := strings.TrimSpace(choice)
		cleanedChoice = strings.Replace(cleanedChoice, "\n", "", -1)

		button := tgbotapi.NewInlineKeyboardButtonData(cleanedChoice, k)
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(button))
	}
	buttonCancel := tgbotapi.NewInlineKeyboardButtonData("Cancel", "cancel")
	buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(buttonCancel))
	buttonsRow := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	return buttonsRow
}

func userNew(user *tgbotapi.User) bool {

	//curl localhost:5000/bolt/pubSubTg/146445941
	//{"215921701":true,"377950061":true}
	//858 banned
	urlUsr := params.Users + strconv.Itoa(user.ID)
	log.Println("userNew", urlUsr)
	b, _ := json.Marshal(user)
	httputils.HttpPut(params.UserName+user.UserName, nil, b)
	res := httputils.HttpPut(urlUsr, nil, b)
	//telefeedbot
	if user.ID > 0 {
		pubSubTgAdd(146445941, "telefeedbot", nil, false, int64(user.ID))
	}
	return res
}

func channelNew(chat *tgbotapi.Chat) bool {
	url := params.Users + strconv.FormatInt(chat.ID, 10)
	log.Println("channelNew", url)
	b, _ := json.Marshal(chat)
	httputils.HttpPut(params.UserName+chat.UserName, nil, b)
	return httputils.HttpPut(url, nil, b)
}

func pubFind(msg *tgbotapi.Message, txt string, userid int64) {
	log.Println("pubFind")
	var delete = false
	var tmp = strings.Replace(txt, "\n", " ", -1)
	tmp = strings.Replace(tmp, "\r", "", -1)
	tmp = strings.TrimSpace(tmp)
	words := strings.Split(tmp, " ")

	for i := range words {
		var word = strings.TrimSpace(words[i])
		if word == "delete" || word == "Delete" {
			delete = true
			continue
		}
		if strings.HasPrefix(word, "@") {
			chanName := strings.Replace(word, "@", "", -1)

			url := params.UserName + chanName
			b := httputils.HttpGet(url, nil)
			var chat *tgbotapi.Chat
			json.Unmarshal(b, &chat)
			if chat != nil {
				userChannelsUrl := params.Channels + strconv.FormatInt(userid, 10)

				userChannelsbody := httputils.HttpGet(userChannelsUrl, nil)
				userChannels := make(map[int64]*tgbotapi.Chat)
				json.Unmarshal(userChannelsbody, &userChannels)
				if userChannels[chat.ID] != nil {
					userid = chat.ID
				} else {
					bot.Send(tgbotapi.NewMessage(userid, chanName+" not yours"))
				}

			}
			continue
		}
		if strings.HasPrefix(word, "http") == false {
			//default sheme is https
			word = "https://" + word
		}
		urls, err := url.Parse(word)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Rss feed on domain:'"+word+"'\n"+params.NotFound+params.Example))
			return
		}
		mainDomain, _ := publicsuffix.EffectiveTLDPlusOne(urls.Host)

		switch mainDomain {
		case "t.me":
			parts := strings.Split(urls.Path, "/")
			if len(parts) > 1 {
				channelName := "@" + parts[len(parts)-1]
				m := tgbotapi.NewMessageToChannel(channelName, "Ok")
				m.DisableWebPagePreview = true
				reply, err := bot.Send(m)
				if err != nil {
					s := err.Error()
					if strings.Contains(s, "orbidden") {
						m := tgbotapi.NewMessage(msg.Chat.ID, "Add @telefeedbot as admin 2 channel: "+channelName)
						bot.Send(m)
					} else {
						m := tgbotapi.NewMessage(msg.Chat.ID, s)
						bot.Send(m)
					}
				} else {

					channel := reply.Chat
					//fmt.Printf("Reply:%+v\n", reply.Chat.ID)
					addChannel(msg.Chat.ID, channel, false)
				}
			}
		case "twitter.com":
			parts := strings.Split(urls.Path, "/")
			for _, part := range parts {
				if part != "" {
					findFeed("https://twitrss.me/twitter_user_to_rss/?user="+part, msg, delete, userid)
				}
			}
		case "instagram.com":
			parts := strings.Split(urls.Path, "/")
			for _, part := range parts {
				if part != "" {
					findFeed("https://web.stagram.com/rss/n/"+part, msg, delete, userid)
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
								pubSubTgAdd(groupVk.Gid, groupVk.ScreenName, msg, delete, userid)
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
						pubSubTgAdd(groupDb.Gid, groupDb.ScreenName, msg, delete, userid)
					}
				}
			}
		default:
			findFeed(word, msg, delete, userid)
		}
	}
}

func addChannel(userId int64, channel *tgbotapi.Chat, isDelete bool) {
	if channel == nil {
		return
	}
	url := params.Channels + strconv.FormatInt(userId, 10)

	body := httputils.HttpGet(url, nil)
	channels := make(map[int64]*tgbotapi.Chat)
	json.Unmarshal(body, &channels)
	channels[channel.ID] = channel
	delete(channels, channel.ID)

	if !isDelete {
		channels[channel.ID] = channel
	}
	log.Println("channels ", channels)

	data, err := json.Marshal(channels)
	if err == nil {
		result := httputils.HttpPut(url, nil, data)
		if result == true {
			if isDelete {
				bot.Send(tgbotapi.NewMessage(userId, "👍 Removed: "+channel.UserName+"\n\n"))
			} else {
				//add channel as User
				if channelNew(channel) {
					bot.Send(tgbotapi.NewMessage(userId, channel.UserName+" 👍\n\nUse /channels for list of channels\n\nSend @"+
						channel.UserName+" http://url for add url 2 channel"))
				}
			}
		}
	}
}

func findFeed(word string, msg *tgbotapi.Message, isDelete bool, userid int64) {
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
		feedSubTgAdd(feedlink, msg, isDelete, userid)
	} else {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, word+"\n"+params.NotFound))
	}
}

func feedSubTgAdd(feedlink string, msg *tgbotapi.Message, isDelete bool, userid int64) {
	url := params.FeedSubs + GetMD5Hash(feedlink)
	log.Println("feedSubTgAdd", url)
	body := httputils.HttpGet(url, nil)
	users := make(map[int64]bool)
	json.Unmarshal(body, &users)
	delete(users, userid)
	if !isDelete {
		users[userid] = true
	}
	log.Println("feedSubTgAdd users ", users)

	//user subs
	usersub(params.Feed+GetMD5Hash(feedlink), userid, isDelete)

	data, err := json.Marshal(users)
	if err == nil {
		log.Println("feedSubTgAdd data ", string(data))
		result := httputils.HttpPut(url, nil, data)
		if result == true {
			if isDelete {
				bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "👍 Removed: "+feedlink+"\n\n"))
			} else {
				bot.Send(tgbotapi.NewMessage(msg.Chat.ID, feedlink+" 👍\n\n"+
					params.Psst))
			}
		}
	}
}

func usersub(url string, userid int64, isDelete bool) map[string]bool {
	suburl := params.UserSubs + strconv.FormatInt(userid, 10)
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
	hash := md5.Sum([]byte(strings.TrimSpace(text)))
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

func pubSubTgAdd(gId int, screenName string, msg *tgbotapi.Message, isDelete bool, userid int64) {

	gid := strconv.Itoa(gId)
	url := params.Subs + gid
	log.Println("pubSubTgAdd", url)
	body := httputils.HttpGet(url, nil)

	users := make(map[int64]bool)
	json.Unmarshal(body, &users)
	delete(users, userid)
	if !isDelete {
		users[userid] = true
	}
	log.Println("pubSubTgAdd users ", users)
	data, err := json.Marshal(users)
	if err == nil {
		log.Println("pubSubTgAdd data ", string(data))
		result := httputils.HttpPut(url, nil, data)
		if result == true {
			if msg != nil {
				if isDelete {
					bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "👍 Removed: https://vk.com/"+screenName+"\n"))
				} else {
					bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "👍 https://vk.com/"+screenName+"\n"+
						params.Psst))
				}
			}
			//user subs
			//log.Println("👍 Removed", userid, isDelete)
			usersub(params.PubNames+screenName, userid, isDelete)
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
