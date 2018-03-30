package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/crc64"
	"image"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/publicsuffix"

	"strings"

	"strconv"

	"github.com/disintegration/imaging"
	"github.com/go-redis/redis"
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

var (
	bot *tgbotapi.BotAPI
	//botYa botan.Botan
	// Здесь будем хранить время последней отправки сообщения для каждого пользователя
	lastMessageTimes = cmap.New()
	forbidden        = cmap.New()
	imgs             = cmap.New()
	videos           = cmap.New()
	red              *redis.Client
)

func initBot() {
	var err error
	tlgrmtoken, err := ioutil.ReadFile(params.Telefeedfile)
	if err != nil {
		log.Fatal(err)
	}
	/*
		fathertoken, err := ioutil.ReadFile(params.ChannelsFatherfile)
		if err != nil {
			log.Fatal(err)
		}*/
	tgtoken := strings.Replace(string(tlgrmtoken), "\n", "", -1)
	//ftoken := strings.Replace(string(fathertoken), "\n", "", -1)
	bot, err = tgbotapi.NewBotAPI(tgtoken)
	if err != nil {
		log.Fatal(err)
	}
	/*
		father, err = tgbotapi.NewBotAPI(ftoken)
		if err != nil {
			log.Fatal(err)
		}*/
	bot.Debug = false
	if err != nil {
		log.Fatal(err)
	}
	red = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := red.Ping().Result()
	fmt.Println(pong, err)
	if err != nil {
		panic(err)
	}
}

func main() {
	log.Println("postsrv")
	initBot()
	//botYa = botan.New(params.YaToken)
	argsWithProg := os.Args
	var param = ""
	var usr = ""
	if len(argsWithProg) > 1 {
		param = strings.TrimSpace(os.Args[1])
	}
	if len(argsWithProg) > 2 {
		usr = strings.TrimSpace(os.Args[2])
	}
	fmt.Printf("params:" + param + " " + usr + "!\n")
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		onclose()
		os.Exit(1)
	}()
	go forever(param, usr)
	select {} // block forever

}

func onclose() {
	fmt.Println("OnClose")
	/*
		for _, key := range vkpost.Keys() {
			t, ok := vkpost.Get(key)
			if ok {
				gid, _ := strconv.Atoi(key)
				lastPostIdSet(gid, t.(int), false)
				fmt.Println("OnClose:", gid, t.(int))
			}
		}*/
	/*
		for _, keyFeed := range feedpost.Keys() {
			_, ok := feedpost.Get(keyFeed)
			if ok {
				httputils.HttpPut(params.Links+keyFeed, nil, []byte(" "))
				fmt.Println("OnClose:", keyFeed)
			}
		}*/

	fmt.Println("OnClose end")
}

func forever(param string, usr string) {
	for {
		fmt.Printf("%v+\n", time.Now())
		switch param {
		case "publics":
			parsePublics(usr)
		case "feeds":
			parseFeeds()
			//default:
			//parseFeeds()
			//parsePublics()
		}
		fmt.Printf("Sleep\n")

		for _, key := range forbidden.Keys() {
			forbidden.Remove(key)
		}
		time.Sleep(400 * time.Second)
	}
}

func send(msgtype string, users map[int64]bool, txt string, bytes []byte, fileName string) []tgMessage {
	msgs := make([]tgMessage, 0, 0)
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
		//chQueueMsg <- msg
		msgs = append(msgs, msg)
	}
	return msgs
}

func checkErr(m tgMessage, msg tgbotapi.Message, err error, userId int64) {
	if err != nil {
		s := err.Error()
		if strings.Contains(s, "orbidden") {
			fmt.Printf("orbidden %d\n", userId)
			forbidden.Set(strconv.FormatInt(userId, 10), true)
		} else {
			if strings.Contains(s, "Too Many") {
				fmt.Printf("Error: msg:%+v userId:%d err:%s\n", m.msgtype, userId, s)
				forbidden.Set(strconv.FormatInt(userId, 10), true)
				time.Sleep(5 * time.Second)
			} else {
				fmt.Printf("Error: msgty:%s fn:%s txt:%s userId:%d err:%s\n", m.msgtype, m.fileName, m.txt, userId, s)
			}
		}

	}
}

func sendMsg(msg tgMessage) {

	userId := msg.userId
	_, forbid := forbidden.Get(strconv.FormatInt(userId, 10))
	if forbid {
		return
	}
	if userCanReceiveMessage(userId) {

		//log.Println(msg.msgtype)
		txt := msg.txt
		msgtype := msg.msgtype
		fileName := msg.fileName
		msgBytes := msg.bytes
		var res tgbotapi.Message
		var err error
		switch msgtype {
		case "photo":
			crc64Int := crc64.Checksum(msgBytes, crc64.MakeTable(0xC96C5795D7870F42))
			crcHash := strconv.FormatUint(crc64Int, 16)
			imgHash, ok := imgs.Get(crcHash)
			if ok {
				//send already uploaded photo
				//fmt.Printf("Send PhotoUpload: res:%s \n", imgHash.(string))
				newmsg := tgbotapi.NewPhotoShare(userId, imgHash.(string))
				newmsg.DisableNotification = true
				newmsg.Caption = trimTo(txt, 200)
				bot.Send(newmsg)
			} else {
				//new Photo
				m := tgbotapi.NewPhotoUpload(userId, tgbotapi.FileBytes{Name: fileName, Bytes: msgBytes})
				m.DisableNotification = true
				m.Caption = trimTo(txt, 200)
				res, err = bot.Send(m)
				//fmt.Printf("NewPhotoUpload: res:%+v \n", res.Photo)
				if res.Photo != nil {
					var files string
					for _, photos := range *res.Photo {
						files = photos.FileID
					}
					//fmt.Printf("NewPhotoUpload: res:%s\n", files)

					//fmt.Printf("crcHash:%s\n", crcHash)
					imgs.Set(crcHash, files)
				}
			}
			//checkErr(msg, res, err, userId)
		case "video":

			crc64Int := crc64.Checksum(msgBytes, crc64.MakeTable(0xC96C5795D7870F42))
			crcHash := strconv.FormatUint(crc64Int, 16)
			videoHash, ok := videos.Get(crcHash)
			if ok {
				//send already uploaded video
				//fmt.Printf("Send videoUpload: res:%s \n", videoHash.(string))
				newmsg := tgbotapi.NewDocumentShare(userId, videoHash.(string))
				newmsg.DisableNotification = true
				newmsg.Caption = trimTo(txt, 200)
				res, err = bot.Send(newmsg)
				//fmt.Printf("err: %s\n", err.Error)
			} else {
				//new video
				m := tgbotapi.NewVideoUpload(userId, tgbotapi.FileBytes{Name: fileName, Bytes: msgBytes})
				m.DisableNotification = true
				m.Caption = trimTo(txt, 200)
				res, err = bot.Send(m)
				//fmt.Printf("NewvideoUpload: res:%+v \n", res)
				if res.Document != nil {
					files := *res.Document
					fileId := files.FileID
					//fmt.Printf("NewDocumentUpload: res:%s\n", fileId)
					//fmt.Printf("crcHash:%s\n", crcHash)
					videos.Set(crcHash, fileId)
				}
			}
			//fmt.Printf("video: res:%+v \n", res.Video)
			//checkErr(msg, res, err, userId)
		case "doc":
			m := tgbotapi.NewDocumentUpload(userId, tgbotapi.FileBytes{Name: fileName, Bytes: msgBytes})
			m.DisableNotification = true
			m.Caption = trimTo(txt, 200)
			res, err = bot.Send(m)
			//checkErr(msg, res, err, userId)
		case "link":
			m := tgbotapi.NewMessage(userId, trimTo(txt, 3500))
			m.DisableNotification = true
			m.DisableWebPagePreview = false
			//m.ParseMode = tgbotapi.ModeHTML // "Markdown"
			res, err = bot.Send(m)
			//checkErr(msg, res, err, userId)
		default:
			//txt
			m := tgbotapi.NewMessage(userId, trimTo(txt, 3500))
			m.DisableNotification = true
			m.DisableWebPagePreview = true
			res, err = bot.Send(m)
			//checkErr(msg, res, err, userId)
		}
		checkErr(msg, res, err, userId)
		lastMessageTimes.Set(strconv.FormatInt(userId, 10), time.Now().UnixNano())
		lastMessageTimes.Set("0", time.Now().UnixNano())
		//fmt.Printf("%s Ok Userid:%d\n", time.Now().Format("15:04:05"), userId)
		//break
		//time.Sleep(33 * time.Millisecond)
	}
}

func parsePublics(order string) {
	var publics []string
	/*
		if usr != "" {
			suburl := params.UserSubs + usr
			bodysub := httputils.HttpGet(suburl, nil)
			subs := make(map[string]bool)
			json.Unmarshal(bodysub, &subs)
			for s := range subs {
				if strings.HasPrefix(s, params.PubNames) {
					publics = append(publics, strings.Replace(s, params.PubNames, "", -1))
				}
			}
		} else {
			calcPubTop()
			publics = getPubNames()
		}*/
	calcPubTop()
	if order == "desc" {
		publics = getPubNames("desc")
	} else {
		publics = getPubNames("asc")
	}
	//fmt.Println(publics)
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
				lastMessageTimes.Set("vk", time.Now().UnixNano())
				msgs := getPubPosts(public, pubusers)

				if len(msgs) > 0 {
					for _, msg := range msgs {
						m := msg
						sendMsg(m)
					}
					//chMsgs := make(chan tgMessage, len(msgs))
					//workers
					//workerMsg(chMsgs)

					//for _, msg := range msgs {
					//chMsgs <- msg
					//}
					//all jobs send
					//close(chMsgs)

				}

				lastVk, ok := lastMessageTimes.Get("vk")
				if ok {
					delay := (lastVk.(int64) + int64(time.Millisecond*334)) - time.Now().UnixNano()
					if delay > 0 {
						time.Sleep(time.Duration(delay))
					}
				}

				//log.Println("msgs pub send")
			}
		}
	}
	time.Sleep(600 * time.Second)
}

func parseFeeds() {

	feeds := getFeedNames()

	//arr of job
	//jobs := make([]feedJob, 0, 0)
	//var wg sync.WaitGroup
	//var activejobs = 0
	for _, hash := range feeds {
		//log.Println("getfeed", url, hash)
		b := httputils.HttpGet(params.Feeds+hash, nil)
		if b != nil {
			url := string(b)
			feedUsers := feedUsers(hash)
			if len(feedUsers) == 0 {
				continue
			}
			//wg.Add(1)
			//activejobs++
			getFeedPosts(url, feedUsers, nil)
			//if activejobs >= 3 {
			//activejobs = 0
			//wg.Wait()
			//}
			//log.Println("getfeed", url)
			//getFeedPosts(url, feedUsers)
			//feedJob := feedJob{
			//		link:  url,
			//	users: feedUsers,
			//}
			//jobs = append(jobs, feedJob)
		}
	}
	//wg.Wait()

	//channels
	//chFeedJobs := make(chan feedJob, len(jobs))
	//chFeedResults := make(chan bool, len(jobs))
	//workers
	//for w := 1; w <= 1; w++ {
	//go workerFeed(w, chFeedJobs)
	//}

	//for _, job := range jobs {
	//chFeedJobs <- job
	//}
	//all jobs send
	//close(chFeedJobs)
	log.Println("jobs send")

	time.Sleep(time.Duration(180) * time.Second)
	fmt.Printf("feed done\n")
}

func workerMsg(msgs <-chan tgMessage) {
	for msg := range msgs {
		m := msg
		sendMsg(m)
	}
}

/*
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
*/
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

func getPubNames(order string) (domainNames []string) {
	var url = params.BaseUri + "pubNames/First?cnt=1000000&order=asc"
	if order == "desc" {
		url = params.BaseUri + "pubNames/Last?cnt=1000000&order=desc"
	}
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
				fmt.Println("getFeedNames", err)
			}
		}
	}
	return
}

func getPubPosts(domain vkapi.Group, users map[int64]bool) []tgMessage {
	msgs := make([]tgMessage, 0, 0)
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered in f %v\n", r)
		}
	}()
	//todo uncom

	var lastPost = lastPostIdGet(domain)
	posts := vkapi.WallGet(domain.Gid * (-1))

	last := len(posts) - 1
	if last > 5 {
		last = 5
	}

	var maxPostId = lastPost
	for i := range posts {
		if i > last {
			break
		}
		post := posts[last-i]

		if post.Id > maxPostId {
			maxPostId = post.Id
		}
	}
	if maxPostId == lastPost {
		return msgs
	}
	maxPostId = lastPostIdSet(domain.Gid, maxPostId, true)
	for i := range posts {
		if i > last {
			break
		}
		post := posts[last-i]
		if post.Id <= lastPost {
			continue
		}

		//ads
		if post.MarkedAsAds == 1 {
			continue
		}
		if len(post.Attachments) == 0 && post.Text == "" {
			// no text no attachments
			continue
		}
		secFromPub := int(time.Now().Unix() - int64(post.Date))
		if secFromPub > 86400 {
			continue
		}
		viralityVal := virality(post.Date, post.Views.Count, post.Likes.Count, post.Reposts.Count)
		if viralityVal >= 100 {
			//fmt.Printf("good! https://vk.com/wall-%d_%d\n", post.OwnerID*(-1), post.Id)
			//url := fmt.Sprintf(params.Bests+"%s_%d_%d", time.Now().Format("20060102150405"), post.OwnerID*(-1), post.Id)
			//b, _ := json.Marshal(post)
			//httputils.HttpPut(url, nil, b)
			if viralityVal >= 150 {
				users[-1001099791579] = true
			}
		} else {
			delete(users, -1001099791579)
		}
		postmsgs := pubpost(domain, post, users)
		for _, postmsg := range postmsgs {
			msgs = append(msgs, postmsg)
		}
	}
	return msgs
}

func postVirality(post vkapi.Post) int {
	//post.Date, post.Views.Count, post.Likes.Count, post.Reposts.Count
	//return 101
	view := post.Views.Count
	like := post.Likes.Count
	repost := post.Reposts.Count
	secFromPub := int(time.Now().Unix() - int64(post.Date))
	if view < 1000 || like < 100 || repost < 50 || secFromPub < 100 {
		return 0
	}
	//fmt.Printf("secFromPub:%d\n", secFromPub)
	virality := (like * 1000 / view) + (repost*10000/view)*((view/2)/secFromPub)/10
	//fmt.Printf("vir:%d\n", virality)
	return virality
}

func virality(date int, view int, like int, repost int) int {
	//return 101
	secFromPub := int(time.Now().Unix() - int64(date))
	if view < 1000 || like < 100 || repost < 50 || secFromPub < 100 {
		return 0
	}
	if like/secFromPub < 1 {
		return 0
	}
	//fmt.Printf("secFromPub:%d\n", secFromPub)
	virality := (like * 1000 / view) + (repost*10000/view)*((view/2)/secFromPub)/10
	//fmt.Printf("vir:%d\n", virality)
	return virality
}

func lastPostIdSet(domainID int, lastPostId int, memory bool) int {

	_, err := red.Set(strconv.Itoa(domainID), strconv.Itoa(lastPostId), (3600*24*7)*time.Second).Result()
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	return lastPostId

}

func lastPostIdGet(domain vkapi.Group) int {
	postId := MinInt
	strval, err := red.Get(strconv.Itoa(domain.Gid)).Result()
	if err == redis.Nil {
		mask := params.LastPost + "%d"
		url := fmt.Sprintf(mask, domain.Gid)
		b := httputils.HttpGet(url, nil)
		if b != nil {
			json.Unmarshal(b, &postId)
		}
		return postId
	} else if err != nil {
		fmt.Println(err)
	} else {
		val, err := strconv.Atoi(strval)
		if err != nil {
			return MinInt
		}
		return val
	}
	return postId
}

func userCanReceiveMessage(userId int64) bool {
	//by default - second ago
	var lastUsrMsg = time.Now().UnixNano() - int64(time.Second)
	var lastAllMsg = lastUsrMsg
	for {
		t, ok := lastMessageTimes.Get(strconv.FormatInt(userId, 10))
		if ok {
			//update if set
			lastUsrMsg = t.(int64)
		}
		//if more then sec ago
		if lastUsrMsg+int64(time.Second) <= time.Now().UnixNano() {
			//fmt.Println("ressult", ok)
			//fmt.Printf("lastUsrMsg:%d  fut:%d now:%d \n", lastUsrMsg, lastUsrMsg+int64(2*time.Second), time.Now().UnixNano())
			t, ok := lastMessageTimes.Get("0")
			if ok {
				lastAllMsg = t.(int64)
			}
			//if more then 1/20 sec ago
			result := lastAllMsg+(int64(time.Second/20)) <= time.Now().UnixNano()
			if result {
				break
			}
		}
		time.Sleep(time.Second / 20)
	}
	return true
}

func histogramHash(b []byte) string {

	img, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		return ""
	}
	img = imaging.Resize(img, 4, 4, imaging.Box)
	//img = grayscale.Convert(img, grayscale.ToGrayAverage)
	bounds := img.Bounds()
	var histogram [16][4]int

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			// A color's RGBA method returns values in the range [0, 65535].
			// Shifting by 12 reduces this to the range [0, 15].
			histogram[r>>12][0]++
			histogram[g>>12][1]++
			histogram[b>>12][2]++
			histogram[a>>12][3]++
		}
	}

	var buffer bytes.Buffer
	for _, x := range histogram {
		buffer.WriteString(strconv.FormatInt(int64(x[0]+x[1]+x[2]+x[3]), 10))
	}
	hash := md5.Sum(buffer.Bytes())
	strhash := hex.EncodeToString(hash[:])
	return strhash
}

func pubpost(domain vkapi.Group, p vkapi.Post, users map[int64]bool) []tgMessage {
	//panic("Panic")

	msgs := make([]tgMessage, 0, 0)
	if len(users) == 0 {
		return msgs
	}
	fmt.Printf("%s domain:%s Post:%d\n", time.Now().Format("15:04:05"), domain.ScreenName, p.Id)
	var t = strings.Replace(p.Text, "&lt;br&gt;", "\n", -1)
	t = strings.Replace(t, "<br>", "\n", -1)
	if t != "" {
		t = t + "\n"
	}
	link := fmt.Sprintf("vk.com/wall%d_%d", domain.Gid*(-1), p.Id)
	tag := strings.Replace(domain.ScreenName, ".", "", -1)
	var virality = postVirality(p) / 25
	if virality > 9 {
		virality = 9
	}

	var likes string
	switch virality {
	case 1:
		likes = "❤️"
	case 2:
		likes = "❤️❤️"
	case 3:
		likes = "❤️❤️❤️"
	case 4:
		likes = "🔥"
	case 5:
		likes = "🔥🔥"
	case 6:
		likes = "🔥🔥🔥"
	case 7:
		likes = "💣"
	case 8:
		likes = "💣💣"
	case 9:
		likes = "💣💣💣"
	}
	appendix := strings.TrimSpace(fmt.Sprintf("#%s 🔗 %s %s", tag, link, likes))
	if len(p.Attachments) == 0 || len([]rune(t)) > 200 {
		msgtxt := trimTo(t, 3900-len([]rune(appendix))-10) + appendix
		for _, m := range send("txt", users, msgtxt, nil, "") {
			msgs = append(msgs, m)
		}
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
				for _, m := range send("photo", users, photoCaption, b, photo) {
					msgs = append(msgs, m)
				}
			}
		case "video":
			urlv := fmt.Sprintf("https://vk.com/video%d_%d", att.Video.OwnerID, att.Video.ID)
			if att.Video.Duration > 600 {
				//send url
				for _, m := range send("txt", users, urlv+"\n"+appendix, nil, "") {
					msgs = append(msgs, m)
				}
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
					if len(cnt) > poshttp && len(cnt) > pos360+8 && (pos360+8) > poshttp {
						s := strings.Replace(cnt[poshttp:pos360+8], "\\/", "/", -1)
						if s != "" {
							//post video
							vidb := httputils.HttpGet(s, nil)
							if vidb != nil {
								for _, m := range send("video", users, trimTo(t, 200-len([]rune(appendix))-10)+"\n"+appendix, vidb, "1.mp4") {
									msgs = append(msgs, m)
								}
							}
						}
					}

				} else {
					log.Println("video", cnt, len(cnt), ":", poshttp, ":", pos360+8)
				}
			}
		case "doc":
			b := httputils.HttpGet(att.Doc.URL, nil)
			if b != nil {
				for _, m := range send("doc", users, trimTo(t, 200-len([]rune(appendix))-10)+"\n"+appendix, b, "tmp."+att.Doc.Ext) {
					msgs = append(msgs, m)
				}
			}
		case "link":
			var hasPhoto = false
			if att.Link.Photo.Photo604 != "" && att.Link.Photo.Width > 400 && att.Link.Photo.Height > 400 && len([]rune(p.Text)) < 200 {
				//link with photo
				b := httputils.HttpGet(att.Link.Photo.Photo604, nil)
				if b != nil {
					hasPhoto = true
					trimlen := 190 - len([]rune(appendix))
					msgCaption := trimTo(att.Link.Title+"\n"+att.Link.Description, trimlen) + "\n" + appendix

					log.Println("caption", msgCaption)
					//fmt.Printf("Photo %s\n", att.Link.Photo.Photo604)
					for _, m := range send("photo", users, msgCaption, b, att.Link.Photo.Photo604) {
						msgs = append(msgs, m)
					}
				}

			}
			//not send as text or photo
			if !hasPhoto && len([]rune(p.Text)) < 200 {
				var desc = ""
				desc = att.Link.Title + "\n" + att.Link.URL + "\n" + appendix
				for _, m := range send("link", users, desc, nil, "") {
					msgs = append(msgs, m)
				}
			}

		}
	}
	return msgs
}

func pubFeed(domain string, p *gofeed.Item, users map[int64]bool, feedLink string) []tgMessage {
	msgs := make([]tgMessage, 0, 0)
	//http://www.farsroid.com/sleep-as-android-full/
	if strings.Contains(p.Link, "farsroid") {
		return msgs
	}
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered in f %v\n", r)
		}
	}()
	fmt.Printf("%s Feed: %s\n", time.Now().Format("15:04:05"), p.Link)
	//var vkcnt int64 = -1001067277325 //myakotka
	//log.Println("pubpost", p.GUID)

	//fmt.Printf("1. %s Feed: %s\n", time.Now().Format("15:04:05"), p.Link)
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
	//fmt.Printf("1.1 %s Feed: %s\n", time.Now().Format("15:04:05"), p.Link)
	imgs := getImgs(links)
	//fmt.Printf("2. %s Feed: %s\n", time.Now().Format("15:04:05"), p.Link)
	var max = 0
	var photo = ""
	for img, len := range imgs {
		if len > max {
			max = len
			photo = img
		}
	}
	var title = strings.Trim(strings.Join(extractText(p.Title), " "), "\n ")

	var description = strings.Trim(strings.Join(extractText(p.Description), " "), "\n ")

	var caption = title
	var link = p.Link
	if strings.HasPrefix(link, "/") {
		maindom, err := url.Parse(feedLink)
		if err == nil {
			link = maindom.Scheme + "://" + maindom.Host + link
		}
	}
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

	//fmt.Printf("3. %s Feed: %s\n", time.Now().Format("15:04:05"), p.Link)
	//get short url
	short := "" //shortenUrl(link)
	if short != "" {
		link = short
	}
	appendix := fmt.Sprintf("\n%s🔗 %s", tag, link)

	//video
	video := getVideo(links)
	//fmt.Printf("4. %s Feed: %s\n", time.Now().Format("15:04:05"), p.Link)
	if video != "" {
		//post video
		caption = trimTo(caption+"\n"+description, 190-len([]rune(appendix))) + "\n" + appendix
		vidb := httputils.HttpGet(video, nil)
		if vidb != nil {
			crc64Int := crc64.Checksum(vidb, crc64.MakeTable(0xC96C5795D7870F42))
			crcHash := strconv.FormatUint(crc64Int, 16)
			old, err := red.Get(params.Video + crcHash).Result()
			if err != redis.Nil {
				if !strings.Contains(old, video) && !strings.Contains(old, "\n") {

					m := tgbotapi.NewVideoUpload(-1001114648696, tgbotapi.FileBytes{Name: "video.mp4", Bytes: vidb})
					m.DisableNotification = true
					m.Caption = trimTo(old+"\n"+video, 200)
					bot.Send(m)
					red.Set(params.Video+crcHash, old+"\n"+video, (3600*24*10)*time.Second).Result()
				}
			} else {
				red.Set(params.Video+crcHash, video, (3600*24*10)*time.Second).Result()
			}

			for _, m := range send("video", users, caption, vidb, "video.mp4") {
				msgs = append(msgs, m)
			}
			return msgs
		}
	}

	if photo != "" {
		caption = trimTo(caption+"\n"+description, 190-len([]rune(appendix))) + "\n" + appendix
		log.Println("caption", caption)

		b := httputils.HttpGet(photo, nil)
		if b != nil {

			if strings.HasSuffix(photo, ".gif") {
				for _, m := range send("doc", users, caption, b, "photo.gif") {
					msgs = append(msgs, m)
				}
			} else {
				for _, m := range send("photo", users, caption, b, "photo.png") {
					msgs = append(msgs, m)
				}
			}
		}
	} else {

		msgtxt := trimTo(title+"\n"+description, 3900-len([]rune(appendix))) + appendix

		for _, m := range send("link", users, msgtxt, nil, "") {
			msgs = append(msgs, m)
		}
	}
	return msgs
}

func getFeedPosts(link string, users map[int64]bool, wg *sync.WaitGroup) bool {

	var defHeaders = make(map[string]string)
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered in f %v\n", r)
			//wg.Done()
		}
	}()
	defHeaders["User-Agent"] = "script::recoilme:v1"
	defHeaders["Authorization"] = "Client-ID 4191ffe3736cfcb"
	b := httputils.HttpGet(link, defHeaders)
	if b == nil {
		//wg.Done()
		return false
	}
	fp := gofeed.NewParser()
	feed, err := fp.Parse(bytes.NewReader(b))
	if err != nil {
		//wg.Done()
		return false
	}
	var maxItems = 10
	if strings.Contains(link, "reddit") {
		maxItems = 20
	}
	var last = len(feed.Items) - 1
	if last > maxItems {
		last = maxItems
	}
	for i := range feed.Items {
		if i > last {
			break
		}
		item := feed.Items[last-i]

		key := GetMD5Hash(link) + "_" + GetMD5Hash(item.Link)
		//_, ok := feedpost.Get(key)
		//if ok {
		//TODO uncomment!
		//continue
		//} else {
		exists, err := red.Exists(key).Result()
		if err != nil {
			fmt.Println("redis", err)
			continue
		}
		if exists > 0 {
			continue
		}
		//body := httputils.HttpGet(params.Links+key, nil)
		//if body != nil {
		//TODO uncomment!
		//continue
		//}
		//}
		//body := httputils.HttpGet(params.Links+key, nil)
		//if body != nil {
		//TODO uncomment!
		//continue
		//}
		// pub feed

		//b, err := json.Marshal(item)
		//if err != nil {
		//continue
		//}
		red.Set(key, " ", (3600*24*30)*time.Second).Result()

		//saved := httputils.HttpPut(params.Links+key, nil, b)
		//feedpost.Set(key, " ")
		if link != "" && item.Link != "" {

			msgs := pubFeed(link, item, users, feed.FeedLink)
			chMsgs := make(chan tgMessage, len(msgs))
			//workers
			go workerMsg(chMsgs)

			for _, msg := range msgs {
				chMsgs <- msg
			}
			//all jobs send
			close(chMsgs)

			//log.Println("msgs send")
			//for _, msg := range msgs {
			//m := msg
			//go sendMsg(m)
			//}
		}
	}
	//wg.Done()
	return true
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
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
		//fmt.Printf("1.2 %s Feed: %s\n", time.Now().Format("15:04:05"), link)
		/*
			fmt.Printf("1.2 %s Feed: %s\n", time.Now().Format("15:04:05"), link)
			resp, err := http.Head(link)

			//	link:= "http://feeds.feedburner.com/~ff/ettoday/realtime?a=5V9lxn-vAwo:_Qjs5D4yEIg:yIl2AUoC8zA"
			//	resp, err := http.Head(u)
			//	fmt.Printf("1.3 %s Feed: %s\n", time.Now().Format("15:04:05"), link)

			fmt.Printf("1.3 %s Feed: %s\n", time.Now().Format("15:04:05"), link)
			if err != nil {
				continue
			}
			len, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
		*/
		len := int(httputils.HttpHead(link, nil))
		//fmt.Printf("1.4 %s Feed: %s\n", time.Now().Format("15:04:05"), link)
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
		//isImg := strings.HasPrefix(resp.Header.Get("Content-Type"), "image")
		//fmt.Printf("1.5 %s Feed: \n", time.Now().Format("15:04:05"))
		if len > 0 { //isImg {
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
	rand.Seed(time.Now().Unix())
	key := params.GooglKeys[rand.Intn(len(params.GooglKeys))]
	resp, err := http.Post(params.ShortUrl+key, "application/json", strings.NewReader(s))
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
				//https://thumbs.gfycat.com/FamiliarAgitatedDavidstiger-mobile.mp4
				break
			}
			if url.Host == "9gag.com" {
				video = strings.Replace(link, "9gag.com/gag", "img-9gag-fun.9cache.com/photo", -1)
				video = video + "_460sv.mp4"
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

func calcPubTop() {

	type kv struct {
		Key   vkapi.Group
		Value int
	}

	var ss []kv

	publics := getPubNames("asc")
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
			}
		}
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	for i, kv := range ss {
		if i == 30 {
			break
		}
		_ = kv
		//top.Set(strconv.Itoa(kv.Key.Gid), true)
	}
}
