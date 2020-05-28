package main

import (
	"context"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type crawler struct {
	ID            int
	URL           string
	MaxDepth      int
	AllowedDomain string
	Stats         struct {
		StartTime    time.Time
		CrawledLinks [16]int
		ErrorLinks   [16]int
		//TotalLinks   int
	}
}

//Client config
//var mongodbURI = os.Getenv("MONGODB_URI") + "?retryWrites=false"
var ClientOptions = options.Client().ApplyURI(MONGODBURI)

func (c *crawler) init(id int, url string, allowedDomain string, maxDepth int) {
	c.Stats.StartTime = time.Now()
	c.ID = id
	c.URL = url
	c.AllowedDomain = allowedDomain
	c.MaxDepth = maxDepth

	//init stats
	for i := 0; i < maxDepth; i++ {
		c.Stats.CrawledLinks[i] = 0
		c.Stats.ErrorLinks[i] = 0
	}

	//init mongo client for the crawler instance
	client, err := mongo.Connect(context.TODO(), ClientOptions)
	if err != nil {
		log.Fatal("cannot connect to mongodb")
	}
	defer client.Disconnect(context.TODO())
	model := mongo.IndexModel{
		Keys: bson.M{
			"title": 1,
		}, Options: nil,
	}
	crawlColly := client.Database(WIKIDB).Collection(CRAWLRESULTS)
	crawlColly.Indexes().CreateOne(context.TODO(), model)
}

func (c *crawler) end() {
	//defer c.summary()
}

func (c *crawler) summary() {
	elapsed := time.Since(c.Stats.StartTime)
	log.Printf("\ncrawler[%d] - seed: %s, time: %s", c.ID, c.URL, elapsed)
	total := 0
	for i := 0; i < c.MaxDepth; i++ {
		perLevelTotal := c.Stats.CrawledLinks[i] + c.Stats.ErrorLinks[i]
		total += perLevelTotal
		log.Printf("\n\tLevel[%d] - crawled: %d, error: %d, total: %d  ", i+1, c.Stats.CrawledLinks[i], c.Stats.ErrorLinks[i], perLevelTotal)
	}
	log.Printf("\nTotal links crawled: %d", total)
}

func Crawl(id int, url string, allowedDomain string, maxDepth int) {
	//Initialize a crawler
	var crawler crawler
	crawler.init(id, url, allowedDomain, maxDepth)
	defer crawler.end()

	c := colly.NewCollector(
		colly.AllowedDomains(allowedDomain),
		colly.Async(true),
		colly.MaxDepth(maxDepth),
		//colly.Debugger(&debug.LogDebugger{}),
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
		crawler.Stats.CrawledLinks[e.Request.Depth-1]++
		doc := WikiDoc{
			strings.TrimSuffix(e.Text, " - Wikipedia"),
			e.Request.URL.String(),
			crawler.URL,
			e.Request.Depth,
		}
		//doc.SeedUrl = "a"
		go MongoInsertDoc(doc)
	})

	c.OnError(func(r *colly.Response, err error) {
		//fmt.Println("ERROR Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
		crawler.Stats.ErrorLinks[r.Request.Depth-1]++
	})

	c.Visit(url)
	c.Wait()
}
