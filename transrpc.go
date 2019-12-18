// Package transrpc provides a client for Transmission RPC hosts.
//
// See: https://github.com/transmission/transmission/blob/master/extras/rpc-spec.txt
package transrpc

//go:generate stringer -type Status -trimprefix Status
//go:generate stringer -type Priority -trimprefix Priority
//go:generate stringer -type Mode -trimprefix Mode
//go:generate stringer -type State -trimprefix State

import (
	"context"
	"fmt"
	"strconv"
	"strings"
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

// Status are torrent statuses.
type Status int64

// Statuses.
const (
	StatusStopped Status = iota
	StatusCheckWait
	StatusChecking
	StatusDownloadWait
	StatusDownloading
	StatusSeedWait
	StatusSeeding
)

// UnmarshalJSON satisfies the json.Unmarshaler interface.
func (s *Status) UnmarshalJSON(buf []byte) error {
	if len(buf) == 0 {
		return ErrInvalidStatus
	}
	i, err := strconv.ParseInt(string(buf), 10, 64)
	if err != nil {
		return err
	}
	switch x := Status(i); x {
	case StatusStopped, StatusCheckWait, StatusChecking, StatusDownloadWait, StatusDownloading, StatusSeedWait, StatusSeeding:
		*s = x
		return nil
	}
	return ErrInvalidStatus
}

// Mode are idle/ratio modes.
type Mode int64

// Modes.
const (
	ModeGlobal Mode = iota
	ModeSingle
	ModeUnlimited
)

// UnmarshalJSON satisfies the json.Unmarshaler interface.
func (m *Mode) UnmarshalJSON(buf []byte) error {
	if len(buf) == 0 {
		return ErrInvalidMode
	}
	i, err := strconv.ParseInt(string(buf), 10, 64)
	if err != nil {
		return err
	}
	switch x := Mode(i); x {
	case ModeGlobal, ModeSingle, ModeUnlimited:
		*m = x
		return nil
	}
	return ErrInvalidMode
}

// State are tracker states.
type State int64

// States.
const (
	StateInactive State = iota
	StateWaiting
	StateQueued
	StateActive
)

// UnmarshalJSON satisfies the json.Unmarshaler interface.
func (s *State) UnmarshalJSON(buf []byte) error {
	if len(buf) == 0 {
		return ErrInvalidState
	}
	i, err := strconv.ParseInt(string(buf), 10, 64)
	if err != nil {
		return err
	}
	switch x := State(i); x {
	case StateInactive, StateWaiting, StateQueued, StateActive:
		*s = x
		return nil
	}
	return ErrInvalidState
}

// Time wraps time.Time.
type Time time.Time

// String satisfies the fmt.Stringer interface.
func (t Time) String() string {
	if time.Time(t).IsZero() {
		return "-"
	}
	return time.Time(t).Format("2006-01-02 15:04:05")
}

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

// MarshalJSON satisfies the json.Marshaler interface.
func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(time.Time(t).Unix(), 10)), nil
}

// MarshalYAML satisfies the yaml.Marshaler interface.
func (t Time) MarshalYAML() (interface{}, error) {
	return time.Time(t).Unix(), nil
}

// Duration wraps time.Duration.
type Duration time.Duration

// String satisfies the fmt.Stringer interface.
func (d Duration) String() string {
	switch v := time.Duration(d); v {
	case -1 * time.Second:
		return "Done"
	case -2 * time.Second:
		return "Unknown"
	default:
		return v.String()
	}
}

// UnmarshalJSON satisfies the json.Unmarshaler interface.
func (d *Duration) UnmarshalJSON(buf []byte) error {
	if len(buf) == 0 {
		return ErrInvalidDuration
	}
	i, err := strconv.ParseInt(string(buf), 10, 64)
	if err != nil {
		return err
	}
	if i == 0 {
		return nil
	}
	*d = Duration(i * int64(time.Second))
	return nil
}

// MarshalJSON satisfies the json.Marshaler interface.
func (d Duration) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(int64(time.Duration(d)/time.Second), 10)), nil
}

// MarshalYAML satisfies the yaml.Marshaler interface.
func (d Duration) MarshalYAML() (interface{}, error) {
	return int64(time.Duration(d) / time.Second), nil
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
		*b = Bool(s == "true")
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

// MarshalJSON satisfies the json.Marshaler interface.
func (b Bool) MarshalJSON() ([]byte, error) {
	if bool(b) {
		return []byte("1"), nil
	}
	return []byte("0"), nil
}

// MarshalYAML satisfies the yaml.Marshaler interface.
func (b Bool) MarshalYAML() (interface{}, error) {
	return bool(b), nil
}

// ByteCount wraps a byte count as int64.
type ByteCount int64

// String satisfies the fmt.Stringer interface.
func (bc ByteCount) String() string {
	return bc.Format(true, 2, "")
}

// Format formats the byte count.
func (bc ByteCount) Format(asIEC bool, prec int, suffix string) string {
	c, sizes, end := int64(1000), "kMGTPEZY", "B"
	if asIEC {
		c, end, sizes = 1024, "iB", "KMGTPEZY"
	}

	if int64(bc) < c {
		return fmt.Sprintf("%d B%s", bc, suffix)
	}

	exp, div := 0, c
	for n := int64(bc) / c; n >= c; n /= c {
		div *= c
		exp++
	}
	return fmt.Sprintf("%."+fmt.Sprintf("%d", prec)+"f %c%s%s", float64(bc)/float64(div), sizes[exp], end, suffix)
}

// Percent wraps a float64.
type Percent float64

// String satisfies the fmt.Stringer interface.
func (p Percent) String() string {
	return fmt.Sprintf("%.f%%", float64(p)*100)
}

// Torrent holds information about a torrent.
type Torrent struct {
	ActivityDate      Time      `json:"activityDate,omitempty" yaml:"activityDate,omitempty"`           // tr_stat
	AddedDate         Time      `json:"addedDate,omitempty" yaml:"addedDate,omitempty"`                 // tr_stat
	BandwidthPriority int64     `json:"bandwidthPriority,omitempty" yaml:"bandwidthPriority,omitempty"` // tr_priority_t
	Comment           string    `json:"comment,omitempty" yaml:"comment,omitempty"`                     // tr_info
	CorruptEver       ByteCount `json:"corruptEver,omitempty" yaml:"corruptEver,omitempty"`             // tr_stat
	Creator           string    `json:"creator,omitempty" yaml:"creator,omitempty"`                     // tr_info
	DateCreated       Time      `json:"dateCreated,omitempty" yaml:"dateCreated,omitempty"`             // tr_info
	DesiredAvailable  ByteCount `json:"desiredAvailable,omitempty" yaml:"desiredAvailable,omitempty"`   // tr_stat
	DoneDate          Time      `json:"doneDate,omitempty" yaml:"doneDate,omitempty"`                   // tr_stat
	DownloadDir       string    `json:"downloadDir,omitempty" yaml:"downloadDir,omitempty"`             // tr_torrent
	DownloadedEver    ByteCount `json:"downloadedEver,omitempty" yaml:"downloadedEver,omitempty"`       // tr_stat
	DownloadLimit     ByteCount `json:"downloadLimit,omitempty" yaml:"downloadLimit,omitempty"`         // tr_torrent
	DownloadLimited   bool      `json:"downloadLimited,omitempty" yaml:"downloadLimited,omitempty"`     // tr_torrent
	Error             int64     `json:"error,omitempty" yaml:"error,omitempty"`                         // tr_stat
	ErrorString       string    `json:"errorString,omitempty" yaml:"errorString,omitempty"`             // tr_stat
	Eta               Duration  `json:"eta,omitempty" yaml:"eta,omitempty"`                             // tr_stat
	EtaIdle           Duration  `json:"etaIdle,omitempty" yaml:"etaIdle,omitempty"`                     // tr_stat
	Files             []struct {
		BytesCompleted ByteCount `json:"bytesCompleted,omitempty" yaml:"bytesCompleted,omitempty"` // tr_torrent
		Length         ByteCount `json:"length,omitempty" yaml:"length,omitempty"`                 // tr_info
		Name           string    `json:"name,omitempty" yaml:"name,omitempty"`                     // tr_info
	} `json:"files,omitempty" yaml:"files,omitempty"` // n/a
	FileStats []struct {
		BytesCompleted ByteCount `json:"bytesCompleted,omitempty" yaml:"bytesCompleted,omitempty"` // tr_torrent
		Wanted         bool      `json:"wanted,omitempty" yaml:"wanted,omitempty"`                 // tr_info
		Priority       Priority  `json:"priority,omitempty" yaml:"priority,omitempty"`             // tr_info
	} `json:"fileStats,omitempty" yaml:"fileStats,omitempty"` // n/a
	HashString              string    `json:"hashString,omitempty" yaml:"hashString,omitempty"`                           // tr_info
	HaveUnchecked           ByteCount `json:"haveUnchecked,omitempty" yaml:"haveUnchecked,omitempty"`                     // tr_stat
	HaveValid               ByteCount `json:"haveValid,omitempty" yaml:"haveValid,omitempty"`                             // tr_stat
	HonorsSessionLimits     bool      `json:"honorsSessionLimits,omitempty" yaml:"honorsSessionLimits,omitempty"`         // tr_torrent
	ID                      int64     `json:"id,omitempty" yaml:"id,omitempty"`                                           // tr_torrent
	IsFinished              bool      `json:"isFinished,omitempty" yaml:"isFinished,omitempty"`                           // tr_stat
	IsPrivate               bool      `json:"isPrivate,omitempty" yaml:"isPrivate,omitempty"`                             // tr_torrent
	IsStalled               bool      `json:"isStalled,omitempty" yaml:"isStalled,omitempty"`                             // tr_stat
	Labels                  []string  `json:"labels,omitempty" yaml:"labels,omitempty"`                                   // tr_torrent
	LeftUntilDone           ByteCount `json:"leftUntilDone,omitempty" yaml:"leftUntilDone,omitempty"`                     // tr_stat
	MagnetLink              string    `json:"magnetLink,omitempty" yaml:"magnetLink,omitempty"`                           // n/a
	ManualAnnounceTime      Duration  `json:"manualAnnounceTime,omitempty" yaml:"manualAnnounceTime,omitempty"`           // tr_stat
	MaxConnectedPeers       int64     `json:"maxConnectedPeers,omitempty" yaml:"maxConnectedPeers,omitempty"`             // tr_torrent
	MetadataPercentComplete Percent   `json:"metadataPercentComplete,omitempty" yaml:"metadataPercentComplete,omitempty"` // tr_stat
	Name                    string    `json:"name,omitempty" yaml:"name,omitempty"`                                       // tr_info
	PeerLimit               int64     `json:"peer-limit,omitempty" yaml:"peer-limit,omitempty"`                           // tr_torrent
	Peers                   []struct {
		Address            string    `json:"address,omitempty" yaml:"address,omitempty"`                       // tr_peer_stat
		ClientName         string    `json:"clientName,omitempty" yaml:"clientName,omitempty"`                 // tr_peer_stat
		ClientIsChoked     bool      `json:"clientIsChoked,omitempty" yaml:"clientIsChoked,omitempty"`         // tr_peer_stat
		ClientIsInterested bool      `json:"clientIsInterested,omitempty" yaml:"clientIsInterested,omitempty"` // tr_peer_stat
		FlagStr            string    `json:"flagStr,omitempty" yaml:"flagStr,omitempty"`                       // tr_peer_stat
		IsDownloadingFrom  bool      `json:"isDownloadingFrom,omitempty" yaml:"isDownloadingFrom,omitempty"`   // tr_peer_stat
		IsEncrypted        bool      `json:"isEncrypted,omitempty" yaml:"isEncrypted,omitempty"`               // tr_peer_stat
		IsIncoming         bool      `json:"isIncoming,omitempty" yaml:"isIncoming,omitempty"`                 // tr_peer_stat
		IsUploadingTo      bool      `json:"isUploadingTo,omitempty" yaml:"isUploadingTo,omitempty"`           // tr_peer_stat
		IsUTP              bool      `json:"isUTP,omitempty" yaml:"isUTP,omitempty"`                           // tr_peer_stat
		PeerIsChoked       bool      `json:"peerIsChoked,omitempty" yaml:"peerIsChoked,omitempty"`             // tr_peer_stat
		PeerIsInterested   bool      `json:"peerIsInterested,omitempty" yaml:"peerIsInterested,omitempty"`     // tr_peer_stat
		Port               int64     `json:"port,omitempty" yaml:"port,omitempty"`                             // tr_peer_stat
		Progress           Percent   `json:"progress,omitempty" yaml:"progress,omitempty"`                     // tr_peer_stat
		RateToClient       ByteCount `json:"rateToClient,omitempty" yaml:"rateToClient,omitempty"`             // tr_peer_stat
		RateToPeer         ByteCount `json:"rateToPeer,omitempty" yaml:"rateToPeer,omitempty"`                 // tr_peer_stat
	} `json:"peers,omitempty" yaml:"peers,omitempty"` // n/a
	PeersConnected int64 `json:"peersConnected,omitempty" yaml:"peersConnected,omitempty"` // tr_stat
	PeersFrom      struct {
		FromCache    int64 `json:"fromCache,omitempty" yaml:"fromCache,omitempty"`       // tr_stat
		FromDht      int64 `json:"fromDht,omitempty" yaml:"fromDht,omitempty"`           // tr_stat
		FromIncoming int64 `json:"fromIncoming,omitempty" yaml:"fromIncoming,omitempty"` // tr_stat
		FromLpd      int64 `json:"fromLpd,omitempty" yaml:"fromLpd,omitempty"`           // tr_stat
		FromLtep     int64 `json:"fromLtep,omitempty" yaml:"fromLtep,omitempty"`         // tr_stat
		FromPex      int64 `json:"fromPex,omitempty" yaml:"fromPex,omitempty"`           // tr_stat
		FromTracker  int64 `json:"fromTracker,omitempty" yaml:"fromTracker,omitempty"`   // tr_stat
	} `json:"peersFrom,omitempty" yaml:"peersFrom,omitempty"` // n/a
	PeersGettingFromUs int64      `json:"peersGettingFromUs,omitempty" yaml:"peersGettingFromUs,omitempty"` // tr_stat
	PeersSendingToUs   int64      `json:"peersSendingToUs,omitempty" yaml:"peersSendingToUs,omitempty"`     // tr_stat
	PercentDone        Percent    `json:"percentDone,omitempty" yaml:"percentDone,omitempty"`               // tr_stat
	Pieces             []byte     `json:"pieces,omitempty" yaml:"pieces,omitempty"`                         // tr_torrent
	PieceCount         int64      `json:"pieceCount,omitempty" yaml:"pieceCount,omitempty"`                 // tr_info
	PieceSize          ByteCount  `json:"pieceSize,omitempty" yaml:"pieceSize,omitempty"`                   // tr_info
	Priorities         []Priority `json:"priorities,omitempty" yaml:"priorities,omitempty"`                 // n/a
	QueuePosition      int64      `json:"queuePosition,omitempty" yaml:"queuePosition,omitempty"`           // tr_stat
	RateDownload       ByteCount  `json:"rateDownload,omitempty" yaml:"rateDownload,omitempty"`             // tr_stat
	RateUpload         ByteCount  `json:"rateUpload,omitempty" yaml:"rateUpload,omitempty"`                 // tr_stat
	RecheckProgress    Percent    `json:"recheckProgress,omitempty" yaml:"recheckProgress,omitempty"`       // tr_stat
	SecondsDownloading Duration   `json:"secondsDownloading,omitempty" yaml:"secondsDownloading,omitempty"` // tr_stat
	SecondsSeeding     Duration   `json:"secondsSeeding,omitempty" yaml:"secondsSeeding,omitempty"`         // tr_stat
	SeedIdleLimit      int64      `json:"seedIdleLimit,omitempty" yaml:"seedIdleLimit,omitempty"`           // tr_torrent
	SeedIdleMode       Mode       `json:"seedIdleMode,omitempty" yaml:"seedIdleMode,omitempty"`             // tr_inactvelimit
	SeedRatioLimit     float64    `json:"seedRatioLimit,omitempty" yaml:"seedRatioLimit,omitempty"`         // tr_torrent
	SeedRatioMode      Mode       `json:"seedRatioMode,omitempty" yaml:"seedRatioMode,omitempty"`           // tr_ratiolimit
	SizeWhenDone       ByteCount  `json:"sizeWhenDone,omitempty" yaml:"sizeWhenDone,omitempty"`             // tr_stat
	StartDate          Time       `json:"startDate,omitempty" yaml:"startDate,omitempty"`                   // tr_stat
	Status             Status     `json:"status,omitempty" yaml:"status,omitempty"`                         // tr_stat
	Trackers           []struct {
		Announce string `json:"announce,omitempty" yaml:"announce,omitempty"` // tr_tracker_info
		ID       int64  `json:"id,omitempty" yaml:"id,omitempty"`             // tr_tracker_info
		Scrape   string `json:"scrape,omitempty" yaml:"scrape,omitempty"`     // tr_tracker_info
		Tier     int64  `json:"tier,omitempty" yaml:"tier,omitempty"`         // tr_tracker_info
	} `json:"trackers,omitempty" yaml:"trackers,omitempty"` // n/a
	TrackerStats []struct {
		Announce              string `json:"announce,omitempty" yaml:"announce,omitempty"`                           // tr_tracker_stat
		AnnounceState         State  `json:"announceState,omitempty" yaml:"announceState,omitempty"`                 // tr_tracker_stat
		DownloadCount         int64  `json:"downloadCount,omitempty" yaml:"downloadCount,omitempty"`                 // tr_tracker_stat
		HasAnnounced          bool   `json:"hasAnnounced,omitempty" yaml:"hasAnnounced,omitempty"`                   // tr_tracker_stat
		HasScraped            bool   `json:"hasScraped,omitempty" yaml:"hasScraped,omitempty"`                       // tr_tracker_stat
		Host                  string `json:"host,omitempty" yaml:"host,omitempty"`                                   // tr_tracker_stat
		ID                    int64  `json:"id,omitempty" yaml:"id,omitempty"`                                       // tr_tracker_stat
		IsBackup              bool   `json:"isBackup,omitempty" yaml:"isBackup,omitempty"`                           // tr_tracker_stat
		LastAnnouncePeerCount int64  `json:"lastAnnouncePeerCount,omitempty" yaml:"lastAnnouncePeerCount,omitempty"` // tr_tracker_stat
		LastAnnounceResult    string `json:"lastAnnounceResult,omitempty" yaml:"lastAnnounceResult,omitempty"`       // tr_tracker_stat
		LastAnnounceStartTime Time   `json:"lastAnnounceStartTime,omitempty" yaml:"lastAnnounceStartTime,omitempty"` // tr_tracker_stat
		LastAnnounceSucceeded bool   `json:"lastAnnounceSucceeded,omitempty" yaml:"lastAnnounceSucceeded,omitempty"` // tr_tracker_stat
		LastAnnounceTime      Time   `json:"lastAnnounceTime,omitempty" yaml:"lastAnnounceTime,omitempty"`           // tr_tracker_stat
		LastAnnounceTimedOut  bool   `json:"lastAnnounceTimedOut,omitempty" yaml:"lastAnnounceTimedOut,omitempty"`   // tr_tracker_stat
		LastScrapeResult      string `json:"lastScrapeResult,omitempty" yaml:"lastScrapeResult,omitempty"`           // tr_tracker_stat
		LastScrapeStartTime   Time   `json:"lastScrapeStartTime,omitempty" yaml:"lastScrapeStartTime,omitempty"`     // tr_tracker_stat
		LastScrapeSucceeded   bool   `json:"lastScrapeSucceeded,omitempty" yaml:"lastScrapeSucceeded,omitempty"`     // tr_tracker_stat
		LastScrapeTime        Time   `json:"lastScrapeTime,omitempty" yaml:"lastScrapeTime,omitempty"`               // tr_tracker_stat
		LastScrapeTimedOut    int64  `json:"lastScrapeTimedOut,omitempty" yaml:"lastScrapeTimedOut,omitempty"`       // tr_tracker_stat
		LeecherCount          int64  `json:"leecherCount,omitempty" yaml:"leecherCount,omitempty"`                   // tr_tracker_stat
		NextAnnounceTime      Time   `json:"nextAnnounceTime,omitempty" yaml:"nextAnnounceTime,omitempty"`           // tr_tracker_stat
		NextScrapeTime        Time   `json:"nextScrapeTime,omitempty" yaml:"nextScrapeTime,omitempty"`               // tr_tracker_stat
		Scrape                string `json:"scrape,omitempty" yaml:"scrape,omitempty"`                               // tr_tracker_stat
		ScrapeState           State  `json:"scrapeState,omitempty" yaml:"scrapeState,omitempty"`                     // tr_tracker_stat
		SeederCount           int64  `json:"seederCount,omitempty" yaml:"seederCount,omitempty"`                     // tr_tracker_stat
		Tier                  int64  `json:"tier,omitempty" yaml:"tier,omitempty"`                                   // tr_tracker_stat
	} `json:"trackerStats,omitempty" yaml:"trackerStats,omitempty"` // n/a
	TotalSize           ByteCount `json:"totalSize,omitempty" yaml:"totalSize,omitempty"`                     // tr_info
	TorrentFile         string    `json:"torrentFile,omitempty" yaml:"torrentFile,omitempty"`                 // tr_info
	UploadedEver        ByteCount `json:"uploadedEver,omitempty" yaml:"uploadedEver,omitempty"`               // tr_stat
	UploadLimit         int64     `json:"uploadLimit,omitempty" yaml:"uploadLimit,omitempty"`                 // tr_torrent
	UploadLimited       bool      `json:"uploadLimited,omitempty" yaml:"uploadLimited,omitempty"`             // tr_torrent
	UploadRatio         float64   `json:"uploadRatio,omitempty" yaml:"uploadRatio,omitempty"`                 // tr_stat
	Wanted              []Bool    `json:"wanted,omitempty" yaml:"wanted,omitempty"`                           // n/a
	Webseeds            []string  `json:"webseeds,omitempty" yaml:"webseeds,omitempty"`                       // n/a
	WebseedsSendingToUs int64     `json:"webseedsSendingToUs,omitempty" yaml:"webseedsSendingToUs,omitempty"` // tr_stat
}

// ShortHash returns the short hash of the torrent.
func (t Torrent) ShortHash() string {
	if len(t.HashString) < 7 {
		return ""
	}
	return t.HashString[:7]
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
// Passed IDs can be any of type of int{,8,16,32,64}, [40]byte, []byte, or
// string.
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
	params := map[string]interface{}{}
	if ids != nil {
		params["ids"] = ids
	}
	return cl.Do(ctx, req.method, params, nil)
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
	changed map[string]bool

	BandwidthPriority   int64         `json:"bandwidthPriority,omitempty" yaml:"bandwidthPriority,omitempty"`     // this torrent's bandwidth tr_priority_t
	DownloadLimit       int64         `json:"downloadLimit,omitempty" yaml:"downloadLimit,omitempty"`             // maximum download speed (KBps)
	DownloadLimited     bool          `json:"downloadLimited,omitempty" yaml:"downloadLimited,omitempty"`         // true if downloadLimit is honored
	FilesWanted         []int64       `json:"files-wanted,omitempty" yaml:"files-wanted,omitempty"`               // indices of file(s) to download
	FilesUnwanted       []int64       `json:"files-unwanted,omitempty" yaml:"files-unwanted,omitempty"`           // indices of file(s) to not download
	HonorsSessionLimits bool          `json:"honorsSessionLimits,omitempty" yaml:"honorsSessionLimits,omitempty"` // true if session upload limits are honored
	IDs                 []interface{} `json:"ids,omitempty" yaml:"ids,omitempty"`                                 // torrent list, as described in 3.1
	Labels              []string      `json:"labels,omitempty" yaml:"labels,omitempty"`                           // array of string labels
	Location            string        `json:"location,omitempty" yaml:"location,omitempty"`                       // new location of the torrent's content
	PeerLimit           int64         `json:"peer-limit,omitempty" yaml:"peer-limit,omitempty"`                   // maximum int64 of peers
	PriorityHigh        []int64       `json:"priority-high,omitempty" yaml:"priority-high,omitempty"`             // indices of high-priority file(s)
	PriorityLow         []int64       `json:"priority-low,omitempty" yaml:"priority-low,omitempty"`               // indices of low-priority file(s)
	PriorityNormal      []int64       `json:"priority-normal,omitempty" yaml:"priority-normal,omitempty"`         // indices of normal-priority file(s)
	QueuePosition       int64         `json:"queuePosition,omitempty" yaml:"queuePosition,omitempty"`             // position of this torrent in its queue [0...n)
	SeedIdleLimit       int64         `json:"seedIdleLimit,omitempty" yaml:"seedIdleLimit,omitempty"`             // torrent-level int64 of minutes of seeding inactivity
	SeedIdleMode        Mode          `json:"seedIdleMode,omitempty" yaml:"seedIdleMode,omitempty"`               // which seeding inactivity to use.  See tr_idlelimit
	SeedRatioLimit      float64       `json:"seedRatioLimit,omitempty" yaml:"seedRatioLimit,omitempty"`           // torrent-level seeding ratio
	SeedRatioMode       Mode          `json:"seedRatioMode,omitempty" yaml:"seedRatioMode,omitempty"`             // which ratio to use.  See tr_ratiolimit
	TrackerAdd          []string      `json:"trackerAdd,omitempty" yaml:"trackerAdd,omitempty"`                   // strings of announce URLs to add
	TrackerRemove       []int64       `json:"trackerRemove,omitempty" yaml:"trackerRemove,omitempty"`             // ids of trackers to remove
	TrackerReplace      []interface{} `json:"trackerReplace,omitempty" yaml:"trackerReplace,omitempty"`           // pairs of <trackerId/new announce URLs>
	UploadLimit         int64         `json:"uploadLimit,omitempty" yaml:"uploadLimit,omitempty"`                 // maximum upload speed (KBps)
	UploadLimited       bool          `json:"uploadLimited,omitempty" yaml:"uploadLimited,omitempty"`             // true if uploadLimit is honored
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
	if len(req.changed) == 0 {
		return nil
	}

	ids, err := checkIdentifierList(req.IDs...)
	if err != nil {
		return err
	}

	// build params
	params := map[string]interface{}{
		"ids": ids,
	}
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
func (req TorrentSetRequest) WithSeedIdleMode(seedIdleMode Mode) *TorrentSetRequest {
	req.SeedIdleMode = seedIdleMode
	return req.WithChanged("SeedIdleMode")
}

// WithSeedRatioLimit sets torrent-level seeding ratio.
func (req TorrentSetRequest) WithSeedRatioLimit(seedRatioLimit float64) *TorrentSetRequest {
	req.SeedRatioLimit = seedRatioLimit
	return req.WithChanged("SeedRatioLimit")
}

// WithSeedRatioMode sets which ratio to use.  See tr_ratiolimit.
func (req TorrentSetRequest) WithSeedRatioMode(seedRatioMode Mode) *TorrentSetRequest {
	req.SeedRatioMode = seedRatioMode
	return req.WithChanged("SeedRatioMode")
}

// WithTrackerAdd sets strings of announce URLs to add.
func (req TorrentSetRequest) WithTrackerAdd(trackerAdd ...string) *TorrentSetRequest {
	req.TrackerAdd = trackerAdd
	return req.WithChanged("TrackerAdd")
}

// WithTrackerRemove sets ids of trackers to remove.
func (req TorrentSetRequest) WithTrackerRemove(trackerRemove ...int64) *TorrentSetRequest {
	req.TrackerRemove = trackerRemove
	return req.WithChanged("TrackerRemove")
}

// WithTrackerReplace sets pairs of <trackerId/new announce URLs>.
func (req TorrentSetRequest) WithTrackerReplace(trackerReplace ...interface{}) *TorrentSetRequest {
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
	fields []string
}

// TorrentGet creates a torrent get request for the specified torrent ids.
func TorrentGet(ids ...interface{}) *TorrentGetRequest {
	return &TorrentGetRequest{ids: ids}
}

// WithFields indicates the fields for the host to return.
func (req TorrentGetRequest) WithFields(fields ...string) *TorrentGetRequest {
	req.fields = append(req.fields, fields...)
	return &req
}

// Do executes the torrent get request against the provided context and client.
func (req *TorrentGetRequest) Do(ctx context.Context, cl *Client) (*TorrentGetResponse, error) {
	fields := req.fields
	if fields == nil {
		fields = DefaultTorrentGetFields()
	}
	params := map[string]interface{}{
		"fields": fields,
	}
	ids, err := checkIdentifierList(req.ids...)
	if err != nil {
		return nil, err
	}
	if ids != nil {
		params["ids"] = ids
	}
	res := new(TorrentGetResponse)
	if err := cl.Do(ctx, "torrent-get", params, res); err != nil {
		return nil, err
	}
	return res, nil
}

// DefaultTorrentGetFields returns the list of all torrent field names.
func DefaultTorrentGetFields() []string {
	return []string{
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
	}
}

// TorrentGetResponse is the torrent get response.
type TorrentGetResponse struct {
	Torrents []Torrent     `json:"torrents,omitempty" yaml:"torrents,omitempty"` // contains the key/value pairs matching the request's "fields" argument
	Removed  []interface{} `json:"removed,omitempty" yaml:"removed,omitempty"`   // populated when the requested id was "recently-active"
}

// TorrentAddRequest is the torrent add request.
type TorrentAddRequest struct {
	Cookies           string  `json:"cookies,omitempty" yaml:"cookies,omitempty"`                     // pointer to a string of one or more cookies.
	DownloadDir       string  `json:"download-dir,omitempty" yaml:"download-dir,omitempty"`           // path to download the torrent to
	Filename          string  `json:"filename,omitempty" yaml:"filename,omitempty"`                   // filename or URL of the .torrent file
	Metainfo          []byte  `json:"metainfo,omitempty" yaml:"metainfo,omitempty"`                   // base64-encoded .torrent content
	Paused            bool    `json:"paused,omitempty" yaml:"paused,omitempty"`                       // if true, don't start the torrent
	PeerLimit         int64   `json:"peer-limit,omitempty" yaml:"peer-limit,omitempty"`               // maximum int64 of peers
	BandwidthPriority int64   `json:"bandwidthPriority,omitempty" yaml:"bandwidthPriority,omitempty"` // torrent's bandwidth tr_priority_t
	FilesWanted       []int64 `json:"files-wanted,omitempty" yaml:"files-wanted,omitempty"`           // indices of file(s) to download
	FilesUnwanted     []int64 `json:"files-unwanted,omitempty" yaml:"files-unwanted,omitempty"`       // indices of file(s) to not download
	PriorityHigh      []int64 `json:"priority-high,omitempty" yaml:"priority-high,omitempty"`         // indices of high-priority file(s)
	PriorityLow       []int64 `json:"priority-low,omitempty" yaml:"priority-low,omitempty"`           // indices of low-priority file(s)
	PriorityNormal    []int64 `json:"priority-normal,omitempty" yaml:"priority-normal,omitempty"`     // indices of normal-priority file(s)
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

// WithCookies sets cookies to a string of one or more cookies
func (req TorrentAddRequest) WithCookies(cookies string) *TorrentAddRequest {
	req.Cookies = cookies
	return &req
}

// WithCookiesList sets cookies to the set of key, value strings.
func (req TorrentAddRequest) WithCookiesList(cookies ...string) *TorrentAddRequest {
	if len(cookies)%2 != 0 {
		panic("cookies length must be even")
	}
	s := make([]string, len(cookies)/2)
	for i := 0; i < len(cookies); i += 2 {
		s = append(s, fmt.Sprintf("%s=%s", cookies[i], cookies[i+1]))
	}
	req.Cookies = strings.Join(s, "; ")
	return &req
}

// WithCookiesMap sets cookies from a map.
func (req TorrentAddRequest) WithCookiesMap(cookies map[string]string) *TorrentAddRequest {
	s := make([]string, len(cookies))
	for k, v := range cookies {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}
	req.Cookies = strings.Join(s, "; ")
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
	TorrentAdded     *Torrent `json:"torrent-added" yaml:"torrent-added"`
	TorrentDuplicate *Torrent `json:"torrent-duplicate" yaml:"torrent-duplicate"`
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
func TorrentSetLocation(location string, move bool, ids ...interface{}) *TorrentSetLocationRequest {
	return &TorrentSetLocationRequest{
		ids:      ids,
		location: location,
		move:     move,
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
	changed map[string]bool

	AltSpeedDown              int64      `json:"alt-speed-down,omitempty" yaml:"alt-speed-down,omitempty"`                             // max global download speed (KBps)
	AltSpeedEnabled           bool       `json:"alt-speed-enabled,omitempty" yaml:"alt-speed-enabled,omitempty"`                       // true means use the alt speeds
	AltSpeedTimeBegin         int64      `json:"alt-speed-time-begin,omitempty" yaml:"alt-speed-time-begin,omitempty"`                 // when to turn on alt speeds (units: minutes after midnight)
	AltSpeedTimeEnabled       bool       `json:"alt-speed-time-enabled,omitempty" yaml:"alt-speed-time-enabled,omitempty"`             // true means the scheduled on/off times are used
	AltSpeedTimeEnd           int64      `json:"alt-speed-time-end,omitempty" yaml:"alt-speed-time-end,omitempty"`                     // when to turn off alt speeds (units: same)
	AltSpeedTimeDay           int64      `json:"alt-speed-time-day,omitempty" yaml:"alt-speed-time-day,omitempty"`                     // what day(s) to turn on alt speeds (look at tr_sched_day)
	AltSpeedUp                int64      `json:"alt-speed-up,omitempty" yaml:"alt-speed-up,omitempty"`                                 // max global upload speed (KBps)
	BlocklistURL              string     `json:"blocklist-url,omitempty" yaml:"blocklist-url,omitempty"`                               // location of the blocklist to use for "blocklist-update"
	BlocklistEnabled          bool       `json:"blocklist-enabled,omitempty" yaml:"blocklist-enabled,omitempty"`                       // true means enabled
	BlocklistSize             int64      `json:"blocklist-size,omitempty" yaml:"blocklist-size,omitempty"`                             // number of rules in the blocklist
	CacheSizeMb               int64      `json:"cache-size-mb,omitempty" yaml:"cache-size-mb,omitempty"`                               // maximum size of the disk cache (MB)
	ConfigDir                 string     `json:"config-dir,omitempty" yaml:"config-dir,omitempty"`                                     // location of transmission's configuration directory
	DownloadDir               string     `json:"download-dir,omitempty" yaml:"download-dir,omitempty"`                                 // default path to download torrents
	DownloadQueueSize         int64      `json:"download-queue-size,omitempty" yaml:"download-queue-size,omitempty"`                   // max number of torrents to download at once (see download-queue-enabled)
	DownloadQueueEnabled      bool       `json:"download-queue-enabled,omitempty" yaml:"download-queue-enabled,omitempty"`             // if true, limit how many torrents can be downloaded at once
	DownloadDirFreeSpace      int64      `json:"download-dir-free-space,omitempty" yaml:"download-dir-free-space,omitempty"`           // ---- not documented ----
	DhtEnabled                bool       `json:"dht-enabled,omitempty" yaml:"dht-enabled,omitempty"`                                   // true means allow dht in public torrents
	Encryption                Encryption `json:"encryption,omitempty" yaml:"encryption,omitempty"`                                     // "required", "preferred", "tolerated"
	IdleSeedingLimit          int64      `json:"idle-seeding-limit,omitempty" yaml:"idle-seeding-limit,omitempty"`                     // torrents we're seeding will be stopped if they're idle for this long
	IdleSeedingLimitEnabled   bool       `json:"idle-seeding-limit-enabled,omitempty" yaml:"idle-seeding-limit-enabled,omitempty"`     // true if the seeding inactivity limit is honored by default
	IncompleteDir             string     `json:"incomplete-dir,omitempty" yaml:"incomplete-dir,omitempty"`                             // path for incomplete torrents, when enabled
	IncompleteDirEnabled      bool       `json:"incomplete-dir-enabled,omitempty" yaml:"incomplete-dir-enabled,omitempty"`             // true means keep torrents in incomplete-dir until done
	LpdEnabled                bool       `json:"lpd-enabled,omitempty" yaml:"lpd-enabled,omitempty"`                                   // true means allow Local Peer Discovery in public torrents
	PeerLimitGlobal           int64      `json:"peer-limit-global,omitempty" yaml:"peer-limit-global,omitempty"`                       // maximum global number of peers
	PeerLimitPerTorrent       int64      `json:"peer-limit-per-torrent,omitempty" yaml:"peer-limit-per-torrent,omitempty"`             // maximum global number of peers
	PexEnabled                bool       `json:"pex-enabled,omitempty" yaml:"pex-enabled,omitempty"`                                   // true means allow pex in public torrents
	PeerPort                  int64      `json:"peer-port,omitempty" yaml:"peer-port,omitempty"`                                       // port number
	PeerPortRandomOnStart     bool       `json:"peer-port-random-on-start,omitempty" yaml:"peer-port-random-on-start,omitempty"`       // true means pick a random peer port on launch
	PortForwardingEnabled     bool       `json:"port-forwarding-enabled,omitempty" yaml:"port-forwarding-enabled,omitempty"`           // true means enabled
	QueueStalledEnabled       bool       `json:"queue-stalled-enabled,omitempty" yaml:"queue-stalled-enabled,omitempty"`               // whether or not to consider idle torrents as stalled
	QueueStalledMinutes       int64      `json:"queue-stalled-minutes,omitempty" yaml:"queue-stalled-minutes,omitempty"`               // torrents that are idle for N minutes aren't counted toward seed-queue-size or download-queue-size
	RenamePartialFiles        bool       `json:"rename-partial-files,omitempty" yaml:"rename-partial-files,omitempty"`                 // true means append ".part" to incomplete files
	RPCVersion                int64      `json:"rpc-version,omitempty" yaml:"rpc-version,omitempty"`                                   // the current RPC API version
	RPCVersionMinimum         int64      `json:"rpc-version-minimum,omitempty" yaml:"rpc-version-minimum,omitempty"`                   // the minimum RPC API version supported
	ScriptTorrentDoneFilename string     `json:"script-torrent-done-filename,omitempty" yaml:"script-torrent-done-filename,omitempty"` // filename of the script to run
	ScriptTorrentDoneEnabled  bool       `json:"script-torrent-done-enabled,omitempty" yaml:"script-torrent-done-enabled,omitempty"`   // whether or not to call the "done" script
	SeedRatioLimit            float64    `json:"seedRatioLimit,omitempty" yaml:"seedRatioLimit,omitempty"`                             // the default seed ratio for torrents to use
	SeedRatioLimited          bool       `json:"seedRatioLimited,omitempty" yaml:"seedRatioLimited,omitempty"`                         // true if seedRatioLimit is honored by default
	SeedQueueSize             int64      `json:"seed-queue-size,omitempty" yaml:"seed-queue-size,omitempty"`                           // max number of torrents to uploaded at once (see seed-queue-enabled)
	SeedQueueEnabled          bool       `json:"seed-queue-enabled,omitempty" yaml:"seed-queue-enabled,omitempty"`                     // if true, limit how many torrents can be uploaded at once
	SpeedLimitDown            int64      `json:"speed-limit-down,omitempty" yaml:"speed-limit-down,omitempty"`                         // max global download speed (KBps)
	SpeedLimitDownEnabled     bool       `json:"speed-limit-down-enabled,omitempty" yaml:"speed-limit-down-enabled,omitempty"`         // true means enabled
	SpeedLimitUp              int64      `json:"speed-limit-up,omitempty" yaml:"speed-limit-up,omitempty"`                             // max global upload speed (KBps)
	SpeedLimitUpEnabled       bool       `json:"speed-limit-up-enabled,omitempty" yaml:"speed-limit-up-enabled,omitempty"`             // true means enabled
	StartAddedTorrents        bool       `json:"start-added-torrents,omitempty" yaml:"start-added-torrents,omitempty"`                 // true means added torrents will be started right away
	TrashOriginalTorrentFiles bool       `json:"trash-original-torrent-files,omitempty" yaml:"trash-original-torrent-files,omitempty"` // true means the .torrent file of added torrents will be deleted
	Units                     Units      `json:"units,omitempty" yaml:"units,omitempty"`                                               // see units below
	UtpEnabled                bool       `json:"utp-enabled,omitempty" yaml:"utp-enabled,omitempty"`                                   // true means allow utp
	Version                   string     `json:"version,omitempty" yaml:"version,omitempty"`                                           // long version string "$version ($revision)"
}

// Units are session units.
type Units struct {
	SpeedUnits  []string `json:"speed-units,omitempty" yaml:"speed-units,omitempty"`   // 4 strings: KB/s, MB/s, GB/s, TB/s
	SpeedBytes  int64    `json:"speed-bytes,omitempty" yaml:"speed-bytes,omitempty"`   // number of bytes in a KB (1000 for kB; 1024 for KiB)
	SizeUnits   []string `json:"size-units,omitempty" yaml:"size-units,omitempty"`     // 4 strings: KB/s, MB/s, GB/s, TB/s
	SizeBytes   int64    `json:"size-bytes,omitempty" yaml:"size-bytes,omitempty"`     // number of bytes in a KB (1000 for kB; 1024 for KiB)
	MemoryUnits []string `json:"memory-units,omitempty" yaml:"memory-units,omitempty"` // 4 strings: KB/s, MB/s, GB/s, TB/s
	MemoryBytes int64    `json:"memory-bytes,omitempty" yaml:"memory-bytes,omitempty"` // number of bytes in a KB (1000 for kB; 1024 for KiB)
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
	if req.changed["RPCVersion"] {
		params["rpc-version"] = req.RPCVersion
	}
	if req.changed["RPCVersionMinimum"] {
		params["rpc-version-minimum"] = req.RPCVersionMinimum
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
	ActiveTorrentCount int64     `json:"activeTorrentCount,omitempty" yaml:"activeTorrentCount,omitempty"`
	DownloadSpeed      ByteCount `json:"downloadSpeed,omitempty" yaml:"downloadSpeed,omitempty"`
	PausedTorrentCount int64     `json:"pausedTorrentCount,omitempty" yaml:"pausedTorrentCount,omitempty"`
	TorrentCount       int64     `json:"torrentCount,omitempty" yaml:"torrentCount,omitempty"`
	UploadSpeed        ByteCount `json:"uploadSpeed,omitempty" yaml:"uploadSpeed,omitempty"`
	CumulativeStats    struct {
		UploadedBytes   ByteCount `json:"uploadedBytes,omitempty" yaml:"uploadedBytes,omitempty"`     // tr_session_stats
		DownloadedBytes ByteCount `json:"downloadedBytes,omitempty" yaml:"downloadedBytes,omitempty"` // tr_session_stats
		FilesAdded      int64     `json:"filesAdded,omitempty" yaml:"filesAdded,omitempty"`           // tr_session_stats
		SessionCount    int64     `json:"sessionCount,omitempty" yaml:"sessionCount,omitempty"`       // tr_session_stats
		SecondsActive   Duration  `json:"secondsActive,omitempty" yaml:"secondsActive,omitempty"`     // tr_session_stats
	} `json:"cumulative-stats,omitempty" yaml:"cumulative-stats,omitempty"`
	CurrentStats struct {
		UploadedBytes   ByteCount `json:"uploadedBytes,omitempty" yaml:"uploadedBytes,omitempty"`     // tr_session_stats
		DownloadedBytes ByteCount `json:"downloadedBytes,omitempty" yaml:"downloadedBytes,omitempty"` // tr_session_stats
		FilesAdded      int64     `json:"filesAdded,omitempty" yaml:"filesAdded,omitempty"`           // tr_session_stats
		SessionCount    int64     `json:"sessionCount,omitempty" yaml:"sessionCount,omitempty"`       // tr_session_stats
		SecondsActive   Duration  `json:"secondsActive,omitempty" yaml:"secondsActive,omitempty"`     // tr_session_stats
	} `json:"current-stats,omitempty" yaml:"current-stats,omitempty"`
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
		BlocklistSize int64 `json:"blocklist-size,omitempty" yaml:"blocklist-size,omitempty"`
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
		PortIsOpen bool `json:"port-is-open,omitempty" yaml:"port-is-open,omitempty"`
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
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
}

// FreeSpace creates a free space request.
func FreeSpace(path string) *FreeSpaceRequest {
	return &FreeSpaceRequest{
		Path: path,
	}
}

// Do executes the free space request against the provided context and client.
func (req *FreeSpaceRequest) Do(ctx context.Context, cl *Client) (ByteCount, error) {
	var res struct {
		Path      string    `json:"path,omitempty" yaml:"path,omitempty"`
		SizeBytes ByteCount `json:"size-bytes,omitempty" yaml:"size-bytes,omitempty"`
	}
	if err := cl.Do(ctx, "free-space", req, &res); err != nil {
		return 0, err
	}
	return res.SizeBytes, nil
}
