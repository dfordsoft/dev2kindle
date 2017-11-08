package source

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"time"

	"github.com/dfordsoft/dev2kindle/config"
	"github.com/dfordsoft/dev2kindle/httputil"
)

func init() {
	config.RegisterInitializer(func() {
		if config.Data.GeekCSDNEnabled {
			t := &geekCSDN{}
			config.RegisterSource(t.fetch)
		}
	})
}

type geekCSDN struct {
}

func (c *geekCSDN) extractURLs(link chan string, html string) {
	regex := regexp.MustCompile(`<a href="([^"]+)" class="title" target="_blank">`)
	list := regex.FindAllStringSubmatch(html, -1)
	for _, l := range list {
		if len(l[1]) > 0 {
			link <- l[1]
		}
	}
}

func (c *geekCSDN) fetch(link chan string) {
	getValues := url.Values{
		"username": {""},
		"from":     {"-"},
		"size":     {"20"},
		"type":     {"HackCount"},
		"_":        {fmt.Sprintf("%d", time.Now().Unix())},
	}
	u := `http://geek.csdn.net/service/news/get_news_list?` + getValues.Encode()
	content := httputil.HttpGet(u)

	if len(content) == 0 {
		return
	}
	var newsList struct {
		HTML string `json:"html"`
	}

	if err := json.Unmarshal(content, &newsList); err != nil {
		log.Println("unmarshalling failed", err)
		return
	}

	c.extractURLs(link, newsList.HTML)
}
