package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

// RegisterInitializer add source initializer
func RegisterInitializer(i func()) {
	Initializer = append(Initializer, i)
}

// RegisterSource add source action function
func RegisterSource(s func(chan string)) {
	Sources = append(Sources, s)
}

// Config stores all items from configuration file
type Config struct {
	Kindle               string   `json:"kindle"`
	Username             string   `json:"instapaper_username"`
	Password             string   `json:"instapaper_password"`
	SegmentFaultEnabled  bool     `json:"segmentfault_enabled"`
	GeekCSDNEnabled      bool     `json:"geekcsdn_enabled"`
	GoldXituEnabled      bool     `json:"goldxitu_enabled"`
	ToutiaoEnabled       bool     `json:"toutiaoio_enabled"`
	JinTianKanShaEnabled bool     `json:"jintiankansha_enabled"`
	WeixinYiduEnabled    bool     `json:"weixinyidu_enabled"`
	RSSEnabled           bool     `json:"rss_enabled"`
	ToutiaoSubjects      []int    `json:"toutiaoio_subjects"`
	JinTianKanShaColumns []string `json:"jintiankansha_columns"`
	WeixinYiduIDs        []string `json:"weixinyidu_ids"`
	RSSFeeds             []string `json:"rss_feeds"`
}

var (
	// Data stores all items from main configuration file
	Data Config
	// Initializer all source initializers
	Initializer []func()
	// Sources all source action functions
	Sources []func(chan string)
)

// LoadConfig loads main configuration file to Data variable as Config struct
func LoadConfig(configPath string) {
	fh, err := os.Open(configPath)
	if err != nil {
		log.Fatal("opening ", configPath, " failed")
		return
	}
	defer fh.Close()
	configcontent, err := ioutil.ReadAll(fh)
	if err != nil {
		log.Fatal("reading ", configPath, " failed")
		return
	}

	if err = json.Unmarshal(configcontent, &Data); err != nil {
		log.Fatal("parsing ", configPath, " failed")
		return
	}
}
