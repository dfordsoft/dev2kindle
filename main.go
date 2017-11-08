package main

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/dfordsoft/dev2kindle/config"
	flag "github.com/ogier/pflag"
)

var (
	kindleMailbox         string
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

func collectLink(u string) {

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

	// initialize sources
	for _, i := range config.Initializer {
		i()
	}

	log.Println("start fetch articles...")

	for _, s := range config.Sources {
		go s(link)
	}

	go func() {
		sigchan := make(chan os.Signal, 10)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan
		log.Println("Program killed !")

		os.Exit(0)
	}()

	for {
		select {
		case <-hourTicker.C:
			for _, s := range config.Sources {
				go s(link)
			}
		}
	}
}
