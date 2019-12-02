// Command transctl is a command-line utility to manage transmission rpc hosts.
package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kingpin"
)

// Version is the command version.
var Version = "0.0.0-dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

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
		f = doStart
	case "stop":
		f = doStop
	case "move":
		f = doMove
	case "remove":
		f = doRemove
	case "verify":
		f = doVerify
	case "reannounce":
		f = doReannounce
	case "session":
		f = doSession
	case "session set":
		f = doSessionSet
	}

	return f(args)
}
