package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
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
		fmt.Println("Could not parse login request:", err)
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
		fmt.Println("Could not send login request:", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != 302 {
		fmt.Println("login request not 302:", resp.StatusCode)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("cannot read login content", err)
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

	retry := 0
doRequest:
	req, err := http.NewRequest("POST", "https://www.instapaper.com/edit", strings.NewReader(postValues.Encode()))
	if err != nil {
		fmt.Println("Could not parse edit link request:", err)
		return
	}

	req.SetBasicAuth(i.Username, i.Password)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Could not send edit link request:", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		fmt.Println("edit link request not 201", resp.StatusCode)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("cannot read edit link content", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}
}

func (i *Instapaper) AddUrl(u string) {
	postValues := url.Values{
		"url": {u},
	}

	retry := 0
doRequest:
	req, err := http.NewRequest("POST", "https://www.instapaper.com/api/add", strings.NewReader(postValues.Encode()))
	if err != nil {
		fmt.Println("Could not parse add link request:", err)
		return
	}

	req.SetBasicAuth(i.Username, i.Password)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Could not send add link request:", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		fmt.Println("add link request not 201", resp.StatusCode)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("cannot read add link content", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}
}

func (i *Instapaper) getIDs(u string) (res []int) {
	retry := 0
doRequest:
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		fmt.Println("Could not parse get IDs request:", err)
		return nil
	}

	req.Header.Set("Referer", "http://blog.instapaper.com/post/152600596211")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("accept-language", `en-US,en;q=0.5`)
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Cookie", i.cookie)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Could not send get IDs request:", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return nil
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Println("get IDs request not 200")
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return nil
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("cannot read get IDs content", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return nil
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
doRequest:
	req, err := http.NewRequest("POST", "https://www.instapaper.com/delete_articles", strings.NewReader(fmt.Sprintf("[%d]", id)))
	if err != nil {
		fmt.Println("Could not parse remove link request:", err)
		return
	}

	retry := 0

	req.Header.Set("Referer", "http://www.instapaper.com/u")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("accept-language", `en-US,en;q=0.5`)
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Cookie", i.cookie)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Could not send remove link request:", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Println("remove link request not 200")
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("cannot read remove link content", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}
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

	retry := 0
doRequest:
	req, err := http.NewRequest("POST", "https://www.instapaper.com/user/kindle_send_now", strings.NewReader(postValues.Encode()))
	if err != nil {
		fmt.Println("Could not parse add link request:", err)
		return
	}

	req.Header.Set("Referer", "http://www.instapaper.com/u")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("accept-language", `en-US,en;q=0.5`)
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Cookie", i.cookie)

	resp, err := noRedirectClient.Do(req)
	if err != nil {
		fmt.Println("Could not send add link request:", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != 302 {
		fmt.Println("add link request not 302", resp.StatusCode)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("cannot read add link content", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}
}

func (i *Instapaper) GetFormKey() {
	retry := 0
doRequest:
	req, err := http.NewRequest("GET", "https://www.instapaper.com/user", nil)
	if err != nil {
		fmt.Println("Could not parse get form_key request:", err)
		return
	}

	req.Header.Set("Referer", "https://www.instapaper.com/u")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:45.0) Gecko/20100101 Firefox/45.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("accept-language", `en-US,en;q=0.5`)
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Cookie", i.cookie)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Could not send get form_key request:", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Println("get form_key request not 200")
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("cannot read get form_key content", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return
	}
	r := regexp.MustCompile(`<input id="form_key" name="form_key" type="hidden" value="([0-9a-zA-Z]+)" />`)

	ss := r.FindAllSubmatch(content, -1)
	i.formKey = string(ss[0][1])
}
