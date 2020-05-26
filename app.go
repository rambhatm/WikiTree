package main

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gocolly/colly"
)

var url string = "https://en.wikipedia.org/wiki/Coronavirus_disease_2019"
var allowedDomain string = "en.wikipedia.org"
var numLinksFound int = 0

func crawlerSummary(start time.Time) {
	elapsed := time.Since(start)
	log.Printf("\nCrawler Summary\nTop-level URL\t%s\nTime taken\t%s\nFound\t%d\n", url, elapsed, numLinksFound)
}

func main() {
	defer crawlerSummary(time.Now())
	db, err := bolt.Open("my.db", 0600, nil)
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

	// Find and visit all links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
	})

	c.OnHTML("title", func(e *colly.HTMLElement) {
		numLinksFound = numLinksFound + 1 //TODO hmmm, is this goroutine safe?
		fmt.Printf("URL: %s Title: %s\n", e.Request.URL, e.Text)
	})

	c.Visit(url)
	c.Wait()
}
