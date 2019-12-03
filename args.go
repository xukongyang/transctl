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

	"github.com/alecthomas/kingpin"
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

	// CredentialsWasSet is true when credentials were set on command line.
	CredentialsWasSet bool

	// NetRC toggles enabling .netrc loading.
	Netrc bool

	// NetRCFile is the NetRCFile to use.
	NetrcFile string

	// Context is the global context name.
	Context string

	// Config is the loaded settings from the config file.
	Config *ini.File

	// ConfigParams are the config params.
	ConfigParams struct {
		Name  string
		Value string
		Unset bool
	}

	// AddParams are the add params.
	AddParams struct {
		Cookies           map[string]string
		DownloadDir       string
		Paused            bool
		PeerLimit         int64
		BandwidthPriority int64
		Remove            bool
	}

	// Output is the output format type.
	Output string

	// All is the all toggle.
	All bool

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

	args := &Args{}
	args.AddParams.Cookies = make(map[string]string)

	// global options
	kingpin.Flag("verbose", "toggle verbose").Short('v').Default("false").BoolVar(&args.Verbose)
	kingpin.Flag("config", "config file").Short('C').Default(configFile).Envar("TRANSCONFIG").PlaceHolder("<file>").StringVar(&args.ConfigFile)
	kingpin.Flag("context", "config context").Short('c').Envar("TRANSCONTEXT").PlaceHolder("<context>").StringVar(&args.Context)
	kingpin.Flag("url", "transmission rpc url").Short('U').Envar("TRANSURL").PlaceHolder("<url>").URLVar(&args.URL)
	kingpin.Flag("netrc", "enable netrc loading (enabled by default)").Short('n').Default("true").BoolVar(&args.Netrc)
	kingpin.Flag("netrc-file", "netrc file path").Default(netrcFile).PlaceHolder("<file>").StringVar(&args.NetrcFile)
	kingpin.Flag("proto", "protocol to use").Default("http").PlaceHolder("<proto>").StringVar(&args.Proto)
	kingpin.Flag("user", "username and password").Short('u').PlaceHolder("<user:pass>").IsSetByUser(&args.CredentialsWasSet).StringVar(&args.Credentials)
	kingpin.Flag("host", "remote host").Short('h').Default("localhost:9091").PlaceHolder("<host>").StringVar(&args.Host)
	kingpin.Flag("rpc-path", "rpc path").Default("/transmission/rpc/").PlaceHolder("<path>").StringVar(&args.RpcPath)

	// config command
	configCmd := kingpin.Command("config", "Get and set transctl configuration")
	configCmd.Arg("name", "option name").Required().StringVar(&args.ConfigParams.Name)
	configCmd.Arg("value", "value").StringVar(&args.ConfigParams.Value)
	configCmd.Flag("unset", "unset value").BoolVar(&args.ConfigParams.Unset)

	// context-set command
	contextSetCmd := kingpin.Command("context-set", "Set default context")
	contextSetCmd.Arg("context", "context name").Required().StringVar(&args.Context)

	// get command
	getCmd := kingpin.Command("get", "Get information about torrents")
	getCmd.Flag("output", "output format").Short('o').PlaceHolder("<format>").EnumVar(&args.Output, "table", "wide", "json", "yaml")
	getCmd.Flag("all", "all torrents").BoolVar(&args.All)
	getCmd.Arg("torrents", "torrent name or identifier").StringsVar(&args.Args)

	// add command
	addCmd := kingpin.Command("add", "Add torrents")
	addCmd.Flag("cookies", "cookies").Short('k').PlaceHolder("<name>=<v>").StringMapVar(&args.AddParams.Cookies)
	addCmd.Flag("download-dir", "download directory").Short('d').PlaceHolder("<dir>").StringVar(&args.AddParams.DownloadDir)
	addCmd.Flag("paused", "start torrent paused").Short('P').BoolVar(&args.AddParams.Paused)
	addCmd.Flag("peer-limit", "peer limit").Short('l').PlaceHolder("<limit>").Int64Var(&args.AddParams.PeerLimit)
	addCmd.Flag("bandwidth-priority", "bandwidth priority").Short('b').PlaceHolder("<bw>").Int64Var(&args.AddParams.BandwidthPriority)
	addCmd.Flag("rm", "remove torrents after adding").BoolVar(&args.AddParams.Remove)
	addCmd.Arg("torrents", "torrent file or URL").StringsVar(&args.Args)

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
		context = "context." + context
	} else {
		context = "default"
	}
	return args.Config.GetKey(context + "." + name)
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

	// add credentials
	if u.User == nil && args.CredentialsWasSet && args.Credentials != "" {
		creds := strings.SplitN(args.Credentials, ":", 2)
		if len(creds) == 2 {
			u.User = url.UserPassword(creds[0], creds[1])
		} else {
			u.User = url.User(creds[0])
		}
	}

	// build options
	opts := []transrpc.ClientOption{
		transrpc.WithUserAgent("transctl/" + version + " (" + runtime.GOOS + "/" + runtime.GOARCH + ")"),
		transrpc.WithURL(u.String()),
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

// logf creates a new log func with the specified prefix.
func (args *Args) logf(w io.Writer, prefix string) func(string, ...interface{}) {
	return func(s string, v ...interface{}) {
		s = strings.TrimSuffix(fmt.Sprintf(s, v...), "\n")
		fmt.Fprintln(w, prefix+strings.Replace(s, "\n", "\n"+prefix, -1)+"\n")
	}
}
