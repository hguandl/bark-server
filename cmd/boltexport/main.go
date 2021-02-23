package main

import (
	"flag"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

func main() {
	dataFile := flag.String("f", "/data/bark.db", "data file")
	flag.Parse()
	boltDB, err := bolt.Open(*dataFile, 0600, nil)
	if err != nil {
		panic(err)
	}

	err = boltDB.View(func(t *bolt.Tx) error {
		b := t.Bucket([]byte("device"))

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("\"%v\": \"%v\"\n", string(k), string(v))
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	boltDB.Close()
}
