package utorrent

import (
	"context"

	"github.com/kenshaw/transctl/providers"
)

type Torrent struct {
}

func init() {
	providers.Register("utorrent", New)
}

func New() providers.Provider {
	return &Provider{}
}

type Provider struct {
}

func (p *Provider) FindTorrents(ctx context.Context, fields ...string) ([]Torrent, error) {
	return nil, nil
}
