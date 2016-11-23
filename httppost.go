package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

func httpPostBasicAuth(u string, data string, username string, password string) (content []byte) {
	retry := 0
doRequest:
	req, err := http.NewRequest("POST", u, strings.NewReader(data))
	if err != nil {
		log.Println("Could not parse post request:", err)
		return nil
	}

	req.SetBasicAuth(username, password)

	var resp *http.Response
	resp, err = client.Do(req)
	if err != nil {
		log.Println("Could not send post request:", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return nil
	}

	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Println("post request not 200:", resp.StatusCode)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return nil
	}
	content, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("cannot read post content", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return nil
	}
	return content
}

func httpPostCustomHeader(u string, data string, headers map[string]string, noRedirect bool) (content []byte) {
	retry := 0
doRequest:
	req, err := http.NewRequest("POST", u, strings.NewReader(data))
	if err != nil {
		log.Println("Could not parse post request:", err)
		return nil
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	var resp *http.Response
	if noRedirect {
		resp, err = noRedirectClient.Do(req)
	} else {
		resp, err = client.Do(req)
	}
	if err != nil {
		log.Println("Could not send post request:", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return nil
	}

	defer resp.Body.Close()
	if noRedirect {
		if resp.StatusCode < 300 || resp.StatusCode >= 400 {
			log.Println("post request not 302:", resp.StatusCode)
			retry++
			if retry < 3 {
				time.Sleep(3 * time.Second)
				goto doRequest
			}
			return nil
		}
	} else {
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			log.Println("post request not 200:", resp.StatusCode)
			retry++
			if retry < 3 {
				time.Sleep(3 * time.Second)
				goto doRequest
			}
			return nil
		}
	}
	content, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("cannot read post content", err)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return nil
	}
	return content
}
