// _examples/qbtweb.go

// Example qbtweb client use.
package main

import (
	"context"
	"log"

	"github.com/kenshaw/torctl/qbtweb"
)

func main() {
	cl := qbtweb.NewClient(
		qbtweb.WithHost("admin:adminadmin@localhost:8080"),
		qbtweb.WithLogf(log.Printf),
	)

	torrents, err := cl.TorrentsInfo(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	for i, torrent := range torrents {
		log.Printf("> ID: %d Name: %s Hash: %s", i, torrent.Name, torrent.Hash)
	}
}
