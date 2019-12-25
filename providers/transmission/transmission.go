package transmission

import (
	"context"
	"reflect"
	"sort"

	"github.com/kenshaw/torctl/providers"
	"github.com/kenshaw/torctl/transrpc"
)

func init() {
	providers.Register("transmission", New)
}

/*
	"config":             doConfig,
	"add":                doAdd,
	"get":                doGet,
	"set":                doSet,
	"start":              doReq(transrpc.TorrentStart),
	"stop":               doReq(transrpc.TorrentStop),
	"move":               doMove,
	"remove":             doRemove,
	"verify":             doReq(transrpc.TorrentVerify),
	"reannounce":         doReq(transrpc.TorrentReannounce),
	"queue top":          doReq(transrpc.QueueMoveTop),
	"queue bottom":       doReq(transrpc.QueueMoveBottom),
	"queue up":           doReq(transrpc.QueueMoveUp),
	"queue down":         doReq(transrpc.QueueMoveDown),
	"peers get":          doPeersGet,
	"files get":          doFilesGet,
	"files set-priority": doFilesSetPriority,
	"files set-wanted":   doFilesSet("FilesWanted"),
	"files set-unwanted": doFilesSet("FilesUnwanted"),
	"files rename":       doFilesRename,
	"trackers get":       doTrackersGet,
	"trackers add":       doTrackersAdd,
	"trackers replace":   doTrackersReplace,
	"trackers remove":    doTrackersRemove,
	"stats":              doStats,
	"shutdown":           doShutdown,
	"free-space":         doFreeSpace,
	"blocklist-update":   doBlocklistUpdate,
	"port-test":          doPortTest,
*/

// Provider is a transmission rpc host provider.
type Provider struct {
}

// New creates a new transmission rpc host provdier.
func New(args *providers.Args) (providers.Provider, error) {
	return &Provider{}, nil
}

// RemoteConfigStore wraps setting configuration for the transrpc rpc host.
type RemoteConfigStore struct {
	cl      *transrpc.Client
	session *transrpc.Session
	setKeys []string
}

// NewRemoteConfigStore creates a new remote config store.
func (p *Provider) NewRemoteConfigStore(ctx context.Context) (providers.RemoveConfigStore, error) {
	session, err := p.cl.SessionGet(ctx)
	if err != nil {
		return nil, err
	}
	return &RemoteConfigStore{cl: p.cl, session: session}, nil
}

// GetKey satisfies the ConfigStore interface.
func (r *RemoteConfigStore) GetKey(key string) string {
	return r.GetMapFlat()[key]
}

// SetKey satisfies the ConfigStore interface.
func (r *RemoteConfigStore) SetKey(key, value string) {
	r.setKeys = append(r.setKeys, key, value)
}

// GetMapFlat satisfies the ConfigStore interface.
func (r *RemoteConfigStore) GetMapFlat() map[string]string {
	m := make(map[string]string)
	addFieldsToMap(m, "", reflect.ValueOf(*r.session))
	return m
}

// GetAllFlat satisfies the ConfigStore interface.
func (r *RemoteConfigStore) GetAllFlat() []string {
	m := make(map[string]string)
	addFieldsToMap(m, "", reflect.ValueOf(*r.session))
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var ret []string
	for _, k := range keys {
		ret = append(ret, k, m[k])
	}
	return ret
}

// Write satisfies the ConfigStore interface.
func (r *RemoteConfigStore) Write(string) error {
	return doWithAndExecute(r.cl, transrpc.SessionSet(), "--remote config", r.setKeys...)
}
