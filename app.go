package main

import (
	"context"
	"log"
	"time"

	//"net/http"
	//_ "net/http/pprof"

	"go.mongodb.org/mongo-driver/mongo"
)

const (
	MONGODBURI   = "mongodb://localhost:27017/?retryWrites=false"
	WIKIDB       = "w2"
	CRAWLRESULTS = "crawlResultCollection"
)

type WikiDoc struct {
	Title           string
	Url             string
	SeedUrl         string
	DistanceFromUrl int
}

var url string = "https://en.wikipedia.org/wiki/Coronavirus_disease_2019"

func MongoInsertDoc(c *mongo.Client, doc WikiDoc) {
	//ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	//client, err := mongo.Connect(ctx, ClientOptions)
	//if err != nil {
	//	log.Fatal("cannot connect to mongodb")
	//}
	//defer client.Disconnect(context.TODO())
	crawlColly := c.Database(WIKIDB).Collection(CRAWLRESULTS)

	// Declare Context type object for managing multiple API requests
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	_, err := crawlColly.InsertOne(ctx, doc)
	if err != nil {
		log.Fatal(err)
	}

}
func main() {
	/*
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()*/
	Crawl(1, url, "en.wikipedia.org", 3)
}
