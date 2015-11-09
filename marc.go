package marc

import (
	"fmt"
	"io"
)

type Record struct {
	leader     [24]byte
	ctrlFields []cField
	dataFields []dField
}

type cField struct {
	Tag   string // 3 chars
	Field string // if Tag == "000"; 40 chars
}

type dField struct {
	Tag       string // 3 chars
	Ind1      string // 1 char
	Ind2      string // 1 char
	SubFields []subField
}

type subField struct {
	Code  string // 1 char
	Value string
}

func (r Record) DumpTo(w io.Writer, colors bool) {
	bold, reset, faint := "", "", ""
	if colors {
		bold = "\x1b[1m"
		reset = "\x1b[0m"
		faint = "\x1b[2m"
	}

	orBlank := func(s string) string {
		if len(s) == 0 || s == " " {
			return "_"
		}
		return s
	}
	fmt.Fprintf(w, "leader\n")
	for _, c := range r.ctrlFields {
		fmt.Fprintf(w, "%s%s%s %s\n", bold, c.Tag, reset, c.Field)
	}
	for _, d := range r.dataFields {
		fmt.Fprintf(w, "%s%s %s%s%s%s ",
			bold, d.Tag, faint, orBlank(d.Ind1), orBlank(d.Ind2), reset)
		for _, s := range d.SubFields {
			fmt.Fprintf(w, "%s[%s] %s%s ", faint, s.Code, reset, s.Value)
		}
		fmt.Fprintf(w, "\n")
	}
	fmt.Fprintf(w, "\n")
}
