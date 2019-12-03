package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/kenshaw/transrpc"
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
		// build request
		req := transrpc.TorrentAdd().
			WithCookiesMap(args.AddParams.Cookies).
			WithDownloadDir(args.AddParams.DownloadDir).
			WithPaused(args.AddParams.Paused).
			WithPeerLimit(args.AddParams.PeerLimit).
			WithBandwidthPriority(args.AddParams.BandwidthPriority)

		// determine each arg is magnet link or file on disk
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
		fmt.Fprintf(os.Stdout, "added %d %q (%s)\n", t.ID, t.Name, t.HashString[:defaultShortHashLen])
	}

	return nil
}

// doSet is the high-level entry point for 'set'.
func doSet(args *Args) error {
	return nil
}

// doGet is the high-level entry point for 'get'.
func doGet(args *Args) error {
	_, torrents, err := args.findTorrents()
	if err != nil {
		return err
	}
	return NewTorrentResult(torrents).Encode(os.Stdout, args)
}

// do is the high-level entry point for 'start'.
func doReq(f func(...interface{}) *transrpc.Request) func(*Args) error {
	return func(args *Args) error {
		cl, torrents, err := args.findTorrents()
		if err != nil {
			return err
		}
		return f(convTorrentIDs(torrents)...).Do(context.Background(), cl)
	}
}

// doMove is the high-level entry point for 'move'.
func doMove(args *Args) error {
	cl, torrents, err := args.findTorrents()
	if err != nil {
		return err
	}
	return transrpc.TorrentSetLocation(
		args.MoveParams.Dest, true, convTorrentIDs(torrents)...,
	).Do(context.Background(), cl)
}

// doRemove is the high-level entry point for 'remove'.
func doRemove(args *Args) error {
	cl, torrents, err := args.findTorrents()
	if err != nil {
		return err
	}
	return transrpc.TorrentRemove(
		args.RemoveParams.Remove, convTorrentIDs(torrents)...,
	).Do(context.Background(), cl)
}

// doSession is the high-level entry point for 'session'.
func doSession(args *Args) error {
	return nil
}

// doSessionSet is the high-level entry point for 'session-set'.
func doSessionSet(args *Args) error {
	return nil
}
