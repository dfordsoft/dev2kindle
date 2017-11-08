package source

import (
	"log"

	"github.com/dfordsoft/dev2kindle/config"
	"github.com/mmcdole/gofeed"
)

func init() {
	config.RegisterInitializer(func() {
		if config.Data.RSSEnabled && len(config.Data.RSSFeeds) > 0 {
			t := &rssFeed{}
			config.RegisterSource(t.fetch)
		}
	})
}

type rssFeed struct {
}

func (r *rssFeed) fetch(link chan string) {
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
