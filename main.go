// Command transctl is a command-line utility to manage transmission rpc hosts.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/kenshaw/transrpc"
)

// version is the command version.
var version = "0.0.0-dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// run executes the command logic.
func run() error {
	// build args
	args, err := NewArgs()
	if err != nil {
		return err
	}

	cmd := kingpin.Parse()

	// load config
	if err = args.loadConfig(); err != nil {
		return err
	}

	log.Printf(">>> cmd: %s", cmd)

	switch cmd {
	case "queue":

	case "get", "start", "stop", "move", "remove", "verify", "reannounce",
		"queue top", "queue bottom", "queue up", "queue down":
		// check exactly one of --recent, --all, or len(args.Args) > 0 conditions
		switch {
		case args.All && args.Recent,
			args.All && len(args.Args) != 0,
			args.Recent && len(args.Args) != 0:
			return ErrMustSpecifyAllRecentOrAtLeastOneTorrent
		}
	}

	var f func(*Args) error
	switch cmd {
	case "config":
		f = doConfig
	case "context-set":
		f = doContextSet
	case "add":
		f = doAdd
	case "set":
		f = doSet
	case "get":
		f = doGet
	case "start":
		if args.StartParams.Now {
			f = doReq(transrpc.TorrentStartNow)
		} else {
			f = doReq(transrpc.TorrentStart)
		}
	case "stop":
		f = doReq(transrpc.TorrentStop)
	case "move":
		f = doMove
	case "remove":
		f = doRemove
	case "verify":
		f = doReq(transrpc.TorrentVerify)
	case "reannounce":
		f = doReq(transrpc.TorrentReannounce)
	case "queue top":
		f = doReq(transrpc.QueueMoveTop)
	case "queue bottom":
		f = doReq(transrpc.QueueMoveBottom)
	case "queue up":
		f = doReq(transrpc.QueueMoveUp)
	case "queue down":
		f = doReq(transrpc.QueueMoveDown)

	case "session":
		f = doSession
	case "session set":
		f = doSessionSet
	}

	return f(args)
}
