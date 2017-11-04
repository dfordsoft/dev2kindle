package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	flag "github.com/ogier/pflag"
)

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
	client                *http.Client
	noRedirectClient      *http.Client
	instapaper            *Instapaper
	kindleMailbox         string
	instapaperUsername    string
	instapaperPassword    string
	db                    *sql.DB
	linkCountInInstapaper int
	config                Config
)

func isFileExists(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err == nil {
		if stat.Mode()&os.ModeType == 0 {
			return true, nil
		}
		return false, errors.New(path + " exists but is not regular file")
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func openDatabase() {
	tableCreated := true
	if existed, _ := isFileExists("dev2kindle.db"); existed == false {
		tableCreated = false
	}
	var err error
	db, err = sql.Open("sqlite3", "dev2kindle.db")
	if err != nil {
		log.Fatal("opening sqlite3 database failed", err)
		return
	}

	if !tableCreated {
		sqlStmt := `create table links (id integer not null primary key, url text, instapaper bool);`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			log.Println("creating table failed", err)
		}
	}
}

func existsInDatabase(u string) bool {
	// query from sqlite
	rows, err := db.Query(fmt.Sprintf("select id from links where url = '%s'", u))
	if err != nil {
		log.Println("querying from sqlite to check existing failed", err)
		return false
	}
	defer rows.Close()
	for rows.Next() {
		return true
	}
	err = rows.Err()
	if err != nil {
		log.Println("reading rows to check existing failed", err)
		return false
	}

	return false
}

func insertIntoDatabase(u string) bool {
	_, err := db.Exec(fmt.Sprintf("insert into links(url, instapaper) values('%s', 'false')", u))
	if err != nil {
		log.Println("can't insert into database", err)
		return false
	}
	return true
}

func addLinksToInstapaper() {
	rows, err := db.Query("select id,url from links where instapaper='false' order by id limit 50;")
	if err != nil {
		log.Println("querying from database failed", err)
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		var u string
		err = rows.Scan(&id, &u)
		if err != nil {
			log.Println("scanning rows failed", err)
			continue
		}
		ids = append(ids, id)
		// add to instapaer
		instapaper.AddUrl(u)
		linkCountInInstapaper++
		if linkCountInInstapaper >= 50 {
			break
		}
	}
	err = rows.Err()
	if err != nil {
		log.Println("reading rows failed", err)
	}
	for _, id := range ids {
		db.Exec("update links set instapaper='true' where id=?", id)
	}
}

func collectLink(u string) {
	if existsInDatabase(u) {
		// just drop it
		return
	}
	// if not exists in sqlite, then insert into sqlite
	insertIntoDatabase(u)
}

func formatURL(theURL *url.URL) (u string) {
	if theURL.Host == "mp.weixin.qq.com" {
		query := theURL.Query().Encode()
		queries := strings.Split(query, "&")
		var newQueries []string
		needs := map[string]bool{
			"__biz": true,
			"sn":    true,
			"mid":   true,
			"idx":   true,
		}
		for _, q := range queries {
			qq := strings.Split(q, "=")
			if len(qq) == 2 {
				if _, ok := needs[qq[0]]; ok {
					newQueries = append(newQueries, q)
				}
			}
		}
		u = fmt.Sprintf("%s://%s%s?%s",
			theURL.Scheme, theURL.Host, theURL.Path, strings.Join(newQueries, "&"))
	} else {
		blacklist := []string{
			"weekly.manong.io",
			"github.com",
			"passport.weibo.com",
		}
		inBlacklist := false
		for _, b := range blacklist {
			if b == theURL.Host {
				inBlacklist = true
				return
			}
		}
		if inBlacklist {
			return
		}
		query := theURL.Query().Encode()
		queries := strings.Split(query, "&")
		var newQueries []string
		for _, q := range queries {
			qq := strings.Split(q, "=")
			if len(qq) == 2 && qq[1] == "toutiao.io" {
				continue
			}
			newQueries = append(newQueries, q)
		}
		if len(newQueries) == 0 {
			u = fmt.Sprintf("%s://%s%s", theURL.Scheme, theURL.Host, theURL.Path)
		} else {
			u = fmt.Sprintf("%s://%s%s?%s",
				theURL.Scheme, theURL.Host, theURL.Path, strings.Join(newQueries, "&"))
		}
	}
	return
}

func pushLinksFromInstapaperToKindle() {
	if linkCountInInstapaper >= 50 {
		// remove old links
		instapaper.RemoveAllLinks()
		linkCountInInstapaper = 0
	}
	// add new links
	addLinksToInstapaper()
	if linkCountInInstapaper >= 50 {
		// push to kindle
		instapaper.PushToKindle()
	}
}

func main() {
	quitAfterPushed := false
	clearInstapaper := false
	pushToKindle := false
	configPath := "config.json"
	flag.StringVar(&configPath, "config", "config.json", "specify the config file path")
	flag.StringVar(&kindleMailbox, "kindle", "", "kindle mailbox")
	flag.StringVar(&instapaperUsername, "username", "", "instapaper username")
	flag.StringVar(&instapaperPassword, "password", "", "instapaper password")
	flag.BoolVar(&quitAfterPushed, "quitAfterPushed", false, "quit after pushed")
	flag.BoolVar(&clearInstapaper, "clearInstapaper", false, "clear instapaper article list")
	flag.BoolVar(&pushToKindle, "pushToKindle", false, "push articles in instapaer to kindle now")
	flag.Parse()

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

	if err = json.Unmarshal(configcontent, &config); err != nil {
		log.Fatal("parsing ", configPath, " failed")
		return
	}

	if len(kindleMailbox) == 0 && len(config.Kindle) != 0 {
		kindleMailbox = config.Kindle
	}
	if len(instapaperUsername) == 0 && len(config.Username) != 0 {
		instapaperUsername = config.Username
	}
	if len(instapaperPassword) == 0 && len(config.Password) != 0 {
		instapaperPassword = config.Password
	}

	if len(kindleMailbox) == 0 || len(instapaperPassword) == 0 || len(instapaperUsername) == 0 {
		log.Println("missing kindle mailbox or instapaer username/password")
		flag.Usage()
		return
	}
	if len(config.ToutiaoSubjects) == 0 {
		config.ToutiaoEnabled = false
	}
	if len(config.RSSFeeds) == 0 {
		config.RSSEnabled = false
	}
	if len(config.JinTianKanShaColumns) == 0 {
		config.JinTianKanShaEnabled = false
	}
	if len(config.WeixinYiduIDs) == 0 {
		config.WeixinYiduEnabled = false
	}

	log.Println("kindle mailbox:", kindleMailbox)
	log.Println("Instapaper username:", instapaperUsername)
	log.Println("Instapaper password:", instapaperPassword)
	log.Println("Quit after pushed:", quitAfterPushed)
	log.Println("Clear Instapaper articles:", clearInstapaper)
	log.Println("Push To Kindle:", pushToKindle)

	client = &http.Client{
		Timeout: 30 * time.Second,
	}
	noRedirectClient = &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	instapaper = &Instapaper{
		Username: instapaperUsername,
		Password: instapaperPassword,
	}
	instapaper.Login()
	instapaper.GetFormKey()

	if pushToKindle {
		instapaper.PushToKindle()
		return
	}

	if clearInstapaper {
		instapaper.RemoveAllLinks()
		return
	}

	openDatabase()

	link := make(chan string, 10)
	quit := make(chan bool)
	go func() {
		for {
			select {
			case <-quit:
				return
			case u := <-link:
				if theURL, e := url.Parse(u); e == nil && theURL.Host != "" {
					if u = formatURL(theURL); u != "" {
						collectLink(u)
					}
				}
			}
		}
	}()

	halfAnHourTicker := time.NewTicker(30 * time.Minute)
	hourTicker := time.NewTicker(60 * time.Minute)
	defer func() {
		hourTicker.Stop()
	}()

	log.Println("start fetch articles...")
	go pushLinksFromInstapaperToKindle()
	var t *Toutiao
	if config.ToutiaoEnabled {
		t = &Toutiao{}
		go t.Fetch(link)
	}
	var x *Xitu
	if config.GoldXituEnabled {
		x = &Xitu{}
		go x.Fetch(link)
	}
	var c *GeekCSDN
	if config.GeekCSDNEnabled {
		c = &GeekCSDN{}
		go c.Fetch(link)
	}
	var s *SegmentFault
	if config.SegmentFaultEnabled {
		s = &SegmentFault{}
		go s.Fetch(link)
	}
	var r *RSSFeed
	if config.RSSEnabled {
		r = &RSSFeed{}
		go r.Fetch(link)
	}
	var j *JinTianKanSha
	if config.JinTianKanShaEnabled {
		j = &JinTianKanSha{}
		go j.Fetch(link)
	}
	var w *WeixinYidu
	if config.WeixinYiduEnabled {
		w = &WeixinYidu{}
		go w.Fetch(link)
	}

	if quitAfterPushed {
		quit <- true
		return
	}

	go func() {
		sigchan := make(chan os.Signal, 10)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan
		log.Println("Program killed !")

		if db != nil {
			db.Close()
		}

		os.Exit(0)
	}()

	for {
		select {
		case <-halfAnHourTicker.C:
			pushLinksFromInstapaperToKindle()
		case <-hourTicker.C:
			if config.ToutiaoEnabled {
				go t.Fetch(link)
			}
			if config.GoldXituEnabled {
				go x.Fetch(link)
			}
			if config.GeekCSDNEnabled {
				go c.Fetch(link)
			}
			if config.SegmentFaultEnabled {
				go s.Fetch(link)
			}
			if config.RSSEnabled {
				go r.Fetch(link)
			}
			if config.WeixinYiduEnabled {
				go w.Fetch(link)
			}
			if config.JinTianKanShaEnabled {
				go j.Fetch(link)
			}
		}
	}
}
