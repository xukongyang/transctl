package delrpc

import (
	"encoding/base64"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/gdm85/go-rencode"
)

// Error is a transrpc error.
type Error string

// Error satisfies the error interface.
func (err Error) Error() string {
	return string(err)
}

// Error values.
const (
	// ErrMismatchedRequestAndResponseIDs is the mismatched request and response ids error.
	ErrMismatchedRequestAndResponseIDs Error = "mismatched request and response ids"
)

// appendParams
func appendParams(z []interface{}, v reflect.Value, depth int) []interface{} {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	var ret []interface{}
	switch v.Kind() {
	case reflect.Struct:
		d := new(rencode.Dictionary)
		for i := 0; i < v.NumField(); i++ {
			tag := strings.SplitN(v.Type().Field(i).Tag.Get("json"), ",", 2)
			if tag[0] == "" || tag[1] == "-" {
				continue
			}
			if depth == 0 {
				ret = appendParams(ret, v.Field(i), depth+1)
			} else {
				var omitempty bool
				if len(tag) > 1 {
					omitempty = contains(strings.Split(tag[1], ","), "omitempty")
				}
				f := v.Field(i)
				if !omitempty || !f.IsZero() {
					d.Add(tag[0], f.Interface())
				}
			}
		}
		if depth != 0 {
			ret = append(ret, d)
		}

	case reflect.Slice:
		switch x := v.Interface().(type) {
		case []interface{}:
			ret = append(ret, x...)
		case []string:
			for i := 0; i < len(x); i++ {
				ret = append(ret, x[i])
			}
		case []byte:
			ret = append(ret, base64.RawStdEncoding.EncodeToString(x))
		default:
			panic(fmt.Sprintf("unsupported slice type %T", v.Interface()))
		}

	case reflect.Map:
		d := new(rencode.Dictionary)
		iter := v.MapRange()
		for iter.Next() {
			d.Add(iter.Key().Interface(), iter.Value().Interface())
		}
		ret = append(ret, d)

	default:
		ret = append(ret, v.Interface())
	}
	return append(z, ret...)
}

// encode encodes a deluge rpc request to the writer.
func encode(w io.Writer, id int64, method string, v interface{}) error {
	params := rencode.NewList()
	params.Add(appendParams([]interface{}{id, method}, reflect.ValueOf(v), 0)...)
	l := rencode.NewList()
	l.Add(params)
	enc := rencode.NewEncoder(w)
	return enc.Encode(l)
}

// decode decodes a deluge rpc response.
func decode(buf []byte, v interface{}) (int64, error) {
	return 0, nil
}

// contains determines if needle is contained in haystack.
func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if needle == s {
			return true
		}
	}
	return false
}
