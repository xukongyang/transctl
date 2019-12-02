package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"

	"github.com/kenshaw/transrpc"
	"github.com/xo/tblfmt"
)

/*
get
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
	switch {
	case args.ConfigParams.Unset && args.ConfigParams.Value != "":
		return ErrCannotSpecifyUnsetWhileTryingToSetAValueWithConfig

	case args.ConfigParams.Unset:
		args.Config.RemoveKey(args.ConfigParams.Name)

	case args.ConfigParams.Value == "":
		fmt.Fprintln(os.Stdout, args.Config.GetKey(args.ConfigParams.Name))
		return nil

	case args.ConfigParams.Value != "":
		args.Config.SetKey(args.ConfigParams.Name, args.ConfigParams.Value)
	}

	return args.Config.Write(args.ConfigFile)
}

// doContextSet is the high-level entry point for 'context-set'.
func doContextSet(args *Args) error {
	args.Config.SetKey("default.context", args.Context)
	return args.Config.Write(args.ConfigFile)
}

var magnetRE = regexp.MustCompile(`(?i)^magnet:\?`)

// doAdd is the high-level entry point for 'add'.
func doAdd(args *Args) error {
	if len(args.Args) < 1 {
		return ErrMustSpecifyAtLeastOneTorrentOrURI
	}

	cl, err := args.newClient()
	if err != nil {
		return err
	}

	var added []transrpc.Torrent
	for _, v := range args.Args {
		req := transrpc.TorrentAdd().
			WithCookiesMap(args.AddParams.Cookies).
			WithDownloadDir(args.AddParams.DownloadDir).
			WithPaused(args.AddParams.Paused).
			WithPeerLimit(args.AddParams.PeerLimit).
			WithBandwidthPriority(args.AddParams.BandwidthPriority)

		isMagnet := magnetRE.MatchString(v)
		fi, err := os.Stat(v)
		switch {
		case err != nil && os.IsNotExist(err) && !isMagnet:
			return fmt.Errorf("file not found: %s", v)
		case err != nil && os.IsNotExist(err) && isMagnet:
			req.Filename = v
		case err != nil:
			return err
		case err == nil && fi.IsDir():
			return fmt.Errorf("cannot add directory %s as torrent", v)
		case err == nil:
			req.Metainfo, err = ioutil.ReadFile(v)
			if err != nil {
				return err
			}
		}

		// execute
		res, err := req.Do(context.Background(), cl)
		if err != nil {
			return err
		}
		if res.TorrentAdded != nil {
			added = append(added, *res.TorrentAdded)
		}
		if res.TorrentDuplicate != nil {
			added = append(added, *res.TorrentDuplicate)
		}
	}

	// remove
	if args.AddParams.Remove {
		for _, v := range args.Args {
			if magnetRE.MatchString(v) {
				continue
			}
			if err = os.Remove(v); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}

	for _, t := range added {
		fmt.Fprintf(os.Stdout, "added %d %q (%s)\n", t.ID, t.Name, t.HashString[:7])
	}

	return nil
}

// doSet is the high-level entry point for 'set'.
func doSet(args *Args) error {
	return nil
}

var intRE = regexp.MustCompile(`^[0-9]+$`)

// doGet is the high-level entry point for 'get'.
func doGet(args *Args) error {
	cl, err := args.newClient()
	if err != nil {
		return err
	}

	if len(args.Args) == 0 && !args.All {
		return ErrMustSpecifyAllOrAtLeastOneTorrent
	}

	var id int
	var ids []interface{}
	for _, v := range args.Args {
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

// doSessionSet is the high-level entry point for 'session-set'.
func doSessionSet(args *Args) error {
	return nil
}
