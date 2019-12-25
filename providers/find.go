package providers

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"

	"github.com/PaesslerAG/gval"
	"github.com/gobwas/glob"
	"github.com/knq/snaker"

	"github.com/kenshaw/torctl/tctypes"
)

// findTorrents finds torrents based on the identifier args.
func findTorrents(args *Args) (Provider, []tctypes.Torrent, error) {
	p, err := args.NewProvider()
	if err != nil {
		return nil, nil, err
	}

	var req *TorrentGetRequest
	var fields map[string][]string
	switch {
	case args.Filter.Recent:
		req = p.TorrentGet([]string{"hashString"}, "recently-active")
	case args.Filter.ListAll:
		req = p.TorrentGet([]string{"hashString"})
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
		req = p.TorrentGet(fieldnames)
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
	var torrents []tctypes.Torrent
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

// extractVars extracts the var names from the provided expression.
func extractVars(args *Args) (map[string][]string, error) {
	// build column mappings
	inverseCols := make(map[string]string, len(getColumnNames))
	for _, n := range getColumnNames {
		k := strings.SplitN(n, "=", 2)
		inverseCols[k[1]] = k[0]
	}

	keys := map[string]bool{"hashString": true}
	typ := reflect.TypeOf(tctypes.Torrent{})
	b, err := gval.Evaluate(
		args.Filter.Filter,
		nil,
		buildQueryLanguage(),
		gval.VariableSelector(func(path gval.Evaluables) gval.Evaluable {
			k, err := path.EvalStrings(context.Background(), nil)
			if err == nil {
				key := strings.Join(k, ".")
				if _, ok := sizeConsts[key]; !ok && key != "identifier" {
					keys[key] = true
				}
			}
			return func(ctx context.Context, v interface{}) (interface{}, error) {
				// evaluate the key name, return it's zero value
				k, err := path.EvalStrings(ctx, v)
				if err != nil {
					return nil, err
				}
				key := strings.Join(k, ".")
				if c, ok := inverseCols[key]; ok {
					key = c
				}
				if key == "identifier" {
					return "", nil
				}
				if _, ok := sizeConsts[key]; ok {
					return int64(0), nil
				}
				f, ok := readFieldOrMethodType(typ, snaker.ForceCamelIdentifier(key))
				if !ok {
					return nil, fmt.Errorf("unknown filter field or method %q", key)
				}
				return reflect.Zero(f).Interface(), nil
			}
		}),
	)
	if err != nil {
		return nil, err
	}
	if _, ok := b.(bool); !ok {
		return nil, ErrFilterMustReturnBool
	}

	// build field => variable map
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

// buildJSONMap uses reflect to build a map of v's fields, using it's json tag
// as key.
func buildJSONMap(v interface{}, fields map[string][]string) map[string]interface{} {
	res := map[string]interface{}{}
	for k, v := range sizeConsts {
		res[k] = v
	}
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
func appendMatch(torrents []tctypes.Torrent, args *Args, t tctypes.Torrent, m map[string]interface{}, l gval.Language) ([]tctypes.Torrent, error) {
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
		gval.Function("strlen", strlenFunc),
	)
}

// globOperator is the gval glob operator.
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

// prefixOperator is the gval string prefix operator.
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

// strlenFunc is the gval strlen function.
func strlenFunc(args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, ErrInvalidStrlenArguments
	}
	s, ok := args[0].(string)
	if !ok {
		return nil, ErrInvalidStrlenArguments
	}
	return float64(len(s)), nil
}

// sizeConsts are size constants used in filter expressions.
var sizeConsts map[string]int64

func init() {
	sizeConsts = make(map[string]int64)
	for i, r := range "kMGTPE" {
		sizeConsts[string(r)+"B"] = int64(math.Pow(1000, float64(i+1)))
		sizeConsts[strings.ToUpper(string(r))+"iB"] = int64(math.Pow(1024, float64(i+1)))
	}
}
