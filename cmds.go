package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/xo/tblfmt"
)

/*
config
get
add
start
stop
move
remove
verify
reannounce
session
*/

// doConfig is the high-level entry point for 'config'.
func doConfig(args *Args) error {
	if args.ConfigParams.Value == "" {
		fmt.Fprintln(os.Stdout, args.Config.GetKey(args.ConfigParams.Name))
		return nil
	}
	args.Config.SetKey(args.ConfigParams.Name, args.ConfigParams.Value)
	return args.Config.Write(args.ConfigFile)
}

var intRE = regexp.MustCompile(`^[0-9]+$`)

// doGet is the high-level entry point for 'get'.
func doGet(args *Args) error {
	cl, err := args.newClient()
	if err != nil {
		return err
	}

	if len(args.IDs) == 0 && !args.All {
		return ErrMustSpecifyAllOrAtLeastOneTorrent
	}

	var id int
	var ids []interface{}
	for _, v := range args.IDs {
		if intRE.MatchString(v) {
			id, err = strconv.Atoi(v)
			if err != nil {
				return err
			}
			ids = append(ids, id)
		} else {
			ids = append(ids, v)
		}
	}
	res, err := cl.TorrentGet(context.Background(), ids...)
	if err != nil {
		return err
	}

	return tblfmt.EncodeTable(os.Stdout, NewTorrentResult(res.Torrents))
}

// doAdd is the high-level entry point for 'add'.
func doAdd(args *Args) error {
	return nil
}

// doStart is the high-level entry point for 'start'.
func doStart(args *Args) error {
	return nil
}

// doStop is the high-level entry point for 'stop'.
func doStop(args *Args) error {
	return nil
}

// doMove is the high-level entry point for 'move'.
func doMove(args *Args) error {
	return nil
}

// doRemove is the high-level entry point for 'remove'.
func doRemove(args *Args) error {
	return nil
}

// doVerify is the high-level entry point for 'verify'.
func doVerify(args *Args) error {
	return nil
}

// doReannounce is the high-level entry point for 'reannounce'.
func doReannounce(args *Args) error {
	return nil
}

// doSession is the high-level entry point for 'session'.
func doSession(args *Args) error {
	return nil
}
