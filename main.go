// Command transctl is a command-line utility to manage transmission rpc hosts.
package main

import (
	"fmt"
	"os"

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
	args, cmd, err := NewArgs()
	if err != nil {
		return err
	}

	f := map[string]func(*Args) error{
		"config":           doConfig,
		"add":              doAdd,
		"get":              doGet,
		"set":              doSet,
		"start":            doReq(transrpc.TorrentStart),
		"stop":             doReq(transrpc.TorrentStop),
		"move":             doMove,
		"remove":           doRemove,
		"verify":           doReq(transrpc.TorrentVerify),
		"reannounce":       doReq(transrpc.TorrentReannounce),
		"queue top":        doReq(transrpc.QueueMoveTop),
		"queue bottom":     doReq(transrpc.QueueMoveBottom),
		"queue up":         doReq(transrpc.QueueMoveUp),
		"queue down":       doReq(transrpc.QueueMoveDown),
		"stats":            doStats,
		"shutdown":         doShutdown,
		"free-space":       doFreeSpace,
		"blocklist-update": doBlocklistUpdate,
		"port-test":        doPortTest,
	}[cmd]

	// start --now special case
	if cmd == "start" && args.StartParams.Now {
		f = doReq(transrpc.TorrentStartNow)
	}

	return f(args)
}
