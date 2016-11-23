package main

import (
	"fmt"
	"regexp"
)

var (
	crawlListIDs = []int{
		4868,  // 架构师之路
		2317,  // 聊聊架构
		1929,  // 架构师
		1922,  // 高可用架构
		4956,  // 移动开发前线
		169,   // CPP开发者
		488,   // iOS开发
		5187,  // IOS开发精选
		6737,  // Android开发中文站
		6990,  // Mac开发
		1199,  // iOS大全
		5054,  // iOS Developer
		3567,  // AndroidDeveloper
		405,   // 程序员的自我修养
		790,   // 郭志敏的程序员书屋
		6128,  // 程序视界
		10712, // 一个程序员的日常
		8947,  // 程序员进阶
		1261,  // 高效运维
		18,    // InfoQ
		5213,  // ArchSummit技术关注
	}
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
	for _, id := range crawlListIDs {
		u := fmt.Sprintf("http://iwgc.cn/list/%d", id)
		i.fetchArticles(link, u)
	}
}
