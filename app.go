package main

import (
	"fmt"
	"log"
	"regexp"

	"github.com/boltdb/bolt"
	"github.com/gocolly/colly"
)

var url string = "https://en.wikipedia.org/wiki/Coronavirus_disease_2019"
var allowedDomain string = "en.wikipedia.org"

func main() {
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	c := colly.NewCollector(
		colly.AllowedDomains(allowedDomain),
		colly.MaxDepth(2),
		colly.URLFilters(
			regexp.MustCompile("https://en.wikipedia\\.org/wiki/"),
		),
	)

	// Find and visit all links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))

	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.Visit(url)
}
