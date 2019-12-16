package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/kenshaw/transrpc"
)

// doConfig is the high-level entry point for 'config'.
func doConfig(args *Args) error {
	var store ConfigStore = args.Config
	if args.ConfigParams.Remote {
		var err error
		store, err = NewRemoteConfigStore(args)
		if err != nil {
			return err
		}
	}

	switch {
	case args.ConfigParams.Unset:
		store.RemoveKey(args.ConfigParams.Name)
		return store.Write(args.ConfigFile)

	case args.ConfigParams.Name != "" && args.ConfigParams.Value == "":
		fmt.Fprintln(os.Stdout, store.GetKey(args.ConfigParams.Name))
		return nil

	case args.ConfigParams.Value != "":
		store.SetKey(args.ConfigParams.Name, args.ConfigParams.Value)
		return store.Write(args.ConfigFile)
	}

	// list all
	all := store.GetAllFlat()
	for i := 0; i < len(all); i += 2 {
		fmt.Fprintf(os.Stdout, "%s=%s\n", strings.TrimSpace(all[i]), strings.TrimSpace(all[i+1]))
	}

	return nil
}

// magnetRE is a regexp for magnet URLs.
var magnetRE = regexp.MustCompile(`(?i)^magnet:\?`)

// doAdd is the high-level entry point for 'add'.
func doAdd(args *Args) error {
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
		case err != nil && isMagnet:
			req.Filename = v
		case err != nil && os.IsNotExist(err) && !isMagnet:
			return fmt.Errorf("file not found: %s", v)
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

	return NewTorrentResult(added).Encode(os.Stdout, args, cl)
}

// doGet is the high-level entry point for 'get'.
func doGet(args *Args) error {
	cl, torrents, err := findTorrents(args)
	if err != nil {
		return err
	}
	return NewTorrentResult(torrents).Encode(os.Stdout, args, cl)
}

// doSet is the high-level entry point for 'set'.
func doSet(args *Args) error {
	cl, torrents, err := findTorrents(args)
	if err != nil {
		return err
	}
	if len(torrents) == 0 {
		return nil
	}
	return doWithAndExecute(cl, transrpc.TorrentSet(convTorrentIDs(torrents)...), "torrent", args.ConfigParams.Name, args.ConfigParams.Value)
}

// doReq creates the high-level entry points for general torrent manipulation
// requests.
func doReq(f func(...interface{}) *transrpc.Request) func(*Args) error {
	return func(args *Args) error {
		cl, torrents, err := findTorrents(args)
		if err != nil {
			return err
		}
		if len(torrents) == 0 {
			return nil
		}
		return f(convTorrentIDs(torrents)...).Do(context.Background(), cl)
	}
}

// doMove is the high-level entry point for 'move'.
func doMove(args *Args) error {
	cl, torrents, err := findTorrents(args)
	if err != nil {
		return err
	}
	if len(torrents) == 0 {
		return nil
	}
	return transrpc.TorrentSetLocation(
		args.MoveParams.Dest, true, convTorrentIDs(torrents)...,
	).Do(context.Background(), cl)
}

// doRemove is the high-level entry point for 'remove'.
func doRemove(args *Args) error {
	cl, torrents, err := findTorrents(args)
	if err != nil {
		return err
	}
	if len(torrents) == 0 {
		return nil
	}
	return transrpc.TorrentRemove(
		args.RemoveParams.Remove, convTorrentIDs(torrents)...,
	).Do(context.Background(), cl)
}

// doPeersGet is the high-level entry point for 'peers get'.
func doPeersGet(args *Args) error {
	return nil
}

// doFilesGet is the high-level entry point for 'files get'.
func doFilesGet(args *Args) error {
	return nil
}

// doFilesSetPriority is the high-level entry point for 'files set-priority'.
func doFilesSetPriority(args *Args) error {
	return nil
}

// doFilesSetLocation is the high-level entry point for 'files set-location'.
func doFilesSetLocation(args *Args) error {
	return nil
}

// doTrackersGet is the high-level entry point for 'trackers get'.
func doTrackersGet(args *Args) error {
	return nil
}

// doTrackersAdd is the high-level entry point for 'trackers add'.
func doTrackersAdd(args *Args) error {
	cl, torrents, err := findTorrents(args)
	if err != nil {
		return err
	}
	if len(torrents) == 0 {
		return nil
	}
	return transrpc.TorrentSet(convTorrentIDs(torrents)...).
		WithTrackerAdd([]string{args.Tracker}).Do(context.Background(), cl)
}

// doTrackersReplace is the high-level entry point for 'trackers replace'.
func doTrackersReplace(args *Args) error {
	cl, torrents, err := findTorrents(args)
	if err != nil {
		return err
	}
	if len(torrents) == 0 {
		return nil
	}
	return transrpc.TorrentSet(convTorrentIDs(torrents)...).
		WithTrackerReplace([]string{args.Tracker, args.TrackersReplaceParams.Replace}).Do(context.Background(), cl)
}

// doTrackersRemove is the high-level entry point for 'trackers remove'.
func doTrackersRemove(args *Args) error {
	cl, torrents, err := findTorrents(args)
	if err != nil {
		return err
	}
	if len(torrents) == 0 {
		return nil
	}
	for _, t := range torrents {
		for _, tracker := range t.Trackers {
			if tracker.Announce == args.Tracker {
				if err := transrpc.TorrentSet(t.HashString).WithTrackerRemove([]int64{tracker.ID}).Do(context.Background(), cl); err != nil {
					return fmt.Errorf("could not remove tracker %d (%s) from %s: %w", tracker.ID, args.Tracker, t.HashString, err)
				}
			}
		}
	}
	return nil
}

// doStats is the high-level entry point for 'stats'.
func doStats(args *Args) error {
	cl, err := args.newClient()
	if err != nil {
		return err
	}
	stats, err := cl.SessionStats(context.Background())
	if err != nil {
		return err
	}
	m := make(map[string]string)
	addFieldsToMap(m, "", reflect.ValueOf(*stats))
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(os.Stdout, "%s=%s\n", strings.TrimSpace(k), strings.TrimSpace(m[k]))
	}
	return nil
}

// doShutdown is the high-level entry point for 'shutdown'.
func doShutdown(args *Args) error {
	cl, err := args.newClient()
	if err != nil {
		return err
	}
	return cl.SessionClose(context.Background())
}

// doFreeSpace is the high-level entry point for 'free-space'.
func doFreeSpace(args *Args) error {
	cl, err := args.newClient()
	if err != nil {
		return err
	}
	for _, path := range args.Args {
		size, err := cl.FreeSpace(context.Background(), path)
		var sz string
		switch {
		case err != nil:
			if e, ok := err.(*transrpc.ErrRequestFailed); ok {
				sz = "error: " + e.Err
			} else {
				sz = "error: " + err.Error()
			}
		case args.Output.Human == "true" || args.Output.Human == "1" || args.Output.SI:
			sz = size.Format(!args.Output.SI, 2, "")
		default:
			sz = strconv.FormatInt(int64(size), 10)
		}
		fmt.Fprintf(os.Stdout, "%s\t%s\n", path, sz)
	}
	return nil
}

// doBlocklistUpdate is the high-level entry point for 'blocklist-update'.
func doBlocklistUpdate(args *Args) error {
	cl, err := args.newClient()
	if err != nil {
		return err
	}
	count, err := cl.BlocklistUpdate(context.Background())
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, count)
	return nil
}

// doPortTest is the high-level entry point for 'port-test'.
func doPortTest(args *Args) error {
	cl, err := args.newClient()
	if err != nil {
		return err
	}
	status, err := cl.PortTest(context.Background())
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "%t\n", status)
	return nil
}
