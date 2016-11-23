package main

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/mmcdole/gofeed"
)

const (
	sfFeedURL = "https://segmentfault.com/news/feeds"
)

type SegmentFault struct {
}

func (s *SegmentFault) extractFinalURL(u string) string {
	content := httpGet(u)

	if len(content) == 0 {
		return ""
	}
	regex := regexp.MustCompile(`window.location.href= "([^"]+)`)
	list := regex.FindAllSubmatch(content, -1)
	for _, l := range list {
		if url, e := strconv.Unquote(string(l[1])); e != nil {
			fmt.Println("unquoting string failed", e)
		} else {
			return url
		}
	}
	return ""
}

func (s *SegmentFault) resolveFinalURL(link chan string, u string) {
	content := httpGet(u)

	if len(content) == 0 {
		return
	}
	regex := regexp.MustCompile(`data\url="([^"]+)`)
	list := regex.FindAllSubmatch(content, -1)
	for _, l := range list {
		followURL := fmt.Sprintf("https://segmentfault.com%s", string(l[1]))
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
		fmt.Println("parsing feed URL failed", err)
		return
	}
	for _, item := range feed.Items {
		s.resolveFinalURL(link, item.Link)
	}
}
