package main

import (
	"github.com/kenshaw/transrpc"
)

const (
	defaultURL = `http://transmission:transmission@localhost:9091/transmission/rpc/`

	defaultConfig = `[default]
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
	// ErrMustSpecifyAllOrAtLeastOneTorrent is the must specify all or at least
	// one torrent error.
	ErrMustSpecifyAllOrAtLeastOneTorrent Error = "must specify --all or at least one torrent"

	// ErrConfigFileCannotBeADirectory is the config file cannot be a directory
	// error.
	ErrConfigFileCannotBeADirectory Error = "config file cannot be a directory"

	// ErrMustSpecifyAtLeastOneTorrentOrURI is the must specify at least one
	// torrent or uri error.
	ErrMustSpecifyAtLeastOneTorrentOrURI Error = "must specify at least one torrent or URI"

	// ErrCannotSpecifyUnsetWhileTryingToSetAValueWithConfig is the cannot
	// specify unsite while trying to set a value with config error.
	ErrCannotSpecifyUnsetWhileTryingToSetAValueWithConfig Error = "cannot specify --unset while trying to set a value with config"
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
	*(v[2].(*interface{})) = tr.torrents[tr.index].HashString[:8]
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
