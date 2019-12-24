// _example/transrpc.go

// Example transrpc client use.
package main

import (
	"context"
	"log"

	"github.com/kenshaw/transctl/transrpc"
)

func main() {
	cl := transrpc.NewClient(
		transrpc.WithHost("user:pass@my-host:9091"),
	)

	res, err := cl.TorrentGet(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	for _, torrent := range res.Torrents {
		log.Printf("> ID: %d Name: %s Hash: %s", torrent.ID, torrent.Name, torrent.HashString)
	}
}
