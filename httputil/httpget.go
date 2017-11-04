package httputil

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func HttpGet(u string) (content []byte) {
	headers := make(map[string]string)
	return HttpGetCustomHeader(u, headers)
}

func HttpGetCustomHeader(u string, headers map[string]string) (content []byte) {
	retry := 0
doRequest:
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		log.Println("Could not parse get request:", err, u)
		return nil
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Could not send get request:", err, u)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return nil
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Println("get request not 200", u)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return nil
	}
	content, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("cannot read get content", err, u)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
	}
	return content
}
