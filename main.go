package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"
)

var (
	client             *http.Client
	noRedirectClient   *http.Client
	kindleMailbox      string
	instapaperUsername string
	instapaperPassword string
)

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

	i := &Instapaper{
		Username: instapaperUsername,
		Password: instapaperPassword,
	}
	i.Login()
	i.GetFormKey()

	hourTicker := time.NewTicker(1 * time.Hour)
	defer func() {
		hourTicker.Stop()
	}()
	for {
		select {
		case <-hourTicker.C:
		}
	}
}
