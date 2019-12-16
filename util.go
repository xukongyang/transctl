package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/kenshaw/transrpc"
	"github.com/knq/snaker"
)

const (
	defaultConfig = `[default]
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

	// ErrSortByNotInColumnList is the sort by not in column list error.
	ErrSortByNotInColumnList Error = "--sort-by not in column list"

	// ErrMustSpecifyAtLeastOneOutputColumn is the must specify at least one output column error.
	ErrMustSpecifyAtLeastOneOutputColumn Error = "must specify at least one output column"

	// ErrFilterMustReturnBool is the filter must return bool error.
	ErrFilterMustReturnBool Error = "filter must return bool"

	// ErrInvalidStrlenArguments is the invalid strlen arguments error.
	ErrInvalidStrlenArguments Error = "invalid strlen() arguments"
)

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

// doWithAndExecute calls the 'With*' method on the reflected request for the
// provided name, value pairs in vals.
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
