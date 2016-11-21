package main

import "github.com/mmcdole/gofeed"

const (
	feedURL = "http://gold.xitu.io/rss"
)

type Xitu struct {
}

func (x *Xitu) Fetch(link chan string) {
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(feedURL)
	for _, item := range feed.Items {
		link <- item.Link
	}
}
