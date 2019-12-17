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

	"github.com/gobwas/glob"
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

	var result []transrpc.Torrent
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
			result = append(result, *res.TorrentAdded)
		}
		if res.TorrentDuplicate != nil {
			result = append(result, *res.TorrentDuplicate)
		}

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

	return err
}

var (
	defaultTableCols = []string{"id", "name", "status", "eta", "rateDownload", "rateUpload", "haveValid", "percentDone", "shortHash"}
	defaultWideCols  = []string{"id", "name", "peersConnected", "downloadDir", "addedDate", "status", "eta", "rateDownload", "rateUpload", "haveValid", "percentDone", "shortHash"}
)

// doGet is the high-level entry point for 'get'.
func doGet(args *Args) error {
	cl, torrents, err := findTorrents(args)
	if err != nil {
		return err
	}
	var result []transrpc.Torrent
	if len(torrents) != 0 {
		var fields []string
		switch {
		case args.Output.Output == "table":
			fields = defaultTableCols
		case args.Output.Output == "wide":
			fields = defaultWideCols
		case strings.HasPrefix(args.Output.Output, "table="):
			fields = strings.Split(args.Output.Output[6:], ",")
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

type peer struct {
	Address            string             `json:"address,omitempty" yaml:"address,omitempty"`                       // tr_peer_stat
	ClientName         string             `json:"clientName,omitempty" yaml:"clientName,omitempty"`                 // tr_peer_stat
	ClientIsChoked     bool               `json:"clientIsChoked,omitempty" yaml:"clientIsChoked,omitempty"`         // tr_peer_stat
	ClientIsInterested bool               `json:"clientIsInterested,omitempty" yaml:"clientIsInterested,omitempty"` // tr_peer_stat
	FlagStr            string             `json:"flagStr,omitempty" yaml:"flagStr,omitempty"`                       // tr_peer_stat
	IsDownloadingFrom  bool               `json:"isDownloadingFrom,omitempty" yaml:"isDownloadingFrom,omitempty"`   // tr_peer_stat
	IsEncrypted        bool               `json:"isEncrypted,omitempty" yaml:"isEncrypted,omitempty"`               // tr_peer_stat
	IsIncoming         bool               `json:"isIncoming,omitempty" yaml:"isIncoming,omitempty"`                 // tr_peer_stat
	IsUploadingTo      bool               `json:"isUploadingTo,omitempty" yaml:"isUploadingTo,omitempty"`           // tr_peer_stat
	IsUTP              bool               `json:"isUTP,omitempty" yaml:"isUTP,omitempty"`                           // tr_peer_stat
	PeerIsChoked       bool               `json:"peerIsChoked,omitempty" yaml:"peerIsChoked,omitempty"`             // tr_peer_stat
	PeerIsInterested   bool               `json:"peerIsInterested,omitempty" yaml:"peerIsInterested,omitempty"`     // tr_peer_stat
	Port               int64              `json:"port,omitempty" yaml:"port,omitempty"`                             // tr_peer_stat
	Progress           transrpc.Percent   `json:"progress,omitempty" yaml:"progress,omitempty"`                     // tr_peer_stat
	RateToClient       transrpc.ByteCount `json:"rateToClient,omitempty" yaml:"rateToClient,omitempty"`             // tr_peer_stat
	RateToPeer         transrpc.ByteCount `json:"rateToPeer,omitempty" yaml:"rateToPeer,omitempty"`                 // tr_peer_stat
	ID                 int64              `json:"id" yaml:"id"`
	HashString         string             `json:"-" yaml:"-"`
	ShortHash          string             `json:"-" yaml:"-"`
}

// doPeersGet is the high-level entry point for 'peers get'.
func doPeersGet(args *Args) error {
	cl, torrents, err := findTorrents(args)
	if err != nil {
		return err
	}
	var result []peer
	if len(torrents) != 0 {
		res, err := transrpc.TorrentGet(convTorrentIDs(torrents)...).WithFields("hashString", "peers").Do(context.Background(), cl)
		if err != nil {
			return err
		}
		for _, t := range res.Torrents {
			for i, v := range t.Peers {
				result = append(result, peer{
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
					HashString:         t.HashString,
					ShortHash:          t.ShortHash(),
				})
			}
		}
	}
	return NewResult(result, args.ResultOptions(
		TableColumns("address", "clientName", "rateToClient", "rateToPeer", "progress", "shortHash"),
		WideColumns("address", "clientName", "isEncrypted", "port", "rateToClient", "rateToPeer", "progress", "shortHash"),
		YamlName("peers"),
		FlatName("peers"),
		FlatKey("id"),
		FlatIndex("shortHash"),
	)...).Encode(os.Stdout)
}

// file is combined fields of files, fileStats from a torrent.
type file struct {
	BytesCompleted transrpc.ByteCount `json:"bytesCompleted,omitempty" yaml:"bytesCompleted,omitempty"` // tr_torrent
	Length         transrpc.ByteCount `json:"length,omitempty" yaml:"length,omitempty"`                 // tr_info
	Name           string             `json:"name,omitempty" yaml:"name,omitempty"`                     // tr_info
	Wanted         bool               `json:"wanted,omitempty" yaml:"wanted,omitempty"`                 // tr_info
	Priority       transrpc.Priority  `json:"priority,omitempty" yaml:"priority,omitempty"`             // tr_info
	ID             int64              `json:"id" yaml:"id"`
	HashString     string             `json:"-" yaml:"-"`
	ShortHash      string             `json:"-" yaml:"-"`
}

// PercentDone provides the
func (f file) PercentDone() transrpc.Percent {
	if f.Length == 0 {
		return 1.0
	}
	return transrpc.Percent(float64(f.BytesCompleted) / float64(f.Length))
}

// doFilesGet is the high-level entry point for 'files get'.
func doFilesGet(args *Args) error {
	cl, torrents, err := findTorrents(args)
	if err != nil {
		return err
	}
	var result []file
	if len(torrents) != 0 {
		res, err := transrpc.TorrentGet(convTorrentIDs(torrents)...).WithFields("hashString", "files", "fileStats").Do(context.Background(), cl)
		if err != nil {
			return err
		}
		for _, t := range res.Torrents {
			for i, v := range t.Files {
				result = append(result, file{
					BytesCompleted: v.BytesCompleted,
					Length:         v.Length,
					Name:           v.Name,
					Wanted:         t.FileStats[i].Wanted,
					Priority:       t.FileStats[i].Priority,
					ID:             int64(i),
					HashString:     t.HashString,
					ShortHash:      t.ShortHash(),
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

// doTorrentSet generates a torrent set request to change the specified field.
func doFilesSet(field string) func(args *Args) error {
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
		cl, torrents, err := findTorrents(args)
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

// doFilesSetPriority is the high-level entry point for 'files set-priority'.
func doFilesSetPriority(args *Args) error {
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

// tracker is combined fields of trackers, trackerStats from a torrent.
type tracker struct {
	Announce              string         `json:"announce,omitempty" yaml:"announce,omitempty"`                           // tr_tracker_info
	ID                    int64          `json:"id" yaml:"id"`                                                           // tr_tracker_info
	Scrape                string         `json:"scrape,omitempty" yaml:"scrape,omitempty"`                               // tr_tracker_info
	Tier                  int64          `json:"tier,omitempty" yaml:"tier,omitempty"`                                   // tr_tracker_info
	AnnounceState         transrpc.State `json:"announceState,omitempty" yaml:"announceState,omitempty"`                 // tr_tracker_stat
	DownloadCount         int64          `json:"downloadCount,omitempty" yaml:"downloadCount,omitempty"`                 // tr_tracker_stat
	HasAnnounced          bool           `json:"hasAnnounced,omitempty" yaml:"hasAnnounced,omitempty"`                   // tr_tracker_stat
	HasScraped            bool           `json:"hasScraped,omitempty" yaml:"hasScraped,omitempty"`                       // tr_tracker_stat
	Host                  string         `json:"host,omitempty" yaml:"host,omitempty"`                                   // tr_tracker_stat
	IsBackup              bool           `json:"isBackup,omitempty" yaml:"isBackup,omitempty"`                           // tr_tracker_stat
	LastAnnouncePeerCount int64          `json:"lastAnnouncePeerCount,omitempty" yaml:"lastAnnouncePeerCount,omitempty"` // tr_tracker_stat
	LastAnnounceResult    string         `json:"lastAnnounceResult,omitempty" yaml:"lastAnnounceResult,omitempty"`       // tr_tracker_stat
	LastAnnounceStartTime transrpc.Time  `json:"lastAnnounceStartTime,omitempty" yaml:"lastAnnounceStartTime,omitempty"` // tr_tracker_stat
	LastAnnounceSucceeded bool           `json:"lastAnnounceSucceeded,omitempty" yaml:"lastAnnounceSucceeded,omitempty"` // tr_tracker_stat
	LastAnnounceTime      transrpc.Time  `json:"lastAnnounceTime,omitempty" yaml:"lastAnnounceTime,omitempty"`           // tr_tracker_stat
	LastAnnounceTimedOut  bool           `json:"lastAnnounceTimedOut,omitempty" yaml:"lastAnnounceTimedOut,omitempty"`   // tr_tracker_stat
	LastScrapeResult      string         `json:"lastScrapeResult,omitempty" yaml:"lastScrapeResult,omitempty"`           // tr_tracker_stat
	LastScrapeStartTime   transrpc.Time  `json:"lastScrapeStartTime,omitempty" yaml:"lastScrapeStartTime,omitempty"`     // tr_tracker_stat
	LastScrapeSucceeded   bool           `json:"lastScrapeSucceeded,omitempty" yaml:"lastScrapeSucceeded,omitempty"`     // tr_tracker_stat
	LastScrapeTime        transrpc.Time  `json:"lastScrapeTime,omitempty" yaml:"lastScrapeTime,omitempty"`               // tr_tracker_stat
	LastScrapeTimedOut    int64          `json:"lastScrapeTimedOut,omitempty" yaml:"lastScrapeTimedOut,omitempty"`       // tr_tracker_stat
	LeecherCount          int64          `json:"leecherCount,omitempty" yaml:"leecherCount,omitempty"`                   // tr_tracker_stat
	NextAnnounceTime      transrpc.Time  `json:"nextAnnounceTime,omitempty" yaml:"nextAnnounceTime,omitempty"`           // tr_tracker_stat
	NextScrapeTime        transrpc.Time  `json:"nextScrapeTime,omitempty" yaml:"nextScrapeTime,omitempty"`               // tr_tracker_stat
	ScrapeState           transrpc.State `json:"scrapeState,omitempty" yaml:"scrapeState,omitempty"`                     // tr_tracker_stat
	SeederCount           int64          `json:"seederCount,omitempty" yaml:"seederCount,omitempty"`                     // tr_tracker_stat
	HashString            string         `json:"-" yaml:"-"`
	ShortHash             string         `json:"-" yaml:"-"`
}

// doTrackersGet is the high-level entry point for 'trackers get'.
func doTrackersGet(args *Args) error {
	cl, torrents, err := findTorrents(args)
	if err != nil {
		return err
	}
	var result []tracker
	if len(torrents) != 0 {
		res, err := transrpc.TorrentGet(convTorrentIDs(torrents)...).WithFields("hashString", "trackers", "trackerStats").Do(context.Background(), cl)
		if err != nil {
			return err
		}
		for _, t := range res.Torrents {
			for i, v := range t.Trackers {
				result = append(result, tracker{
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
					HashString:            t.HashString,
					ShortHash:             t.ShortHash(),
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
