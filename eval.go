package main

import (
	"context"
	"reflect"
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
	var fields map[string][]string
	switch {
	case args.Filter.Recent:
		req = transrpc.TorrentGet(transrpc.RecentlyActive).WithFields("hashString")
	case args.Filter.ListAll:
		req = transrpc.TorrentGet().WithFields("hashString")
	case args.Filter.Filter != "":
		// evaluate filter expression to build field names
		fields, err = extractVars(args)
		if err != nil {
			return nil, nil, err
		}
		var fieldnames []string
		for k := range fields {
			fieldnames = append(fieldnames, k)
		}
		sort.Strings(fieldnames)
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
		m := buildJSONMap(t, fields)
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

// extractVars extracts the var names from the provided expression against thing.
func extractVars(args *Args) (map[string][]string, error) {
	keys := map[string]bool{"hashString": true}
	_, err := gval.Evaluate(
		args.Filter.Filter,
		map[string]interface{}{},
		buildQueryLanguage(),
		gval.VariableSelector(func(path gval.Evaluables) gval.Evaluable {
			return func(ctx context.Context, v interface{}) (interface{}, error) {
				k, err := path.EvalStrings(ctx, v)
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

	// build column mappings
	inverseCols := make(map[string]string, len(args.Output.ColumnNames))
	for k, v := range args.Output.ColumnNames {
		inverseCols[v] = k
	}
	m := make(map[string][]string)
	for k := range keys {
		v := k
		if c, ok := inverseCols[k]; ok {
			k = c
		}
		m[k] = append(m[k], v)
	}
	return m, nil
}

// buildJSONMap builds a JSON map.
func buildJSONMap(v interface{}, fields map[string][]string) map[string]interface{} {
	res := map[string]interface{}{}
	if v == nil {
		return res
	}
	typ := reflect.TypeOf(v)
	indirect := reflect.Indirect(reflect.ValueOf(v))
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	for i := 0; i < typ.NumField(); i++ {
		tag := strings.TrimSpace(strings.SplitN(typ.Field(i).Tag.Get("json"), ",", 2)[0])
		if tag == "" || tag == "-" {
			continue
		}
		cols, ok := fields[tag]
		if !ok {
			continue
		}
		var y interface{}
		f := indirect.Field(i).Interface()
		switch typ.Field(i).Type.Kind() {
		case reflect.Struct:
			y = buildJSONMap(f, fields)
		case reflect.Slice:
		default:
			y = f
		}
		for _, col := range cols {
			res[col] = y
		}
	}
	return res
}

// appendMatch appends torrents matching the filter, returning the aggregate
// slice.
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
		gval.InfixEvalOperator("%^", prefixOperator),
		gval.Precedence("%^", 40),
	)
}

// globOperator is the glob operator implementation.
func globOperator(a, b gval.Evaluable) (gval.Evaluable, error) {
	if !b.IsConst() {
		return func(ctx context.Context, o interface{}) (interface{}, error) {
			astr, err := a.EvalString(ctx, o)
			if err != nil {
				return nil, err
			}
			bstr, err := b.EvalString(ctx, o)
			if err != nil {
				return nil, err
			}
			g, err := glob.Compile(bstr)
			if err != nil {
				return nil, err
			}
			return g.Match(astr), nil
		}, nil
	}
	bstr, err := b.EvalString(context.TODO(), nil)
	if err != nil {
		return nil, err
	}
	g, err := glob.Compile(bstr)
	if err != nil {
		return nil, err
	}
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		astr, err := a.EvalString(ctx, v)
		if err != nil {
			return nil, err
		}
		return g.Match(astr), nil
	}, nil
}

// prefixOperator is the glob operator implementation.
func prefixOperator(a, b gval.Evaluable) (gval.Evaluable, error) {
	if !b.IsConst() {
		return func(ctx context.Context, o interface{}) (interface{}, error) {
			astr, err := a.EvalString(ctx, o)
			if err != nil {
				return nil, err
			}
			bstr, err := b.EvalString(ctx, o)
			if err != nil {
				return nil, err
			}
			return strings.HasPrefix(strings.ToLower(astr), strings.ToLower(bstr)), nil
		}, nil
	}
	bstr, err := b.EvalString(context.TODO(), nil)
	if err != nil {
		return nil, err
	}
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		astr, err := a.EvalString(ctx, v)
		if err != nil {
			return nil, err
		}
		return strings.HasPrefix(astr, bstr), nil
	}, nil
}
