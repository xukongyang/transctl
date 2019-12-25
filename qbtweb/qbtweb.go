// Package qbtweb provides a idiomatic Go client for qBittorrent web.
//
// See: https://github.com/qbittorrent/qBittorrent/wiki/Web-API-Documentation
package qbtweb

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/url"
	"strconv"
	"strings"

	"github.com/kenshaw/transctl/tcutil"
)

//go:generate stringer -type DaySchedule
//go:generate stringer -type ProxyType -trimprefix Proxy
//go:generate stringer -type ServiceType -trimprefix Service
//go:generate stringer -type BehaviorType -trimprefix Behavior
//go:generate stringer -type Encryption -trimprefix Encryption
//go:generate stringer -type Protocol -trimprefix Protocol
//go:generate stringer -type LogType -trimprefix Log
//go:generate stringer -type TrackerStatus -trimprefix Tracker
//go:generate stringer -type FilePriority -trimprefix FilePriority
//go:generate stringer -type PieceState -trimprefix PieceState

// Type aliases
type (
	// ByteCount is a byte count.
	ByteCount = tcutil.ByteCount

	// Rate is a byte per second rate.
	Rate = tcutil.Rate

	// Limit is a KB/s limit.
	// Limit = tcutil.Limit

	// Limit is a KiB/s limit.
	KiLimit = tcutil.KiLimit

	// Percent is a percent.
	Percent = tcutil.Percent

	// Time wraps time.Time.
	Time = tcutil.Time

	// MilliTime wraps time.Time.
	MilliTime = tcutil.MilliTime

	// Duration wraps time.Duration.
	Duration = tcutil.Duration

	// Bool wraps bool.
	Bool = tcutil.Bool
)

// State is the state enum.
type State string

// State values.
const (
	StateError              State = "error"              // Some error occurred, applies to paused torrents
	StateMissingFiles       State = "missingFiles"       // Torrent data files is missing
	StateUploading          State = "uploading"          // Torrent is being seeded and data is being transferred
	StatePausedUP           State = "pausedUP"           // Torrent is paused and has finished downloading
	StateQueuedUP           State = "queuedUP"           // Queuing is enabled and torrent is queued for upload
	StateStalledUP          State = "stalledUP"          // Torrent is being seeded, but no connection were made
	StateCheckingUP         State = "checkingUP"         // Torrent has finished downloading and is being checked
	StateForcedUP           State = "forcedUP"           // Torrent is forced to uploading and ignore queue limit
	StateAllocating         State = "allocating"         // Torrent is allocating disk space for download
	StateDownloading        State = "downloading"        // Torrent is being downloaded and data is being transferred
	StateMetaDL             State = "metaDL"             // Torrent has just started downloading and is fetching metadata
	StatePausedDL           State = "pausedDL"           // Torrent is paused and has NOT finished downloading
	StateQueuedDL           State = "queuedDL"           // Queuing is enabled and torrent is queued for download
	StateStalledDL          State = "stalledDL"          // Torrent is being downloaded, but no connection were made
	StateCheckingDL         State = "checkingDL"         // Same as checkingUP, but torrent has NOT finished downloading
	StateForceDL            State = "forceDL"            // Torrent is forced to downloading to ignore queue limit
	StateCheckingResumeData State = "checkingResumeData" // Checking resume data on qBt startup
	StateMoving             State = "moving"             // Torrent is moving to another location
	StateUnknown            State = "unknown"            // Unknown status
)

// String satisfies the fmt.Stringer interface.
func (s State) String() string {
	return string(s)
}

// DaySchedule is the day schedule enum.
type DaySchedule int

// Day schedule values.
const (
	EveryDay DaySchedule = iota
	EveryWeekday
	EveryWeekend
	EveryMonday
	EveryTuesday
	EveryWednesday
	EveryThursday
	EveryFriday
	EverySaturday
	EverySunday
)

// ProxyType is the proxy type enum.
type ProxyType int

// Proxy type values.
const (
	ProxyDisabled                  ProxyType = -1
	ProxyHTTPWithoutAuthentication ProxyType = iota + 1
	ProxySOCKS5WithoutAuthentication
	ProxyHTTPWithAuthentication
	ProxySOCKS5WithAuthentication
	ProxySOCKS4WithoutAuthentication
)

// ServiceType is the dyndns service type enum.
type ServiceType int

// Dyndns service values.
const (
	ServiceUseDyDNS ServiceType = iota
	ServiceUseNOIP
)

// BehaviorType is the max ratio stop behavior enum.
type BehaviorType int

// Max ratio stop behavior values.
const (
	BehaviorPause BehaviorType = iota
	BehaviorRemove
)

// Encryption is the encryption enum.
type Encryption int

// Encryption values.
const (
	EncryptionPreferred Encryption = iota
	EncryptionForceOn
	EncryptionForceOff
)

// Protocol is the protocol enum.
type Protocol int

// Protocol values.
const (
	ProtocolBoth Protocol = iota
	ProtocolTCP
	ProtocolUTP
)

// LogType is the log type enum.
type LogType int

// Log type values.
const (
	LogNormal LogType = 1 << iota
	LogInfo
	LogWarning
	LogCritical
)

// ConnectionStatus is the connection status enum.
type ConnectionStatus string

// Connection status values.
const (
	ConnectionConnected    ConnectionStatus = "connected"
	ConnectionFirewalled   ConnectionStatus = "firewalled"
	ConnectionDisconnected ConnectionStatus = "disconnected"
)

// TrackerStatus is the tracker status enum.
type TrackerStatus int

// Tracker status values.
const (
	TrackerDisabled TrackerStatus = iota
	TrackerNotYetContacted
	TrackerContactedAndWorking
	TrackerUpdating
	TrackerNotWorking
)

// Torrent holds information about a torrent.
type Torrent struct {
	AddedOn           Time      `json:"added_on,omitempty" yaml:"added_on,omitempty"`                     // Time (Unix Epoch) when the torrent was added to the client
	AmountLeft        ByteCount `json:"amount_left,omitempty" yaml:"amount_left,omitempty"`               // Amount of data left to download (bytes)
	Availability      Percent   `json:"availability,omitempty" yaml:"availability,omitempty"`             // not documented
	AutoTmm           bool      `json:"auto_tmm,omitempty" yaml:"auto_tmm,omitempty"`                     // Whether this torrent is managed by Automatic Torrent Management
	Category          string    `json:"category,omitempty" yaml:"category,omitempty"`                     // Category of the torrent
	Completed         ByteCount `json:"completed,omitempty" yaml:"completed,omitempty"`                   // Amount of transfer data completed (bytes)
	CompletionOn      Time      `json:"completion_on,omitempty" yaml:"completion_on,omitempty"`           // Time (Unix Epoch) when the torrent completed
	DlLimit           Rate      `json:"dl_limit,omitempty" yaml:"dl_limit,omitempty"`                     // Torrent download speed limit (bytes/s). -1 if ulimited.
	Dlspeed           Rate      `json:"dlspeed,omitempty" yaml:"dlspeed,omitempty"`                       // Torrent download speed (bytes/s)
	Downloaded        ByteCount `json:"downloaded,omitempty" yaml:"downloaded,omitempty"`                 // Amount of data downloaded
	DownloadedSession ByteCount `json:"downloaded_session,omitempty" yaml:"downloaded_session,omitempty"` // Amount of data downloaded this session
	Eta               Duration  `json:"eta,omitempty" yaml:"eta,omitempty"`                               // Torrent ETA (seconds)
	FLPiecePrio       bool      `json:"f_l_piece_prio,omitempty" yaml:"f_l_piece_prio,omitempty"`         // True if first last piece are prioritized
	ForceStart        bool      `json:"force_start,omitempty" yaml:"force_start,omitempty"`               // True if force start is enabled for this torrent
	Hash              string    `json:"hash,omitempty" yaml:"hash,omitempty"`                             // Torrent hash
	LastActivity      Time      `json:"last_activity,omitempty" yaml:"last_activity,omitempty"`           // Last time (Unix Epoch) when a chunk was downloaded/uploaded
	MagnetURI         string    `json:"magnet_uri,omitempty" yaml:"magnet_uri,omitempty"`                 // Magnet URI corresponding to this torrent
	MaxRatio          Percent   `json:"max_ratio,omitempty" yaml:"max_ratio,omitempty"`                   // Maximum share ratio until torrent is stopped from seeding/uploading
	MaxSeedingTime    Duration  `json:"max_seeding_time,omitempty" yaml:"max_seeding_time,omitempty"`     // Maximum seeding time (seconds) until torrent is stopped from seeding
	Name              string    `json:"name,omitempty" yaml:"name,omitempty"`                             // Torrent name
	NumComplete       int64     `json:"num_complete,omitempty" yaml:"num_complete,omitempty"`             // Number of seeds in the swarm
	NumIncomplete     int64     `json:"num_incomplete,omitempty" yaml:"num_incomplete,omitempty"`         // Number of leechers in the swarm
	NumLeechs         int64     `json:"num_leechs,omitempty" yaml:"num_leechs,omitempty"`                 // Number of leechers connected to
	NumSeeds          int64     `json:"num_seeds,omitempty" yaml:"num_seeds,omitempty"`                   // Number of seeds connected to
	Priority          int64     `json:"priority,omitempty" yaml:"priority,omitempty"`                     // Torrent priority. Returns -1 if queuing is disabled or torrent is in seed mode
	Progress          Percent   `json:"progress,omitempty" yaml:"progress,omitempty"`                     // Torrent progress (percentage/100)
	Ratio             Percent   `json:"ratio,omitempty" yaml:"ratio,omitempty"`                           // Torrent share ratio. Max ratio value: 9999.
	RatioLimit        Percent   `json:"ratio_limit,omitempty" yaml:"ratio_limit,omitempty"`               // TODO (what is different from max_ratio?)
	SavePath          string    `json:"save_path,omitempty" yaml:"save_path,omitempty"`                   // Path where this torrent's data is stored
	SeedingTimeLimit  Duration  `json:"seeding_time_limit,omitempty" yaml:"seeding_time_limit,omitempty"` // TODO (what is different from max_seeding_time?)
	SeenComplete      Time      `json:"seen_complete,omitempty" yaml:"seen_complete,omitempty"`           // Time (Unix Epoch) when this torrent was last seen complete
	SeqDl             bool      `json:"seq_dl,omitempty" yaml:"seq_dl,omitempty"`                         // True if sequential download is enabled
	Size              ByteCount `json:"size,omitempty" yaml:"size,omitempty"`                             // Total size (bytes) of files selected for download
	State             State     `json:"state,omitempty" yaml:"state,omitempty"`                           // Torrent state. See table here below for the possible values
	SuperSeeding      bool      `json:"super_seeding,omitempty" yaml:"super_seeding,omitempty"`           // True if super seeding is enabled
	Tags              string    `json:"tags,omitempty" yaml:"tags,omitempty"`                             // Comma-concatenated tag list of the torrent
	TimeActive        Duration  `json:"time_active,omitempty" yaml:"time_active,omitempty"`               // Total active time (seconds)
	TotalSize         ByteCount `json:"total_size,omitempty" yaml:"total_size,omitempty"`                 // Total size (bytes) of all file in this torrent (including unselected ones)
	Tracker           string    `json:"tracker,omitempty" yaml:"tracker,omitempty"`                       // The first tracker with working status. (TODO: what is returned if no tracker is working?)
	UpLimit           Rate      `json:"up_limit,omitempty" yaml:"up_limit,omitempty"`                     // Torrent upload speed limit (bytes/s). -1 if ulimited.
	Uploaded          ByteCount `json:"uploaded,omitempty" yaml:"uploaded,omitempty"`                     // Amount of data uploaded
	UploadedSession   ByteCount `json:"uploaded_session,omitempty" yaml:"uploaded_session,omitempty"`     // Amount of data uploaded this session
	Upspeed           Rate      `json:"upspeed,omitempty" yaml:"upspeed,omitempty"`                       // Torrent upload speed (bytes/s)
}

// FilterType are the filter types.
type FilterType string

// Filter values.
const (
	FilterAll         FilterType = "all"
	FilterDownloading FilterType = "downloading"
	FilterCompleted   FilterType = "completed"
	FilterPaused      FilterType = "paused"
	FilterActive      FilterType = "active"
	FilterInactive    FilterType = "inactive"
	FilterResumed     FilterType = "resumed"
)

/*
// AuthLoginRequest is a login request
type AuthLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthLogin creates a login request.
func AuthLogin(username, password string) *AuthLoginRequest {
	return &LoginRequest{username: username, password: password}
}

// Do executes the request against the provided context and client.
func (req *AuthLoginRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "auth/login", req, nil)
}
*/

// AuthLogoutRequest is a logout request.
type AuthLogoutRequest struct{}

// AuthLogout creates a logout request.
func AuthLogout() *AuthLogoutRequest {
	return &AuthLogoutRequest{}
}

// Do executes the request against the provided context and client.
func (req *AuthLogoutRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "auth/logout", req, nil)
}

// AppVersionRequest is a version request.
type AppVersionRequest struct{}

// AppVersion creates a app verison request.
func AppVersion() *AppVersionRequest {
	return &AppVersionRequest{}
}

// Do executes the request against the provided context and client.
func (req *AppVersionRequest) Do(ctx context.Context, cl *Client) (string, error) {
	var res string
	if err := cl.Do(ctx, "app/version", req, &res); err != nil {
		return "", err
	}
	return res, nil
}

// AppWebapiVersionRequest is a app webapiVersion request.
type AppWebapiVersionRequest struct{}

// AppWebapiVersion creates a app webapiVersion request.
func AppWebapiVersion() *AppWebapiVersionRequest {
	return &AppWebapiVersionRequest{}
}

// Do executes the request against the provided context and client.
func (req *AppWebapiVersionRequest) Do(ctx context.Context, cl *Client) (string, error) {
	var res string
	if err := cl.Do(ctx, "app/webapiVersion", req, &res); err != nil {
		return "", err
	}
	return res, nil
}

// AppBuildInfoRequest is a app buildInfo request.
type AppBuildInfoRequest struct{}

// AppBuildInfo creates a app buildInfo request.
func AppBuildInfo() *AppBuildInfoRequest {
	return &AppBuildInfoRequest{}
}

// Do executes the request against the provided context and client.
func (req *AppBuildInfoRequest) Do(ctx context.Context, cl *Client) (*AppBuildInfoResponse, error) {
	res := new(AppBuildInfoResponse)
	if err := cl.Do(ctx, "app/buildInfo", req, res); err != nil {
		return nil, err
	}
	return res, nil
}

// AppBuildInfoResponse is the app buildInfo response.
type AppBuildInfoResponse struct {
	Qt         string `json:"qt,omitempty" yaml:"qt,omitempty"`                 // QT version
	Libtorrent string `json:"libtorrent,omitempty" yaml:"libtorrent,omitempty"` // libtorrent version
	Boost      string `json:"boost,omitempty" yaml:"boost,omitempty"`           // Boost version
	Openssl    string `json:"openssl,omitempty" yaml:"openssl,omitempty"`       // OpenSSL version
	Bitness    int64  `json:"bitness,omitempty" yaml:"bitness,omitempty"`       // Application bitness (e.g. 64-bit)
	Zlib       string `json:"zlib,omitempty" yaml:"zlib,omitempty"`             // Zlib version
}

// AppShutdownRequest is a app shutdown request.
type AppShutdownRequest struct{}

// AppShutdown creates a app shutdown request.
func AppShutdown() *AppShutdownRequest {
	return &AppShutdownRequest{}
}

// Do executes the request against the provided context and client.
func (req *AppShutdownRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "app/shutdown", req, nil)
}

// AppPreferencesRequest is a app preferences request.
type AppPreferencesRequest struct{}

// AppPreferences creates a app preferences request.
func AppPreferences() *AppPreferencesRequest {
	return &AppPreferencesRequest{}
}

// Do executes the request against the provided context and client.
func (req *AppPreferencesRequest) Do(ctx context.Context, cl *Client) (*AppPreferencesResponse, error) {
	res := new(AppPreferencesResponse)
	if err := cl.Do(ctx, "app/preferences", req, res); err != nil {
		return nil, err
	}
	return res, nil
}

// AppPreferencesResponse is the app preferences response.
type AppPreferencesResponse = Preferences

// Preferences is the app preferences.
type Preferences struct {
	AddTrackers                        string                 `json:"add_trackers" yaml:"add_trackers"`                                                     // not documented
	AddTrackersEnabled                 bool                   `json:"add_trackers_enabled" yaml:"add_trackers_enabled"`                                     // not documented
	AltDlLimit                         KiLimit                `json:"alt_dl_limit" yaml:"alt_dl_limit"`                                                     // Alternative global download speed limit in KiB/s
	AltUpLimit                         KiLimit                `json:"alt_up_limit" yaml:"alt_up_limit"`                                                     // Alternative global upload speed limit in KiB/s
	AlternativeWebuiEnabled            bool                   `json:"alternative_webui_enabled" yaml:"alternative_webui_enabled"`                           // True if an alternative WebUI should be used
	AlternativeWebuiPath               string                 `json:"alternative_webui_path" yaml:"alternative_webui_path"`                                 // File path to the alternative WebUI
	AnnounceIp                         string                 `json:"announce_ip" yaml:"announce_ip"`                                                       // not documented
	AnnounceToAllTiers                 bool                   `json:"announce_to_all_tiers" yaml:"announce_to_all_tiers"`                                   // not documented
	AnnounceToAllTrackers              bool                   `json:"announce_to_all_trackers" yaml:"announce_to_all_trackers"`                             // not documented
	AnonymousMode                      bool                   `json:"anonymous_mode" yaml:"anonymous_mode"`                                                 // If true anonymous mode will be enabled; read more here; this option is only available in qBittorent built against libtorrent version 0.16.X and higher
	AsyncIoThreads                     int64                  `json:"async_io_threads" yaml:"async_io_threads"`                                             // not documented
	AutoDeleteMode                     int64                  `json:"auto_delete_mode" yaml:"auto_delete_mode"`                                             // TODO
	AutoTmmEnabled                     bool                   `json:"auto_tmm_enabled" yaml:"auto_tmm_enabled"`                                             // True if Automatic Torrent Management is enabled by default
	AutorunEnabled                     bool                   `json:"autorun_enabled" yaml:"autorun_enabled"`                                               // True if external program should be run after torrent has finished downloading
	AutorunProgram                     string                 `json:"autorun_program" yaml:"autorun_program"`                                               // Program path/name/arguments to run if autorun_enabled is enabled; path is separated by slashes; you can use %f and %n arguments, which will be expanded by qBittorent as path_to_torrent_file and torrent_name (from the GUI; not the .torrent file name) respectively
	BannedIPs                          string                 `json:"banned_IPs" yaml:"banned_IPs"`                                                         // not documented
	BittorrentProtocol                 Protocol               `json:"bittorrent_protocol" yaml:"bittorrent_protocol"`                                       // not documented
	BypassAuthSubnetWhitelist          string                 `json:"bypass_auth_subnet_whitelist" yaml:"bypass_auth_subnet_whitelist"`                     // (White)list of ipv4/ipv6 subnets for which webui authentication should be bypassed; list entries are separated by commas
	BypassAuthSubnetWhitelistEnabled   bool                   `json:"bypass_auth_subnet_whitelist_enabled" yaml:"bypass_auth_subnet_whitelist_enabled"`     // True if webui authentication should be bypassed for clients whose ip resides within (at least) one of the subnets on the whitelist
	BypassLocalAuth                    bool                   `json:"bypass_local_auth" yaml:"bypass_local_auth"`                                           // True if authentication challenge for loopback address (127.0.0.1) should be disabled
	CategoryChangedTmmEnabled          bool                   `json:"category_changed_tmm_enabled" yaml:"category_changed_tmm_enabled"`                     // True if torrent should be relocated when its Category's save path changes
	CheckingMemoryUse                  int64                  `json:"checking_memory_use" yaml:"checking_memory_use"`                                       // not documented
	CreateSubfolderEnabled             bool                   `json:"create_subfolder_enabled" yaml:"create_subfolder_enabled"`                             // True if a subfolder should be created when adding a torrent
	CurrentInterfaceAddress            string                 `json:"current_interface_address" yaml:"current_interface_address"`                           // not documented
	Dht                                bool                   `json:"dht" yaml:"dht"`                                                                       // True if DHT is enabled
	DlLimit                            KiLimit                `json:"dl_limit" yaml:"dl_limit"`                                                             // Global download speed limit in KiB/s; -1 means no limit is applied
	DontCountSlowTorrents              bool                   `json:"dont_count_slow_torrents" yaml:"dont_count_slow_torrents"`                             // If true torrents w/o any activity (stalled ones) will not be counted towards max_active_* limits; see dont_count_slow_torrents for more information
	DyndnsDomain                       string                 `json:"dyndns_domain" yaml:"dyndns_domain"`                                                   // Your DDNS domain name
	DyndnsEnabled                      bool                   `json:"dyndns_enabled" yaml:"dyndns_enabled"`                                                 // True if server DNS should be updated dynamically
	DyndnsPassword                     string                 `json:"dyndns_password" yaml:"dyndns_password"`                                               // Password for DDNS service
	DyndnsService                      ServiceType            `json:"dyndns_service" yaml:"dyndns_service"`                                                 // See list of possible values here below
	DyndnsUsername                     string                 `json:"dyndns_username" yaml:"dyndns_username"`                                               // Username for DDNS service
	Encryption                         Encryption             `json:"encryption" yaml:"encryption"`                                                         // See list of possible values here below
	ExportDir                          string                 `json:"export_dir" yaml:"export_dir"`                                                         // Path to directory to copy .torrent files to. Slashes are used as path separators
	ExportDirFin                       string                 `json:"export_dir_fin" yaml:"export_dir_fin"`                                                 // Path to directory to copy .torrent files of completed downloads to. Slashes are used as path separators
	IncompleteFilesExt                 bool                   `json:"incomplete_files_ext" yaml:"incomplete_files_ext"`                                     // True if ".!qB" should be appended to incomplete files
	IpFilterEnabled                    bool                   `json:"ip_filter_enabled" yaml:"ip_filter_enabled"`                                           // True if external IP filter should be enabled
	IpFilterPath                       string                 `json:"ip_filter_path" yaml:"ip_filter_path"`                                                 // Path to IP filter file (.dat, .p2p, .p2b files are supported); path is separated by slashes
	IpFilterTrackers                   bool                   `json:"ip_filter_trackers" yaml:"ip_filter_trackers"`                                         // True if IP filters are applied to trackers
	LimitLanPeers                      bool                   `json:"limit_lan_peers" yaml:"limit_lan_peers"`                                               // True if [du]l_limit should be applied to peers on the LAN
	LimitTcpOverhead                   bool                   `json:"limit_tcp_overhead" yaml:"limit_tcp_overhead"`                                         // True if [du]l_limit should be applied to estimated TCP overhead (service data: e.g. packet headers)
	LimitUtpRate                       bool                   `json:"limit_utp_rate" yaml:"limit_utp_rate"`                                                 // True if [du]l_limit should be applied to uTP connections; this option is only available in qBittorent built against libtorrent version 0.16.X and higher
	ListenPort                         int64                  `json:"listen_port" yaml:"listen_port"`                                                       // Port for incoming connections
	Locale                             string                 `json:"locale" yaml:"locale"`                                                                 // Currently selected language (e.g. en_GB for English)
	Lsd                                bool                   `json:"lsd" yaml:"lsd"`                                                                       // True if LSD is enabled
	MailNotificationAuthEnabled        bool                   `json:"mail_notification_auth_enabled" yaml:"mail_notification_auth_enabled"`                 // True if smtp server requires authentication
	MailNotificationEmail              string                 `json:"mail_notification_email" yaml:"mail_notification_email"`                               // e-mail to send notifications to
	MailNotificationEnabled            bool                   `json:"mail_notification_enabled" yaml:"mail_notification_enabled"`                           // True if e-mail notification should be enabled
	MailNotificationPassword           string                 `json:"mail_notification_password" yaml:"mail_notification_password"`                         // Password for smtp authentication
	MailNotificationSender             string                 `json:"mail_notification_sender" yaml:"mail_notification_sender"`                             // e-mail where notifications should originate from
	MailNotificationSmtp               string                 `json:"mail_notification_smtp" yaml:"mail_notification_smtp"`                                 // smtp server for e-mail notifications
	MailNotificationSslEnabled         bool                   `json:"mail_notification_ssl_enabled" yaml:"mail_notification_ssl_enabled"`                   // True if smtp server requires SSL connection
	MailNotificationUsername           string                 `json:"mail_notification_username" yaml:"mail_notification_username"`                         // Username for smtp authentication
	MaxActiveDownloads                 int64                  `json:"max_active_downloads" yaml:"max_active_downloads"`                                     // Maximum number of active simultaneous downloads
	MaxActiveTorrents                  int64                  `json:"max_active_torrents" yaml:"max_active_torrents"`                                       // Maximum number of active simultaneous downloads and uploads
	MaxActiveUploads                   int64                  `json:"max_active_uploads" yaml:"max_active_uploads"`                                         // Maximum number of active simultaneous uploads
	MaxConnec                          int64                  `json:"max_connec" yaml:"max_connec"`                                                         // Maximum global number of simultaneous connections
	MaxConnecPerTorrent                int64                  `json:"max_connec_per_torrent" yaml:"max_connec_per_torrent"`                                 // Maximum number of simultaneous connections per torrent
	MaxRatio                           Percent                `json:"max_ratio" yaml:"max_ratio"`                                                           // Get the global share ratio limit
	MaxRatioAct                        BehaviorType           `json:"max_ratio_act" yaml:"max_ratio_act"`                                                   // Action performed when a torrent reaches the maximum share ratio. See list of possible values here below.
	MaxRatioEnabled                    bool                   `json:"max_ratio_enabled" yaml:"max_ratio_enabled"`                                           // True if share ratio limit is enabled
	MaxUploads                         int64                  `json:"max_uploads" yaml:"max_uploads"`                                                       // Maximum number of upload slots
	MaxUploadsPerTorrent               int64                  `json:"max_uploads_per_torrent" yaml:"max_uploads_per_torrent"`                               // Maximum number of upload slots per torrent
	Pex                                bool                   `json:"pex" yaml:"pex"`                                                                       // True if PeX is enabled
	PreallocateAll                     bool                   `json:"preallocate_all" yaml:"preallocate_all"`                                               // True if disk space should be pre-allocated for all files
	ProxyAuthEnabled                   bool                   `json:"proxy_auth_enabled" yaml:"proxy_auth_enabled"`                                         // True proxy requires authentication; doesn't apply to SOCKS4 proxies
	ProxyIp                            string                 `json:"proxy_ip" yaml:"proxy_ip"`                                                             // Proxy IP address or domain name
	ProxyPassword                      string                 `json:"proxy_password" yaml:"proxy_password"`                                                 // Password for proxy authentication
	ProxyPeerConnections               bool                   `json:"proxy_peer_connections" yaml:"proxy_peer_connections"`                                 // True if peer and web seed connections should be proxified; this option will have any effect only in qBittorent built against libtorrent version 0.16.X and higher
	ProxyPort                          int64                  `json:"proxy_port" yaml:"proxy_port"`                                                         // Proxy port
	ProxyType                          ProxyType              `json:"proxy_type" yaml:"proxy_type"`                                                         // See list of possible values here below
	ProxyUsername                      string                 `json:"proxy_username" yaml:"proxy_username"`                                                 // Username for proxy authentication
	QueueingEnabled                    bool                   `json:"queueing_enabled" yaml:"queueing_enabled"`                                             // True if torrent queuing is enabled
	RandomPort                         bool                   `json:"random_port" yaml:"random_port"`                                                       // True if the port is randomly selected
	RssAutoDownloadingEnabled          bool                   `json:"rss_auto_downloading_enabled" yaml:"rss_auto_downloading_enabled"`                     // Enable auto-downloading of torrents from the RSS feeds
	RssMaxArticlesPerFeed              int64                  `json:"rss_max_articles_per_feed" yaml:"rss_max_articles_per_feed"`                           // Max stored articles per RSS feed
	RssProcessingEnabled               bool                   `json:"rss_processing_enabled" yaml:"rss_processing_enabled"`                                 // Enable processing of RSS feeds
	RssRefreshInterval                 int64                  `json:"rss_refresh_interval" yaml:"rss_refresh_interval"`                                     // RSS refresh interval
	SavePath                           string                 `json:"save_path" yaml:"save_path"`                                                           // Default save path for torrents, separated by slashes
	SavePathChangedTmmEnabled          bool                   `json:"save_path_changed_tmm_enabled" yaml:"save_path_changed_tmm_enabled"`                   // True if torrent should be relocated when the default save path changes
	ScanDirs                           map[string]interface{} `json:"scan_dirs" yaml:"scan_dirs"`                                                           // Property: directory to watch for torrent files, value: where torrents loaded from this directory should be downloaded to (see list of possible values below). Slashes are used as path separators; multiple key/value pairs can be specified
	ScheduleFromHour                   int64                  `json:"schedule_from_hour" yaml:"schedule_from_hour"`                                         // Scheduler starting hour
	ScheduleFromMin                    int64                  `json:"schedule_from_min" yaml:"schedule_from_min"`                                           // Scheduler starting minute
	ScheduleToHour                     int64                  `json:"schedule_to_hour" yaml:"schedule_to_hour"`                                             // Scheduler ending hour
	ScheduleToMin                      int64                  `json:"schedule_to_min" yaml:"schedule_to_min"`                                               // Scheduler ending minute
	SchedulerDays                      DaySchedule            `json:"scheduler_days" yaml:"scheduler_days"`                                                 // Scheduler days. See possible values here below
	SchedulerEnabled                   bool                   `json:"scheduler_enabled" yaml:"scheduler_enabled"`                                           // True if alternative limits should be applied according to schedule
	SlowTorrentDlRateThreshold         KiLimit                `json:"slow_torrent_dl_rate_threshold" yaml:"slow_torrent_dl_rate_threshold"`                 // Download rate in KiB/s for a torrent to be considered "slow"
	SlowTorrentInactiveTimer           Duration               `json:"slow_torrent_inactive_timer" yaml:"slow_torrent_inactive_timer"`                       // Seconds a torrent should be inactive before considered "slow"
	SlowTorrentUlRateThreshold         KiLimit                `json:"slow_torrent_ul_rate_threshold" yaml:"slow_torrent_ul_rate_threshold"`                 // Upload rate in KiB/s for a torrent to be considered "slow"
	StartPausedEnabled                 bool                   `json:"start_paused_enabled" yaml:"start_paused_enabled"`                                     // True if torrents should be added in a Paused state
	TempPath                           string                 `json:"temp_path" yaml:"temp_path"`                                                           // Path for incomplete torrents, separated by slashes
	TempPathEnabled                    bool                   `json:"temp_path_enabled" yaml:"temp_path_enabled"`                                           // True if folder for incomplete torrents is enabled
	TorrentChangedTmmEnabled           bool                   `json:"torrent_changed_tmm_enabled" yaml:"torrent_changed_tmm_enabled"`                       // True if torrent should be relocated when its Category changes
	UpLimit                            KiLimit                `json:"up_limit" yaml:"up_limit"`                                                             // Global upload speed limit in KiB/s; -1 means no limit is applied
	Upnp                               bool                   `json:"upnp" yaml:"upnp"`                                                                     // True if UPnP/NAT-PMP is enabled
	UseHttps                           bool                   `json:"use_https" yaml:"use_https"`                                                           // True if WebUI HTTPS access is enabled
	WebUiAddress                       string                 `json:"web_ui_address" yaml:"web_ui_address"`                                                 // IP address to use for the WebUI
	WebUiClickjackingProtectionEnabled bool                   `json:"web_ui_clickjacking_protection_enabled" yaml:"web_ui_clickjacking_protection_enabled"` // True if WebUI clickjacking protection is enabled
	WebUiCsrfProtectionEnabled         bool                   `json:"web_ui_csrf_protection_enabled" yaml:"web_ui_csrf_protection_enabled"`                 // True if WebUI CSRF protection is enabled
	WebUiDomainList                    string                 `json:"web_ui_domain_list" yaml:"web_ui_domain_list"`                                         // Comma-separated list of domains to accept when performing Host header validation
	WebUiPort                          int64                  `json:"web_ui_port" yaml:"web_ui_port"`                                                       // WebUI port
	WebUiUpnp                          bool                   `json:"web_ui_upnp" yaml:"web_ui_upnp"`                                                       // True if UPnP is used for the WebUI port
	WebUiUsername                      string                 `json:"web_ui_username" yaml:"web_ui_username"`                                               // WebUI username
	CurrentNetworkInterface            string                 `json:"current_network_interface" yaml:"current_network_interface"`                           // not completed
	DiskCache                          int64                  `json:"disk_cache" yaml:"disk_cache"`                                                         // not completed
	DiskCacheTtl                       Duration               `json:"disk_cache_ttl" yaml:"disk_cache_ttl"`                                                 // not completed
	EmbeddedTrackerPort                int64                  `json:"embedded_tracker_port" yaml:"embedded_tracker_port"`                                   // not completed
	EnableCoalesceReadWrite            bool                   `json:"enable_coalesce_read_write" yaml:"enable_coalesce_read_write"`                         // not completed
	EnableEmbeddedTracker              bool                   `json:"enable_embedded_tracker" yaml:"enable_embedded_tracker"`                               // not completed
	EnableMultiConnectionsFromSameIp   bool                   `json:"enable_multi_connections_from_same_ip" yaml:"enable_multi_connections_from_same_ip"`   // not completed
	EnableOsCache                      bool                   `json:"enable_os_cache" yaml:"enable_os_cache"`                                               // not completed
	EnableSuperSeeding                 bool                   `json:"enable_super_seeding" yaml:"enable_super_seeding"`                                     // not completed
	EnableUploadSuggestions            bool                   `json:"enable_upload_suggestions" yaml:"enable_upload_suggestions"`                           // not completed
	FilePoolSize                       int64                  `json:"file_pool_size" yaml:"file_pool_size"`                                                 // not completed
	MaxSeedingTime                     Duration               `json:"max_seeding_time" yaml:"max_seeding_time"`                                             // not completed
	MaxSeedingTimeEnabled              bool                   `json:"max_seeding_time_enabled" yaml:"max_seeding_time_enabled"`                             // not completed
	OutgoingPortsMax                   int64                  `json:"outgoing_ports_max" yaml:"outgoing_ports_max"`                                         // not completed
	OutgoingPortsMin                   int64                  `json:"outgoing_ports_min" yaml:"outgoing_ports_min"`                                         // not completed
	ProxyTorrentsOnly                  bool                   `json:"proxy_torrents_only" yaml:"proxy_torrents_only"`                                       // not completed
	RecheckCompletedTorrents           bool                   `json:"recheck_completed_torrents" yaml:"recheck_completed_torrents"`                         // not completed
	ResolvePeerCountries               bool                   `json:"resolve_peer_countries" yaml:"resolve_peer_countries"`                                 // not completed
	SaveResumeDataInterval             Duration               `json:"save_resume_data_interval" yaml:"save_resume_data_interval"`                           // not completed
	SendBufferLowWatermark             int64                  `json:"send_buffer_low_watermark" yaml:"send_buffer_low_watermark"`                           // not completed
	SendBufferWatermark                int64                  `json:"send_buffer_watermark" yaml:"send_buffer_watermark"`                                   // not completed
	SendBufferWatermarkFactor          int64                  `json:"send_buffer_watermark_factor" yaml:"send_buffer_watermark_factor"`                     // not completed
	SocketBacklogSize                  int64                  `json:"socket_backlog_size" yaml:"socket_backlog_size"`                                       // not completed
	UploadChokingAlgorithm             int64                  `json:"upload_choking_algorithm" yaml:"upload_choking_algorithm"`                             // not completed
	UploadSlotsBehavior                int64                  `json:"upload_slots_behavior" yaml:"upload_slots_behavior"`                                   // not completed
	UtpTcpMixedMode                    int64                  `json:"utp_tcp_mixed_mode" yaml:"utp_tcp_mixed_mode"`                                         // not completed
	WebUiHostHeaderValidationEnabled   bool                   `json:"web_ui_host_header_validation_enabled" yaml:"web_ui_host_header_validation_enabled"`   // not completed
	WebUiHttpsCertPath                 string                 `json:"web_ui_https_cert_path" yaml:"web_ui_https_cert_path"`                                 // not completed
	WebUiHttpsKeyPath                  string                 `json:"web_ui_https_key_path" yaml:"web_ui_https_key_path"`                                   // not completed
	WebUiSessionTimeout                Duration               `json:"web_ui_session_timeout" yaml:"web_ui_session_timeout"`                                 // not completed

	// DhtPort                            int64                  `json:"dht_port" yaml:"dht_port"`                                                             // DHT port if dhtSameAsBT is false
	// DhtSameAsBT                        bool                   `json:"dhtSameAsBT" yaml:"dhtSameAsBT"`                                                       // True if DHT port should match TCP port
	// EnableUtp                          bool                   `json:"enable_utp" yaml:"enable_utp"`                                                         // True if uTP protocol should be enabled; this option is only available in qBittorent built against libtorrent version 0.16.X and higher
	// ForceProxy                         bool                   `json:"force_proxy" yaml:"force_proxy"`                                                       // True if the connections not supported by the proxy are disabled
	// SslCert                            string                 `json:"ssl_cert" yaml:"ssl_cert"`                                                             // SSL certificate contents (this is a not a path)
	// SslKey                             string                 `json:"ssl_key" yaml:"ssl_key"`                                                               // SSL keyfile contents (this is a not a path)
	// WebUiPassword                      string                 `json:"web_ui_password" yaml:"web_ui_password"`                                               // For API ≥ v2.3.0: Plaintext WebUI password, not readable, write-only. For API < v2.3.0: MD5 hash of WebUI password, hash is generated from the following string: username:Web UI Access:plain_text_web_ui_password
}

// AppSetPreferencesRequest is a app setPreferences request.
type AppSetPreferencesRequest struct {
	Preferences
	changed map[string]bool
}

// AppSetPreferences creates a app setPreferences request.
func AppSetPreferences() *AppSetPreferencesRequest {
	return &AppSetPreferencesRequest{
		changed: make(map[string]bool),
	}
}

// Do executes the request against the provided context and client.
func (req *AppSetPreferencesRequest) Do(ctx context.Context, cl *Client) error {
	if len(req.changed) == 0 {
		return nil
	}
	params := make(map[string]interface{})
	if req.changed["AddTrackers"] {
		params["add_trackers"] = req.AddTrackers
	}
	if req.changed["AddTrackersEnabled"] {
		params["add_trackers_enabled"] = req.AddTrackersEnabled
	}
	if req.changed["AltDlLimit"] {
		params["alt_dl_limit"] = req.AltDlLimit
	}
	if req.changed["AltUpLimit"] {
		params["alt_up_limit"] = req.AltUpLimit
	}
	if req.changed["AlternativeWebuiEnabled"] {
		params["alternative_webui_enabled"] = req.AlternativeWebuiEnabled
	}
	if req.changed["AlternativeWebuiPath"] {
		params["alternative_webui_path"] = req.AlternativeWebuiPath
	}
	if req.changed["AnnounceIp"] {
		params["announce_ip"] = req.AnnounceIp
	}
	if req.changed["AnnounceToAllTiers"] {
		params["announce_to_all_tiers"] = req.AnnounceToAllTiers
	}
	if req.changed["AnnounceToAllTrackers"] {
		params["announce_to_all_trackers"] = req.AnnounceToAllTrackers
	}
	if req.changed["AnonymousMode"] {
		params["anonymous_mode"] = req.AnonymousMode
	}
	if req.changed["AsyncIoThreads"] {
		params["async_io_threads"] = req.AsyncIoThreads
	}
	if req.changed["AutoDeleteMode"] {
		params["auto_delete_mode"] = req.AutoDeleteMode
	}
	if req.changed["AutoTmmEnabled"] {
		params["auto_tmm_enabled"] = req.AutoTmmEnabled
	}
	if req.changed["AutorunEnabled"] {
		params["autorun_enabled"] = req.AutorunEnabled
	}
	if req.changed["AutorunProgram"] {
		params["autorun_program"] = req.AutorunProgram
	}
	if req.changed["BannedIPs"] {
		params["banned_IPs"] = req.BannedIPs
	}
	if req.changed["BittorrentProtocol"] {
		params["bittorrent_protocol"] = req.BittorrentProtocol
	}
	if req.changed["BypassAuthSubnetWhitelist"] {
		params["bypass_auth_subnet_whitelist"] = req.BypassAuthSubnetWhitelist
	}
	if req.changed["BypassAuthSubnetWhitelistEnabled"] {
		params["bypass_auth_subnet_whitelist_enabled"] = req.BypassAuthSubnetWhitelistEnabled
	}
	if req.changed["BypassLocalAuth"] {
		params["bypass_local_auth"] = req.BypassLocalAuth
	}
	if req.changed["CategoryChangedTmmEnabled"] {
		params["category_changed_tmm_enabled"] = req.CategoryChangedTmmEnabled
	}
	if req.changed["CheckingMemoryUse"] {
		params["checking_memory_use"] = req.CheckingMemoryUse
	}
	if req.changed["CreateSubfolderEnabled"] {
		params["create_subfolder_enabled"] = req.CreateSubfolderEnabled
	}
	if req.changed["CurrentInterfaceAddress"] {
		params["current_interface_address"] = req.CurrentInterfaceAddress
	}
	if req.changed["Dht"] {
		params["dht"] = req.Dht
	}
	if req.changed["DlLimit"] {
		params["dl_limit"] = req.DlLimit
	}
	if req.changed["DontCountSlowTorrents"] {
		params["dont_count_slow_torrents"] = req.DontCountSlowTorrents
	}
	if req.changed["DyndnsDomain"] {
		params["dyndns_domain"] = req.DyndnsDomain
	}
	if req.changed["DyndnsEnabled"] {
		params["dyndns_enabled"] = req.DyndnsEnabled
	}
	if req.changed["DyndnsPassword"] {
		params["dyndns_password"] = req.DyndnsPassword
	}
	if req.changed["DyndnsService"] {
		params["dyndns_service"] = req.DyndnsService
	}
	if req.changed["DyndnsUsername"] {
		params["dyndns_username"] = req.DyndnsUsername
	}
	if req.changed["Encryption"] {
		params["encryption"] = req.Encryption
	}
	if req.changed["ExportDir"] {
		params["export_dir"] = req.ExportDir
	}
	if req.changed["ExportDirFin"] {
		params["export_dir_fin"] = req.ExportDirFin
	}
	if req.changed["IncompleteFilesExt"] {
		params["incomplete_files_ext"] = req.IncompleteFilesExt
	}
	if req.changed["IpFilterEnabled"] {
		params["ip_filter_enabled"] = req.IpFilterEnabled
	}
	if req.changed["IpFilterPath"] {
		params["ip_filter_path"] = req.IpFilterPath
	}
	if req.changed["IpFilterTrackers"] {
		params["ip_filter_trackers"] = req.IpFilterTrackers
	}
	if req.changed["LimitLanPeers"] {
		params["limit_lan_peers"] = req.LimitLanPeers
	}
	if req.changed["LimitTcpOverhead"] {
		params["limit_tcp_overhead"] = req.LimitTcpOverhead
	}
	if req.changed["LimitUtpRate"] {
		params["limit_utp_rate"] = req.LimitUtpRate
	}
	if req.changed["ListenPort"] {
		params["listen_port"] = req.ListenPort
	}
	if req.changed["Locale"] {
		params["locale"] = req.Locale
	}
	if req.changed["Lsd"] {
		params["lsd"] = req.Lsd
	}
	if req.changed["MailNotificationAuthEnabled"] {
		params["mail_notification_auth_enabled"] = req.MailNotificationAuthEnabled
	}
	if req.changed["MailNotificationEmail"] {
		params["mail_notification_email"] = req.MailNotificationEmail
	}
	if req.changed["MailNotificationEnabled"] {
		params["mail_notification_enabled"] = req.MailNotificationEnabled
	}
	if req.changed["MailNotificationPassword"] {
		params["mail_notification_password"] = req.MailNotificationPassword
	}
	if req.changed["MailNotificationSender"] {
		params["mail_notification_sender"] = req.MailNotificationSender
	}
	if req.changed["MailNotificationSmtp"] {
		params["mail_notification_smtp"] = req.MailNotificationSmtp
	}
	if req.changed["MailNotificationSslEnabled"] {
		params["mail_notification_ssl_enabled"] = req.MailNotificationSslEnabled
	}
	if req.changed["MailNotificationUsername"] {
		params["mail_notification_username"] = req.MailNotificationUsername
	}
	if req.changed["MaxActiveDownloads"] {
		params["max_active_downloads"] = req.MaxActiveDownloads
	}
	if req.changed["MaxActiveTorrents"] {
		params["max_active_torrents"] = req.MaxActiveTorrents
	}
	if req.changed["MaxActiveUploads"] {
		params["max_active_uploads"] = req.MaxActiveUploads
	}
	if req.changed["MaxConnec"] {
		params["max_connec"] = req.MaxConnec
	}
	if req.changed["MaxConnecPerTorrent"] {
		params["max_connec_per_torrent"] = req.MaxConnecPerTorrent
	}
	if req.changed["MaxRatio"] {
		params["max_ratio"] = req.MaxRatio
	}
	if req.changed["MaxRatioAct"] {
		params["max_ratio_act"] = req.MaxRatioAct
	}
	if req.changed["MaxRatioEnabled"] {
		params["max_ratio_enabled"] = req.MaxRatioEnabled
	}
	if req.changed["MaxUploads"] {
		params["max_uploads"] = req.MaxUploads
	}
	if req.changed["MaxUploadsPerTorrent"] {
		params["max_uploads_per_torrent"] = req.MaxUploadsPerTorrent
	}
	if req.changed["Pex"] {
		params["pex"] = req.Pex
	}
	if req.changed["PreallocateAll"] {
		params["preallocate_all"] = req.PreallocateAll
	}
	if req.changed["ProxyAuthEnabled"] {
		params["proxy_auth_enabled"] = req.ProxyAuthEnabled
	}
	if req.changed["ProxyIp"] {
		params["proxy_ip"] = req.ProxyIp
	}
	if req.changed["ProxyPassword"] {
		params["proxy_password"] = req.ProxyPassword
	}
	if req.changed["ProxyPeerConnections"] {
		params["proxy_peer_connections"] = req.ProxyPeerConnections
	}
	if req.changed["ProxyPort"] {
		params["proxy_port"] = req.ProxyPort
	}
	if req.changed["ProxyType"] {
		params["proxy_type"] = req.ProxyType
	}
	if req.changed["ProxyUsername"] {
		params["proxy_username"] = req.ProxyUsername
	}
	if req.changed["QueueingEnabled"] {
		params["queueing_enabled"] = req.QueueingEnabled
	}
	if req.changed["RandomPort"] {
		params["random_port"] = req.RandomPort
	}
	if req.changed["RssAutoDownloadingEnabled"] {
		params["rss_auto_downloading_enabled"] = req.RssAutoDownloadingEnabled
	}
	if req.changed["RssMaxArticlesPerFeed"] {
		params["rss_max_articles_per_feed"] = req.RssMaxArticlesPerFeed
	}
	if req.changed["RssProcessingEnabled"] {
		params["rss_processing_enabled"] = req.RssProcessingEnabled
	}
	if req.changed["RssRefreshInterval"] {
		params["rss_refresh_interval"] = req.RssRefreshInterval
	}
	if req.changed["SavePath"] {
		params["save_path"] = req.SavePath
	}
	if req.changed["SavePathChangedTmmEnabled"] {
		params["save_path_changed_tmm_enabled"] = req.SavePathChangedTmmEnabled
	}
	if req.changed["ScanDirs"] {
		params["scan_dirs"] = req.ScanDirs
	}
	if req.changed["ScheduleFromHour"] {
		params["schedule_from_hour"] = req.ScheduleFromHour
	}
	if req.changed["ScheduleFromMin"] {
		params["schedule_from_min"] = req.ScheduleFromMin
	}
	if req.changed["ScheduleToHour"] {
		params["schedule_to_hour"] = req.ScheduleToHour
	}
	if req.changed["ScheduleToMin"] {
		params["schedule_to_min"] = req.ScheduleToMin
	}
	if req.changed["SchedulerDays"] {
		params["scheduler_days"] = req.SchedulerDays
	}
	if req.changed["SchedulerEnabled"] {
		params["scheduler_enabled"] = req.SchedulerEnabled
	}
	if req.changed["SlowTorrentDlRateThreshold"] {
		params["slow_torrent_dl_rate_threshold"] = req.SlowTorrentDlRateThreshold
	}
	if req.changed["SlowTorrentInactiveTimer"] {
		params["slow_torrent_inactive_timer"] = req.SlowTorrentInactiveTimer
	}
	if req.changed["SlowTorrentUlRateThreshold"] {
		params["slow_torrent_ul_rate_threshold"] = req.SlowTorrentUlRateThreshold
	}
	if req.changed["StartPausedEnabled"] {
		params["start_paused_enabled"] = req.StartPausedEnabled
	}
	if req.changed["TempPath"] {
		params["temp_path"] = req.TempPath
	}
	if req.changed["TempPathEnabled"] {
		params["temp_path_enabled"] = req.TempPathEnabled
	}
	if req.changed["TorrentChangedTmmEnabled"] {
		params["torrent_changed_tmm_enabled"] = req.TorrentChangedTmmEnabled
	}
	if req.changed["UpLimit"] {
		params["up_limit"] = req.UpLimit
	}
	if req.changed["Upnp"] {
		params["upnp"] = req.Upnp
	}
	if req.changed["UseHttps"] {
		params["use_https"] = req.UseHttps
	}
	if req.changed["WebUiAddress"] {
		params["web_ui_address"] = req.WebUiAddress
	}
	if req.changed["WebUiClickjackingProtectionEnabled"] {
		params["web_ui_clickjacking_protection_enabled"] = req.WebUiClickjackingProtectionEnabled
	}
	if req.changed["WebUiCsrfProtectionEnabled"] {
		params["web_ui_csrf_protection_enabled"] = req.WebUiCsrfProtectionEnabled
	}
	if req.changed["WebUiDomainList"] {
		params["web_ui_domain_list"] = req.WebUiDomainList
	}
	if req.changed["WebUiPort"] {
		params["web_ui_port"] = req.WebUiPort
	}
	if req.changed["WebUiUpnp"] {
		params["web_ui_upnp"] = req.WebUiUpnp
	}
	if req.changed["WebUiUsername"] {
		params["web_ui_username"] = req.WebUiUsername
	}
	if req.changed["CurrentNetworkInterface"] {
		params["current_network_interface"] = req.CurrentNetworkInterface
	}
	if req.changed["DiskCache"] {
		params["disk_cache"] = req.DiskCache
	}
	if req.changed["DiskCacheTtl"] {
		params["disk_cache_ttl"] = req.DiskCacheTtl
	}
	if req.changed["EmbeddedTrackerPort"] {
		params["embedded_tracker_port"] = req.EmbeddedTrackerPort
	}
	if req.changed["EnableCoalesceReadWrite"] {
		params["enable_coalesce_read_write"] = req.EnableCoalesceReadWrite
	}
	if req.changed["EnableEmbeddedTracker"] {
		params["enable_embedded_tracker"] = req.EnableEmbeddedTracker
	}
	if req.changed["EnableMultiConnectionsFromSameIp"] {
		params["enable_multi_connections_from_same_ip"] = req.EnableMultiConnectionsFromSameIp
	}
	if req.changed["EnableOsCache"] {
		params["enable_os_cache"] = req.EnableOsCache
	}
	if req.changed["EnableSuperSeeding"] {
		params["enable_super_seeding"] = req.EnableSuperSeeding
	}
	if req.changed["EnableUploadSuggestions"] {
		params["enable_upload_suggestions"] = req.EnableUploadSuggestions
	}
	if req.changed["FilePoolSize"] {
		params["file_pool_size"] = req.FilePoolSize
	}
	if req.changed["MaxSeedingTime"] {
		params["max_seeding_time"] = req.MaxSeedingTime
	}
	if req.changed["MaxSeedingTimeEnabled"] {
		params["max_seeding_time_enabled"] = req.MaxSeedingTimeEnabled
	}
	if req.changed["OutgoingPortsMax"] {
		params["outgoing_ports_max"] = req.OutgoingPortsMax
	}
	if req.changed["OutgoingPortsMin"] {
		params["outgoing_ports_min"] = req.OutgoingPortsMin
	}
	if req.changed["ProxyTorrentsOnly"] {
		params["proxy_torrents_only"] = req.ProxyTorrentsOnly
	}
	if req.changed["RecheckCompletedTorrents"] {
		params["recheck_completed_torrents"] = req.RecheckCompletedTorrents
	}
	if req.changed["ResolvePeerCountries"] {
		params["resolve_peer_countries"] = req.ResolvePeerCountries
	}
	if req.changed["SaveResumeDataInterval"] {
		params["save_resume_data_interval"] = req.SaveResumeDataInterval
	}
	if req.changed["SendBufferLowWatermark"] {
		params["send_buffer_low_watermark"] = req.SendBufferLowWatermark
	}
	if req.changed["SendBufferWatermark"] {
		params["send_buffer_watermark"] = req.SendBufferWatermark
	}
	if req.changed["SendBufferWatermarkFactor"] {
		params["send_buffer_watermark_factor"] = req.SendBufferWatermarkFactor
	}
	if req.changed["SocketBacklogSize"] {
		params["socket_backlog_size"] = req.SocketBacklogSize
	}
	if req.changed["UploadChokingAlgorithm"] {
		params["upload_choking_algorithm"] = req.UploadChokingAlgorithm
	}
	if req.changed["UploadSlotsBehavior"] {
		params["upload_slots_behavior"] = req.UploadSlotsBehavior
	}
	if req.changed["UtpTcpMixedMode"] {
		params["utp_tcp_mixed_mode"] = req.UtpTcpMixedMode
	}
	if req.changed["WebUiHostHeaderValidationEnabled"] {
		params["web_ui_host_header_validation_enabled"] = req.WebUiHostHeaderValidationEnabled
	}
	if req.changed["WebUiHttpsCertPath"] {
		params["web_ui_https_cert_path"] = req.WebUiHttpsCertPath
	}
	if req.changed["WebUiHttpsKeyPath"] {
		params["web_ui_https_key_path"] = req.WebUiHttpsKeyPath
	}
	if req.changed["WebUiSessionTimeout"] {
		params["web_ui_session_timeout"] = req.WebUiSessionTimeout
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(params); err != nil {
		return err
	}
	return cl.Do(ctx, "app/setPreferences", map[string]interface{}{"json": buf.String()}, nil)
}

// WithChanged marks the fields that were changed.
func (req *AppSetPreferencesRequest) WithChanged(fields ...string) *AppSetPreferencesRequest {
	for _, field := range fields {
		req.changed[field] = true
	}
	return req
}

// WithAddTrackers sets add_trackers.
func (req AppSetPreferencesRequest) WithAddTrackers(addTrackers string) *AppSetPreferencesRequest {
	req.AddTrackers = addTrackers
	return req.WithChanged("AddTrackers")
}

// WithAddTrackersEnabled sets add_trackers_enabled.
func (req AppSetPreferencesRequest) WithAddTrackersEnabled(addTrackersEnabled bool) *AppSetPreferencesRequest {
	req.AddTrackersEnabled = addTrackersEnabled
	return req.WithChanged("AddTrackersEnabled")
}

// WithAltDlLimit sets alt_dl_limit.
func (req AppSetPreferencesRequest) WithAltDlLimit(altDlLimit KiLimit) *AppSetPreferencesRequest {
	req.AltDlLimit = altDlLimit
	return req.WithChanged("AltDlLimit")
}

// WithAltUpLimit sets alt_up_limit.
func (req AppSetPreferencesRequest) WithAltUpLimit(altUpLimit KiLimit) *AppSetPreferencesRequest {
	req.AltUpLimit = altUpLimit
	return req.WithChanged("AltUpLimit")
}

// WithAlternativeWebuiEnabled sets alternative_webui_enabled.
func (req AppSetPreferencesRequest) WithAlternativeWebuiEnabled(alternativeWebuiEnabled bool) *AppSetPreferencesRequest {
	req.AlternativeWebuiEnabled = alternativeWebuiEnabled
	return req.WithChanged("AlternativeWebuiEnabled")
}

// WithAlternativeWebuiPath sets alternative_webui_path.
func (req AppSetPreferencesRequest) WithAlternativeWebuiPath(alternativeWebuiPath string) *AppSetPreferencesRequest {
	req.AlternativeWebuiPath = alternativeWebuiPath
	return req.WithChanged("AlternativeWebuiPath")
}

// WithAnnounceIp sets announce_ip.
func (req AppSetPreferencesRequest) WithAnnounceIp(announceIp string) *AppSetPreferencesRequest {
	req.AnnounceIp = announceIp
	return req.WithChanged("AnnounceIp")
}

// WithAnnounceToAllTiers sets announce_to_all_tiers.
func (req AppSetPreferencesRequest) WithAnnounceToAllTiers(announceToAllTiers bool) *AppSetPreferencesRequest {
	req.AnnounceToAllTiers = announceToAllTiers
	return req.WithChanged("AnnounceToAllTiers")
}

// WithAnnounceToAllTrackers sets announce_to_all_trackers.
func (req AppSetPreferencesRequest) WithAnnounceToAllTrackers(announceToAllTrackers bool) *AppSetPreferencesRequest {
	req.AnnounceToAllTrackers = announceToAllTrackers
	return req.WithChanged("AnnounceToAllTrackers")
}

// WithAnonymousMode sets anonymous_mode.
func (req AppSetPreferencesRequest) WithAnonymousMode(anonymousMode bool) *AppSetPreferencesRequest {
	req.AnonymousMode = anonymousMode
	return req.WithChanged("AnonymousMode")
}

// WithAsyncIoThreads sets async_io_threads.
func (req AppSetPreferencesRequest) WithAsyncIoThreads(asyncIoThreads int64) *AppSetPreferencesRequest {
	req.AsyncIoThreads = asyncIoThreads
	return req.WithChanged("AsyncIoThreads")
}

// WithAutoDeleteMode sets auto_delete_mode.
func (req AppSetPreferencesRequest) WithAutoDeleteMode(autoDeleteMode int64) *AppSetPreferencesRequest {
	req.AutoDeleteMode = autoDeleteMode
	return req.WithChanged("AutoDeleteMode")
}

// WithAutoTmmEnabled sets auto_tmm_enabled.
func (req AppSetPreferencesRequest) WithAutoTmmEnabled(autoTmmEnabled bool) *AppSetPreferencesRequest {
	req.AutoTmmEnabled = autoTmmEnabled
	return req.WithChanged("AutoTmmEnabled")
}

// WithAutorunEnabled sets autorun_enabled.
func (req AppSetPreferencesRequest) WithAutorunEnabled(autorunEnabled bool) *AppSetPreferencesRequest {
	req.AutorunEnabled = autorunEnabled
	return req.WithChanged("AutorunEnabled")
}

// WithAutorunProgram sets autorun_program.
func (req AppSetPreferencesRequest) WithAutorunProgram(autorunProgram string) *AppSetPreferencesRequest {
	req.AutorunProgram = autorunProgram
	return req.WithChanged("AutorunProgram")
}

// WithBannedIPs sets banned_IPs.
func (req AppSetPreferencesRequest) WithBannedIPs(bannedIPs string) *AppSetPreferencesRequest {
	req.BannedIPs = bannedIPs
	return req.WithChanged("BannedIPs")
}

// WithBittorrentProtocol sets bittorrent_protocol.
func (req AppSetPreferencesRequest) WithBittorrentProtocol(bittorrentProtocol Protocol) *AppSetPreferencesRequest {
	req.BittorrentProtocol = bittorrentProtocol
	return req.WithChanged("BittorrentProtocol")
}

// WithBypassAuthSubnetWhitelist sets bypass_auth_subnet_whitelist.
func (req AppSetPreferencesRequest) WithBypassAuthSubnetWhitelist(bypassAuthSubnetWhitelist string) *AppSetPreferencesRequest {
	req.BypassAuthSubnetWhitelist = bypassAuthSubnetWhitelist
	return req.WithChanged("BypassAuthSubnetWhitelist")
}

// WithBypassAuthSubnetWhitelistEnabled sets bypass_auth_subnet_whitelist_enabled.
func (req AppSetPreferencesRequest) WithBypassAuthSubnetWhitelistEnabled(bypassAuthSubnetWhitelistEnabled bool) *AppSetPreferencesRequest {
	req.BypassAuthSubnetWhitelistEnabled = bypassAuthSubnetWhitelistEnabled
	return req.WithChanged("BypassAuthSubnetWhitelistEnabled")
}

// WithBypassLocalAuth sets bypass_local_auth.
func (req AppSetPreferencesRequest) WithBypassLocalAuth(bypassLocalAuth bool) *AppSetPreferencesRequest {
	req.BypassLocalAuth = bypassLocalAuth
	return req.WithChanged("BypassLocalAuth")
}

// WithCategoryChangedTmmEnabled sets category_changed_tmm_enabled.
func (req AppSetPreferencesRequest) WithCategoryChangedTmmEnabled(categoryChangedTmmEnabled bool) *AppSetPreferencesRequest {
	req.CategoryChangedTmmEnabled = categoryChangedTmmEnabled
	return req.WithChanged("CategoryChangedTmmEnabled")
}

// WithCheckingMemoryUse sets checking_memory_use.
func (req AppSetPreferencesRequest) WithCheckingMemoryUse(checkingMemoryUse int64) *AppSetPreferencesRequest {
	req.CheckingMemoryUse = checkingMemoryUse
	return req.WithChanged("CheckingMemoryUse")
}

// WithCreateSubfolderEnabled sets create_subfolder_enabled.
func (req AppSetPreferencesRequest) WithCreateSubfolderEnabled(createSubfolderEnabled bool) *AppSetPreferencesRequest {
	req.CreateSubfolderEnabled = createSubfolderEnabled
	return req.WithChanged("CreateSubfolderEnabled")
}

// WithCurrentInterfaceAddress sets current_interface_address.
func (req AppSetPreferencesRequest) WithCurrentInterfaceAddress(currentInterfaceAddress string) *AppSetPreferencesRequest {
	req.CurrentInterfaceAddress = currentInterfaceAddress
	return req.WithChanged("CurrentInterfaceAddress")
}

// WithDht sets dht.
func (req AppSetPreferencesRequest) WithDht(dht bool) *AppSetPreferencesRequest {
	req.Dht = dht
	return req.WithChanged("Dht")
}

// WithDlLimit sets dl_limit.
func (req AppSetPreferencesRequest) WithDlLimit(dlLimit KiLimit) *AppSetPreferencesRequest {
	req.DlLimit = dlLimit
	return req.WithChanged("DlLimit")
}

// WithDontCountSlowTorrents sets dont_count_slow_torrents.
func (req AppSetPreferencesRequest) WithDontCountSlowTorrents(dontCountSlowTorrents bool) *AppSetPreferencesRequest {
	req.DontCountSlowTorrents = dontCountSlowTorrents
	return req.WithChanged("DontCountSlowTorrents")
}

// WithDyndnsDomain sets dyndns_domain.
func (req AppSetPreferencesRequest) WithDyndnsDomain(dyndnsDomain string) *AppSetPreferencesRequest {
	req.DyndnsDomain = dyndnsDomain
	return req.WithChanged("DyndnsDomain")
}

// WithDyndnsEnabled sets dyndns_enabled.
func (req AppSetPreferencesRequest) WithDyndnsEnabled(dyndnsEnabled bool) *AppSetPreferencesRequest {
	req.DyndnsEnabled = dyndnsEnabled
	return req.WithChanged("DyndnsEnabled")
}

// WithDyndnsPassword sets dyndns_password.
func (req AppSetPreferencesRequest) WithDyndnsPassword(dyndnsPassword string) *AppSetPreferencesRequest {
	req.DyndnsPassword = dyndnsPassword
	return req.WithChanged("DyndnsPassword")
}

// WithDyndnsService sets dyndns_service.
func (req AppSetPreferencesRequest) WithDyndnsService(dyndnsService ServiceType) *AppSetPreferencesRequest {
	req.DyndnsService = dyndnsService
	return req.WithChanged("DyndnsService")
}

// WithDyndnsUsername sets dyndns_username.
func (req AppSetPreferencesRequest) WithDyndnsUsername(dyndnsUsername string) *AppSetPreferencesRequest {
	req.DyndnsUsername = dyndnsUsername
	return req.WithChanged("DyndnsUsername")
}

// WithEncryption sets encryption.
func (req AppSetPreferencesRequest) WithEncryption(encryption Encryption) *AppSetPreferencesRequest {
	req.Encryption = encryption
	return req.WithChanged("Encryption")
}

// WithExportDir sets export_dir.
func (req AppSetPreferencesRequest) WithExportDir(exportDir string) *AppSetPreferencesRequest {
	req.ExportDir = exportDir
	return req.WithChanged("ExportDir")
}

// WithExportDirFin sets export_dir_fin.
func (req AppSetPreferencesRequest) WithExportDirFin(exportDirFin string) *AppSetPreferencesRequest {
	req.ExportDirFin = exportDirFin
	return req.WithChanged("ExportDirFin")
}

// WithIncompleteFilesExt sets incomplete_files_ext.
func (req AppSetPreferencesRequest) WithIncompleteFilesExt(incompleteFilesExt bool) *AppSetPreferencesRequest {
	req.IncompleteFilesExt = incompleteFilesExt
	return req.WithChanged("IncompleteFilesExt")
}

// WithIpFilterEnabled sets ip_filter_enabled.
func (req AppSetPreferencesRequest) WithIpFilterEnabled(ipFilterEnabled bool) *AppSetPreferencesRequest {
	req.IpFilterEnabled = ipFilterEnabled
	return req.WithChanged("IpFilterEnabled")
}

// WithIpFilterPath sets ip_filter_path.
func (req AppSetPreferencesRequest) WithIpFilterPath(ipFilterPath string) *AppSetPreferencesRequest {
	req.IpFilterPath = ipFilterPath
	return req.WithChanged("IpFilterPath")
}

// WithIpFilterTrackers sets ip_filter_trackers.
func (req AppSetPreferencesRequest) WithIpFilterTrackers(ipFilterTrackers bool) *AppSetPreferencesRequest {
	req.IpFilterTrackers = ipFilterTrackers
	return req.WithChanged("IpFilterTrackers")
}

// WithLimitLanPeers sets limit_lan_peers.
func (req AppSetPreferencesRequest) WithLimitLanPeers(limitLanPeers bool) *AppSetPreferencesRequest {
	req.LimitLanPeers = limitLanPeers
	return req.WithChanged("LimitLanPeers")
}

// WithLimitTcpOverhead sets limit_tcp_overhead.
func (req AppSetPreferencesRequest) WithLimitTcpOverhead(limitTcpOverhead bool) *AppSetPreferencesRequest {
	req.LimitTcpOverhead = limitTcpOverhead
	return req.WithChanged("LimitTcpOverhead")
}

// WithLimitUtpRate sets limit_utp_rate.
func (req AppSetPreferencesRequest) WithLimitUtpRate(limitUtpRate bool) *AppSetPreferencesRequest {
	req.LimitUtpRate = limitUtpRate
	return req.WithChanged("LimitUtpRate")
}

// WithListenPort sets listen_port.
func (req AppSetPreferencesRequest) WithListenPort(listenPort int64) *AppSetPreferencesRequest {
	req.ListenPort = listenPort
	return req.WithChanged("ListenPort")
}

// WithLocale sets locale.
func (req AppSetPreferencesRequest) WithLocale(locale string) *AppSetPreferencesRequest {
	req.Locale = locale
	return req.WithChanged("Locale")
}

// WithLsd sets lsd.
func (req AppSetPreferencesRequest) WithLsd(lsd bool) *AppSetPreferencesRequest {
	req.Lsd = lsd
	return req.WithChanged("Lsd")
}

// WithMailNotificationAuthEnabled sets mail_notification_auth_enabled.
func (req AppSetPreferencesRequest) WithMailNotificationAuthEnabled(mailNotificationAuthEnabled bool) *AppSetPreferencesRequest {
	req.MailNotificationAuthEnabled = mailNotificationAuthEnabled
	return req.WithChanged("MailNotificationAuthEnabled")
}

// WithMailNotificationEmail sets mail_notification_email.
func (req AppSetPreferencesRequest) WithMailNotificationEmail(mailNotificationEmail string) *AppSetPreferencesRequest {
	req.MailNotificationEmail = mailNotificationEmail
	return req.WithChanged("MailNotificationEmail")
}

// WithMailNotificationEnabled sets mail_notification_enabled.
func (req AppSetPreferencesRequest) WithMailNotificationEnabled(mailNotificationEnabled bool) *AppSetPreferencesRequest {
	req.MailNotificationEnabled = mailNotificationEnabled
	return req.WithChanged("MailNotificationEnabled")
}

// WithMailNotificationPassword sets mail_notification_password.
func (req AppSetPreferencesRequest) WithMailNotificationPassword(mailNotificationPassword string) *AppSetPreferencesRequest {
	req.MailNotificationPassword = mailNotificationPassword
	return req.WithChanged("MailNotificationPassword")
}

// WithMailNotificationSender sets mail_notification_sender.
func (req AppSetPreferencesRequest) WithMailNotificationSender(mailNotificationSender string) *AppSetPreferencesRequest {
	req.MailNotificationSender = mailNotificationSender
	return req.WithChanged("MailNotificationSender")
}

// WithMailNotificationSmtp sets mail_notification_smtp.
func (req AppSetPreferencesRequest) WithMailNotificationSmtp(mailNotificationSmtp string) *AppSetPreferencesRequest {
	req.MailNotificationSmtp = mailNotificationSmtp
	return req.WithChanged("MailNotificationSmtp")
}

// WithMailNotificationSslEnabled sets mail_notification_ssl_enabled.
func (req AppSetPreferencesRequest) WithMailNotificationSslEnabled(mailNotificationSslEnabled bool) *AppSetPreferencesRequest {
	req.MailNotificationSslEnabled = mailNotificationSslEnabled
	return req.WithChanged("MailNotificationSslEnabled")
}

// WithMailNotificationUsername sets mail_notification_username.
func (req AppSetPreferencesRequest) WithMailNotificationUsername(mailNotificationUsername string) *AppSetPreferencesRequest {
	req.MailNotificationUsername = mailNotificationUsername
	return req.WithChanged("MailNotificationUsername")
}

// WithMaxActiveDownloads sets max_active_downloads.
func (req AppSetPreferencesRequest) WithMaxActiveDownloads(maxActiveDownloads int64) *AppSetPreferencesRequest {
	req.MaxActiveDownloads = maxActiveDownloads
	return req.WithChanged("MaxActiveDownloads")
}

// WithMaxActiveTorrents sets max_active_torrents.
func (req AppSetPreferencesRequest) WithMaxActiveTorrents(maxActiveTorrents int64) *AppSetPreferencesRequest {
	req.MaxActiveTorrents = maxActiveTorrents
	return req.WithChanged("MaxActiveTorrents")
}

// WithMaxActiveUploads sets max_active_uploads.
func (req AppSetPreferencesRequest) WithMaxActiveUploads(maxActiveUploads int64) *AppSetPreferencesRequest {
	req.MaxActiveUploads = maxActiveUploads
	return req.WithChanged("MaxActiveUploads")
}

// WithMaxConnec sets max_connec.
func (req AppSetPreferencesRequest) WithMaxConnec(maxConnec int64) *AppSetPreferencesRequest {
	req.MaxConnec = maxConnec
	return req.WithChanged("MaxConnec")
}

// WithMaxConnecPerTorrent sets max_connec_per_torrent.
func (req AppSetPreferencesRequest) WithMaxConnecPerTorrent(maxConnecPerTorrent int64) *AppSetPreferencesRequest {
	req.MaxConnecPerTorrent = maxConnecPerTorrent
	return req.WithChanged("MaxConnecPerTorrent")
}

// WithMaxRatio sets max_ratio.
func (req AppSetPreferencesRequest) WithMaxRatio(maxRatio Percent) *AppSetPreferencesRequest {
	req.MaxRatio = maxRatio
	return req.WithChanged("MaxRatio")
}

// WithMaxRatioAct sets max_ratio_act.
func (req AppSetPreferencesRequest) WithMaxRatioAct(maxRatioAct BehaviorType) *AppSetPreferencesRequest {
	req.MaxRatioAct = maxRatioAct
	return req.WithChanged("MaxRatioAct")
}

// WithMaxRatioEnabled sets max_ratio_enabled.
func (req AppSetPreferencesRequest) WithMaxRatioEnabled(maxRatioEnabled bool) *AppSetPreferencesRequest {
	req.MaxRatioEnabled = maxRatioEnabled
	return req.WithChanged("MaxRatioEnabled")
}

// WithMaxUploads sets max_uploads.
func (req AppSetPreferencesRequest) WithMaxUploads(maxUploads int64) *AppSetPreferencesRequest {
	req.MaxUploads = maxUploads
	return req.WithChanged("MaxUploads")
}

// WithMaxUploadsPerTorrent sets max_uploads_per_torrent.
func (req AppSetPreferencesRequest) WithMaxUploadsPerTorrent(maxUploadsPerTorrent int64) *AppSetPreferencesRequest {
	req.MaxUploadsPerTorrent = maxUploadsPerTorrent
	return req.WithChanged("MaxUploadsPerTorrent")
}

// WithPex sets pex.
func (req AppSetPreferencesRequest) WithPex(pex bool) *AppSetPreferencesRequest {
	req.Pex = pex
	return req.WithChanged("Pex")
}

// WithPreallocateAll sets preallocate_all.
func (req AppSetPreferencesRequest) WithPreallocateAll(preallocateAll bool) *AppSetPreferencesRequest {
	req.PreallocateAll = preallocateAll
	return req.WithChanged("PreallocateAll")
}

// WithProxyAuthEnabled sets proxy_auth_enabled.
func (req AppSetPreferencesRequest) WithProxyAuthEnabled(proxyAuthEnabled bool) *AppSetPreferencesRequest {
	req.ProxyAuthEnabled = proxyAuthEnabled
	return req.WithChanged("ProxyAuthEnabled")
}

// WithProxyIp sets proxy_ip.
func (req AppSetPreferencesRequest) WithProxyIp(proxyIp string) *AppSetPreferencesRequest {
	req.ProxyIp = proxyIp
	return req.WithChanged("ProxyIp")
}

// WithProxyPassword sets proxy_password.
func (req AppSetPreferencesRequest) WithProxyPassword(proxyPassword string) *AppSetPreferencesRequest {
	req.ProxyPassword = proxyPassword
	return req.WithChanged("ProxyPassword")
}

// WithProxyPeerConnections sets proxy_peer_connections.
func (req AppSetPreferencesRequest) WithProxyPeerConnections(proxyPeerConnections bool) *AppSetPreferencesRequest {
	req.ProxyPeerConnections = proxyPeerConnections
	return req.WithChanged("ProxyPeerConnections")
}

// WithProxyPort sets proxy_port.
func (req AppSetPreferencesRequest) WithProxyPort(proxyPort int64) *AppSetPreferencesRequest {
	req.ProxyPort = proxyPort
	return req.WithChanged("ProxyPort")
}

// WithProxyType sets proxy_type.
func (req AppSetPreferencesRequest) WithProxyType(proxyType ProxyType) *AppSetPreferencesRequest {
	req.ProxyType = proxyType
	return req.WithChanged("ProxyType")
}

// WithProxyUsername sets proxy_username.
func (req AppSetPreferencesRequest) WithProxyUsername(proxyUsername string) *AppSetPreferencesRequest {
	req.ProxyUsername = proxyUsername
	return req.WithChanged("ProxyUsername")
}

// WithQueueingEnabled sets queueing_enabled.
func (req AppSetPreferencesRequest) WithQueueingEnabled(queueingEnabled bool) *AppSetPreferencesRequest {
	req.QueueingEnabled = queueingEnabled
	return req.WithChanged("QueueingEnabled")
}

// WithRandomPort sets random_port.
func (req AppSetPreferencesRequest) WithRandomPort(randomPort bool) *AppSetPreferencesRequest {
	req.RandomPort = randomPort
	return req.WithChanged("RandomPort")
}

// WithRssAutoDownloadingEnabled sets rss_auto_downloading_enabled.
func (req AppSetPreferencesRequest) WithRssAutoDownloadingEnabled(rssAutoDownloadingEnabled bool) *AppSetPreferencesRequest {
	req.RssAutoDownloadingEnabled = rssAutoDownloadingEnabled
	return req.WithChanged("RssAutoDownloadingEnabled")
}

// WithRssMaxArticlesPerFeed sets rss_max_articles_per_feed.
func (req AppSetPreferencesRequest) WithRssMaxArticlesPerFeed(rssMaxArticlesPerFeed int64) *AppSetPreferencesRequest {
	req.RssMaxArticlesPerFeed = rssMaxArticlesPerFeed
	return req.WithChanged("RssMaxArticlesPerFeed")
}

// WithRssProcessingEnabled sets rss_processing_enabled.
func (req AppSetPreferencesRequest) WithRssProcessingEnabled(rssProcessingEnabled bool) *AppSetPreferencesRequest {
	req.RssProcessingEnabled = rssProcessingEnabled
	return req.WithChanged("RssProcessingEnabled")
}

// WithRssRefreshInterval sets rss_refresh_interval.
func (req AppSetPreferencesRequest) WithRssRefreshInterval(rssRefreshInterval int64) *AppSetPreferencesRequest {
	req.RssRefreshInterval = rssRefreshInterval
	return req.WithChanged("RssRefreshInterval")
}

// WithSavePath sets save_path.
func (req AppSetPreferencesRequest) WithSavePath(savePath string) *AppSetPreferencesRequest {
	req.SavePath = savePath
	return req.WithChanged("SavePath")
}

// WithSavePathChangedTmmEnabled sets save_path_changed_tmm_enabled.
func (req AppSetPreferencesRequest) WithSavePathChangedTmmEnabled(savePathChangedTmmEnabled bool) *AppSetPreferencesRequest {
	req.SavePathChangedTmmEnabled = savePathChangedTmmEnabled
	return req.WithChanged("SavePathChangedTmmEnabled")
}

// WithScanDirs sets scan_dirs.
func (req AppSetPreferencesRequest) WithScanDirs(scanDirs map[string]interface{}) *AppSetPreferencesRequest {
	req.ScanDirs = scanDirs
	return req.WithChanged("ScanDirs")
}

// WithScheduleFromHour sets schedule_from_hour.
func (req AppSetPreferencesRequest) WithScheduleFromHour(scheduleFromHour int64) *AppSetPreferencesRequest {
	req.ScheduleFromHour = scheduleFromHour
	return req.WithChanged("ScheduleFromHour")
}

// WithScheduleFromMin sets schedule_from_min.
func (req AppSetPreferencesRequest) WithScheduleFromMin(scheduleFromMin int64) *AppSetPreferencesRequest {
	req.ScheduleFromMin = scheduleFromMin
	return req.WithChanged("ScheduleFromMin")
}

// WithScheduleToHour sets schedule_to_hour.
func (req AppSetPreferencesRequest) WithScheduleToHour(scheduleToHour int64) *AppSetPreferencesRequest {
	req.ScheduleToHour = scheduleToHour
	return req.WithChanged("ScheduleToHour")
}

// WithScheduleToMin sets schedule_to_min.
func (req AppSetPreferencesRequest) WithScheduleToMin(scheduleToMin int64) *AppSetPreferencesRequest {
	req.ScheduleToMin = scheduleToMin
	return req.WithChanged("ScheduleToMin")
}

// WithSchedulerDays sets scheduler_days.
func (req AppSetPreferencesRequest) WithSchedulerDays(schedulerDays DaySchedule) *AppSetPreferencesRequest {
	req.SchedulerDays = schedulerDays
	return req.WithChanged("SchedulerDays")
}

// WithSchedulerEnabled sets scheduler_enabled.
func (req AppSetPreferencesRequest) WithSchedulerEnabled(schedulerEnabled bool) *AppSetPreferencesRequest {
	req.SchedulerEnabled = schedulerEnabled
	return req.WithChanged("SchedulerEnabled")
}

// WithSlowTorrentDlRateThreshold sets slow_torrent_dl_rate_threshold.
func (req AppSetPreferencesRequest) WithSlowTorrentDlRateThreshold(slowTorrentDlRateThreshold KiLimit) *AppSetPreferencesRequest {
	req.SlowTorrentDlRateThreshold = slowTorrentDlRateThreshold
	return req.WithChanged("SlowTorrentDlRateThreshold")
}

// WithSlowTorrentInactiveTimer sets slow_torrent_inactive_timer.
func (req AppSetPreferencesRequest) WithSlowTorrentInactiveTimer(slowTorrentInactiveTimer Duration) *AppSetPreferencesRequest {
	req.SlowTorrentInactiveTimer = slowTorrentInactiveTimer
	return req.WithChanged("SlowTorrentInactiveTimer")
}

// WithSlowTorrentUlRateThreshold sets slow_torrent_ul_rate_threshold.
func (req AppSetPreferencesRequest) WithSlowTorrentUlRateThreshold(slowTorrentUlRateThreshold KiLimit) *AppSetPreferencesRequest {
	req.SlowTorrentUlRateThreshold = slowTorrentUlRateThreshold
	return req.WithChanged("SlowTorrentUlRateThreshold")
}

// WithStartPausedEnabled sets start_paused_enabled.
func (req AppSetPreferencesRequest) WithStartPausedEnabled(startPausedEnabled bool) *AppSetPreferencesRequest {
	req.StartPausedEnabled = startPausedEnabled
	return req.WithChanged("StartPausedEnabled")
}

// WithTempPath sets temp_path.
func (req AppSetPreferencesRequest) WithTempPath(tempPath string) *AppSetPreferencesRequest {
	req.TempPath = tempPath
	return req.WithChanged("TempPath")
}

// WithTempPathEnabled sets temp_path_enabled.
func (req AppSetPreferencesRequest) WithTempPathEnabled(tempPathEnabled bool) *AppSetPreferencesRequest {
	req.TempPathEnabled = tempPathEnabled
	return req.WithChanged("TempPathEnabled")
}

// WithTorrentChangedTmmEnabled sets torrent_changed_tmm_enabled.
func (req AppSetPreferencesRequest) WithTorrentChangedTmmEnabled(torrentChangedTmmEnabled bool) *AppSetPreferencesRequest {
	req.TorrentChangedTmmEnabled = torrentChangedTmmEnabled
	return req.WithChanged("TorrentChangedTmmEnabled")
}

// WithUpLimit sets up_limit.
func (req AppSetPreferencesRequest) WithUpLimit(upLimit KiLimit) *AppSetPreferencesRequest {
	req.UpLimit = upLimit
	return req.WithChanged("UpLimit")
}

// WithUpnp sets upnp.
func (req AppSetPreferencesRequest) WithUpnp(upnp bool) *AppSetPreferencesRequest {
	req.Upnp = upnp
	return req.WithChanged("Upnp")
}

// WithUseHttps sets use_https.
func (req AppSetPreferencesRequest) WithUseHttps(useHttps bool) *AppSetPreferencesRequest {
	req.UseHttps = useHttps
	return req.WithChanged("UseHttps")
}

// WithWebUiAddress sets web_ui_address.
func (req AppSetPreferencesRequest) WithWebUiAddress(webUiAddress string) *AppSetPreferencesRequest {
	req.WebUiAddress = webUiAddress
	return req.WithChanged("WebUiAddress")
}

// WithWebUiClickjackingProtectionEnabled sets web_ui_clickjacking_protection_enabled.
func (req AppSetPreferencesRequest) WithWebUiClickjackingProtectionEnabled(webUiClickjackingProtectionEnabled bool) *AppSetPreferencesRequest {
	req.WebUiClickjackingProtectionEnabled = webUiClickjackingProtectionEnabled
	return req.WithChanged("WebUiClickjackingProtectionEnabled")
}

// WithWebUiCsrfProtectionEnabled sets web_ui_csrf_protection_enabled.
func (req AppSetPreferencesRequest) WithWebUiCsrfProtectionEnabled(webUiCsrfProtectionEnabled bool) *AppSetPreferencesRequest {
	req.WebUiCsrfProtectionEnabled = webUiCsrfProtectionEnabled
	return req.WithChanged("WebUiCsrfProtectionEnabled")
}

// WithWebUiDomainList sets web_ui_domain_list.
func (req AppSetPreferencesRequest) WithWebUiDomainList(webUiDomainList string) *AppSetPreferencesRequest {
	req.WebUiDomainList = webUiDomainList
	return req.WithChanged("WebUiDomainList")
}

// WithWebUiPort sets web_ui_port.
func (req AppSetPreferencesRequest) WithWebUiPort(webUiPort int64) *AppSetPreferencesRequest {
	req.WebUiPort = webUiPort
	return req.WithChanged("WebUiPort")
}

// WithWebUiUpnp sets web_ui_upnp.
func (req AppSetPreferencesRequest) WithWebUiUpnp(webUiUpnp bool) *AppSetPreferencesRequest {
	req.WebUiUpnp = webUiUpnp
	return req.WithChanged("WebUiUpnp")
}

// WithWebUiUsername sets web_ui_username.
func (req AppSetPreferencesRequest) WithWebUiUsername(webUiUsername string) *AppSetPreferencesRequest {
	req.WebUiUsername = webUiUsername
	return req.WithChanged("WebUiUsername")
}

// WithCurrentNetworkInterface sets current_network_interface.
func (req AppSetPreferencesRequest) WithCurrentNetworkInterface(currentNetworkInterface string) *AppSetPreferencesRequest {
	req.CurrentNetworkInterface = currentNetworkInterface
	return req.WithChanged("CurrentNetworkInterface")
}

// WithDiskCache sets disk_cache.
func (req AppSetPreferencesRequest) WithDiskCache(diskCache int64) *AppSetPreferencesRequest {
	req.DiskCache = diskCache
	return req.WithChanged("DiskCache")
}

// WithDiskCacheTtl sets disk_cache_ttl.
func (req AppSetPreferencesRequest) WithDiskCacheTtl(diskCacheTtl Duration) *AppSetPreferencesRequest {
	req.DiskCacheTtl = diskCacheTtl
	return req.WithChanged("DiskCacheTtl")
}

// WithEmbeddedTrackerPort sets embedded_tracker_port.
func (req AppSetPreferencesRequest) WithEmbeddedTrackerPort(embeddedTrackerPort int64) *AppSetPreferencesRequest {
	req.EmbeddedTrackerPort = embeddedTrackerPort
	return req.WithChanged("EmbeddedTrackerPort")
}

// WithEnableCoalesceReadWrite sets enable_coalesce_read_write.
func (req AppSetPreferencesRequest) WithEnableCoalesceReadWrite(enableCoalesceReadWrite bool) *AppSetPreferencesRequest {
	req.EnableCoalesceReadWrite = enableCoalesceReadWrite
	return req.WithChanged("EnableCoalesceReadWrite")
}

// WithEnableEmbeddedTracker sets enable_embedded_tracker.
func (req AppSetPreferencesRequest) WithEnableEmbeddedTracker(enableEmbeddedTracker bool) *AppSetPreferencesRequest {
	req.EnableEmbeddedTracker = enableEmbeddedTracker
	return req.WithChanged("EnableEmbeddedTracker")
}

// WithEnableMultiConnectionsFromSameIp sets enable_multi_connections_from_same_ip.
func (req AppSetPreferencesRequest) WithEnableMultiConnectionsFromSameIp(enableMultiConnectionsFromSameIp bool) *AppSetPreferencesRequest {
	req.EnableMultiConnectionsFromSameIp = enableMultiConnectionsFromSameIp
	return req.WithChanged("EnableMultiConnectionsFromSameIp")
}

// WithEnableOsCache sets enable_os_cache.
func (req AppSetPreferencesRequest) WithEnableOsCache(enableOsCache bool) *AppSetPreferencesRequest {
	req.EnableOsCache = enableOsCache
	return req.WithChanged("EnableOsCache")
}

// WithEnableSuperSeeding sets enable_super_seeding.
func (req AppSetPreferencesRequest) WithEnableSuperSeeding(enableSuperSeeding bool) *AppSetPreferencesRequest {
	req.EnableSuperSeeding = enableSuperSeeding
	return req.WithChanged("EnableSuperSeeding")
}

// WithEnableUploadSuggestions sets enable_upload_suggestions.
func (req AppSetPreferencesRequest) WithEnableUploadSuggestions(enableUploadSuggestions bool) *AppSetPreferencesRequest {
	req.EnableUploadSuggestions = enableUploadSuggestions
	return req.WithChanged("EnableUploadSuggestions")
}

// WithFilePoolSize sets file_pool_size.
func (req AppSetPreferencesRequest) WithFilePoolSize(filePoolSize int64) *AppSetPreferencesRequest {
	req.FilePoolSize = filePoolSize
	return req.WithChanged("FilePoolSize")
}

// WithMaxSeedingTime sets max_seeding_time.
func (req AppSetPreferencesRequest) WithMaxSeedingTime(maxSeedingTime Duration) *AppSetPreferencesRequest {
	req.MaxSeedingTime = maxSeedingTime
	return req.WithChanged("MaxSeedingTime")
}

// WithMaxSeedingTimeEnabled sets max_seeding_time_enabled.
func (req AppSetPreferencesRequest) WithMaxSeedingTimeEnabled(maxSeedingTimeEnabled bool) *AppSetPreferencesRequest {
	req.MaxSeedingTimeEnabled = maxSeedingTimeEnabled
	return req.WithChanged("MaxSeedingTimeEnabled")
}

// WithOutgoingPortsMax sets outgoing_ports_max.
func (req AppSetPreferencesRequest) WithOutgoingPortsMax(outgoingPortsMax int64) *AppSetPreferencesRequest {
	req.OutgoingPortsMax = outgoingPortsMax
	return req.WithChanged("OutgoingPortsMax")
}

// WithOutgoingPortsMin sets outgoing_ports_min.
func (req AppSetPreferencesRequest) WithOutgoingPortsMin(outgoingPortsMin int64) *AppSetPreferencesRequest {
	req.OutgoingPortsMin = outgoingPortsMin
	return req.WithChanged("OutgoingPortsMin")
}

// WithProxyTorrentsOnly sets proxy_torrents_only.
func (req AppSetPreferencesRequest) WithProxyTorrentsOnly(proxyTorrentsOnly bool) *AppSetPreferencesRequest {
	req.ProxyTorrentsOnly = proxyTorrentsOnly
	return req.WithChanged("ProxyTorrentsOnly")
}

// WithRecheckCompletedTorrents sets recheck_completed_torrents.
func (req AppSetPreferencesRequest) WithRecheckCompletedTorrents(recheckCompletedTorrents bool) *AppSetPreferencesRequest {
	req.RecheckCompletedTorrents = recheckCompletedTorrents
	return req.WithChanged("RecheckCompletedTorrents")
}

// WithResolvePeerCountries sets resolve_peer_countries.
func (req AppSetPreferencesRequest) WithResolvePeerCountries(resolvePeerCountries bool) *AppSetPreferencesRequest {
	req.ResolvePeerCountries = resolvePeerCountries
	return req.WithChanged("ResolvePeerCountries")
}

// WithSaveResumeDataInterval sets save_resume_data_interval.
func (req AppSetPreferencesRequest) WithSaveResumeDataInterval(saveResumeDataInterval Duration) *AppSetPreferencesRequest {
	req.SaveResumeDataInterval = saveResumeDataInterval
	return req.WithChanged("SaveResumeDataInterval")
}

// WithSendBufferLowWatermark sets send_buffer_low_watermark.
func (req AppSetPreferencesRequest) WithSendBufferLowWatermark(sendBufferLowWatermark int64) *AppSetPreferencesRequest {
	req.SendBufferLowWatermark = sendBufferLowWatermark
	return req.WithChanged("SendBufferLowWatermark")
}

// WithSendBufferWatermark sets send_buffer_watermark.
func (req AppSetPreferencesRequest) WithSendBufferWatermark(sendBufferWatermark int64) *AppSetPreferencesRequest {
	req.SendBufferWatermark = sendBufferWatermark
	return req.WithChanged("SendBufferWatermark")
}

// WithSendBufferWatermarkFactor sets send_buffer_watermark_factor.
func (req AppSetPreferencesRequest) WithSendBufferWatermarkFactor(sendBufferWatermarkFactor int64) *AppSetPreferencesRequest {
	req.SendBufferWatermarkFactor = sendBufferWatermarkFactor
	return req.WithChanged("SendBufferWatermarkFactor")
}

// WithSocketBacklogSize sets socket_backlog_size.
func (req AppSetPreferencesRequest) WithSocketBacklogSize(socketBacklogSize int64) *AppSetPreferencesRequest {
	req.SocketBacklogSize = socketBacklogSize
	return req.WithChanged("SocketBacklogSize")
}

// WithUploadChokingAlgorithm sets upload_choking_algorithm.
func (req AppSetPreferencesRequest) WithUploadChokingAlgorithm(uploadChokingAlgorithm int64) *AppSetPreferencesRequest {
	req.UploadChokingAlgorithm = uploadChokingAlgorithm
	return req.WithChanged("UploadChokingAlgorithm")
}

// WithUploadSlotsBehavior sets upload_slots_behavior.
func (req AppSetPreferencesRequest) WithUploadSlotsBehavior(uploadSlotsBehavior int64) *AppSetPreferencesRequest {
	req.UploadSlotsBehavior = uploadSlotsBehavior
	return req.WithChanged("UploadSlotsBehavior")
}

// WithUtpTcpMixedMode sets utp_tcp_mixed_mode.
func (req AppSetPreferencesRequest) WithUtpTcpMixedMode(utpTcpMixedMode int64) *AppSetPreferencesRequest {
	req.UtpTcpMixedMode = utpTcpMixedMode
	return req.WithChanged("UtpTcpMixedMode")
}

// WithWebUiHostHeaderValidationEnabled sets web_ui_host_header_validation_enabled.
func (req AppSetPreferencesRequest) WithWebUiHostHeaderValidationEnabled(webUiHostHeaderValidationEnabled bool) *AppSetPreferencesRequest {
	req.WebUiHostHeaderValidationEnabled = webUiHostHeaderValidationEnabled
	return req.WithChanged("WebUiHostHeaderValidationEnabled")
}

// WithWebUiHttpsCertPath sets web_ui_https_cert_path.
func (req AppSetPreferencesRequest) WithWebUiHttpsCertPath(webUiHttpsCertPath string) *AppSetPreferencesRequest {
	req.WebUiHttpsCertPath = webUiHttpsCertPath
	return req.WithChanged("WebUiHttpsCertPath")
}

// WithWebUiHttpsKeyPath sets web_ui_https_key_path.
func (req AppSetPreferencesRequest) WithWebUiHttpsKeyPath(webUiHttpsKeyPath string) *AppSetPreferencesRequest {
	req.WebUiHttpsKeyPath = webUiHttpsKeyPath
	return req.WithChanged("WebUiHttpsKeyPath")
}

// WithWebUiSessionTimeout sets web_ui_session_timeout.
func (req AppSetPreferencesRequest) WithWebUiSessionTimeout(webUiSessionTimeout Duration) *AppSetPreferencesRequest {
	req.WebUiSessionTimeout = webUiSessionTimeout
	return req.WithChanged("WebUiSessionTimeout")
}

// AppDefaultSavePathRequest is a app defaultSavePath request.
type AppDefaultSavePathRequest struct{}

// AppDefaultSavePath creates a app defaultSavePath request.
func AppDefaultSavePath() *AppDefaultSavePathRequest {
	return &AppDefaultSavePathRequest{}
}

// Do executes the request against the provided context and client.
func (req *AppDefaultSavePathRequest) Do(ctx context.Context, cl *Client) (string, error) {
	var res string
	if err := cl.Do(ctx, "app/defaultSavePath", req, &res); err != nil {
		return "", err
	}
	return res, nil
}

// LogMainRequest is a log main request.
type LogMainRequest struct {
	Normal      bool  `json:"normal" yaml:"normal"`               // Include normal messages (default: true)
	Info        bool  `json:"info" yaml:"info"`                   // Include info messages (default: true)
	Warning     bool  `json:"warning" yaml:"warning"`             // Include warning messages (default: true)
	Critical    bool  `json:"critical" yaml:"critical"`           // Include critical messages (default: true)
	LastKnownID int64 `json:"last_known_id" yaml:"last_known_id"` // Exclude messages with "message id" <= last_known_id (default: -1)
}

// LogMain creates a log main request.
func LogMain() *LogMainRequest {
	return &LogMainRequest{
		Normal:      true,
		Info:        true,
		Warning:     true,
		Critical:    true,
		LastKnownID: -1,
	}
}

// Do executes the request against the provided context and client.
func (req *LogMainRequest) Do(ctx context.Context, cl *Client) ([]MainLogEntry, error) {
	var res []MainLogEntry
	if err := cl.Do(ctx, "log/main", req, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// WithNormal sets include normal messages (default: true).
func (req LogMainRequest) WithNormal(normal bool) *LogMainRequest {
	req.Normal = normal
	return &req
}

// WithInfo sets include info messages (default: true).
func (req LogMainRequest) WithInfo(info bool) *LogMainRequest {
	req.Info = info
	return &req
}

// WithWarning sets include warning messages (default: true).
func (req LogMainRequest) WithWarning(warning bool) *LogMainRequest {
	req.Warning = warning
	return &req
}

// WithCritical sets include critical messages (default: true).
func (req LogMainRequest) WithCritical(critical bool) *LogMainRequest {
	req.Critical = critical
	return &req
}

// WithLastKnownID sets exclude messages with "message id" <= last_known_id (default: -1).
func (req LogMainRequest) WithLastKnownID(lastKnownID int64) *LogMainRequest {
	req.LastKnownID = lastKnownID
	return &req
}

// MainLogEntry is the a main log entry.
type MainLogEntry struct {
	ID        int64     `json:"id,omitempty" yaml:"id,omitempty"`               // ID of the message
	Message   string    `json:"message,omitempty" yaml:"message,omitempty"`     // Text of the message
	Timestamp MilliTime `json:"timestamp,omitempty" yaml:"timestamp,omitempty"` // Milliseconds since epoch
	Type      LogType   `json:"type,omitempty" yaml:"type,omitempty"`           // Type of the message: Log::NORMAL: 1, Log::INFO: 2, Log::WARNING: 4, Log::CRITICAL: 8
}

// LogPeersRequest is a log peers request.
type LogPeersRequest struct {
	LastKnownID int64 `json:"last_known_id" yaml:"last_known_id"` // Exclude messages with "message id" <= last_known_id (default: -1)
}

// LogPeers creates a log peers request.
func LogPeers() *LogPeersRequest {
	return &LogPeersRequest{
		LastKnownID: -1,
	}
}

// Do executes the request against the provided context and client.
func (req *LogPeersRequest) Do(ctx context.Context, cl *Client) ([]PeersLogEntry, error) {
	var res []PeersLogEntry
	if err := cl.Do(ctx, "log/peers", req, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// WithLastKnownID sets exclude messages with "message id" <= last_known_id (default: -1).
func (req LogPeersRequest) WithLastKnownID(lastKnownID int64) *LogPeersRequest {
	req.LastKnownID = lastKnownID
	return &req
}

// PeersLogEntry is a peers log entry.
type PeersLogEntry struct {
	ID        int64     `json:"id,omitempty" yaml:"id,omitempty"`               // ID of the peer
	Ip        string    `json:"ip,omitempty" yaml:"ip,omitempty"`               // IP of the peer
	Timestamp MilliTime `json:"timestamp,omitempty" yaml:"timestamp,omitempty"` // Milliseconds since epoch
	Blocked   bool      `json:"blocked,omitempty" yaml:"blocked,omitempty"`     // Whether or not the peer was blocked
	Reason    string    `json:"reason,omitempty" yaml:"reason,omitempty"`       // Reason of the block
}

// SyncMaindataRequest is a sync maindata request.
type SyncMaindataRequest struct {
	Rid int64 `json:"rid,omitempty" yaml:"rid,omitempty"` // Response ID. If not provided, rid=0 will be assumed. If the given rid is different from the one of last server reply, full_update will be true (see the server reply details for more info)
}

// SyncMaindata creates a sync maindata request.
func SyncMaindata() *SyncMaindataRequest {
	return &SyncMaindataRequest{}
}

// Do executes the request against the provided context and client.
func (req *SyncMaindataRequest) Do(ctx context.Context, cl *Client) (*SyncMaindataResponse, error) {
	res := new(SyncMaindataResponse)
	if err := cl.Do(ctx, "sync/maindata", req, res); err != nil {
		return nil, err
	}
	return res, nil
}

// WithResponseID sets response ID. If not provided, rid=0 will be assumed. If
// the given rid is different from the one of last server reply, full_update
// will be true (see the server reply details for more info).
func (req SyncMaindataRequest) WithResponseID(rid int64) *SyncMaindataRequest {
	req.Rid = rid
	return &req
}

// SyncMaindataResponse is the sync maindata response.
type SyncMaindataResponse struct {
	Rid               int64               `json:"rid,omitempty" yaml:"rid,omitempty"`                               // Response ID
	FullUpdate        bool                `json:"full_update,omitempty" yaml:"full_update,omitempty"`               // Whether the response contains all the data or partial data
	Torrents          []Torrent           `json:"torrents,omitempty" yaml:"torrents,omitempty"`                     // Property: torrent hash, value: same as torrent list
	TorrentsRemoved   []string            `json:"torrents_removed,omitempty" yaml:"torrents_removed,omitempty"`     // List of hashes of torrents removed since last request
	Categories        map[string]Category `json:"categories,omitempty" yaml:"categories,omitempty"`                 // Info for categories added since last request
	CategoriesRemoved []string            `json:"categories_removed,omitempty" yaml:"categories_removed,omitempty"` // List of categories removed since last request
	Tags              []string            `json:"tags,omitempty" yaml:"tags,omitempty"`                             // List of tags added since last request
	TagsRemoved       []string            `json:"tags_removed,omitempty" yaml:"tags_removed,omitempty"`             // List of tags removed since last request
	ServerState       State               `json:"server_state,omitempty" yaml:"server_state,omitempty"`             // Global transfer info
}

// Category is a category.
type Category struct {
	Name     string `json:"name,omitempty" yaml:"name,omitempty"`         // Name of the category
	SavePath string `json:"savePath,omitempty" yaml:"savePath,omitempty"` // Save path of the category
}

// SyncTorrentPeersRequest is a sync torrentPeers request.
type SyncTorrentPeersRequest struct {
	Hash string `json:"hash,omitempty" yaml:"hash,omitempty"` // Torrent hash
	Rid  int64  `json:"rid,omitempty" yaml:"rid,omitempty"`   // Response ID. If not provided, rid=0 will be assumed. If the given rid is different from the one of last server reply, full_update will be true (see the server reply details for more info)
}

// SyncTorrentPeers creates a sync torrentPeers request.
func SyncTorrentPeers(hash string) *SyncTorrentPeersRequest {
	return &SyncTorrentPeersRequest{
		Hash: hash,
	}
}

// Do executes the request against the provided context and client.
func (req *SyncTorrentPeersRequest) Do(ctx context.Context, cl *Client) (*SyncTorrentPeersResponse, error) {
	res := new(SyncTorrentPeersResponse)
	if err := cl.Do(ctx, "sync/torrentPeers", req, res); err != nil {
		return nil, err
	}
	return res, nil
}

// SyncTorrentPeersResponse is the sync torrentPeers response.
//
// TODO: not documented.
type SyncTorrentPeersResponse struct {
}

// WithResponseID sets the response id. If not provided, rid=0 will be assumed.
// If the given rid is different from the one of last server reply, full_update
// will be true (see the server reply details for more info)
func (req SyncTorrentPeersRequest) WithResponseID(rid int64) *SyncTorrentPeersRequest {
	req.Rid = rid
	return &req
}

// TransferInfoRequest is a transfer info request.
type TransferInfoRequest struct{}

// TransferInfo creates a transfer info request.
func TransferInfo() *TransferInfoRequest {
	return &TransferInfoRequest{}
}

// Do executes the request against the provided context and client.
func (req *TransferInfoRequest) Do(ctx context.Context, cl *Client) (*TransferInfoResponse, error) {
	res := new(TransferInfoResponse)
	if err := cl.Do(ctx, "transfer/info", req, res); err != nil {
		return nil, err
	}
	return res, nil
}

// TransferInfoResponse is the transfer info response.
type TransferInfoResponse struct {
	DlInfoSpeed      Rate             `json:"dl_info_speed,omitempty" yaml:"dl_info_speed,omitempty"`         // Global download rate (bytes/s)
	DlInfoData       ByteCount        `json:"dl_info_data,omitempty" yaml:"dl_info_data,omitempty"`           // Data downloaded this session (bytes)
	UpInfoSpeed      Rate             `json:"up_info_speed,omitempty" yaml:"up_info_speed,omitempty"`         // Global upload rate (bytes/s)
	UpInfoData       ByteCount        `json:"up_info_data,omitempty" yaml:"up_info_data,omitempty"`           // Data uploaded this session (bytes)
	DlRateLimit      Rate             `json:"dl_rate_limit,omitempty" yaml:"dl_rate_limit,omitempty"`         // Download rate limit (bytes/s)
	UpRateLimit      Rate             `json:"up_rate_limit,omitempty" yaml:"up_rate_limit,omitempty"`         // Upload rate limit (bytes/s)
	DhtNodes         int64            `json:"dht_nodes,omitempty" yaml:"dht_nodes,omitempty"`                 // DHT nodes connected to
	ConnectionStatus ConnectionStatus `json:"connection_status,omitempty" yaml:"connection_status,omitempty"` // Connection status. See possible values here below
}

// TransferSpeedLimitsModeRequest is a transfer speedLimitsMode request.
type TransferSpeedLimitsModeRequest struct{}

// TransferSpeedLimitsMode creates a transfer speedLimitsMode request.
func TransferSpeedLimitsMode() *TransferSpeedLimitsModeRequest {
	return &TransferSpeedLimitsModeRequest{}
}

// Do executes the request against the provided context and client.
func (req *TransferSpeedLimitsModeRequest) Do(ctx context.Context, cl *Client) (bool, error) {
	var res Bool
	if err := cl.Do(ctx, "transfer/speedLimitsMode", req, &res); err != nil {
		return false, err
	}
	return bool(res), nil
}

// TransferToggleSpeedLimitsModeRequest is a transfer toggleSpeedLimitsMode request.
type TransferToggleSpeedLimitsModeRequest struct{}

// TransferToggleSpeedLimitsMode creates a transfer toggleSpeedLimitsMode request.
func TransferToggleSpeedLimitsMode() *TransferToggleSpeedLimitsModeRequest {
	return &TransferToggleSpeedLimitsModeRequest{}
}

// Do executes the request against the provided context and client.
func (req *TransferToggleSpeedLimitsModeRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "transfer/toggleSpeedLimitsMode", req, nil)
}

// TransferDownloadLimitRequest is a transfer downloadLimit request.
type TransferDownloadLimitRequest struct{}

// TransferDownloadLimit creates a transfer downloadLimit request.
func TransferDownloadLimit() *TransferDownloadLimitRequest {
	return &TransferDownloadLimitRequest{}
}

// Do executes the request against the provided context and client.
func (req *TransferDownloadLimitRequest) Do(ctx context.Context, cl *Client) (Rate, error) {
	var res Rate
	if err := cl.Do(ctx, "transfer/downloadLimit", req, &res); err != nil {
		return 0, err
	}
	return res, nil
}

// TransferSetDownloadLimitRequest is a transfer setDownloadLimit request.
type TransferSetDownloadLimitRequest struct {
	Limit Rate
}

// TransferSetDownloadLimit creates a transfer setDownloadLimit request.
func TransferSetDownloadLimit(limit Rate) *TransferSetDownloadLimitRequest {
	return &TransferSetDownloadLimitRequest{
		Limit: limit,
	}
}

// Do executes the request against the provided context and client.
func (req *TransferSetDownloadLimitRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "transfer/setDownloadLimit", req, nil)
}

// TransferUploadLimitRequest is a transfer uploadLimit request.
type TransferUploadLimitRequest struct{}

// TransferUploadLimit creates a transfer uploadLimit request.
func TransferUploadLimit() *TransferUploadLimitRequest {
	return &TransferUploadLimitRequest{}
}

// Do executes the request against the provided context and client.
func (req *TransferUploadLimitRequest) Do(ctx context.Context, cl *Client) (Rate, error) {
	var res Rate
	if err := cl.Do(ctx, "transfer/uploadLimit", req, &res); err != nil {
		return 0, err
	}
	return res, nil
}

// TransferSetUploadLimitRequest is a transfer setUploadLimit request.
type TransferSetUploadLimitRequest struct {
	Limit Rate
}

// TransferSetUploadLimit creates a transfer setUploadLimit request.
func TransferSetUploadLimit(limit Rate) *TransferSetUploadLimitRequest {
	return &TransferSetUploadLimitRequest{
		Limit: limit,
	}
}

// Do executes the request against the provided context and client.
func (req *TransferSetUploadLimitRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "transfer/setUploadLimit", req, nil)
}

// TransferBanPeersRequest is a transfer banPeers request.
type TransferBanPeersRequest struct {
	Peers []string `json:"peers,omitempty" yaml:"peers,omitempty"` // The peer to ban, or multiple peers separated by a pipe |. Each peer is a colon-separated host:port
}

// TransferBanPeers creates a transfer banPeers request.
func TransferBanPeers(peers ...string) *TransferBanPeersRequest {
	return &TransferBanPeersRequest{
		Peers: peers,
	}
}

// Do executes the request against the provided context and client.
func (req *TransferBanPeersRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "transfer/banPeers", req, nil)
}

// TorrentsInfoRequest is a torrents info request.
type TorrentsInfoRequest struct {
	Filter   FilterType `json:"filter,omitempty" yaml:"filter,omitempty"`
	Category string     `json:"category,omitempty" yaml:"category,omitempty"`
	Sort     string     `json:"sort,omitempty" yaml:"sort,omitempty"`
	Reverse  bool       `json:"reverse,omitempty" yaml:"reverse,omitempty"`
	Limit    int64      `json:"limit,omitempty" yaml:"limit,omitempty"`
	Offset   int64      `json:"offset,omitempty" yaml:"offset,omitempty"`
	Hashes   []string   `json:"hashes,omitempty" yaml:"hashes,omitempty"`
}

// TorrentsInfo creates a torrents info request.
func TorrentsInfo(hashes ...string) *TorrentsInfoRequest {
	return &TorrentsInfoRequest{
		Hashes: hashes,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsInfoRequest) Do(ctx context.Context, cl *Client) ([]Torrent, error) {
	var res TorrentsInfoResponse
	if err := cl.Do(ctx, "torrents/info", req, &res); err != nil {
		return nil, err
	}
	return []Torrent(res), nil
}

// WithFilter sets the filter value.
func (req TorrentsInfoRequest) WithFilter(filter FilterType) *TorrentsInfoRequest {
	req.Filter = filter
	return &req
}

// WithCategory sets the category value.
func (req TorrentsInfoRequest) WithCategory(category string) *TorrentsInfoRequest {
	req.Category = category
	return &req
}

// WithSort sets the sort value.
func (req TorrentsInfoRequest) WithSort(sort string) *TorrentsInfoRequest {
	req.Sort = sort
	return &req
}

// WithReverse sets the reverse value.
func (req TorrentsInfoRequest) WithReverse(reverse bool) *TorrentsInfoRequest {
	req.Reverse = reverse
	return &req
}

// WithLimit sets the limit value.
func (req TorrentsInfoRequest) WithLimit(limit int64) *TorrentsInfoRequest {
	req.Limit = limit
	return &req
}

// WithOffset sets the offset value.
func (req TorrentsInfoRequest) WithOffset(offset int64) *TorrentsInfoRequest {
	req.Offset = offset
	return &req
}

// TorrentsInfoResponse is the torrents info response.
type TorrentsInfoResponse []Torrent

// TorrentsPropertiesRequest is a torrents properties request.
type TorrentsPropertiesRequest struct {
	Hash string `json:"hash" yaml:"hash"` // The hash of the torrent you want to get the contents of
}

// TorrentsProperties creates a torrents properties request.
func TorrentsProperties(hash string) *TorrentsPropertiesRequest {
	return &TorrentsPropertiesRequest{
		Hash: hash,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsPropertiesRequest) Do(ctx context.Context, cl *Client) (*TorrentProperties, error) {
	res := new(TorrentProperties)
	if err := cl.Do(ctx, "torrents/properties", req, res); err != nil {
		return nil, err
	}
	return res, nil
}

// TorrentProperties are torrent properties.
type TorrentProperties struct {
	SavePath               string    `json:"save_path,omitempty" yaml:"save_path,omitempty"`                               // Torrent save path
	CreationDate           Time      `json:"creation_date,omitempty" yaml:"creation_date,omitempty"`                       // Torrent creation date (Unix timestamp)
	PieceSize              ByteCount `json:"piece_size,omitempty" yaml:"piece_size,omitempty"`                             // Torrent piece size (bytes)
	Comment                string    `json:"comment,omitempty" yaml:"comment,omitempty"`                                   // Torrent comment
	TotalWasted            ByteCount `json:"total_wasted,omitempty" yaml:"total_wasted,omitempty"`                         // Total data wasted for torrent (bytes)
	TotalUploaded          ByteCount `json:"total_uploaded,omitempty" yaml:"total_uploaded,omitempty"`                     // Total data uploaded for torrent (bytes)
	TotalUploadedSession   ByteCount `json:"total_uploaded_session,omitempty" yaml:"total_uploaded_session,omitempty"`     // Total data uploaded this session (bytes)
	TotalDownloaded        ByteCount `json:"total_downloaded,omitempty" yaml:"total_downloaded,omitempty"`                 // Total data downloaded for torrent (bytes)
	TotalDownloadedSession ByteCount `json:"total_downloaded_session,omitempty" yaml:"total_downloaded_session,omitempty"` // Total data downloaded this session (bytes)
	UpLimit                Rate      `json:"up_limit,omitempty" yaml:"up_limit,omitempty"`                                 // Torrent upload limit (bytes/s)
	DlLimit                Rate      `json:"dl_limit,omitempty" yaml:"dl_limit,omitempty"`                                 // Torrent download limit (bytes/s)
	TimeElapsed            Duration  `json:"time_elapsed,omitempty" yaml:"time_elapsed,omitempty"`                         // Torrent elapsed time (seconds)
	SeedingTime            Duration  `json:"seeding_time,omitempty" yaml:"seeding_time,omitempty"`                         // Torrent elapsed time while complete (seconds)
	NbConnections          int64     `json:"nb_connections,omitempty" yaml:"nb_connections,omitempty"`                     // Torrent connection count
	NbConnectionsLimit     int64     `json:"nb_connections_limit,omitempty" yaml:"nb_connections_limit,omitempty"`         // Torrent connection count limit
	ShareRatio             Percent   `json:"share_ratio,omitempty" yaml:"share_ratio,omitempty"`                           // Torrent share ratio
	AdditionDate           Time      `json:"addition_date,omitempty" yaml:"addition_date,omitempty"`                       // When this torrent was added (unix timestamp)
	CompletionDate         Time      `json:"completion_date,omitempty" yaml:"completion_date,omitempty"`                   // Torrent completion date (unix timestamp)
	CreatedBy              string    `json:"created_by,omitempty" yaml:"created_by,omitempty"`                             // Torrent creator
	DlSpeedAvg             Rate      `json:"dl_speed_avg,omitempty" yaml:"dl_speed_avg,omitempty"`                         // Torrent average download speed (bytes/second)
	DlSpeed                Rate      `json:"dl_speed,omitempty" yaml:"dl_speed,omitempty"`                                 // Torrent download speed (bytes/second)
	Eta                    Duration  `json:"eta,omitempty" yaml:"eta,omitempty"`                                           // Torrent ETA (seconds)
	LastSeen               Time      `json:"last_seen,omitempty" yaml:"last_seen,omitempty"`                               // Last seen complete date (unix timestamp)
	Peers                  int64     `json:"peers,omitempty" yaml:"peers,omitempty"`                                       // Number of peers connected to
	PeersTotal             int64     `json:"peers_total,omitempty" yaml:"peers_total,omitempty"`                           // Number of peers in the swarm
	PiecesHave             int64     `json:"pieces_have,omitempty" yaml:"pieces_have,omitempty"`                           // Number of pieces owned
	PiecesNum              int64     `json:"pieces_num,omitempty" yaml:"pieces_num,omitempty"`                             // Number of pieces of the torrent
	Reannounce             Duration  `json:"reannounce,omitempty" yaml:"reannounce,omitempty"`                             // Number of seconds until the next announce
	Seeds                  int64     `json:"seeds,omitempty" yaml:"seeds,omitempty"`                                       // Number of seeds connected to
	SeedsTotal             int64     `json:"seeds_total,omitempty" yaml:"seeds_total,omitempty"`                           // Number of seeds in the swarm
	TotalSize              ByteCount `json:"total_size,omitempty" yaml:"total_size,omitempty"`                             // Torrent total size (bytes)
	UpSpeedAvg             Rate      `json:"up_speed_avg,omitempty" yaml:"up_speed_avg,omitempty"`                         // Torrent average upload speed (bytes/second)
	UpSpeed                Rate      `json:"up_speed,omitempty" yaml:"up_speed,omitempty"`                                 // Torrent upload speed (bytes/second)
}

// TorrentsTrackersRequest is a torrents trackers request.
type TorrentsTrackersRequest struct {
	Hash string `json:"hash" yaml:"hash"` // The hash of the torrent you want to get the contents of
}

// TorrentsTrackers creates a torrents trackers request.
func TorrentsTrackers(hash string) *TorrentsTrackersRequest {
	return &TorrentsTrackersRequest{
		Hash: hash,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsTrackersRequest) Do(ctx context.Context, cl *Client) ([]Tracker, error) {
	var res []Tracker
	if err := cl.Do(ctx, "torrents/trackers", req, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// Tracker holds information about a torrent tracker.
type Tracker struct {
	URL           string        `json:"url,omitempty" yaml:"url,omitempty"`                       // Tracker url
	Status        TrackerStatus `json:"status,omitempty" yaml:"status,omitempty"`                 // Tracker status. See the table below for possible values
	Tier          int64         `json:"tier,omitempty" yaml:"tier,omitempty"`                     // Tracker priority tier. Lower tier trackers are tried before higher tiers
	NumPeers      int64         `json:"num_peers,omitempty" yaml:"num_peers,omitempty"`           // Number of peers for current torrent, as reported by the tracker
	NumSeeds      int64         `json:"num_seeds,omitempty" yaml:"num_seeds,omitempty"`           // Number of seeds for current torrent, asreported by the tracker
	NumLeeches    int64         `json:"num_leeches,omitempty" yaml:"num_leeches,omitempty"`       // Number of leeches for current torrent, as reported by the tracker
	NumDownloaded int64         `json:"num_downloaded,omitempty" yaml:"num_downloaded,omitempty"` // Number of completed downlods for current torrent, as reported by the tracker
	Msg           string        `json:"msg,omitempty" yaml:"msg,omitempty"`                       // Tracker message (there is no way of knowing what this message is - it's up to tracker admins)
}

// TorrentsWebseedsRequest is a torrents webseeds request.
type TorrentsWebseedsRequest struct {
	Hash string `json:"hash" yaml:"hash"` // The hash of the torrent you want to get the contents of
}

// TorrentsWebseeds creates a torrents webseeds request.
func TorrentsWebseeds(hash string) *TorrentsWebseedsRequest {
	return &TorrentsWebseedsRequest{
		Hash: hash,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsWebseedsRequest) Do(ctx context.Context, cl *Client) ([]Webseed, error) {
	var res []Webseed
	if err := cl.Do(ctx, "torrents/webseeds", req, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// Webseed is a webseed.
type Webseed struct {
	URL string `json:"url,omitempty" yaml:"url,omitempty"` // URL of the web seed
}

// TorrentsFilesRequest is a torrents files request.
type TorrentsFilesRequest struct {
	Hash string `json:"hash" yaml:"hash"` // The hash of the torrent you want to get the contents of
}

// TorrentsFiles creates a torrents files request.
func TorrentsFiles(hash string) *TorrentsFilesRequest {
	return &TorrentsFilesRequest{
		Hash: hash,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsFilesRequest) Do(ctx context.Context, cl *Client) ([]File, error) {
	var res []File
	if err := cl.Do(ctx, "torrents/files", req, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// File is a file.
type File struct {
	Name         string       `json:"name,omitempty" yaml:"name,omitempty"`                 // File name (including relative path)
	Size         ByteCount    `json:"size,omitempty" yaml:"size,omitempty"`                 // File size (bytes)
	Progress     Percent      `json:"progress,omitempty" yaml:"progress,omitempty"`         // File progress (percentage/100)
	Priority     FilePriority `json:"priority,omitempty" yaml:"priority,omitempty"`         // File priority. See possible values here below
	IsSeed       bool         `json:"is_seed,omitempty" yaml:"is_seed,omitempty"`           // True if file is seeding/complete
	PieceRange   []int64      `json:"piece_range,omitempty" yaml:"piece_range,omitempty"`   // The first number is the starting piece index and the second number is the ending piece index (inclusive)
	Availability Percent      `json:"availability,omitempty" yaml:"availability,omitempty"` // Percentage of file pieces currently available
}

// FilePriority is the file priority enum.
type FilePriority int

// File priority values.
const (
	FilePriorityDoNotDownload FilePriority = 0
	FilePriorityNormal        FilePriority = 1
	FilePriorityHigh          FilePriority = 6
	FilePriorityMaximal       FilePriority = 7
)

// TorrentsPieceStatesRequest is a torrents pieceStates request.
type TorrentsPieceStatesRequest struct {
	Hash string `json:"hash" yaml:"hash"` // The hash of the torrent you want to get the contents of
}

// TorrentsPieceStates creates a torrents pieceStates request.
func TorrentsPieceStates(hash string) *TorrentsPieceStatesRequest {
	return &TorrentsPieceStatesRequest{
		Hash: hash,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsPieceStatesRequest) Do(ctx context.Context, cl *Client) ([]PieceState, error) {
	var res []PieceState
	if err := cl.Do(ctx, "torrents/pieceStates", req, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// PieceState is the piece state enum.
type PieceState int

// Piece state values.
const (
	PieceStateNotDownloadedYet PieceState = iota
	PieceStateDownloading
	PieceStateDone
)

// TorrentsPieceHashesRequest is a torrents pieceHashes request.
type TorrentsPieceHashesRequest struct {
	Hash string `json:"hash" yaml:"hash"` // The hash of the torrent you want to get the contents of
}

// TorrentsPieceHashes creates a torrents pieceHashes request.
func TorrentsPieceHashes(hash string) *TorrentsPieceHashesRequest {
	return &TorrentsPieceHashesRequest{
		Hash: hash,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsPieceHashesRequest) Do(ctx context.Context, cl *Client) ([]string, error) {
	var res []string
	if err := cl.Do(ctx, "torrents/pieceHashes", req, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// TorrentsPauseRequest is a torrents pause request.
type TorrentsPauseRequest struct {
	Hashes []string `json:"hashes" yaml:"hashes"` // The hashes of the torrents you want to pause. hashes can contain multiple hashes separated by |, to pause multiple torrents, or set to all, to pause all torrents.
}

// TorrentsPause creates a torrents pause request.
func TorrentsPause(hashes ...string) *TorrentsPauseRequest {
	return &TorrentsPauseRequest{
		Hashes: hashes,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsPauseRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/pause", req, nil)
}

// TorrentsResumeRequest is a torrents resume request.
type TorrentsResumeRequest struct {
	Hashes []string `json:"hashes" yaml:"hashes"` // The hashes of the torrents you want to pause. hashes can contain multiple hashes separated by |, to pause multiple torrents, or set to all, to pause all torrents.
}

// TorrentsResume creates a torrents resume request.
func TorrentsResume(hashes ...string) *TorrentsResumeRequest {
	return &TorrentsResumeRequest{
		Hashes: hashes,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsResumeRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/resume", req, nil)
}

// TorrentsDeleteRequest is a torrents delete request.
type TorrentsDeleteRequest struct {
	Hashes      []string `json:"hashes" yaml:"hashes"`           // The hashes of the torrents you want to pause. hashes can contain multiple hashes separated by |, to pause multiple torrents, or set to all, to pause all torrents.
	DeleteFiles bool     `json:"deleteFiles" yaml:"deleteFiles"` // If set to true, the downloaded data will also be deleted, otherwise has no effect.
}

// TorrentsDelete creates a torrents delete request.
func TorrentsDelete(deleteFiles bool, hashes ...string) *TorrentsDeleteRequest {
	return &TorrentsDeleteRequest{
		Hashes:      hashes,
		DeleteFiles: deleteFiles,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsDeleteRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/delete", req, nil)
}

// TorrentsRecheckRequest is a torrents reannounce request.
type TorrentsRecheckRequest struct {
	Hashes []string `json:"hashes" yaml:"hashes"` // The hashes of the torrents you want to pause. hashes can contain multiple hashes separated by |, to pause multiple torrents, or set to all, to pause all torrents.
}

// TorrentsRecheck creates a torrents reannounce request.
func TorrentsRecheck(hashes ...string) *TorrentsRecheckRequest {
	return &TorrentsRecheckRequest{
		Hashes: hashes,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsRecheckRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/recheck", req, nil)
}

// TorrentsReannounceRequest is a torrents reannounce request.
type TorrentsReannounceRequest struct {
	Hashes []string `json:"hashes" yaml:"hashes"` // The hashes of the torrents you want to pause. hashes can contain multiple hashes separated by |, to pause multiple torrents, or set to all, to pause all torrents.
}

// TorrentsReannounce creates a torrents reannounce request.
func TorrentsReannounce(hashes ...string) *TorrentsReannounceRequest {
	return &TorrentsReannounceRequest{
		Hashes: hashes,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsReannounceRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/reannounce", req, nil)
}

// TorrentsAddRequest is a torrents add request.
type TorrentsAddRequest struct {
	URLs               []string          `json:"urls,omitempty" yaml:"urls,omitempty"`                             // URLs separated with newlines
	Torrents           map[string][]byte `json:"torrents,omitempty" yaml:"torrents,omitempty"`                     // Raw data of torrent file. torrents can be presented multiple times.
	Savepath           string            `json:"savepath,omitempty" yaml:"savepath,omitempty"`                     // Download folder
	Cookie             url.Values        `json:"cookie,omitempty" yaml:"cookie,omitempty"`                         // Cookie sent to download the .torrent file
	Category           string            `json:"category,omitempty" yaml:"category,omitempty"`                     // Category for the torrent
	SkipChecking       bool              `json:"skip_checking,omitempty" yaml:"skip_checking,omitempty"`           // Skip hash checking. Possible values are true, false (default)
	Paused             bool              `json:"paused,omitempty" yaml:"paused,omitempty"`                         // Add torrents in the paused state. Possible values are true, false (default)
	RootFolder         RootType          `json:"root_folder,omitempty" yaml:"root_folder,omitempty"`               // Create the root folder. Possible values are true, false, unset (default)
	Rename             string            `json:"rename,omitempty" yaml:"rename,omitempty"`                         // Rename torrent
	UpLimit            Rate              `json:"upLimit,omitempty" yaml:"upLimit,omitempty"`                       // Set torrent upload speed limit. Unit in bytes/second
	DlLimit            Rate              `json:"dlLimit,omitempty" yaml:"dlLimit,omitempty"`                       // Set torrent download speed limit. Unit in bytes/second
	AutoTMM            bool              `json:"autoTMM,omitempty" yaml:"autoTMM,omitempty"`                       // Whether Automatic Torrent Management should be used
	SequentialDownload bool              `json:"sequentialDownload,omitempty" yaml:"sequentialDownload,omitempty"` // Enable sequential download. Possible values are true, false (default)
	FirstLastPiecePrio bool              `json:"firstLastPiecePrio,omitempty" yaml:"firstLastPiecePrio,omitempty"` // Prioritize download first last piece. Possible values are true, false (default)
}

// EncodeFormData encodes the torrent add request as form data.
func (req *TorrentsAddRequest) EncodeFormData(w io.Writer) (string, error) {
	m := multipart.NewWriter(w)

	// add torrent file data
	for n, buf := range req.Torrents {
		f, err := m.CreateFormFile("torrents", n)
		if err != nil {
			return "", nil
		}
		if _, err = f.Write(buf); err != nil {
			return "", err
		}
	}

	// build vals
	vals := make(map[string]string)
	if len(req.URLs) != 0 {
		vals["urls"] = strings.Join(req.URLs, "\n")
	}
	if req.Savepath != "" {
		vals["savepath"] = req.Savepath
	}
	if req.Cookie != nil {
		vals["cookie"] = req.Cookie.Encode()
	}
	if req.Category != "" {
		vals["category"] = req.Category
	}
	if req.SkipChecking {
		vals["skip_checking"] = "true"
	}
	if req.Paused {
		vals["paused"] = "true"
	}
	if req.RootFolder != "" {
		vals["root_folder"] = string(req.RootFolder)
	}
	if req.Rename != "" {
		vals["rename"] = req.Rename
	}
	if req.UpLimit != 0 {
		vals["upLimit"] = strconv.FormatInt(int64(req.UpLimit), 10)
	}
	if req.DlLimit != 0 {
		vals["dlLimit"] = strconv.FormatInt(int64(req.DlLimit), 10)
	}
	if req.AutoTMM {
		vals["autoTMM"] = "true"
	}
	if req.SequentialDownload {
		vals["sequentialDownload"] = "true"
	}
	if req.FirstLastPiecePrio {
		vals["firstLastPiecePrio"] = "true"
	}

	// encode vals
	for k, v := range vals {
		p, err := m.CreateFormField(k)
		if err != nil {
			return "", err
		}
		if _, err = p.Write([]byte(v)); err != nil {
			return "", err
		}
	}
	if err := m.Close(); err != nil {
		return "", err
	}

	return m.FormDataContentType(), nil
}

// RootType is the root folder type enum.
type RootType string

// Root folder values.
const (
	RootFolderTrue  RootType = "true"
	RootFolderFalse RootType = "false"
	RootFolderUnset RootType = "unset"
)

// TorrentsAdd creates a torrents add request.
func TorrentsAdd() *TorrentsAddRequest {
	return &TorrentsAddRequest{
		Torrents: make(map[string][]byte),
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsAddRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/add", req, nil)
}

// WithTorrent adds a torrent to the add request.
func (req *TorrentsAddRequest) WithTorrent(name string, data []byte) *TorrentsAddRequest {
	req.Torrents[name] = data
	return req
}

// WithURLs sets urls separated with newlines.
func (req TorrentsAddRequest) WithURLs(urls []string) *TorrentsAddRequest {
	req.URLs = urls
	return &req
}

// WithSavepath sets download folder.
func (req TorrentsAddRequest) WithSavepath(savepath string) *TorrentsAddRequest {
	req.Savepath = savepath
	return &req
}

// WithCookie sets cookie sent to download the .torrent file.
func (req TorrentsAddRequest) WithCookie(name, value string) *TorrentsAddRequest {
	req.Cookie = url.Values{name: []string{value}}
	return &req
}

// WithCategory sets category for the torrent.
func (req TorrentsAddRequest) WithCategory(category string) *TorrentsAddRequest {
	req.Category = category
	return &req
}

// WithSkipChecking sets skip hash checking. Possible values are true, false (default).
func (req TorrentsAddRequest) WithSkipChecking(skipChecking bool) *TorrentsAddRequest {
	req.SkipChecking = skipChecking
	return &req
}

// WithPaused sets add torrents in the paused state. Possible values are true, false (default).
func (req TorrentsAddRequest) WithPaused(paused bool) *TorrentsAddRequest {
	req.Paused = paused
	return &req
}

// WithRootFolder sets create the root folder. Possible values are true, false, unset (default).
func (req TorrentsAddRequest) WithRootFolder(rootFolder RootType) *TorrentsAddRequest {
	req.RootFolder = rootFolder
	return &req
}

// WithRename sets rename torrent.
func (req TorrentsAddRequest) WithRename(rename string) *TorrentsAddRequest {
	req.Rename = rename
	return &req
}

// WithUpLimit sets set torrent upload speed limit. Unit in bytes/second.
func (req TorrentsAddRequest) WithUpLimit(upLimit Rate) *TorrentsAddRequest {
	req.UpLimit = upLimit
	return &req
}

// WithDlLimit sets set torrent download speed limit. Unit in bytes/second.
func (req TorrentsAddRequest) WithDlLimit(dlLimit Rate) *TorrentsAddRequest {
	req.DlLimit = dlLimit
	return &req
}

// WithAutoTMM sets whether Automatic Torrent Management should be used.
func (req TorrentsAddRequest) WithAutoTMM(autoTMM bool) *TorrentsAddRequest {
	req.AutoTMM = autoTMM
	return &req
}

// WithSequentialDownload sets enable sequential download. Possible values are true, false (default).
func (req TorrentsAddRequest) WithSequentialDownload(sequentialDownload bool) *TorrentsAddRequest {
	req.SequentialDownload = sequentialDownload
	return &req
}

// WithFirstLastPiecePrio sets prioritize download first last piece. Possible values are true, false (default).
func (req TorrentsAddRequest) WithFirstLastPiecePrio(firstLastPiecePrio bool) *TorrentsAddRequest {
	req.FirstLastPiecePrio = firstLastPiecePrio
	return &req
}

// TorrentsAddTrackersRequest is a torrents addTrackers request.
type TorrentsAddTrackersRequest struct {
	Hash string   `json:"hash" yaml:"hash"` // The hash of the torrent you want to get the contents of
	URLs []string `json:"urls" yaml:"urls"`
}

// TorrentsAddTrackers creates a torrents addTrackers request.
func TorrentsAddTrackers(hash string, urls ...string) *TorrentsAddTrackersRequest {
	return &TorrentsAddTrackersRequest{
		Hash: hash,
		URLs: urls,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsAddTrackersRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/addTrackers", map[string]interface{}{
		"hash": req.Hash,
		"urls": strings.Join(req.URLs, "\n"),
	}, nil)
}

// WithURLs sets urls separated with newlines.
func (req TorrentsAddTrackersRequest) WithURLs(urls []string) *TorrentsAddTrackersRequest {
	req.URLs = urls
	return &req
}

// TorrentsEditTrackerRequest is a torrents editTracker request.
type TorrentsEditTrackerRequest struct {
	Hash    string `json:"hash" yaml:"hash"`       // The hash of the torrent you want to get the contents of
	OrigURL string `json:"origUrl" yaml:"origUrl"` // The tracker URL you want to edit
	NewURL  string `json:"newUrl" yaml:"newUrl"`   // The new URL to replace the origUrl
}

// TorrentsEditTracker creates a torrents editTracker request.
func TorrentsEditTracker(hash, origURL, newURL string) *TorrentsEditTrackerRequest {
	return &TorrentsEditTrackerRequest{
		Hash:    hash,
		OrigURL: origURL,
		NewURL:  newURL,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsEditTrackerRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/editTracker", req, nil)
}

// TorrentsRemoveTrackersRequest is a torrents removeTrackers request.
type TorrentsRemoveTrackersRequest struct {
	Hash string   `json:"hash" yaml:"hash"` // The hash of the torrent you want to get the contents of
	URLs []string `json:"urls" yaml:"urls"`
}

// TorrentsRemoveTrackers creates a torrents removeTrackers request.
func TorrentsRemoveTrackers(hash string, urls ...string) *TorrentsRemoveTrackersRequest {
	return &TorrentsRemoveTrackersRequest{
		Hash: hash,
		URLs: urls,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsRemoveTrackersRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/removeTrackers", req, nil)
}

// WithURLs sets urls separated with newlines.
func (req TorrentsRemoveTrackersRequest) WithURLs(urls []string) *TorrentsRemoveTrackersRequest {
	req.URLs = urls
	return &req
}

// TorrentsAddPeersRequest is a torrents addPeers request.
type TorrentsAddPeersRequest struct {
	Hashes []string `json:"hashes" yaml:"hashes"` // The hashes of the torrents you want to pause. hashes can contain multiple hashes separated by |, to pause multiple torrents, or set to all, to pause all torrents.
	Peers  []string `json:"peers" yaml:"peers"`   // The peer to add, or multiple peers separated by a pipe |. Each peer is a colon-separated host:port
}

// TorrentsAddPeers creates a torrents addPeers request.
func TorrentsAddPeers(hashes ...string) *TorrentsAddPeersRequest {
	return &TorrentsAddPeersRequest{
		Hashes: hashes,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsAddPeersRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/addPeers", req, nil)
}

// WithPeers sets peers to add.
func (req TorrentsAddPeersRequest) WithPeers(peers []string) *TorrentsAddPeersRequest {
	req.Peers = peers
	return &req
}

// TorrentsIncreasePrioRequest is a torrents increasePrio request.
type TorrentsIncreasePrioRequest struct {
	Hashes []string `json:"hashes" yaml:"hashes"` // The hashes of the torrents you want to pause. hashes can contain multiple hashes separated by |, to pause multiple torrents, or set to all, to pause all torrents.
}

// TorrentsIncreasePrio creates a torrents increasePrio request.
func TorrentsIncreasePrio(hashes ...string) *TorrentsIncreasePrioRequest {
	return &TorrentsIncreasePrioRequest{
		Hashes: hashes,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsIncreasePrioRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/increasePrio", req, nil)
}

// TorrentsDecreasePrioRequest is a torrents decreasePrio request.
type TorrentsDecreasePrioRequest struct {
	Hashes []string `json:"hashes" yaml:"hashes"` // The hashes of the torrents you want to pause. hashes can contain multiple hashes separated by |, to pause multiple torrents, or set to all, to pause all torrents.
}

// TorrentsDecreasePrio creates a torrents decreasePrio request.
func TorrentsDecreasePrio(hashes ...string) *TorrentsDecreasePrioRequest {
	return &TorrentsDecreasePrioRequest{
		Hashes: hashes,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsDecreasePrioRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/decreasePrio", req, nil)
}

// TorrentsTopPrioRequest is a torrents topPrio request.
type TorrentsTopPrioRequest struct {
	Hashes []string `json:"hashes" yaml:"hashes"` // The hashes of the torrents you want to pause. hashes can contain multiple hashes separated by |, to pause multiple torrents, or set to all, to pause all torrents.
}

// TorrentsTopPrio creates a torrents topPrio request.
func TorrentsTopPrio(hashes ...string) *TorrentsTopPrioRequest {
	return &TorrentsTopPrioRequest{
		Hashes: hashes,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsTopPrioRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/topPrio", req, nil)
}

// TorrentsBottomPrioRequest is a torrents bottomPrio request.
type TorrentsBottomPrioRequest struct {
	Hashes []string `json:"hashes" yaml:"hashes"` // The hashes of the torrents you want to pause. hashes can contain multiple hashes separated by |, to pause multiple torrents, or set to all, to pause all torrents.
}

// TorrentsBottomPrio creates a torrents bottomPrio request.
func TorrentsBottomPrio(hashes ...string) *TorrentsBottomPrioRequest {
	return &TorrentsBottomPrioRequest{
		Hashes: hashes,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsBottomPrioRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/bottomPrio", req, nil)
}

// TorrentsFilePrioRequest is a torrents filePrio request.
type TorrentsFilePrioRequest struct {
	Hash     string       `json:"hash" yaml:"hash"`         // The hash of the torrent you want to get the contents of
	ID       []string     `json:"id" yaml:"id"`             // File ids, separated by |
	Priority FilePriority `json:"priority" yaml:"priority"` // File priority to set
}

// TorrentsFilePrio creates a torrents filePrio request.
func TorrentsFilePrio(hash string, priority FilePriority, id ...string) *TorrentsFilePrioRequest {
	return &TorrentsFilePrioRequest{
		Hash:     hash,
		ID:       id,
		Priority: priority,
	}
}

// WithID sets file ids, separated by |.
func (req TorrentsFilePrioRequest) WithID(id []string) *TorrentsFilePrioRequest {
	req.ID = id
	return &req
}

// WithPriority sets file priority to set.
func (req TorrentsFilePrioRequest) WithPriority(priority FilePriority) *TorrentsFilePrioRequest {
	req.Priority = priority
	return &req
}

// Do executes the request against the provided context and client.
func (req *TorrentsFilePrioRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/filePrio", req, nil)
}

// TorrentsDownloadLimitRequest is a torrents downloadLimit request.
type TorrentsDownloadLimitRequest struct {
	Hashes []string `json:"hashes" yaml:"hashes"` // The hashes of the torrents you want to pause. hashes can contain multiple hashes separated by |, to pause multiple torrents, or set to all, to pause all torrents.
}

// TorrentsDownloadLimit creates a torrents downloadLimit request.
func TorrentsDownloadLimit(hashes ...string) *TorrentsDownloadLimitRequest {
	return &TorrentsDownloadLimitRequest{
		Hashes: hashes,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsDownloadLimitRequest) Do(ctx context.Context, cl *Client) (map[string]Rate, error) {
	res := make(map[string]Rate)
	if err := cl.Do(ctx, "torrents/downloadLimit", req, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// TorrentsSetDownloadLimitRequest is a torrents setDownloadLimit request.
type TorrentsSetDownloadLimitRequest struct {
	Hashes []string `json:"hashes" yaml:"hashes"` // The hashes of the torrents you want to pause. hashes can contain multiple hashes separated by |, to pause multiple torrents, or set to all, to pause all torrents.
	Limit  Rate     `json:"limit" yaml:"limit"`
}

// TorrentsSetDownloadLimit creates a torrents setDownloadLimit request.
func TorrentsSetDownloadLimit(limit Rate, hashes ...string) *TorrentsSetDownloadLimitRequest {
	return &TorrentsSetDownloadLimitRequest{
		Hashes: hashes,
		Limit:  limit,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsSetDownloadLimitRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/setDownloadLimit", req, nil)
}

// TorrentsSetShareLimitsRequest is a torrents setShareLimits request.
type TorrentsSetShareLimitsRequest struct {
	Hashes           []string `json:"hashes" yaml:"hashes"`                     // The hashes of the torrents you want to pause. hashes can contain multiple hashes separated by |, to pause multiple torrents, or set to all, to pause all torrents.
	RatioLimit       Percent  `json:"ratioLimit" yaml:"ratioLimit"`             // The max ratio the torrent should be seeded until. -2 means the global limit should be used, -1 means no limit.
	SeedingTimeLimit Duration `json:"seedingTimeLimit" yaml:"seedingTimeLimit"` // The max amount of time the torrent should be seeded. -2 means the global limit should be used, -1 means no limit.
}

// TorrentsSetShareLimits creates a torrents setShareLimits request.
func TorrentsSetShareLimits(hashes ...string) *TorrentsSetShareLimitsRequest {
	return &TorrentsSetShareLimitsRequest{
		Hashes: hashes,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsSetShareLimitsRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/setShareLimits", req, nil)
}

// WithRatioLimit sets the max ratio the torrent should be seeded until. -2 means the global limit should be used, -1 means no limit.
func (req TorrentsSetShareLimitsRequest) WithRatioLimit(ratioLimit Percent) *TorrentsSetShareLimitsRequest {
	req.RatioLimit = ratioLimit
	return &req
}

// WithSeedingTimeLimit sets the max amount of time the torrent should be seeded. -2 means the global limit should be used, -1 means no limit.
func (req TorrentsSetShareLimitsRequest) WithSeedingTimeLimit(seedingTimeLimit Duration) *TorrentsSetShareLimitsRequest {
	req.SeedingTimeLimit = seedingTimeLimit
	return &req
}

// TorrentsUploadLimitRequest is a torrents uploadLimit request.
type TorrentsUploadLimitRequest struct {
	Hashes []string `json:"hashes" yaml:"hashes"` // The hashes of the torrents you want to pause. hashes can contain multiple hashes separated by |, to pause multiple torrents, or set to all, to pause all torrents.
}

// TorrentsUploadLimit creates a torrents uploadLimit request.
func TorrentsUploadLimit(hashes ...string) *TorrentsUploadLimitRequest {
	return &TorrentsUploadLimitRequest{
		Hashes: hashes,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsUploadLimitRequest) Do(ctx context.Context, cl *Client) (map[string]Rate, error) {
	res := make(map[string]Rate)
	if err := cl.Do(ctx, "torrents/uploadLimit", req, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// TorrentsSetUploadLimitRequest is a torrents setUploadLimit request.
type TorrentsSetUploadLimitRequest struct {
	Hashes []string `json:"hashes" yaml:"hashes"` // The hashes of the torrents you want to pause. hashes can contain multiple hashes separated by |, to pause multiple torrents, or set to all, to pause all torrents.
	Limit  Rate     `json:"limit" yaml:"limit"`
}

// TorrentsSetUploadLimit creates a torrents setUploadLimit request.
func TorrentsSetUploadLimit(limit Rate, hashes ...string) *TorrentsSetUploadLimitRequest {
	return &TorrentsSetUploadLimitRequest{
		Hashes: hashes,
		Limit:  limit,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsSetUploadLimitRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/setUploadLimit", req, nil)
}

// TorrentsSetLocationRequest is a torrents setLocation request.
type TorrentsSetLocationRequest struct {
	Hashes   []string `json:"hashes" yaml:"hashes"`     // The hashes of the torrents you want to pause. hashes can contain multiple hashes separated by |, to pause multiple torrents, or set to all, to pause all torrents.
	Location string   `json:"location" yaml:"location"` // The location to download the torrent to. If the location doesn't exist, the torrent's location is unchanged.
}

// TorrentsSetLocation creates a torrents setLocation request.
func TorrentsSetLocation(location string, hashes ...string) *TorrentsSetLocationRequest {
	return &TorrentsSetLocationRequest{
		Hashes:   hashes,
		Location: location,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsSetLocationRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/setLocation", req, nil)
}

// TorrentsRenameRequest is a torrents rename request.
type TorrentsRenameRequest struct {
	Hash string `json:"hash" yaml:"hash"` // The hash of the torrent you want to get the contents of
	Name string `json:"name" yaml:"name"` // Name to set.
}

// TorrentsRename creates a torrents rename request.
func TorrentsRename(hash string, name string) *TorrentsRenameRequest {
	return &TorrentsRenameRequest{
		Hash: hash,
		Name: name,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsRenameRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/rename", req, nil)
}

// TorrentsSetCategoryRequest is a torrents setCategory request.
type TorrentsSetCategoryRequest struct {
	Hashes   []string `json:"hashes" yaml:"hashes"`     // The hashes of the torrents you want to pause. hashes can contain multiple hashes separated by |, to pause multiple torrents, or set to all, to pause all torrents.
	Category string   `json:"category" yaml:"category"` // The torrent category you want to set.
}

// TorrentsSetCategory creates a torrents setCategory request.
func TorrentsSetCategory(category string, hashes ...string) *TorrentsSetCategoryRequest {
	return &TorrentsSetCategoryRequest{
		Hashes:   hashes,
		Category: category,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsSetCategoryRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/setCategory", req, nil)
}

// TorrentsCategoriesRequest is a torrents categories request.
type TorrentsCategoriesRequest struct{}

// TorrentsCategories creates a torrents categories request.
func TorrentsCategories() *TorrentsCategoriesRequest {
	return &TorrentsCategoriesRequest{}
}

// Do executes the request against the provided context and client.
func (req *TorrentsCategoriesRequest) Do(ctx context.Context, cl *Client) (map[string]Category, error) {
	res := make(map[string]Category)
	if err := cl.Do(ctx, "torrents/categories", req, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// TorrentsCreateCategoryRequest is a torrents createCategory request.
type TorrentsCreateCategoryRequest struct {
	Category string `json:"category" yaml:"category"` // The category you want to create.
	SavePath string `json:"savePath" yaml:"savePath"` // Category save path.
}

// TorrentsCreateCategory creates a torrents createCategory request.
func TorrentsCreateCategory(category, savePath string) *TorrentsCreateCategoryRequest {
	return &TorrentsCreateCategoryRequest{
		Category: category,
		SavePath: savePath,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsCreateCategoryRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/createCategory", req, nil)
}

// TorrentsEditCategoryRequest is a torrents editCategory request.
type TorrentsEditCategoryRequest struct {
	Category string `json:"category" yaml:"category"` // The category you want to create.
	SavePath string `json:"savePath" yaml:"savePath"` // Category save path.
}

// TorrentsEditCategory creates a torrents editCategory request.
func TorrentsEditCategory(category, savePath string) *TorrentsEditCategoryRequest {
	return &TorrentsEditCategoryRequest{
		Category: category,
		SavePath: savePath,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsEditCategoryRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/editCategory", req, nil)
}

// TorrentsRemoveCategoriesRequest is a torrents removeCategories request.
type TorrentsRemoveCategoriesRequest struct {
	Categories []string `json:"categories" yaml:"categories"`
}

// TorrentsRemoveCategories creates a torrents removeCategories request.
func TorrentsRemoveCategories(categories ...string) *TorrentsRemoveCategoriesRequest {
	return &TorrentsRemoveCategoriesRequest{}
}

// Do executes the request against the provided context and client.
func (req *TorrentsRemoveCategoriesRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/removeCategories", map[string]interface{}{
		"categories": strings.Join(req.Categories, "\n"),
	}, nil)
}

// TorrentsAddTagsRequest is a torrents addTags request.
type TorrentsAddTagsRequest struct {
	Hashes []string `json:"hashes" yaml:"hashes"` // The hashes of the torrents you want to pause. hashes can contain multiple hashes separated by |, to pause multiple torrents, or set to all, to pause all torrents.
	Tags   []string `json:"tags" yaml:"tags"`     // The list of tags you want to add to passed torrents
}

// TorrentsAddTags creates a torrents addTags request.
func TorrentsAddTags(hashes ...string) *TorrentsAddTagsRequest {
	return &TorrentsAddTagsRequest{
		Hashes: hashes,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsAddTagsRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/addTags", map[string]interface{}{
		"hashes": strings.Join(req.Hashes, "|"),
		"tags":   strings.Join(req.Tags, ","),
	}, nil)
}

// WithTags sets the list of tags you want to add to passed torrents.
func (req TorrentsAddTagsRequest) WithTags(tags []string) *TorrentsAddTagsRequest {
	req.Tags = tags
	return &req
}

// TorrentsRemoveTagsRequest is a torrents removeTags request.
type TorrentsRemoveTagsRequest struct {
	Hashes []string `json:"hashes" yaml:"hashes"` // The hashes of the torrents you want to pause. hashes can contain multiple hashes separated by |, to pause multiple torrents, or set to all, to pause all torrents.
	Tags   []string `json:"tags" yaml:"tags"`     // The list of tags you want to add to passed torrents
}

// TorrentsRemoveTags creates a torrents removeTags request.
func TorrentsRemoveTags(hashes ...string) *TorrentsRemoveTagsRequest {
	return &TorrentsRemoveTagsRequest{
		Hashes: hashes,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsRemoveTagsRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/removeTags", map[string]interface{}{
		"hashes": strings.Join(req.Hashes, "|"),
		"tags":   strings.Join(req.Tags, ","),
	}, nil)
}

// WithTags sets the list of tags you want to add to passed torrents.
func (req TorrentsRemoveTagsRequest) WithTags(tags []string) *TorrentsRemoveTagsRequest {
	req.Tags = tags
	return &req
}

// TorrentsTagsRequest is a torrents tags request.
type TorrentsTagsRequest struct{}

// TorrentsTags creates a torrents tags request.
func TorrentsTags() *TorrentsTagsRequest {
	return &TorrentsTagsRequest{}
}

// Do executes the request against the provided context and client.
func (req *TorrentsTagsRequest) Do(ctx context.Context, cl *Client) ([]string, error) {
	var res []string
	if err := cl.Do(ctx, "torrents/tags", req, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// TorrentsCreateTagsRequest is a torrents createTags request.
type TorrentsCreateTagsRequest struct {
	Tags []string `json:"tags" yaml:"tags"` // The list of tags you want to add to passed torrents
}

// TorrentsCreateTags creates a torrents createTags request.
func TorrentsCreateTags(tags ...string) *TorrentsCreateTagsRequest {
	return &TorrentsCreateTagsRequest{
		Tags: tags,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsCreateTagsRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/createTags", map[string]interface{}{
		"tags": strings.Join(req.Tags, ","),
	}, nil)
}

// TorrentsDeleteTagsRequest is a torrents deleteTags request.
type TorrentsDeleteTagsRequest struct {
	Tags []string `json:"tags" yaml:"tags"` // The list of tags you want to add to passed torrents
}

// TorrentsDeleteTags creates a torrents deleteTags request.
func TorrentsDeleteTags(tags ...string) *TorrentsDeleteTagsRequest {
	return &TorrentsDeleteTagsRequest{
		Tags: tags,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsDeleteTagsRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/deleteTags", map[string]interface{}{
		"tags": strings.Join(req.Tags, ","),
	}, nil)
}

// TorrentsSetAutoManagementRequest is a torrents setAutoManagement request.
type TorrentsSetAutoManagementRequest struct {
	Hashes []string `json:"hashes" yaml:"hashes"` // The hashes of the torrents you want to pause. hashes can contain multiple hashes separated by |, to pause multiple torrents, or set to all, to pause all torrents.
	Enable bool     `json:"enable" yaml:"enable"` // affects the torrents listed in hashes, default is false
}

// TorrentsSetAutoManagement creates a torrents setAutoManagement request.
func TorrentsSetAutoManagement(enable bool, hashes ...string) *TorrentsSetAutoManagementRequest {
	return &TorrentsSetAutoManagementRequest{
		Hashes: hashes,
		Enable: enable,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsSetAutoManagementRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/setAutoManagement", req, nil)
}

// TorrentsToggleSequentialDownloadRequest is a torrents toggleSequentialDownload request.
type TorrentsToggleSequentialDownloadRequest struct {
	Hashes []string `json:"hashes" yaml:"hashes"` // The hashes of the torrents you want to pause. hashes can contain multiple hashes separated by |, to pause multiple torrents, or set to all, to pause all torrents.
}

// TorrentsToggleSequentialDownload creates a torrents toggleSequentialDownload request.
func TorrentsToggleSequentialDownload(hashes ...string) *TorrentsToggleSequentialDownloadRequest {
	return &TorrentsToggleSequentialDownloadRequest{
		Hashes: hashes,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsToggleSequentialDownloadRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/toggleSequentialDownload", req, nil)
}

// TorrentsToggleFirstLastPiecePrioRequest is a torrents toggleFirstLastPiecePrio request.
type TorrentsToggleFirstLastPiecePrioRequest struct {
	Hashes []string `json:"hashes" yaml:"hashes"` // The hashes of the torrents you want to pause. hashes can contain multiple hashes separated by |, to pause multiple torrents, or set to all, to pause all torrents.
}

// TorrentsToggleFirstLastPiecePrio creates a torrents toggleFirstLastPiecePrio request.
func TorrentsToggleFirstLastPiecePrio(hashes ...string) *TorrentsToggleFirstLastPiecePrioRequest {
	return &TorrentsToggleFirstLastPiecePrioRequest{
		Hashes: hashes,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsToggleFirstLastPiecePrioRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/toggleFirstLastPiecePrio", req, nil)
}

// TorrentsSetForceStartRequest is a torrents setForceStart request.
type TorrentsSetForceStartRequest struct {
	Hashes []string `json:"hashes" yaml:"hashes"` // The hashes of the torrents you want to pause. hashes can contain multiple hashes separated by |, to pause multiple torrents, or set to all, to pause all torrents.
	Value  bool     `json:"value" yaml:"value"`   // affects the torrents listed in hashes, default is false
}

// TorrentsSetForceStart creates a torrents setForceStart request.
func TorrentsSetForceStart(value bool, hashes ...string) *TorrentsSetForceStartRequest {
	return &TorrentsSetForceStartRequest{
		Hashes: hashes,
		Value:  value,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsSetForceStartRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/setForceStart", req, nil)
}

// TorrentsSetSuperSeedingRequest is a torrents setSuperSeeding request.
type TorrentsSetSuperSeedingRequest struct {
	Hashes []string `json:"hashes" yaml:"hashes"` // The hashes of the torrents you want to pause. hashes can contain multiple hashes separated by |, to pause multiple torrents, or set to all, to pause all torrents.
	Value  bool     `json:"value" yaml:"value"`   // affects the torrents listed in hashes, default is false
}

// TorrentsSetSuperSeeding creates a torrents setSuperSeeding request.
func TorrentsSetSuperSeeding(value bool, hashes ...string) *TorrentsSetSuperSeedingRequest {
	return &TorrentsSetSuperSeedingRequest{
		Hashes: hashes,
		Value:  value,
	}
}

// Do executes the request against the provided context and client.
func (req *TorrentsSetSuperSeedingRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "torrents/setSuperSeeding", req, nil)
}

// RssAddFolderRequest is a rss addFolder request.
type RssAddFolderRequest struct {
	Path string `json:"path" yaml:"path"` // Full path of added folder (e.g. "The Pirate Bay\Top100")
}

// RssAddFolder creates a rss addFolder request.
func RssAddFolder(path string) *RssAddFolderRequest {
	return &RssAddFolderRequest{
		Path: path,
	}
}

// Do executes the request against the provided context and client.
func (req *RssAddFolderRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "rss/addFolder", req, nil)
}

// RssAddFeedRequest is a rss addFeed request.
type RssAddFeedRequest struct {
	URL  string `json:"url" yaml:"url"`                       // URL of RSS feed (e.g. "http://thepiratebay.org/rss//top100/200")
	Path string `json:"path,omitempty" yaml:"path,omitempty"` // Full path of added folder (e.g. "The Pirate Bay\Top100")
}

// RssAddFeed creates a rss addFeed request.
func RssAddFeed(urlstr string) *RssAddFeedRequest {
	return &RssAddFeedRequest{
		URL: urlstr,
	}
}

// Do executes the request against the provided context and client.
func (req *RssAddFeedRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "rss/addFeed", req, nil)
}

// WithPath sets full path of added folder (e.g. "The Pirate Bay\Top100").
func (req RssAddFeedRequest) WithPath(path string) *RssAddFeedRequest {
	req.Path = path
	return &req
}

// RssRemoveItemRequest is a rss removeItem request.
type RssRemoveItemRequest struct {
	Path string `json:"path" yaml:"path"` // Full path of added folder (e.g. "The Pirate Bay\Top100")
}

// RssRemoveItem creates a rss removeItem request.
func RssRemoveItem(path string) *RssRemoveItemRequest {
	return &RssRemoveItemRequest{
		Path: path,
	}
}

// Do executes the request against the provided context and client.
func (req *RssRemoveItemRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "rss/removeItem", req, nil)
}

// RssMoveItemRequest is a rss moveItem request.
type RssMoveItemRequest struct {
	ItemPath string `json:"itemPath" yaml:"itemPath"` // Current full path of item (e.g. "The Pirate Bay\Top100")
	DestPath string `json:"destPath" yaml:"destPath"` // New full path of item (e.g. "The Pirate Bay")
}

// RssMoveItem creates a rss moveItem request.
func RssMoveItem(itemPath, destPath string) *RssMoveItemRequest {
	return &RssMoveItemRequest{
		ItemPath: itemPath,
		DestPath: destPath,
	}
}

// Do executes the request against the provided context and client.
func (req *RssMoveItemRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "rss/moveItem", req, nil)
}

// RssItemsRequest is a rss items request.
type RssItemsRequest struct {
	WithData bool `json:"withData" yaml:"withData"`
}

// RssItems creates a rss items request.
func RssItems(withData bool) *RssItemsRequest {
	return &RssItemsRequest{
		WithData: withData,
	}
}

// Do executes the request against the provided context and client.
func (req *RssItemsRequest) Do(ctx context.Context, cl *Client) (map[string]interface{}, error) {
	res := make(map[string]interface{})
	if err := cl.Do(ctx, "rss/items", req, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// RssSetRuleRequest is a rss setRule request.
type RssSetRuleRequest struct {
	RuleName string `json:"ruleName,omitempty" yaml:"ruleName,omitempty"` // Rule name (e.g. "Punisher")
	RuleDef  Rule   `json:"ruleDef,omitempty" yaml:"ruleDef,omitempty"`   // JSON encoded rule definition
}

// RssSetRule creates a rss setRule request.
func RssSetRule(ruleName string, ruleDef Rule) *RssSetRuleRequest {
	return &RssSetRuleRequest{
		RuleName: ruleName,
		RuleDef:  ruleDef,
	}
}

// Do executes the request against the provided context and client.
func (req *RssSetRuleRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "rss/setRule", req, nil)
}

type Rule struct {
	Enabled                   bool        `json:"enabled,omitempty" yaml:"enabled,omitempty"`                                     // Whether the rule is enabled
	MustContain               string      `json:"mustContain,omitempty" yaml:"mustContain,omitempty"`                             // The substring that the torrent name must contain
	MustNotContain            string      `json:"mustNotContain,omitempty" yaml:"mustNotContain,omitempty"`                       // The substring that the torrent name must not contain
	UseRegex                  bool        `json:"useRegex,omitempty" yaml:"useRegex,omitempty"`                                   // Enable regex mode in "mustContain" and "mustNotContain"
	EpisodeFilter             string      `json:"episodeFilter,omitempty" yaml:"episodeFilter,omitempty"`                         // Episode filter definition
	SmartFilter               bool        `json:"smartFilter,omitempty" yaml:"smartFilter,omitempty"`                             // Enable smart episode filter
	PreviouslyMatchedEpisodes []string    `json:"previouslyMatchedEpisodes,omitempty" yaml:"previouslyMatchedEpisodes,omitempty"` // The list of episode IDs already matched by smart filter
	AffectedFeeds             []string    `json:"affectedFeeds,omitempty" yaml:"affectedFeeds,omitempty"`                         // The feed URLs the rule applied to
	IgnoreDays                DaySchedule `json:"ignoreDays,omitempty" yaml:"ignoreDays,omitempty"`                               // Ignore sunsequent rule matches
	LastMatch                 string      `json:"lastMatch,omitempty" yaml:"lastMatch,omitempty"`                                 // The rule last match time
	AddPaused                 bool        `json:"addPaused,omitempty" yaml:"addPaused,omitempty"`                                 // Add matched torrent in paused mode
	AssignedCategory          string      `json:"assignedCategory,omitempty" yaml:"assignedCategory,omitempty"`                   // Assign category to the torrent
	SavePath                  string      `json:"savePath,omitempty" yaml:"savePath,omitempty"`                                   // Save torrent to the given directory
}

// RssRenameRuleRequest is a rss renameRule request.
type RssRenameRuleRequest struct {
	RuleName    string `json:"ruleName,omitempty" yaml:"ruleName,omitempty"`       // Rule name (e.g. "Punisher")
	NewRuleName string `json:"newRuleName,omitempty" yaml:"newRuleName,omitempty"` // New rule name (e.g. "The Punisher")
}

// RssRenameRule creates a rss renameRule request.
func RssRenameRule(ruleName, newRuleName string) *RssRenameRuleRequest {
	return &RssRenameRuleRequest{
		RuleName:    ruleName,
		NewRuleName: newRuleName,
	}
}

// Do executes the request against the provided context and client.
func (req *RssRenameRuleRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "rss/renameRule", req, nil)
}

// RssRemoveRuleRequest is a rss removeRule request.
type RssRemoveRuleRequest struct {
	RuleName string `json:"ruleName,omitempty" yaml:"ruleName,omitempty"` // Rule name (e.g. "Punisher")
}

// RssRemoveRule creates a rss removeRule request.
func RssRemoveRule(ruleName string) *RssRemoveRuleRequest {
	return &RssRemoveRuleRequest{
		RuleName: ruleName,
	}
}

// Do executes the request against the provided context and client.
func (req *RssRemoveRuleRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "rss/removeRule", req, nil)
}

// RssRulesRequest is a rss rules request.
type RssRulesRequest struct{}

// RssRules creates a rss rules request.
func RssRules() *RssRulesRequest {
	return &RssRulesRequest{}
}

// Do executes the request against the provided context and client.
func (req *RssRulesRequest) Do(ctx context.Context, cl *Client) (map[string]Rule, error) {
	res := make(map[string]Rule)
	if err := cl.Do(ctx, "rss/rules", req, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// SearchStartRequest is a search start request.
type SearchStartRequest struct {
	Pattern  string   `json:"pattern,omitempty" yaml:"pattern,omitempty"`   // Pattern to search for (e.g. "Ubuntu 18.04")
	Plugins  []string `json:"plugins,omitempty" yaml:"plugins,omitempty"`   // Plugins to use for searching (e.g. "legittorrents"). Supports multiple plugins separated by |. Also supports all and enabled
	Category string   `json:"category,omitempty" yaml:"category,omitempty"` // Categories to limit your search to (e.g. "legittorrents"). Available categories depend on the specified plugins. Also supports all
}

// SearchStart creates a search start request.
func SearchStart() *SearchStartRequest {
	return &SearchStartRequest{}
}

// Do executes the request against the provided context and client.
func (req *SearchStartRequest) Do(ctx context.Context, cl *Client) (*SearchStartResponse, error) {
	res := new(SearchStartResponse)
	if err := cl.Do(ctx, "search/start", req, res); err != nil {
		return nil, err
	}
	return res, nil
}

// WithPattern sets pattern to search for (e.g. "Ubuntu 18.04").
func (req SearchStartRequest) WithPattern(pattern string) *SearchStartRequest {
	req.Pattern = pattern
	return &req
}

// WithPlugins sets plugins to use for searching (e.g. "legittorrents"). Supports multiple plugins separated by |. Also supports all and enabled.
func (req SearchStartRequest) WithPlugins(plugins []string) *SearchStartRequest {
	req.Plugins = plugins
	return &req
}

// WithCategory sets categories to limit your search to (e.g. "legittorrents"). Available categories depend on the specified plugins. Also supports all.
func (req SearchStartRequest) WithCategory(category string) *SearchStartRequest {
	req.Category = category
	return &req
}

// SearchStartResponse is the search start response.
type SearchStartResponse struct {
	ID int64 `json:"id,omitempty" yaml:"id,omitempty"`
}

// SearchStopRequest is a search stop request.
type SearchStopRequest struct {
	ID int64 `json:"id,omitempty" yaml:"id,omitempty"`
}

// SearchStop creates a search stop request.
func SearchStop(id int64) *SearchStopRequest {
	return &SearchStopRequest{
		ID: id,
	}
}

// Do executes the request against the provided context and client.
func (req *SearchStopRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "search/stop", req, nil)
}

// SearchStatusRequest is a search status request.
type SearchStatusRequest struct {
	ID int64 `json:"id,omitempty" yaml:"id,omitempty"`
}

// SearchStatus creates a search status request.
func SearchStatus(id int64) *SearchStatusRequest {
	return &SearchStatusRequest{
		ID: id,
	}
}

// Do executes the request against the provided context and client.
func (req *SearchStatusRequest) Do(ctx context.Context, cl *Client) ([]SearchStatusInfo, error) {
	var res []SearchStatusInfo
	if err := cl.Do(ctx, "search/status", req, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// SearchStatusInfo is a search status.
type SearchStatusInfo struct {
	ID     int64  `json:"id,omitempty" yaml:"id,omitempty"`         // ID of the search job
	Status string `json:"status,omitempty" yaml:"status,omitempty"` // Current status of the search job (either Running or Stopped)
	Total  int64  `json:"total,omitempty" yaml:"total,omitempty"`   // Total number of results. If the status is Running this number may contineu to increase
}

// SearchResultsRequest is a search results request.
type SearchResultsRequest struct {
	ID     int64 `json:"id" yaml:"id"`                                               // ID of the search job
	Limit  int64 `json:"limit optional,omitempty" yaml:"limit optional,omitempty"`   // max number of results to return. 0 or negative means no limit
	Offset int64 `json:"offset optional,omitempty" yaml:"offset optional,omitempty"` // result to start at. A negative number means count backwards (e.g. -2 returns the 2 most recent results)
}

// SearchResults creates a search results request.
func SearchResults() *SearchResultsRequest {
	return &SearchResultsRequest{}
}

// Do executes the request against the provided context and client.
func (req *SearchResultsRequest) Do(ctx context.Context, cl *Client) (*SearchResultsResponse, error) {
	res := new(SearchResultsResponse)
	if err := cl.Do(ctx, "search/results", req, res); err != nil {
		return nil, err
	}
	return res, nil
}

// WithLimit sets max number of results to return. 0 or negative means no limit.
func (req SearchResultsRequest) WithLimit(limit int64) *SearchResultsRequest {
	req.Limit = limit
	return &req
}

// WithOffset sets result to start at. A negative number means count backwards (e.g. -2 returns the 2 most recent results).
func (req SearchResultsRequest) WithOffset(offset int64) *SearchResultsRequest {
	req.Offset = offset
	return &req
}

// SearchResultsResponse is the search results response.
type SearchResultsResponse struct {
	Results []SearchResult `json:"results,omitempty" yaml:"results,omitempty"` // Array of result objects- see table below
	Status  string         `json:"status,omitempty" yaml:"status,omitempty"`   // Current status of the search job (either Running or Stopped)
	Total   int64          `json:"total,omitempty" yaml:"total,omitempty"`     // Total number of results. If the status is Running this number may continue to increase
}

// SearchResult is a search result.
type SearchResult struct {
	DescrLink  string `json:"descrLink,omitempty" yaml:"descrLink,omitempty"`   // URL of the torrent's description page
	FileName   string `json:"fileName,omitempty" yaml:"fileName,omitempty"`     // Name of the file
	FileSize   int64  `json:"fileSize,omitempty" yaml:"fileSize,omitempty"`     // Size of the file in Bytes
	FileUrl    string `json:"fileUrl,omitempty" yaml:"fileUrl,omitempty"`       // Torrent download link (usually either .torrent file or magnet link)
	NbLeechers int64  `json:"nbLeechers,omitempty" yaml:"nbLeechers,omitempty"` // Number of leechers
	NbSeeders  int64  `json:"nbSeeders,omitempty" yaml:"nbSeeders,omitempty"`   // Number of seeders
	SiteUrl    string `json:"siteUrl,omitempty" yaml:"siteUrl,omitempty"`       // URL of the torrent site
}

// SearchDeleteRequest is a search delete request.
type SearchDeleteRequest struct {
	ID int64 `json:"id" yaml:"id"` // ID of the search job
}

// SearchDelete creates a search delete request.
func SearchDelete(id int64) *SearchDeleteRequest {
	return &SearchDeleteRequest{
		ID: id,
	}
}

// Do executes the request against the provided context and client.
func (req *SearchDeleteRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "search/delete", req, nil)
}

// SearchCategoriesRequest is a search categories request.
type SearchCategoriesRequest struct {
	PluginName string `json:"pluginName,omitempty" yaml:"pluginName,omitempty"` // name of the plugin (e.g. "legittorrents"). Also supports all and enabled
}

// SearchCategories creates a search categories request.
func SearchCategories() *SearchCategoriesRequest {
	return &SearchCategoriesRequest{}
}

// Do executes the request against the provided context and client.
func (req *SearchCategoriesRequest) Do(ctx context.Context, cl *Client) ([]string, error) {
	var res []string
	if err := cl.Do(ctx, "search/categories", req, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// WithPluginName sets name of the plugin (e.g. "legittorrents"). Also supports all and enabled.
func (req SearchCategoriesRequest) WithPluginName(pluginName string) *SearchCategoriesRequest {
	req.PluginName = pluginName
	return &req
}

// SearchPluginsRequest is a search plugins request.
type SearchPluginsRequest struct{}

// SearchPlugins creates a search plugins request.
func SearchPlugins() *SearchPluginsRequest {
	return &SearchPluginsRequest{}
}

// Do executes the request against the provided context and client.
func (req *SearchPluginsRequest) Do(ctx context.Context, cl *Client) ([]Plugin, error) {
	var res []Plugin
	if err := cl.Do(ctx, "search/plugins", req, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// Plugin holds information about a plugin.
type Plugin struct {
	Enabled             bool     `json:"enabled,omitempty" yaml:"enabled,omitempty"`                         // Whether the plugin is enabled
	FullName            string   `json:"fullName,omitempty" yaml:"fullName,omitempty"`                       // Full name of the plugin
	Name                string   `json:"name,omitempty" yaml:"name,omitempty"`                               // Short name of the plugin
	SupportedCategories []string `json:"supportedCategories,omitempty" yaml:"supportedCategories,omitempty"` // List of supported categories as strings
	URL                 string   `json:"url,omitempty" yaml:"url,omitempty"`                                 // URL of the torrent site
	Version             string   `json:"version,omitempty" yaml:"version,omitempty"`                         // Installed version of the plugin
}

// SearchInstallPluginRequest is a search installPlugin request.
type SearchInstallPluginRequest struct {
	Sources []string `json:"sources" yaml:"sources"` // Url or file path of the plugin to install (e.g. "https://raw.githubusercontent.com/qbittorrent/search-plugins/master/nova3/engines/legittorrents.py"). Supports multiple sources separated by |
}

// SearchInstallPlugin creates a search installPlugin request.
func SearchInstallPlugin(sources ...string) *SearchInstallPluginRequest {
	return &SearchInstallPluginRequest{
		Sources: sources,
	}
}

// Do executes the request against the provided context and client.
func (req *SearchInstallPluginRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "search/installPlugin", req, nil)
}

// SearchUninstallPluginRequest is a search uninstallPlugin request.
type SearchUninstallPluginRequest struct {
	Names []string `json:"names" yaml:"names"` // Name of the plugin to uninstall (e.g. "legittorrents"). Supports multiple names separated by |
}

// SearchUninstallPlugin creates a search uninstallPlugin request.
func SearchUninstallPlugin(names ...string) *SearchUninstallPluginRequest {
	return &SearchUninstallPluginRequest{
		Names: names,
	}
}

// Do executes the request against the provided context and client.
func (req *SearchUninstallPluginRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "search/uninstallPlugin", req, nil)
}

// SearchEnablePluginRequest is a search enablePlugin request.
type SearchEnablePluginRequest struct {
	Names  []string `json:"names" yaml:"names"`   // Name of the plugin to uninstall (e.g. "legittorrents"). Supports multiple names separated by |
	Enable bool     `json:"enable" yaml:"enable"` // Whether the plugins should be enabled
}

// SearchEnablePlugin creates a search enablePlugin request.
func SearchEnablePlugin(enable bool, names ...string) *SearchEnablePluginRequest {
	return &SearchEnablePluginRequest{
		Names:  names,
		Enable: enable,
	}
}

// Do executes the request against the provided context and client.
func (req *SearchEnablePluginRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "search/enablePlugin", req, nil)
}

// SearchUpdatePluginsRequest is a search updatePlugins request.
type SearchUpdatePluginsRequest struct{}

// SearchUpdatePlugins creates a search updatePlugins request.
func SearchUpdatePlugins() *SearchUpdatePluginsRequest {
	return &SearchUpdatePluginsRequest{}
}

// Do executes the request against the provided context and client.
func (req *SearchUpdatePluginsRequest) Do(ctx context.Context, cl *Client) error {
	return cl.Do(ctx, "search/updatePlugins", req, nil)
}
