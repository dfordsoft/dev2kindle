package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dfordsoft/dev2kindle/httputil"
)

var (
	client           *http.Client
	noRedirectClient *http.Client
)

type Instapaper struct {
	Username string
	Password string
	cookie   string
	formKey  string
}

func (i *Instapaper) Login() {
	postValues := url.Values{
		"username":       {i.Username},
		"password":       {i.Password},
		"keep_logged_in": {"yes"},
	}

	retry := 0
doRequest:
	req, err := http.NewRequest("POST", "https://www.instapaper.com/user/login", strings.NewReader(postValues.Encode()))
	if err != nil {
		log.Println("Could not parse login request:", err)
		return
	}

	req.Header.Set("Referer", "http://www.instapaper.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("accept-language", `en-US,en;q=0.5`)
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := noRedirectClient.Do(req)
	if err != nil {
		log.Println("Could not send login request:", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != 302 {
		log.Println("login request not 302:", resp.StatusCode)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("cannot read login content", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}

	cookies := resp.Cookies()
	var cookie []string
	for _, v := range cookies {
		ss := strings.Split(v.String(), ";")
		for _, c := range ss {
			if strings.Contains(c, "pfp=") || strings.Contains(c, "pfu=") || strings.Contains(c, "pfh=") {
				i := strings.Index(c, "=")
				cookie = append(cookie, c[i-3:])
			}
		}
	}
	i.cookie = strings.Join(cookie, "; ")
}

func (i *Instapaper) EditUrl(u string) {
	postValues := url.Values{
		"bookmark[url]": {u},
		"form_key":      {i.formKey},
	}

	httputil.HttpPostBasicAuth("https://www.instapaper.com/edit", postValues.Encode(), i.Username, i.Password)
}

func (i *Instapaper) AddUrl(u string) {
	postValues := url.Values{
		"url": {u},
	}

	httputil.HttpPostBasicAuth("https://www.instapaper.com/api/add", postValues.Encode(), i.Username, i.Password)
}

func (i *Instapaper) getIDs(u string) (res []int) {
	headers := map[string]string{
		"Referer":                   "http://blog.instapaper.com/post/152600596211",
		"User-Agent":                "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0",
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"accept-language":           `en-US,en;q=0.5`,
		"Upgrade-Insecure-Requests": "1",
		"Connection":                "keep-alive",
		"Cache-Control":             "max-age=0",
		"Cookie":                    i.cookie,
	}
	content := httputil.HttpGetCustomHeader(u, headers)
	if len(content) == 0 {
		return res
	}

	r := regexp.MustCompile(`/delete/([0-9]+)`)
	ss := r.FindAllSubmatch(content, -1)
	for _, s := range ss {
		id, e := strconv.Atoi(string(s[1]))
		if e == nil {
			res = append(res, id)
		}
	}
	return res
}

func (i *Instapaper) removeLink(id int) {
	headers := map[string]string{
		"Referer":                   "http://www.instapaper.com/u",
		"User-Agent":                "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0",
		"Accept":                    "*/*",
		"accept-language":           `en-US,en;q=0.5`,
		"Upgrade-Insecure-Requests": "1",
		"Connection":                "keep-alive",
		"Cache-Control":             "max-age=0",
		"Content-Type":              "application/x-www-form-urlencoded; charset=UTF-8",
		"X-Requested-With":          "XMLHttpRequest",
		"Cookie":                    i.cookie,
	}
	httputil.HttpPostCustomHeader("https://www.instapaper.com/delete_articles", fmt.Sprintf("[%d]", id), headers, false)
}

func (i *Instapaper) RemoveAllLinks() {
	var ids []int
	for index := 1; index < 10; index++ {
		u := fmt.Sprintf("https://www.instapaper.com/u/%d", index)
		l := i.getIDs(u)
		if len(l) == 0 {
			break
		}
		ids = append(ids, l...)
	}
	for _, id := range ids {
		i.removeLink(id)
	}
}

func (i *Instapaper) PushToKindle() {
	postValues := url.Values{
		"form_key": {i.formKey},
		"submit":   {"1"},
	}

	headers := map[string]string{
		"Referer":                   "http://www.instapaper.com/u",
		"User-Agent":                "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0",
		"Accept":                    "*/*",
		"accept-language":           `en-US,en;q=0.5`,
		"Upgrade-Insecure-Requests": "1",
		"Connection":                "keep-alive",
		"Cache-Control":             "max-age=0",
		"Content-Type":              "application/x-www-form-urlencoded; charset=UTF-8",
		"X-Requested-With":          "XMLHttpRequest",
		"Cookie":                    i.cookie,
	}
	httputil.HttpPostCustomHeader("https://www.instapaper.com/user/kindle_send_now", postValues.Encode(), headers, true)
}

func (i *Instapaper) GetFormKey() {
	headers := map[string]string{
		"Referer":                   "https://www.instapaper.com/u",
		"User-Agent":                "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0",
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"accept-language":           `en-US,en;q=0.5`,
		"Upgrade-Insecure-Requests": "1",
		"Connection":                "keep-alive",
		"Cache-Control":             "max-age=0",
		"Cookie":                    i.cookie,
	}
	content := httputil.HttpGetCustomHeader("https://www.instapaper.com/user", headers)
	if len(content) == 0 {
		return
	}
	r := regexp.MustCompile(`<input id="form_key" name="form_key" type="hidden" value="([0-9a-zA-Z]+)" />`)
	ss := r.FindAllSubmatch(content, -1)
	i.formKey = string(ss[0][1])
}
