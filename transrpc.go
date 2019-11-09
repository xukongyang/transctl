// Package transrpc provides a client for the Transmission RPC.
//
// See: https://github.com/transmission/transmission/blob/master/extras/rpc-spec.txt
package transrpc

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

// Priority are file priorities.
type Priority int64

// Priorities.
const (
	PriorityLow    Priority = -1
	PriorityNormal Priority = 0
	PriorityHigh   Priority = 1
)

// String satisfies the fmt.Stringer interface.
func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "Low"
	case PriorityNormal:
		return "Normal"
	case PriorityHigh:
		return "High"
	}
	return fmt.Sprintf("Priority(%d)", p)
}

// UnmarshalJSON satisfies the json.Unmarshaler interface.
func (p *Priority) UnmarshalJSON(buf []byte) error {
	if len(buf) == 0 {
		return ErrInvalidPriority
	}
	i, err := strconv.ParseInt(string(buf), 10, 64)
	if err != nil {
		return err
	}
	switch x := Priority(i); x {
	case PriorityLow, PriorityNormal, PriorityHigh:
		*p = x
		return nil
	}
	return ErrInvalidPriority
}

// Time wraps time.Time.
type Time time.Time

// UnmarshalJSON satisfies the json.Unmarshaler interface.
func (t *Time) UnmarshalJSON(buf []byte) error {
	if len(buf) == 0 {
		return ErrInvalidTime
	}
	i, err := strconv.ParseInt(string(buf), 10, 64)
	if err != nil {
		return err
	}
	if i == 0 {
		return nil
	}
	*t = Time(time.Unix(i, 0))
	return nil
}

// Bool wraps int64.
type Bool bool

// UnmarshalJSON satisfies the json.Unmarshaler interface.
func (b *Bool) UnmarshalJSON(buf []byte) error {
	if len(buf) == 0 {
		return ErrInvalidBool
	}

	// check if string
	if s := string(buf); s == "false" || s == "true" {
		v, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		*b = Bool(v)
		return nil
	}

	// true when != 0
	i, err := strconv.ParseInt(string(buf), 10, 64)
	if err != nil {
		return err
	}
	if i != 0 {
		*b = true
	}
	return nil
}

// Torrent holds information about a torrent.
type Torrent struct {
	ActivityDate      Time   `json:"activityDate,omitempty"`      // tr_stat
	AddedDate         Time   `json:"addedDate,omitempty"`         // tr_stat
	BandwidthPriority int64  `json:"bandwidthPriority,omitempty"` // tr_priority_t
	Comment           string `json:"comment,omitempty"`           // tr_info
	CorruptEver       int64  `json:"corruptEver,omitempty"`       // tr_stat
	Creator           string `json:"creator,omitempty"`           // tr_info
	DateCreated       Time   `json:"dateCreated,omitempty"`       // tr_info
	DesiredAvailable  int64  `json:"desiredAvailable,omitempty"`  // tr_stat
	DoneDate          Time   `json:"doneDate,omitempty"`          // tr_stat
	DownloadDir       string `json:"downloadDir,omitempty"`       // tr_torrent
	DownloadedEver    int64  `json:"downloadedEver,omitempty"`    // tr_stat
	DownloadLimit     int64  `json:"downloadLimit,omitempty"`     // tr_torrent
	DownloadLimited   bool   `json:"downloadLimited,omitempty"`   // tr_torrent
	Error             int64  `json:"error,omitempty"`             // tr_stat
	ErrorString       string `json:"errorString,omitempty"`       // tr_stat
	Eta               int64  `json:"eta,omitempty"`               // tr_stat
	EtaIdle           int64  `json:"etaIdle,omitempty"`           // tr_stat
	Files             []struct {
		BytesCompleted int64  `json:"bytesCompleted,omitempty"` // tr_torrent
		Length         int64  `json:"length,omitempty"`         // tr_info
		Name           string `json:"name,omitempty"`           // tr_info
	} `json:"files,omitempty"` // n/a
	FileStats []struct {
		BytesCompleted int64    `json:"bytesCompleted,omitempty"` // tr_torrent
		Wanted         bool     `json:"wanted,omitempty"`         // tr_info
		Priority       Priority `json:"priority,omitempty"`       // tr_info
	} `json:"fileStats,omitempty"` // n/a
	HashString              string   `json:"hashString,omitempty"`              // tr_info
	HaveUnchecked           int64    `json:"haveUnchecked,omitempty"`           // tr_stat
	HaveValid               int64    `json:"haveValid,omitempty"`               // tr_stat
	HonorsSessionLimits     bool     `json:"honorsSessionLimits,omitempty"`     // tr_torrent
	ID                      int64    `json:"id,omitempty"`                      // tr_torrent
	IsFinished              bool     `json:"isFinished,omitempty"`              // tr_stat
	IsPrivate               bool     `json:"isPrivate,omitempty"`               // tr_torrent
	IsStalled               bool     `json:"isStalled,omitempty"`               // tr_stat
	Labels                  []string `json:"labels,omitempty"`                  // tr_torrent
	LeftUntilDone           int64    `json:"leftUntilDone,omitempty"`           // tr_stat
	MagnetLink              string   `json:"magnetLink,omitempty"`              // n/a
	ManualAnnounceTime      int64    `json:"manualAnnounceTime,omitempty"`      // tr_stat
	MaxConnectedPeers       int64    `json:"maxConnectedPeers,omitempty"`       // tr_torrent
	MetadataPercentComplete float64  `json:"metadataPercentComplete,omitempty"` // tr_stat
	Name                    string   `json:"name,omitempty"`                    // tr_info
	PeerLimit               int64    `json:"peer-limit,omitempty"`              // tr_torrent
	Peers                   []struct {
		Address              string  `json:"address,omitempty"`              // tr_peer_stat
		ClientName           string  `json:"clientName,omitempty"`           // tr_peer_stat
		ClientIsChoked       bool    `json:"clientIsChoked,omitempty"`       // tr_peer_stat
		ClientIsint64erested bool    `json:"clientIsint64erested,omitempty"` // tr_peer_stat
		FlagStr              string  `json:"flagStr,omitempty"`              // tr_peer_stat
		IsDownloadingFrom    bool    `json:"isDownloadingFrom,omitempty"`    // tr_peer_stat
		IsEncrypted          bool    `json:"isEncrypted,omitempty"`          // tr_peer_stat
		IsIncoming           bool    `json:"isIncoming,omitempty"`           // tr_peer_stat
		IsUploadingTo        bool    `json:"isUploadingTo,omitempty"`        // tr_peer_stat
		IsUTP                bool    `json:"isUTP,omitempty"`                // tr_peer_stat
		PeerIsChoked         bool    `json:"peerIsChoked,omitempty"`         // tr_peer_stat
		PeerIsint64erested   bool    `json:"peerIsint64erested,omitempty"`   // tr_peer_stat
		Port                 int64   `json:"port,omitempty"`                 // tr_peer_stat
		Progress             float64 `json:"progress,omitempty"`             // tr_peer_stat
		RateToClient         int64   `json:"rateToClient,omitempty"`         // tr_peer_stat
		RateToPeer           int64   `json:"rateToPeer,omitempty"`           // tr_peer_stat
	} `json:"peers,omitempty"` // n/a
	PeersConnected int64 `json:"peersConnected,omitempty"` // tr_stat
	PeersFrom      struct {
		FromCache    int64 `json:"fromCache,omitempty"`    // tr_stat
		FromDht      int64 `json:"fromDht,omitempty"`      // tr_stat
		FromIncoming int64 `json:"fromIncoming,omitempty"` // tr_stat
		FromLpd      int64 `json:"fromLpd,omitempty"`      // tr_stat
		FromLtep     int64 `json:"fromLtep,omitempty"`     // tr_stat
		FromPex      int64 `json:"fromPex,omitempty"`      // tr_stat
		FromTracker  int64 `json:"fromTracker,omitempty"`  // tr_stat
	} `json:"peersFrom,omitempty"` // n/a
	PeersGettingFromUs int64      `json:"peersGettingFromUs,omitempty"` // tr_stat
	PeersSendingToUs   int64      `json:"peersSendingToUs,omitempty"`   // tr_stat
	PercentDone        float64    `json:"percentDone,omitempty"`        // tr_stat
	Pieces             []byte     `json:"pieces,omitempty"`             // tr_torrent
	PieceCount         int64      `json:"pieceCount,omitempty"`         // tr_info
	PieceSize          int64      `json:"pieceSize,omitempty"`          // tr_info
	Priorities         []Priority `json:"priorities,omitempty"`         // n/a
	QueuePosition      int64      `json:"queuePosition,omitempty"`      // tr_stat
	RateDownload       int64      `json:"rateDownload,omitempty"`       // tr_stat
	RateUpload         int64      `json:"rateUpload,omitempty"`         // tr_stat
	RecheckProgress    float64    `json:"recheckProgress,omitempty"`    // tr_stat
	SecondsDownloading int64      `json:"secondsDownloading,omitempty"` // tr_stat
	SecondsSeeding     int64      `json:"secondsSeeding,omitempty"`     // tr_stat
	SeedIdleLimit      int64      `json:"seedIdleLimit,omitempty"`      // tr_torrent
	SeedIdleMode       int64      `json:"seedIdleMode,omitempty"`       // tr_inactvelimit
	SeedRatioLimit     float64    `json:"seedRatioLimit,omitempty"`     // tr_torrent
	SeedRatioMode      int64      `json:"seedRatioMode,omitempty"`      // tr_ratiolimit
	SizeWhenDone       int64      `json:"sizeWhenDone,omitempty"`       // tr_stat
	StartDate          Time       `json:"startDate,omitempty"`          // tr_stat
	Status             int64      `json:"status,omitempty"`             // tr_stat
	Trackers           []struct {
		Announce string `json:"announce,omitempty"` // tr_tracker_info
		ID       int64  `json:"id,omitempty"`       // tr_tracker_info
		Scrape   string `json:"scrape,omitempty"`   // tr_tracker_info
		Tier     int64  `json:"tier,omitempty"`     // tr_tracker_info
	} `json:"trackers,omitempty"` // n/a
	TrackerStats []struct {
		Announce              string `json:"announce,omitempty"`              // tr_tracker_stat
		AnnounceState         int64  `json:"announceState,omitempty"`         // tr_tracker_stat
		DownloadCount         int64  `json:"downloadCount,omitempty"`         // tr_tracker_stat
		HasAnnounced          bool   `json:"hasAnnounced,omitempty"`          // tr_tracker_stat
		HasScraped            bool   `json:"hasScraped,omitempty"`            // tr_tracker_stat
		Host                  string `json:"host,omitempty"`                  // tr_tracker_stat
		ID                    int64  `json:"id,omitempty"`                    // tr_tracker_stat
		IsBackup              bool   `json:"isBackup,omitempty"`              // tr_tracker_stat
		LastAnnouncePeerCount int64  `json:"lastAnnouncePeerCount,omitempty"` // tr_tracker_stat
		LastAnnounceResult    string `json:"lastAnnounceResult,omitempty"`    // tr_tracker_stat
		LastAnnounceStartTime Time   `json:"lastAnnounceStartTime,omitempty"` // tr_tracker_stat
		LastAnnounceSucceeded bool   `json:"lastAnnounceSucceeded,omitempty"` // tr_tracker_stat
		LastAnnounceTime      Time   `json:"lastAnnounceTime,omitempty"`      // tr_tracker_stat
		LastAnnounceTimedOut  bool   `json:"lastAnnounceTimedOut,omitempty"`  // tr_tracker_stat
		LastScrapeResult      string `json:"lastScrapeResult,omitempty"`      // tr_tracker_stat
		LastScrapeStartTime   Time   `json:"lastScrapeStartTime,omitempty"`   // tr_tracker_stat
		LastScrapeSucceeded   bool   `json:"lastScrapeSucceeded,omitempty"`   // tr_tracker_stat
		LastScrapeTime        Time   `json:"lastScrapeTime,omitempty"`        // tr_tracker_stat
		LastScrapeTimedOut    int64  `json:"lastScrapeTimedOut,omitempty"`    // tr_tracker_stat
		LeecherCount          int64  `json:"leecherCount,omitempty"`          // tr_tracker_stat
		NextAnnounceTime      Time   `json:"nextAnnounceTime,omitempty"`      // tr_tracker_stat
		NextScrapeTime        Time   `json:"nextScrapeTime,omitempty"`        // tr_tracker_stat
		Scrape                string `json:"scrape,omitempty"`                // tr_tracker_stat
		ScrapeState           int64  `json:"scrapeState,omitempty"`           // tr_tracker_stat
		SeederCount           int64  `json:"seederCount,omitempty"`           // tr_tracker_stat
		Tier                  int64  `json:"tier,omitempty"`                  // tr_tracker_stat
	} `json:"trackerStats,omitempty"` // n/a
	TotalSize           int64    `json:"totalSize,omitempty"`           // tr_info
	TorrentFile         string   `json:"torrentFile,omitempty"`         // tr_info
	UploadedEver        int64    `json:"uploadedEver,omitempty"`        // tr_stat
	UploadLimit         int64    `json:"uploadLimit,omitempty"`         // tr_torrent
	UploadLimited       bool     `json:"uploadLimited,omitempty"`       // tr_torrent
	UploadRatio         float64  `json:"uploadRatio,omitempty"`         // tr_stat
	Wanted              []Bool   `json:"wanted,omitempty"`              // n/a
	Webseeds            []string `json:"webseeds,omitempty"`            // n/a
	WebseedsSendingToUs int64    `json:"webseedsSendingToUs,omitempty"` // tr_stat
}

// Request is a generic request used when working with a list of torrent
// identifiers.
type Request struct {
	method string        // rpc request method to call
	ids    []interface{} // the torrent torrent list, as described in 3.1
}

// NewRequest creates a generic request for a list of torrent
// identifiers and the named method.
//
// Passed IDs can be any of type of int{,8,16,32,64}, []byte, or string.
//
// Used for torrent-{start,start-now,stop,verify,reannounce} methods.
func NewRequest(method string, ids ...interface{}) *Request {
	return &Request{method: method, ids: ids}
}

// Do executes the torrent request using the provided context and client.
func (req *Request) Do(ctx context.Context, cl *Client) error {
	ids, err := checkIdentifierList(req.ids...)
	if err != nil {
		return err
	}
	return cl.Do(ctx, req.method, map[string]interface{}{
		"ids": ids,
	}, nil)
}

// TorrentStartRequest is a torrent start request.
type TorrentStartRequest = Request

// TorrentStart creates a torrent start request for the specified ids.
func TorrentStart(ids ...interface{}) *TorrentStartRequest {
	return NewRequest("torrent-start", ids...)
}

// TorrentStartNowRequest is the torrent start now request.
type TorrentStartNowRequest = Request

// TorrentStartNow creates a torrent start now request for the specified ids.
func TorrentStartNow(ids ...interface{}) *TorrentStartNowRequest {
	return NewRequest("torrent-start-now", ids...)
}

// TorrentStopRequest is the torrent stop request.
type TorrentStopRequest = Request

// TorrentStop creates a torrent stop request for the specified ids.
func TorrentStop(ids ...interface{}) *TorrentStopRequest {
	return NewRequest("torrent-stop", ids...)
}

// TorrentVerifyRequest is the torrent verify request.
type TorrentVerifyRequest = Request

// TorrentVerify creates a torrent verify request for the specified ids
func TorrentVerify(ids ...interface{}) *TorrentVerifyRequest {
	return NewRequest("torrent-verify", ids...)
}

// TorrentReannounceRequest is a torrent reannounce request.
type TorrentReannounceRequest = Request

// TorrentReannounce creates a torrent reannounce request for the specified ids.
func TorrentReannounce(ids ...interface{}) *TorrentReannounceRequest {
	return NewRequest("torrent-reannounce", ids...)
}

// TorrentSetRequest is a torrent set request.
type TorrentSetRequest struct {
	changed             map[string]bool `json:"-"`                             // fields marked as changed
	BandwidthPriority   int64           `json:"bandwidthPriority,omitempty"`   // this torrent's bandwidth tr_priority_t
	DownloadLimit       int64           `json:"downloadLimit,omitempty"`       // maximum download speed (KBps)
	DownloadLimited     bool            `json:"downloadLimited,omitempty"`     // true if downloadLimit is honored
	FilesWanted         []int64         `json:"files-wanted,omitempty"`        // indices of file(s) to download
	FilesUnwanted       []int64         `json:"files-unwanted,omitempty"`      // indices of file(s) to not download
	HonorsSessionLimits bool            `json:"honorsSessionLimits,omitempty"` // true if session upload limits are honored
	IDs                 []interface{}   `json:"ids,omitempty"`                 // torrent list, as described in 3.1
	Labels              []string        `json:"labels,omitempty"`              // array of string labels
	Location            string          `json:"location,omitempty"`            // new location of the torrent's content
	PeerLimit           int64           `json:"peer-limit,omitempty"`          // maximum int64 of peers
	PriorityHigh        []int64         `json:"priority-high,omitempty"`       // indices of high-priority file(s)
	PriorityLow         []int64         `json:"priority-low,omitempty"`        // indices of low-priority file(s)
	PriorityNormal      []int64         `json:"priority-normal,omitempty"`     // indices of normal-priority file(s)
	QueuePosition       int64           `json:"queuePosition,omitempty"`       // position of this torrent in its queue [0...n)
	SeedIdleLimit       int64           `json:"seedIdleLimit,omitempty"`       // torrent-level int64 of minutes of seeding inactivity
	SeedIdleMode        int64           `json:"seedIdleMode,omitempty"`        // which seeding inactivity to use.  See tr_idlelimit
	SeedRatioLimit      float64         `json:"seedRatioLimit,omitempty"`      // torrent-level seeding ratio
	SeedRatioMode       int64           `json:"seedRatioMode,omitempty"`       // which ratio to use.  See tr_ratiolimit
	TrackerAdd          []string        `json:"trackerAdd,omitempty"`          // strings of announce URLs to add
	TrackerRemove       []int64         `json:"trackerRemove,omitempty"`       // ids of trackers to remove
	TrackerReplace      []string        `json:"trackerReplace,omitempty"`      // pairs of <trackerId/new announce URLs>
	UploadLimit         int64           `json:"uploadLimit,omitempty"`         // maximum upload speed (KBps)
	UploadLimited       bool            `json:"uploadLimited,omitempty"`       // true if uploadLimit is honored
}

// TorrentSet creates a torrent set request.
func TorrentSet(ids ...interface{}) *TorrentSetRequest {
	return &TorrentSetRequest{
		changed: make(map[string]bool),
		IDs:     ids,
	}
}

// Do executes the torrent set request using the provided context and client.
func (req *TorrentSetRequest) Do(ctx context.Context, cl *Client) error {
	params := make(map[string]interface{})

	if req.changed["BandwidthPriority"] {
		params["bandwidthPriority"] = req.BandwidthPriority
	}
	if req.changed["DownloadLimit"] {
		params["downloadLimit"] = req.DownloadLimit
	}
	if req.changed["DownloadLimited"] {
		params["downloadLimited"] = req.DownloadLimited
	}
	if req.changed["FilesWanted"] {
		params["files-wanted"] = req.FilesWanted
	}
	if req.changed["FilesUnwanted"] {
		params["files-unwanted"] = req.FilesUnwanted
	}
	if req.changed["HonorsSessionLimits"] {
		params["honorsSessionLimits"] = req.HonorsSessionLimits
	}
	if req.changed["IDs"] {
		params["ids"] = req.IDs
	}
	if req.changed["Labels"] {
		params["labels"] = req.Labels
	}
	if req.changed["Location"] {
		params["location"] = req.Location
	}
	if req.changed["PeerLimit"] {
		params["peer-limit"] = req.PeerLimit
	}
	if req.changed["PriorityHigh"] {
		params["priority-high"] = req.PriorityHigh
	}
	if req.changed["PriorityLow"] {
		params["priority-low"] = req.PriorityLow
	}
	if req.changed["PriorityNormal"] {
		params["priority-normal"] = req.PriorityNormal
	}
	if req.changed["QueuePosition"] {
		params["queuePosition"] = req.QueuePosition
	}
	if req.changed["SeedIdleLimit"] {
		params["seedIdleLimit"] = req.SeedIdleLimit
	}
	if req.changed["SeedIdleMode"] {
		params["seedIdleMode"] = req.SeedIdleMode
	}
	if req.changed["SeedRatioLimit"] {
		params["seedRatioLimit"] = req.SeedRatioLimit
	}
	if req.changed["SeedRatioMode"] {
		params["seedRatioMode"] = req.SeedRatioMode
	}
	if req.changed["TrackerAdd"] {
		params["trackerAdd"] = req.TrackerAdd
	}
	if req.changed["TrackerRemove"] {
		params["trackerRemove"] = req.TrackerRemove
	}
	if req.changed["TrackerReplace"] {
		params["trackerReplace"] = req.TrackerReplace
	}
	if req.changed["UploadLimit"] {
		params["uploadLimit"] = req.UploadLimit
	}
	if req.changed["UploadLimited"] {
		params["uploadLimited"] = req.UploadLimited
	}

	// check identifiers
	if v, ok := params["ids"]; ok {
		ids, err := checkIdentifierList(v.([]interface{})...)
		if err != nil {
			return err
		}
		params["ids"] = ids
	}

	return cl.Do(ctx, "torrent-set", params, nil)
}

// WithChanged marks the fields that were changed.
func (req TorrentSetRequest) WithChanged(fields ...string) *TorrentSetRequest {
	for _, field := range fields {
		req.changed[field] = true
	}
	return &req
}

// WithBandwidthPriority sets this torrent's bandwidth tr_priority_t.
func (req TorrentSetRequest) WithBandwidthPriority(bandwidthPriority int64) *TorrentSetRequest {
	req.BandwidthPriority = bandwidthPriority
	return req.WithChanged("BandwidthPriority")
}

// WithDownloadLimit sets maximum download speed (KBps).
func (req TorrentSetRequest) WithDownloadLimit(downloadLimit int64) *TorrentSetRequest {
	req.DownloadLimit = downloadLimit
	return req.WithChanged("DownloadLimit")
}

// WithDownloadLimited sets true if downloadLimit is honored.
func (req TorrentSetRequest) WithDownloadLimited(downloadLimited bool) *TorrentSetRequest {
	req.DownloadLimited = downloadLimited
	return req.WithChanged("DownloadLimited")
}

// WithFilesWanted sets indices of file(s) to download.
func (req TorrentSetRequest) WithFilesWanted(filesWanted []int64) *TorrentSetRequest {
	req.FilesWanted = filesWanted
	return req.WithChanged("FilesWanted")
}

// WithFilesUnwanted sets indices of file(s) to not download.
func (req TorrentSetRequest) WithFilesUnwanted(filesUnwanted []int64) *TorrentSetRequest {
	req.FilesUnwanted = filesUnwanted
	return req.WithChanged("FilesUnwanted")
}

// WithHonorsSessionLimits sets true if session upload limits are honored.
func (req TorrentSetRequest) WithHonorsSessionLimits(honorsSessionLimits bool) *TorrentSetRequest {
	req.HonorsSessionLimits = honorsSessionLimits
	return req.WithChanged("HonorsSessionLimits")
}

// WithIDs sets torrent list, as described in 3.1.
func (req TorrentSetRequest) WithIDs(ids ...interface{}) *TorrentSetRequest {
	req.IDs = ids
	return req.WithChanged("IDs")
}

// WithLabels sets array of string labels.
func (req TorrentSetRequest) WithLabels(labels []string) *TorrentSetRequest {
	req.Labels = labels
	return req.WithChanged("Labels")
}

// WithLocation sets new location of the torrent's content.
func (req TorrentSetRequest) WithLocation(location string) *TorrentSetRequest {
	req.Location = location
	return req.WithChanged("Location")
}

// WithPeerLimit sets maximum int64 of peers.
func (req TorrentSetRequest) WithPeerLimit(peerLimit int64) *TorrentSetRequest {
	req.PeerLimit = peerLimit
	return req.WithChanged("PeerLimit")
}

// WithPriorityHigh sets indices of high-priority file(s).
func (req TorrentSetRequest) WithPriorityHigh(priorityHigh []int64) *TorrentSetRequest {
	req.PriorityHigh = priorityHigh
	return req.WithChanged("PriorityHigh")
}

// WithPriorityLow sets indices of low-priority file(s).
func (req TorrentSetRequest) WithPriorityLow(priorityLow []int64) *TorrentSetRequest {
	req.PriorityLow = priorityLow
	return req.WithChanged("PriorityLow")
}

// WithPriorityNormal sets indices of normal-priority file(s).
func (req TorrentSetRequest) WithPriorityNormal(priorityNormal []int64) *TorrentSetRequest {
	req.PriorityNormal = priorityNormal
	return req.WithChanged("PriorityNormal")
}

// WithQueuePosition sets position of this torrent in its queue [0...n).
func (req TorrentSetRequest) WithQueuePosition(queuePosition int64) *TorrentSetRequest {
	req.QueuePosition = queuePosition
	return req.WithChanged("QueuePosition")
}

// WithSeedIdleLimit sets torrent-level int64 of minutes of seeding inactivity.
func (req TorrentSetRequest) WithSeedIdleLimit(seedIdleLimit int64) *TorrentSetRequest {
	req.SeedIdleLimit = seedIdleLimit
	return req.WithChanged("SeedIdleLimit")
}

// WithSeedIdleMode sets which seeding inactivity to use.  See tr_idlelimit.
func (req TorrentSetRequest) WithSeedIdleMode(seedIdleMode int64) *TorrentSetRequest {
	req.SeedIdleMode = seedIdleMode
	return req.WithChanged("SeedIdleMode")
}

// WithSeedRatioLimit sets torrent-level seeding ratio.
func (req TorrentSetRequest) WithSeedRatioLimit(seedRatioLimit float64) *TorrentSetRequest {
	req.SeedRatioLimit = seedRatioLimit
	return req.WithChanged("SeedRatioLimit")
}

// WithSeedRatioMode sets which ratio to use.  See tr_ratiolimit.
func (req TorrentSetRequest) WithSeedRatioMode(seedRatioMode int64) *TorrentSetRequest {
	req.SeedRatioMode = seedRatioMode
	return req.WithChanged("SeedRatioMode")
}

// WithTrackerAdd sets strings of announce URLs to add.
func (req TorrentSetRequest) WithTrackerAdd(trackerAdd []string) *TorrentSetRequest {
	req.TrackerAdd = trackerAdd
	return req.WithChanged("TrackerAdd")
}

// WithTrackerRemove sets ids of trackers to remove.
func (req TorrentSetRequest) WithTrackerRemove(trackerRemove []int64) *TorrentSetRequest {
	req.TrackerRemove = trackerRemove
	return req.WithChanged("TrackerRemove")
}

// WithTrackerReplace sets pairs of <trackerId/new announce URLs>.
func (req TorrentSetRequest) WithTrackerReplace(trackerReplace []string) *TorrentSetRequest {
	req.TrackerReplace = trackerReplace
	return req.WithChanged("TrackerReplace")
}

// WithUploadLimit sets maximum upload speed (KBps).
func (req TorrentSetRequest) WithUploadLimit(uploadLimit int64) *TorrentSetRequest {
	req.UploadLimit = uploadLimit
	return req.WithChanged("UploadLimit")
}

// WithUploadLimited sets true if uploadLimit is honored.
func (req TorrentSetRequest) WithUploadLimited(uploadLimited bool) *TorrentSetRequest {
	req.UploadLimited = uploadLimited
	return req.WithChanged("UploadLimited")
}

// TorrentGetRequest is the torrent get request.
type TorrentGetRequest struct {
	ids    []interface{} // An optional "ids" array as described in 3.1
	fields []string      // A required "fields" array of keys
}

// TorrentGet creates a torrent get request for the specified torrent ids.
func TorrentGet(ids ...interface{}) *TorrentGetRequest {
	return &TorrentGetRequest{
		ids: ids,
		fields: []string{
			"activityDate", "addedDate", "bandwidthPriority",
			"comment", "corruptEver", "creator",
			"dateCreated", "desiredAvailable", "doneDate",
			"downloadDir", "downloadedEver", "downloadLimit",
			"downloadLimited", "error", "errorString",
			"eta", "etaIdle", "files",
			"fileStats", "hashString", "haveUnchecked",
			"haveValid", "honorsSessionLimits", "id",
			"isFinished", "isPrivate", "isStalled",
			"labels", "leftUntilDone", "magnetLink",
			"manualAnnounceTime", "maxConnectedPeers", "metadataPercentComplete",
			"name", "peer-limit", "peers",
			"peersConnected", "peersFrom", "peersGettingFromUs",
			"peersSendingToUs", "percentDone", "pieces",
			"pieceCount", "pieceSize", "priorities",
			"queuePosition", "rateDownload", "rateUpload",
			"recheckProgress", "secondsDownloading", "secondsSeeding",
			"seedIdleLimit", "seedIdleMode", "seedRatioLimit",
			"seedRatioMode", "sizeWhenDone", "startDate",
			"status", "trackers", "trackerStats",
			"totalSize", "torrentFile", "uploadedEver",
			"uploadLimit", "uploadLimited", "uploadRatio",
			"wanted", "webseeds", "webseedsSendingToUs",
		},
	}
}

// Do executes the torrent get request against the provided context and client.
func (req *TorrentGetRequest) Do(ctx context.Context, cl *Client) (*TorrentGetResponse, error) {
	ids, err := checkIdentifierList(req.ids...)
	if err != nil {
		return nil, err
	}
	res := new(TorrentGetResponse)
	if err := cl.Do(ctx, "torrent-get", map[string]interface{}{
		"ids":    ids,
		"fields": req.fields,
	}, res); err != nil {
		return nil, err
	}
	return res, nil
}

// TorrentGetResponse is the torrent get response.
type TorrentGetResponse struct {
	Torrents []Torrent     `json:"torrents,omitempty"` // contains the key/value pairs matching the request's "fields" argument
	Removed  []interface{} `json:"removed,omitempty"`  // populated when the requested id was "recently-active"
}

// TorrentAddRequest is the torrent add request.
type TorrentAddRequest struct {
	Cookies           string  `json:"cookies,omitempty"`           // pointer to a string of one or more cookies.
	DownloadDir       string  `json:"download-dir,omitempty"`      // path to download the torrent to
	Filename          string  `json:"filename,omitempty"`          // filename or URL of the .torrent file
	Metainfo          []byte  `json:"metainfo,omitempty"`          // base64-encoded .torrent content
	Paused            bool    `json:"paused,omitempty"`            // if true, don't start the torrent
	PeerLimit         int64   `json:"peer-limit,omitempty"`        // maximum int64 of peers
	BandwidthPriority int64   `json:"bandwidthPriority,omitempty"` // torrent's bandwidth tr_priority_t
	FilesWanted       []int64 `json:"files-wanted,omitempty"`      // indices of file(s) to download
	FilesUnwanted     []int64 `json:"files-unwanted,omitempty"`    // indices of file(s) to not download
	PriorityHigh      []int64 `json:"priority-high,omitempty"`     // indices of high-priority file(s)
	PriorityLow       []int64 `json:"priority-low,omitempty"`      // indices of low-priority file(s)
	PriorityNormal    []int64 `json:"priority-normal,omitempty"`   // indices of normal-priority file(s)
}

// TorrentAdd creates a torrent add request.
func TorrentAdd() *TorrentAddRequest {
	return &TorrentAddRequest{}
}

// Do executes the torrent add request against the provided context and client.
func (req *TorrentAddRequest) Do(ctx context.Context, cl *Client) (*TorrentAddResponse, error) {
	res := new(TorrentAddResponse)
	if err := cl.Do(ctx, "torrent-add", req, res); err != nil {
		return nil, err
	}
	return res, nil
}

// WithCookies sets pointer to a string of one or more cookies..
func (req TorrentAddRequest) WithCookies(cookies string) *TorrentAddRequest {
	req.Cookies = cookies
	return &req
}

// WithDownloadDir sets path to download the torrent to.
func (req TorrentAddRequest) WithDownloadDir(downloadDir string) *TorrentAddRequest {
	req.DownloadDir = downloadDir
	return &req
}

// WithFilename sets filename or URL of the .torrent file.
func (req TorrentAddRequest) WithFilename(filename string) *TorrentAddRequest {
	req.Filename = filename
	return &req
}

// WithMetainfo sets base64-encoded .torrent content.
func (req TorrentAddRequest) WithMetainfo(metainfo []byte) *TorrentAddRequest {
	req.Metainfo = metainfo
	return &req
}

// WithPaused sets if true, don't start the torrent.
func (req TorrentAddRequest) WithPaused(paused bool) *TorrentAddRequest {
	req.Paused = paused
	return &req
}

// WithPeerLimit sets maximum int64 of peers.
func (req TorrentAddRequest) WithPeerLimit(peerLimit int64) *TorrentAddRequest {
	req.PeerLimit = peerLimit
	return &req
}

// WithBandwidthPriority sets torrent's bandwidth tr_priority_t.
func (req TorrentAddRequest) WithBandwidthPriority(bandwidthPriority int64) *TorrentAddRequest {
	req.BandwidthPriority = bandwidthPriority
	return &req
}

// WithFilesWanted sets indices of file(s) to download.
func (req TorrentAddRequest) WithFilesWanted(filesWanted []int64) *TorrentAddRequest {
	req.FilesWanted = filesWanted
	return &req
}

// WithFilesUnwanted sets indices of file(s) to not download.
func (req TorrentAddRequest) WithFilesUnwanted(filesUnwanted []int64) *TorrentAddRequest {
	req.FilesUnwanted = filesUnwanted
	return &req
}

// WithPriorityHigh sets indices of high-priority file(s).
func (req TorrentAddRequest) WithPriorityHigh(priorityHigh []int64) *TorrentAddRequest {
	req.PriorityHigh = priorityHigh
	return &req
}

// WithPriorityLow sets indices of low-priority file(s).
func (req TorrentAddRequest) WithPriorityLow(priorityLow []int64) *TorrentAddRequest {
	req.PriorityLow = priorityLow
	return &req
}

// WithPriorityNormal sets indices of normal-priority file(s).
func (req TorrentAddRequest) WithPriorityNormal(priorityNormal []int64) *TorrentAddRequest {
	req.PriorityNormal = priorityNormal
	return &req
}

// TorrentAddResponse is the torrent add response.
type TorrentAddResponse struct {
	TorrentAdded     *Torrent `json:"torrent-added"`
	TorrentDuplicate *Torrent `json:"torrent-duplicate"`
}

// TorrentRemoveRequest is the torrent remove request.
type TorrentRemoveRequest struct {
	ids             []interface{} // torrent list, as described in 3.1
	deleteLocalData bool          // delete local data. (default: false)
}

// TorrentRemove creates a torrent remove request.
func TorrentRemove(deleteLocalData bool, ids ...interface{}) *TorrentRemoveRequest {
	return &TorrentRemoveRequest{
		ids:             ids,
		deleteLocalData: deleteLocalData,
	}
}

// Do executes the torrent remove request against the provided context and
// client.
func (req *TorrentRemoveRequest) Do(ctx context.Context, cl *Client) error {
	ids, err := checkIdentifierList(req.ids...)
	if err != nil {
		return err
	}
	return cl.Do(ctx, "torrent-remove", map[string]interface{}{
		"ids":               ids,
		"delete-local-data": req.deleteLocalData,
	}, nil)
}

// WithDeleteLocalData sets delete local data. (default: false).
func (req TorrentRemoveRequest) WithDeleteLocalData(deleteLocalData bool) *TorrentRemoveRequest {
	req.deleteLocalData = deleteLocalData
	return &req
}

// TorrentSetLocationRequest is the torrent set location request.
type TorrentSetLocationRequest struct {
	ids      []interface{} // torrent list, as described in 3.1
	location string        // the new torrent location
	move     bool          // if true, move from previous location. otherwise, search "location" for files (default: false)
}

// TorrentSetLocation creates a torrent set location request.
func TorrentSetLocation(location string, ids ...interface{}) *TorrentSetLocationRequest {
	return &TorrentSetLocationRequest{
		ids:      ids,
		location: location,
	}
}

// Do executes the torrent set location request against the provided context
// and client.
func (req *TorrentSetLocationRequest) Do(ctx context.Context, cl *Client) error {
	ids, err := checkIdentifierList(req.ids...)
	if err != nil {
		return err
	}
	return cl.Do(ctx, "torrent-set-location", map[string]interface{}{
		"ids":      ids,
		"location": req.location,
		"move":     req.move,
	}, nil)
}

// WithMove sets if true, move from previous location. otherwise, search
// "location" for files (default: false).
func (req TorrentSetLocationRequest) WithMove(move bool) *TorrentSetLocationRequest {
	req.move = move
	return &req
}

// TorrentRenamePathRequest is the torrent rename path request.
type TorrentRenamePathRequest struct {
	ids  []interface{} // the torrent torrent list, as described in 3.1 (must only be 1 torrent)
	path string        // the path to the file or folder that will be renamed
	name string        // the file or folder's new name
}

// TorrentRenamePath creates a torrent rename path request.
func TorrentRenamePath(path, name string, ids ...interface{}) *TorrentRenamePathRequest {
	return &TorrentRenamePathRequest{
		ids:  ids,
		path: path,
		name: name,
	}
}

// Do executes the torrent rename path request against the provided context and
// client.
func (req *TorrentRenamePathRequest) Do(ctx context.Context, cl *Client) error {
	if len(req.ids) != 1 {
		return ErrTorrentRenamePathCanOnlyBeUsedWithOneTorrentIdentifier
	}
	ids, err := checkIdentifierList(req.ids...)
	if err != nil {
		return err
	}
	return cl.Do(ctx, "torrent-rename-path", map[string]interface{}{
		"ids":  ids,
		"path": req.path,
		"name": req.name,
	}, nil)
}

// Encryption is the list of encryption types.
type Encryption string

// Encryption types
const (
	EncryptionRequired  Encryption = "required"
	EncryptionPreferred Encryption = "preferred"
	EncryptionTolerated Encryption = "tolerated"
)

// Strings satisfies the fmt.Stringer interface.
func (t Encryption) String() string {
	return string(t)
}

// UnmarshalJSON satisfies the json.Unmarshaler interface.
func (t *Encryption) UnmarshalJSON(buf []byte) error {
	if len(buf) == 0 {
		return ErrInvalidEncryption
	}
	// unquote
	s, err := strconv.Unquote(string(buf))
	if err != nil {
		return err
	}
	switch x := Encryption(s); x {
	case EncryptionRequired, EncryptionPreferred, EncryptionTolerated:
		*t = x
		return nil
	}
	return ErrInvalidEncryption
}

// Session holds transmission rpc session arguments.
type Session struct {
	changed map[string]bool `json:"-"` // fields marked as changed

	AltSpeedDown              int64      `json:"alt-speed-down,omitempty"`               // max global download speed (KBps)
	AltSpeedEnabled           bool       `json:"alt-speed-enabled,omitempty"`            // true means use the alt speeds
	AltSpeedTimeBegin         int64      `json:"alt-speed-time-begin,omitempty"`         // when to turn on alt speeds (units: minutes after midnight)
	AltSpeedTimeEnabled       bool       `json:"alt-speed-time-enabled,omitempty"`       // true means the scheduled on/off times are used
	AltSpeedTimeEnd           int64      `json:"alt-speed-time-end,omitempty"`           // when to turn off alt speeds (units: same)
	AltSpeedTimeDay           int64      `json:"alt-speed-time-day,omitempty"`           // what day(s) to turn on alt speeds (look at tr_sched_day)
	AltSpeedUp                int64      `json:"alt-speed-up,omitempty"`                 // max global upload speed (KBps)
	BlocklistURL              string     `json:"blocklist-url,omitempty"`                // location of the blocklist to use for "blocklist-update"
	BlocklistEnabled          bool       `json:"blocklist-enabled,omitempty"`            // true means enabled
	BlocklistSize             int64      `json:"blocklist-size,omitempty"`               // number of rules in the blocklist
	CacheSizeMb               int64      `json:"cache-size-mb,omitempty"`                // maximum size of the disk cache (MB)
	ConfigDir                 string     `json:"config-dir,omitempty"`                   // location of transmission's configuration directory
	DownloadDir               string     `json:"download-dir,omitempty"`                 // default path to download torrents
	DownloadQueueSize         int64      `json:"download-queue-size,omitempty"`          // max number of torrents to download at once (see download-queue-enabled)
	DownloadQueueEnabled      bool       `json:"download-queue-enabled,omitempty"`       // if true, limit how many torrents can be downloaded at once
	DownloadDirFreeSpace      int64      `json:"download-dir-free-space,omitempty"`      // ---- not documented ----
	DhtEnabled                bool       `json:"dht-enabled,omitempty"`                  // true means allow dht in public torrents
	Encryption                Encryption `json:"encryption,omitempty"`                   // "required", "preferred", "tolerated"
	IdleSeedingLimit          int64      `json:"idle-seeding-limit,omitempty"`           // torrents we're seeding will be stopped if they're idle for this long
	IdleSeedingLimitEnabled   bool       `json:"idle-seeding-limit-enabled,omitempty"`   // true if the seeding inactivity limit is honored by default
	IncompleteDir             string     `json:"incomplete-dir,omitempty"`               // path for incomplete torrents, when enabled
	IncompleteDirEnabled      bool       `json:"incomplete-dir-enabled,omitempty"`       // true means keep torrents in incomplete-dir until done
	LpdEnabled                bool       `json:"lpd-enabled,omitempty"`                  // true means allow Local Peer Discovery in public torrents
	PeerLimitGlobal           int64      `json:"peer-limit-global,omitempty"`            // maximum global number of peers
	PeerLimitPerTorrent       int64      `json:"peer-limit-per-torrent,omitempty"`       // maximum global number of peers
	PexEnabled                bool       `json:"pex-enabled,omitempty"`                  // true means allow pex in public torrents
	PeerPort                  int64      `json:"peer-port,omitempty"`                    // port number
	PeerPortRandomOnStart     bool       `json:"peer-port-random-on-start,omitempty"`    // true means pick a random peer port on launch
	PortForwardingEnabled     bool       `json:"port-forwarding-enabled,omitempty"`      // true means enabled
	QueueStalledEnabled       bool       `json:"queue-stalled-enabled,omitempty"`        // whether or not to consider idle torrents as stalled
	QueueStalledMinutes       int64      `json:"queue-stalled-minutes,omitempty"`        // torrents that are idle for N minutes aren't counted toward seed-queue-size or download-queue-size
	RenamePartialFiles        bool       `json:"rename-partial-files,omitempty"`         // true means append ".part" to incomplete files
	RpcVersion                int64      `json:"rpc-version,omitempty"`                  // the current RPC API version
	RpcVersionMinimum         int64      `json:"rpc-version-minimum,omitempty"`          // the minimum RPC API version supported
	ScriptTorrentDoneFilename string     `json:"script-torrent-done-filename,omitempty"` // filename of the script to run
	ScriptTorrentDoneEnabled  bool       `json:"script-torrent-done-enabled,omitempty"`  // whether or not to call the "done" script
	SeedRatioLimit            float64    `json:"seedRatioLimit,omitempty"`               // the default seed ratio for torrents to use
	SeedRatioLimited          bool       `json:"seedRatioLimited,omitempty"`             // true if seedRatioLimit is honored by default
	SeedQueueSize             int64      `json:"seed-queue-size,omitempty"`              // max number of torrents to uploaded at once (see seed-queue-enabled)
	SeedQueueEnabled          bool       `json:"seed-queue-enabled,omitempty"`           // if true, limit how many torrents can be uploaded at once
	SpeedLimitDown            int64      `json:"speed-limit-down,omitempty"`             // max global download speed (KBps)
	SpeedLimitDownEnabled     bool       `json:"speed-limit-down-enabled,omitempty"`     // true means enabled
	SpeedLimitUp              int64      `json:"speed-limit-up,omitempty"`               // max global upload speed (KBps)
	SpeedLimitUpEnabled       bool       `json:"speed-limit-up-enabled,omitempty"`       // true means enabled
	StartAddedTorrents        bool       `json:"start-added-torrents,omitempty"`         // true means added torrents will be started right away
	TrashOriginalTorrentFiles bool       `json:"trash-original-torrent-files,omitempty"` // true means the .torrent file of added torrents will be deleted
	Units                     Units      `json:"units,omitempty"`                        // see units below
	UtpEnabled                bool       `json:"utp-enabled,omitempty"`                  // true means allow utp
	Version                   string     `json:"version,omitempty"`                      // long version string "$version ($revision)"
}

// Units are session units.
type Units struct {
	SpeedUnits  []string `json:"speed-units,omitempty"`  // 4 strings: KB/s, MB/s, GB/s, TB/s
	SpeedBytes  int64    `json:"speed-bytes,omitempty"`  // number of bytes in a KB (1000 for kB; 1024 for KiB)
	SizeUnits   []string `json:"size-units,omitempty"`   // 4 strings: KB/s, MB/s, GB/s, TB/s
	SizeBytes   int64    `json:"size-bytes,omitempty"`   // number of bytes in a KB (1000 for kB; 1024 for KiB)
	MemoryUnits []string `json:"memory-units,omitempty"` // 4 strings: KB/s, MB/s, GB/s, TB/s
	MemoryBytes int64    `json:"memory-bytes,omitempty"` // number of bytes in a KB (1000 for kB; 1024 for KiB)
}

// SessionSetRequest is the session set request.
type SessionSetRequest Session

// SessionSet creates a session set request.
func SessionSet() *SessionSetRequest {
	return &SessionSetRequest{
		changed: make(map[string]bool),
	}
}

// Do executes a request for the session values against the provided context and client.
func (req *SessionSetRequest) Do(ctx context.Context, cl *Client) error {
	params := make(map[string]interface{})

	if req.changed["AltSpeedDown"] {
		params["alt-speed-down"] = req.AltSpeedDown
	}
	if req.changed["AltSpeedEnabled"] {
		params["alt-speed-enabled"] = req.AltSpeedEnabled
	}
	if req.changed["AltSpeedTimeBegin"] {
		params["alt-speed-time-begin"] = req.AltSpeedTimeBegin
	}
	if req.changed["AltSpeedTimeEnabled"] {
		params["alt-speed-time-enabled"] = req.AltSpeedTimeEnabled
	}
	if req.changed["AltSpeedTimeEnd"] {
		params["alt-speed-time-end"] = req.AltSpeedTimeEnd
	}
	if req.changed["AltSpeedTimeDay"] {
		params["alt-speed-time-day"] = req.AltSpeedTimeDay
	}
	if req.changed["AltSpeedUp"] {
		params["alt-speed-up"] = req.AltSpeedUp
	}
	if req.changed["BlocklistURL"] {
		params["blocklist-url"] = req.BlocklistURL
	}
	if req.changed["BlocklistEnabled"] {
		params["blocklist-enabled"] = req.BlocklistEnabled
	}
	if req.changed["BlocklistSize"] {
		params["blocklist-size"] = req.BlocklistSize
	}
	if req.changed["CacheSizeMb"] {
		params["cache-size-mb"] = req.CacheSizeMb
	}
	if req.changed["ConfigDir"] {
		params["config-dir"] = req.ConfigDir
	}
	if req.changed["DownloadDir"] {
		params["download-dir"] = req.DownloadDir
	}
	if req.changed["DownloadQueueSize"] {
		params["download-queue-size"] = req.DownloadQueueSize
	}
	if req.changed["DownloadQueueEnabled"] {
		params["download-queue-enabled"] = req.DownloadQueueEnabled
	}
	if req.changed["DownloadDirFreeSpace"] {
		params["download-dir-free-space"] = req.DownloadDirFreeSpace
	}
	if req.changed["DhtEnabled"] {
		params["dht-enabled"] = req.DhtEnabled
	}
	if req.changed["Encryption"] {
		params["encryption"] = req.Encryption
	}
	if req.changed["IdleSeedingLimit"] {
		params["idle-seeding-limit"] = req.IdleSeedingLimit
	}
	if req.changed["IdleSeedingLimitEnabled"] {
		params["idle-seeding-limit-enabled"] = req.IdleSeedingLimitEnabled
	}
	if req.changed["IncompleteDir"] {
		params["incomplete-dir"] = req.IncompleteDir
	}
	if req.changed["IncompleteDirEnabled"] {
		params["incomplete-dir-enabled"] = req.IncompleteDirEnabled
	}
	if req.changed["LpdEnabled"] {
		params["lpd-enabled"] = req.LpdEnabled
	}
	if req.changed["PeerLimitGlobal"] {
		params["peer-limit-global"] = req.PeerLimitGlobal
	}
	if req.changed["PeerLimitPerTorrent"] {
		params["peer-limit-per-torrent"] = req.PeerLimitPerTorrent
	}
	if req.changed["PexEnabled"] {
		params["pex-enabled"] = req.PexEnabled
	}
	if req.changed["PeerPort"] {
		params["peer-port"] = req.PeerPort
	}
	if req.changed["PeerPortRandomOnStart"] {
		params["peer-port-random-on-start"] = req.PeerPortRandomOnStart
	}
	if req.changed["PortForwardingEnabled"] {
		params["port-forwarding-enabled"] = req.PortForwardingEnabled
	}
	if req.changed["QueueStalledEnabled"] {
		params["queue-stalled-enabled"] = req.QueueStalledEnabled
	}
	if req.changed["QueueStalledMinutes"] {
		params["queue-stalled-minutes"] = req.QueueStalledMinutes
	}
	if req.changed["RenamePartialFiles"] {
		params["rename-partial-files"] = req.RenamePartialFiles
	}
	if req.changed["RpcVersion"] {
		params["rpc-version"] = req.RpcVersion
	}
	if req.changed["RpcVersionMinimum"] {
		params["rpc-version-minimum"] = req.RpcVersionMinimum
	}
	if req.changed["ScriptTorrentDoneFilename"] {
		params["script-torrent-done-filename"] = req.ScriptTorrentDoneFilename
	}
	if req.changed["ScriptTorrentDoneEnabled"] {
		params["script-torrent-done-enabled"] = req.ScriptTorrentDoneEnabled
	}
	if req.changed["SeedRatioLimit"] {
		params["seedRatioLimit"] = req.SeedRatioLimit
	}
	if req.changed["SeedRatioLimited"] {
		params["seedRatioLimited"] = req.SeedRatioLimited
	}
	if req.changed["SeedQueueSize"] {
		params["seed-queue-size"] = req.SeedQueueSize
	}
	if req.changed["SeedQueueEnabled"] {
		params["seed-queue-enabled"] = req.SeedQueueEnabled
	}
	if req.changed["SpeedLimitDown"] {
		params["speed-limit-down"] = req.SpeedLimitDown
	}
	if req.changed["SpeedLimitDownEnabled"] {
		params["speed-limit-down-enabled"] = req.SpeedLimitDownEnabled
	}
	if req.changed["SpeedLimitUp"] {
		params["speed-limit-up"] = req.SpeedLimitUp
	}
	if req.changed["SpeedLimitUpEnabled"] {
		params["speed-limit-up-enabled"] = req.SpeedLimitUpEnabled
	}
	if req.changed["StartAddedTorrents"] {
		params["start-added-torrents"] = req.StartAddedTorrents
	}
	if req.changed["TrashOriginalTorrentFiles"] {
		params["trash-original-torrent-files"] = req.TrashOriginalTorrentFiles
	}
	if req.changed["Units"] {
		params["units"] = req.Units
	}
	if req.changed["UtpEnabled"] {
		params["utp-enabled"] = req.UtpEnabled
	}
	if req.changed["Version"] {
		params["version"] = req.Version
	}

	return cl.Do(ctx, "session-set", params, nil)
}

// WithChanged marks the fields that were changed.
func (req SessionSetRequest) WithChanged(fields ...string) *SessionSetRequest {
	for _, field := range fields {
		req.changed[field] = true
	}
	return &req
}

// WithAltSpeedDown sets max global download speed (KBps).
func (req SessionSetRequest) WithAltSpeedDown(altSpeedDown int64) *SessionSetRequest {
	req.AltSpeedDown = altSpeedDown
	return req.WithChanged("AltSpeedDown")
}

// WithAltSpeedEnabled sets true means use the alt speeds.
func (req SessionSetRequest) WithAltSpeedEnabled(altSpeedEnabled bool) *SessionSetRequest {
	req.AltSpeedEnabled = altSpeedEnabled
	return req.WithChanged("AltSpeedEnabled")
}

// WithAltSpeedTimeBegin sets when to turn on alt speeds (units: minutes after midnight).
func (req SessionSetRequest) WithAltSpeedTimeBegin(altSpeedTimeBegin int64) *SessionSetRequest {
	req.AltSpeedTimeBegin = altSpeedTimeBegin
	return req.WithChanged("AltSpeedTimeBegin")
}

// WithAltSpeedTimeEnabled sets true means the scheduled on/off times are used.
func (req SessionSetRequest) WithAltSpeedTimeEnabled(altSpeedTimeEnabled bool) *SessionSetRequest {
	req.AltSpeedTimeEnabled = altSpeedTimeEnabled
	return req.WithChanged("AltSpeedTimeEnabled")
}

// WithAltSpeedTimeEnd sets when to turn off alt speeds (units: same).
func (req SessionSetRequest) WithAltSpeedTimeEnd(altSpeedTimeEnd int64) *SessionSetRequest {
	req.AltSpeedTimeEnd = altSpeedTimeEnd
	return req.WithChanged("AltSpeedTimeEnd")
}

// WithAltSpeedTimeDay sets what day(s) to turn on alt speeds (look at tr_sched_day).
func (req SessionSetRequest) WithAltSpeedTimeDay(altSpeedTimeDay int64) *SessionSetRequest {
	req.AltSpeedTimeDay = altSpeedTimeDay
	return req.WithChanged("AltSpeedTimeDay")
}

// WithAltSpeedUp sets max global upload speed (KBps).
func (req SessionSetRequest) WithAltSpeedUp(altSpeedUp int64) *SessionSetRequest {
	req.AltSpeedUp = altSpeedUp
	return req.WithChanged("AltSpeedUp")
}

// WithBlocklistURL sets location of the blocklist to use for "blocklist-update".
func (req SessionSetRequest) WithBlocklistURL(blocklistURL string) *SessionSetRequest {
	req.BlocklistURL = blocklistURL
	return req.WithChanged("BlocklistURL")
}

// WithBlocklistEnabled sets true means enabled.
func (req SessionSetRequest) WithBlocklistEnabled(blocklistEnabled bool) *SessionSetRequest {
	req.BlocklistEnabled = blocklistEnabled
	return req.WithChanged("BlocklistEnabled")
}

// WithCacheSizeMb sets maximum size of the disk cache (MB).
func (req SessionSetRequest) WithCacheSizeMb(cacheSizeMb int64) *SessionSetRequest {
	req.CacheSizeMb = cacheSizeMb
	return req.WithChanged("CacheSizeMb")
}

// WithDownloadDir sets default path to download torrents.
func (req SessionSetRequest) WithDownloadDir(downloadDir string) *SessionSetRequest {
	req.DownloadDir = downloadDir
	return req.WithChanged("DownloadDir")
}

// WithDownloadQueueSize sets max number of torrents to download at once (see download-queue-enabled).
func (req SessionSetRequest) WithDownloadQueueSize(downloadQueueSize int64) *SessionSetRequest {
	req.DownloadQueueSize = downloadQueueSize
	return req.WithChanged("DownloadQueueSize")
}

// WithDownloadQueueEnabled sets if true, limit how many torrents can be downloaded at once.
func (req SessionSetRequest) WithDownloadQueueEnabled(downloadQueueEnabled bool) *SessionSetRequest {
	req.DownloadQueueEnabled = downloadQueueEnabled
	return req.WithChanged("DownloadQueueEnabled")
}

// WithDhtEnabled sets true means allow dht in public torrents.
func (req SessionSetRequest) WithDhtEnabled(dhtEnabled bool) *SessionSetRequest {
	req.DhtEnabled = dhtEnabled
	return req.WithChanged("DhtEnabled")
}

// WithEncryption sets "required", "preferred", "tolerated".
func (req SessionSetRequest) WithEncryption(encryption Encryption) *SessionSetRequest {
	req.Encryption = encryption
	return req.WithChanged("Encryption")
}

// WithIdleSeedingLimit sets torrents we're seeding will be stopped if they're idle for this long.
func (req SessionSetRequest) WithIdleSeedingLimit(idleSeedingLimit int64) *SessionSetRequest {
	req.IdleSeedingLimit = idleSeedingLimit
	return req.WithChanged("IdleSeedingLimit")
}

// WithIdleSeedingLimitEnabled sets true if the seeding inactivity limit is honored by default.
func (req SessionSetRequest) WithIdleSeedingLimitEnabled(idleSeedingLimitEnabled bool) *SessionSetRequest {
	req.IdleSeedingLimitEnabled = idleSeedingLimitEnabled
	return req.WithChanged("IdleSeedingLimitEnabled")
}

// WithIncompleteDir sets path for incomplete torrents, when enabled.
func (req SessionSetRequest) WithIncompleteDir(incompleteDir string) *SessionSetRequest {
	req.IncompleteDir = incompleteDir
	return req.WithChanged("IncompleteDir")
}

// WithIncompleteDirEnabled sets true means keep torrents in incomplete-dir until done.
func (req SessionSetRequest) WithIncompleteDirEnabled(incompleteDirEnabled bool) *SessionSetRequest {
	req.IncompleteDirEnabled = incompleteDirEnabled
	return req.WithChanged("IncompleteDirEnabled")
}

// WithLpdEnabled sets true means allow Local Peer Discovery in public torrents.
func (req SessionSetRequest) WithLpdEnabled(lpdEnabled bool) *SessionSetRequest {
	req.LpdEnabled = lpdEnabled
	return req.WithChanged("LpdEnabled")
}

// WithPeerLimitGlobal sets maximum global number of peers.
func (req SessionSetRequest) WithPeerLimitGlobal(peerLimitGlobal int64) *SessionSetRequest {
	req.PeerLimitGlobal = peerLimitGlobal
	return req.WithChanged("PeerLimitGlobal")
}

// WithPeerLimitPerTorrent sets maximum global number of peers.
func (req SessionSetRequest) WithPeerLimitPerTorrent(peerLimitPerTorrent int64) *SessionSetRequest {
	req.PeerLimitPerTorrent = peerLimitPerTorrent
	return req.WithChanged("PeerLimitPerTorrent")
}

// WithPexEnabled sets true means allow pex in public torrents.
func (req SessionSetRequest) WithPexEnabled(pexEnabled bool) *SessionSetRequest {
	req.PexEnabled = pexEnabled
	return req.WithChanged("PexEnabled")
}

// WithPeerPort sets port number.
func (req SessionSetRequest) WithPeerPort(peerPort int64) *SessionSetRequest {
	req.PeerPort = peerPort
	return req.WithChanged("PeerPort")
}

// WithPeerPortRandomOnStart sets true means pick a random peer port on launch.
func (req SessionSetRequest) WithPeerPortRandomOnStart(peerPortRandomOnStart bool) *SessionSetRequest {
	req.PeerPortRandomOnStart = peerPortRandomOnStart
	return req.WithChanged("PeerPortRandomOnStart")
}

// WithPortForwardingEnabled sets true means enabled.
func (req SessionSetRequest) WithPortForwardingEnabled(portForwardingEnabled bool) *SessionSetRequest {
	req.PortForwardingEnabled = portForwardingEnabled
	return req.WithChanged("PortForwardingEnabled")
}

// WithQueueStalledEnabled sets whether or not to consider idle torrents as stalled.
func (req SessionSetRequest) WithQueueStalledEnabled(queueStalledEnabled bool) *SessionSetRequest {
	req.QueueStalledEnabled = queueStalledEnabled
	return req.WithChanged("QueueStalledEnabled")
}

// WithQueueStalledMinutes sets torrents that are idle for N minutes aren't counted toward seed-queue-size or download-queue-size.
func (req SessionSetRequest) WithQueueStalledMinutes(queueStalledMinutes int64) *SessionSetRequest {
	req.QueueStalledMinutes = queueStalledMinutes
	return req.WithChanged("QueueStalledMinutes")
}

// WithRenamePartialFiles sets true means append ".part" to incomplete files.
func (req SessionSetRequest) WithRenamePartialFiles(renamePartialFiles bool) *SessionSetRequest {
	req.RenamePartialFiles = renamePartialFiles
	return req.WithChanged("RenamePartialFiles")
}

// WithScriptTorrentDoneFilename sets filename of the script to run.
func (req SessionSetRequest) WithScriptTorrentDoneFilename(scriptTorrentDoneFilename string) *SessionSetRequest {
	req.ScriptTorrentDoneFilename = scriptTorrentDoneFilename
	return req.WithChanged("ScriptTorrentDoneFilename")
}

// WithScriptTorrentDoneEnabled sets whether or not to call the "done" script.
func (req SessionSetRequest) WithScriptTorrentDoneEnabled(scriptTorrentDoneEnabled bool) *SessionSetRequest {
	req.ScriptTorrentDoneEnabled = scriptTorrentDoneEnabled
	return req.WithChanged("ScriptTorrentDoneEnabled")
}

// WithSeedRatioLimit sets the default seed ratio for torrents to use.
func (req SessionSetRequest) WithSeedRatioLimit(seedRatioLimit float64) *SessionSetRequest {
	req.SeedRatioLimit = seedRatioLimit
	return req.WithChanged("SeedRatioLimit")
}

// WithSeedRatioLimited sets true if seedRatioLimit is honored by default.
func (req SessionSetRequest) WithSeedRatioLimited(seedRatioLimited bool) *SessionSetRequest {
	req.SeedRatioLimited = seedRatioLimited
	return req.WithChanged("SeedRatioLimited")
}

// WithSeedQueueSize sets max number of torrents to uploaded at once (see seed-queue-enabled).
func (req SessionSetRequest) WithSeedQueueSize(seedQueueSize int64) *SessionSetRequest {
	req.SeedQueueSize = seedQueueSize
	return req.WithChanged("SeedQueueSize")
}

// WithSeedQueueEnabled sets if true, limit how many torrents can be uploaded at once.
func (req SessionSetRequest) WithSeedQueueEnabled(seedQueueEnabled bool) *SessionSetRequest {
	req.SeedQueueEnabled = seedQueueEnabled
	return req.WithChanged("SeedQueueEnabled")
}

// WithSpeedLimitDown sets max global download speed (KBps).
func (req SessionSetRequest) WithSpeedLimitDown(speedLimitDown int64) *SessionSetRequest {
	req.SpeedLimitDown = speedLimitDown
	return req.WithChanged("SpeedLimitDown")
}

// WithSpeedLimitDownEnabled sets true means enabled.
func (req SessionSetRequest) WithSpeedLimitDownEnabled(speedLimitDownEnabled bool) *SessionSetRequest {
	req.SpeedLimitDownEnabled = speedLimitDownEnabled
	return req.WithChanged("SpeedLimitDownEnabled")
}

// WithSpeedLimitUp sets max global upload speed (KBps).
func (req SessionSetRequest) WithSpeedLimitUp(speedLimitUp int64) *SessionSetRequest {
	req.SpeedLimitUp = speedLimitUp
	return req.WithChanged("SpeedLimitUp")
}

// WithSpeedLimitUpEnabled sets true means enabled.
func (req SessionSetRequest) WithSpeedLimitUpEnabled(speedLimitUpEnabled bool) *SessionSetRequest {
	req.SpeedLimitUpEnabled = speedLimitUpEnabled
	return req.WithChanged("SpeedLimitUpEnabled")
}

// WithStartAddedTorrents sets true means added torrents will be started right away.
func (req SessionSetRequest) WithStartAddedTorrents(startAddedTorrents bool) *SessionSetRequest {
	req.StartAddedTorrents = startAddedTorrents
	return req.WithChanged("StartAddedTorrents")
}

// WithTrashOriginalTorrentFiles sets true means the .torrent file of added torrents will be deleted.
func (req SessionSetRequest) WithTrashOriginalTorrentFiles(trashOriginalTorrentFiles bool) *SessionSetRequest {
	req.TrashOriginalTorrentFiles = trashOriginalTorrentFiles
	return req.WithChanged("TrashOriginalTorrentFiles")
}

// WithUnits sets see units below.
func (req SessionSetRequest) WithUnits(units Units) *SessionSetRequest {
	req.Units = units
	return req.WithChanged("Units")
}

// WithUtpEnabled sets true means allow utp.
func (req SessionSetRequest) WithUtpEnabled(utpEnabled bool) *SessionSetRequest {
	req.UtpEnabled = utpEnabled
	return req.WithChanged("UtpEnabled")
}

// SessionGetRequest is the session get request.
type SessionGetRequest struct{}

// SessionGet creates a session get request.
func SessionGet() *SessionGetRequest {
	return &SessionGetRequest{}
}

// Do executes the session get request against the provided context and client.
func (req *SessionGetRequest) Do(ctx context.Context, cl *Client) (*SessionGetResponse, error) {
	res := new(SessionGetResponse)
	if err := cl.Do(ctx, "session-get", req, res); err != nil {
		return nil, err
	}
	return res, nil
}

// SessionGetResponse is the session get response.
type SessionGetResponse = Session

// SessionStatsRequest is the session stats request.
type SessionStatsRequest struct{}

// SessionStats creates a session stats request.
func SessionStats() *SessionStatsRequest {
	return &SessionStatsRequest{}
}

// Do executes the session stats request against the provided context and
// client.
func (req *SessionStatsRequest) Do(ctx context.Context, cl *Client) (*SessionStatsResponse, error) {
	res := new(SessionStatsResponse)
	if err := cl.Do(ctx, "session-stats", req, res); err != nil {
		return nil, err
	}
	return res, nil
}

// SessionStatsResponse is the session stats response.
type SessionStatsResponse struct {
	ActiveTorrentCount int64 `json:"activeTorrentCount,omitempty"`
	DownloadSpeed      int64 `json:"downloadSpeed,omitempty"`
	PausedTorrentCount int64 `json:"pausedTorrentCount,omitempty"`
	TorrentCount       int64 `json:"torrentCount,omitempty"`
	UploadSpeed        int64 `json:"uploadSpeed,omitempty"`
	CumulativeStats    struct {
		UploadedBytes   int64 `json:"uploadedBytes,omitempty"`   // tr_session_stats
		DownloadedBytes int64 `json:"downloadedBytes,omitempty"` // tr_session_stats
		FilesAdded      int64 `json:"filesAdded,omitempty"`      // tr_session_stats
		SessionCount    int64 `json:"sessionCount,omitempty"`    // tr_session_stats
		SecondsActive   int64 `json:"secondsActive,omitempty"`   // tr_session_stats
	} `json:"cumulative-stats,omitempty"`
	CurrentStats struct {
		UploadedBytes   int64 `json:"uploadedBytes,omitempty"`   // tr_session_stats
		DownloadedBytes int64 `json:"downloadedBytes,omitempty"` // tr_session_stats
		FilesAdded      int64 `json:"filesAdded,omitempty"`      // tr_session_stats
		SessionCount    int64 `json:"sessionCount,omitempty"`    // tr_session_stats
		SecondsActive   int64 `json:"secondsActive,omitempty"`   // tr_session_stats
	} `json:"current-stats,omitempty"`
}

// BlocklistUpdateRequest is the blocklist update request.
type BlocklistUpdateRequest struct{}

// BlocklistUpdate creates a blocklist update request.
func BlocklistUpdate() *BlocklistUpdateRequest {
	return &BlocklistUpdateRequest{}
}

// Do executes the blocklist update request against the provided context and
// client.
func (req *BlocklistUpdateRequest) Do(ctx context.Context, cl *Client) (int64, error) {
	var res struct {
		BlocklistSize int64 `json:"blocklist-size,omitempty"`
	}
	if err := cl.Do(ctx, "blocklist-update", req, &res); err != nil {
		return 0, err
	}
	return res.BlocklistSize, nil
}

// PortTestRequest is the port test request.
type PortTestRequest struct{}

// PortTest creates a port test request.
func PortTest() *PortTestRequest {
	return &PortTestRequest{}
}

// Do executes the port test request against the provided context and
// client.
func (req *PortTestRequest) Do(ctx context.Context, cl *Client) (bool, error) {
	var res struct {
		PortIsOpen bool `json:"port-is-open,omitempty"`
	}
	if err := cl.Do(ctx, "port-test", req, &res); err != nil {
		return false, err
	}
	return res.PortIsOpen, nil
}

// SessionCloseRequest is the session close request.
type SessionCloseRequest struct{}

// SessionClose creates a session close request.
func SessionClose() *SessionCloseRequest {
	return &SessionCloseRequest{}
}

// SessionShutdown creates a session close request.
//
// Alias for SessionClose.
func SessionShutdown() *SessionCloseRequest {
	return SessionClose()
}

// Do executes the session close request against the provided context and
// client.
func (req *SessionCloseRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "session-close", req, nil)
}

// QueueMoveTopRequest is a queue move top request.
type QueueMoveTopRequest = Request

// QueueMoveTop creates a queue move top request for the specified ids.
func QueueMoveTop(ids ...interface{}) *QueueMoveTopRequest {
	return NewRequest("queue-move-top", ids...)
}

// QueueMoveUpRequest is a queue move up request.
type QueueMoveUpRequest = Request

// QueueMoveUp creates a queue move up request for the specified ids.
func QueueMoveUp(ids ...interface{}) *QueueMoveUpRequest {
	return NewRequest("queue-move-up", ids...)
}

// QueueMoveDownRequest is a queue move down request.
type QueueMoveDownRequest = Request

// QueueMoveDown creates a queue move down request for the specified ids.
func QueueMoveDown(ids ...interface{}) *QueueMoveDownRequest {
	return NewRequest("queue-move-down", ids...)
}

// QueueMoveBottomRequest is a queue move bottom request.
type QueueMoveBottomRequest = Request

// QueueMoveBottom creates a queue move bottom request for the specified ids.
func QueueMoveBottom(ids ...interface{}) *QueueMoveBottomRequest {
	return NewRequest("queue-move-bottom", ids...)
}

// FreeSpaceRequest is the free space request.
type FreeSpaceRequest struct {
	Path string `json:"path,omitempty"`
}

// FreeSpace creates a free space request.
func FreeSpace(path string) *FreeSpaceRequest {
	return &FreeSpaceRequest{
		Path: path,
	}
}

// Do executes the free space request against the provided context and client.
func (req *FreeSpaceRequest) Do(ctx context.Context, cl *Client) (int64, error) {
	var res struct {
		Path      string `json:"path,omitempty"`
		SizeBytes int64  `json:"size-bytes,omitempty"`
	}
	if err := cl.Do(ctx, "free-space", req, &res); err != nil {
		return 0, err
	}
	return res.SizeBytes, nil
}
