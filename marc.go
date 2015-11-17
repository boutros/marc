package marc

import (
	"fmt"
	"io"
)

// Record represents a bibliographic record, serializable to MARC formt
type Record struct {
	leader     [24]byte
	ctrlFields []cField
	dataFields []dField
}

type cField struct {
	Tag   string // 3 chars
	Value string // if Tag == "000"; 40 chars
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

// DumpTo dumps a Record to the give writer
func (r Record) DumpTo(w io.Writer, colors bool) {
	bold, reset, faint, green := "", "", "", ""
	if colors {
		bold = "\x1b[1m"
		reset = "\x1b[0m"
		faint = "\x1b[2m"
		green = "\x1b[32m"
	}

	orBlank := func(s string) string {
		if len(s) == 0 || s == " " {
			return "_"
		}
		return s
	}
	fmt.Fprintf(w, "leader\n")
	for _, c := range r.ctrlFields {
		fmt.Fprintf(w, "%s%s%s %s\n", bold, c.Tag, reset, c.Value)
	}
	for _, d := range r.dataFields {
		fmt.Fprintf(w, "%s%s %s%s%s%s ",
			bold, d.Tag, faint, orBlank(d.Ind1), orBlank(d.Ind2), reset)
		for _, s := range d.SubFields {
			//if i > 0 {
			//	fmt.Fprintf(w, "\n       ")
			//}
			fmt.Fprintf(w, "%s|%s %s%s ", green, s.Code, reset, s.Value)
		}
		fmt.Fprintf(w, "\n")
	}
	fmt.Fprintf(w, "\n")
}
