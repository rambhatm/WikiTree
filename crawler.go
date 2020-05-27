package main

import (
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
)

type crawler struct {
	URL           string
	MaxDepth      int
	AllowedDomain string
	Stats         struct {
		CrawledLinks int
		ErrorLinks   int
		TotalLinks   int
	}
}

func (c *crawler) Summary(start time.Time) {
	elapsed := time.Since(start)
	c.Stats.TotalLinks = c.Stats.CrawledLinks + c.Stats.ErrorLinks
	log.Printf("\nCrawler Summary\nTop-level URL\t%s\nTime taken\t%s\nStats\t%+v\n", url, elapsed, c.Stats)
}

func Crawl(url string, allowedDomain string, maxDepth int) {
	//Initialize a crawler
	var crawler crawler
	crawler.URL = url
	crawler.AllowedDomain = allowedDomain
	crawler.MaxDepth = maxDepth
	defer crawler.Summary(time.Now())

	db, err := bolt.Open("wikiTree.bolt", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	c := colly.NewCollector(
		colly.AllowedDomains(allowedDomain),
		colly.Async(true),
		colly.MaxDepth(maxDepth),
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
	extensions.Referer(c)
	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 4})

	// Find and visit all links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
	})

	c.OnHTML("title", func(e *colly.HTMLElement) {
		crawler.Stats.CrawledLinks++
		url := e.Request.URL.String()
		title := strings.TrimSuffix(e.Text, " - Wikipedia")
		NewDoc(db, url, title)
	})

	c.OnError(func(r *colly.Response, err error) {
		//fmt.Println("ERROR Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
		crawler.Stats.ErrorLinks++
	})

	c.Visit(url)
	c.Wait()
}
