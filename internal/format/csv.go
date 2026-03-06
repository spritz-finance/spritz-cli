package format

import (
	"encoding/csv"
	"io"
)

type CSV struct {
	w        io.Writer
	noHeader bool
}

func (f *CSV) Write(headers []string, rows [][]string) error {
	w := csv.NewWriter(f.w)
	if !f.noHeader {
		if err := w.Write(headers); err != nil {
			return err
		}
	}
	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func (f *CSV) WriteObject(obj map[string]string) error {
	w := csv.NewWriter(f.w)
	var headers, values []string
	for k, v := range obj {
		headers = append(headers, k)
		values = append(values, v)
	}
	if !f.noHeader {
		if err := w.Write(headers); err != nil {
			return err
		}
	}
	if err := w.Write(values); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}
