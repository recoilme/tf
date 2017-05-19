package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"strings"

	"github.com/recoilme/tf/httputils"
	"github.com/recoilme/tf/params"
	"github.com/recoilme/tf/vkapi"
)

func TestDonvk(t *testing.T) {
	log.Println("TestDonvk")
	posts := vkapi.WallGet(1257785 * (-1))
	log.Println("len", len(posts))

	domains := getDomainNames()
	for _, domainName := range domains {

		//log.Println(domainName)

		b := httputils.HttpGet(params.Publics+domainName, nil)
		if b != nil {
			var domain vkapi.Group
			err := json.Unmarshal(b, &domain)
			if err == nil {
				log.Println(domain.ScreenName)
				users := domUsers(domain)
				if len(users) > 0 {
					log.Println("saveposts", domain.ScreenName)
					//saveposts(domain, users)
				}
			}
		}

		//time.Sleep(100 * time.Millisecond)
	}

	log.Println("len dom", len(domains))
}

func TestFeeds(t *testing.T) {
	log.Println("TesRandd")
	rand.Seed(time.Now().Unix())
	log.Println("Magic 8-Ball says:", params.Stores[rand.Intn(len(params.Stores))])

	log.Println("TestFeeds")
	var url = "https://api.vk.com/method/wall.get?domain=driveru&v=5.63"
	var offset = 0
	//count := 50
	for i := 0; i < 1; i++ {
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
			//log.Println(href + "\t" + strconv.Itoa(virality))
		}
		i++
	}
}
