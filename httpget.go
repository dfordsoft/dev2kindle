package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func httpGet(u string) (content []byte) {
	retry := 0
doRequest:
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		fmt.Println("Could not parse get request:", err, u)
		return nil
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Could not send get request:", err, u)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return nil
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Println("get request not 200", u)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
		return nil
	}
	content, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("cannot read get content", err, u)
		retry++
		if retry < 3 {
			time.Sleep(3 * time.Second)
			goto doRequest
		}
	}
	return content
}
