package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

type WikiDoc struct {
	Title string
}

func NewDoc(db *bolt.DB, key string, title string) {
	var value bytes.Buffer
	encoder := gob.NewEncoder(&value)

	//encode
	err := encoder.Encode(wikiNode{title})
	if err != nil {
		log.Fatal("encode error:", err)
		return
	}

	//commit
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
	crawler.Crawl("https://en.wikipedia.org/wiki/Coronavirus_disease_2019", "en.wikipedia.org", 2)
}
