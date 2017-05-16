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
	Example      = "\nExample: \nhttps://www.reddit.com/r/gifs/top/\nhttps://vk.com/evil_incorparate\n\nMore examples: http://telegra.ph/telefeedbot-05-12\n "
	SomeErr      = "🇬🇧 Something going wrong. Try later.. 🇷🇺 Ошибка, мать её!"
	Hello        = "🇬🇧 Send me link.\n\n🇷🇺 Отправь мне ссылку.\n\n" + Example
	Psst         = "🇬🇧 As soon as there are new articles here - i will  send them\nSend delete link_on_domain for unsubscribe\n\n🇷🇺 Я отправлю новый пост, как только он выйдет\nПришли delete ссылка_на_домен, чтобы отписаться"
	NotFound     = "🇬🇧 Not found\n🇷🇺 Домен не найден"
)

func init() {
	log.SetOutput(ioutil.Discard)
}
