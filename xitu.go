package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"github.com/mmcdole/gofeed"
)

const (
	feedURL = "http://gold.xitu.io/rss"
)

type Xitu struct {
}

func (x *Xitu) resolveFinalURL(link chan string, u string) string {
	retry := 0
doRequest:
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		fmt.Println("Could not parse get article content request:", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Could not send get article content request:", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Println("get article content request not 200")
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("cannot read get article content content", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
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
	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		fmt.Println("parsing feed URL failed", err)
		return
	}
	for _, item := range feed.Items {
		x.resolveFinalURL(link, item.Link)
	}
}
