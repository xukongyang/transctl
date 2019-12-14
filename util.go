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

func init() {
	if err := snaker.AddInitialisms("UTP"); err != nil {
		panic(err)
	}
}

// Error is the error type.
type Error string

// Error satisfies the error interface.
func (err Error) Error() string {
	return string(err)
}

const (
	// ErrMustSpecifyListRecentFilterOrAtLeastOneTorrent is the must specify
	// list, recent, filter or at least one torrent error.
	ErrMustSpecifyListRecentFilterOrAtLeastOneTorrent Error = "must specify --list, --recent, --filter or at least one torrent"

	// ErrMustSpecifyListOrOptionName is the must specify list or option name
	// error.
	ErrMustSpecifyListOrOptionName Error = "must specify --list or option name"

	// ErrConfigFileCannotBeADirectory is the config file cannot be a directory
	// error.
	ErrConfigFileCannotBeADirectory Error = "config file cannot be a directory"

	// ErrMustSpecifyAtLeastOneLocation is the must specify at least one
	// location error.
	ErrMustSpecifyAtLeastOneLocation Error = "must specify at least one location"

	// ErrCannotSpecifyUnsetAndAlsoSetAnOptionValue is the cannot specify unset
	// and also set an option value error.
	ErrCannotSpecifyUnsetAndAlsoSetAnOptionValue Error = "cannot specify --unset and also set an option value"

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

// NewTorrentResult creates a new torrent result output encoder for the passed
// torrents.
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
	/*
	*(v[0].(*interface{})) = tr.torrents[tr.index].ID
	*(v[1].(*interface{})) = tr.torrents[tr.index].Name
	*(v[2].(*interface{})) = tr.torrents[tr.index].HashString[:defaultShortHashLen]
	 */
	tr.index++
	return nil
}

// Columns satisfies the tblfmt.ResultSet interface.
func (*TorrentResult) Columns() ([]string, error) {
	return []string{}, nil
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
	switch args.Output.Output {
	case "table":
		f = tr.encodeTable("id", "name", "status", "eta", "rateDownload", "rateUpload", "haveValid", "percentDone", "shorthash")
	case "wide":
		f = tr.encodeTable("id", "name", "status", "eta", "rateDownload", "rateUpload", "haveValid", "percentDone", "shorthash")
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
func (tr *TorrentResult) encodeTable(cols ...string) func(io.Writer, *Args, *transrpc.Client) error {
	return func(w io.Writer, args *Args, cl *transrpc.Client) error {
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
		var hasTotals bool
		headers, fieldnames, totals, display := make([]string, len(cols)), make([]string, len(cols)), make([]transrpc.ByteCount, len(cols)), make([]bool, len(cols))
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
		if !args.Output.NoHeaders {
			tbl.SetHeader(headers)
		}
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

				hasTotals = true
				totals[i] += x
				display[i] = true

				suffix, prec := "", 2
				if headers[i] == "UP" || headers[i] == "DOWN" {
					suffix = "/s"
				}
				if args.Output.Human == "true" || args.Output.Human == "1" || args.Output.SI {
					if args.Output.SI && int64(x) < 1024*1024 || !args.Output.SI && int64(x) < 1000*1000 {
						prec = 0
					}
					row[i] = x.Format(!args.Output.SI, prec, suffix)
				} else {
					row[i] = fmt.Sprintf("%d%s", x, suffix)
				}
			}
			tbl.Append(row)
		}

		if !args.Output.NoTotals && hasTotals && len(torrents) > 0 {
			row := make([]string, len(cols))
			for i := 0; i < len(totals); i++ {
				if !display[i] {
					continue
				}
				x := totals[i]
				suffix, prec := "", 2
				if headers[i] == "UP" || headers[i] == "DOWN" {
					suffix = "/s"
				}
				if args.Output.Human == "true" || args.Output.Human == "1" || args.Output.SI {
					if args.Output.SI && int64(x) < 1024*1024 || !args.Output.SI && int64(x) < 1000*1000 {
						prec = 0
					}
					row[i] = x.Format(!args.Output.SI, prec, suffix)
				} else {
					row[i] = fmt.Sprintf("%d%s", x, suffix)
				}
			}
			tbl.Append(row)
		}

		tbl.Render()
		return nil
	}
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

// Write satisfies the ConfigStore interface.
func (r *RemoteConfigStore) Write(string) error {
	return doWithAndExecute(r.cl, transrpc.SessionSet(), "--remote config", r.setKeys...)
}

// convTorrentIDs converts torrent list to a hash string identifier list.
func convTorrentIDs(torrents []transrpc.Torrent) []interface{} {
	ids := make([]interface{}, len(torrents))
	for i := 0; i < len(torrents); i++ {
		ids[i] = torrents[i].HashString
	}
	return ids
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

// executor interface is the common interface for settable requests.
type executor interface {
	Do(context.Context, *transrpc.Client) error
}

// doWithAndExecute calls the
func doWithAndExecute(cl *transrpc.Client, req executor, errMsg string, vals ...string) error {
	if len(vals)%2 != 0 {
		panic("invalid vals")
	}
	v := reflect.ValueOf(req)
	for i := 0; i < len(vals); i += 2 {
		name := "With" + snaker.ForceCamelIdentifier(vals[i])
		f := v.MethodByName(name)
		if f.Kind() == reflect.Invalid {
			return fmt.Errorf("unsupported setting %s option %q", errMsg, vals[i])
		}
		args := make([]reflect.Value, 1)
		switch f.Type().In(0).Kind() {
		case reflect.String:
			args[0] = reflect.ValueOf(vals[i+1])
		case reflect.Int64:
			z, err := strconv.ParseInt(vals[i+1], 10, 64)
			if err != nil {
				return err
			}
			args[0] = reflect.ValueOf(z)
		case reflect.Float64:
			z, err := strconv.ParseFloat(vals[i+1], 64)
			if err != nil {
				return err
			}
			args[0] = reflect.ValueOf(z)
		case reflect.Bool:
			b, err := strconv.ParseBool(vals[i+1])
			if err != nil {
				return err
			}
			args[0] = reflect.ValueOf(b)

		case reflect.Slice:
			// split values
			z := strings.Split(vals[i+1], ",")
			for j := range z {
				z[j] = strings.TrimSpace(z[j])
			}

			// make slice
			args[0] = reflect.Zero(f.Type().In(0))
			switch args[0].Interface().(type) {
			case []string:
				args[0] = reflect.ValueOf(z)
			case []int64:
				y := make([]int64, len(z))
				for a := range z {
					var err error
					y[a], err = strconv.ParseInt(z[a], 10, 64)
					if err != nil {
						return err
					}
				}
				args[0] = reflect.ValueOf(y)

			default:
				panic(fmt.Sprintf("unknown slice type %v", f.Type().In(0)))
			}

		default:
			panic(fmt.Sprintf("unknown type %v", f.Type().In(0)))
		}
		req = f.Call(args)[0].Interface().(executor)
	}
	return req.Do(context.Background(), cl)
}
