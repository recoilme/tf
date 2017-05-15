package params

import (
	"io/ioutil"
	"log"
)

const (
	api = "http://badtobefat.ru/bolt"
	//api          = "http://localhost:5000/bolt"
	//Telefeedfile = "telefeedtst.bot"
	Telefeedfile = "telefeed.bot"
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
	BaseUri      = api + "/"
	Publics      = api + pubNames
	Feeds        = api + feeds
	Links        = api + links
	Users        = api + users
	UserName     = api + username
	Subs         = api + pubSubTg
	FeedSubs     = api + feedSubTg
	UserSubs     = api + userSubTg
	LastPost     = api + lastPost
	Example      = "\nExample: \nhttps://vk.com/myakotkapub\n"
	SomeErr      = "🇬🇧 Something going wrong. Try later.. 🇷🇺 Ошибка, мать её!"
	Hello        = "🇬🇧 Send me links to public pages from vk.com, and I will send you new articles.\n🇷🇺 Отправь мне ссылки на общедоступные страницы c vk.com, и я буду присылать тебе новые статьи с них.\n" + Example + "\nContacts: @recoilme"
	Psst         = "🇬🇧 Psst. As soon as there are new articles here - I will immediately send them\nSend delete link_on_domain for unsubscribe\n\n🇷🇺 Псст. Как только появятся новые статьи здесь -  я их сразу пришлю\nПришли delete ссылка_на_домен, чтобы отписаться"
	NotFound     = "🇬🇧 Not found\n🇷🇺 Домен не найден"
)

func init() {
	log.SetOutput(ioutil.Discard)
}
