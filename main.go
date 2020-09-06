// Command transctl is a command-line utility to manage transmission rpc hosts.
package main

import (
	"fmt"
	"os"

	"github.com/knq/snaker"

	"github.com/kenshaw/transctl/providers"
	_ "github.com/kenshaw/transctl/providers/deluge"
	_ "github.com/kenshaw/transctl/providers/qbittorrent"
	_ "github.com/kenshaw/transctl/providers/rtorrent"
	_ "github.com/kenshaw/transctl/providers/transmission"
	_ "github.com/kenshaw/transctl/providers/utorrent"
)

// version is the command version.
var (
	name    = "transctl"
	version = "0.0.0-dev"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// run executes the command logic.
func run() error {
	// build args
	args, cmd, err := providers.NewArgs()
	if err != nil {
		return err
	}

	f := map[string]func(*providers.Args) error{
		"config":             providers.DoConfig,
		"add":                providers.DoAdd,
		"get":                providers.DoGet,
		"set":                providers.DoSet,
		"start":              providers.DoReq,
		"stop":               providers.DoReq,
		"move":               providers.DoMove,
		"remove":             providers.DoRemove,
		"verify":             providers.DoReq,
		"reannounce":         providers.DoReq,
		"queue top":          providers.DoReq,
		"queue bottom":       providers.DoReq,
		"queue up":           providers.DoReq,
		"queue down":         providers.DoReq,
		"peers get":          providers.DoPeersGet,
		"files get":          providers.DoFilesGet,
		"files set-priority": providers.DoFilesSet,
		"files set-wanted":   providers.DoFilesSet,
		"files set-unwanted": providers.DoFilesSet,
		"files rename":       providers.DoFilesRename,
		"trackers get":       providers.DoTrackersGet,
		"trackers add":       providers.DoTrackersAdd,
		"trackers replace":   providers.DoTrackersReplace,
		"trackers remove":    providers.DoTrackersRemove,
		"stats":              providers.DoStats,
		"shutdown":           providers.DoShutdown,
		"free-space":         providers.DoFreeSpace,
		"blocklist-update":   providers.DoBlocklistUpdate,
		"port-test":          providers.DoPortTest,
	}[cmd]

	return f(args, cmd)
}

func init() {
	if err := snaker.AddInitialisms("UTP"); err != nil {
		panic(err)
	}
}
