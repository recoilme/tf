package vkapi

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"strings"

	"log"

	"github.com/recoilme/tf/httputils"
	"github.com/recoilme/tf/params"
)

type PostResponse struct {
	Response struct {
		Count int    `json:"count"`
		Posts []Post `json:"items"`
	} `json:"response"`
}

type GroupResponse struct {
	Groups []Group `json:"response"`
}

type Group struct {
	Gid          int    `json:"gid"`
	Name         string `json:"name"`
	ScreenName   string `json:"screen_name"`
	IsClosed     int    `json:"is_closed"`
	Type         string `json:"type"`
	MembersCount int    `json:"members_count"`
	Description  string `json:"description"`
	Photo        string `json:"photo"`
	PhotoMedium  string `json:"photo_medium"`
	PhotoBig     string `json:"photo_big"`
}

type Post struct {
	Id          int    `json:"id"`
	FromId      int    `json:"from_id"`
	OwnerID     int    `json:"owner_id"`
	ToId        int    `json:"to_id"`
	Date        int    `json:"date"`
	MarkedAsAds int8   `json:"marked_as_ads"`
	PostType    string `json:"post_type"`
	Text        string `json:"text"`
	SignerId    int    `json:"signer_id"`
	IsPinned    int8   `json:"is_pinned"`
	//Attachment  Attachment   `json:"attachment"`
	Attachments []Attachment `json:"attachments"`
	Comments    struct {
		Count int `json:"count"`
	} `json:"comments"`
	Likes struct {
		Count int `json:"count"`
	} `json:"likes"`
	Reposts struct {
		Count int `json:"count"`
	} `json:"reposts"`
	Views struct {
		Count int `json:"count"`
	} `json:"views"`
}

type Attachment struct {
	Type  string `json:"type"`
	Photo *Photo `json:"photo"`
	Link  *Link  `json:"link"`
	Video *Video `json:"video"`
	Doc   *Doc   `json:"doc"`
}

type Doc struct {
	ID      int    `json:"id"`
	OwnerID int    `json:"owner_id"`
	Title   string `json:"title"`
	Size    int    `json:"size"`
	Ext     string `json:"ext"`
	URL     string `json:"url"`
	Date    int    `json:"date"`
	Type    int    `json:"type"`
	Preview struct {
		Photo struct {
			Sizes []struct {
				Src    string `json:"src"`
				Width  int    `json:"width"`
				Height int    `json:"height"`
				Type   string `json:"type"`
			} `json:"sizes"`
		} `json:"photo"`
		Video struct {
			Src      string `json:"src"`
			Width    int    `json:"width"`
			Height   int    `json:"height"`
			FileSize int    `json:"file_size"`
		} `json:"video"`
	} `json:"preview"`
	AccessKey string `json:"access_key"`
}

type Photo struct {
	ID        int    `json:"id"`
	AlbumID   int    `json:"album_id"`
	OwnerID   int    `json:"owner_id"`
	UserID    int    `json:"user_id"`
	Photo75   string `json:"photo_75"`
	Photo130  string `json:"photo_130"`
	Photo604  string `json:"photo_604"`
	Photo807  string `json:"photo_807"`
	Photo1280 string `json:"photo_1280"`
	Photo2560 string `json:"photo_2560"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Text      string `json:"text"`
	Date      int    `json:"date"`
	AccessKey string `json:"access_key"`
}

type Video struct {
	ID            int    `json:"id"`
	OwnerID       int    `json:"owner_id"`
	Title         string `json:"title"`
	Duration      int    `json:"duration"`
	Description   string `json:"description"`
	Date          int    `json:"date"`
	Comments      int    `json:"comments"`
	Views         int    `json:"views"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	Photo130      string `json:"photo_130"`
	Photo320      string `json:"photo_320"`
	Photo800      string `json:"photo_800"`
	AccessKey     string `json:"access_key"`
	Repeat        int    `json:"repeat"`
	FirstFrame320 string `json:"first_frame_320"`
	FirstFrame160 string `json:"first_frame_160"`
	FirstFrame130 string `json:"first_frame_130"`
	FirstFrame800 string `json:"first_frame_800"`
	CanAdd        int    `json:"can_add"`
}
type Link struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Caption     string `json:"caption"`
	Description string `json:"description"`
	Photo       struct {
		ID       int    `json:"id"`
		AlbumID  int    `json:"album_id"`
		OwnerID  int    `json:"owner_id"`
		Photo75  string `json:"photo_75"`
		Photo130 string `json:"photo_130"`
		Photo604 string `json:"photo_604"`
		Width    int    `json:"width"`
		Height   int    `json:"height"`
		Text     string `json:"text"`
		Date     int    `json:"date"`
	} `json:"photo"`
}

// WallGet return array of Post by domain name
// get ownerid or screenname as param
func WallGet(domain interface{}) []Post {
	//549d16f616cced9143b53d66d28e0a4348e3a452bf026b08f5ef2411ec3c55587c50cdcaa3e77236ac007 //burnnews
	//fd203e2d8a7a4af5c59ac52a6a297460ab5d4136b2f89c57b399c86e11706a226fc8a521af90f923e63c6 //fontanka
	//d091f7d5f6e97966aa4d8cf461f01d0e0373d8d83a5f2c52523fd434d72b6d5f895b1ee782bf097e9516e //telefeed //blocked
	//7f31f1c447cf23d335b02d4d195f22b34249bb4a942f8a86c7594fb10ec29a30e140713f494cfeb882cac //telefeedbot //blocked
	//e197dc3dc93a61be680b74781b961932b9c99a19430c5c9dd68d863192fcfdd3920c346c29e7558c1a45f //telefeedbot2 //blocked
	//https://oauth.vk.com/authorize?client_id=5589502&scope=groups%2Cwall%2Coffline%2Cphotos%2Cvideos%2Caudios%2Cdocuments&redirect_uri=https://oauth.vk.com/blank.html&display=page&v=5.63&response_type=token
	//7f31f1c447cf23d335b02d4d195f22b34249bb4a942f8a86c7594fb10ec29a30e140713f494cfeb882cac

	apikey := params.Tokens[rand.Intn(len(params.Tokens))]
	token := "&access_token=" + apikey
	var url string
	//https://api.vk.com/method/wall.get?owner_id=-125698500&v=5.63
	switch domain.(type) {
	case int:
		ownerid := domain.(int)
		if ownerid > 0 {
			ownerid = ownerid * (-1)
		}
		url = fmt.Sprintf("https://api.vk.com/method/wall.get?owner_id=%d&v=5.63%s", ownerid, token)
	case string:
		url = fmt.Sprintf("https://api.vk.com/method/wall.get?domain=%s&v=5.63%s", domain.(string), token)
	default:
		return make([]Post, 0, 0)
	}
	return PostsGet(url)
}

func PostsGet(url string) []Post {
	posts := make([]Post, 0, 20)
	body := httputils.HttpGet(url, nil)
	if body != nil {
		var postRes PostResponse
		err := json.Unmarshal(body, &postRes)
		if err == nil {
			for i := range postRes.Response.Posts {
				posts = append(posts, postRes.Response.Posts[i])
			}
		}
	}
	return posts
}

// GroupsGetById return groups, where name = shortname or vk public id
func GroupsGetById(name string) (groups []Group) {
	if strings.HasPrefix(name, "public") {
		name = name[len("public"):]
		log.Println("name", name)
	}
	url := "https://api.vk.com/method/groups.getById?group_id=" + name + "&v=5.63&fields=members_count,description"
	body := httputils.HttpGet(url, nil)
	if body != nil {
		var groupRes GroupResponse
		err := json.Unmarshal(body, &groupRes)
		if err == nil {
			groups = groupRes.Groups
		}
	}
	return
}
