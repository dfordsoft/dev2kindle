package source

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/dfordsoft/dev2kindle/config"
	"github.com/dfordsoft/dev2kindle/httputil"
)

func init() {
	config.RegisterInitializer(func() {
		if config.Data.ToutiaoEnabled && len(config.Data.ToutiaoSubjects) > 0 {
			t := &toutiao{}
			config.RegisterSource(t.fetch)
		}
	})
}

type toutiao struct {
}

func (t *toutiao) resolveFinalURL(u string) string {
	retry := 0
doResolve:
	resp, err := http.Get(u)
	if err != nil {
		log.Printf("resolving url %s failed => %v\n", u, err)
		// try to extract hyperlink from err.Error()
		regex := regexp.MustCompile(`https?://[^:]+`)
		list := regex.FindAllString(err.Error(), -1)
		if len(list) > 0 {
			return list[0]
		}
		if retry < 3 {
			retry++
			time.Sleep(3 * time.Second)
			goto doResolve
		} else {
			return ""
		}
	}
	defer resp.Body.Close()
	finalURL := resp.Request.URL.String()
	return finalURL
}

func (t *toutiao) fetch(link chan string) {
	now := time.Now()
	u := fmt.Sprintf("https://toutiao.io/prev/%4.4d-%2.2d-%2.2d", now.Year(), now.Month(), now.Day())
	t.fetchArticles(link, u)

	for _, id := range config.Data.ToutiaoSubjects {
		u := fmt.Sprintf("https://toutiao.io/subjects/%d?f=new", id)
		t.fetchArticles(link, u)
	}
}

func (t *toutiao) fetchArticles(link chan string, u string) {
	content := httputil.HttpGet(u)

	if len(content) == 0 {
		return
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
