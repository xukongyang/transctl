package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kenshaw/transrpc"
	"github.com/knq/snaker"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

// TorrentResult is a wrapper type for slice of *transrpc.Torrent's that
// satisfies the tblfmt.ResultSet interface.
type TorrentResult struct {
	torrents []transrpc.Torrent
}

// NewTorrentResult creates a new torrent result output encoder for the passed
// torrents.
func NewTorrentResult(torrents []transrpc.Torrent) *TorrentResult {
	return &TorrentResult{
		torrents: torrents,
	}
}

// Encode encodes the torrent result using the settings in args to the
// io.Writer.
func (tr *TorrentResult) Encode(w io.Writer, args *Args, cl *transrpc.Client) error {
	var f func(io.Writer, *Args, *transrpc.Client) error
	switch output := strings.SplitN(args.Output.Output, "=", 2); output[0] {
	case "table":
		cols := []string{"id", "name", "status", "eta", "rateDownload", "rateUpload", "haveValid", "percentDone", "shortHash"}
		if len(output) > 1 {
			cols = strings.Split(output[1], ",")
		}
		f = tr.encodeTable(cols...)
	case "wide":
		f = tr.encodeTable("id", "name", "peersConnected", "downloadDir", "addedDate", "status", "eta", "rateDownload", "rateUpload", "haveValid", "percentDone", "shortHash")
	case "json":
		f = tr.encodeJSON
	case "yaml":
		f = tr.encodeYaml
	case "flat":
		f = tr.encodeFlat
	default:
		return ErrInvalidOutputOptionSpecified
	}
	return f(w, args, cl)
}

// encodeTableColumns encodes the specified table results with the included
// columns.
func (tr *TorrentResult) encodeTable(columns ...string) func(io.Writer, *Args, *transrpc.Client) error {
	// typ := reflect.TypeOf(transrpc.Torrent{})
	return func(w io.Writer, args *Args, cl *transrpc.Client) error {
		// check that at least one column was non-empty
		var cols []string
		for i := 0; i < len(columns); i++ {
			c := strings.TrimSpace(columns[i])
			if c == "" {
				continue
			}
			cols = append(cols, c)
		}
		if len(cols) < 1 {
			return ErrMustSpecifyAtLeastOneOutputColumn
		}

		// build column mappings
		inverseCols := make(map[string]string, len(args.Output.ColumnNames))
		for k, v := range args.Output.ColumnNames {
			inverseCols[v] = k
		}
		headers := make([]string, len(cols))
		fieldnames := make([]string, len(cols))
		colnames := make([]string, len(cols))
		sortByField := ""
		sortBy := strings.TrimSpace(args.Output.SortBy)
		for i := 0; i < len(cols); i++ {
			if c, ok := inverseCols[cols[i]]; ok {
				cols[i] = c
			}
			headers[i] = cols[i]
			if h, ok := args.Output.ColumnNames[cols[i]]; ok {
				headers[i] = h
			}
			headers[i] = strings.ToUpper(headers[i])
			if cols[i] == "shortHash" {
				fieldnames[i] = "hashString"
			} else {
				fieldnames[i] = cols[i]
			}
			colnames[i] = snaker.ForceCamelIdentifier(cols[i])
			if sortBy == cols[i] || strings.EqualFold(sortBy, headers[i]) {
				sortByField = colnames[i]
			}
		}

		switch {
		case sortByField == "" && !args.Output.SortByWasSet:
			sortByField = colnames[0]
		case sortByField == "":
			return ErrSortByNotInColumnList
		}

		// build base request
		var torrents []transrpc.Torrent
		if len(tr.torrents) != 0 {
			req := transrpc.TorrentGet(convTorrentIDs(tr.torrents)...).WithFields(fieldnames...)
			res, err := req.Do(context.Background(), cl)
			if err != nil {
				return err
			}
			torrents = res.Torrents
		}

		// sort
		sort.Slice(torrents, func(i, j int) bool {
			a := reflect.ValueOf(torrents[i]).FieldByName(sortByField)
			if a.Kind() == reflect.Invalid {
				a = reflect.ValueOf(torrents[i]).MethodByName(sortByField)
				if a.Kind() == reflect.Invalid {
					panic("unknown torrent field or method " + args.Output.SortBy)
				}
				a = a.Call([]reflect.Value{})[0]
			}
			b := reflect.ValueOf(torrents[j]).FieldByName(sortByField)
			if b.Kind() == reflect.Invalid {
				b = reflect.ValueOf(torrents[j]).MethodByName(sortByField)
				if b.Kind() == reflect.Invalid {
					panic("unknown torrent field or method " + args.Output.SortBy)
				}
				b = b.Call([]reflect.Value{})[0]
			}

			switch x := a.Interface().(type) {
			case string:
				if args.Output.SortOrder == "desc" {
					return x > b.Interface().(string)
				}
				return x < b.Interface().(string)
			case int64:
				if args.Output.SortOrder == "desc" {
					return x > b.Interface().(int64)
				}
				return x < b.Interface().(int64)
			case float64:
				if args.Output.SortOrder == "desc" {
					return x > b.Interface().(float64)
				}
				return x < b.Interface().(float64)
			case transrpc.Percent:
				if args.Output.SortOrder == "desc" {
					return x > b.Interface().(transrpc.Percent)
				}
				return x < b.Interface().(transrpc.Percent)
			case transrpc.Status:
				if args.Output.SortOrder == "desc" {
					return x > b.Interface().(transrpc.Status)
				}
				return x < b.Interface().(transrpc.Status)
			case transrpc.Priority:
				if args.Output.SortOrder == "desc" {
					return x > b.Interface().(transrpc.Priority)
				}
				return x < b.Interface().(transrpc.Priority)
			case transrpc.Encryption:
				if args.Output.SortOrder == "desc" {
					return x > b.Interface().(transrpc.Encryption)
				}
				return x < b.Interface().(transrpc.Encryption)
			case transrpc.ByteCount:
				if args.Output.SortOrder == "desc" {
					return x > b.Interface().(transrpc.ByteCount)
				}
				return x < b.Interface().(transrpc.ByteCount)
			case transrpc.Duration:
				if args.Output.SortOrder == "desc" {
					return x > b.Interface().(transrpc.Duration)
				}
				return x < b.Interface().(transrpc.Duration)
			case transrpc.Time:
				if args.Output.SortOrder == "desc" {
					return time.Time(x).After(time.Time(b.Interface().(transrpc.Time)))
				}
				return time.Time(x).Before(time.Time(b.Interface().(transrpc.Time)))
			case transrpc.Bool:
				return false
			default:
				panic(fmt.Sprintf("unknown comparison type %T", a.Interface()))
			}
		})

		// tablewriter package is temporary until tblfmt is fixed
		tbl := tablewriter.NewWriter(w)
		if !args.Output.NoHeaders {
			tbl.SetHeader(headers)
		}
		tbl.SetAutoWrapText(false)
		tbl.SetAutoFormatHeaders(true)
		tbl.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		tbl.SetAlignment(tablewriter.ALIGN_LEFT)
		tbl.SetCenterSeparator("")
		tbl.SetColumnSeparator("")
		tbl.SetRowSeparator("")
		tbl.SetHeaderLine(false)
		tbl.SetBorder(false)
		tbl.SetTablePadding("\t") // pad with tabs
		tbl.SetNoWhiteSpace(true)

		// process torrents
		hasTotals := false
		display, totals := make([]bool, len(cols)), make([]transrpc.ByteCount, len(cols))
		for _, t := range torrents {
			row := make([]string, len(cols))
			for i := 0; i < len(cols); i++ {
				v := reflect.ValueOf(t).FieldByName(colnames[i])
				if v.Kind() == reflect.Invalid {
					v = reflect.ValueOf(t).MethodByName(colnames[i])
					if v.Kind() == reflect.Invalid {
						return fmt.Errorf("unknown field or method %s", cols[i])
					}
					v = v.Call([]reflect.Value{})[0]
				}
				x, ok := v.Interface().(transrpc.ByteCount)
				if !ok {
					row[i] = fmt.Sprintf("%v", v)
					continue
				}
				totals[i] += x
				hasTotals, display[i] = true, true
				row[i] = args.formatByteCount(x, strings.Contains(cols[i], "rate"))
			}
			tbl.Append(row)
		}

		if !args.Output.NoTotals && hasTotals && len(torrents) > 0 {
			row := make([]string, len(cols))
			for i := 0; i < len(totals); i++ {
				if !display[i] {
					continue
				}
				row[i] = args.formatByteCount(totals[i], strings.Contains(cols[i], "rate"))
			}
			tbl.Append(row)
		}

		tbl.Render()
		return nil
	}
}

// encodeJSON encodes the torrent results to the writer as a table.
func (tr *TorrentResult) encodeJSON(w io.Writer, args *Args, cl *transrpc.Client) error {
	res, err := cl.TorrentGet(context.Background(), convTorrentIDs(tr.torrents)...)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(res.Torrents)
}

// encodeYaml encodes the torrent results to the writer as a table.
func (tr *TorrentResult) encodeYaml(w io.Writer, args *Args, cl *transrpc.Client) error {
	res, err := cl.TorrentGet(context.Background(), convTorrentIDs(tr.torrents)...)
	if err != nil {
		return err
	}
	for _, t := range res.Torrents {
		fmt.Fprintln(w, "---")
		if err = yaml.NewEncoder(w).Encode(t); err != nil {
			return err
		}
	}
	return nil
}

// encodeFlat encodes the torrent results to the writer as a flat key map.
func (tr *TorrentResult) encodeFlat(w io.Writer, args *Args, cl *transrpc.Client) error {
	res, err := cl.TorrentGet(context.Background(), convTorrentIDs(tr.torrents)...)
	if err != nil {
		return err
	}
	for i, t := range res.Torrents {
		if i != 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "[torrent %s]\n", t.ShortHash())
		m := make(map[string]string)
		addFieldsToMap(m, "", reflect.ValueOf(t))
		var keys []string
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if m[k] == "" {
				continue
			}
			fmt.Fprintf(w, "%s=%s\n", k, m[k])
		}
	}
	return nil
}

// PeerResult is a wrapper type for slice of *transrpc.Torrent's that
// satisfies the tblfmt.ResultSet interface.
type PeerResult struct {
	torrents []transrpc.Torrent
}

// NewPeerResult creates a new torrent result output encoder for the passed
// torrents.
func NewPeerResult(torrents []transrpc.Torrent) *PeerResult {
	return &PeerResult{
		torrents: torrents,
	}
}

// Encode encodes the torrent result using the settings in args to the
// io.Writer.
func (tr *PeerResult) Encode(w io.Writer, args *Args, cl *transrpc.Client) error {
	var f func(io.Writer, *Args, *transrpc.Client) error
	switch output := strings.SplitN(args.Output.Output, "=", 2); output[0] {
	case "table":
		cols := []string{"address", "clientName", "rateToClient", "rateToPeer", "progress", "shortHash"}
		if len(output) > 1 {
			cols = strings.Split(output[1], ",")
		}
		f = tr.encodeTable(cols...)
	case "wide":
		f = tr.encodeTable("address", "clientName", "isEncrypted", "port", "rateToClient", "rateToPeer", "progress", "shortHash")
	case "json":
		f = tr.encodeJSON
	case "yaml":
		f = tr.encodeYaml
	case "flat":
		f = tr.encodeFlat
	default:
		return ErrInvalidOutputOptionSpecified
	}
	return f(w, args, cl)
}

// encodeTableColumns encodes the specified table results with the included
// columns.
func (tr *PeerResult) encodeTable(columns ...string) func(io.Writer, *Args, *transrpc.Client) error {
	return func(w io.Writer, args *Args, cl *transrpc.Client) error {
		// check that at least one column was non-empty
		var cols []string
		for i := 0; i < len(columns); i++ {
			c := strings.TrimSpace(columns[i])
			if c == "" {
				continue
			}
			cols = append(cols, c)
		}
		if len(cols) < 1 {
			return ErrMustSpecifyAtLeastOneOutputColumn
		}

		// build column mappings
		inverseCols := make(map[string]string, len(args.Output.ColumnNames))
		for k, v := range args.Output.ColumnNames {
			inverseCols[v] = k
		}
		headers := make([]string, len(cols))
		fieldnames := make([]string, len(cols))
		colnames := make([]string, len(cols))
		sortByField := ""
		sortBy := strings.TrimSpace(args.Output.SortBy)
		for i := 0; i < len(cols); i++ {
			if c, ok := inverseCols[cols[i]]; ok {
				cols[i] = c
			}
			headers[i] = cols[i]
			if h, ok := args.Output.ColumnNames[cols[i]]; ok {
				headers[i] = h
			}
			headers[i] = strings.ToUpper(headers[i])
			if cols[i] == "shortHash" {
				fieldnames[i] = "hashString"
			} else {
				fieldnames[i] = cols[i]
			}
			colnames[i] = snaker.ForceCamelIdentifier(cols[i])
			if sortBy == cols[i] || strings.EqualFold(sortBy, headers[i]) {
				sortByField = colnames[i]
			}
		}

		switch {
		case sortByField == "" && !args.Output.SortByWasSet:
			sortByField = colnames[0]
		case sortByField == "":
			return ErrSortByNotInColumnList
		}

		// make peers
		peers := reflect.MakeSlice(reflect.ValueOf(transrpc.Torrent{}).FieldByName("Peers").Type(), 0, 0)
		var torrenthashes []string
		for _, t := range tr.torrents {
			peers = reflect.AppendSlice(peers, reflect.ValueOf(t.Peers))
			hash := t.ShortHash()
			for _ = range t.Peers {
				torrenthashes = append(torrenthashes, hash)
			}
		}

		// sort
		sort.Slice(peers.Interface(), func(i, j int) bool {
			a := peers.Index(i).FieldByName(sortByField)
			if a.Kind() == reflect.Invalid {
				a = peers.Index(i).MethodByName(sortByField)
				if a.Kind() == reflect.Invalid {
					panic("unknown torrent field or method " + args.Output.SortBy)
				}
				a = a.Call([]reflect.Value{})[0]
			}
			b := peers.Index(j).FieldByName(sortByField)
			if b.Kind() == reflect.Invalid {
				b = peers.Index(j).MethodByName(sortByField)
				if b.Kind() == reflect.Invalid {
					panic("unknown torrent field or method " + args.Output.SortBy)
				}
				b = b.Call([]reflect.Value{})[0]
			}

			switch x := a.Interface().(type) {
			case string:
				if args.Output.SortOrder == "desc" {
					return x > b.Interface().(string)
				}
				return x < b.Interface().(string)
			case int64:
				if args.Output.SortOrder == "desc" {
					return x > b.Interface().(int64)
				}
				return x < b.Interface().(int64)
			case float64:
				if args.Output.SortOrder == "desc" {
					return x > b.Interface().(float64)
				}
				return x < b.Interface().(float64)
			case transrpc.Percent:
				if args.Output.SortOrder == "desc" {
					return x > b.Interface().(transrpc.Percent)
				}
				return x < b.Interface().(transrpc.Percent)
			case transrpc.Status:
				if args.Output.SortOrder == "desc" {
					return x > b.Interface().(transrpc.Status)
				}
				return x < b.Interface().(transrpc.Status)
			case transrpc.Priority:
				if args.Output.SortOrder == "desc" {
					return x > b.Interface().(transrpc.Priority)
				}
				return x < b.Interface().(transrpc.Priority)
			case transrpc.Encryption:
				if args.Output.SortOrder == "desc" {
					return x > b.Interface().(transrpc.Encryption)
				}
				return x < b.Interface().(transrpc.Encryption)
			case transrpc.ByteCount:
				if args.Output.SortOrder == "desc" {
					return x > b.Interface().(transrpc.ByteCount)
				}
				return x < b.Interface().(transrpc.ByteCount)
			case transrpc.Duration:
				if args.Output.SortOrder == "desc" {
					return x > b.Interface().(transrpc.Duration)
				}
				return x < b.Interface().(transrpc.Duration)
			case transrpc.Time:
				if args.Output.SortOrder == "desc" {
					return time.Time(x).After(time.Time(b.Interface().(transrpc.Time)))
				}
				return time.Time(x).Before(time.Time(b.Interface().(transrpc.Time)))
			case transrpc.Bool:
				return false
			default:
				panic(fmt.Sprintf("unknown comparison type %T", a.Interface()))
			}
		})

		// tablewriter package is temporary until tblfmt is fixed
		tbl := tablewriter.NewWriter(w)
		if !args.Output.NoHeaders {
			tbl.SetHeader(headers)
		}
		tbl.SetAutoWrapText(false)
		tbl.SetAutoFormatHeaders(true)
		tbl.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		tbl.SetAlignment(tablewriter.ALIGN_LEFT)
		tbl.SetCenterSeparator("")
		tbl.SetColumnSeparator("")
		tbl.SetRowSeparator("")
		tbl.SetHeaderLine(false)
		tbl.SetBorder(false)
		tbl.SetTablePadding("\t") // pad with tabs
		tbl.SetNoWhiteSpace(true)

		// process
		hasTotals := false
		display, totals := make([]bool, len(cols)), make([]transrpc.ByteCount, len(cols))
		for j := 0; j < peers.Len(); j++ {
			row := make([]string, len(cols))
			for i := 0; i < len(cols); i++ {
				if cols[i] == "shortHash" {
					row[i] = torrenthashes[j]
					continue
				}

				v := peers.Index(j).FieldByName(colnames[i])
				if v.Kind() == reflect.Invalid {
					v = peers.Index(j).MethodByName(colnames[i])
					if v.Kind() == reflect.Invalid {
						return fmt.Errorf("unknown field or method %s", cols[i])
					}
					v = v.Call([]reflect.Value{})[0]
				}

				x, ok := v.Interface().(transrpc.ByteCount)
				if !ok {
					row[i] = fmt.Sprintf("%v", v)
					continue
				}

				totals[i] += x
				hasTotals, display[i] = true, true
				row[i] = args.formatByteCount(x, strings.Contains(cols[i], "rate"))
			}
			tbl.Append(row)
		}

		if !args.Output.NoTotals && hasTotals && peers.Len() > 0 {
			row := make([]string, len(cols))
			for i := 0; i < len(totals); i++ {
				if !display[i] {
					continue
				}
				row[i] = args.formatByteCount(totals[i], strings.Contains(cols[i], "rate"))
			}
			tbl.Append(row)
		}

		tbl.Render()
		return nil
	}
}

// encodeJSON encodes the torrent results to the writer as a table.
func (tr *PeerResult) encodeJSON(w io.Writer, args *Args, cl *transrpc.Client) error {
	m := make(map[string]interface{})
	for _, t := range tr.torrents {
		m[t.HashString] = t.Peers
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(m)
}

// encodeYaml encodes the torrent results to the writer as a table.
func (tr *PeerResult) encodeYaml(w io.Writer, args *Args, cl *transrpc.Client) error {
	for _, t := range tr.torrents {
		fmt.Fprintln(w, "---")
		if err := yaml.NewEncoder(w).Encode(map[string]interface{}{
			"hashString": t.HashString,
			"peers":      t.Peers,
		}); err != nil {
			return err
		}
	}
	return nil
}

// encodeFlat encodes the torrent results to the writer as a flat key map.
func (tr *PeerResult) encodeFlat(w io.Writer, args *Args, cl *transrpc.Client) error {
	res, err := cl.TorrentGet(context.Background(), convTorrentIDs(tr.torrents)...)
	if err != nil {
		return err
	}
	for j, t := range res.Torrents {
		fmt.Fprintf(w, "[peers %s]\n", t.ShortHash())
		if j != 0 {
			fmt.Fprintln(w)
		}
		for i, p := range t.Peers {
			m := make(map[string]string)
			addFieldsToMap(m, strconv.Itoa(i)+".", reflect.ValueOf(p))
			var keys []string
			for k := range m {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				if m[k] == "" {
					continue
				}
				fmt.Fprintf(w, "%s=%s\n", k, m[k])
			}
		}
	}
	return nil
}

// FileResult is a wrapper type for slice of *transrpc.Torrent.
type FileResult struct {
	torrents []transrpc.Torrent
}

// NewFileResult creates a new torrent result output encoder for the passed
// torrents.
func NewFileResult(torrents []transrpc.Torrent) *FileResult {
	return &FileResult{
		torrents: torrents,
	}
}

// Encode encodes the torrent result using the settings in args to the
// io.Writer.
func (tr *FileResult) Encode(w io.Writer, args *Args, cl *transrpc.Client) error {
	return nil
}
