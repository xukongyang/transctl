package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/PaesslerAG/gval"
	"github.com/PaesslerAG/jsonpath"
	"github.com/gobwas/glob"
	"github.com/kenshaw/transrpc"
)

// findTorrents finds torrents based on the identifier args.
func findTorrents(args *Args) (*transrpc.Client, []transrpc.Torrent, error) {
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

	// evaluate filter expression to build fieldnames to retrieve
	var fieldnames []string

	// limit returned fields to match fields only
	req := transrpc.TorrentGet(ids...).WithFields(fieldnames...)
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
				t, id = t, id
			}
		}
	}

	return cl, torrents, nil
}

// buildQueryLanguage builds the jsonpath language used for queries.
func buildQueryLanguage() gval.Language {
	return gval.NewLanguage(
		jsonpath.Language(),
		gval.InfixOperator("%%", func(a, b interface{}) (interface{}, error) {
			g, err := glob.Compile(fmt.Sprintf("%s", b))
			if err != nil {
				return false, err
			}
			return g.Match(fmt.Sprintf("%s", a)), nil
		}),
		gval.Function("hasPrefix", func(s, prefix string) bool {
			return strings.HasPrefix(strings.ToLower(s), strings.ToLower(prefix))
		}),
	)
}
