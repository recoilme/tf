package params

const (
	api      = "http://badtobefat.ru/bolt"
	users    = "/usertg/"
	pubNames = "/pubNames/"
	pubSubTg = "/pubSubTg/"
	lastPost = "/vkpublastpost/"
	BaseUri  = api + "/"
	Publics  = api + pubNames
	Users    = api + users
	Subs     = api + pubSubTg
	LastPost = api + lastPost
	SomeErr  = "Something going wrong. Try later.. Ошибка, мать её!"
	Hello    = "🇬🇧 Send me links to public pages from vk.com, and I will send you new articles.\n🇷🇺 Отправь мне ссылки на общедоступные страницы c vk.com, и я буду присылать тебе новые статьи.\nExample: https://vk.com/myakotkapub\nContacts: @recoilme"
	Psst     = "🇬🇧 Psst. As soon as there are new articles here - I will immediately send them\n🇷🇺 Псст. Как только появятся новые статьи здесь -  я их сразу пришлю"
)

func init() {
	//log.SetOutput(ioutil.Discard)
}
