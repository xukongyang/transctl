package providers

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gobwas/glob"
	"github.com/kenshaw/transctl/tctypes"
	"github.com/kenshaw/transctl/transrpc"
)

/*
	"config":             doConfig,
	"add":                doAdd,
	"get":                doGet,
	"set":                doSet,
	"start":              doReq(transrpc.TorrentStart),
	"stop":               doReq(transrpc.TorrentStop),
	"move":               doMove,
	"remove":             doRemove,
	"verify":             doReq(transrpc.TorrentVerify),
	"reannounce":         doReq(transrpc.TorrentReannounce),
	"queue top":          doReq(transrpc.QueueMoveTop),
	"queue bottom":       doReq(transrpc.QueueMoveBottom),
	"queue up":           doReq(transrpc.QueueMoveUp),
	"queue down":         doReq(transrpc.QueueMoveDown),
	"peers get":          doPeersGet,
	"files get":          doFilesGet,
	"files set-priority": doFilesSetPriority,
	"files set-wanted":   doFilesSet("FilesWanted"),
	"files set-unwanted": doFilesSet("FilesUnwanted"),
	"files rename":       doFilesRename,
	"trackers get":       doTrackersGet,
	"trackers add":       doTrackersAdd,
	"trackers replace":   doTrackersReplace,
	"trackers remove":    doTrackersRemove,
	"stats":              doStats,
	"shutdown":           doShutdown,
	"free-space":         doFreeSpace,
	"blocklist-update":   doBlocklistUpdate,
	"port-test":          doPortTest,
*/

// DoConfig is the high-level entry point for 'config'.
func DoConfig(args *Args, cmd string) error {
	var store ConfigStore = args.Config
	if args.ConfigParams.Remote {
		var err error
		p, err := args.NewProvider()
		if err != nil {
			return err
		}
		store, err = p.NewRemoteConfigStore(context.Background())
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

// DoAdd is the high-level entry point for 'add'.
func DoAdd(args *Args, cmd string) error {
	p, err := args.NewProvider()
	if err != nil {
		return err
	}

	var files []interface{}
	for _, v := range args.Args {
		// determine each arg is magnet link or file on disk
		isMagnet := magnetRE.MatchString(v)
		fi, err := os.Stat(v)
		switch {
		case err != nil && isMagnet:
			files = append(files, v)
		case err != nil && os.IsNotExist(err) && !isMagnet:
			return fmt.Errorf("file not found: %s", v)
		case err != nil:
			return err
		case err == nil && fi.IsDir():
			return fmt.Errorf("cannot add directory %s as torrent", v)
		case err == nil:
			var buf []byte
			buf, err = ioutil.ReadFile(v)
			if err != nil {
				return err
			}
			files = append(files, buf)
		}
	}

	/*
		result, err := p.Add(files...)
			// build request
			req := transrpc.TorrentAdd().
				WithCookiesMap(args.AddParams.Cookies).
				WithDownloadDir(args.AddParams.DownloadDir).
				WithPaused(args.AddParams.Paused).
				WithPeerLimit(args.AddParams.PeerLimit).
				WithBandwidthPriority(args.AddParams.BandwidthPriority)

			// execute
			res, err := req.Do(context.Background(), cl)
			if err != nil {
				return err
			}
			if res.TorrentAdded != nil {
				result = append(result, *res.TorrentAdded)
			}
			if res.TorrentDuplicate != nil {
				result = append(result, *res.TorrentDuplicate)
			}
	*/

	// execute
	result, err := p.Add(context.Background(), files...)
	if err != nil {
		return err
	}
	if err = NewResult(result, args.ResultOptions(
		TableColumns(defaultTableCols...),
		WideColumns(defaultWideCols...),
		FlatName("torrent"),
		FlatIndex("shortHash"),
	)...).Encode(os.Stdout); err != nil {
		return err
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
	return nil
}

var (
	defaultTableCols = []string{"id", "name", "status", "eta", "rateDownload", "rateUpload", "haveValid", "percentDone", "shortHash"}
	defaultWideCols  = []string{"id", "name", "peersConnected", "downloadDir", "addedDate", "status", "eta", "rateDownload", "rateUpload", "haveValid", "percentDone", "shortHash"}
)

// DoGet is the high-level entry point for 'get'.
func DoGet(args *Args, cmd string) error {
	p, torrents, err := findTorrents(args)
	if err != nil {
		return err
	}
	var result []tctypes.Torrent
	if len(torrents) != 0 {
		var fields []string
		switch {
		case args.Output.Output == "table":
			fields = defaultTableCols
		case args.Output.Output == "wide":
			fields = defaultWideCols
		case strings.HasPrefix(args.Output.Output, "cols="):
			fields = strings.Split(args.Output.Output[5:], ",")
		}
		req := transrpc.TorrentGet(convTorrentIDs(torrents)...)
		if len(fields) != 0 {
			var cols []string
			// inverse lookup
			m := make(map[string]string)
			for k, v := range args.Output.ColumnNames {
				m[v] = k
			}
			for _, col := range fields {
				col = strings.TrimSpace(col)
				if c, ok := m[col]; ok {
					col = c
				}
				if col == "shortHash" {
					col = "hashString"
				}
				cols = append(cols, col)
			}
			req = req.WithFields(cols...)
		}
		res, err := req.Do(context.Background(), cl)
		if err != nil {
			return err
		}
		result = res.Torrents
	}
	return NewResult(result, args.ResultOptions(
		TableColumns(defaultTableCols...),
		WideColumns(defaultWideCols...),
		FlatName("torrent"),
		FlatIndex("shortHash"),
	)...).Encode(os.Stdout)
}

// DoSet is the high-level entry point for 'set'.
func DoSet(args *Args, cmd string) error {
	p, torrents, err := findTorrents(args)
	if err != nil {
		return err
	}
	if len(torrents) == 0 {
		return nil
	}
	return doWithAndExecute(cl, transrpc.TorrentSet(convTorrentIDs(torrents)...), "torrent", args.ConfigParams.Name, args.ConfigParams.Value)
}

// DoReq creates the high-level entry points for general torrent manipulation
// requests.
func DoReq(f func(...interface{}) *transrpc.Request) func(*Args) error {
	return func(args *Args) error {
		p, torrents, err := findTorrents(args)
		if err != nil {
			return err
		}
		if len(torrents) == 0 {
			return nil
		}
		return f(convTorrentIDs(torrents)...).Do(context.Background(), p)
	}
}

// DoMove is the high-level entry point for 'move'.
func DoMove(args *Args) error {
	p, torrents, err := findTorrents(args)
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

// DoRemove is the high-level entry point for 'remove'.
func DoRemove(args *Args) error {
	p, torrents, err := findTorrents(args)
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

// DoPeersGet is the high-level entry point for 'peers get'.
func DoPeersGet(args *Args) error {
	p, torrents, err := findTorrents(args)
	if err != nil {
		return err
	}
	var result []tctypes.Peer
	if len(torrents) != 0 {
		res, err := transrpc.TorrentGet(convTorrentIDs(torrents)...).WithFields("name", "hashString", "peers").Do(context.Background(), cl)
		if err != nil {
			return err
		}
		for _, t := range res.Torrents {
			for i, v := range t.Peers {
				result = append(result, tctypes.Peer{
					Address:            v.Address,
					ClientName:         v.ClientName,
					ClientIsChoked:     v.ClientIsChoked,
					ClientIsInterested: v.ClientIsInterested,
					FlagStr:            v.FlagStr,
					IsDownloadingFrom:  v.IsDownloadingFrom,
					IsEncrypted:        v.IsEncrypted,
					IsIncoming:         v.IsIncoming,
					IsUploadingTo:      v.IsUploadingTo,
					IsUTP:              v.IsUTP,
					PeerIsChoked:       v.PeerIsChoked,
					PeerIsInterested:   v.PeerIsInterested,
					Port:               v.Port,
					Progress:           v.Progress,
					RateToClient:       v.RateToClient,
					RateToPeer:         v.RateToPeer,
					ID:                 int64(i),
					Torrent:            t.Name,
					HashString:         t.HashString,
				})
			}
		}
	}
	return NewResult(result, args.ResultOptions(
		TableColumns("address", "clientName", "rateToClient", "rateToPeer", "progress", "shortHash"),
		WideColumns("address", "port", "clientName", "flagStr", "clientIsInterested", "isEncrypted", "rateToClient", "rateToPeer", "progress", "shortHash"),
		YamlName("peers"),
		FlatName("peers"),
		FlatKey("id"),
		FlatIndex("shortHash"),
	)...).Encode(os.Stdout)
}

// DoFilesGet is the high-level entry point for 'files get'.
func DoFilesGet(args *Args) error {
	p, torrents, err := findTorrents(args)
	if err != nil {
		return err
	}
	var result []tctypes.File
	if len(torrents) != 0 {
		res, err := transrpc.TorrentGet(convTorrentIDs(torrents)...).WithFields("name", "hashString", "files", "fileStats").Do(context.Background(), cl)
		if err != nil {
			return err
		}
		for _, t := range res.Torrents {
			for i, v := range t.Files {
				result = append(result, tctypes.File{
					BytesCompleted: v.BytesCompleted,
					Length:         v.Length,
					Name:           v.Name,
					Wanted:         t.FileStats[i].Wanted,
					Priority:       t.FileStats[i].Priority.String(),
					ID:             int64(i),
					Torrent:        t.Name,
					HashString:     t.HashString,
				})
			}
		}
	}
	return NewResult(result, args.ResultOptions(
		TableColumns("name", "priority", "bytesCompleted", "percentDone", "shortHash"),
		WideColumns("name", "priority", "wanted", "bytesCompleted", "length", "percentDone", "id", "shortHash"),
		YamlName("files"),
		FlatName("files"),
		FlatKey("id"),
		FlatIndex("shortHash"),
	)...).Encode(os.Stdout)
}

// DoFilesSet generates a torrent set request to change the specified field.
func DoFilesSet(field string) func(args *Args) error {
	var errmsg string
	switch {
	case strings.HasPrefix(field, "Files"):
		errmsg = "files set-" + strings.ToLower(strings.TrimPrefix(field, "Files"))
	case strings.HasPrefix(field, "Priority"):
		errmsg = "files set-priority " + strings.ToLower(strings.TrimPrefix(field, "Priority"))
	}
	return func(args *Args) error {
		g, err := glob.Compile(args.FileMask)
		if err != nil {
			return err
		}
		p, torrents, err := findTorrents(args)
		if err != nil {
			return err
		}
		if len(torrents) == 0 {
			return nil
		}
		res, err := transrpc.TorrentGet(convTorrentIDs(torrents)...).WithFields("hashString", "files").Do(context.Background(), cl)
		if err != nil {
			return nil
		}
		for _, t := range res.Torrents {
			var files []string
			for i := 0; i < len(t.Files); i++ {
				if g.Match(t.Files[i].Name) {
					files = append(files, strconv.Itoa(i))
				}
			}
			if len(files) != 0 {
				if err = doWithAndExecute(cl, transrpc.TorrentSet(t.HashString), errmsg, field, strings.Join(files, ",")); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

// DoFilesSetPriority is the high-level entry point for 'files set-priority'.
func DoFilesSetPriority(args *Args) error {
	switch args.FilesSetPriorityParams.Priority {
	case "low":
		return doFilesSet("PriorityLow")(args)
	case "normal":
		return doFilesSet("PriorityNormal")(args)
	case "high":
		return doFilesSet("PriorityHigh")(args)
	}
	return nil
}

// DoFilesRename is the high-level entry point for 'files rename'.
func DoFilesRename(args *Args) error {
	p, torrents, err := findTorrents(args)
	if err != nil {
		return err
	}
	if len(torrents) == 0 {
		return nil
	}
	req := transrpc.TorrentGet(convTorrentIDs(torrents)...).WithFields("hashString", "files")
	res, err := req.Do(context.Background(), cl)
	if err != nil {
		return err
	}
	for _, t := range res.Torrents {
		for _, f := range t.Files {
			if f.Name == args.FilesRenameParams.OldPath {
				if err = transrpc.TorrentRenamePath(
					args.FilesRenameParams.OldPath, args.FilesRenameParams.NewPath, t.HashString,
				).Do(context.Background(), cl); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// DoTrackersGet is the high-level entry point for 'trackers get'.
func DoTrackersGet(args *Args) error {
	p, torrents, err := findTorrents(args)
	if err != nil {
		return err
	}
	var result []tctypes.Tracker
	if len(torrents) != 0 {
		res, err := transrpc.TorrentGet(convTorrentIDs(torrents)...).WithFields("name", "hashString", "trackers", "trackerStats").Do(context.Background(), cl)
		if err != nil {
			return err
		}
		for _, t := range res.Torrents {
			for i, v := range t.Trackers {
				result = append(result, tctypes.Tracker{
					Announce:              v.Announce,
					ID:                    v.ID,
					Scrape:                v.Scrape,
					Tier:                  v.Tier,
					AnnounceState:         t.TrackerStats[i].AnnounceState,
					DownloadCount:         t.TrackerStats[i].DownloadCount,
					HasAnnounced:          t.TrackerStats[i].HasAnnounced,
					HasScraped:            t.TrackerStats[i].HasScraped,
					Host:                  t.TrackerStats[i].Host,
					IsBackup:              t.TrackerStats[i].IsBackup,
					LastAnnouncePeerCount: t.TrackerStats[i].LastAnnouncePeerCount,
					LastAnnounceResult:    t.TrackerStats[i].LastAnnounceResult,
					LastAnnounceStartTime: t.TrackerStats[i].LastAnnounceStartTime,
					LastAnnounceSucceeded: t.TrackerStats[i].LastAnnounceSucceeded,
					LastAnnounceTime:      t.TrackerStats[i].LastAnnounceTime,
					LastAnnounceTimedOut:  t.TrackerStats[i].LastAnnounceTimedOut,
					LastScrapeResult:      t.TrackerStats[i].LastScrapeResult,
					LastScrapeStartTime:   t.TrackerStats[i].LastScrapeStartTime,
					LastScrapeSucceeded:   t.TrackerStats[i].LastScrapeSucceeded,
					LastScrapeTime:        t.TrackerStats[i].LastScrapeTime,
					LastScrapeTimedOut:    t.TrackerStats[i].LastScrapeTimedOut,
					LeecherCount:          t.TrackerStats[i].LeecherCount,
					NextAnnounceTime:      t.TrackerStats[i].NextAnnounceTime,
					NextScrapeTime:        t.TrackerStats[i].NextScrapeTime,
					ScrapeState:           t.TrackerStats[i].ScrapeState,
					SeederCount:           t.TrackerStats[i].SeederCount,
					Torrent:               t.Name,
					HashString:            t.HashString,
				})
			}
		}
	}
	return NewResult(result, args.ResultOptions(
		TableColumns("announce", "lastAnnounceResult", "lastAnnouncePeerCount", "seederCount", "shortHash"),
		WideColumns("announce", "announceState", "lastAnnounceResult", "lastAnnounceTime", "nextAnnounceTime", "lastAnnouncePeerCount", "seederCount", "tier", "shortHash"),
		YamlName("trackers"),
		FlatName("trackers"),
		FlatKey("id"),
		FlatIndex("shortHash"),
	)...).Encode(os.Stdout)
}

// DoTrackersAdd is the high-level entry point for 'trackers add'.
func DoTrackersAdd(args *Args) error {
	p, torrents, err := findTorrents(args)
	if err != nil {
		return err
	}
	if len(torrents) == 0 {
		return nil
	}
	return transrpc.TorrentSet(convTorrentIDs(torrents)...).
		WithTrackerAdd(args.Tracker).Do(context.Background(), cl)
}

// DoTrackersReplace is the high-level entry point for 'trackers replace'.
func DoTrackersReplace(args *Args) error {
	p, torrents, err := findTorrents(args)
	if err != nil {
		return err
	}
	if len(torrents) == 0 {
		return nil
	}
	req := transrpc.TorrentGet(convTorrentIDs(torrents)...).WithFields("hashString", "trackers")
	res, err := req.Do(context.Background(), cl)
	if err != nil {
		return err
	}
	for _, t := range res.Torrents {
		for _, tracker := range t.Trackers {
			if tracker.Announce == args.Tracker {
				if err := transrpc.TorrentSet(t.HashString).WithTrackerReplace(
					tracker.ID, args.TrackersReplaceParams.Replace,
				).Do(context.Background(), cl); err != nil {
					return fmt.Errorf("could not replace tracker %d (%s) with %s for %s: %w",
						tracker.ID, args.Tracker, args.TrackersReplaceParams.Replace, t.HashString, err)
				}
			}
		}
	}
	return nil
}

// DoTrackersRemove is the high-level entry point for 'trackers remove'.
func DoTrackersRemove(args *Args) error {
	p, torrents, err := findTorrents(args)
	if err != nil {
		return err
	}
	if len(torrents) == 0 {
		return nil
	}
	req := transrpc.TorrentGet(convTorrentIDs(torrents)...).WithFields("hashString", "trackers")
	res, err := req.Do(context.Background(), cl)
	if err != nil {
		return err
	}
	for _, t := range res.Torrents {
		for _, tracker := range t.Trackers {
			if tracker.Announce == args.Tracker {
				if err := transrpc.TorrentSet(t.HashString).WithTrackerRemove(
					tracker.ID,
				).Do(context.Background(), cl); err != nil {
					return fmt.Errorf("could not remove tracker %d (%s) from %s: %w", tracker.ID, args.Tracker, t.HashString, err)
				}
			}
		}
	}
	return nil
}

/*
type keypair struct {
	HashString string      `json:"-" yaml:"-"`
	Name       string      `json:"name,omitempty" yaml:"name,omitempty"`
	Key        string      `json:"-" yaml:"-"`
	Value      interface{} `json:"value,omitempty" yaml:"value,omitempty"`
	ID         int64       `json:"id" yaml:"id"`
}
*/

// DoStats is the high-level entry point for 'stats'.
func DoStats(args *Args) error {
	p, err := args.NewProvider()
	if err != nil {
		return err
	}
	res, err := p.Stats(context.Background())
	if err != nil {
		return err
	}
	/*
		[]keypair{
			{"session-stats", "Active Torrent Count", "active-torrent-count", res.ActiveTorrentCount, 0},
			{"session-stats", "Download Speed", "download-speed", res.DownloadSpeed, 1},
			{"session-stats", "Paused Torrent Count", "paused-torrent-count", res.PausedTorrentCount, 2},
			{"session-stats", "Torrent Count", "torrent-count", res.TorrentCount, 3},
			{"session-stats", "Upload Speed", "upload-speed", res.UploadSpeed, 4},
			{"session-stats", "Cumulative Uploaded", "cumulative-stats.uploaded-bytes", res.CumulativeStats.UploadedBytes, 5},
			{"session-stats", "Cumulative Downloaded", "cumulative-stats.downloaded-bytes", res.CumulativeStats.DownloadedBytes, 6},
			{"session-stats", "Cumulative Files Added", "cumulative-stats.files-added", res.CumulativeStats.FilesAdded, 7},
			{"session-stats", "Cumulative Session Count", "cumulative-stats.session-count", res.CumulativeStats.SessionCount, 8},
			{"session-stats", "Cumulative Seconds Active", "cumulative-stats.seconds-active", res.CumulativeStats.SecondsActive, 9},
			{"session-stats", "Current Uploaded", "current-stats.uploaded-bytes", res.CurrentStats.UploadedBytes, 10},
			{"session-stats", "Current Downloaded", "current-stats.downloaded-bytes", res.CurrentStats.DownloadedBytes, 11},
			{"session-stats", "Current Files Added", "current-stats.files-added", res.CurrentStats.FilesAdded, 12},
			{"session-stats", "Current Session Count", "current-stats.session-count", res.CurrentStats.SessionCount, 13},
			{"session-stats", "Current Seconds Active", "current-stats.seconds-active", res.CurrentStats.SecondsActive, 14},
		},
	*/
	return NewResult(
		res,
		args.ResultOptions(
			TableColumns("name", "value"),
			WideColumns("name", "key", "value"),
			YamlName("session-stats"),
			FlatName("session-stats"),
			FlatKey("id"),
			FlatIndex("hashString"),
			NoTotals(true),
		)...,
	).Encode(os.Stdout)
}

// DoShutdown is the high-level entry point for 'shutdown'.
func DoShutdown(args *Args) error {
	p, err := args.NewProvider()
	if err != nil {
		return err
	}
	return p.SessionClose(context.Background())
}

// DoFreeSpace is the high-level entry point for 'free-space'.
func DoFreeSpace(args *Args) error {
	p, err := args.NewProvider()
	if err != nil {
		return err
	}
	for _, path := range args.Args {
		size, err := p.FreeSpace(context.Background(), path)
		var sz string
		switch {
		case err != nil:
			if e, ok := err.(*transrpc.ErrRequestFailed); ok {
				sz = "error: " + e.Err
			} else {
				sz = "error: " + err.Error()
			}
		case args.Output.Human == "true" || args.Output.Human == "1" || args.Output.SI:
			sz = size.Format(!args.Output.SI, 2)
		default:
			sz = strconv.FormatInt(int64(size), 10)
		}
		fmt.Fprintf(os.Stdout, "%s\t%s\n", path, sz)
	}
	return nil
}

// DoBlocklistUpdate is the high-level entry point for 'blocklist-update'.
func DoBlocklistUpdate(args *Args) error {
	p, err := args.NewProvider()
	if err != nil {
		return err
	}
	count, err := p.BlocklistUpdate(context.Background())
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, count)
	return nil
}

// DoPortTest is the high-level entry point for 'port-test'.
func DoPortTest(args *Args) error {
	p, err := args.NewProvider()
	if err != nil {
		return err
	}
	status, err := p.PortTest(context.Background())
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "%t\n", status)
	return nil
}
