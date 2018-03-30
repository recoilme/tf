package params

import (
	"io/ioutil"
	"log"
)

const (
	host         = "localhost:5000"
	Api          = "http://" + host + "/bolt"
	Vkwriterfile = "vkwriter.bot"
	users        = "/usertg/"
	channels     = "/channels/"
	PubNames     = "/pubNames/"
	username     = "/usernametg/"
	pubSubTg     = "/pubSubTg/"
	feedSubTg    = "/feedSubTg/"
	userSubTg    = "/userSubTg/"
	lastPost     = "/vkpublastpost/"
	Feed         = "/feeds/"
	best         = "/best/"
	viralhash    = "/viralhash/"
	links        = "/links/"
	video        = "/video/"
	ShortUrl     = "https://www.googleapis.com/urlshortener/v1/url?key="
	TgApi        = "https://api.telegram.org"
	//MyakotkaId   = -1001067277325
	BaseUri   = Api + "/"
	Publics   = Api + PubNames
	Feeds     = Api + Feed
	Links     = Api + links
	Video     = Api + video
	Users     = Api + users
	Channels  = Api + channels
	Bests     = Api + best
	ViralHash = Api + viralhash
	UserName  = Api + username
	Subs      = Api + pubSubTg
	FeedSubs  = Api + feedSubTg
	UserSubs  = Api + userSubTg
	LastPost  = Api + lastPost
	//YaToken    = "1309689c-a584-41f2-a0e8-299747bb6326"
	Example    = "\nExample: \nhttps://www.reddit.com/r/gifs/top/\nhttps://vk.com/evil_incorparate\n\nMore examples: http://telegra.ph/telefeedbot-05-12\n\nЛучшее из вконтакте - собрано здесь: @memefeed\n\nТоп пабликов и фидов по подписчикам: /top"
	SomeErr    = "🇬🇧 Something going wrong. Try later.. 🇷🇺 Ошибка, мать её!"
	Hello      = "🇬🇧 Send me a link on domain/rss.\n\n🇷🇺 Отправь мне ссылку на домен или rss.\n\n" + Example
	Psst       = "🇬🇧 As soon as there are new articles here - i will send them, but with some delay (1-3 hour)\n\n🇷🇺 Я отправлю новые посты, но с некоторой задержкой (1-3 часа)"
	NotFound   = "🇬🇧 Rss feed not found\nPls send me direct link on rss\n\n🇷🇺 Rss поток не найден\nПришли, пожалуйста, прямую ссылку на rss\n"
	NewChannel = "🇬🇧 Add @telefeedbot as admin in channel\nSend me link on channel, example: https://t.me/channel\n\n 🇷🇺 Добавь @telefeedbot как админа в канал\nПришли ссылку на канал в формате: https://t.me/channel\n"
	SubsHelp   = "🇬🇧 Commands:\nAdd url:\n@channelname url(s)\nDelete url(s):\n@channelname delete url"
	Rate       = "Please rate me here ❤️❤️❤️:\nhttps://storebot.me/bot/telefeedbot\n\nSupport(если что-то не работает кликай сюда): https://t.me/joinchat/AAAAAEMFJOGkHNVp8qKQ1g"
	TopLinks   = `
	Top rss feeds (by subscribers):

0 https://www.reddit.com/r/gifs/top/.rss

1 http://itc.ua/feed/

2 http://pikabu.ru/xmlfeeds.php?cmd=popular

3 https://web.stagram.com/rss/n/p

4 https://wylsa.com/feed/

5 http://news.liga.net/all/rss.xml

6 http://www.opennet.ru/opennews/opennews_all.rss

7 http://feeds.feedburner.com/macdigger/

8 http://droider.ru/feed/

9 https://xakep.ru/feed/

Топ пабликов вконтакте (по подписчикам):

0 Самое полайканное из вконтакте - собрано в отдельный канал: @memefeed

1 https://vk.com/evil_incorparate
Журнал сарказма, подстебай своих друзей. 

2 https://vk.com/4ch

3 https://vk.com/mudakoff
Не забывай свои корни МДК | MDK | ЭМДИКЕЙ 

4 https://vk.com/leprum
Ле́пра (Проказа) — хроническое инфекционное заболевание, протекающие с преимущественным поражением ..

5 https://vk.com/pikabu
Мы хотим, чтобы вы проводили время с интересом. Наши посты проходят жесткий отбор: сначала ..

6 https://vk.com/oldlentach

7 https://vk.com/tnull

8 https://vk.com/chotkiy_paca

9 https://vk.com/borsch
 
10 https://vk.com/tj
Новости интернета. Ты либо в тренде, либо уходи. 

11 https://vk.com/typical_kiev

12 https://vk.com/leprazo

13 https://vk.com/ru9gag
Just for fun!Repost of the best jokes from 9gag.com9GAG is your best source of happiness ..

14 https://vk.com/science_technology
Будущее рядом 

15 https://vk.com/designmdk
Хуевый графический дизайн.Самое охуевшее сообщество. 

16 https://vk.com/marvel_dc

17 https://vk.com/paper.comics

18 https://vk.com/pn6
Рекомендовано для 18+Привет, друг мой. Если ты здесь, значит тебе есть что рассказать о ..

19 https://vk.com/overhear
Здесь только то, что присылают наши читатели.Поделиться откровением совершенно анонимно ..

20 https://vk.com/dzenpub

21 https://vk.com/igm
Самый популярный паблик для геймеров! 

22 https://vk.com/oko_mag
ОКО видит всё, в том числе невидимое и вытесняемое — "некрасивое", "неприличное", "неинтересное". 

23 https://vk.com/chop.choppp

24 https://vk.com/fuck_humor

25 https://vk.com/cliqque
CLIQUE — это поп-культурное оружие c полным магазином патронов. 24 часа в сутки и 7 дней в неделю ..

26 https://vk.com/sci
Первый познавательный паблик ВКонтакте. 

27 https://vk.com/stolbn

28 https://vk.com/rapnewrap
(с) Больше, чем про рэпВещаем с 2011 года..

29 https://vk.com/ukrlit_memes
Вiдроджуємо українську лiтературу 

30 https://vk.com/styd.pozor
Паблик анонимных рассказов от реальных людей. Не каждой историей можно поделиться с родными, ..

More examples: http://telegra.ph/telefeedbot-05-12
	
`
)

var (
	Telefeedfile = "./telefeed.bot"
	//Telefeedfile       = "./telefeedtst.bot"
	ChannelsFatherfile = "./channelsfather.bot"
	Tokens             = [...]string{"token1", "token2"}
	Stores             = [...]string{"@telefeedcontent1", "@telefeedcontent2", "@telefeedcontent3"}
	StoreIds           = [...]int64{-1001140338639, -1001144965998, -1001122084977, -1001121449455, -1001147806509, -1001069985583, -1001128095164}
	GooglKeys          = [...]string{"AIzaSyCTmUsTGqjl7iWJLiJisrejgTNamp7bfIA", "AIzaSyAZaQiAkSmYFZLMjUOxtKOj3R29TPs81X0"}
)

func init() {
	/*
		f, err := os.OpenFile("./testlogfile.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.SetOutput(f)
		}
		defer f.Close()*/

	log.SetOutput(ioutil.Discard)
}
