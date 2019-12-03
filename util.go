package main

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/kenshaw/transrpc"
	"github.com/knq/snaker"
	"github.com/xo/tblfmt"
)

const (
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

	// ErrInvalidMatchOrder is the invalid match order error.
	ErrInvalidMatchOrder Error = "invalid match order"

	// ErrCannotListAllOptionsAndUnset is the cannot list all options and unset
	// error.
	ErrCannotListAllOptionsAndUnset Error = "cannot --list all options and --unset"

	// ErrCannotUnsetARemoteConfigOption is the cannot unset a remote config
	// option error.
	ErrCannotUnsetARemoteConfigOption Error = "cannot --unset a --remote config option"

	// ErrMustSpecifyConfigOptionNameToUnset is the must specify config option
	// name to unset error.
	ErrMustSpecifyConfigOptionNameToUnset Error = "must specify config option name to --unset"
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

// ConfigStore is the interface for config stores.
type ConfigStore interface {
	GetKey(string) string
	SetKey(string, string)
	RemoveKey(string)
	GetMapFlat() map[string]string
	Write(string) error
}

// RemoteConfigStore wraps the transrpc client.
type RemoteConfigStore struct {
	cl      *transrpc.Client
	session *transrpc.Session
	setKeys []string
}

// NewRemoteConfigStore creates a new remote config store.
func NewRemoteConfigStore(args *Args) (*RemoteConfigStore, error) {
	cl, err := args.newClient()
	if err != nil {
		return nil, err
	}

	session, err := cl.SessionGet(context.Background())
	if err != nil {
		return nil, err
	}

	return &RemoteConfigStore{cl: cl, session: session}, nil
}

// GetKey satisfies the ConfigStore interface.
func (r *RemoteConfigStore) GetKey(key string) string {
	return r.GetMapFlat()[key]
}

// SetKey satisfies the ConfigStore interface.
func (r *RemoteConfigStore) SetKey(key, value string) {
	r.setKeys = append(r.setKeys, key, value)
}

// RemoveKey satisfies the ConfigStore interface.
func (r *RemoteConfigStore) RemoveKey(key string) {
	panic("cannot remove remote session config option")
}

// GetMapFlat satisfies the ConfigStore interface.
func (r *RemoteConfigStore) GetMapFlat() map[string]string {
	m := make(map[string]string)
	addFieldsToMap(m, "", reflect.ValueOf(*r.session))
	return m
}

// addFieldsToMap adds reflected field values to the map.
func addFieldsToMap(m map[string]string, prefix string, v reflect.Value) {
	t := v.Type()
	count := t.NumField()
	for i := 0; i < count; i++ {
		tag := t.Field(i).Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		name := strings.SplitN(tag, ",", 2)[0]
		f := v.Field(i)
		switch f.Kind() {
		case reflect.String:
			m[prefix+name] = f.String()
		case reflect.Int64:
			m[prefix+name] = strconv.FormatInt(f.Int(), 10)
		case reflect.Float64:
			m[prefix+name] = fmt.Sprintf("%f", f.Float())
		case reflect.Bool:
			m[prefix+name] = strconv.FormatBool(f.Bool())
		case reflect.Struct:
			addFieldsToMap(m, name+".", f)
		case reflect.Slice:
			s, ok := f.Interface().([]string)
			if !ok {
				panic("not a []string")
			}
			m[prefix+name] = strings.Join(s, ",")

		default:
			panic(fmt.Sprintf("unknown type: %d // %v", i, f))
		}
	}
}

// Write satisfies the ConfigStore interface.
func (r *RemoteConfigStore) Write(string) error {
	req := transrpc.SessionSet()
	v := reflect.ValueOf(req)
	for i := 0; i < len(r.setKeys); i += 2 {
		name := "With" + snaker.ForceCamelIdentifier(r.setKeys[i])
		f := v.MethodByName(name)
		if f.Kind() == reflect.Invalid {
			return fmt.Errorf("unsupported setting --remote option %q", r.setKeys[i])
		}
		args := make([]reflect.Value, 1)
		switch f.Type().In(0).Kind() {
		case reflect.String:
			args[0] = reflect.ValueOf(r.setKeys[i+1])
		case reflect.Int64:
			z, err := strconv.ParseInt(r.setKeys[i+1], 10, 64)
			if err != nil {
				return err
			}
			args[0] = reflect.ValueOf(z)
		case reflect.Float64:
			z, err := strconv.ParseFloat(r.setKeys[i+1], 64)
			if err != nil {
				return err
			}
			args[0] = reflect.ValueOf(z)
		case reflect.Bool:
			b, err := strconv.ParseBool(r.setKeys[i+1])
			if err != nil {
				return err
			}
			args[0] = reflect.ValueOf(b)
		}
		req = f.Call(args)[0].Interface().(*transrpc.SessionSetRequest)
	}
	return req.Do(context.Background(), r.cl)
}
