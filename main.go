// Command transctl is a command-line utility to manage transmission rpc hosts.
package main

import (
	"fmt"
	"os"
	"strings"

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

	// change flags from config file, if not set
	if v := strings.ToLower(strings.TrimSpace(args.Config.GetKey("default.output"))); v != "" && !args.OutputWasSet {
		args.Output = v
	}
	if v := strings.ToLower(strings.TrimSpace(args.Config.GetKey("default.si"))); v != "" && !args.SIWasSet {
		args.SI = v == "true" || v == "1"
	}
	if v := args.getContextKey("match-order"); v != "" && !args.MatchOrderWasSet {
		args.MatchOrder = strings.Split(v, ",")
		for i := 0; i < len(args.MatchOrder); i++ {
			args.MatchOrder[i] = strings.ToLower(strings.TrimSpace(args.MatchOrder[i]))
		}
	}
	if v := strings.ToLower(strings.TrimSpace(args.Config.GetKey("command.add.rm"))); v != "" && !args.AddParams.RemoveWasSet {
		args.AddParams.Remove = v == "true" || v == "1"
	}

	switch cmd {
	case "get", "set", "start", "stop", "move", "remove", "verify",
		"reannounce", "queue top", "queue bottom", "queue up", "queue down":
		// check exactly one of --recent, --all, or len(args.Args) > 0 conditions
		switch {
		case args.ListAll && args.Recent,
			args.ListAll && len(args.Args) != 0,
			args.Recent && len(args.Args) != 0,
			!args.ListAll && !args.Recent && len(args.Args) == 0:
			return ErrMustSpecifyAllRecentOrAtLeastOneTorrent
		}
	}

	var f func(*Args) error
	switch cmd {
	case "config":
		f = doConfig
	case "add":
		f = doAdd
	case "get":
		f = doGet
	case "set":
		f = doSet
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
	case "stats":
		f = doStats
	case "shutdown":
		f = doShutdown
	case "free-space":
		f = doFreeSpace
	case "blocklist-update":
		f = doBlocklistUpdate
	case "port-test":
		f = doPortTest
	}

	return f(args)
}
