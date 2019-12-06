package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/kenshaw/transrpc"
	"github.com/knq/snaker"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
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

	// ErrInvalidOutputOptionSpecified is the invalid output option specified
	// error.
	ErrInvalidOutputOptionSpecified Error = "invalid --output option specified"
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
	// TODO: fix this and use tblfmt again
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
func (tr *TorrentResult) Encode(w io.Writer, args *Args, cl *transrpc.Client) error {
	var f func(io.Writer, *Args, *transrpc.Client) error
	switch args.Output {
	case "table":
		f = tr.encodeTable
	case "wide":
		f = tr.encodeWide
	case "json":
		f = tr.encodeJSON
	case "yaml":
		f = tr.encodeYaml
	case "flat":
		f = tr.encodeFlat
	default:
		return ErrInvalidOutputOptionSpecified
	}
	return f(w, args, cl)
}

// headerNames are column header names.
var headerNames = map[string]string{
	"RATE DOWNLOAD": "DOWN",
	"RATE UPLOAD":   "UP",
	"HAVE VALID":    "HAVE",
	"PERCENT DONE":  "%",
}

// encodeTableColumns encodes the specified table results with the included
// columns.
func (tr *TorrentResult) encodeTableColumns(w io.Writer, args *Args, cl *transrpc.Client, cols ...string) error {
	// check field names
	typ := reflect.TypeOf(transrpc.Torrent{})
	fields := make(map[string]int, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		tag := typ.Field(i).Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		fields[strings.SplitN(tag, ",", 2)[0]] = i
	}

	// build headers and field names
	headers, fieldnames := make([]string, len(cols)), make([]string, len(cols))
	for i, c := range cols {
		if c == "shorthash" {
			headers[i], fieldnames[i] = "HASH", "hashString"
			continue
		}
		n, ok := fields[c]
		if !ok {
			return fmt.Errorf("invalid torrent field %q", c)
		}
		headers[i] = strings.ReplaceAll(strings.ToUpper(snaker.CamelToSnakeIdentifier(typ.Field(n).Name)), "_", " ")
		if h, ok := headerNames[headers[i]]; ok {
			headers[i] = h
		}
		fieldnames[i] = c
	}

	// build base request
	var torrents []transrpc.Torrent
	if len(tr.torrents) != 0 {
		req := transrpc.TorrentGet(convTorrentIDs(tr.torrents)...).WithFields(fieldnames...)
		res, err := req.Do(context.Background(), cl)
		if err != nil {
			return err
		}
		torrents = res.Torrents
	}

	// tablewriter package is temporary until tblfmt is fixed
	tbl := tablewriter.NewWriter(w)
	tbl.SetHeader(headers)
	tbl.SetAutoWrapText(false)
	tbl.SetAutoFormatHeaders(true)
	tbl.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	tbl.SetAlignment(tablewriter.ALIGN_LEFT)
	tbl.SetCenterSeparator("")
	tbl.SetColumnSeparator("")
	tbl.SetRowSeparator("")
	tbl.SetHeaderLine(false)
	tbl.SetBorder(false)
	tbl.SetTablePadding("\t") // pad with tabs
	tbl.SetNoWhiteSpace(true)

	// add torrents
	for _, t := range torrents {
		row := make([]string, len(cols))
		for i := 0; i < len(cols); i++ {
			if cols[i] == "shorthash" {
				row[i] = t.HashString[:defaultShortHashLen]
				continue
			}

			v := reflect.ValueOf(t).Field(fields[cols[i]]).Interface()
			x, ok := v.(transrpc.ByteCount)
			if !ok {
				row[i] = fmt.Sprintf("%v", v)
				continue
			}

			suffix, prec := "", 2
			if headers[i] == "UP" || headers[i] == "DOWN" {
				suffix = "/s"
			}
			if args.Human == "true" || args.Human == "1" || args.SI {
				if args.SI && int64(x) < 1024*1024 || !args.SI && int64(x) < 1000*1000 {
					prec = 0
				}
				row[i] = x.Format(!args.SI, prec, suffix)
			} else {
				row[i] = fmt.Sprintf("%d%s", x, suffix)
			}
		}
		tbl.Append(row)
	}

	tbl.Render()
	return nil
}

// encodeTable encodes the torrent results to the writer as a table.
func (tr *TorrentResult) encodeTable(w io.Writer, args *Args, cl *transrpc.Client) error {
	return tr.encodeTableColumns(w, args, cl, "id", "name", "status", "eta", "rateDownload", "rateUpload", "haveValid", "percentDone")
}

// encodeWide encodes the torrent results to the writer as a table.
func (tr *TorrentResult) encodeWide(w io.Writer, args *Args, cl *transrpc.Client) error {
	return tr.encodeTableColumns(w, args, cl, "id", "name", "status", "eta", "rateDownload", "rateUpload", "haveValid", "percentDone")
}

// encodeJSON encodes the torrent results to the writer as a table.
func (tr *TorrentResult) encodeJSON(w io.Writer, args *Args, cl *transrpc.Client) error {
	res, err := cl.TorrentGet(context.Background(), convTorrentIDs(tr.torrents)...)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(res.Torrents)
}

// encodeYaml encodes the torrent results to the writer as a table.
func (tr *TorrentResult) encodeYaml(w io.Writer, args *Args, cl *transrpc.Client) error {
	res, err := cl.TorrentGet(context.Background(), convTorrentIDs(tr.torrents)...)
	if err != nil {
		return err
	}
	for _, t := range res.Torrents {
		fmt.Fprintln(w, "---")
		if err = yaml.NewEncoder(w).Encode(t); err != nil {
			return err
		}
	}
	return nil
}

// encodeFlat encodes the torrent results to the writer as a flat key map.
func (tr *TorrentResult) encodeFlat(w io.Writer, args *Args, cl *transrpc.Client) error {
	res, err := cl.TorrentGet(context.Background(), convTorrentIDs(tr.torrents)...)
	if err != nil {
		return err
	}
	for i, t := range res.Torrents {
		if i != 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "[torrent %s]\n", t.HashString[:defaultShortHashLen])
		m := make(map[string]string)
		addFieldsToMap(m, "", reflect.ValueOf(t))
		var keys []string
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if m[k] == "" {
				continue
			}
			fmt.Fprintf(w, "%s=%s\n", k, m[k])
		}
	}
	return nil
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
	GetAllFlat() []string
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

// GetAllFlat satisfies the ConfigStore interface.
func (r *RemoteConfigStore) GetAllFlat() []string {
	m := make(map[string]string)
	addFieldsToMap(m, "", reflect.ValueOf(*r.session))
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var ret []string
	for _, k := range keys {
		ret = append(ret, k, m[k])
	}
	return ret
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
		name := strings.ReplaceAll(snaker.CamelToSnakeIdentifier(strings.SplitN(tag, ",", 2)[0]), "_", "-")
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
			var s []string
			switch x := f.Interface().(type) {
			case []string:
				s = x
			case []byte:
				s = append(s, base64.StdEncoding.EncodeToString(x))
			case []int64:
				for _, v := range x {
					s = append(s, strconv.FormatInt(v, 10))
				}
			case []transrpc.Priority:
				for _, v := range x {
					s = append(s, fmt.Sprintf("%d", v))
				}
			case []transrpc.Bool:
				for _, v := range x {
					if bool(v) {
						s = append(s, "1")
					} else {
						s = append(s, "0")
					}
				}
			default:
				if reflect.TypeOf(x).Elem().Kind() != reflect.Struct {
					panic(fmt.Sprintf("unknown type for field %q", prefix+name))
				}
				for i := 0; i < f.Len(); i++ {
					z := make(map[string]string)
					addFieldsToMap(z, "", f.Index(i))
					var keys []string
					for k := range z {
						keys = append(keys, k)
					}
					sort.Strings(keys)
					var a string
					for i, k := range keys {
						if i != 0 {
							a += ","
						}
						a += fmt.Sprintf("%s:%s", strings.TrimSpace(k), strings.TrimSpace(z[k]))
					}
					s = append(s, "{"+a+"}")
				}
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
			return fmt.Errorf("unsupported setting --remote config option %q", r.setKeys[i])
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
