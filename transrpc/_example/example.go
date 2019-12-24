// _example/example.go
package main

import (
	"context"
	"log"

	"github.com/kenshaw/transrpc"
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
