package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"time"
)

type NewsList struct {
	HTML string `json:"html"`
}

type GeekCSDN struct {
}

func (c *GeekCSDN) extractURLs(link chan string, html string) {
	regex := regexp.MustCompile(`<a href="([^"]+)" class="title" target="_blank">`)
	list := regex.FindAllStringSubmatch(html, -1)
	for _, l := range list {
		if len(l[1]) > 0 {
			link <- l[1]
		}
	}
}

func (c *GeekCSDN) Fetch(link chan string) {
	getValues := url.Values{
		"username": {""},
		"from":     {"-"},
		"size":     {"20"},
		"type":     {"HackCount"},
		"_":        {fmt.Sprintf("%d", time.Now().Unix())},
	}
	u := `http://geek.csdn.net/service/news/get_news_list?` + getValues.Encode()
	content := httpGet(u)

	if len(content) == 0 {
		return
	}
	var newsList NewsList
	if err := json.Unmarshal(content, &newsList); err != nil {
		fmt.Println("unmarshalling failed", err)
		return
	}

	c.extractURLs(link, newsList.HTML)
}
