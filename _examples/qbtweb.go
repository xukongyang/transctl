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
		qbtweb.WithHost("user:pass@my-host:8080"),
	)

	torrents, err := cl.TorrentsInfo(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	for _, torrent := range torrents {
		log.Printf("> ID: %d Name: %s Hash: %s", torrent.ID, torrent.Name, torrent.HashString)
	}
}
