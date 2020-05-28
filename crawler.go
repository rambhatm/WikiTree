package main

import (
	"context"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	Mongo struct {
		Client *mongo.Client
	}
}

//Client config
//var mongodbURI = os.Getenv("MONGODB_URI") + "?retryWrites=false"
var ClientOptions = options.Client().ApplyURI(MONGODBURI)

func (c *crawler) init(url string, allowedDomain string, maxDepth int) {
	c.URL = url
	c.AllowedDomain = allowedDomain
	c.MaxDepth = maxDepth

	//init mongo client for the crawler instance
	var err error
	c.Mongo.Client, err = mongo.Connect(context.TODO(), ClientOptions)
	if err != nil {
		log.Fatal("cannot connect to mongodb")
	}
}

func (c *crawler) end(start time.Time) {
	elapsed := time.Since(start)
	c.Mongo.Client.Disconnect(context.TODO())
	c.Stats.TotalLinks = c.Stats.CrawledLinks + c.Stats.ErrorLinks
	log.Printf("\nCrawler Summary\nTop-level URL\t%s\nTime taken\t%s\nStats\t%+v\n", url, elapsed, c.Stats)
}

func Crawl(url string, allowedDomain string, maxDepth int) {
	//Initialize a crawler
	var crawler crawler
	crawler.init(url, allowedDomain, maxDepth)
	defer crawler.end(time.Now())

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
		doc := WikiDoc{
			strings.TrimSuffix(e.Text, " - Wikipedia"),
			e.Request.URL.String(),
			e.Request.Depth,
		}
		doc.InsertDB(crawler.Mongo.Client)
	})

	c.OnError(func(r *colly.Response, err error) {
		//fmt.Println("ERROR Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
		crawler.Stats.ErrorLinks++
	})

	c.Visit(url)
	c.Wait()
}
