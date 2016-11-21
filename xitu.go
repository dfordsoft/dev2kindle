package main

import (
	"fmt"

	"github.com/mmcdole/gofeed"
)

const (
	feedURL = "http://gold.xitu.io/rss"
)

type Xitu struct {
}

func (x *Xitu) Fetch(link chan string) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		fmt.Println("parsing feed URL failed", err)
		return
	}
	for _, item := range feed.Items {
		link <- item.Link
	}
}
