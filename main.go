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

	// change flags from config file, if not set by command line flags
	if v := strings.ToLower(strings.TrimSpace(args.Config.GetKey("default.output"))); v != "" && !args.Output.OutputWasSet {
		args.Output.Output = v
	}
	if v := strings.ToLower(strings.TrimSpace(args.Config.GetKey("default.si"))); v != "" && !args.Output.SIWasSet {
		args.Output.SI = v == "true" || v == "1"
	}
	if v := args.getContextKey("match-order"); v != "" && !args.Filter.MatchOrderWasSet {
		args.Filter.MatchOrder = strings.Split(v, ",")
		for i := 0; i < len(args.Filter.MatchOrder); i++ {
			args.Filter.MatchOrder[i] = strings.ToLower(strings.TrimSpace(args.Filter.MatchOrder[i]))
		}
	}
	if v := strings.ToLower(strings.TrimSpace(args.Config.GetKey("command.add.rm"))); v != "" && !args.AddParams.RemoveWasSet {
		args.AddParams.Remove = v == "true" || v == "1"
	}
	if v := strings.TrimSpace(args.getContextKey("free-space")); cmd == "free-space" && v != "" && len(args.Args) == 0 {
		args.Args = strings.Split(v, ",")
		for i := 0; i < len(args.Args); i++ {
			args.Args[i] = strings.TrimSpace(args.Args[i])
		}
	}

	switch cmd {
	case "get", "set", "start", "stop", "move", "remove", "verify",
		"reannounce", "queue top", "queue bottom", "queue up", "queue down":
		// check exactly one of --recent, --all, or len(args.Args) > 0 conditions
		switch {
		case args.Filter.ListAll && args.Filter.Recent,
			args.Filter.ListAll && len(args.Args) != 0,
			args.Filter.Recent && len(args.Args) != 0,
			!args.Filter.ListAll && !args.Filter.Recent && len(args.Args) == 0:
			return ErrMustSpecifyAllRecentOrAtLeastOneTorrent
		}
	case "free-space":
		if len(args.Args) == 0 {
			return ErrMustSpecifyAtLeastOneLocation
		}
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

	//
	if cmd == "start" && args.StartParams.Now {
		f = doReq(transrpc.TorrentStartNow)
	}

	return f(args)
}
