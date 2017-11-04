package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/dfordsoft/dev2kindle/config"
	"github.com/dfordsoft/dev2kindle/source"
	_ "github.com/mattn/go-sqlite3"
	flag "github.com/ogier/pflag"
)

var (
	kindleMailbox         string
	db                    *sql.DB
	linkCountInInstapaper int
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

func main() {
	configPath := "config.json"
	flag.StringVar(&configPath, "config", "config.json", "specify the config file path")
	flag.StringVar(&kindleMailbox, "kindle", "", "kindle mailbox")
	flag.Parse()

	config.LoadConfig(configPath)

	if len(kindleMailbox) == 0 && len(config.Data.Kindle) != 0 {
		kindleMailbox = config.Data.Kindle
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

	hourTicker := time.NewTicker(60 * time.Minute)
	defer func() {
		hourTicker.Stop()
	}()

	log.Println("start fetch articles...")

	var t *source.Toutiao
	if config.Data.ToutiaoEnabled {
		t = &source.Toutiao{}
		go t.Fetch(link)
	}
	var x *source.Xitu
	if config.Data.GoldXituEnabled {
		x = &source.Xitu{}
		go x.Fetch(link)
	}
	var c *source.GeekCSDN
	if config.Data.GeekCSDNEnabled {
		c = &source.GeekCSDN{}
		go c.Fetch(link)
	}
	var s *source.SegmentFault
	if config.Data.SegmentFaultEnabled {
		s = &source.SegmentFault{}
		go s.Fetch(link)
	}
	var r *source.RSSFeed
	if config.Data.RSSEnabled {
		r = &source.RSSFeed{}
		go r.Fetch(link)
	}
	var j *source.JinTianKanSha
	if config.Data.JinTianKanShaEnabled {
		j = &source.JinTianKanSha{}
		go j.Fetch(link)
	}
	var w *source.WeixinYidu
	if config.Data.WeixinYiduEnabled {
		w = &source.WeixinYidu{}
		go w.Fetch(link)
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
		case <-hourTicker.C:
			if config.Data.ToutiaoEnabled {
				go t.Fetch(link)
			}
			if config.Data.GoldXituEnabled {
				go x.Fetch(link)
			}
			if config.Data.GeekCSDNEnabled {
				go c.Fetch(link)
			}
			if config.Data.SegmentFaultEnabled {
				go s.Fetch(link)
			}
			if config.Data.RSSEnabled {
				go r.Fetch(link)
			}
			if config.Data.WeixinYiduEnabled {
				go w.Fetch(link)
			}
			if config.Data.JinTianKanShaEnabled {
				go j.Fetch(link)
			}
		}
	}
}
