package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

var (
	crawlSubjectIDs = []int{
		4755,   // 三高：高可用高性能高并发
		1996,   // JAVA 程序员技术分享
		46756,  // redis
		101482, // 进击的架构师
		50375,  // APP 后端开发
		14908,  // fir.im
	}
)

type Toutiao struct {
}

func (t *Toutiao) resolveFinalURL(u string) string {
	retry := 0
doResolve:
	resp, err := http.Get(u)
	if err != nil {
		fmt.Printf("resolving url %s failed => %v\n", u, err.Error())
		if retry < 3 {
			retry++
			time.Sleep(3 * time.Second)
			goto doResolve
		} else {
			return ""
		}
	}

	finalURL := resp.Request.URL.String()
	return finalURL
}

func (t *Toutiao) Fetch(link chan string) {
	now := time.Now()
	u := fmt.Sprintf("https://toutiao.io/prev/%4.4d-%2.2d-%2.2d", now.Year(), now.Month(), now.Day())
	t.fetchArticles(link, u)

	for _, id := range crawlSubjectIDs {
		u := fmt.Sprintf("https://toutiao.io/subjects/%d", id)
		t.fetchArticles(link, u)
	}
}

func (t *Toutiao) fetchArticles(link chan string, u string) {
	retry := 0
doRequest:
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		fmt.Println("Could not parse get page list request:", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Could not send get page list request:", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Println("get page list request not 200")
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("cannot read get page list content", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
	}

	regex := regexp.MustCompile(`/k/([0-9a-zA-Z]+)`)
	list := regex.FindAllSubmatch(content, -1)
	for _, l := range list {
		lnk := fmt.Sprintf("https://toutiao.io/k/%s", string(l[1]))
		resolvedURL := t.resolveFinalURL(lnk)
		if resolvedURL != "" {
			link <- resolvedURL
		}
	}
}
