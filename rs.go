package main

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/kenshaw/transrpc"
	"github.com/knq/snaker"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

// Result is a wraper for a slice of structs. Uses reflection to iterate the
// items and output based on
type Result struct {
	// res is the value to iterate.
	res reflect.Value

	// tableCols are the default table columns to display.
	tableCols []string

	// wideCols are the wide table columns to display.
	wideCols []string

	// output is the output type.
	output string

	// index is the field name to use for associative output.
	index string

	// sortBy is the field to sort by.
	sortBy string

	// sortByWasSet is the sort by was set toggle.
	sortByWasSet bool

	// sortOrder is the sort order, either asc or desc.
	sortOrder string

	// yamlName is the yaml key to encode with, otherwise encodes highest level.
	yamlName string

	// flatName is the flat name to encode with.
	flatName string

	// flatKey is the flat index key to use.
	flatKey string

	// columnNames is the column name map.
	columnNames map[string]string

	// formatByteCount is the byte count format func.
	formatByteCount func(transrpc.ByteCount, bool) string

	// noHeaders is the no headers output toggle.
	noHeaders bool

	// noTotals is the no totals output toggle.
	noTotals bool
}

// NewResult creates a new reflection result for v.
func NewResult(v interface{}, options ...ResultOption) *Result {
	z := reflect.ValueOf(v)
	if z.Kind() != reflect.Slice && z.Elem().Kind() != reflect.Struct {
		panic("v must be []struct")
	}
	res := &Result{res: z, index: "shortHash"}
	for _, o := range options {
		o(res)
	}
	return res
}

// Encode encodes the result using the settings in args to the io.Writer.
func (res *Result) Encode(w io.Writer) error {
	var f func(io.Writer) error
	switch {
	case res.output == "table":
		f = res.encodeTable(res.tableCols...)
	case res.output == "wide":
		f = res.encodeTable(res.wideCols...)
	case res.output == "json":
		f = res.encodeJSON
	case res.output == "yaml":
		f = res.encodeYaml
	case res.output == "flat":
		f = res.encodeFlat
	case strings.HasPrefix(res.output, "table="):
		f = res.encodeTable(strings.Split(res.output[6:], ",")...)
	default:
		return ErrInvalidOutputOptionSpecified
	}
	return f(w)
}

// encodeTable encodes the results to the writer as at table.
func (res *Result) encodeTable(columns ...string) func(w io.Writer) error {
	return func(w io.Writer) error {
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
		inverseCols := make(map[string]string, len(res.columnNames))
		for k, v := range res.columnNames {
			inverseCols[v] = k
		}
		headers := make([]string, len(cols))
		colnames := make([]string, len(cols))
		sortByField := ""
		sortBy := strings.TrimSpace(res.sortBy)
		for i := 0; i < len(cols); i++ {
			if c, ok := inverseCols[cols[i]]; ok {
				cols[i] = c
			}
			headers[i] = cols[i]
			if h, ok := res.columnNames[cols[i]]; ok {
				headers[i] = h
			}

			if headers[i] == "percent" {
			}

			headers[i] = strings.ToUpper(headers[i])
			colnames[i] = snaker.ForceCamelIdentifier(cols[i])
			if sortBy == cols[i] || strings.EqualFold(sortBy, headers[i]) {
				sortByField = colnames[i]
			}
		}

		switch {
		case sortByField == "" && !res.sortByWasSet:
			sortByField = colnames[0]
		case sortByField == "":
			return ErrSortByNotInColumnList
		}

		res.sort(sortByField)

		// tablewriter package is temporary until tblfmt is fixed
		tbl := tablewriter.NewWriter(w)
		if !res.noHeaders {
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
		for j := 0; j < res.res.Len(); j++ {
			row := make([]string, len(cols))
			for i := 0; i < len(cols); i++ {
				v, err := readFieldOrMethod(res.res.Index(j), colnames[i])
				if err != nil {
					return err
				}
				x, ok := v.(transrpc.ByteCount)
				if !ok {
					row[i] = fmt.Sprintf("%v", v)
					continue
				}
				totals[i] += x
				hasTotals, display[i] = true, true
				row[i] = res.formatByteCount(x, strings.Contains(cols[i], "rate"))
			}
			tbl.Append(row)
		}

		if !res.noTotals && hasTotals && res.res.Len() > 0 {
			row := make([]string, len(cols))
			for i := 0; i < len(totals); i++ {
				if !display[i] {
					continue
				}
				row[i] = res.formatByteCount(totals[i], strings.Contains(cols[i], "rate"))
			}
			tbl.Append(row)
		}

		tbl.Render()
		return nil
	}
}

// sort sorts the results based on the the specified sort by field.
func (res *Result) sort(sortByField string) {
	if res.res.Len() == 0 {
		return
	}
	sort.Slice(res.res.Interface(), func(i, j int) bool {
		a, err := readFieldOrMethod(res.res.Index(i), sortByField)
		if err != nil {
			panic(err)
		}
		b, err := readFieldOrMethod(res.res.Index(j), sortByField)
		if err != nil {
			panic(err)
		}
		switch x := a.(type) {
		case string:
			if res.sortOrder == "desc" {
				return x > b.(string)
			}
			return x < b.(string)
		case int64:
			if res.sortOrder == "desc" {
				return x > b.(int64)
			}
			return x < b.(int64)
		case float64:
			if res.sortOrder == "desc" {
				return x > b.(float64)
			}
			return x < b.(float64)
		case transrpc.Percent:
			if res.sortOrder == "desc" {
				return x > b.(transrpc.Percent)
			}
			return x < b.(transrpc.Percent)
		case transrpc.Status:
			if res.sortOrder == "desc" {
				return x > b.(transrpc.Status)
			}
			return x < b.(transrpc.Status)
		case transrpc.Priority:
			if res.sortOrder == "desc" {
				return x > b.(transrpc.Priority)
			}
			return x < b.(transrpc.Priority)
		case transrpc.Encryption:
			if res.sortOrder == "desc" {
				return x > b.(transrpc.Encryption)
			}
			return x < b.(transrpc.Encryption)
		case transrpc.ByteCount:
			if res.sortOrder == "desc" {
				return x > b.(transrpc.ByteCount)
			}
			return x < b.(transrpc.ByteCount)
		case transrpc.Duration:
			if res.sortOrder == "desc" {
				return x > b.(transrpc.Duration)
			}
			return x < b.(transrpc.Duration)
		case transrpc.Time:
			if res.sortOrder == "desc" {
				return time.Time(x).After(time.Time(b.(transrpc.Time)))
			}
			return time.Time(x).Before(time.Time(b.(transrpc.Time)))
		case transrpc.Bool:
			return false
		default:
			panic(fmt.Sprintf("unknown comparison type %T", a))
		}
	})
}

// encodeJSON encodes the results to the writer as JSON.
func (res *Result) encodeJSON(w io.Writer) error {
	m := make(map[string]interface{})
	for i := 0; i < res.res.Len(); i++ {
		v := res.res.Index(i)
		key, err := readFieldOrMethodString(v, res.index)
		if err != nil {
			return err
		}
		m[key] = v.Interface()
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(m)
}

// encodeYaml encodes the results to the writer as YAML.
func (res *Result) encodeYaml(w io.Writer) error {
	for i := 0; i < res.res.Len(); i++ {
		fmt.Fprintln(w, "---")
		v := res.res.Index(i)
		z := v.Interface()
		if res.yamlName != "" {
			key, err := readFieldOrMethod(v, res.index)
			if err != nil {
				return err
			}
			z = map[string]interface{}{
				res.index:    key,
				res.yamlName: z,
			}
		}
		if err := yaml.NewEncoder(w).Encode(z); err != nil {
			return err
		}
	}
	return nil
}

// encodeFlat encodes the results to the writer as a flat key map.
func (res *Result) encodeFlat(w io.Writer) error {
	var last string
	for i := 0; i < res.res.Len(); i++ {
		key, err := readFieldOrMethodString(res.res.Index(i), res.index)
		if err != nil {
			return err
		}
		if last != key {
			if i != 0 {
				fmt.Fprintln(w)
			}
			fmt.Fprintf(w, "[%s %q]\n", res.flatName, key)
		}
		last = key
		m := make(map[string]string)
		var prefix string
		if res.flatKey != "" {
			s, err := readFieldOrMethodString(res.res.Index(i), res.flatKey)
			if err != nil {
				return err
			}
			prefix = s + "."
		}
		addFieldsToMap(m, prefix, res.res.Index(i))
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

// ResultOption is a result option.
type ResultOption = func(*Result)

// TableColumns is a result option to set the table column output.
func TableColumns(tableCols ...string) ResultOption {
	return func(res *Result) {
		res.tableCols = tableCols
	}
}

// WideColumns is a result option to set the wide column output.
func WideColumns(wideCols ...string) ResultOption {
	return func(res *Result) {
		res.wideCols = wideCols
	}
}

// Output is a result option to set the output format.
func Output(output string) ResultOption {
	return func(res *Result) {
		res.output = output
	}
}

// SortBy is a result option to set the sort by field.
func SortBy(sortBy string, sortByWasSet bool) ResultOption {
	return func(res *Result) {
		res.sortBy, res.sortByWasSet = sortBy, sortByWasSet
	}
}

// SortOrder is a result option to set the sort order direction (asc or desc).
func SortOrder(sortOrder string) ResultOption {
	return func(res *Result) {
		res.sortOrder = sortOrder
	}
}

// YamlName is a result option to set the yaml key used for yaml output.
func YamlName(yamlName string) ResultOption {
	return func(res *Result) {
		res.yamlName = yamlName
	}
}

// FlatName is a result option to set the flat key used for output.
func FlatName(flatName string) ResultOption {
	return func(res *Result) {
		res.flatName = flatName
	}
}

// FlatKey is a result option to set the flat key field.
func FlatKey(flatKey string) ResultOption {
	return func(res *Result) {
		res.flatKey = flatKey
	}
}

// ColumnNames is a result option to set the column names map.
func ColumnNames(columnNames map[string]string) ResultOption {
	return func(res *Result) {
		res.columnNames = columnNames
	}
}

// FormatByteCount is a result option to set the func used to format byte
// counts.
func FormatByteCount(formatByteCount func(transrpc.ByteCount, bool) string) ResultOption {
	return func(res *Result) {
		res.formatByteCount = formatByteCount
	}
}

// NoHeaders is a result option to set the no headers toggle.
func NoHeaders(noHeaders bool) ResultOption {
	return func(res *Result) {
		res.noHeaders = noHeaders
	}
}

// NoTotals is a result option to set the no totals toggle.
func NoTotals(noTotals bool) ResultOption {
	return func(res *Result) {
		res.noTotals = noTotals
	}
}

// readFieldOrMethod returns the field or method name declared on v.
func readFieldOrMethod(x reflect.Value, name string) (interface{}, error) {
	name = snaker.ForceCamelIdentifier(name)
	v := x.FieldByName(name)
	if v.Kind() == reflect.Invalid {
		v = x.MethodByName(name)
		if v.Kind() == reflect.Invalid {
			return nil, fmt.Errorf("unknown field or method %q", name)
		}
		v = v.Call([]reflect.Value{})[0]
	}
	return v.Interface(), nil
}

// readFieldOrMethodString returns the field or method name declared on v as a string
func readFieldOrMethodString(x reflect.Value, name string) (string, error) {
	v, err := readFieldOrMethod(x, name)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", v), nil
}
