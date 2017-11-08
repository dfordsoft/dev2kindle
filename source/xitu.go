package source

import (
	"log"
	"regexp"

	"github.com/dfordsoft/dev2kindle/config"
	"github.com/dfordsoft/dev2kindle/httputil"
	"github.com/mmcdole/gofeed"
)

func init() {
	config.RegisterInitializer(func() {
		if config.Data.GoldXituEnabled {
			t := &xitu{}
			config.RegisterSource(t.fetch)
		}
	})
}

const (
	xituFeedURL = "http://gold.xitu.io/rss"
)

type xitu struct {
}

func (x *xitu) resolveFinalURL(link chan string, u string) {
	content := httputil.HttpGet(u)

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

func (x *xitu) fetch(link chan string) {
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
