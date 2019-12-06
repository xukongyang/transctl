package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
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
	// Verbose is the global verbose toggle.
	Verbose bool

	// ConfigFile is the global config file.
	ConfigFile string

	// URL is the global URL to work with.
	URL *url.URL

	// Proto is the global proto to use for building URLs.
	Proto string

	// Host is the global host and port to use for building URLs.
	Host string

	// RpcPath is the global rpc path to use for building URLs.
	RpcPath string

	// Credentials is the global user:pass credentials to work with.
	Credentials string

	// CredentialsWasSet is the credentials was set toggle.
	CredentialsWasSet bool

	// NetRC toggles enabling .netrc loading.
	Netrc bool

	// NetRCFile is the NetRCFile to use.
	NetrcFile string

	// Context is the global context name.
	Context string

	// Config is the loaded settings from the config file.
	Config *ini.File

	// Timeout is the rpc host request timeout.
	Timeout time.Duration

	// Human is the toggle to display sizes in powers of 1024 (ie, 1023MiB).
	Human string

	// SI is the toggle to display sizes in powers of 1000 (ie, 1.1GB).
	SI bool

	// SIWasSet is the si was set toggle.
	SIWasSet bool

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

	// Output is the output format type.
	Output string

	// OutputWasSet is the output was set toggle.
	OutputWasSet bool

	// ListAll is the all toggle.
	ListAll bool

	// Recent is the recent toggle.
	Recent bool

	// MatchOrder is torrent identifier match order.
	MatchOrder []string

	// MatchOrderWasSet is the match order was set toggle.
	MatchOrderWasSet bool

	// Args are torrent identifiers to use.
	Args []string
}

// NewArgs creates the command args.
func NewArgs() (*Args, error) {
	// determine netrc path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	netrcFile := filepath.Join(homeDir, ".netrc")
	if runtime.GOOS == "windows" {
		netrcFile = filepath.Join(homeDir, "_netrc")
	}

	// determine config file path
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	configFile := filepath.Join(configDir, "transctl", "config.ini")

	// kingpin settings
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate)

	// create args
	args := &Args{}
	args.AddParams.Cookies = make(map[string]string)

	// global options
	kingpin.Flag("verbose", "toggle verbose (default: false)").Short('v').Default("false").BoolVar(&args.Verbose)
	kingpin.Flag("config", "config file").Short('C').Default(configFile).Envar("TRANSCONFIG").PlaceHolder("<file>").StringVar(&args.ConfigFile)
	kingpin.Flag("context", "config context").Short('c').Envar("TRANSCONTEXT").PlaceHolder("<context>").StringVar(&args.Context)
	kingpin.Flag("url", "remote host url").Short('U').Envar("TRANSURL").PlaceHolder("<url>").URLVar(&args.URL)
	kingpin.Flag("netrc", "enable netrc loading (default: true)").Short('n').Default("true").BoolVar(&args.Netrc)
	kingpin.Flag("netrc-file", "netrc file path").Default(netrcFile).PlaceHolder("<file>").StringVar(&args.NetrcFile)
	kingpin.Flag("proto", "protocol to use (default: http)").Default("http").PlaceHolder("<proto>").StringVar(&args.Proto)
	kingpin.Flag("user", "username and password").Short('u').PlaceHolder("<user:pass>").IsSetByUser(&args.CredentialsWasSet).StringVar(&args.Credentials)
	kingpin.Flag("host", "remote host (default: localhost:9091)").Short('h').PlaceHolder("<host>").StringVar(&args.Host)
	kingpin.Flag("rpc-path", "rpc path (default: /transmission/rpc/)").Default("/transmission/rpc/").PlaceHolder("<path>").StringVar(&args.RpcPath)
	kingpin.Flag("timeout", "request timeout (default: 25s)").Default("25s").PlaceHolder("<dur>").DurationVar(&args.Timeout)

	// config command
	configCmd := kingpin.Command("config", "Get and set local and remote config")
	configCmd.Flag("remote", "get or set remote config option").BoolVar(&args.ConfigParams.Remote)
	configCmd.Arg("name", "option name").StringVar(&args.ConfigParams.Name)
	configCmd.Arg("value", "option value").StringVar(&args.ConfigParams.Value)
	configCmd.Flag("unset", "unset value").BoolVar(&args.ConfigParams.Unset)
	configCmd.Flag("list", "list all options").Short('l').BoolVar(&args.ListAll)
	configCmd.Flag("all", "list all options").Hidden().BoolVar(&args.ListAll)

	// add command
	addCmd := kingpin.Command("add", "Add torrents")
	addCmd.Flag("output", "output format (default: table)").Short('o').PlaceHolder("<format>").IsSetByUser(&args.OutputWasSet).EnumVar(&args.Output, "table", "wide", "json", "yaml", "flat")
	addCmd.Flag("human", "print sizes in powers of 1024 (e.g., 1023MiB) (default: true)").Default("true").StringVar(&args.Human)
	addCmd.Flag("si", "print sizes in powers of 1000 (e.g., 1.1GB)").IsSetByUser(&args.SIWasSet).BoolVar(&args.SI)
	addCmd.Flag("cookies", "cookies").Short('k').PlaceHolder("<name>=<v>").StringMapVar(&args.AddParams.Cookies)
	addCmd.Flag("download-dir", "download directory").Short('d').PlaceHolder("<dir>").StringVar(&args.AddParams.DownloadDir)
	addCmd.Flag("paused", "start torrent paused").Short('P').BoolVar(&args.AddParams.Paused)
	addCmd.Flag("peer-limit", "peer limit").Short('L').PlaceHolder("<limit>").Int64Var(&args.AddParams.PeerLimit)
	addCmd.Flag("bandwidth-priority", "bandwidth priority").Short('b').PlaceHolder("<bw>").Int64Var(&args.AddParams.BandwidthPriority)
	addCmd.Flag("rm", "remove torrents after adding").IsSetByUser(&args.AddParams.RemoveWasSet).BoolVar(&args.AddParams.Remove)
	addCmd.Arg("torrents", "torrent file or URL").StringsVar(&args.Args)

	// add retrieval/manipulation commands
	commands := []string{
		"get", "Get information about torrents",
		"start", "Start torrents",
		"stop", "Stop torrents",
		"move", "Move torrent location",
		"remove", "Remove torrents",
		"verify", "Verify torrents",
		"reannounce", "Reannounce torrents",
		"queue bottom", "Move torrents to bottom of queue",
		"queue top", "Move torrents to top of queue",
		"queue up", "Move torrents up in queue",
		"queue down", "Move torrents down in queue",
	}

	var queueCmd *kingpin.CmdClause
	for i := 0; i < len(commands); i += 2 {
		f := kingpin.Command

		// handle queue command creation
		if strings.HasPrefix(commands[i], "queue ") {
			if queueCmd == nil {
				queueCmd = kingpin.Command("queue", "Change torrent queue position")
			}
			f = queueCmd.Command
		}

		// add command
		cmd := f(strings.TrimPrefix(commands[i], "queue "), commands[i+1])
		cmd.Flag("output", "output format (default: table)").Short('o').PlaceHolder("<format>").IsSetByUser(&args.OutputWasSet).EnumVar(&args.Output, "table", "wide", "json", "yaml", "flat")
		cmd.Flag("list", "list all torrents").Short('l').BoolVar(&args.ListAll)
		cmd.Flag("human", "print sizes in powers of 1024 (e.g., 1023MiB) (default: true)").Default("true").StringVar(&args.Human)
		cmd.Flag("si", "print sizes in powers of 1000 (e.g., 1.1GB)").IsSetByUser(&args.SIWasSet).BoolVar(&args.SI)
		cmd.Flag("all", "list all torrents").Hidden().BoolVar(&args.ListAll)
		cmd.Flag("recent", "recently active torrents").Short('R').BoolVar(&args.Recent)
		cmd.Flag("active", "recently active torrents").Hidden().BoolVar(&args.Recent)
		cmd.Flag("match-order", "match order (default: hash,id,glob)").Short('m').PlaceHolder("<m>,<m>").Default("hash", "id", "glob").EnumsVar(&args.MatchOrder, "hash", "id", "glob")

		switch commands[i] {
		case "start":
			cmd.Flag("now", "start now").BoolVar(&args.StartParams.Now)
		case "move":
			cmd.Flag("dest", "move destination").Short('d').PlaceHolder("<dir>").StringVar(&args.MoveParams.Dest)
		case "remove":
			cmd.Flag("rm", "remove downloaded files").BoolVar(&args.RemoveParams.Remove)
		}

		cmd.Arg("torrents", "torrent name or identifier").StringsVar(&args.Args)
	}

	// stats command
	_ = kingpin.Command("stats", "Get session statistics")

	// shutdown command
	_ = kingpin.Command("shutdown", "Shutdown remote host")

	// free-space command
	freeSpaceCmd := kingpin.Command("free-space", "Retrieve free space")
	freeSpaceCmd.Flag("human", "print sizes in powers of 1024 (e.g., 1023MiB) (default: true)").Default("true").StringVar(&args.Human)
	freeSpaceCmd.Flag("si", "print sizes in powers of 1000 (e.g., 1.1GB)").BoolVar(&args.SI)
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
	}).Short('V').Bool()

	return args, nil
}

// loadConfig loads the configuration file from disk.
func (args *Args) loadConfig() error {
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
	u := args.URL

	// check if host is specified
	if u == nil && args.Host != "" {
		u, err = url.Parse(args.Proto + "://" + args.Host + args.RpcPath)
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
		host := args.Host
		if host == "" {
			host = "localhost:9091"
		}
		u, err = url.Parse(args.Proto + "://" + host + args.RpcPath)
		if err != nil {
			return nil, err
		}
	}

	// add credentials
	if u.User == nil && args.CredentialsWasSet && args.Credentials != "" {
		creds := strings.SplitN(args.Credentials, ":", 2)
		if len(creds) == 2 {
			u.User = url.UserPassword(creds[0], creds[1])
		} else {
			u.User = url.User(creds[0])
		}
	}

	// get timeout
	timeout := args.Timeout
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
	if args.Netrc && !args.CredentialsWasSet {
		fi, err := os.Stat(args.NetrcFile)
		if err == nil && !fi.IsDir() {
			if n, err := netrc.Parse(args.NetrcFile); err == nil {
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
	if !setFallback && !args.CredentialsWasSet && u.Hostname() == "localhost" {
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
	case args.Recent:
		ids = append(ids, transrpc.RecentlyActive)
	case args.ListAll:
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
	if args.ListAll || args.Recent {
		torrents = res.Torrents
	} else {
		for _, t := range res.Torrents {
			for _, id := range args.Args {
				g, err := glob.Compile(id)
				for _, m := range args.MatchOrder {
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
						if err == nil && g.Match(t.Name) {
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
