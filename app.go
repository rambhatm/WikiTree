package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

const (
	MONGODBURI   = "mongodb://localhost:27017/?retryWrites=false"
	WIKIDB       = "wikiDB"
	CRAWLRESULTS = "crawlResultCollection"
)

type WikiDoc struct {
	Title string
	Url   string
	Depth int
}

var url string = "https://en.wikipedia.org/wiki/Coronavirus_disease_2019"

func (doc WikiDoc) InsertDB(client *mongo.Client) {
	crawlColly := client.Database(WIKIDB).Collection(CRAWLRESULTS)

	// Declare Context type object for managing multiple API requests
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	_, err := crawlColly.InsertOne(ctx, doc)
	if err != nil {
		log.Fatal(err)
	}

}
func main() {
	Crawl(url, "en.wikipedia.org", 2)
}
