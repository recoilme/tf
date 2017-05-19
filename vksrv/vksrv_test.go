package main

import (
	"log"
	"strconv"
	"testing"
	"time"

	"strings"

	"github.com/recoilme/tf/vkapi"
)

func TestFeeds(t *testing.T) {
	var url = "https://api.vk.com/method/wall.get?domain=driveru&v=5.63"
	var offset = 0
	//count := 50
	for i := 0; i < 50; i++ {
		feed := url + "&offset=" + strconv.Itoa(offset)
		//fmt.Println(feed)
		TesPosts(feed)

		offset = offset + 20
		time.Sleep(1 * time.Second)
	}

}

func TesPosts(url string) {
	//url := "https://api.vk.com/method/wall.get?domain=driveru&v=5.63"
	posts := vkapi.PostsGet(url)
	var i = 1
	for _, post := range posts {
		likes := post.Likes.Count
		views := post.Views.Count
		var virality = 0
		if likes != 0 && views > 100 {
			virality = (likes) * 10000 / (views)
		}
		var href = ""
		if len(post.Attachments) > 0 {
			for _, att := range post.Attachments {
				if att.Link != nil {
					href = strings.Replace(att.Link.URL, "http://drive.ru", "https://www.drive.ru", -1)
					break
				}
			}
		}
		if href != "" && virality > 0 {
			log.Println(href + "\t" + strconv.Itoa(virality))
		}
		i++
	}
}
