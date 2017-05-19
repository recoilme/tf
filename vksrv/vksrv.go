package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
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
	bot, wrbot *tgbotapi.BotAPI
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
		parseVk()
		time.Sleep(1200 * time.Second)
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
	/*
		domains := vkdomains()
		for i := range domains {
			domain := domains[i]
			log.Println(domain.ScreenName)
			users := domUsers(domains[i])
			if len(users) > 0 {
				saveposts(domain, users)
			}
			time.Sleep(1 * time.Second)
		}*/
}

func domUsers(domain vkapi.Group) (users map[int]bool) {
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

func saveposts(domain vkapi.Group, users map[int]bool) {
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
		break
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
	url := params.BaseUri + "pubNames/Last?cnt=1000000&order=desc&vals=false"
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
	url := params.BaseUri + "pubNames/Last?cnt=1000000&order=desc&vals=false"
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

func forward(users map[int]bool, msgID int, e error, storeId int64) {
	if e != nil {
		fmt.Printf("Error post to myakotka: %s\n", e.Error)
		return
	}
	time.Sleep(300 * time.Millisecond)
	var counter = 0
	for user := range users {
		//log.Println(user)
		_, err := bot.Send(tgbotapi.NewForward(int64(user), storeId, msgID))
		if err != nil {
			fmt.Printf("Error post to user:%d %s\n", int64(user), err)
		}
		counter = counter + 1
		if counter%10 == 0 {
			time.Sleep(300 * time.Millisecond)
		}
	}
}

func getStoreId() int64 {
	rand.Seed(time.Now().Unix())
	return params.StoreIds[rand.Intn(len(params.StoreIds))]
}

func pubpost(domain vkapi.Group, p vkapi.Post, users map[int]bool) {
	log.Println("pubpost", p.Id)
	var storeId = getStoreId()
	time.Sleep(2 * time.Second)
	//var vkcnt int64 = -1001067277325 //myakotka
	//var fwd int64 = 366035536        //telefeed
	var t = strings.Replace(p.Text, "&lt;br&gt;", "\n", -1)
	if t != "" {
		t = t + "\n"
	}
	link := fmt.Sprintf("vk.com/wall%d_%d", domain.Gid*(-1), p.Id)
	tag := strings.Replace(domain.ScreenName, ".", "", -1)
	appendix := fmt.Sprintf("#%s 🔗 %s", tag, link)
	if len(p.Attachments) == 0 || len([]rune(t)) > 200 {
		msg := tgbotapi.NewMessage(storeId, t+appendix)
		t = ""
		msg.DisableWebPagePreview = true
		msg.DisableNotification = true
		res, err := wrbot.Send(msg)

		forward(users, res.MessageID, err, storeId)

	}
	for i := range p.Attachments {
		time.Sleep(500 * time.Millisecond)
		storeId = getStoreId()
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
				res, err := wrbot.Send(msg)

				forward(users, res.MessageID, err, storeId)

			}
		case "video":
			//fmt.Printf("%+v\n", att.Video)
			urlv := fmt.Sprintf("https://vk.com/video%d_%d", att.Video.OwnerID, att.Video.ID)
			if att.Video.Duration > 600 {
				//send url
				msg := tgbotapi.NewMessage(storeId, urlv+"\n"+appendix)
				msg.DisableWebPagePreview = false
				msg.DisableNotification = true
				res, err := wrbot.Send(msg)

				forward(users, res.MessageID, err, storeId)

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
						res, err := wrbot.Send(msg)

						forward(users, res.MessageID, err, storeId)

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
				res, err := wrbot.Send(msg)

				forward(users, res.MessageID, err, storeId)

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
					res, err := wrbot.Send(msg)

					forward(users, res.MessageID, err, storeId)

				}

			} else {
				var desc = ""
				desc = att.Link.Title + "\n" + att.Link.URL + "\n" + appendix

				msg := tgbotapi.NewMessage(storeId, desc)
				msg.DisableWebPagePreview = false
				msg.DisableNotification = true
				res, err := wrbot.Send(msg)
				forward(users, res.MessageID, err, storeId)

			}
		}
	}

}
