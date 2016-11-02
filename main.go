package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"
)

var (
	client             *http.Client
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
	client = &http.Client{
		Timeout: 30 * time.Second,
	}
}
