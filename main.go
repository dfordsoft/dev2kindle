package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	client             *http.Client
	noRedirectClient   *http.Client
	i                  *Instapaper
	kindleMailbox      string
	instapaperUsername string
	instapaperPassword string
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

func openDatabase() (db *sql.DB) {
	tableCreated := true
	if existed, _ := isFileExists("dev2kindle.db"); existed == false {
		tableCreated = false
	}

	db, err := sql.Open("sqlite3", "dev2kindle.db")
	if err != nil {
		log.Fatal(err)
		return nil
	}

	if !tableCreated {
		sqlStmt := `create table links (id integer not null primary key, url text);`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			log.Printf("%q: %s\n", err, sqlStmt)
			return nil
		}
	}

	return db
}

func existsInDatabase(u string, db *sql.DB) bool {
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

func insertIntoDatabase(u string, db *sql.DB) bool {
	_, err := db.Exec(fmt.Sprintf("insert into foo(url) values('%s')", u))
	if err != nil {
		log.Fatal(err)
		return false
	}
	return true
}

func collectLink(u string) {
	db := openDatabase()
	if existsInDatabase(u, db) {
		// just drop it
		return
	}
	// if not exists in sqlite, then add to instapaer and insert into sqlite
	i.AddUrl(u)
	insertIntoDatabase(u, db)
}

func main() {
	fmt.Println("push articles for developers to kindle")
	flag.StringVar(&kindleMailbox, "kindle", "", "kindle mailbox")
	flag.StringVar(&instapaperUsername, "username", "", "instapaper username")
	flag.StringVar(&instapaperPassword, "password", "", "instapaper password")
	flag.Parse()

	if len(kindleMailbox) == 0 || len(instapaperPassword) == 0 || len(instapaperUsername) == 0 {
		fmt.Println("missing kindle mailbox or instapaer username/password")
		flag.Usage()
		return
	}

	client = &http.Client{
		Timeout: 30 * time.Second,
	}
	noRedirectClient = &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	i = &Instapaper{
		Username: instapaperUsername,
		Password: instapaperPassword,
	}
	i.Login()
	i.GetFormKey()

	link := make(chan string, 10)

	go func() {
		addLinkCount := 0
		for {
			select {
			case u := <-link:
				collectLink(u)
				addLinkCount++
				if addLinkCount > 50 {
					i.PushToKindle()
					time.Sleep(30 * time.Minute) // remove after all links are pushed to kindle
					i.RemoveAllLinks()
					addLinkCount = 0
				}
			}
		}
	}()

	hourTicker := time.NewTicker(60 * time.Minute)
	defer func() {
		hourTicker.Stop()
	}()
	for {
		select {
		case <-hourTicker.C:
			go fetchToutiao(link)
			go fetchXitu(link)
			go fetchGank(link)
			go fetchGeekCSDN(link)
			go fetchIwgc(link)
			go fetchSegmentFault(link)
		}
	}
}
