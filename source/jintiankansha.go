package source

import "github.com/dfordsoft/dev2kindle/config"

func init() {
	config.RegisterInitializer(func() {
		if config.Data.JinTianKanShaEnabled && len(config.Data.JinTianKanShaColumns) > 0 {
			t := &jinTianKanSha{}
			config.RegisterSource(t.fetch)
		}
	})
}

type jinTianKanSha struct {
}

func (t *jinTianKanSha) fetch(link chan string) {
}
