# transrpc [![GoDoc][godoc]][godoc-link] [![Build Status][travis-ci]][travis-ci-link]

Package `transrpc` provides a Go idiomatic [Transmission RPC][transmission-spec]
client.

[transmission-spec]: https://github.com/transmission/transmission/blob/master/extras/rpc-spec.txt

[godoc]: https://godoc.org/github.com/kenshaw/transrpc?status.svg (GoDoc)
[godoc-link]: https://godoc.org/github.com/kenshaw/transrpc
[travis-ci]: https://travis-ci.org/kenshaw/transrpc.svg?branch=master (Travis CI)
[travis-ci-link]: https://travis-ci.org/kenshaw/transrpc

## Example

```go
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
```
