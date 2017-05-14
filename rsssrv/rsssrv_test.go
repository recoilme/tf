package main

import (
	"log"
	"testing"
)

func TestRssdomains(t *testing.T) {
	//domains := rssdomains()
	//log.Println("group1", domains)
}

func TestRss2(t *testing.T) {
	/*s := `<img src="https://habrastorage.org/web/85d/cb2/cc2/85dcb2cc27ac4eeb9e2d034d8c3e3e28.png"/><br/>
	Расскажу, как классификация текста помогла мне в поиске квартиры, а также почему я отказался от регулярных выражений и нейронных сетей и стал использовать лексический анализатор.<br/>
	 <a href="https://habrahabr.ru/post/328282/?utm_source=habrahabr&amp;utm_medium=rss&amp;utm_campaign=best#habracut">Читать дальше &rarr;</a>`
	*/
	//s := `oft <a href="http://1.ru">12</a>Build 2017. <img src="https://png.cmtt.space/paper-preview-fox/m/ic/microsoft-build-announcements/8d1c780b2eba-original.jpg">`
	s := ` по улучшению взаимодействия машин и человека сейчас актуальна как никогда. Появились технические возможности для перехода от модели «100 кликов» к парадигме «скажи, что ты хочешь». Да, я имею в виду различные боты, которые уже несколько лет разрабатывают все кому не лень. К примеру, многие крупные компании, не только технологические, но и retail, logistics, банки в данный момент ведут активный Research&Design в этой области. <br/>
<br/>
Простой пример, как, например, происходит процесс выбора товаров в каком-либо интернет магазине? Куча списков, категорий, в которых я роюсь и что-то выбираю. It suck's. Или, допустим, заходя в интернет банк, я сталкиваюсь с различными меню, если я хочу сделать перевод, то я должен выбрать соответствующие пункты в меню и ввести кучу данных, если же я хочу посмотреть список транзакций, то опять таки, я должен напрягать как мозг, так и указательный палец. Гораздо проще и удобнее было бы зайти на страницу, и просто сказать: «Я хочу купить литр молока и пол-литра водки», или просто спросить у банка: «Что с деньгами?».<br/>
<br/>
В список профессий, которым грозит вымирание в достаточно близкой перспективе, добавляются: теллеры, операторы call центров, и многие другие. И на простом примере, реализовать который у меня заняло часов 7, я покажу, как можно достаточно просто сделать интеграцию распознавания речи, и выявления сущностей, на примере открытого Wit.Ai (Google Speech API интеграция также включена)<br/>
<img src="https://habrastorage.org/web/f1c/f84/d32/f1cf84d327f444cd8023382c4b313463.jpg"/><br/>
 <a href="https://habrahabr.ru/post/328612/?utm_source=habrahabr&amp;utm_medium=rss&amp;utm_campaign=best#habracut">Читать дальше &rarr;</a>`
	links := extractLinks(s)
	log.Println("!!", links)
	imgs := getImgs(links)
	var max = 0
	var maximg = ""
	for img, len := range imgs {
		if len > max {
			max = len
			maximg = img
		}
		log.Println(img, "len", len)
	}
	log.Println(maximg)
}
