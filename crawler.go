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
	C       *mongo.Client
	Scraper *colly.Collector
}

//Client config
//var mongodbURI = os.Getenv("MONGODB_URI") + "?retryWrites=false"
var ClientOptions = options.Client().ApplyURI(MONGODBURI)
var Collector = colly.NewCollector(
	colly.AllowedDomains("en.wikipedia.org"),
	colly.Async(true),
	colly.MaxDepth(1),
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
	var err error
	c.C, err = mongo.Connect(context.TODO(), ClientOptions)
	if err != nil {
		log.Fatal("cannot connect to mongodb")
	}
	//defer client.Disconnect(context.TODO())
	model := mongo.IndexModel{
		Keys: bson.M{
			"title": 1,
		}, Options: nil,
	}
	crawlColly := c.C.Database(WIKIDB).Collection(CRAWLRESULTS)
	crawlColly.Indexes().CreateOne(context.TODO(), model)

	c.Scraper = Collector.Clone()
}

func (c *crawler) end() {
	c.C.Disconnect(context.TODO())
	defer c.summary()
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

func Crawl(id int, url string) {
	//Initialize a crawler
	var crawler crawler
	crawler.init(id, url)
	defer crawler.end()

	var doc WikiDoc

	extensions.RandomUserAgent(crawler.Scrapper)
	extensions.Referer(c)
	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 32, RandomDelay: 25 * time.Millisecond})

	// Find and visit all links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		doc.PageLinkspageLinks = append(doc.PageLinks, e.Attr("href"))
		//e.Request.Visit(e.Attr("href"))
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
		go MongoInsertDoc(crawler.C, doc)
	})

	c.OnError(func(r *colly.Response, err error) {
		//fmt.Println("ERROR Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
		crawler.Stats.ErrorLinks[r.Request.Depth-1]++
	})

	var pageLinkMap = map[string]string
	for link = delete()
	c.Visit(url)
	c.Wait()

}
