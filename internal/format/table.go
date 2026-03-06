package format

import (
	"io"

	"github.com/olekukonko/tablewriter"
)

type Table struct {
	w io.Writer
}

func (f *Table) Write(headers []string, rows [][]string) error {
	t := tablewriter.NewWriter(f.w)
	t.Header(toAny(headers)...)
	for _, row := range rows {
		t.Append(toAny(row)...)
	}
	return t.Render()
}

func (f *Table) WriteObject(obj map[string]string) error {
	var headers, values []string
	for k, v := range obj {
		headers = append(headers, k)
		values = append(values, v)
	}
	return f.Write(headers, [][]string{values})
}

func toAny(s []string) []any {
	a := make([]any, len(s))
	for i, v := range s {
		a[i] = v
	}
	return a
}
