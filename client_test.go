package btsite_test

import (
	"encoding/json"
	"github.com/heibizi/go-btsite"
	"os"
	"testing"
	"time"
)

var client btsite.Client

func TestMain(m *testing.M) {
	btsite.InitConfig(os.Getenv("GO_BTSITE_CONFIGS_PATH"))
	client, _ = btsite.NewClient(&btsite.Site{
		Code:      os.Getenv("GO_BTSITE_CODE"),
		Name:      os.Getenv("GO_BTSITE_NAME"),
		UserAgent: os.Getenv("GO_BTSITE_UA"),
		Cookie:    os.Getenv("GO_BTSITE_COOKIE"),
		RssUrl:    os.Getenv("GO_BTSITE_RSS_URL"),
	})
	m.Run()
}

func TestUserBasicInfo(t *testing.T) {
	info, err := client.UserBasicInfo()
	log(info, err, t)
	time.Sleep(1 * time.Second)
}

func TestUserDetails(t *testing.T) {
	details, err := client.UserDetails()
	log(details, err, t)
	time.Sleep(1 * time.Second)
}

func TestSeedingStatistics(t *testing.T) {
	statistics, err := client.SeedingStatistics()
	log(statistics, err, t)
	time.Sleep(1 * time.Second)
}

func TestFavicon(t *testing.T) {
	favicon, err := client.Favicon()
	log(favicon, err, t)
	time.Sleep(1 * time.Second)
}

func TestMyHr(t *testing.T) {
	hr, err := client.MyHr()
	log(hr, err, t)
	time.Sleep(1 * time.Second)
}

func TestUnreadMessage(t *testing.T) {
	messages, err := client.UnreadMessages(true)
	log(messages, err, t)
	time.Sleep(1 * time.Second)
}

func TestLatestNotice(t *testing.T) {
	notice, err := client.LatestNotice()
	log(notice, err, t)
	time.Sleep(1 * time.Second)
}

func TestSignIn(t *testing.T) {
	r, err := client.SignIn()
	log(r, err, t)
	time.Sleep(1 * time.Second)
}

func TestDetails(t *testing.T) {
	details, err := client.Details(os.Getenv("GO_BTSITE_TORRENT_ID"))
	log(details, err, t)
	time.Sleep(1 * time.Second)
}

func TestSearch(t *testing.T) {
	torrents, err := client.Search(btsite.SearchParams{
		Keyword:   "",
		MediaType: btsite.Movie,
		Page:      0,
	})
	log(torrents, err, t)
	if len(torrents) > 0 {
		url, err := client.GetDownloadUrl(torrents[0])
		log(url, err, t)
	}
	time.Sleep(1 * time.Second)
}

func TestRss(t *testing.T) {
	rss, err := client.Rss()
	log(rss, err, t)
	time.Sleep(1 * time.Second)
}

func log(v any, err error, t *testing.T) {
	if err != nil {
		t.Log(err)
		return
	}
	j, _ := json.MarshalIndent(v, "", "    ")
	t.Log(string(j))
}
