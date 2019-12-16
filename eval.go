package main

import (
	"context"
	"sort"
	"strings"

	"github.com/PaesslerAG/gval"
	"github.com/gobwas/glob"
	"github.com/kenshaw/transrpc"
)

// findTorrents finds torrents based on the identifier args.
func findTorrents(args *Args) (*transrpc.Client, []transrpc.Torrent, error) {
	cl, err := args.newClient()
	if err != nil {
		return nil, nil, err
	}

	var req *transrpc.TorrentGetRequest
	switch {
	case args.Filter.Recent:
		req = transrpc.TorrentGet(transrpc.RecentlyActive).WithFields("hashString")
	case args.Filter.ListAll:
		req = transrpc.TorrentGet().WithFields("hashString")
	case args.Filter.Filter != "":
		// evaluate filter expression to build field names
		fieldnames, err := extractVars(args)
		if err != nil {
			return nil, nil, err
		}
		req = transrpc.TorrentGet().WithFields(fieldnames...)
	default:
		return nil, nil, ErrMustSpecifyListRecentFilterOrAtLeastOneTorrent
	}

	res, err := req.Do(context.Background(), cl)
	if err != nil {
		return nil, nil, err
	}
	if args.Filter.ListAll || args.Filter.Recent {
		return cl, res.Torrents, nil
	}

	l := buildQueryLanguage()

	// filter torrents
	var torrents []transrpc.Torrent
	for _, t := range res.Torrents {
		m := buildJSONMap(t)
		if len(args.Args) == 0 {
			torrents, err = appendMatch(torrents, args, t, m, l)
			if err != nil {
				return nil, nil, err
			}
		} else {
			for _, identifier := range args.Args {
				m["identifier"] = identifier
				torrents, err = appendMatch(torrents, args, t, m, l)
				if err != nil {
					return nil, nil, err
				}
			}
		}
	}

	return cl, torrents, nil
}

// appendMatch
func appendMatch(torrents []transrpc.Torrent, args *Args, t transrpc.Torrent, m map[string]interface{}, l gval.Language) ([]transrpc.Torrent, error) {
	match, err := gval.Evaluate(args.Filter.Filter, m, l)
	if err != nil {
		return nil, err
	}
	v, ok := match.(bool)
	if !ok {
		return nil, ErrFilterMustReturnBool
	}
	if v {
		return append(torrents, t), nil
	}
	return torrents, nil
}

// buildQueryLanguage builds the jsonpath language used for queries.
func buildQueryLanguage() gval.Language {
	return gval.NewLanguage(
		gval.Full(),
		gval.InfixEvalOperator("%%", globOperator),
		gval.Precedence("%%", 40),
		gval.Function("hasPrefix", hasPrefixFunc),
	)
}

// globOperator is the glob operator implementation.
func globOperator(a, b gval.Evaluable) (gval.Evaluable, error) {
	if !b.IsConst() {
		return func(c context.Context, o interface{}) (interface{}, error) {
			a, err := a.EvalString(c, o)
			if err != nil {
				return nil, err
			}
			b, err := b.EvalString(c, o)
			if err != nil {
				return nil, err
			}
			g, err := glob.Compile(b)
			if err != nil {
				return nil, err
			}
			return g.Match(a), nil
		}, nil
	}
	s, err := b.EvalString(nil, nil)
	if err != nil {
		return nil, err
	}
	g, err := glob.Compile(s)
	if err != nil {
		return nil, err
	}
	return func(c context.Context, v interface{}) (interface{}, error) {
		s, err := a.EvalString(c, v)
		if err != nil {
			return nil, err
		}
		return g.Match(s), nil
	}, nil
}

// hasPrefixFunc is the hasPrefix function implementation.
func hasPrefixFunc(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, ErrHasPrefixTakesExactlyTwoArguments
	}
	a, ok := args[0].(string)
	if !ok {
		return nil, ErrSMustBeAString
	}
	b, ok := args[1].(string)
	if !ok {
		return nil, ErrPrefixMustBeAString
	}
	return strings.HasPrefix(strings.ToLower(a), strings.ToLower(b)), nil
}

// extractVars extracts the var names from the provided expression against thing.
func extractVars(args *Args) ([]string, error) {
	// build column mappings
	inverseCols := make(map[string]string, len(args.Output.ColumnNames))
	for k, v := range args.Output.ColumnNames {
		inverseCols[v] = k
	}

	keys := map[string]bool{"hashString": true}
	_, err := gval.Evaluate(
		args.Filter.Filter,
		map[string]interface{}{},
		buildQueryLanguage(),
		gval.VariableSelector(func(path gval.Evaluables) gval.Evaluable {
			return func(c context.Context, v interface{}) (interface{}, error) {
				k, err := path.EvalStrings(c, v)
				if err != nil {
					return nil, err
				}
				key := strings.Join(k, ".")
				if key != "identifier" {
					keys[key] = true
				}
				return key, nil
			}
		}),
	)
	if err != nil {
		return nil, err
	}
	var ret []string
	for k := range keys {
		if c, ok := inverseCols[k]; ok {
			k = c
		}
		ret = append(ret, k)
	}
	sort.Strings(ret)
	return ret, nil
}
