package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/recoilme/tf/httputils"
	"github.com/recoilme/tf/params"
	"github.com/recoilme/tf/vkapi"
)

const (
	MaxInt = int(^uint(0) >> 1)
	MinInt = -MaxInt - 1
)

var (
	bot, wrbot  *tgbotapi.BotAPI
	forbidden         = map[int64]bool{}
	lastUID     int64 = 0
	fwdPostTime       = time.Now()
)

func main() {
	//botfile:= "telefeed.bot"
	log.Println("main")
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
	//wrbot.Debug = true
	//bot.Debug = true
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("%v+\n", time.Now())
	go forever()
	select {} // block forever
}

func forever() {
	for {
		//fmt.Printf("%v+\n", time.Now())
		parseVk()
		time.Sleep(1200 * time.Second)
	}
}

func sendAll() {
	var users []string
	url := params.BaseUri + "usertg"
	log.Println("sendAll", url)
	resp, err := http.Post(url, "application/json", nil)
	if err == nil {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(body, &users)

		b, _ := tgbotapi.NewBotAPI("364483768:AAFhyU95D609MLVMQNNFzd3ZOAxIgyIHMN0")
		var start = false
		for _, user := range users {
			if user == "397314920" || start {
				start = true
			} else {
				continue
			}
			uid, _ := strconv.ParseInt(user, 10, 64)
			log.Println(user, uid)

			txt := "Извините, пожалуйста, за перебои в работе сервиса @telefeedbot(.\n\nЕжесекундно в ВКонтакте появляются сотни новых постов, и мы не успеваем их отправлять всем подписчикам всвязи с ограничениями апи телеграм (ну или из-за кривых рук).\n\nПытаюсь найти решение.\n\nСпасибо за Ваше терпение!"
			msg := tgbotapi.NewMessage(uid, txt)
			msg.DisableWebPagePreview = true
			msg.DisableNotification = true
			_, err := b.Send(msg)
			if err == nil {
				log.Println("Ok")
			} else {
				log.Println("Error")
			}
			time.Sleep(1000 * time.Millisecond)
			break

		}
	}

}

func parseVk() {

	domains := getDomainNames()
	for _, domainName := range domains {

		//log.Println(domainName)

		b := httputils.HttpGet(params.Publics+domainName, nil)
		if b != nil {
			var domain vkapi.Group
			err := json.Unmarshal(b, &domain)
			if err == nil {
				log.Println(domain.ScreenName)
				users := domUsers(domain)
				if len(users) > 0 {
					log.Println("saveposts", domain.ScreenName)
					saveposts(domain, users)
				}
			}
		}

		//time.Sleep(100 * time.Millisecond)
	}
}

func domUsers(domain vkapi.Group) (users map[int64]bool) {
	mask := params.Subs + "%d"
	url := fmt.Sprintf(mask, domain.Gid)
	log.Println(url)
	b := httputils.HttpGet(url, nil)
	if b != nil {
		json.Unmarshal(b, &users)
	}
	return users
}

func lastPostIdGet(domain vkapi.Group) int {
	postId := MinInt
	mask := params.LastPost + "%d"
	url := fmt.Sprintf(mask, domain.Gid)
	b := httputils.HttpGet(url, nil)
	if b != nil {
		json.Unmarshal(b, &postId)
	}
	return postId
}

func lastPostIdSet(domain vkapi.Group, lastPostId int) int {
	postId := MinInt
	mask := params.LastPost + "%d"
	url := fmt.Sprintf(mask, domain.Gid)
	b, err := json.Marshal(lastPostId)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		postId = lastPostId
	} else {
		log.Println(err)
	}
	return postId
}

func saveposts(domain vkapi.Group, users map[int64]bool) {
	log.Println(domain)
	var lastPost = lastPostIdGet(domain)
	log.Println("last", lastPost)
	posts := vkapi.WallGet(domain.Gid * (-1))

	last := len(posts) - 1
	if last > 5 {
		last = 5
	}
	for i := range posts {
		if i > last {
			break
		}
		post := posts[last-i]
		if post.Id <= lastPost {
			continue
		}
		lastPost = lastPostIdSet(domain, post.Id)
		//ads
		if post.MarkedAsAds == 1 {
			continue
		}
		if len(post.Attachments) == 0 && post.Text == "" {
			// no text no attachments
			continue
		}
		//fmt.Printf("Post: %+v\n", post)
		url := fmt.Sprintf("http://badtobefat.ru/bolt/%d/%s", post.OwnerID*(-1), fmt.Sprintf("%010d", post.Id))
		b, _ := json.Marshal(post)
		httputils.HttpPut(url, nil, b)
		log.Println(post.Id)
		pubpost(domain, post, users)
		//break
	}
}

// for testing
func getpost() (post vkapi.Post) {
	postid := "126993367/0000001170"

	url := params.BaseUri + postid
	resp, err := http.Get(url)
	if err == nil {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			err := json.Unmarshal(body, &post)
			if err == nil {
				return
			}
		}
	}
	return
}

func getDomainNames() (domainNames []string) {
	url := params.BaseUri + "pubNames/First?cnt=1000000&order=asc"
	log.Println("vkdomains", url)
	resp, err := http.Post(url, "application/json", nil)
	if err == nil {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		err := json.Unmarshal(body, &domainNames)
		if err == nil {
			return
		} else {
			log.Println(err)
		}
	}
	return
}

func vkdomains() (domains []vkapi.Group) {
	var domainNames []string
	url := params.BaseUri + "pubNames/First?cnt=1000000&order=asc"
	log.Println("vkdomains", url)
	resp, err := http.Post(url, "application/json", nil)
	if err == nil {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		err := json.Unmarshal(body, &domainNames)
		if err == nil {
			for i := range domainNames {
				domainName := domainNames[i]
				log.Println("domainName", domainName)
				b := httputils.HttpGet(params.Publics+domainName, nil)
				if b != nil {
					var domain vkapi.Group
					err := json.Unmarshal(b, &domain)
					if err == nil {
						domains = append(domains, domain)
					}
				}
			}
		} else {
			log.Println(err)
		}
	}
	return
}

func Forward(users map[int64]bool, fwdmsg tgbotapi.Message, e error, disableWebPagePreview bool, storeId int64) {
	if len(users) <= 0 {
		return
	}

	//time.Sleep(1 * time.Second)
	//fmt.Printf("msg: %+v\n", fwdmsg)
	if e != nil {
		se := e.Error()
		fmt.Printf("Error post to myakotka: %+v\n", se)
		if strings.Contains(se, "orbidden") {
			forbidden[storeId] = true
		}
		return
	}
	var postNum = 0
	for user := range users {
		uid := int64(user)
		if forbidden[uid] == true {
			continue
		}
		if uid == storeId {
			continue
		}

		if lastUID == uid {
			fromLastPost := int64((time.Now().Sub(fwdPostTime)).Seconds() * 1000)
			//fmt.Printf("From Last post%d\n", fromLastPost)
			dif := time.Duration(1200-fromLastPost) * time.Millisecond
			//fmt.Printf("Sleep time ms %s\n", dif)
			time.Sleep(dif)
		}
		lastUID = uid
		var err error

		var msgType = "text"
		if fwdmsg.Photo != nil {
			msgType = "photo"
		} else if fwdmsg.Video != nil {
			msgType = "video"
		} else if fwdmsg.Document != nil {
			msgType = "doc"
		}
		switch msgType {
		case "photo":
			var photos []tgbotapi.PhotoSize
			photos = *fwdmsg.Photo
			var fileID = ""
			for _, photo := range photos {
				fileID = photo.FileID
			}
			if fileID == "" {
				continue
			}
			msg := tgbotapi.NewPhotoShare(uid, fileID)
			msg.Caption = fwdmsg.Caption
			msg.DisableNotification = true
			_, err = bot.Send(msg)
		case "video":
			vid := *fwdmsg.Video
			msg := tgbotapi.NewVideoShare(uid, vid.FileID)
			msg.DisableNotification = true

			msg.Caption = fwdmsg.Caption
			_, err = bot.Send(msg)
		case "doc":
			doc := fwdmsg.Document
			msg := tgbotapi.NewDocumentShare(uid, doc.FileID)
			msg.Caption = fwdmsg.Caption
			msg.DisableNotification = true
			_, err = bot.Send(msg)
		default:
			//txt
			msg := tgbotapi.NewMessage(uid, fwdmsg.Text)
			msg.DisableNotification = true
			msg.DisableWebPagePreview = disableWebPagePreview
			_, err = bot.Send(msg)
		}

		fwdPostTime = time.Now()
		if err != nil {
			s := err.Error()
			fmt.Printf("Error post to user:%d %s\n", uid, s)
			if strings.Contains(s, "Many") {
				fmt.Printf("%s to many request\n", time.Now().Format("15:04:05"))
				time.Sleep(600 * time.Second)
			} else {
				if strings.Contains(s, "orbidden") {
					forbidden[int64(user)] = true
				}
			}
		} else {
			fmt.Printf("%s Ok, uid:%d\n", time.Now().Format("15:04:05"), uid)
		}
		postNum++
		if postNum%10 == 0 {
			time.Sleep(1 * time.Second)
		}
	}
}

func pubpost(domain vkapi.Group, p vkapi.Post, users map[int64]bool) {

	log.Println("pubpost", p.Id)
	if len(users) == 0 {
		return
	}
	//first user - store content
	var storeId int64
	//fmt.Printf("storeid%d len %d\n", storeId, len(users))
	for user := range users {
		if forbidden[user] == true {
			continue
		}
		storeId = user
		break
	}
	if storeId == 0 {
		//all users blocks bot
		return
	}
	//delete(users, storeId)
	//fmt.Printf("storeid%d len %d\n", storeId, len(users))
	time.Sleep(1 * time.Second)
	var t = strings.Replace(p.Text, "&lt;br&gt;", "\n", -1)
	if t != "" {
		t = t + "\n"
	}
	link := fmt.Sprintf("vk.com/wall%d_%d", domain.Gid*(-1), p.Id)
	tag := strings.Replace(domain.ScreenName, ".", "", -1)
	fmt.Printf("%s tag:%s\n", time.Now().Format("15:04:05"), tag)
	appendix := fmt.Sprintf("#%s 🔗 %s", tag, link)
	if len(p.Attachments) == 0 || len([]rune(t)) > 200 {
		msg := tgbotapi.NewMessage(storeId, t+appendix)
		t = ""
		msg.DisableWebPagePreview = true
		msg.DisableNotification = true
		res, err := bot.Send(msg)

		Forward(users, res, err, true, storeId)

	}
	for i := range p.Attachments {
		time.Sleep(1 * time.Second)
		att := p.Attachments[i]
		log.Println(att.Type)
		switch att.Type {
		case "photo":
			if att.Photo.Width < 100 {
				continue
			}
			if att.Photo.Height < 100 {
				continue
			}
			var photo = att.Photo.Photo1280
			if photo == "" {
				photo = att.Photo.Photo604
			}
			log.Println(photo)
			b := httputils.HttpGet(photo, nil)
			if b != nil {
				bb := tgbotapi.FileBytes{Name: photo, Bytes: b}
				msg := tgbotapi.NewPhotoUpload(storeId, bb)
				if i == 0 {
					msg.Caption = t + appendix
				} else {
					msg.Caption = appendix
				}
				msg.DisableNotification = true
				res, err := bot.Send(msg)

				Forward(users, res, err, true, storeId)

			}
		case "video":
			//fmt.Printf("%+v\n", att.Video)
			urlv := fmt.Sprintf("https://vk.com/video%d_%d", att.Video.OwnerID, att.Video.ID)
			if att.Video.Duration > 600 {
				//send url
				msg := tgbotapi.NewMessage(storeId, urlv+"\n"+appendix)
				msg.DisableWebPagePreview = false
				msg.DisableNotification = true
				res, err := bot.Send(msg)

				Forward(users, res, err, false, storeId)

				continue
			}
			b := httputils.HttpGet(urlv, nil)
			if b != nil {
				cnt := string(b)
				var pos360 = strings.Index(cnt, ".360.mp4")
				if pos360 < 0 {
					pos360 = strings.Index(cnt, ".240.mp4")
				}
				if pos360 < 0 || pos360 < 200 {
					break
				}
				poshttp := strings.Index(cnt[pos360-200:], "https") + pos360 - 200 //cnt.find("https:",pos360-200)
				if poshttp > 0 {
					s := strings.Replace(cnt[poshttp:pos360+8], "\\/", "/", -1)
					if s != "" {
						//post video
						vidb := httputils.HttpGet(s, nil)
						bb := tgbotapi.FileBytes{Name: s, Bytes: vidb}
						msg := tgbotapi.NewVideoUpload(storeId, bb)
						msg.Caption = appendix
						msg.DisableNotification = true
						res, err := bot.Send(msg)

						Forward(users, res, err, false, storeId)

					}
				}
			}
		case "doc":
			//fmt.Printf("%+v\n", att.Doc)
			b := httputils.HttpGet(att.Doc.URL, nil)
			if b != nil {
				bb := tgbotapi.FileBytes{Name: "tmp." + att.Doc.Ext, Bytes: b}
				msg := tgbotapi.NewDocumentUpload(storeId, bb)
				msg.Caption = appendix
				msg.DisableNotification = true
				res, err := bot.Send(msg)

				Forward(users, res, err, false, storeId)

			}
		case "link":
			if att.Link.Photo.Photo604 != "" && att.Link.Photo.Width > 400 && att.Link.Photo.Height > 400 {
				//link with photo
				b := httputils.HttpGet(att.Link.Photo.Photo604, nil)
				if b != nil {
					bb := tgbotapi.FileBytes{Name: att.Link.Photo.Photo604, Bytes: b}

					msg := tgbotapi.NewPhotoUpload(storeId, bb)
					msg.Caption = att.Link.Title + "\n" + att.Link.Description + "\n" + att.Link.URL + "\n" + appendix
					msg.DisableNotification = true
					res, err := bot.Send(msg)

					Forward(users, res, err, true, storeId)

				}

			} else {
				var desc = ""
				desc = att.Link.Title + "\n" + att.Link.URL + "\n" + appendix

				msg := tgbotapi.NewMessage(storeId, desc)
				msg.DisableWebPagePreview = false
				msg.DisableNotification = true
				res, err := bot.Send(msg)
				Forward(users, res, err, false, storeId)

			}
		}
	}

}
