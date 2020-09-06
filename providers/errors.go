package providers

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
