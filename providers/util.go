package providers

import (
	"context"
	"encoding/base64"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/kenshaw/transctl/transrpc"
	"github.com/knq/snaker"
)

const (
	defaultConfig = `[default]
	output=table
`
)

// ConvertTorrentIDs converts torrent list to a hash string identifier list.
func ConvertTorrentIDs(torrents []transrpc.Torrent) []interface{} {
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
		for f.Kind() == reflect.Interface {
			f = f.Elem()
		}
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
			panic(fmt.Sprintf("unknown type: %d // %v", i, f.Type()))
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
