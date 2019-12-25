package qbtweb

import (
	"fmt"
	"io"
	"net/url"
	"reflect"
	"regexp"
	"strings"
)

// Error is an error.
type Error string

// Error satisfies the error interface.
func (err Error) Error() string {
	return string(err)
}

// Error values.
const (
	// ErrUnauthorizedUser is the unauthorized user error.
	ErrUnauthorizedUser Error = "unauthorized user"

	// ErrTorrentNotFound is the torrent not found error.
	ErrTorrentNotFound Error = "torrent not found"

	// ErrTorrentFileInvalid is the torrent file invalid error.
	ErrTorrentFileInvalid Error = "torrent file invalid"

	// ErrRequestFailed is the request failed error.
	ErrRequestFailed Error = "request failed"
)

// sha1RE is a regexp to verify a SHA1 hash in string form.
var sha1RE = regexp.MustCompile(`(?i)^[0-9a-f]{40}$`)

// buildParamMap converts z into a map[string]interface{}.
func buildParamMap(z interface{}) (map[string]interface{}, error) {
	if m, ok := z.(map[string]interface{}); ok {
		return m, nil
	}

	// ensure z is *struct
	v := reflect.ValueOf(z)
	if reflect.Ptr != v.Kind() {
		return nil, fmt.Errorf("expected pointer to struct, got: %T", z)
	}
	v = v.Elem()
	if reflect.Struct != v.Kind() {
		return nil, fmt.Errorf("expected pointer to struct, got: %T", z)
	}

	// build params
	typ := v.Type()
	params := make(map[string]interface{})
	for i := 0; i < v.NumField(); i++ {
		f := typ.Field(i)
		tag := strings.SplitN(f.Tag.Get("json"), ",", 2)
		if tag[0] == "" || tag[0] == "-" {
			continue
		}
		var omitempty bool
		if len(tag) > 1 {
			omitempty = contains(strings.Split(tag[1], ","), "omitempty")
		}
		if reflect.Slice == f.Type.Kind() && reflect.String == f.Type.Elem().Kind() {
			s := strings.Join(v.Field(i).Interface().([]string), "|")
			if !omitempty || s != "" {
				params[tag[0]] = s
			}
		} else {
			x := v.Field(i)
			if !omitempty || !x.IsZero() {
				params[tag[0]] = x.Interface()
			}
		}
	}
	return params, nil
}

// buildRequestBody builds the request body for the passed params.
func buildRequestBody(w io.Writer, params map[string]interface{}) (string, error) {
	x := make(url.Values)
	for k, v := range params {
		x.Add(k, fmt.Sprintf("%v", v))
	}
	if _, err := w.Write([]byte(x.Encode())); err != nil {
		return "", err
	}
	return "application/x-www-form-urlencoded", nil
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
