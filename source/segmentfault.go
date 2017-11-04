package source

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/dfordsoft/dev2kindle/httputil"
	"github.com/mmcdole/gofeed"
)

const (
	sfFeedURL = "https://segmentfault.com/news/feeds"
)

type SegmentFault struct {
}

func (s *SegmentFault) extractFinalURL(u string) string {
	content := httputil.HttpGet(u)

	if len(content) == 0 {
		return ""
	}
	regex := regexp.MustCompile(`window.location.href= "([^"]+)`)
	list := regex.FindAllSubmatch(content, -1)
	for _, l := range list {
		return strings.Replace(string(l[1]), `\/`, `/`, -1)
	}
	return ""
}

func (s *SegmentFault) resolveFinalURL(link chan string, u string) {
	content := httputil.HttpGet(u)

	if len(content) == 0 {
		return
	}
	regex := regexp.MustCompile(`data\-url="([^"]+)`)
	list := regex.FindAllSubmatch(content, -1)
	for _, l := range list {
		followURL := string(l[1])
		if strings.Index(followURL, "https://segmentfault.com") != 0 {
			followURL = fmt.Sprintf("https://segmentfault.com%s", string(l[1]))
		}

		url := s.extractFinalURL(followURL)
		if len(url) > 0 {
			link <- url
			return
		}
	}
}

func (s *SegmentFault) Fetch(link chan string) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(sfFeedURL)
	if err != nil {
		log.Println("parsing feed URL failed", err)
		return
	}
	for _, item := range feed.Items {
		s.resolveFinalURL(link, item.Link)
	}
}
