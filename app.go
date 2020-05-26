package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
)

var url string = "https://en.wikipedia.org/wiki/Coronavirus_disease_2019"
var allowedDomain string = "en.wikipedia.org"
var numLinksFound int = 0

func crawlerSummary(start time.Time) {
	elapsed := time.Since(start)
	log.Printf("\nCrawler Summary\nTop-level URL\t%s\nTime taken\t%s\nFound\t%d\n", url, elapsed, numLinksFound)
}

type wikiNode struct {
	Title string
}

func createNode(db *bolt.DB, key string, title string) {

	var value bytes.Buffer
	encoder := gob.NewEncoder(&value)

	err := encoder.Encode(wikiNode{title})
	if err != nil {
		log.Fatal("encode error:", err)
		return
	}

	db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(url))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		err = b.Put([]byte(key), value.Bytes())
		return err
	})

}

func main() {
	defer crawlerSummary(time.Now())
	db, err := bolt.Open("wikiTree.bolt", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	c := colly.NewCollector(
		colly.AllowedDomains(allowedDomain),
		colly.Async(true),
		colly.MaxDepth(2),
		colly.URLFilters(
			regexp.MustCompile("https://en.wikipedia\\.org/wiki/"),
		),

		colly.DisallowedURLFilters(
			regexp.MustCompile("https://en.wikipedia\\.org/wiki/File\\:"),
			regexp.MustCompile("https://en.wikipedia\\.org/wiki/Template\\:"),
			regexp.MustCompile("https://en.wikipedia\\.org/wiki/Help\\:"),
			regexp.MustCompile("https://en.wikipedia\\.org/wiki/VideoWiki\\:"),
			regexp.MustCompile("https://en.wikipedia\\.org/wiki/Wikipedia\\:"),
			regexp.MustCompile("https://en.wikipedia\\.org/wiki/Special\\:"),
			regexp.MustCompile("https://en.wikipedia\\.org/wiki/Category\\:"),
			regexp.MustCompile("https://en.wikipedia\\.org/wiki/Template_talk\\:"),
			regexp.MustCompile("https://en.wikipedia\\.org/wiki/Portal\\:"),
			regexp.MustCompile("https://en.wikipedia\\.org/wiki/Talk\\:"),
		),
	)

	extensions.RandomUserAgent(c)

	// Find and visit all links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
	})

	c.OnHTML("title", func(e *colly.HTMLElement) {
		numLinksFound = numLinksFound + 1 //TODO hmmm, is this goroutine safe?
		url := e.Request.URL.String()
		title := strings.TrimSuffix(e.Text, " - Wikipedia")
		createNode(db, url, title)

		//fmt.Printf("URL: %s Title: %s\n", url, title)
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("ERROR Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	c.Visit(url)
	c.Wait()
}
