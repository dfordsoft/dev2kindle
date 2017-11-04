package source

import (
	"log"

	"github.com/dfordsoft/dev2kindle/config"
	"github.com/mmcdole/gofeed"
)

type RSSFeed struct {
}

func (r *RSSFeed) Fetch(link chan string) {
	fp := gofeed.NewParser()
	for _, f := range config.Data.RSSFeeds {
		feed, err := fp.ParseURL(f)
		if err != nil {
			log.Println("parsing feed URL failed", err)
			return
		}
		for _, item := range feed.Items {
			link <- item.Link
		}
	}
}
