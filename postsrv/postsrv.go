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
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/publicsuffix"

	"strings"

	"strconv"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/mmcdole/gofeed"
	"github.com/orcaman/concurrent-map"
	"github.com/recoilme/tf/httputils"
	"github.com/recoilme/tf/params"
	"github.com/recoilme/tf/vkapi"
)

const (
	MaxInt = int(^uint(0) >> 1)
	MinInt = -MaxInt - 1
)

type tgMessage struct {
	msgtype  string
	userId   int64
	txt      string
	bytes    []byte
	fileName string
}

type shortUrl struct {
	Kind    string `json:"kind"`
	ID      string `json:"id"`
	LongURL string `json:"longUrl"`
}

type feedJob struct {
	link  string
	users map[int64]bool
}

var (
	bot *tgbotapi.BotAPI
	// Здесь будем хранить время последней отправки сообщения для каждого пользователя
	lastMessageTimes = cmap.New()
	// Здесь будем хранить время последней отправки сообщения в целом
	//lastMessageTime int64
	timer      = time.NewTicker(time.Second / 30)
	forbidden  = cmap.New()
	chQueueMsg = make(chan tgMessage, 100000)
)

func initBot() {
	var err error
	tlgrmtoken, err := ioutil.ReadFile(params.Telefeedfile)
	if err != nil {
		log.Fatal(err)
	}
	tgtoken := strings.Replace(string(tlgrmtoken), "\n", "", -1)
	bot, err = tgbotapi.NewBotAPI(tgtoken)
	if err != nil {
		log.Fatal(err)
	}
	bot.Debug = false
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	log.Println("postsrv")
	//var timer = time.NewTicker(time.Second / 30)
	initBot()
	//parse()
	//log.Println("end")

	go popQueueMsg()
	go forever()
	select {} // block forever
}

func forever() {
	for {
		fmt.Printf("%v+\n", time.Now())
		parse()
		time.Sleep(120 * time.Second)
	}
}

func popQueueMsg() {

	for {
		select {
		case msg := <-chQueueMsg:
			sendMsg(msg)
		}
	}
	//for msg := range chQueueMsg {
	//sendMsg(msg)
	//time.Sleep(100 * time.Millisecond)
	//}
}

func send(msgtype string, users map[int64]bool, txt string, bytes []byte, fileName string) {
	for user := range users {
		_, forbid := forbidden.Get(strconv.FormatInt(user, 10))
		if forbid {
			continue
		}
		msg := tgMessage{
			msgtype:  msgtype,
			userId:   user,
			txt:      txt,
			bytes:    bytes,
			fileName: fileName,
		}
		//go sendMsg(msg)
		chQueueMsg <- msg
	}
}

func checkErr(msg tgbotapi.Message, err error, userId int64) {
	if err != nil {
		s := err.Error()
		if strings.Contains(s, "orbidden") {
			forbidden.Set(strconv.FormatInt(userId, 10), true)
		} else {
			fmt.Printf("Error: msg:%+v userId:%d err:%s", nil, userId, s)
		}
	}
}

func sendMsg(msg tgMessage) {

	//for { //range timer.C {
	userId := msg.userId
	if userCanReceiveMessage(userId) {
		//log.Println(msg.msgtype)

		switch msg.msgtype {
		case "photo":
			m := tgbotapi.NewPhotoUpload(userId, tgbotapi.FileBytes{Name: msg.fileName, Bytes: msg.bytes})
			m.DisableNotification = true
			m.Caption = msg.txt
			res, err := bot.Send(m)
			checkErr(res, err, userId)
		case "video":
			m := tgbotapi.NewVideoUpload(userId, tgbotapi.FileBytes{Name: msg.fileName, Bytes: msg.bytes})
			m.DisableNotification = true
			m.Caption = msg.txt
			res, err := bot.Send(m)
			checkErr(res, err, userId)
		case "doc":
			m := tgbotapi.NewDocumentUpload(userId, tgbotapi.FileBytes{Name: msg.fileName, Bytes: msg.bytes})
			m.DisableNotification = true
			m.Caption = msg.txt
			res, err := bot.Send(m)
			checkErr(res, err, userId)
		case "link":
			m := tgbotapi.NewMessage(userId, msg.txt)
			m.DisableNotification = true
			m.DisableWebPagePreview = false
			m.ParseMode = "Markdown"
			res, err := bot.Send(m)
			checkErr(res, err, userId)
		default:
			//txt
			m := tgbotapi.NewMessage(userId, msg.txt)
			m.DisableNotification = true
			m.DisableWebPagePreview = true
			res, err := bot.Send(m)
			checkErr(res, err, userId)
		}

		lastMessageTimes.Set(strconv.FormatInt(userId, 10), time.Now().UnixNano())
		lastMessageTimes.Set("0", time.Now().UnixNano())
		fmt.Printf("%s Ok Userid:%d\n", time.Now().Format("15:04:05"), userId)
		//break
	}
	//}
}

func parse() {

	feeds := getFeedNames()

	//arr of job
	jobs := make([]feedJob, 0, 0)

	for _, hash := range feeds {
		//log.Println("getfeed", url, hash)
		b := httputils.HttpGet(params.Feeds+hash, nil)
		if b != nil {
			url := string(b)
			feedUsers := feedUsers(hash)
			if len(feedUsers) == 0 {
				continue
			}
			//log.Println("getfeed", url)
			//getFeedPosts(url, feedUsers)
			feedJob := feedJob{
				link:  url,
				users: feedUsers,
			}
			jobs = append(jobs, feedJob)
		}
	}

	//channels
	chFeedJobs := make(chan feedJob, len(jobs))
	//chFeedResults := make(chan bool, len(jobs))
	//workers
	for w := 1; w <= 3; w++ {
		go workerFeed(w, chFeedJobs)
	}

	for _, job := range jobs {
		chFeedJobs <- job
	}
	//all jobs send
	close(chFeedJobs)

	log.Println("jobs send")
	//for r := 0; r < len(jobs); r++ {
	//	res := <-feedResults
	//	fmt.Println("finished with res:", res)
	//}
	//close(feedResults)

	time.Sleep(time.Duration(10) * time.Second)
	log.Println("feed done")

	publics := getPubNames()
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
				log.Println("getpub", public.ScreenName)
				//go MakeRequestDeferred(i, "22", nil, "", nil)
				getPubPosts(public, pubusers)
			}
		}
	}
	time.Sleep(30 * time.Second)

}

func workerFeed(id int, jobs <-chan feedJob) {

	for job := range jobs {

		feed := job
		link := feed.link
		users := feed.users
		getFeedPosts(link, users)

		//rate limits wor workers
		time.Sleep(time.Duration(1) * time.Second)
		//results <- result
	}

}

func pubUsers(domain vkapi.Group) (users map[int64]bool) {
	mask := params.Subs + "%d"
	url := fmt.Sprintf(mask, domain.Gid)
	//log.Println(url)
	b := httputils.HttpGet(url, nil)
	if b != nil {
		json.Unmarshal(b, &users)
	}
	return users
}

func feedUsers(hash string) (users map[int64]bool) {
	mask := params.FeedSubs + "%s"
	url := fmt.Sprintf(mask, hash)
	//log.Println(url)
	b := httputils.HttpGet(url, nil)
	if b != nil {
		json.Unmarshal(b, &users)
	}
	return users
}

func getPubNames() (domainNames []string) {
	url := params.BaseUri + "pubNames/First?cnt=1000000&order=asc"
	//log.Println("vkdomains", url)
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

func getFeedNames() (feedNames []string) {

	url := params.BaseUri + "feeds/First?cnt=1000000&order=asc"
	//log.Println("rssdomains", url)
	resp, err := http.Post(url, "application/json", nil)
	if err == nil {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			err := json.Unmarshal(body, &feedNames)
			if err == nil {
				return
			} else {
				fmt.Println(err)
			}
		}
	}
	return
}

func getPubPosts(domain vkapi.Group, users map[int64]bool) {
	var lastPost = lastPostIdGet(domain)
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
		pubpost(domain, post, users)
	}
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

// Проверка может ли уже пользователь получить следующее сообщение
func userCanReceiveMessage(userId int64) (result bool) {
	for {
		t, ok := lastMessageTimes.Get(strconv.FormatInt(userId, 10))

		result = !ok || t.(int64)+int64(time.Second) <= time.Now().UnixNano()
		if result == true {
			//if we may send to this user check all limit
			t, ok := lastMessageTimes.Get("0")
			result = !ok || t.(int64)+(int64(time.Second/30)) <= time.Now().UnixNano()
		}
		if result {
			break
		} else {
			time.Sleep(20 * time.Millisecond)
		}
	}
	return
}

func pubpost(domain vkapi.Group, p vkapi.Post, users map[int64]bool) {
	if len(users) == 0 {
		return
	}
	fmt.Printf("%s domain:%s Post:%d\n", time.Now().Format("15:04:05"), domain.ScreenName, p.Id)
	var t = strings.Replace(p.Text, "&lt;br&gt;", "\n", -1)
	if t != "" {
		t = t + "\n"
	}
	link := fmt.Sprintf("vk.com/wall%d_%d", domain.Gid*(-1), p.Id)
	tag := strings.Replace(domain.ScreenName, ".", "", -1)

	appendix := fmt.Sprintf("#%s 🔗 %s", tag, link)
	if len(p.Attachments) == 0 || len([]rune(t)) > 200 {
		send("txt", users, t+appendix, nil, "")
		t = ""
	}
	for i := range p.Attachments {
		att := p.Attachments[i]
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
			b := httputils.HttpGet(photo, nil)
			if b != nil {
				var photoCaption string
				if i == 0 {
					photoCaption = t + appendix
				} else {
					photoCaption = appendix
				}
				send("photo", users, photoCaption, b, photo)
			}
		case "video":
			urlv := fmt.Sprintf("https://vk.com/video%d_%d", att.Video.OwnerID, att.Video.ID)
			if att.Video.Duration > 600 {
				//send url
				send("txt", users, urlv+"\n"+appendix, nil, "")
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
						if vidb != nil {
							send("video", users, appendix, vidb, s)
						}

					}
				}
			}
		case "doc":
			b := httputils.HttpGet(att.Doc.URL, nil)
			if b != nil {
				send("doc", users, appendix, b, "tmp."+att.Doc.Ext)
			}
		case "link":
			if att.Link.Photo.Photo604 != "" && att.Link.Photo.Width > 400 && att.Link.Photo.Height > 400 {
				//link with photo
				b := httputils.HttpGet(att.Link.Photo.Photo604, nil)
				if b != nil {
					msgCaption := att.Link.Title + "\n" + att.Link.Description + "\n" + att.Link.URL + "\n" + appendix
					send("photo", users, msgCaption, b, "")
				}

			} else {
				var desc = ""
				desc = att.Link.Title + "\n" + att.Link.URL + "\n" + appendix
				send("link", users, desc, nil, "")
			}
		}
	}

}

func getFeedPosts(link string, users map[int64]bool) bool {

	var defHeaders = make(map[string]string)
	defHeaders["User-Agent"] = "script::recoilme:v1"
	defHeaders["Authorization"] = "Client-ID 4191ffe3736cfcb"
	b := httputils.HttpGet(link, defHeaders)
	if b == nil {
		return false
	}
	fp := gofeed.NewParser()
	feed, err := fp.Parse(bytes.NewReader(b))
	if err != nil {
		return false
	}

	var last = len(feed.Items) - 1
	if last > 10 {
		last = 10
	}
	for i := range feed.Items {
		if i > last {
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
		pubFeed(link, item, users)

	}
	return true
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func pubFeed(domain string, p *gofeed.Item, users map[int64]bool) {

	fmt.Printf("%s Feed: %s\n", time.Now().Format("15:04:05"), p.Link)
	//var vkcnt int64 = -1001067277325 //myakotka
	//log.Println("pubpost", p.GUID)

	var links = extractLinks(p.Title + " " + p.Description + " " + p.Content)
	//log.Println("lin1", links)
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
			//log.Println("photo", photo, "len", len)
		}
	}
	//log.Println("phot:", photo)
	var title = strings.Trim(strings.Join(extractText(p.Title), " "), "\n ")

	var description = strings.Trim(strings.Join(extractText(p.Description), " "), "\n ")

	var caption = title
	var link = p.Link
	urls, err := url.Parse(link)
	var tag = ""
	if err == nil {
		mainDomain, err := publicsuffix.EffectiveTLDPlusOne(urls.Host)
		if err == nil {
			if strings.LastIndex(mainDomain, ".") != -1 {
				tag = "#" + mainDomain[:strings.LastIndex(mainDomain, ".")] + " "
			} else {
				tag = "#" + mainDomain + " "
			}
		}
	}
	tag = strings.Replace(tag, "-", "", -1)
	tag = strings.Replace(tag, ".", "", -1)

	//get short url
	short := shortenUrl(link)
	if short != "" {
		link = short
	}
	appendix := fmt.Sprintf("\n%s🔗 %s", tag, link)

	//video
	video := getVideo(links)
	if video != "" {
		//post video
		vidb := httputils.HttpGet(video, nil)
		if vidb != nil {
			send("video", users, appendix, vidb, video)
		} else {
			return
		}

	}

	if photo != "" {
		var maxlen = 190 - len([]rune(caption)) - len([]rune(appendix))
		descr := description
		caption = caption + trimTo(descr, maxlen)

		caption = caption + appendix
		log.Println("caption", caption)

		b := httputils.HttpGet(photo, nil)
		if b != nil {

			if strings.HasSuffix(photo, ".gif") {
				send("doc", users, caption, b, photo)
			} else {
				send("photo", users, caption, b, photo)
			}
		}
	} else {
		description = trimTo(description, 4000-len([]rune(title))-len([]rune(appendix))-10)
		msgtxt := "*" + title + "*\n" + description + appendix

		//if len([]rune(msgtxt)) < 250 {
		send("link", users, msgtxt, nil, "")
		//} else {
		//send("txt", users, msgtxt, nil, "")
		//}
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
			//log.Println("err tok")
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

func getImgs(links []string) (imgs map[string]int) {
	imgs = make(map[string]int)
	var maxlen = 2000000
	for _, link := range links {
		resp, err := http.Head(link)
		if err != nil {
			continue
		}
		len, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
		// 10 - 500kb~?
		//gif hack
		if strings.HasSuffix(link, ".gif") {
			maxlen = 20000000
		} else {
			maxlen = 2000000
		}
		if len < 7000 || len > maxlen {
			continue
		}
		isImg := strings.HasPrefix(resp.Header.Get("Content-Type"), "image")
		if isImg {
			imgs[link] = len
		}
	}
	return imgs
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
func shortenUrl(url string) (id string) {
	s := "{\"longUrl\": \"" + url + "\"}"
	resp, err := http.Post(params.ShortUrl, "application/json", strings.NewReader(s))
	if err == nil {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			//log.Println(string(body))
			var sh shortUrl
			err := json.Unmarshal(body, &sh)
			if err == nil {
				id = sh.ID
			}
		}
	}
	return
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

func trimTo(s string, lenstart int) (result string) {
	var maxlen = lenstart
	words := strings.Split(s, " ")
	for i, word := range words {
		if i == 0 {
			result = result + "\n"
		}
		if len([]rune(word)) < maxlen {
			maxlen = maxlen - len([]rune(word)) - 1
			result = result + word + " "
		} else {
			result = result + ".."
			break
		}
	}
	return
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
