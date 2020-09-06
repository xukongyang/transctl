package providers

import (
	"context"

	"github.com/kenshaw/transctl/tctypes"
)

// Provider is a torrent provider.
type Provider interface {
	// NewRemoteConfigStore creates a config store for the remote host.
	NewRemoteConfigStore(context.Context) (ConfigStore, error)

	// Add adds a torrent ([]byte) or magnet link (string).
	Add(context.Context, ...interface{}) ([]tctypes.Torrent, error)

	// Get returns a semi-populated torrent list, with the provided fields.
	Get(context.Context, []string, ...interface{}) ([]tctypes.Torrent, error)

	// Set sets a configuration option on a torrent.
	Set(context.Context, map[string]interface{}) error

	// Start starts the provided identifiers.
	Start(context.Context, ...interface{}) error

	// Stop stops the provided identifiers.
	Stop(context.Context, ...interface{}) error

	// Move moves the provided identifiers
	Move(context.Context, string, ...interface{}) error

	// Remove removes the provided identifiers.
	Remove(context.Context, bool, ...interface{}) error

	// Verify verifies the provided identifiers.
	Verify(context.Context, ...interface{}) error

	// Reannounce reannounces the provided identifiers.
	Reannounce(context.Context, ...interface{}) error

	// Queue sets the queue position (top, bottom, up, down) for the provided identifiers.
	Queue(context.Context, string, ...interface{}) error

	// PeersGet returns the peers for the provided identifiers.
	PeersGet(context.Context, ...interface{}) ([]tctypes.Peer, error)

	// FilesGet returns the files for the provided identifiers.
	FilesGet(context.Context, ...interface{}) ([]tctypes.File, error)

	// FilesSet sets a file config option on the provided identifiers.
	FilesSet(context.Context, map[string]interface{}, ...interface{}) error

	// FilesRename renames a file on the provided identifiers.
	FilesRename(context.Context, string, string, ...interface{}) error

	// TrackersGet returns the trackers for the provided identifiers.
	TrackersGet(context.Context, ...interface{}) ([]tctypes.Tracker, error)

	// TrackersAdd adds a tracker to the provided identifiers.
	TrackersAdd(context.Context, string, ...interface{}) error

	// TrackersReplace replaces a tracker on the provided identifiers.
	TrackersReplace(context.Context, string, string, ...interface{}) error

	// TrackersRemove removes a tracker from the provided identifiers.
	TrackersRemove(context.Context, string, ...interface{}) error

	// Stats returns the stats for the remote host.
	Stats(context.Context) (map[string]interface{}, error)

	// Shutdown shuts down the remote host.
	Shutdown(context.Context) error

	// FreeSpace returns the free space for the provided location.
	FreeSpace(context.Context, string) (tctypes.ByteCount, error)

	// BlocklistUpdate tells the remote host to update its blocklist.
	BlocklistUpdate(context.Context) (int64, error)

	// PortTest tells the remote to do a port test.
	PortTest(context.Context) (bool, error)
}

// providers are the registered providers.
var providers map[string]func(*Args) (Provider, error)

// Register registers a provider.
func Register(name string, f func(*Args) (Provider, error)) {
	providers[name] = f
}

func init() {
	providers = make(map[string]func() Provider)
}

// ConfigStore is the interface for config stores.
type ConfigStore interface {
	GetKey(string) string
	SetKey(string, string)
	RemoveKey(string)
	GetMapFlat() map[string]string
	GetAllFlat() []string
	Write(string) error
}
