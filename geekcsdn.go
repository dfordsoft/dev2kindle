package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

type NewsList struct {
	HTML string `json:"html"`
}

type GeekCSDN struct {
}

func (c *GeekCSDN) extractURLs(link chan string, html string) {
	regex := regexp.MustCompile(`<a href="([^"]+)" class="title" target="_blank">`)
	list := regex.FindAllStringSubmatch(html, -1)
	for _, l := range list {
		if len(l[1]) > 0 {
			link <- l[1]
		}
	}
}

func (c *GeekCSDN) Fetch(link chan string) {
	retry := 0
doRequest:
	getValues := url.Values{
		"username": {""},
		"from":     {"-"},
		"size":     {"20"},
		"type":     {"HackCount"},
		"_":        {fmt.Sprintf("%d", time.Now().Unix())},
	}
	req, err := http.NewRequest("GET", `http://geek.csdn.net/service/news/get_news_list?`+getValues.Encode(), nil)
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

	var newsList NewsList
	if err = json.Unmarshal(content, &newsList); err != nil {
		fmt.Println("unmarshalling failed", err)
		return
	}

	c.extractURLs(link, newsList.HTML)
}
