package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kingpin"
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

	// Context is the global context name.
	Context string

	// Config is the loaded settings from the config file.
	Config *ini.File

	// ConfigParams are the config params.
	ConfigParams struct {
		Name  string
		Value string
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
	// determine config dir
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	configFile := filepath.Join(configDir, "transctl", "config.ini")

	kingpin.UsageTemplate(kingpin.CompactUsageTemplate)

	args := &Args{}
	args.AddParams.Cookies = make(map[string]string)

	// global options
	kingpin.Version(Version)
	kingpin.Flag("verbose", "toggle verbose").Short('v').Default("false").BoolVar(&args.Verbose)
	kingpin.Flag("config", "config file").Short('C').Default(configFile).Envar("TRANSCONFIG").PlaceHolder("<file>").StringVar(&args.ConfigFile)
	kingpin.Flag("url", "transmission rpc url").Short('U').PlaceHolder("<url>").URLVar(&args.URL)

	// config command
	configCmd := kingpin.Command("config", "Get and set configuration options")
	configCmd.Arg("name", "option name").Required().StringVar(&args.ConfigParams.Name)
	configCmd.Arg("value", "value").StringVar(&args.ConfigParams.Value)

	// get command
	getCmd := kingpin.Command("get", "Get information about torrents")
	getCmd.Flag("output", "output format").Short('o').StringVar(&args.Output)
	getCmd.Flag("all", "all torrents").BoolVar(&args.All)
	getCmd.Arg("torrents", "torrent name or identifier").StringsVar(&args.Args)

	addCmd := kingpin.Command("add", "Add torrents")
	addCmd.Flag("cookies", "cookies").Short('k').PlaceHolder("NAME=VALUE").StringMapVar(&args.AddParams.Cookies)
	addCmd.Flag("download-dir", "download directory").Short('d').PlaceHolder("<dir>").StringVar(&args.AddParams.DownloadDir)
	addCmd.Flag("paused", "add torrent paused").Short('P').BoolVar(&args.AddParams.Paused)
	addCmd.Flag("peer-limit", "peer limit").Short('l').PlaceHolder("<limit>").Int64Var(&args.AddParams.PeerLimit)
	addCmd.Flag("bandwidth-priority", "bandwidth priority").Short('b').PlaceHolder("<bw>").Int64Var(&args.AddParams.BandwidthPriority)
	addCmd.Flag("rm", "remove file after adding").BoolVar(&args.AddParams.Remove)
	addCmd.Arg("torrents", "torrent file or URL").StringsVar(&args.Args)

	return args, nil
}

// loadConfig loads the configuration file from disk.
func (args *Args) loadConfig() error {
	// check if the file exists
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
	// determine URL
	u := args.URL
	if u == nil {
		urlstr := args.getContextKey("url")
		if urlstr == "" {
			urlstr = defaultURL
		}
		var err error
		u, err = url.Parse(urlstr)
		if err != nil {
			return nil, err
		}
	}

	// build options
	opts := []transrpc.ClientOption{
		transrpc.WithUserAgent("transctl/" + Version),
		transrpc.WithURL(u.String()),
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
