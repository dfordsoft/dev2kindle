package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	client             *http.Client
	noRedirectClient   *http.Client
	instapaper         *Instapaper
	kindleMailbox      string
	instapaperUsername string
	instapaperPassword string
	db                 *sql.DB
)

func init() {
}

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
		log.Fatal(err)
		return
	}

	if !tableCreated {
		sqlStmt := `create table links (id integer not null primary key, url text, instapaper bool);`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			log.Printf("%q: %s\n", err, sqlStmt)
		}
	}
}

func existsInDatabase(u string) bool {
	// query from sqlite
	rows, err := db.Query(fmt.Sprintf("select id from links where url = '%s'", u))
	if err != nil {
		log.Fatal(err)
		return false
	}
	defer rows.Close()
	for rows.Next() {
		return true
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
		return false
	}

	return false
}

func insertIntoDatabase(u string) bool {
	_, err := db.Exec(fmt.Sprintf("insert into links(url, instapaper) values('%s', 'false')", u))
	if err != nil {
		log.Fatal(err)
		return false
	}
	return true
}

func addLinksToInstapaper() {
	rows, err := db.Query("select id,url from links where instapaper='false' order by id limit 50;")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		var u string
		err = rows.Scan(&id, &u)
		if err != nil {
			log.Fatal(err)
		}
		ids = append(ids, id)
		// add to instapaer
		instapaper.AddUrl(u)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
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
	// remove old links
	instapaper.RemoveAllLinks()
	// add new links
	addLinksToInstapaper()
	// push to kindle
	instapaper.PushToKindle()
}

func main() {
	quitAfterPushed := false
	clearInstapaper := false
	pushToKindle := false
	flag.StringVar(&kindleMailbox, "kindle", "", "kindle mailbox")
	flag.StringVar(&instapaperUsername, "username", "", "instapaper username")
	flag.StringVar(&instapaperPassword, "password", "", "instapaper password")
	flag.BoolVar(&quitAfterPushed, "quitAfterPushed", false, "quit after pushed")
	flag.BoolVar(&clearInstapaper, "clearInstapaper", false, "clear instapaper article list")
	flag.BoolVar(&pushToKindle, "pushToKindle", false, "push articles in instapaer to kindle now")
	flag.Parse()

	if len(kindleMailbox) == 0 || len(instapaperPassword) == 0 || len(instapaperUsername) == 0 {
		fmt.Println("missing kindle mailbox or instapaer username/password")
		flag.Usage()
		return
	}

	fmt.Println("kindle mailbox:", kindleMailbox)
	fmt.Println("Instapaper username:", instapaperUsername)
	fmt.Println("Instapaper password:", instapaperPassword)
	fmt.Println("Quit after pushed:", quitAfterPushed)
	fmt.Println("Clear Instapaper articles:", clearInstapaper)
	fmt.Println("Push To Kindle:", pushToKindle)

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
	t := &Toutiao{}
	i := &Iwgc{}
	x := &Xitu{}
	g := &Gank{}
	c := &GeekCSDN{}
	s := &SegmentFault{}

	fmt.Println("start fetch articles...")
	go pushLinksFromInstapaperToKindle()
	go t.Fetch(link)
	go i.Fetch(link)
	go x.Fetch(link)
	go g.Fetch(link)
	go c.Fetch(link)
	go s.Fetch(link)

	if quitAfterPushed {
		quit <- true
		return
	}
	for {
		select {
		case <-halfAnHourTicker.C:
			pushLinksFromInstapaperToKindle()
		case <-hourTicker.C:
			go t.Fetch(link)
			go i.Fetch(link)
			go x.Fetch(link)
			go g.Fetch(link)
			go c.Fetch(link)
			go s.Fetch(link)
		}
	}
}
