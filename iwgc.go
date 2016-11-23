package main

import (
	"fmt"
	"regexp"
)

type Iwgc struct {
}

func (i *Iwgc) resolveFinalURL(u string) (resolvedURL string) {
	content := httpGet(u)

	if len(content) == 0 {
		return
	}
	regex := regexp.MustCompile(`http://mp.weixin.qq.com[^']+`)
	list := regex.FindAll(content, -1)
	if len(list) > 0 {
		resolvedURL = string(list[0])
	}
	return
}

func (i *Iwgc) fetchArticles(link chan string, u string) {
	content := httpGet(u)

	if len(content) == 0 {
		return
	}
	regex := regexp.MustCompile(`/link/([0-9]+)`)
	list := regex.FindAllSubmatch(content, -1)
	for _, l := range list {
		lnk := fmt.Sprintf("http://iwgc.cn/link/%s", string(l[1]))
		resolvedURL := i.resolveFinalURL(lnk)
		if resolvedURL != "" {
			link <- resolvedURL
		}
	}
}

func (i *Iwgc) Fetch(link chan string) {
	for _, id := range config.IwgcLists {
		u := fmt.Sprintf("http://iwgc.cn/list/%d", id)
		i.fetchArticles(link, u)
	}
}
