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
	"net/url"
	"strconv"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/publicsuffix"

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
	tok4vid    string
)

type ShortUrl struct {
	Kind    string `json:"kind"`
	ID      string `json:"id"`
	LongURL string `json:"longUrl"`
}

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
	tok4vid = wrtoken
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
	go forever()
	//parseRss()
	select {} // block forever
}

func forever() {
	for {
		fmt.Printf("%v+\n", time.Now())
		parseRss()
		time.Sleep(600 * time.Second)
	}
}

func parseRss() {
	domains := rssdomains()

	for hash, url := range domains {

		log.Println(hash)
		users := domUsers(hash)
		if len(users) == 0 {
			continue
		}
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
	log.Println("link", link, "userslen", len(users))

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

		key := GetMD5Hash(link) + "_" + GetMD5Hash(item.Link)
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
	//fmt.Printf("%+v\n", p)
	var vkcnt int64 = -1001067277325 //myakotka
	log.Println("pubpost", p.GUID)

	var links = extractLinks(p.Title + " " + p.Description + " " + p.Content)
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
		//gif hack
		if strings.HasSuffix(img, ".gif") {
			photo = img
			break
		}
		if len > max {
			max = len
			photo = img
			log.Println("photo", photo, "len", len)
		}
	}
	log.Println("phot:", photo)

	var title = strings.Trim(strings.Join(extractText(p.Title), " "), "\n ")

	var description = strings.Trim(strings.Join(extractText(p.Description), " "), "\n ")

	var caption = title
	var link = p.Link
	urls, err := url.Parse(link)
	var tag = ""
	if err == nil {
		link = urls.Host + urls.Path
		mainDomain, err := publicsuffix.EffectiveTLDPlusOne(urls.Host)
		if err == nil {
			if strings.LastIndex(mainDomain, ".") != -1 {
				tag = "#" + mainDomain[:strings.LastIndex(mainDomain, ".")] + " "
			} else {
				tag = "#" + mainDomain + " "
			}
		}
	}

	//get short url
	short := shortenUrl(link)
	if short != "" {
		link = short
	}
	appendix := fmt.Sprintf("\n%s🔗 %s", tag, link)

	//video
	//video
	//video := "https://i.imgur.com/JOwvswE.mp4"
	video := getVideo(links)
	if video != "" {
		ok := sendVideo(video, title+appendix, users, vkcnt)
		if ok {
			return
		}
	}

	if photo != "" {
		var maxlen = 190 - len(caption) - len(appendix)
		descr := description
		words := strings.Split(descr, " ")
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
		log.Println("caption", caption)

		b := httputils.HttpGet(photo, nil)
		if b != nil {

			bb := tgbotapi.FileBytes{Name: photo, Bytes: b}
			msg := tgbotapi.NewPhotoUpload(vkcnt, bb)
			msg.Caption = caption
			msg.DisableNotification = true
			res, err := wrbot.Send(msg)
			if err == nil {
				for user := range users {
					bot.Send(tgbotapi.NewForward(int64(user), vkcnt, res.MessageID))
				}
			} else {
				log.Println("\n\n\n\n\nError", err)
				fmt.Printf("%+v\n", msg)
			}

		}
	} else {
		if len(description) > 3000 {
			description = ""
		}
		msg := tgbotapi.NewMessage(vkcnt, "*"+title+"*\n"+description+appendix)
		msg.DisableWebPagePreview = true
		msg.DisableNotification = true
		msg.ParseMode = "Markdown"
		res, err := wrbot.Send(msg)
		if err == nil {
			for user := range users {
				bot.Send(tgbotapi.NewForward(int64(user), vkcnt, res.MessageID))
			}
		} else {
			log.Println("\n\n\n\n\nError", err)
			fmt.Printf("%+v\n", msg)
		}
	}
}

func sendVideo(video string, caption string, users map[int]bool, vkcnt int64) (ok bool) {
	captionQ := url.QueryEscape(caption)
	vidurl := params.TgApi + "/bot" + tok4vid + "/sendVideo?chat_id=@myakotkapub&disable_notification=true&caption=" + captionQ + "&video=" + video
	//https://api.telegram.org/bot332514590:AAFq1wVBFZDMbKPoVQ1Oq6U1SijLGuYsZC0/sendVideo?chat_id=@myakotkapub&video=http://i.imgur.com/JOwvswE.mp4
	//log.Println("vidurl", vidurl)
	b := httputils.HttpGet(vidurl, nil)
	if b != nil {
		type apires struct {
			Ok     bool `json:"ok"`
			Result struct {
				MessageID int `json:"message_id"`
				Chat      struct {
					ID       int64  `json:"id"`
					Title    string `json:"title"`
					Username string `json:"username"`
					Type     string `json:"type"`
				} `json:"chat"`
				Date     int `json:"date"`
				Document struct {
					FileName string `json:"file_name"`
					MimeType string `json:"mime_type"`
					Thumb    struct {
						FileID   string `json:"file_id"`
						FileSize int    `json:"file_size"`
						Width    int    `json:"width"`
						Height   int    `json:"height"`
					} `json:"thumb"`
					FileID   string `json:"file_id"`
					FileSize int    `json:"file_size"`
				} `json:"document"`
			} `json:"result"`
		}
		var result apires
		err := json.Unmarshal(b, &result)
		if err == nil {
			ok = true
			for user := range users {
				bot.Send(tgbotapi.NewForward(int64(user), vkcnt, result.Result.MessageID))
			}
		}
	}
	return
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

func extractText(s string) (texts []string) {
	var wasNl = false
	z := html.NewTokenizer(strings.NewReader(s))
	for {
		tt := z.Next()
		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			//log.Println("err tok")
			return
		case tt == html.TextToken:
			t := z.Token()
			text := strings.TrimSpace(t.Data)
			if text != "" {
				texts = append(texts, strings.TrimSpace(t.Data))
				wasNl = false
				//log.Println("text:", "'"+strings.TrimSpace(t.Data)+"'")
			}

		case tt == html.StartTagToken || tt == html.SelfClosingTagToken:
			t := z.Token()
			switch t.Data {
			case "br":
				if !wasNl {
					texts = append(texts, "\n")
				}
				wasNl = true
			}
		}
	}
}

func extractLinks(s string) (links []string) {
	checkLink := func(s string) {
		url, err := url.Parse(s)
		if err == nil {
			if url.Host == "" {
				return
			}
			if url.Host == "imgur.com" {
				//"https://imgur.com/bgtwwY2"
				paths := strings.Split(url.Path, "/")
				if len(paths) == 2 && strings.Contains(url.Path, ".") == false {
					s = s + ".png"
				}
			}
			links = append(links, s)
		}
	}
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
				checkLink(href)

				//log.Println(href)
			case "a":
				ok, href := getVal(t, "href")
				if !ok {
					continue
				}
				checkLink(href)
			default:
				//		log.Println("t.data", t.Data)
				continue
			}
		}
	}
}

// Dirty little hacks(
func getVideo(links []string) (video string) {

	for _, link := range links {
		url, err := url.Parse(link)
		if err == nil {
			if url.Host == "" {
				continue
			}
			log.Println("url.Host", url.Host)
			if url.Host == "imgur.com" || url.Host == "i.imgur.com" {
				if strings.HasSuffix(link, "gifv") {
					video = strings.Replace(link, ".gifv", ".mp4", -1)
					break
				}
			}
			if url.Host == "gfycat.com" {
				video = strings.Replace(link, "gfycat.com", "thumbs.gfycat.com", -1)
				video = video + "-mobile.mp4"
				break
			}
		}
	}
	return video
}

func getImgs(links []string) (imgs map[string]int) {
	imgs = make(map[string]int)
	for _, link := range links {
		resp, err := http.Head(link)
		if err != nil {
			continue
		}
		len, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
		// 10 - 500kb~?
		if len < 7000 || len > 2000000 {
			continue
		}
		isImg := strings.HasPrefix(resp.Header.Get("Content-Type"), "image")
		if isImg {
			imgs[link] = len
		}
	}
	return imgs
}

func shortenUrl(url string) (id string) {
	s := "{\"longUrl\": \"" + url + "\"}"
	resp, err := http.Post(params.ShortUrl, "application/json", strings.NewReader(s))
	if err == nil {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			//log.Println(string(body))
			var sh ShortUrl
			err := json.Unmarshal(body, &sh)
			if err == nil {
				id = sh.ID
			}
		}
	}
	return
}
