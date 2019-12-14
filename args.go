package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/gobwas/glob"
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

		// MatchOrder is torrent identifier match order.
		MatchOrder []string

		// MatchOrderWasSet is the match order was set toggle.
		MatchOrderWasSet bool
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

		// NoHeadersWasSet is the no headers was set toggle.
		NoHeadersWasSet bool

		// NoTotals is the toggle to disable the total line on table output.
		NoTotals bool

		// NoTotalsWasSet is the no totals was set toggle.
		NoTotalsWasSet bool
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

	// FilesSetLocationParams are the files set-location params.
	FilesSetLocationParams struct {
		Location string
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
	addCmd.Flag("output", "output format (default: table)").Short('o').PlaceHolder("<format>").IsSetByUser(&args.Output.OutputWasSet).EnumVar(&args.Output.Output, "table", "wide", "json", "yaml", "flat")
	addCmd.Flag("human", "print sizes in powers of 1024 (e.g., 1023MiB) (default: true)").Default("true").PlaceHolder("true").StringVar(&args.Output.Human)
	addCmd.Flag("si", "print sizes in powers of 1000 (e.g., 1.1GB)").IsSetByUser(&args.Output.SIWasSet).BoolVar(&args.Output.SI)
	addCmd.Flag("no-headers", "disable table header output").IsSetByUser(&args.Output.NoHeadersWasSet).BoolVar(&args.Output.NoHeaders)
	addCmd.Flag("no-totals", "disable table total output").IsSetByUser(&args.Output.NoTotalsWasSet).BoolVar(&args.Output.NoTotals)
	addCmd.Flag("bandwidth-priority", "bandwidth priority").Short('b').PlaceHolder("<bw>").Int64Var(&args.AddParams.BandwidthPriority)
	addCmd.Flag("cookies", "cookies").Short('k').PlaceHolder("<name>=<v>").StringMapVar(&args.AddParams.Cookies)
	addCmd.Flag("download-dir", "download directory").Short('d').PlaceHolder("<dir>").StringVar(&args.AddParams.DownloadDir)
	addCmd.Flag("paused", "start torrent paused").Short('P').BoolVar(&args.AddParams.Paused)
	addCmd.Flag("peer-limit", "peer limit").Short('L').PlaceHolder("<limit>").Int64Var(&args.AddParams.PeerLimit)
	addCmd.Flag("rm", "remove torrents after adding").IsSetByUser(&args.AddParams.RemoveWasSet).BoolVar(&args.AddParams.Remove)
	addCmd.Arg("torrents", "torrent file or URL").StringsVar(&args.Args)

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
		"files set-priority", "Set priority for torrent files",
		"files set-location", "Set location for torrent files",
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
		cmd.Flag("match-order", "match order (default: hash,id,glob)").Short('m').PlaceHolder("<m>,<m>").Default("hash", "id", "glob").EnumsVar(&args.Filter.MatchOrder, "hash", "id", "glob", "regexp")

		switch commands[i] {
		case "get", "files get", "trackers get":
			cmd.Flag("output", "output format (table, wide, json, yaml, flat; default: table)").Short('o').PlaceHolder("<format>").IsSetByUser(&args.Output.OutputWasSet).EnumVar(&args.Output.Output, "table", "wide", "json", "yaml", "flat")
			cmd.Flag("human", "print sizes in powers of 1024 (e.g., 1023MiB) (default: true)").Default("true").PlaceHolder("true").StringVar(&args.Output.Human)
			cmd.Flag("si", "print sizes in powers of 1000 (e.g., 1.1GB)").IsSetByUser(&args.Output.SIWasSet).BoolVar(&args.Output.SI)
			cmd.Flag("no-headers", "disable table header output").IsSetByUser(&args.Output.NoHeadersWasSet).BoolVar(&args.Output.NoHeaders)
			cmd.Flag("no-totals", "disable table total output").IsSetByUser(&args.Output.NoTotalsWasSet).BoolVar(&args.Output.NoTotals)

		case "set":
			cmd.Arg("name", "option name").Required().StringVar(&args.ConfigParams.Name)
			cmd.Arg("value", "option value").Required().StringVar(&args.ConfigParams.Value)

		case "start":
			cmd.Flag("now", "start now").BoolVar(&args.StartParams.Now)

		case "move":
			cmd.Flag("dest", "move destination").Short('d').PlaceHolder("<dir>").StringVar(&args.MoveParams.Dest)

		case "remove":
			cmd.Flag("rm", "remove downloaded files").BoolVar(&args.RemoveParams.Remove)

		case "files set-priority":
			cmd.Arg("mask", "file mask").Required().StringVar(&args.FileMask)
			cmd.Arg("priority", "file priority (low, normal, high)").Required().EnumVar(&args.FilesSetPriorityParams.Priority, "low", "normal", "high")

		case "files set-location":
			cmd.Arg("mask", "file mask").Required().StringVar(&args.FileMask)
			cmd.Arg("location", "file location").Required().StringVar(&args.FilesSetLocationParams.Location)

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
	if v := args.getContextKey("match-order"); v != "" && !args.Filter.MatchOrderWasSet {
		args.Filter.MatchOrder = strings.Split(v, ",")
		for i := 0; i < len(args.Filter.MatchOrder); i++ {
			args.Filter.MatchOrder[i] = strings.ToLower(strings.TrimSpace(args.Filter.MatchOrder[i]))
		}
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
	case "get", "set", "start", "stop", "move", "remove", "verify",
		"reannounce", "queue top", "queue bottom", "queue up", "queue down":
		// check exactly one of --recent, --all, or len(args.Args) > 0 conditions
		switch {
		case args.Filter.ListAll && args.Filter.Recent,
			args.Filter.ListAll && len(args.Args) != 0,
			args.Filter.Recent && len(args.Args) != 0,
			!args.Filter.ListAll && !args.Filter.Recent && len(args.Args) == 0:
			return ErrMustSpecifyAllRecentOrAtLeastOneTorrent
		}

	case "free-space":
		// check that either a location was passed as an argument, or specified
		// via config context options
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

// findTorrents finds torrents based on the identifier args.
func (args *Args) findTorrents() (*transrpc.Client, []transrpc.Torrent, error) {
	cl, err := args.newClient()
	if err != nil {
		return nil, nil, err
	}

	var ids []interface{}
	switch {
	case args.Filter.Recent:
		ids = append(ids, transrpc.RecentlyActive)
	case args.Filter.ListAll:
	default:
	}

	// limit returned fields to match fields only
	req := transrpc.TorrentGet(ids...).WithFields("id", "name", "hashString")
	res, err := req.Do(context.Background(), cl)
	if err != nil {
		return nil, nil, err
	}

	// filter torrents
	var torrents []transrpc.Torrent
	if args.Filter.ListAll || args.Filter.Recent {
		torrents = res.Torrents
	} else {
		for _, t := range res.Torrents {
			for _, id := range args.Args {
				g, gerr := glob.Compile(id)
				re, reerr := regexp.Compile(id)
				for _, m := range args.Filter.MatchOrder {
					switch m {
					case "id":
						if id == strconv.FormatInt(t.ID, 10) {
							torrents = append(torrents, t)
						}
					case "hash":
						if len(id) >= minimumHashCompareLen && strings.HasPrefix(t.HashString, id) {
							torrents = append(torrents, t)
						}
					case "glob":
						if gerr == nil && g.Match(t.Name) {
							torrents = append(torrents, t)
						}
					case "regexp":
						if reerr == nil && re.MatchString(t.Name) {
							torrents = append(torrents, t)
						}
					default:
						return nil, nil, ErrInvalidMatchOrder
					}
				}
			}
		}
	}

	return cl, torrents, nil
}

// logf creates a new log func with the specified prefix.
func (args *Args) logf(w io.Writer, prefix string) func(string, ...interface{}) {
	return func(s string, v ...interface{}) {
		s = strings.TrimSuffix(fmt.Sprintf(s, v...), "\n")
		fmt.Fprintln(w, prefix+strings.Replace(s, "\n", "\n"+prefix, -1)+"\n")
	}
}
