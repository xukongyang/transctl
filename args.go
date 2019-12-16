package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/jdxcode/netrc"
	"github.com/kenshaw/transrpc"
	"github.com/knq/ini"
)

// Args holds command args.
type Args struct {
	// ConfigFile is the global config file.
	ConfigFile string

	// Context is the global context name.
	Context string

	// Verbose is the global verbose toggle.
	Verbose bool

	// Host contains the global host configuration.
	Host struct {
		// URL is the URL to work with.
		URL *url.URL

		// Proto is the proto to use for building URLs.
		Proto string

		// Host is the host and port to use for building URLs.
		Host string

		// RpcPath is the rpc path to use for building URLs.
		RpcPath string

		// Credentials is the user:pass credentials to work with.
		Credentials string

		// CredentialsWasSet is the credentials was set toggle.
		CredentialsWasSet bool

		// Timeout is the rpc host request timeout.
		Timeout time.Duration

		// NoNetrc toggles disabling .netrc loading.
		NoNetrc bool

		// NetRCFile is the NetRCFile to use.
		NetrcFile string
	}

	// Filter contains the global filter configuration.
	Filter struct {
		// ListAll is the all toggle.
		ListAll bool

		// Recent is the recent toggle.
		Recent bool

		// Filter is the torrent filter.
		Filter string

		// FilterWasSet is the filter set toggle.
		FilterWasSet bool
	}

	// Output contains the global output configuration.
	Output struct {
		// Output is the output format type.
		Output string

		// OutputWasSet is the output was set toggle.
		OutputWasSet bool

		// Human is the toggle to display sizes in powers of 1024 (ie, 1023MiB).
		Human string

		// SI is the toggle to display sizes in powers of 1000 (ie, 1.1GB).
		SI bool

		// SIWasSet is the si was set toggle.
		SIWasSet bool

		// NoHeaders is the toggle to disable headers on table output.
		NoHeaders bool

		// NoTotals is the toggle to disable the total line on table output.
		NoTotals bool

		// ColumnNames is the column name map.
		ColumnNames map[string]string

		// SortBy is the column to sort by.
		SortBy string

		// SortByWasSet is the sort by set toggle.
		SortByWasSet bool

		// SortOrder ist he sort by order.
		SortOrder string
	}

	// ConfigParams are the config params.
	ConfigParams struct {
		Remote bool
		Name   string
		Value  string
		Unset  bool
	}

	// AddParams are the add params.
	AddParams struct {
		Cookies           map[string]string
		DownloadDir       string
		Paused            bool
		PeerLimit         int64
		BandwidthPriority int64
		Remove            bool
		RemoveWasSet      bool
	}

	// StartParams are the start params.
	StartParams struct {
		Now bool
	}

	// MoveParams are the move params.
	MoveParams struct {
		Dest string
	}

	// RemoveParams are the remove params.
	RemoveParams struct {
		Remove bool
	}

	// FilesSetPriorityParams are the files set-priority params.
	FilesSetPriorityParams struct {
		Priority string
	}

	// TrackersReplacePramas are the trackers replace params.
	TrackersReplaceParams struct {
		Replace string
	}

	// FileMask is the file mask for files operations.
	FileMask string

	// Tracker is the tracker for trackers operations.
	Tracker string

	// Args are remaining global arguments.
	Args []string

	// Config is the loaded settings from the config file.
	Config *ini.File
}

var tableCols = []string{"rateDownload=down", "rateUpload=up", "haveValid=have", "percentDone=done", "shortHash=hash", "addedDate=added", "downloadDir=location", "peersConnected=peers"}

// NewArgs creates the command args.
func NewArgs() (*Args, string, error) {
	// determine netrc path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, "", err
	}
	netrcFile := filepath.Join(homeDir, ".netrc")
	if runtime.GOOS == "windows" {
		netrcFile = filepath.Join(homeDir, "_netrc")
	}

	// determine config file path
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, "", err
	}
	configFile := filepath.Join(configDir, "transctl", "config.ini")

	// kingpin settings
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate)

	// create args
	args := &Args{}
	args.AddParams.Cookies = make(map[string]string)
	args.Output.ColumnNames = make(map[string]string)

	// global options
	kingpin.Flag("verbose", "toggle verbose").Short('v').Default("false").BoolVar(&args.Verbose)
	kingpin.Flag("config", "config file").Short('C').Default(configFile).Envar("TRANSCONFIG").PlaceHolder("<file>").StringVar(&args.ConfigFile)
	kingpin.Flag("context", "config context").Short('c').Envar("TRANSCONTEXT").PlaceHolder("<context>").StringVar(&args.Context)
	kingpin.Flag("url", "remote host url").Short('U').Envar("TRANSURL").PlaceHolder("<url>").URLVar(&args.Host.URL)
	kingpin.Flag("proto", "protocol to use").Default("http").PlaceHolder("http").StringVar(&args.Host.Proto)
	kingpin.Flag("host", "remote host").Short('h').PlaceHolder("localhost:9091").StringVar(&args.Host.Host)
	kingpin.Flag("rpc-path", "rpc path").Default("/transmission/rpc/").PlaceHolder("<path>").StringVar(&args.Host.RpcPath)
	kingpin.Flag("user", "remote host username and password").Short('u').PlaceHolder("<user:pass>").IsSetByUser(&args.Host.CredentialsWasSet).StringVar(&args.Host.Credentials)
	kingpin.Flag("no-netrc", "disable netrc loading").BoolVar(&args.Host.NoNetrc)
	kingpin.Flag("netrc-file", "netrc file path").Default(netrcFile).PlaceHolder("<file>").StringVar(&args.Host.NetrcFile)
	kingpin.Flag("timeout", "rpc request timeout (default: 25s)").Default("25s").PlaceHolder("<dur>").DurationVar(&args.Host.Timeout)

	// config command
	configCmd := kingpin.Command("config", "Get and set local and remote config")
	configCmd.Flag("remote", "get or set remote config option").BoolVar(&args.ConfigParams.Remote)
	configCmd.Flag("list", "list all options").Short('l').BoolVar(&args.Filter.ListAll)
	configCmd.Flag("all", "list all options").Hidden().BoolVar(&args.Filter.ListAll)
	configCmd.Flag("unset", "unset option").BoolVar(&args.ConfigParams.Unset)
	configCmd.Arg("name", "option name").StringVar(&args.ConfigParams.Name)
	configCmd.Arg("value", "option value").StringVar(&args.ConfigParams.Value)

	// add command
	addCmd := kingpin.Command("add", "Add torrents")
	args.addOutputFlags(addCmd, "id", tableCols...)
	addCmd.Flag("bandwidth-priority", "bandwidth priority").Short('b').PlaceHolder("<bw>").Int64Var(&args.AddParams.BandwidthPriority)
	addCmd.Flag("cookies", "cookies").Short('k').PlaceHolder("<name>=<v>").StringMapVar(&args.AddParams.Cookies)
	addCmd.Flag("download-dir", "download directory").Short('d').PlaceHolder("<dir>").StringVar(&args.AddParams.DownloadDir)
	addCmd.Flag("paused", "start torrent paused").Short('P').BoolVar(&args.AddParams.Paused)
	addCmd.Flag("peer-limit", "peer limit").Short('L').PlaceHolder("<limit>").Int64Var(&args.AddParams.PeerLimit)
	addCmd.Flag("rm", "remove torrents after adding").IsSetByUser(&args.AddParams.RemoveWasSet).BoolVar(&args.AddParams.Remove)
	addCmd.Arg("torrents", "torrent file or URL").Required().StringsVar(&args.Args)

	// add retrieval/manipulation commands
	commands := []string{
		"get", "Get information about torrents",
		"set", "Set torrent config options",
		"start", "Start torrents",
		"stop", "Stop torrents",
		"move", "Move torrent location",
		"remove", "Remove torrents",
		"verify", "Verify torrents",
		"reannounce", "Reannounce torrents",
		"queue top", "Move torrents to top of queue",
		"queue bottom", "Move torrents to bottom of queue",
		"queue up", "Move torrents up in queue",
		"queue down", "Move torrents down in queue",
		"peers get", "Get information about peers",
		"files get", "Get information about files",
		"files set-priority", "Set torrent files' priority",
		"files set-wanted", "Set torrent files as wanted",
		"files set-unwanted", "Set torrent files as unwanted",
		"trackers get", "Get information about trackers",
		"trackers add", "Add tracker to torrents",
		"trackers replace", "Replace tracker for torrents",
		"trackers remove", "Remove tracker from torrents",
	}

	cmds := map[string]*kingpin.CmdClause{
		"queue":    kingpin.Command("queue", "Change torrent queue position"),
		"peers":    kingpin.Command("peers", "Retrieve information about peers"),
		"files":    kingpin.Command("files", "Change priority and location of torrent files"),
		"trackers": kingpin.Command("trackers", "Change torrent trackers"),
	}
	for i := 0; i < len(commands); i += 2 {
		f := kingpin.Command
		cmdName := commands[i]
		if s := strings.SplitN(commands[i], " ", 2); len(s) > 1 {
			f, cmdName = cmds[s[0]].Command, s[1]
		}

		// add command
		cmd := f(cmdName, commands[i+1])
		cmd.Flag("list", "list all torrents").Short('l').BoolVar(&args.Filter.ListAll)
		cmd.Flag("all", "list all torrents").Hidden().BoolVar(&args.Filter.ListAll)
		cmd.Flag("recent", "recently active torrents").Short('R').BoolVar(&args.Filter.Recent)
		cmd.Flag("active", "recently active torrents").Hidden().BoolVar(&args.Filter.Recent)
		cmd.Flag("filter", "torrent filter").Short('F').PlaceHolder("<filter>").Default(
			`id == identifier || name %% identifier || (strlen(identifier) >= 5 && hashString %^ identifier)`,
		).IsSetByUser(&args.Filter.FilterWasSet).StringVar(&args.Filter.Filter)

		switch commands[i] {
		case "get":
			args.addOutputFlags(cmd, "id", tableCols...)

		case "set":
			cmd.Arg("name", "option name").Required().StringVar(&args.ConfigParams.Name)
			cmd.Arg("value", "option value").Required().StringVar(&args.ConfigParams.Value)

		case "start":
			cmd.Flag("now", "start now").BoolVar(&args.StartParams.Now)

		case "move":
			cmd.Flag("dest", "move destination").Short('d').PlaceHolder("<dir>").StringVar(&args.MoveParams.Dest)

		case "remove":
			cmd.Flag("rm", "remove downloaded files").BoolVar(&args.RemoveParams.Remove)

		case "peers get":
			args.addOutputFlags(cmd, "address", "clientName=client", "rateToClient=down", "rateToPeer=up", "progress=%", "shortHash=hash")

		case "files get":
			args.addOutputFlags(cmd, "name")

		case "files set-priority":
			cmd.Arg("file mask", "file mask").Required().StringVar(&args.FileMask)
			cmd.Arg("priority", "file priority (low, normal, high)").Required().EnumVar(&args.FilesSetPriorityParams.Priority, "low", "normal", "high")

		case "files set-wanted", "files set-unwanted":
			cmd.Arg("file mask", "file mask").Required().StringVar(&args.FileMask)

		case "trackers get":
			args.addOutputFlags(cmd, "announce")

		case "trackers add", "trackers remove":
			cmd.Arg("tracker", "tracker url").Required().StringVar(&args.Tracker)

		case "trackers replace":
			cmd.Arg("tracker", "tracker url").Required().StringVar(&args.Tracker)
			cmd.Arg("replace", "replace url").Required().StringVar(&args.TrackersReplaceParams.Replace)
		}

		cmd.Arg("torrents", "torrent id, name, or hash").StringsVar(&args.Args)
	}

	// stats command
	_ = kingpin.Command("stats", "Get session statistics")

	// shutdown command
	_ = kingpin.Command("shutdown", "Shutdown remote host")

	// free-space command
	freeSpaceCmd := kingpin.Command("free-space", "Retrieve free space")
	freeSpaceCmd.Flag("human", "print sizes in powers of 1024 (e.g., 1023MiB) (default: true)").Default("true").StringVar(&args.Output.Human)
	freeSpaceCmd.Flag("si", "print sizes in powers of 1000 (e.g., 1.1GB)").BoolVar(&args.Output.SI)
	freeSpaceCmd.Arg("location", "location").StringsVar(&args.Args)

	// blocklist-update command
	_ = kingpin.Command("blocklist-update", "Update blocklist")

	// port-test command
	_ = kingpin.Command("port-test", "Check if external port is open")

	// add --version flag
	kingpin.Flag("version", "display version and exit").PreAction(func(*kingpin.ParseContext) error {
		fmt.Fprintln(os.Stdout, "transctl", version)
		os.Exit(0)
		return nil
	}).Bool()

	cmd := kingpin.Parse()

	// load config
	if err = args.loadConfig(cmd); err != nil {
		return nil, "", err
	}

	return args, cmd, nil
}

// addOutputFlags adds output flags to the cmd.
func (args *Args) addOutputFlags(cmd *kingpin.CmdClause, sortBy string, columnNames ...string) {
	cmd.Flag("output", "output format (table, wide, json, yaml, flat; default: table)").Short('o').PlaceHolder("<format>").IsSetByUser(&args.Output.OutputWasSet).StringVar(&args.Output.Output)
	cmd.Flag("human", "print sizes in powers of 1024 (e.g., 1023MiB) (default: true)").Default("true").PlaceHolder("true").StringVar(&args.Output.Human)
	cmd.Flag("si", "print sizes in powers of 1000 (e.g., 1.1GB)").IsSetByUser(&args.Output.SIWasSet).BoolVar(&args.Output.SI)
	cmd.Flag("no-headers", "disable table header output").BoolVar(&args.Output.NoHeaders)
	cmd.Flag("no-totals", "disable table total output").BoolVar(&args.Output.NoTotals)
	cmd.Flag("column-name", "change output column name").PlaceHolder("<k=v>").Default(columnNames...).StringMapVar(&args.Output.ColumnNames)
	cmd.Flag("sort-by", "sort output order by column").PlaceHolder("<sort>").Default(sortBy).IsSetByUser(&args.Output.SortByWasSet).StringVar(&args.Output.SortBy)
	cmd.Flag("order-by", "sort output order by column").Hidden().PlaceHolder("<sort>").IsSetByUser(&args.Output.SortByWasSet).StringVar(&args.Output.SortBy)
	cmd.Flag("by", "sort output order by column").Hidden().PlaceHolder("<sort>").IsSetByUser(&args.Output.SortByWasSet).StringVar(&args.Output.SortBy)
	cmd.Flag("sort-order", "sort output order (asc, desc; default: asc)").PlaceHolder("<order>").Default("asc").EnumVar(&args.Output.SortOrder, "asc", "desc")
	cmd.Flag("order", "sort output order (asc, desc; default: asc)").Hidden().PlaceHolder("<order>").EnumVar(&args.Output.SortOrder, "asc", "desc")
}

// loadConfig loads the configuration file from disk.
func (args *Args) loadConfig(cmd string) error {
	// check if config file exists, create if not
	fi, err := os.Stat(args.ConfigFile)
	switch {
	case err == nil && fi.IsDir():
		return ErrConfigFileCannotBeADirectory
	case err != nil && os.IsNotExist(err):
		if err = args.createConfig(); err != nil {
			return err
		}
	case err != nil:
		return err
	}

	// load config
	args.Config, err = ini.LoadFile(args.ConfigFile)
	if err != nil {
		return err
	}
	args.Config.SectionManipFunc, args.Config.SectionNameFunc = ini.GitSectionManipFunc, ini.GitSectionNameFunc

	// change flags from config file, if not set by command line flags
	if v := strings.ToLower(strings.TrimSpace(args.Config.GetKey("default.output"))); v != "" && !args.Output.OutputWasSet {
		args.Output.Output = v
	}
	if v := strings.ToLower(strings.TrimSpace(args.Config.GetKey("default.si"))); v != "" && !args.Output.SIWasSet {
		args.Output.SI = v == "true" || v == "1"
	}
	if v := strings.ToLower(strings.TrimSpace(args.Config.GetKey("command.add.rm"))); v != "" && !args.AddParams.RemoveWasSet {
		args.AddParams.Remove = v == "true" || v == "1"
	}
	if v := strings.TrimSpace(args.getContextKey("free-space")); cmd == "free-space" && v != "" && len(args.Args) == 0 {
		args.Args = strings.Split(v, ",")
		for i := 0; i < len(args.Args); i++ {
			args.Args[i] = strings.TrimSpace(args.Args[i])
		}
	}

	// check specific command flags
	switch cmd {
	// check that either a name was passed, or that --all was specified
	case "config":
		switch {
		case args.Filter.ListAll && args.ConfigParams.Unset:
			return ErrCannotListAllOptionsAndUnset
		case args.ConfigParams.Remote && args.ConfigParams.Unset:
			return ErrCannotUnsetARemoteConfigOption
		case args.ConfigParams.Unset && args.ConfigParams.Name == "":
			return ErrMustSpecifyConfigOptionNameToUnset
		case args.ConfigParams.Unset && args.ConfigParams.Value != "":
			return ErrCannotSpecifyUnsetAndAlsoSetAnOptionValue
		case !args.Filter.ListAll && args.ConfigParams.Name == "":
			return ErrMustSpecifyListOrOptionName
		}

	// check exactly one of --list, --recent, --filter, or len(args.Args) > 0 conditions
	case "get", "set", "start", "stop", "move", "remove", "verify", "reannounce",
		"peers get", "files get", "files set-priority", "files set-wanted", "files set-unwanted",
		"trackers get", "trackers add", "trackers replace", "trackers remove",
		"queue top", "queue bottom", "queue up", "queue down":
		switch {
		case args.Filter.ListAll && args.Filter.Recent,
			args.Filter.ListAll && len(args.Args) != 0,
			args.Filter.Recent && len(args.Args) != 0,
			args.Filter.ListAll && args.Filter.FilterWasSet,
			args.Filter.Recent && args.Filter.FilterWasSet,
			!args.Filter.ListAll && !args.Filter.Recent && !args.Filter.FilterWasSet && len(args.Args) == 0:
			return ErrMustSpecifyListRecentFilterOrAtLeastOneTorrent
		}

	// check that either a location was passed as an argument, or specified via
	// config context options
	case "free-space":
		if len(args.Args) == 0 {
			return ErrMustSpecifyAtLeastOneLocation
		}
	}

	return nil
}

// createConfig creates the configuration file.
func (args *Args) createConfig() error {
	var err error
	if err = os.MkdirAll(filepath.Dir(args.ConfigFile), 0700); err != nil {
		return err
	}

	// stat file, create if not present
	_, err = os.Stat(args.ConfigFile)
	switch {
	case err != nil && os.IsNotExist(err):
		if err = ioutil.WriteFile(args.ConfigFile, []byte(defaultConfig), 0600); err != nil {
			return err
		}
	case err != nil:
		return err
	}

	return nil
}

// getContextKey returns the current context's name value from the config, or the default value.
func (args *Args) getContextKey(name string) string {
	context := args.Context
	if context == "" {
		context = args.Config.GetKey("default.context")
	}
	if context != "" {
		if v := args.Config.GetKey("context." + context + "." + name); v != "" {
			return v
		}
	}
	return args.Config.GetKey("default." + name)
}

// newClient builds a transrpc client for use by other commands.
func (args *Args) newClient() (*transrpc.Client, error) {
	var err error

	// choose specified url first
	u := args.Host.URL

	// check if host is specified
	if u == nil && args.Host.Host != "" {
		u, err = url.Parse(args.Host.Proto + "://" + args.Host.Host + args.Host.RpcPath)
		if err != nil {
			return nil, ErrInvalidProtoHostOrRpcPath
		}
	}

	// get context based url
	if urlstr := args.getContextKey("url"); u == nil && urlstr != "" {
		u, err = url.Parse(urlstr)
		if err != nil {
			return nil, err
		}
	}

	// default host
	if u == nil {
		host := args.Host.Host
		if host == "" {
			host = "localhost:9091"
		}
		u, err = url.Parse(args.Host.Proto + "://" + host + args.Host.RpcPath)
		if err != nil {
			return nil, err
		}
	}

	// add credentials
	if u.User == nil && args.Host.CredentialsWasSet && args.Host.Credentials != "" {
		creds := strings.SplitN(args.Host.Credentials, ":", 2)
		if len(creds) == 2 {
			u.User = url.UserPassword(creds[0], creds[1])
		} else {
			u.User = url.User(creds[0])
		}
	}

	// get timeout
	timeout := args.Host.Timeout
	if v := args.getContextKey("timeout"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			timeout = d
		}
	}

	// build options
	opts := []transrpc.ClientOption{
		transrpc.WithUserAgent("transctl/" + version + " (" + runtime.GOOS + "/" + runtime.GOARCH + ")"),
		transrpc.WithURL(u.String()),
		transrpc.WithTimeout(timeout),
	}

	// load netrc credentials
	var setFallback bool
	if !args.Host.NoNetrc && !args.Host.CredentialsWasSet {
		fi, err := os.Stat(args.Host.NetrcFile)
		if err == nil && !fi.IsDir() {
			if n, err := netrc.Parse(args.Host.NetrcFile); err == nil {
				if m := n.Machine(u.Hostname()); m != nil {
					user, pass := m.Get("login"), m.Get("password")
					if user != "" {
						setFallback = true
						opts = append(opts, transrpc.WithCredentialFallback(user, pass))
					}
				}
			}
		}
	}

	// set fallback credentials for localhost when none were specified
	if !setFallback && !args.Host.CredentialsWasSet && u.Hostname() == "localhost" {
		opts = append(opts, transrpc.WithCredentialFallback("transmission", "transmission"))
	}

	if args.Verbose {
		opts = append(opts, transrpc.WithLogf(args.logf(os.Stderr, "> "), args.logf(os.Stderr, "< ")))
	}

	return transrpc.NewClient(opts...), nil
}

// logf creates a new log func with the specified prefix.
func (args *Args) logf(w io.Writer, prefix string) func(string, ...interface{}) {
	return func(s string, v ...interface{}) {
		s = strings.TrimSuffix(fmt.Sprintf(s, v...), "\n")
		fmt.Fprintln(w, prefix+strings.Replace(s, "\n", "\n"+prefix, -1)+"\n")
	}
}

// formatByteCount formats a byte count for display.
func (args *Args) formatByteCount(x transrpc.ByteCount, hasSuffix bool) string {
	suffix, prec := "", 2
	if hasSuffix {
		suffix = "/s"
	}
	if args.Output.Human == "true" || args.Output.Human == "1" || args.Output.SI {
		if args.Output.SI && int64(x) < 1024*1024 || !args.Output.SI && int64(x) < 1000*1000 {
			prec = 0
		}
		return x.Format(!args.Output.SI, prec, suffix)
	}
	return fmt.Sprintf("%d%s", x, suffix)
}
