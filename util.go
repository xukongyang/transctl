package transrpc

import (
	"fmt"
	"regexp"
	"strings"
)

// Error is a transrpc error.
type Error string

// Error satisfies the error interface.
func (err Error) Error() string {
	return string(err)
}

// Error values.
const (
	// ErrInvalidTorrentHash is the invalid torrent hash error.
	ErrInvalidTorrentHash Error = "invalid torrent hash"

	// ErrInvalidIdentifierType is the invalid identifier type error.
	ErrInvalidIdentifierType Error = "invalid identifier type"

	// ErrUnauthorizedUser is the unauthorized user error.
	ErrUnauthorizedUser Error = "unauthorized user"

	// ErrUnknownProblemEncountered is the unknown problem encountered error.
	ErrUnknownProblemEncountered Error = "unknown problem encountered"

	// ErrRequestFailed is the request failed error.
	ErrRequestFailed Error = "request failed"

	// ErrRecentlyActiveCanHaveOnlyOneValue is the recently-active can have only one value error.
	ErrRecentlyActiveCanHaveOnlyOneValue Error = "recently-active can have only one value"

	// ErrInvalidPriority is the invalid priority error.
	ErrInvalidPriority Error = "invalid priority"

	// ErrInvalidTime is the invalid time error.
	ErrInvalidTime Error = "invalid time"

	// ErrInvalidBool is the invalid bool error.
	ErrInvalidBool Error = "invalid bool"

	// ErrInvalidEncryption is the invalid encryption error.
	ErrInvalidEncryption Error = "invalid encryption"

	// ErrTorrentRenamePathCanOnlyBeUsedWithOneTorrentIdentifier is the torrent
	// rename path can only be used with one torrent identifier error.
	ErrTorrentRenamePathCanOnlyBeUsedWithOneTorrentIdentifier Error = "torrent rename path can only be used with one torrent identifier"
)

const (
	// recentlyActive is the recently active identifier.
	recentlyActive = "recently-active"

	// csrfHeader is the CSRF header used for transmission rpc sessions.
	csrfHeader = "X-Transmission-Session-Id"
)

// sha1RE is a regexp to verify a SHA1 hash in string form.
var sha1RE = regexp.MustCompile(`(?i)^[0-9a-f]{40}$`)

// checkIdentifierList processes the passed ids and verifies they are of the
// right type. Allowed identifier types are: int{,8,16,32,64}, string,
// [40]byte, and []byte.
//
// Collapses the passed list of IDs into the accepted "3.1" type defined in the
// transmission rpc spec.
//
// As per the spec identifiers will be changed to:
//
// 		torrent IDs => int64
// 		bytes (hashes) => 40 character hexadecimal string
//      strings (hashes) => 40 character hexadecimal string
// 		"recently-active" string => passed through
//
// A id list with only one value will be flattened. If the identifier is
// "recently-active", then it can be the only identifier in the list.
func checkIdentifierList(ids ...interface{}) (interface{}, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	var v []interface{}
	for _, id := range ids {
		switch x := id.(type) {
		case int:
			v = append(v, int64(x))
		case int8:
			v = append(v, int64(x))
		case int16:
			v = append(v, int64(x))
		case int32:
			v = append(v, int64(x))
		case int64:
			v = append(v, x)

		case [40]byte:
			v = append(v, fmt.Sprintf("%x", x))

		case []byte:
			// convert
			if len(x) != 40 {
				return nil, ErrInvalidTorrentHash
			}
			v = append(v, fmt.Sprintf("%x", x))

		case string:
			// check if "recently-active", if so then no other ids may be
			// present. also check that a string is a valid sha1 hash
			switch {
			case x == recentlyActive && len(ids) != 1:
				return nil, ErrRecentlyActiveCanHaveOnlyOneValue
			case x != recentlyActive && !sha1RE.MatchString(x):
				return nil, ErrInvalidTorrentHash
			}
			v = append(v, strings.ToLower(x))

		default:
			return nil, ErrInvalidIdentifierType
		}
	}

	if len(v) == 1 {
		switch x := v[0].(type) {
		case int64:
			return x, nil
		case string:
			if x == recentlyActive {
				return x, nil
			}
		}
	}

	return v, nil
}
