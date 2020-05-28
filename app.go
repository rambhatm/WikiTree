package main

import (
	"context"
	"log"

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
	depth int
}

var url string = "https://en.wikipedia.org/wiki/Coronavirus_disease_2019"

func (doc WikiDoc) InsertDB(client *mongo.Client) {
	crawlColly := client.Database(WIKIDB).Collection(CRAWLRESULTS)

	//Find a max of 5 results for filter
	_, err := crawlColly.InsertOne(context.TODO(), doc)
	if err != nil {
		log.Fatal(err)
	}

}
func main() {
	Crawl(url, "en.wikipedia.org", 2)
}
