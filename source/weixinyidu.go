package source

import "github.com/dfordsoft/dev2kindle/config"

func init() {
	config.RegisterInitializer(func() {
		if config.Data.WeixinYiduEnabled && len(config.Data.WeixinYiduIDs) > 0 {
			t := &weixinYidu{}
			config.RegisterSource(t.fetch)
		}
	})
}

type weixinYidu struct {
}

func (t *weixinYidu) fetch(link chan string) {
}
