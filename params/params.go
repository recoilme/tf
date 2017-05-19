package params

import (
	"io/ioutil"
	"log"
)

const (
	//host = "localhost:5000"
	host         = "badtobefat.ru"
	api          = "http://" + host + "/bolt"
	Vkwriterfile = "vkwriter.bot"
	users        = "/usertg/"
	pubNames     = "/pubNames/"
	username     = "/usernametg/"
	pubSubTg     = "/pubSubTg/"
	feedSubTg    = "/feedSubTg/"
	userSubTg    = "/userSubTg/"
	lastPost     = "/vkpublastpost/"
	feeds        = "/feeds/"
	links        = "/links/"
	ShortUrl     = "https://www.googleapis.com/urlshortener/v1/url?key=AIzaSyCTmUsTGqjl7iWJLiJisrejgTNamp7bfIA"
	TgApi        = "https://api.telegram.org"
	//MyakotkaId   = -1001067277325
	BaseUri  = api + "/"
	Publics  = api + pubNames
	Feeds    = api + feeds
	Links    = api + links
	Users    = api + users
	UserName = api + username
	Subs     = api + pubSubTg
	FeedSubs = api + feedSubTg
	UserSubs = api + userSubTg
	LastPost = api + lastPost
	Example  = "\nExample: \nhttps://www.reddit.com/r/gifs/top/\nhttps://vk.com/evil_incorparate\n\nMore examples: http://telegra.ph/telefeedbot-05-12\n "
	SomeErr  = "🇬🇧 Something going wrong. Try later.. 🇷🇺 Ошибка, мать её!"
	Hello    = "🇬🇧 Send me a link.\n\n🇷🇺 Отправь мне ссылку.\n\n" + Example
	Psst     = "🇬🇧 As soon as there are new articles here - i will  send them\nSend delete link_on_domain for unsubscribe\n\n🇷🇺 Я отправлю новый пост, как только он выйдет\nПришли delete ссылка_на_домен, чтобы отписаться"
	NotFound = "🇬🇧 Not found\n🇷🇺 Домен не найден"
)

var (
	Telefeedfile = "telefeedtst.bot"
	Stores       = [...]string{"@telefeedcontent1", "@telefeedcontent2", "@telefeedcontent3"}
	StoreIds     = [...]int64{-1001140338639, -1001144965998, -1001122084977}
)

func init() {
	if host == "badtobefat.ru" {
		log.SetOutput(ioutil.Discard)
		Telefeedfile = "telefeed.bot"
	} else {
		//log.SetOutput(ioutil.Discard)

	}
}
