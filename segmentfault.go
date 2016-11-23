package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/mmcdole/gofeed"
)

const (
	sfFeedURL = "https://segmentfault.com/news/feeds"
)

type SegmentFault struct {
}

func (s *SegmentFault) extractFinalURL(u string) string {
	retry := 0
doRequest:
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		fmt.Println("Could not parse get final URL request:", err)
		return ""
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Could not send get final URL request:", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return ""
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Println("get final URL request not 200")
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return ""
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("cannot read get final URL content", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
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
	retry := 0
doRequest:
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		fmt.Println("Could not parse get redirect URL request:", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Could not send get redirect URL request:", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Println("get redirect URL request not 200")
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("cannot read get redirect URL content", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
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
		x.resolveFinalURL(link, item.Link)
	}
}
