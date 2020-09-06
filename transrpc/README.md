# transrpc [![GoDoc][godoc]][godoc-link]

Package `transrpc` provides a Go idiomatic [Transmission RPC][transmission-spec]
client, and is part of the [command-line tool `transctl`][transctl].

[godoc]: https://godoc.org/github.com/kenshaw/transctl/transrpc?status.svg (GoDoc)
[godoc-link]: https://godoc.org/github.com/kenshaw/transctl/transrpc

## Installing

Install in the usual [Go][go-project] fashion:

```sh
$ go get -u github.com/kenshaw/transctl/transrpc
```

## Using

`transrpc` can be used similar to the following:

```go
// _example/example.go
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
```

Additionally, a more complex example used to remove a completed [torrent upon
notification is available here][tcomplete].

[go-project]: https://golang.org/project
[transmission-spec]: https://github.com/transmission/transmission/blob/master/extras/rpc-spec.txt
[transctl]: https://github.com/kenshaw/transctl
[tcomplete]: https://github.com/kenshaw/tcomplete/blob/master/main.go
