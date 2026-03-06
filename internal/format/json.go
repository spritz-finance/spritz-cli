package format

import (
	"encoding/json"
	"io"
)

type JSON struct {
	w io.Writer
}

func (f *JSON) Write(headers []string, rows [][]string) error {
	var objects []map[string]string
	for _, row := range rows {
		obj := make(map[string]string, len(headers))
		for i, h := range headers {
			if i < len(row) {
				obj[h] = row[i]
			}
		}
		objects = append(objects, obj)
	}
	enc := json.NewEncoder(f.w)
	enc.SetIndent("", "  ")
	return enc.Encode(objects)
}

func (f *JSON) WriteObject(obj map[string]string) error {
	enc := json.NewEncoder(f.w)
	enc.SetIndent("", "  ")
	return enc.Encode(obj)
}
