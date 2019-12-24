package main

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/kenshaw/transctl/tcutil"
	"github.com/kenshaw/transctl/transrpc"
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

	// sortOrderWasSet is the sort order was set toggle.
	sortOrderWasSet bool

	// yamlName is the yaml key to encode with, otherwise encodes highest level.
	yamlName string

	// flatName is the flat name to encode with.
	flatName string

	// flatKey is the flat index key to use.
	flatKey string

	// flatIndex is the flat index field.
	flatIndex string

	// columnNames is the column name map.
	columnNames map[string]string

	// formatBytes is the byte format func.
	formatBytes func(tcutil.ByteFormatter) string

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
	res := &Result{res: z, index: "hashString"}
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
	case res.output == "all":
		cols, err := res.buildAllColumns()
		if err != nil {
			return err
		}
		f = res.encodeTable(cols...)
	case res.output == "json":
		f = res.encodeJSON
	case res.output == "yaml":
		f = res.encodeYaml
	case res.output == "flat":
		f = res.encodeFlat
	case strings.HasPrefix(res.output, "cols="):
		f = res.encodeTable(strings.Split(res.output[5:], ",")...)
	default:
		return ErrInvalidOutputOptionSpecified
	}
	return f(w)
}

// buildAllColumns builds all column names from the reflected result type.
func (res *Result) buildAllColumns() ([]string, error) {
	wideMap := make(map[string]bool)
	cols := res.wideCols
	for _, k := range cols {
		wideMap[k] = true
	}
	typ := res.res.Type().Elem()
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		tag := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
		if all := f.Tag.Get("all"); all != "" {
			tag = all
		}
		if tag == "" || tag == "-" || f.Type.Kind() == reflect.Slice || wideMap[tag] {
			continue
		}
		cols = append(cols, tag)
	}
	for i := 0; i < typ.NumMethod(); i++ {
		m := typ.Method(i)
		name := snaker.ForceLowerCamelIdentifier(m.Name)
		if wideMap[name] {
			continue
		}
		if m.Type.NumIn() == 1 && m.Type.NumOut() == 1 && m.Type.Out(0).Kind() == reflect.String {
			cols = append(cols, name)
		}
	}
	return cols, nil
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
			headers[i] = strings.ToUpper(strings.ReplaceAll(snaker.CamelToSnake(headers[i]), "_", " "))
			colnames[i] = snaker.ForceCamelIdentifier(cols[i])
			if sortBy == cols[i] || strings.EqualFold(sortBy, headers[i]) {
				sortByField = colnames[i]
			}
		}

		// determine sort by and order
		switch {
		case sortByField == "" && !res.sortByWasSet:
			sortByField = colnames[0]
		case sortByField == "":
			return ErrSortByNotInColumnList
		}
		dir := res.sortOrder
		if !res.sortOrderWasSet {
			typ, ok := readFieldOrMethodType(res.res.Type().Elem(), sortByField)
			if ok {
				z := reflect.Zero(typ).Interface()
				if _, ok = z.(tcutil.ByteFormatter); ok {
					dir = "desc"
				}
				if _, ok = z.(tcutil.Percent); ok {
					dir = "desc"
				}
			}
		}
		res.sort(sortByField, dir == "desc")

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
		display, totals := make([]bool, len(cols)), make([]tcutil.ByteFormatter, len(cols))
		for j := 0; j < res.res.Len(); j++ {
			row := make([]string, len(cols))
			for i := 0; i < len(cols); i++ {
				v, err := readFieldOrMethod(res.res.Index(j), colnames[i])
				if err != nil {
					return err
				}
				x, ok := v.(tcutil.ByteFormatter)
				if !ok {
					row[i] = fmt.Sprintf("%v", v)
					continue
				}
				row[i] = res.formatBytes(x)
				if !res.noTotals {
					if totals[i] == nil {
						totals[i] = reflect.Zero(reflect.TypeOf(x)).Interface().(tcutil.ByteFormatter)
					}
					totals[i] = totals[i].Add(x).(tcutil.ByteFormatter)
					hasTotals, display[i] = true, true
				}
			}
			tbl.Append(row)
		}

		if !res.noTotals && hasTotals && res.res.Len() > 0 {
			row := make([]string, len(cols))
			for i := 0; i < len(totals); i++ {
				if !display[i] {
					continue
				}
				row[i] = res.formatBytes(totals[i])
			}
			tbl.Append(row)
		}

		tbl.Render()
		return nil
	}
}

// sort sorts the results based on the the specified sort by field.
func (res *Result) sort(sortBy string, sortDesc bool) {
	if res.res.Len() == 0 {
		return
	}
	sort.Slice(res.res.Interface(), func(i, j int) bool {
		a, err := readFieldOrMethod(res.res.Index(i), sortBy)
		if err != nil {
			panic(err)
		}
		b, err := readFieldOrMethod(res.res.Index(j), sortBy)
		if err != nil {
			panic(err)
		}
		switch x := a.(type) {
		case string:
			if sortDesc {
				return x > b.(string)
			}
			return x < b.(string)
		case int64:
			if sortDesc {
				return x > b.(int64)
			}
			return x < b.(int64)
		case float64:
			if sortDesc {
				return x > b.(float64)
			}
			return x < b.(float64)
		case tcutil.ByteFormatter:
			if sortDesc {
				return x.Int64() > b.(tcutil.ByteFormatter).Int64()
			}
			return x.Int64() < b.(tcutil.ByteFormatter).Int64()
		case tcutil.Percent:
			if sortDesc {
				return x > b.(tcutil.Percent)
			}
			return x < b.(tcutil.Percent)
		case transrpc.Status:
			if sortDesc {
				return x > b.(transrpc.Status)
			}
			return x < b.(transrpc.Status)
		case transrpc.Priority:
			if sortDesc {
				return x > b.(transrpc.Priority)
			}
			return x < b.(transrpc.Priority)
		case transrpc.Encryption:
			if sortDesc {
				return x > b.(transrpc.Encryption)
			}
			return x < b.(transrpc.Encryption)
		case transrpc.Duration:
			if sortDesc {
				return x > b.(transrpc.Duration)
			}
			return x < b.(transrpc.Duration)
		case transrpc.Time:
			if sortDesc {
				return time.Time(x).After(time.Time(b.(transrpc.Time)))
			}
			return time.Time(x).Before(time.Time(b.(transrpc.Time)))
		case transrpc.Bool:
			if sortDesc {
				return x != b.(transrpc.Bool)
			}
			return x == b.(transrpc.Bool)
		default:
			panic(fmt.Sprintf("unknown comparison type %T", a))
		}
	})
}

// encodeJSON encodes the results to the writer as JSON.
func (res *Result) encodeJSON(w io.Writer) error {
	if res.res.Len() == 0 {
		return nil
	}
	m := make(map[string]interface{})
	for i := 0; i < res.res.Len(); i++ {
		v := res.res.Index(i)
		key, err := readFieldOrMethodString(v, res.index)
		if err != nil {
			return err
		}
		if res.yamlName == "" && res.flatKey == "" {
			m[key] = v.Interface()
		} else {
			if _, ok := m[key]; !ok {
				m[key] = reflect.MakeSlice(reflect.SliceOf(v.Type()), 0, 0).Interface()
			}
			m[key] = reflect.Append(reflect.ValueOf(m[key]), v).Interface()
		}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(m)
}

// encodeYaml encodes the results to the writer as YAML.
func (res *Result) encodeYaml(w io.Writer) error {
	if res.res.Len() == 0 {
		return nil
	}
	if res.yamlName != "" {
		var last string
		var m map[string]interface{}
		for i := 0; i < res.res.Len(); i++ {
			v := res.res.Index(i)
			key, err := readFieldOrMethodString(v, res.index)
			if err != nil {
				return err
			}
			if last != key {
				if m != nil {
					fmt.Fprintln(w, "---")
					if err = yaml.NewEncoder(w).Encode(m); err != nil {
						return err
					}
				}
				m = map[string]interface{}{
					res.index: key,
				}
			}
			if _, ok := m[res.yamlName]; !ok {
				m[res.yamlName] = reflect.MakeSlice(reflect.SliceOf(v.Type()), 0, 0).Interface()
			}
			m[res.yamlName] = reflect.Append(reflect.ValueOf(m[res.yamlName]), v).Interface()
			last = key
		}
		fmt.Fprintln(w, "---")
		return yaml.NewEncoder(w).Encode(m)
	}

	for i := 0; i < res.res.Len(); i++ {
	}
	return nil
}

// encodeFlat encodes the results to the writer as a flat key map.
func (res *Result) encodeFlat(w io.Writer) error {
	if res.res.Len() == 0 {
		return nil
	}
	var last string
	for i := 0; i < res.res.Len(); i++ {
		key, err := readFieldOrMethodString(res.res.Index(i), res.flatIndex)
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
func SortOrder(sortOrder string, sortOrderWasSet bool) ResultOption {
	return func(res *Result) {
		res.sortOrder, res.sortOrderWasSet = sortOrder, sortOrderWasSet
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

// FlatIndex is a result option to set the flat index key.
func FlatIndex(flatIndex string) ResultOption {
	return func(res *Result) {
		res.flatIndex = flatIndex
	}
}

// ColumnNames is a result option to set the column names map.
func ColumnNames(columnNames map[string]string) ResultOption {
	return func(res *Result) {
		res.columnNames = columnNames
	}
}

// FormatBytes is a result option to set the func used to format bytes.
func FormatBytes(formatBytes func(tcutil.ByteFormatter) string) ResultOption {
	return func(res *Result) {
		res.formatBytes = formatBytes
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

// Index sets the index field to use.
func Index(index string) ResultOption {
	return func(res *Result) {
		res.index = index
	}
}

// readFieldOrMethodType returns the type of the field or method name on x.
func readFieldOrMethodType(x reflect.Type, name string) (reflect.Type, bool) {
	f, ok := x.FieldByName(name)
	if ok {
		return f.Type, true
	}
	m, ok := x.MethodByName(name)
	if ok {
		return m.Type.Out(0), ok
	}
	return nil, false
}

// readFieldOrMethod returns the field or method name declared on x.
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
