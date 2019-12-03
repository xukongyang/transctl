package main

import (
	"io"

	"github.com/kenshaw/transrpc"
	"github.com/xo/tblfmt"
)

const (
	defaultHost           = `localhost:9091`
	defaultShortHashLen   = 7
	minimumHashCompareLen = 5
	defaultConfig         = `[default]
	output=table
`
)

// Error is the error type.
type Error string

// Error satisfies the error interface.
func (err Error) Error() string {
	return string(err)
}

const (
	// ErrMustSpecifyAllRecentOrAtLeastOneTorrent is the must specify all,
	// recent or at least one torrent error.
	ErrMustSpecifyAllRecentOrAtLeastOneTorrent Error = "must specify --all, --recent or at least one torrent"

	// ErrConfigFileCannotBeADirectory is the config file cannot be a directory
	// error.
	ErrConfigFileCannotBeADirectory Error = "config file cannot be a directory"

	// ErrMustSpecifyAtLeastOneTorrentOrURI is the must specify at least one
	// torrent or uri error.
	ErrMustSpecifyAtLeastOneTorrentOrURI Error = "must specify at least one torrent or URI"

	// ErrCannotSpecifyUnsetWhileTryingToSetAValueWithConfig is the cannot
	// specify unsite while trying to set a value with config error.
	ErrCannotSpecifyUnsetWhileTryingToSetAValueWithConfig Error = "cannot specify --unset while trying to set a value with config"

	// ErrInvalidProtoHostOrRpcPath is the invalid proto, host, or rpc-path
	// error.
	ErrInvalidProtoHostOrRpcPath Error = "invalid --proto, --host, or --rpc-path"

	// ErrMustSpecifyURLHostOrConfigureTheContextAndContextURL is the must
	// specify url, host, or configure the context and context.url error.
	ErrMustSpecifyURLHostOrConfigureTheContextAndContextURL Error = "must specify --url, --host, or configure the context and context.url"

	// ErrInvalidMatchOrder is the invalid match order error.
	ErrInvalidMatchOrder Error = "invalid match order"
)

// TorrentResult is a wrapper type for slice of *transrpc.Torrent's that
// satisfies the tblfmt.ResultSet interface.
type TorrentResult struct {
	torrents []transrpc.Torrent
	index    int
}

func NewTorrentResult(torrents []transrpc.Torrent) *TorrentResult {
	return &TorrentResult{
		torrents: torrents,
	}
}

// Next satisfies the tblfmt.ResultSet interface.
func (tr *TorrentResult) Next() bool {
	return tr.index < len(tr.torrents)
}

// Scan satisfies the tblfmt.ResultSet interface.
func (tr *TorrentResult) Scan(v ...interface{}) error {
	*(v[0].(*interface{})) = tr.torrents[tr.index].ID
	*(v[1].(*interface{})) = tr.torrents[tr.index].Name
	*(v[2].(*interface{})) = tr.torrents[tr.index].HashString[:defaultShortHashLen]
	tr.index++
	return nil
}

// Columns satisfies the tblfmt.ResultSet interface.
func (*TorrentResult) Columns() ([]string, error) {
	return []string{
		"ID",
		"Name",
		"Short",
	}, nil
}

// Close satisfies the tblfmt.ResultSet interface.
func (*TorrentResult) Close() error {
	return nil
}

// Err satisfies the tblfmt.ResultSet interface.
func (*TorrentResult) Err() error {
	return nil
}

// NextResultSet satisfies the tblfmt.ResultSet interface.
func (*TorrentResult) NextResultSet() bool {
	return false
}

// Encode encodes the torrent result using the settings in args to the
// io.Writer.
func (tr *TorrentResult) Encode(w io.Writer, args *Args) error {
	return tblfmt.EncodeTable(w, tr)
}

// convTorrentIDs converts torrent list to a hash string identifier list.
func convTorrentIDs(torrents []transrpc.Torrent) []interface{} {
	ids := make([]interface{}, len(torrents))
	for i := 0; i < len(torrents); i++ {
		ids[i] = torrents[i].HashString
	}
	return ids
}
