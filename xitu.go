package main

import (
	"log"
	"regexp"

	"github.com/mmcdole/gofeed"
)

const (
	xituFeedURL = "http://gold.xitu.io/rss"
)

type Xitu struct {
}

func (x *Xitu) resolveFinalURL(link chan string, u string) {
	content := httpGet(u)

	if len(content) == 0 {
		return
	}
	regex := regexp.MustCompile(`<a href="([^"]+)" target="_blank" class="share\-link">`)
	list := regex.FindAllSubmatch(content, -1)
	for _, l := range list {
		if len(l) > 2 {
			link <- string(l[1])
		}
	}
}

func (x *Xitu) Fetch(link chan string) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(xituFeedURL)
	if err != nil {
		log.Println("parsing feed URL failed", err)
		return
	}
	for _, item := range feed.Items {
		x.resolveFinalURL(link, item.Link)
	}
}
