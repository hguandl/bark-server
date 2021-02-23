package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"time"

	bolt "go.etcd.io/bbolt"
)

func main() {
	dataDir := flag.String("d", "/data", "data dir")
	flag.Parse()
	boltDB, err := bolt.Open(filepath.Join(*dataDir, "bark.db"), 0600, nil)
	if err != nil {
		panic(err)
	}

	tokens := make(map[string]interface{})

	err = boltDB.View(func(t *bolt.Tx) error {
		b := t.Bucket([]byte("device"))

		c := b.Cursor()
		for k, deviceToken := c.First(); k != nil; k, deviceToken = c.Next() {
			_, ok := tokens[string(deviceToken)]

			if !ok {
				tokens[string(deviceToken)] = nil
			} else {
				fmt.Printf("Device token \"%v\" duplicated.\n", string(deviceToken))
			}
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	boltDB.Close()

	newDB, err := bolt.Open(filepath.Join(*dataDir, "bark-new.db"), 0600, nil)
	if err != nil {
		panic(err)
	}

	err = newDB.Update(func(t *bolt.Tx) error {
		b, err := t.CreateBucketIfNotExists([]byte("device"))
		if err != nil {
			return err
		}

		for deviceToken, _ := range tokens {
			err = b.Put([]byte(deviceToken), []byte(time.Now().Local().String()))
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	newDB.Close()
}
