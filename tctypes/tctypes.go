// Package tctypes contains the shared types for client implementations.
package tctypes

import (
	"fmt"
	"strconv"
	"time"
)

//go:generate stringer -type Status -trimprefix Status
//go:generate stringer -type Priority -trimprefix Priority
//go:generate stringer -type Mode -trimprefix Mode
//go:generate stringer -type State -trimprefix State

// DefaultAsIEC toggles whether the IEC format is used by default for byte
// counts/rates/limits' string conversion.
var DefaultAsIEC bool = true

// ByteFormatter is the shared interface for byte counts, rates, and limits.
type ByteFormatter interface {
	// Int64 returns the actual value in a int64.
	Int64() int64

	// Format formats the value in a human readable string, with the passed
	// precision.
	Format(bool, int) string

	// String satisfies the fmt.Stringer interface, returning a human readable
	// string.
	String() string

	// Add adds the passed value to itself, returning the sum of the two values.
	Add(interface{}) interface{}
}

// Format formats size i using the supplied precision, as a human readable
// string (1 B, 2 kB, 3 MB, 5 GB, ...). When asIEC is true, will format the
// amount as a IEC size (1 B, 2 KiB, 4 MiB, 5 GiB, ...)
func Format(i int64, asIEC bool, precision int) string {
	c, sizes, end := int64(1000), "kMGTPEZY", "B"
	if asIEC {
		c, sizes, end = 1024, "KMGTPEZY", "iB"
	}
	if i < c {
		return fmt.Sprintf("%d B", i)
	}
	exp, div := 0, c
	for n := i / c; n >= c; n /= c {
		div *= c
		exp++
	}
	return fmt.Sprintf("%."+strconv.Itoa(precision)+"f %c%s", float64(i)/float64(div), sizes[exp], end)
}

// ByteCount wraps a byte count as int64.
type ByteCount int64

// Int64 returns the byte count as an int64.
func (bc ByteCount) Int64() int64 {
	return int64(bc)
}

// Format formats the byte count.
func (bc ByteCount) Format(asIEC bool, prec int) string {
	return Format(int64(bc), asIEC, prec)
}

// String satisfies the fmt.Stringer interface.
func (bc ByteCount) String() string {
	return bc.Format(DefaultAsIEC, 2)
}

// Add adds i to the byte count.
func (bc ByteCount) Add(i interface{}) interface{} {
	return bc + i.(ByteCount)
}

// Rate is a bytes per second rate.
type Rate int64

// Int64 returns the rate as an int64.
func (r Rate) Int64() int64 {
	return int64(r)
}

// Format formats the rate.
func (r Rate) Format(asIEC bool, prec int) string {
	return Format(int64(r), asIEC, prec) + "/s"
}

// String satisfies the fmt.Stringer interface.
func (r Rate) String() string {
	return r.Format(DefaultAsIEC, 2)
}

// Add adds i to the byte count.
func (r Rate) Add(i interface{}) interface{} {
	return r + i.(Rate)
}

// Limit is a K bytes per second limit.
type Limit int64

// Int64 returns the rate as an int64.
func (l Limit) Int64() int64 {
	return int64(l)
}

// Format formats the rate.
func (l Limit) Format(asIEC bool, prec int) string {
	return Format(int64(l*1000), asIEC, prec) + "/s"
}

// String satisfies the fmt.Stringer interface.
func (l Limit) String() string {
	return l.Format(DefaultAsIEC, 2)
}

// Add adds i to the byte count.
func (l Limit) Add(i interface{}) interface{} {
	return l + i.(Limit)
}

// KiLimit is a K bytes per second limit.
type KiLimit int64

// Int64 returns the rate as an int64.
func (l KiLimit) Int64() int64 {
	return int64(l)
}

// Format formats the rate.
func (l KiLimit) Format(asIEC bool, prec int) string {
	return Format(int64(l*1024), asIEC, prec) + "/s"
}

// String satisfies the fmt.Stringer interface.
func (l KiLimit) String() string {
	return l.Format(DefaultAsIEC, 2)
}

// Add adds i to the byte count.
func (l KiLimit) Add(i interface{}) interface{} {
	return l + i.(KiLimit)
}

// Percent wraps a float64.
type Percent float64

// String satisfies the fmt.Stringer interface.
func (p Percent) String() string {
	return fmt.Sprintf("%.f%%", float64(p)*100)
}

// Time wraps time.Time.
type Time time.Time

// String satisfies the fmt.Stringer interface.
func (t Time) String() string {
	if time.Time(t).IsZero() {
		return ""
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
	if i <= 0 {
		return nil
	}
	*t = Time(time.Unix(i, 0))
	return nil
}

// MarshalJSON satisfies the json.Marshaler interface.
func (t Time) MarshalJSON() ([]byte, error) {
	if time.Time(t).IsZero() {
		return []byte("-1"), nil
	}
	return []byte(strconv.FormatInt(time.Time(t).Unix(), 10)), nil
}

// MarshalYAML satisfies the yaml.Marshaler interface.
func (t Time) MarshalYAML() (interface{}, error) {
	if time.Time(t).IsZero() {
		return -1, nil
	}
	return time.Time(t).Unix(), nil
}

// MilliTime wraps time.Time.
type MilliTime time.Time

// String satisfies the fmt.Stringer interface.
func (t MilliTime) String() string {
	if time.Time(t).IsZero() {
		return ""
	}
	return time.Time(t).Format("2006-01-02 15:04:05")
}

// UnmarshalJSON satisfies the json.Unmarshaler interface.
func (t *MilliTime) UnmarshalJSON(buf []byte) error {
	if len(buf) == 0 {
		return ErrInvalidTime
	}
	i, err := strconv.ParseInt(string(buf), 10, 64)
	if err != nil {
		return err
	}
	if i <= 0 {
		return nil
	}
	*t = MilliTime(time.Unix(0, i*int64(time.Millisecond)))
	return nil
}

// MarshalJSON satisfies the json.Marshaler interface.
func (t MilliTime) MarshalJSON() ([]byte, error) {
	if time.Time(t).IsZero() {
		return []byte("-1"), nil
	}
	return []byte(strconv.FormatInt(time.Time(t).Unix()/int64(time.Millisecond), 10)), nil
}

// MarshalYAML satisfies the yaml.Marshaler interface.
func (t MilliTime) MarshalYAML() (interface{}, error) {
	if time.Time(t).IsZero() {
		return -1, nil
	}
	return time.Time(t).UnixNano() / int64(time.Millisecond), nil
}

// Duration wraps time.Duration.
type Duration time.Duration

// String satisfies the fmt.Stringer interface.
func (d Duration) String() string {
	switch v := time.Duration(d); v {
	case -1 * time.Second:
		return "Done"
	case -2 * time.Second:
		return "" // "Unknown"
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

// MilliDuration wraps time.Duration.
type MilliDuration time.Duration

// String satisfies the fmt.Stringer interface.
func (d MilliDuration) String() string {
	switch v := time.Duration(d); v {
	case -1 * time.Millisecond:
		return "Done"
	case -2 * time.Millisecond:
		return "" // "Unknown"
	default:
		return v.String()
	}
}

// UnmarshalJSON satisfies the json.Unmarshaler interface.
func (d *MilliDuration) UnmarshalJSON(buf []byte) error {
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
	*d = MilliDuration(i * int64(time.Millisecond))
	return nil
}

// MarshalJSON satisfies the json.Marshaler interface.
func (d MilliDuration) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(int64(time.Duration(d)/time.Millisecond), 10)), nil
}

// MarshalYAML satisfies the yaml.Marshaler interface.
func (d MilliDuration) MarshalYAML() (interface{}, error) {
	return int64(time.Duration(d) / time.Millisecond), nil
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
	if bool(b) {
		return []byte("1"), nil
	}
	return []byte("0"), nil
}

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

// ErrNo wraps remote errors.
type ErrNo int64

// String satisifies the fmt.Stringer interface.
func (errNo ErrNo) String() string {
	if errNo == 0 {
		return ""
	}
	return strconv.FormatInt(int64(errNo), 10)
}

// Torrent holds information about a torrent.
type Torrent struct {
	ActivityDate      Time      `json:"activityDate,omitempty" yaml:"activityDate,omitempty"`           // tr_stat
	AddedDate         Time      `json:"addedDate,omitempty" yaml:"addedDate,omitempty"`                 // tr_stat
	BandwidthPriority Priority  `json:"bandwidthPriority,omitempty" yaml:"bandwidthPriority,omitempty"` // tr_priority_t
	Comment           string    `json:"comment,omitempty" yaml:"comment,omitempty"`                     // tr_info
	CorruptEver       ByteCount `json:"corruptEver,omitempty" yaml:"corruptEver,omitempty"`             // tr_stat
	Creator           string    `json:"creator,omitempty" yaml:"creator,omitempty"`                     // tr_info
	DateCreated       Time      `json:"dateCreated,omitempty" yaml:"dateCreated,omitempty"`             // tr_info
	DesiredAvailable  ByteCount `json:"desiredAvailable,omitempty" yaml:"desiredAvailable,omitempty"`   // tr_stat
	DoneDate          Time      `json:"doneDate,omitempty" yaml:"doneDate,omitempty"`                   // tr_stat
	DownloadDir       string    `json:"downloadDir,omitempty" yaml:"downloadDir,omitempty"`             // tr_torrent
	DownloadedEver    ByteCount `json:"downloadedEver,omitempty" yaml:"downloadedEver,omitempty"`       // tr_stat
	DownloadLimit     Limit     `json:"downloadLimit,omitempty" yaml:"downloadLimit,omitempty"`         // tr_torrent
	DownloadLimited   bool      `json:"downloadLimited,omitempty" yaml:"downloadLimited,omitempty"`     // tr_torrent
	EditDate          Time      `json:"editDate,omitempty" yaml:"editDate,omitempty"`                   // tr_stat
	Error             ErrNo     `json:"error,omitempty" yaml:"error,omitempty"`                         // tr_stat
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
	ManualAnnounceTime      Time      `json:"manualAnnounceTime,omitempty" yaml:"manualAnnounceTime,omitempty"`           // tr_stat
	MaxConnectedPeers       int64     `json:"maxConnectedPeers,omitempty" yaml:"maxConnectedPeers,omitempty"`             // tr_torrent
	MetadataPercentComplete Percent   `json:"metadataPercentComplete,omitempty" yaml:"metadataPercentComplete,omitempty"` // tr_stat
	Name                    string    `json:"name,omitempty" yaml:"name,omitempty"`                                       // tr_info
	PeerLimit               int64     `json:"peer-limit,omitempty" yaml:"peer-limit,omitempty"`                           // tr_torrent
	Peers                   []struct {
		Address            string  `json:"address,omitempty" yaml:"address,omitempty"`                       // tr_peer_stat
		ClientName         string  `json:"clientName,omitempty" yaml:"clientName,omitempty"`                 // tr_peer_stat
		ClientIsChoked     bool    `json:"clientIsChoked,omitempty" yaml:"clientIsChoked,omitempty"`         // tr_peer_stat
		ClientIsInterested bool    `json:"clientIsInterested,omitempty" yaml:"clientIsInterested,omitempty"` // tr_peer_stat
		FlagStr            string  `json:"flagStr,omitempty" yaml:"flagStr,omitempty"`                       // tr_peer_stat
		IsDownloadingFrom  bool    `json:"isDownloadingFrom,omitempty" yaml:"isDownloadingFrom,omitempty"`   // tr_peer_stat
		IsEncrypted        bool    `json:"isEncrypted,omitempty" yaml:"isEncrypted,omitempty"`               // tr_peer_stat
		IsIncoming         bool    `json:"isIncoming,omitempty" yaml:"isIncoming,omitempty"`                 // tr_peer_stat
		IsUploadingTo      bool    `json:"isUploadingTo,omitempty" yaml:"isUploadingTo,omitempty"`           // tr_peer_stat
		IsUTP              bool    `json:"isUTP,omitempty" yaml:"isUTP,omitempty"`                           // tr_peer_stat
		PeerIsChoked       bool    `json:"peerIsChoked,omitempty" yaml:"peerIsChoked,omitempty"`             // tr_peer_stat
		PeerIsInterested   bool    `json:"peerIsInterested,omitempty" yaml:"peerIsInterested,omitempty"`     // tr_peer_stat
		Port               int64   `json:"port,omitempty" yaml:"port,omitempty"`                             // tr_peer_stat
		Progress           Percent `json:"progress,omitempty" yaml:"progress,omitempty"`                     // tr_peer_stat
		RateToClient       Rate    `json:"rateToClient,omitempty" yaml:"rateToClient,omitempty"`             // tr_peer_stat
		RateToPeer         Rate    `json:"rateToPeer,omitempty" yaml:"rateToPeer,omitempty"`                 // tr_peer_stat
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
	RateDownload       Rate       `json:"rateDownload,omitempty" yaml:"rateDownload,omitempty"`             // tr_stat
	RateUpload         Rate       `json:"rateUpload,omitempty" yaml:"rateUpload,omitempty"`                 // tr_stat
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
	UploadLimit         Limit     `json:"uploadLimit,omitempty" yaml:"uploadLimit,omitempty"`                 // tr_torrent
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

// File is combined fields of files, fileStats from a torrent.
type File struct {
	BytesCompleted ByteCount `json:"bytesCompleted,omitempty" yaml:"bytesCompleted,omitempty"` // tr_torrent
	Length         ByteCount `json:"length,omitempty" yaml:"length,omitempty"`                 // tr_info
	Name           string    `json:"name,omitempty" yaml:"name,omitempty"`                     // tr_info
	Wanted         bool      `json:"wanted,omitempty" yaml:"wanted,omitempty"`                 // tr_info
	Priority       string    `json:"priority,omitempty" yaml:"priority,omitempty"`             // tr_info
	ID             int64     `json:"id" yaml:"id"`
	Torrent        string    `json:"-" yaml:"-" all:"torrent"`
	HashString     string    `json:"-" yaml:"-" all:"hashString"`
}

// ShortHash returns the short hash of the torrent.
func (f File) ShortHash() string {
	if len(f.HashString) < 7 {
		return ""
	}
	return f.HashString[:7]
}

// DoPercentDone calculates the done percent for the file.
func (f File) PercentDone() Percent {
	if f.Length == 0 {
		return 1.0
	}
	return Percent(float64(f.BytesCompleted) / float64(f.Length))
}

// Peer is a peer.
type Peer struct {
	Address            string  `json:"address,omitempty" yaml:"address,omitempty"`                       // tr_peer_stat
	ClientName         string  `json:"clientName,omitempty" yaml:"clientName,omitempty"`                 // tr_peer_stat
	ClientIsChoked     bool    `json:"clientIsChoked,omitempty" yaml:"clientIsChoked,omitempty"`         // tr_peer_stat
	ClientIsInterested bool    `json:"clientIsInterested,omitempty" yaml:"clientIsInterested,omitempty"` // tr_peer_stat
	FlagStr            string  `json:"flagStr,omitempty" yaml:"flagStr,omitempty"`                       // tr_peer_stat
	IsDownloadingFrom  bool    `json:"isDownloadingFrom,omitempty" yaml:"isDownloadingFrom,omitempty"`   // tr_peer_stat
	IsEncrypted        bool    `json:"isEncrypted,omitempty" yaml:"isEncrypted,omitempty"`               // tr_peer_stat
	IsIncoming         bool    `json:"isIncoming,omitempty" yaml:"isIncoming,omitempty"`                 // tr_peer_stat
	IsUploadingTo      bool    `json:"isUploadingTo,omitempty" yaml:"isUploadingTo,omitempty"`           // tr_peer_stat
	IsUTP              bool    `json:"isUTP,omitempty" yaml:"isUTP,omitempty"`                           // tr_peer_stat
	PeerIsChoked       bool    `json:"peerIsChoked,omitempty" yaml:"peerIsChoked,omitempty"`             // tr_peer_stat
	PeerIsInterested   bool    `json:"peerIsInterested,omitempty" yaml:"peerIsInterested,omitempty"`     // tr_peer_stat
	Port               int64   `json:"port,omitempty" yaml:"port,omitempty"`                             // tr_peer_stat
	Progress           Percent `json:"progress,omitempty" yaml:"progress,omitempty"`                     // tr_peer_stat
	RateToClient       Rate    `json:"rateToClient,omitempty" yaml:"rateToClient,omitempty"`             // tr_peer_stat
	RateToPeer         Rate    `json:"rateToPeer,omitempty" yaml:"rateToPeer,omitempty"`                 // tr_peer_stat
	ID                 int64   `json:"id" yaml:"id"`
	Torrent            string  `json:"-" yaml:"-" all:"torrent"`
	HashString         string  `json:"-" yaml:"-" all:"hashString"`
}

// ShortHash returns the short hash of the torrent.
func (p Peer) ShortHash() string {
	if len(p.HashString) < 7 {
		return ""
	}
	return p.HashString[:7]
}

// Tracker is combined fields of trackers, trackerStats from a torrent.
type Tracker struct {
	Announce              string `json:"announce,omitempty" yaml:"announce,omitempty"`                           // tr_tracker_info
	ID                    int64  `json:"id" yaml:"id"`                                                           // tr_tracker_info
	Scrape                string `json:"scrape,omitempty" yaml:"scrape,omitempty"`                               // tr_tracker_info
	Tier                  int64  `json:"tier,omitempty" yaml:"tier,omitempty"`                                   // tr_tracker_info
	AnnounceState         State  `json:"announceState,omitempty" yaml:"announceState,omitempty"`                 // tr_tracker_stat
	DownloadCount         int64  `json:"downloadCount,omitempty" yaml:"downloadCount,omitempty"`                 // tr_tracker_stat
	HasAnnounced          bool   `json:"hasAnnounced,omitempty" yaml:"hasAnnounced,omitempty"`                   // tr_tracker_stat
	HasScraped            bool   `json:"hasScraped,omitempty" yaml:"hasScraped,omitempty"`                       // tr_tracker_stat
	Host                  string `json:"host,omitempty" yaml:"host,omitempty"`                                   // tr_tracker_stat
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
	ScrapeState           State  `json:"scrapeState,omitempty" yaml:"scrapeState,omitempty"`                     // tr_tracker_stat
	SeederCount           int64  `json:"seederCount,omitempty" yaml:"seederCount,omitempty"`                     // tr_tracker_stat
	Torrent               string `json:"-" yaml:"-" all:"torrent"`
	HashString            string `json:"-" yaml:"-" all:"hashString"`
}

// ShortHash returns the short hash of the torrent.
func (t Tracker) ShortHash() string {
	if len(t.HashString) < 7 {
		return ""
	}
	return t.HashString[:7]
}

// Error is an error.
type Error string

// Error satisfies the error interface.
func (err Error) Error() string {
	return string(err)
}

const (
	// ErrInvalidTime is the invalid time error.
	ErrInvalidTime Error = "invalid time"

	// ErrInvalidDuration is the invalid duration error.
	ErrInvalidDuration Error = "invalid duration"

	// ErrInvalidBool is the invalid bool error.
	ErrInvalidBool Error = "invalid bool"

	// ErrInvalidPriority is the invalid priority error.
	ErrInvalidPriority Error = "invalid priority"

	// ErrInvalidStatus is the invalid status error.
	ErrInvalidStatus Error = "invalid status"

	// ErrInvalidMode is the invalid mode error.
	ErrInvalidMode Error = "invalid mode"

	// ErrInvalidState is the invalid state error.
	ErrInvalidState Error = "invalid state"

	// ErrInvalidEncryption is the invalid encryption error.
	ErrInvalidEncryption Error = "invalid encryption"
)
