package format

import (
	"io"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Formatter interface {
	Write(headers []string, rows [][]string) error
	WriteObject(obj map[string]string) error
}

var Global Formatter = &CSV{w: os.Stdout}

func New(format string, noHeader bool, w io.Writer) Formatter {
	switch strings.ToLower(format) {
	case "json":
		return &JSON{w: w}
	case "table":
		return &Table{w: w}
	default:
		return &CSV{w: w, noHeader: noHeader}
	}
}

func ResolveFormat(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}
	if v := os.Getenv("SPRITZ_OUTPUT"); v != "" {
		return v
	}
	if v := viper.GetString("output"); v != "" {
		return v
	}
	return "csv"
}
